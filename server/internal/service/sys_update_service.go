package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

const (
	defaultInstallDir = "/opt/xt-pms"
	defaultService    = "xt-pms"
	updateTmpDir      = "/tmp/xt-pms-update"
	updateManifest    = "/tmp/xt-pms-update-manifest.json"
)

// getServiceName 获取 service 名称
// 优先从配置读取，其次自动检测（读 cgroup），最后 fallback xt-pms
func getServiceName() string {
	if config.GlobalConfig != nil && config.GlobalConfig.App.ServiceName != "" {
		return config.GlobalConfig.App.ServiceName
	}
	if svc := detectServiceName(); svc != "" {
		return svc
	}
	return defaultService
}

// detectServiceName 从 /proc/self/cgroup 中解析 systemd service 名
// 无论 deploy.sh 部署还是 1Panel 部署，systemd 都会在 cgroup 里留下 service 名
func detectServiceName() string {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, ".service") {
			// 格式：0::/system.slice/xt-pms.service 或 0::/system.slice/1panel-xtpms.service
			start := strings.LastIndex(line, "/")
			end := strings.Index(line[start:], ".service")
			if start >= 0 && end >= 0 {
				return line[start+1 : start+end]
			}
		}
	}
	return ""
}

type SysUpdateService struct{}

func NewSysUpdateService() *SysUpdateService {
	return &SysUpdateService{}
}

// InitUpdateService 服务启动时调用，执行待处理的数据库增量迁移
// 迁移 SQL 文件放在 migrations/updates/ 目录下，详见 migration_service.go 顶部注释
func InitUpdateService() error {
	return NewMigrationService().RunPendingMigrations()
}

// releaseResponse 统一适配 Gitee/GitHub Releases API 响应
// Gitee: published_at 为空，用 created_at；GitHub: 用 published_at
type releaseResponse struct {
	TagName     string `json:"tag_name"`
	Name        string `json:"name"`
	Body        string `json:"body"`
	PublishedAt string `json:"published_at"`
	CreatedAt   string `json:"created_at"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// getPublishedAt 返回发布时间，兼容 Gitee (created_at) 和 GitHub (published_at)
func (r *releaseResponse) getPublishedAt() string {
	if r.PublishedAt != "" {
		return r.PublishedAt
	}
	return r.CreatedAt
}

// compareVersions 比较两个 semver 版本号（去掉 v 前缀后比较）
// 支持 pre-release 比较：0.1.0-beta.2 > 0.1.0-beta.1，0.2.0 > 0.2.0-beta.1
func compareVersions(a, b string) int {
	coreA, preA := splitSemver(a)
	coreB, preB := splitSemver(b)

	// 先比较数字段
	partsA := strings.Split(coreA, ".")
	partsB := strings.Split(coreB, ".")

	minLen := len(partsA)
	if len(partsB) < minLen {
		minLen = len(partsB)
	}

	for i := 0; i < minLen; i++ {
		na := atoi(partsA[i])
		nb := atoi(partsB[i])
		if na < nb {
			return -1
		}
		if na > nb {
			return 1
		}
	}

	if len(partsA) < len(partsB) {
		return -1
	}
	if len(partsA) > len(partsB) {
		return 1
	}

	// 数字段相同，比较 pre-release 后缀
	// 无 pre-release > 有 pre-release（正式版大于预发布版）
	if preA == "" && preB != "" {
		return 1
	}
	if preA != "" && preB == "" {
		return -1
	}
	if preA == "" && preB == "" {
		return 0
	}

	// 都有 pre-release，逐段比较
	return comparePreRelease(preA, preB)
}

// splitSemver 分离版本号和 pre-release 后缀
func splitSemver(v string) (core, pre string) {
	idx := strings.Index(v, "-")
	if idx == -1 {
		return v, ""
	}
	return v[:idx], v[idx+1:]
}

// comparePreRelease 比较 pre-release 后缀段
// 纯数字段按数值比较，否则按字符串比较
func comparePreRelease(a, b string) int {
	segsA := strings.Split(a, ".")
	segsB := strings.Split(b, ".")

	minLen := len(segsA)
	if len(segsB) < minLen {
		minLen = len(segsB)
	}

	for i := 0; i < minLen; i++ {
		na, aIsNum := strconv.Atoi(segsA[i])
		nb, bIsNum := strconv.Atoi(segsB[i])

		if aIsNum == nil && bIsNum == nil {
			// 都是数字：数值比较
			if na < nb {
				return -1
			}
			if na > nb {
				return 1
			}
		} else {
			// 非数字：字符串比较
			if segsA[i] < segsB[i] {
				return -1
			}
			if segsA[i] > segsB[i] {
				return 1
			}
		}
	}

	if len(segsA) < len(segsB) {
		return -1
	}
	if len(segsA) > len(segsB) {
		return 1
	}
	return 0
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// CheckUpdate 查询 Releases API 检查是否有新版本
// 默认使用 config.yaml 中配置的 update_source_url（默认 Gitee API）
// overrideURL 可通过前端传参覆盖
func (s *SysUpdateService) CheckUpdate(currentVersion, overrideURL string) (*dto.CheckUpdateResponse, error) {
	current := strings.TrimPrefix(currentVersion, "v")

	// 优先使用前端传入的 URL，其次使用配置文件，最后默认 Gitee
	url := overrideURL
	if url == "" && config.GlobalConfig != nil {
		url = config.GlobalConfig.App.UpdateSourceURL
	}
	if url == "" {
		url = "https://gitee.com/api/v5/repos/kamyang-tech/xt-pms/releases/latest"
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "xt-pms-updater")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求更新源 API 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("更新源 API 返回 %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var release releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析更新源 API 响应失败: %w", err)
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	if latest == "" {
		return nil, fmt.Errorf("无法解析最新版本号")
	}

	// 版本比较：数值化比较各段，正确处理 0.10.0 > 0.2.0
	if compareVersions(current, latest) >= 0 {
		return &dto.CheckUpdateResponse{
			HasUpdate:      false,
			Current:        current,
			Latest:         latest,
		}, nil
	}

	// 匹配下载资源：tar.gz
	var tarAsset *struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	}
	var sha256Asset *struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	}
	for i := range release.Assets {
		a := &release.Assets[i]
		// 只匹配包含 xt-pms 的 tar.gz（排除 Gitee 自动生成的源码包）
		if strings.Contains(a.Name, "xt-pms") && strings.HasSuffix(a.Name, ".tar.gz") {
			tarAsset = a
		}
		if strings.HasSuffix(a.Name, ".sha256") {
			sha256Asset = a
		}
	}

	if tarAsset == nil {
		return nil, fmt.Errorf("未找到 tar.gz 下载资源")
	}

	resp2 := &dto.CheckUpdateResponse{
		HasUpdate:   true,
		Current:     current,
		Latest:      latest,
		Changelog:   release.Body,
		DownloadURL: tarAsset.BrowserDownloadURL,
		FileSize:    tarAsset.Size,
		PublishedAt: release.getPublishedAt(),
	}
	if sha256Asset != nil {
		resp2.SHA256URL = sha256Asset.BrowserDownloadURL
	}

	return resp2, nil
}

// DownloadUpdate 下载更新包，通过 callback 推送下载进度（用于 SSE）
func (s *SysUpdateService) DownloadUpdate(downloadURL string, progressFn func(percent int, downloaded, total int64)) (string, string, error) {
	if err := os.MkdirAll(updateTmpDir, 0755); err != nil {
		return "", "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 从 URL 提取文件名
	parts := strings.Split(downloadURL, "/")
	filename := parts[len(parts)-1]
	if !strings.HasSuffix(filename, ".tar.gz") {
		filename = "xt-pms-update.tar.gz"
	}
	filePath := filepath.Join(updateTmpDir, filename)

	client := &http.Client{Timeout: 30 * time.Minute} // 大文件需要较长超时
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("构造下载请求失败: %w", err)
	}
	req.Header.Set("User-Agent", "xt-pms-updater")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("下载请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	total := resp.ContentLength
	out, err := os.Create(filePath)
	if err != nil {
		return "", "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	var downloaded int64
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := out.Write(buf[:n]); err != nil {
				return "", "", fmt.Errorf("写入文件失败: %w", err)
			}
			downloaded += int64(n)
			if progressFn != nil && total > 0 {
				percent := int(downloaded * 100 / total)
				progressFn(percent, downloaded, total)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", "", fmt.Errorf("下载中断: %w", readErr)
		}
	}

	return filePath, "", nil
}

// ComputeSHA256 计算文件的 SHA256
func (s *SysUpdateService) ComputeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// WriteManifest 写入更新清单文件
func (s *SysUpdateService) WriteManifest(manifest *dto.UpdateManifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(updateManifest, data, 0644)
}

// DetectInstallDir 检测安装目录
func (s *SysUpdateService) DetectInstallDir() (string, error) {
	svc := getServiceName()
	// 1. 从 systemd 服务文件获取
	if out, err := exec.Command("systemctl", "show", "-p", "ExecStart", svc).Output(); err == nil {
		line := strings.TrimSpace(string(out))
		// ExecStart={ path=/opt/xt-pms/xt-pms ; argv[]=...
		if idx := strings.Index(line, "path="); idx != -1 {
			rest := line[idx+5:]
			rest = strings.Split(rest, " ;")[0]
			dir := filepath.Dir(rest)
			if dir != "" && dir != "." {
				return dir, nil
			}
		}
	}

	// 2. 检查默认目录
	if _, err := os.Stat(filepath.Join(defaultInstallDir, "xt-pms")); err == nil {
		return defaultInstallDir, nil
	}

	// 3. 从当前进程推测
	exe, err := os.Executable()
	if err == nil {
		return filepath.Dir(exe), nil
	}

	return "", fmt.Errorf("无法检测安装目录")
}

// RunSupervisorUpdate supervisor 模式更新：主进程 goroutine 自行完成替换，最后 syscall.Exec
// PID 保持不变，1Panel/Go 运行时等 supervisor 完全无感知，无需外部重启
func (s *SysUpdateService) RunSupervisorUpdate(m *dto.UpdateManifest) {
	installDir := m.InstallDir
	if installDir == "" {
		installDir = defaultInstallDir
	}

	// 等 HTTP 响应发完再动手
	time.Sleep(2 * time.Second)

	logger.Info("[supervisor-update] 开始执行更新",
		zap.String("from", m.FromVersion),
		zap.String("to", m.ToVersion),
	)

	// 1. 检查更新包
	if _, err := os.Stat(m.PackagePath); os.IsNotExist(err) {
		s.failUpdate(fmt.Errorf("更新包不存在: %s", m.PackagePath))
		return
	}

	// 2. SHA256 校验
	if m.SHA256 != "" {
		f, err := os.Open(m.PackagePath)
		if err != nil {
			s.failUpdate(fmt.Errorf("打开更新包失败: %w", err))
			return
		}
		h := sha256.New()
		io.Copy(h, f)
		f.Close()
		if fmt.Sprintf("%x", h.Sum(nil)) != m.SHA256 {
		s.failUpdate(fmt.Errorf("SHA256 校验失败"))
			return
		}
		logger.Info("[supervisor-update] SHA256 校验通过")
	}

	// 3. 全量备份
	backupDir := m.BackupDir
	if backupDir == "" {
		backupDir = installDir + ".bak." + time.Now().Format("20060102_150405")
	}
	logger.Info("[supervisor-update] 正在备份", zap.String("dir", backupDir))
	if err := exec.Command("cp", "-a", installDir, backupDir).Run(); err != nil {
		logger.Warn("[supervisor-update] 备份失败（继续更新）", zap.Error(err))
	}

	// 4. 解压到 staging 目录
	stagingDir := installDir + ".staging"
	os.RemoveAll(stagingDir)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		s.failUpdate(fmt.Errorf("创建 staging 目录失败: %w", err))
		return
	}
	defer os.RemoveAll(stagingDir)

	logger.Info("[supervisor-update] 正在解压", zap.String("dest", stagingDir))
	if out, err := exec.Command(
		"tar", "-xzf", m.PackagePath, "-C", stagingDir,
		"--strip-components=1", "--no-xattrs",
	).CombinedOutput(); err != nil {
		s.failUpdate(fmt.Errorf("解压失败: %s, %w", string(out), err))
		return
	}

	// 5. 同步运行时数据
	for _, dir := range []string{"config", "logs", "uploads"} {
		src := filepath.Join(installDir, dir)
		dst := filepath.Join(stagingDir, dir)
		if _, err := os.Stat(src); err == nil {
			os.RemoveAll(dst)
			_ = exec.Command("cp", "-a", src, dst).Run()
		}
	}

	// 6. 替换文件（Docker volume mount 不支持 rename 目录本身，改原地 cp）
	//    cp -a 批量覆盖，运行中的二进制 xt-pms 会跳过，后面单独处理
	if out, err := exec.Command(
		"cp", "-a", stagingDir+"/.", installDir+"/",
	).CombinedOutput(); err != nil {
		// 二进制 Text file busy 会导致 cp 返回非零，其他文件已成功复制
		logger.Warn("[supervisor-update] 批量覆盖完成（二进制可能因锁跳过）",
			zap.String("detail", strings.TrimSpace(string(out))))
	}

	// 单独替换二进制（先删后 cp，Linux 允许删除运行中的二进制，fd 保持）
	binPath := filepath.Join(installDir, "xt-pms")
	os.Remove(binPath)
	if out, err := exec.Command("cp", filepath.Join(stagingDir, "xt-pms"), binPath).CombinedOutput(); err != nil {
		s.failUpdate(fmt.Errorf("替换二进制失败: %s, %w", string(out), err))
		return
	}
	os.Chmod(binPath, 0755)

	os.RemoveAll(stagingDir)
	stagingDir = "" // 防 defer 误删

	// 7. 更新 VERSION
	_ = os.WriteFile(filepath.Join(installDir, "VERSION"), []byte(m.ToVersion), 0644)

	// 9. 清理临时文件
	_ = os.Remove(updateManifest)
	_ = os.RemoveAll(updateTmpDir)

	// 10. 原地变身：替换当前进程为新的 xt-pms（PID 不变，1Panel 无感知）
	logger.Info("[supervisor-update] 更新完成，正在 exec 新版本")
	exe := filepath.Join(installDir, "xt-pms")
	if err := syscall.Exec(exe, []string{exe}, os.Environ()); err != nil {
		logger.Error("[supervisor-update] syscall.Exec 失败", zap.Error(err))
		os.Exit(1)
	}
}

// failUpdate 更新失败时清理并退出
func (s *SysUpdateService) failUpdate(err error) {
	logger.Error("[supervisor-update] 更新失败", zap.Error(err))
	_ = os.Remove(updateManifest)
	_ = os.RemoveAll(updateTmpDir)
}

// ========== update-helper 子进程逻辑 ==========

// RunUpdateHelper 由主进程 spawn 的子进程执行，负责停止服务、替换文件
func RunUpdateHelper() error {
	data, err := os.ReadFile(updateManifest)
	if err != nil {
		return fmt.Errorf("读取更新清单失败: %w", err)
	}

	var manifest dto.UpdateManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("解析更新清单失败: %w", err)
	}

	switch manifest.Action {
	case "update":
		return applyUpdate(&manifest)
	default:
		return fmt.Errorf("未知操作: %s", manifest.Action)
	}
}

func applyUpdate(m *dto.UpdateManifest) error {
	switch m.UpdateMode {
	case "supervisor", "docker":
		// Docker 容器与 supervisor 模式机制相同：
		// 进程退出后由外层（docker / supervisord）自动重新拉起
		return applyUpdateSupervisor(m)
	default:
		return applyUpdateSystemd(m)
	}
}

func applyUpdateSystemd(m *dto.UpdateManifest) error {
	installDir := m.InstallDir
	if installDir == "" {
		installDir = defaultInstallDir
	}
	svcName := m.ServiceName
	if svcName == "" {
		svcName = defaultService
	}

	fmt.Println("[update-helper] 等待父进程完成响应...")
	time.Sleep(3 * time.Second)

	// 1. 检查更新包
	if _, err := os.Stat(m.PackagePath); os.IsNotExist(err) {
		return fmt.Errorf("更新包不存在: %s", m.PackagePath)
	}

	// 2. SHA256 校验
	if m.SHA256 != "" {
		f, err := os.Open(m.PackagePath)
		if err != nil {
			return fmt.Errorf("打开更新包失败: %w", err)
		}
		h := sha256.New()
		io.Copy(h, f)
		f.Close()
		actual := fmt.Sprintf("%x", h.Sum(nil))
		if actual != m.SHA256 {
			return fmt.Errorf("SHA256 校验失败\n期望: %s\n实际: %s", m.SHA256, actual)
		}
		fmt.Println("[update-helper] SHA256 校验通过")
	}

	// 3. 停止服务
	fmt.Printf("[update-helper] 正在停止服务 %s...\n", svcName)
	if out, err := exec.Command("systemctl", "stop", svcName).CombinedOutput(); err != nil {
		return fmt.Errorf("停止服务失败: %s, %w", string(out), err)
	}
	time.Sleep(2 * time.Second)

	// 4. 备份
	backupDir := m.BackupDir
	if backupDir == "" {
		backupDir = installDir + ".bak." + time.Now().Format("20060102_150405")
	}
	fmt.Printf("[update-helper] 正在备份到 %s...\n", backupDir)
	if err := exec.Command("cp", "-a", installDir, backupDir).Run(); err != nil {
		// 备份失败不阻塞，继续更新
		fmt.Printf("[update-helper] 备份失败（继续更新）: %v\n", err)
	}

	// 5. 解压替换
	fmt.Printf("[update-helper] 正在解压更新包 %s...\n", m.PackagePath)
	// 使用 --strip-components=1 去掉顶层目录
	if out, err := exec.Command("tar", "-xzf", m.PackagePath, "-C", installDir, "--strip-components=1", "--overwrite").CombinedOutput(); err != nil {
		return fmt.Errorf("解压更新包失败: %s, %w", string(out), err)
	}

	// 6. 更新 VERSION 文件
	versionFile := filepath.Join(installDir, "VERSION")
	_ = os.WriteFile(versionFile, []byte(m.ToVersion), 0644)

	// 7. 启动服务
	fmt.Printf("[update-helper] 正在启动服务 %s...\n", svcName)
	if out, err := exec.Command("systemctl", "start", svcName).CombinedOutput(); err != nil {
		return fmt.Errorf("启动服务失败: %s, %w", string(out), err)
	}

	// 8. 清理临时文件
	_ = os.Remove(updateManifest)
	_ = os.RemoveAll(updateTmpDir)

	fmt.Printf("[update-helper] 更新完成 v%s\n", m.ToVersion)
	return nil
}

// applyUpdateSupervisor supervisor 模式：解压到 staging → 同步运行时 → 杀进程 → 原子替换
// 由外层 supervisor（1Panel Go 运行时 / supervisord / docker）自动重新拉起
func applyUpdateSupervisor(m *dto.UpdateManifest) error {
	installDir := m.InstallDir
	if installDir == "" {
		installDir = defaultInstallDir
	}

	fmt.Println("[update-helper] 等待父进程完成响应...")
	time.Sleep(3 * time.Second)

	// 1. 检查更新包
	if _, err := os.Stat(m.PackagePath); os.IsNotExist(err) {
		return fmt.Errorf("更新包不存在: %s", m.PackagePath)
	}

	// 2. SHA256 校验
	if m.SHA256 != "" {
		f, err := os.Open(m.PackagePath)
		if err != nil {
			return fmt.Errorf("打开更新包失败: %w", err)
		}
		h := sha256.New()
		io.Copy(h, f)
		f.Close()
		actual := fmt.Sprintf("%x", h.Sum(nil))
		if actual != m.SHA256 {
			return fmt.Errorf("SHA256 校验失败\n期望: %s\n实际: %s", m.SHA256, actual)
		}
		fmt.Println("[update-helper] SHA256 校验通过")
	}

	// 3. 全量备份（先跑，保证出问题能回滚）
	backupDir := m.BackupDir
	if backupDir == "" {
		backupDir = installDir + ".bak." + time.Now().Format("20060102_150405")
	}
	fmt.Printf("[update-helper] 正在备份到 %s...\n", backupDir)
	if err := exec.Command("cp", "-a", installDir, backupDir).Run(); err != nil {
		fmt.Printf("[update-helper] 备份失败（继续更新）: %v\n", err)
	}

	// 4. 解压到 staging 目录（不影响运行中的进程，无 Text file busy 问题）
	stagingDir := installDir + ".staging"
	os.RemoveAll(stagingDir)
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		return fmt.Errorf("创建 staging 目录失败: %w", err)
	}
	defer os.RemoveAll(stagingDir) // 异常退出时兜底清理

	fmt.Printf("[update-helper] 正在解压到 %s...\n", stagingDir)
	if out, err := exec.Command(
		"tar", "-xzf", m.PackagePath, "-C", stagingDir,
		"--strip-components=1", "--no-xattrs",
	).CombinedOutput(); err != nil {
		return fmt.Errorf("解压更新包失败: %s, %w", string(out), err)
	}

	// 5. 同步运行时数据（config/logs/uploads 从当前目录复制到 staging）
	for _, dir := range []string{"config", "logs", "uploads"} {
		src := filepath.Join(installDir, dir)
		dst := filepath.Join(stagingDir, dir)
		if _, err := os.Stat(src); err == nil {
			os.RemoveAll(dst)
			if err := exec.Command("cp", "-a", src, dst).Run(); err != nil {
				fmt.Printf("[update-helper] 同步 %s 失败（继续）: %v\n", dir, err)
			}
		}
	}

	// 6. 先原子替换目录（主进程还跑着，但 Linux 允许 rename 运行目录）
	//    旧进程通过 fd 继续跑在 .old 下，1Panel 重启时会指向新目录
	//    顺序至关重要：先 rename 后 kill，消除 1Panel 快速重启的竞态窗口
	oldDir := installDir + ".old"
	os.RemoveAll(oldDir)
	if err := os.Rename(installDir, oldDir); err != nil {
		return fmt.Errorf("移除旧目录失败: %w", err)
	}
	if err := os.Rename(stagingDir, installDir); err != nil {
		os.Rename(oldDir, installDir) // 失败时回滚
		return fmt.Errorf("激活新版本失败: %w", err)
	}
	stagingDir = "" // 已经 rename 走了，防止 defer 误删
	fmt.Println("[update-helper] 目录替换完成，新文件已就位")

	// 7. 更新 VERSION 文件
	_ = os.WriteFile(filepath.Join(installDir, "VERSION"), []byte(m.ToVersion), 0644)

	// 8. 停止主进程（此时新目录已在位，1Panel 重启就能立刻跑新版本）
	if m.ParentPID > 0 {
		fmt.Printf("[update-helper] 正在停止主进程 (PID=%d)...\n", m.ParentPID)
		if err := syscall.Kill(m.ParentPID, syscall.SIGTERM); err != nil {
			fmt.Printf("[update-helper] SIGTERM 失败: %v，尝试 SIGKILL...\n", err)
			_ = syscall.Kill(m.ParentPID, syscall.SIGKILL)
		}
		for i := 0; i < 20; i++ {
			if err := syscall.Kill(m.ParentPID, 0); err != nil {
				fmt.Println("[update-helper] 主进程已退出")
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

	// 9. 异步清理旧目录和临时文件
	go func(old string) {
		time.Sleep(10 * time.Second)
		os.RemoveAll(old)
	}(oldDir)
	_ = os.Remove(updateManifest)
	_ = os.RemoveAll(updateTmpDir)

	fmt.Printf("[update-helper] 更新完成 v%s，1Panel 将自动重启新版本\n", m.ToVersion)
	return nil
}

// IsSystemdAvailable 检查 systemd 是否可用
func IsSystemdAvailable() bool {
	return DetectUpdateMode() == "systemd"
}

// DetectUpdateMode 检测更新模式
// - systemd: 由 systemd 管理服务生命周期，更新时执行 systemctl stop/start
// - docker:  容器化部署（Docker / Docker Compose），由 docker 守护进程管理生命周期
// - supervisor: 由 1Panel Go 运行时 / supervisord 等管理，更新时替换二进制后 os.Exit(0) 让外层自动重启
// - manual: 手动运行（./xt-pms），不支持自动更新
func DetectUpdateMode() string {
	// 1. 配置文件优先级最高
	if config.GlobalConfig != nil && config.GlobalConfig.App.UpdateMode != "" && config.GlobalConfig.App.UpdateMode != "auto" {
		return config.GlobalConfig.App.UpdateMode
	}

	// 2. 检测是否在 Docker 容器内运行
	//    识别方式：/.dockerenv 文件（最稳定）或 /proc/1/cgroup 中包含 docker/containerd
	if IsRunningInDocker() {
		return "docker"
	}

	// 3. 自动检测：先尝试 systemd
	svc := getServiceName()
	if svc != "" {
		if exec.Command("systemctl", "is-active", "--quiet", svc).Run() == nil {
			return "systemd"
		}
	}

	// 4. 不是 systemd，但二进制文件存在 → supervisor 模式
	//    （1Panel Go 运行时、supervisord 等由外部进程管理生命周期的场景）
	//    只要二进制存在，就假设外部有 supervisor 会在进程退出后重新拉起
	exe, err := os.Executable()
	if err == nil {
		if _, err := os.Stat(exe); err == nil {
			return "supervisor"
		}
	}

	return "manual"
}

// IsRunningInDocker 检测当前进程是否运行在 Docker 容器内
func IsRunningInDocker() bool {
	// 方式 1：/.dockerenv 是 Docker 创建的标志性文件
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// 方式 2：环境变量（docker compose / k8s 通常会设置）
	if os.Getenv("DOCKER_CONTAINER") != "" || os.Getenv("RUNNING_IN_DOCKER") != "" {
		return true
	}

	// 方式 3：解析 /proc/1/cgroup，查找 docker / containerd 标识
	if data, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		content := strings.ToLower(string(data))
		if strings.Contains(content, "docker") ||
			strings.Contains(content, "containerd") ||
			strings.Contains(content, "kubepods") {
			return true
		}
	}

	return false
}

// GetInstallInfo 获取安装信息（安装目录、服务名、更新模式）
// Docker 部署的 canAutoUpdate 也为 false，因为 Docker 应通过镜像升级而非容器内替换文件。
func (s *SysUpdateService) GetInstallInfo() (installDir, serviceName, updateMode string, canAutoUpdate bool) {
	svc := getServiceName()
	dir, err := s.DetectInstallDir()
	if err != nil {
		dir = defaultInstallDir
	}
	mode := DetectUpdateMode()
	// Docker 和 manual 模式均不支持容器内在线更新
	canAutoUpdate = mode != "manual" && mode != "docker"
	return dir, svc, mode, canAutoUpdate
}

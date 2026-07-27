package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/service"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type UpdateController struct {
	updateService *service.SysUpdateService
}

func NewUpdateController() *UpdateController {
	return &UpdateController{
		updateService: service.NewSysUpdateService(),
	}
}

// CheckUpdate 检查是否有新版本
func (uc *UpdateController) CheckUpdate(c *gin.Context) {
	currentVersion := c.Query("current")
	if currentVersion == "" {
		response.Error(c, response.BAD_REQUEST, "缺少 current 参数（当前版本号）")
		return
	}

	// 可选：前端传入自定义更新源 URL
	overrideURL := c.Query("url")

	result, err := uc.updateService.CheckUpdate(currentVersion, overrideURL)
	if err != nil {
		logger.Warn("检查更新失败", zap.Error(err))
		response.Error(c, response.ERROR, "检查更新失败: "+err.Error())
		return
	}

	response.Success(c, result)
}

// DownloadUpdate 下载更新包（SSE 推送进度）
func (uc *UpdateController) DownloadUpdate(c *gin.Context) {
	var req struct {
		DownloadURL string `json:"download_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.DownloadURL == "" {
		response.Error(c, response.BAD_REQUEST, "缺少 download_url 参数")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Error(c, response.ERROR, "不支持 SSE")
		return
	}

	// 进度回调
	progressFn := func(percent int, downloaded, total int64) {
		msg := dto.DownloadProgress{
			Percent:    percent,
			Downloaded: downloaded,
			Total:      total,
			Status:     "downloading",
		}
		data, _ := json.Marshal(msg)
		_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		flusher.Flush()
	}

	filePath, sha256URL, err := uc.updateService.DownloadUpdate(req.DownloadURL, progressFn)
	if err != nil {
		errMsg := dto.DownloadProgress{Status: "failed", Message: err.Error()}
		data, _ := json.Marshal(errMsg)
		_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", data)
		flusher.Flush()
		return
	}

	sha256, _ := uc.updateService.ComputeSHA256(filePath)

	// 下载完成
	done := dto.DownloadProgress{
		Percent:    100,
		Status:     "completed",
		Message:    filePath,
	}
	// 把 sha256 嵌入 message（用分隔符拼接文件路径和 sha256）
	if sha256 != "" {
		done.Message = filePath + "|sha256:" + sha256
	}
	// 如果需要从远程获取 sha256 文件
	if sha256 == "" && sha256URL != "" {
		done.Message = filePath + "|sha256_url:" + sha256URL
	}

	data, _ := json.Marshal(done)
	_, _ = fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	flusher.Flush()
}

// ApplyUpdate 执行更新（生成清单 → spawn helper 子进程）
func (uc *UpdateController) ApplyUpdate(c *gin.Context) {
	var req struct {
		FromVersion string `json:"from_version"`
		ToVersion   string `json:"to_version"`
		PackagePath string `json:"package_path"`
		SHA256      string `json:"sha256"`
		Changelog   string `json:"changelog"`
		FileSize    int64  `json:"file_size"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.PackagePath == "" {
		response.Error(c, response.BAD_REQUEST, "参数不完整")
		return
	}

	// 检测部署环境
	installDir, svcName, updateMode, canAutoUpdate := uc.updateService.GetInstallInfo()

	if !canAutoUpdate {
		response.Error(c, response.ERROR, "当前部署环境不支持在线更新，请手动升级")
		return
	}

	backupDir := installDir + ".bak." + req.ToVersion

	exe, _ := os.Executable()
	manifest := &dto.UpdateManifest{
		Action:      "update",
		FromVersion: req.FromVersion,
		ToVersion:   req.ToVersion,
		PackagePath: req.PackagePath,
		SHA256:      req.SHA256,
		InstallDir:  installDir,
		ServiceName: svcName,
		ParentPID:   os.Getpid(),
		BackupDir:   backupDir,
		UpdateMode:  updateMode,
	}
	if err := uc.updateService.WriteManifest(manifest); err != nil {
		response.Error(c, response.ERROR, "写入更新清单失败")
		return
	}

	// 根据部署环境选择更新策略
	if updateMode == "supervisor" || updateMode == "docker" {
		// supervisor / docker 环境：主进程自行替换 + syscall.Exec
		// PID 不变，外层（supervisord / docker）无感知，无需依赖外部重启
		go uc.updateService.RunSupervisorUpdate(manifest)
	} else {
		// systemd 环境：spawn helper 子进程执行 systemctl stop → 替换 → systemctl start
		go uc.spawnHelper(exe, installDir)
	}

	response.Success(c, gin.H{
		"message": "更新已启动，服务即将自动重启，请稍后刷新页面",
	})
}

// Restart 重启服务（主进程直接 Exit，由 supervisor 自动拉起，通杀所有部署环境）
func (uc *UpdateController) Restart(c *gin.Context) {
	_, _, _, canAutoUpdate := uc.updateService.GetInstallInfo()

	if !canAutoUpdate {
		response.Error(c, response.ERROR, "当前部署环境不支持在线重启")
		return
	}

	response.Success(c, gin.H{"message": "服务即将重启，请稍后刷新页面"})

	go func() {
		time.Sleep(500 * time.Millisecond) // 等 HTTP 响应发完
		os.Exit(0)
	}()
}

// GetInstallInfo 获取当前部署信息
func (uc *UpdateController) GetInstallInfo(c *gin.Context) {
	installDir, svcName, updateMode, canAutoUpdate := uc.updateService.GetInstallInfo()
	response.Success(c, gin.H{
		"install_dir":     installDir,
		"service_name":    svcName,
		"update_mode":     updateMode,
		"can_auto_update": canAutoUpdate,
	})
}

// spawnHelper 启动 helper 子进程（后台独立运行，不依赖父进程）
func (uc *UpdateController) spawnHelper(exePath string, installDir string) {
	cmd := exec.Command(exePath, "update-helper")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	// 日志输出到文件（因为父进程退出后 stdout/stderr 不可用）
	logFile, err := os.OpenFile("/tmp/xt-pms-update-helper.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err == nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	// 工作目录使用检测到的安装目录
	if installDir == "" {
		installDir = "/opt/xt-pms"
	}
	cmd.Dir = installDir

	if err := cmd.Start(); err != nil {
		logger.Error("启动 helper 子进程失败", zap.Error(err))
		return
	}

	// 释放子进程，不等待
	go func() {
		if logFile != nil {
			defer logFile.Close()
		}
		if err := cmd.Wait(); err != nil {
			logger.Error("helper 子进程异常退出", zap.Error(err))
		}
	}()

	logger.Info("helper 子进程已启动，即将重启服务")
}

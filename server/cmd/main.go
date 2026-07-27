package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // 导入 pprof 用于性能分析
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/monitor"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/installer"
	"github.com/KamyangTech/xt-pms/server/internal/job"
	"github.com/KamyangTech/xt-pms/server/internal/router"
	"github.com/KamyangTech/xt-pms/server/internal/service"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

// 编译时注入的版本信息（通过 ldflags -X）
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
	DevBuild  = "false" // true=本地开发编译，跳过安装向导
)

// 当前要求的数据库 Schema 版本（与 init_schema.sql 版本对应）
const CurrentSchemaVersion = 1

// InstalledInfo 记录在 config/.installed 中的部署信息
type InstalledInfo struct {
	Version       string `json:"version"`        // 编译版本号
	BuildTime     string `json:"build_time"`     // 编译时间 (UTC)
	GitCommit     string `json:"git_commit"`     // Git 提交哈希
	GoVersion     string `json:"go_version"`     // Go 编译版本
	SchemaVersion int    `json:"schema_version"` // 数据库 Schema 版本
	InstalledAt   string `json:"installed_at"`   // 首次安装时间（安装后不再变）
}

const installedFile = "config/.installed"

// isDevMode 判断是否为开发模式（go run 或本地 make build）
func isDevMode() bool {
	if DevBuild == "true" {
		return true
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return strings.Contains(exe, "go-build")
}

// isFirstInstall 判断当前是否为首次部署（config/.installed 不存在）
func isFirstInstall() bool {
	_, err := os.Stat(installedFile)
	return os.IsNotExist(err)
}

func loadInstalledInfo() (*InstalledInfo, error) {
	data, err := os.ReadFile(installedFile)
	if err != nil {
		return nil, err
	}
	var info InstalledInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// saveInstalledInfo 写入安装信息（首次安装时用）
func saveInstalledInfo(info *InstalledInfo) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(installedFile, data, 0644)
}

// refreshInstalledInfo 更新版本相关字段（每次成功启动时调用）
func refreshInstalledInfo() {
	if isDevMode() {
		return // 开发模式不写入 .installed
	}
	info, err := loadInstalledInfo()
	if err != nil {
		return
	}
	// 仅更新版本相关的编译时信息，installed_at 保持不变
	info.Version = Version
	info.BuildTime = BuildTime
	info.GitCommit = GitCommit
	info.GoVersion = runtime.Version()
	_ = saveInstalledInfo(info)
}

// Docker 自动创建管理员子命令，完全不影响裸机部署

// runSetupAdmin 从环境变量读取管理员信息并创建账号
func runSetupAdmin() {
	username := os.Getenv("PMS_ADMIN_USERNAME")
	password := os.Getenv("PMS_ADMIN_PASSWORD")
	realName := os.Getenv("PMS_ADMIN_REAL_NAME")
	email := os.Getenv("PMS_ADMIN_EMAIL")
	phone := os.Getenv("PMS_ADMIN_PHONE")

	if username == "" || password == "" {
		fmt.Println("Usage: PMS_ADMIN_USERNAME=xxx PMS_ADMIN_PASSWORD=xxx ./pms-server setup-admin")
		fmt.Println("Optional: PMS_ADMIN_REAL_NAME, PMS_ADMIN_EMAIL, PMS_ADMIN_PHONE")
		os.Exit(1)
	}

	if len(password) < 6 {
		fmt.Println("密码至少 6 位")
		os.Exit(1)
	}

	if err := config.Init(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	if err := dao.InitDB(config.GlobalConfig); err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}

	var count int64
	if err := dao.DB.Table("sys_user").Where("username = ?", username).Count(&count).Error; err != nil {
		fmt.Printf("查询用户失败: %v\n", err)
		os.Exit(1)
	}
	if count > 0 {
		fmt.Printf("管理员 %s 已存在，跳过创建\n", username)
		os.Exit(0)
	}

	pepper := config.GlobalConfig.Security.Pepper
	pepperedPassword := password + pepper
	hash, err := bcrypt.GenerateFromPassword([]byte(pepperedPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("生成密码哈希失败: %v\n", err)
		os.Exit(1)
	}

	if err := dao.DB.Exec(
		"INSERT INTO sys_user (username, password, real_name, email, phone, status) VALUES (?, ?, ?, ?, ?, 1)",
		username, string(hash), realName, email, phone,
	).Error; err != nil {
		fmt.Printf("创建管理员失败: %v\n", err)
		os.Exit(1)
	}

	if err := dao.DB.Exec(
		"INSERT INTO sys_user_role (user_id, role_id) SELECT u.id, r.id FROM sys_user u, sys_role r WHERE u.username = ? AND r.role_name = '超级管理员'",
		username,
	).Error; err != nil {
		fmt.Printf("分配角色失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("管理员 %s 创建成功\n", username)
	os.Exit(0)
}

func runResetPassword() {
	fmt.Println()
	fmt.Println("🔐 重置密码")
	fmt.Println("----------------")

	if err := config.Init(); err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(); err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	if err := dao.InitDB(config.GlobalConfig); err != nil {
		fmt.Printf("连接数据库失败: %v\n", err)
		os.Exit(1)
	}

	admin, err := dao.FindSuperAdmin()
	if err != nil || admin == "" {
		admin = "admin"
	}

	username := promptString(fmt.Sprintf("用户名 [%s]", admin))
	if username == "" {
		username = admin
	}

	password := promptPassword("新密码")
	if password == "" {
		fmt.Println("密码不能为空")
		os.Exit(1)
	}

	password2 := promptPassword("确认密码")
	if password != password2 {
		fmt.Println("两次密码不一致")
		os.Exit(1)
	}
	if len(password) < 6 {
		fmt.Println("密码至少 6 位")
		os.Exit(1)
	}

	pepper := config.GlobalConfig.Security.Pepper
	pepperedPassword := password + pepper
	hash, err := bcrypt.GenerateFromPassword([]byte(pepperedPassword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("生成密码哈希失败: %v\n", err)
		os.Exit(1)
	}

	result := dao.DB.Table("sys_user").Where("username = ?", username).Update("password", string(hash))
	if result.Error != nil {
		fmt.Printf("更新密码失败: %v\n", result.Error)
		os.Exit(1)
	}
	if result.RowsAffected == 0 {
		fmt.Printf("用户 %s 不存在\n", username)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("✅ 用户 %s 密码已重置\n", username)
	os.Exit(0)
}

func runUpdateHelper() {
	if err := service.RunUpdateHelper(); err != nil {
		fmt.Fprintf(os.Stderr, "更新失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("更新操作完成")
	os.Exit(0)
}

func promptString(prompt string) string {
	fmt.Printf("%s: ", prompt)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

// promptPassword 显示提示并读取密码（终端不回显）
func promptPassword(prompt string) string {
	fmt.Printf("%s: ", prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		// 降级：如果终端不支持，用明文读取
		var input string
		fmt.Scanln(&input)
		return strings.TrimSpace(input)
	}
	fmt.Println()
	return string(bytePassword)
}

// safeStaticDir 返回安全的静态文件目录绝对路径，防止路径遍历攻击
func safeStaticDir(relativePath string) string {
	abs, err := filepath.Abs(relativePath)
	if err != nil {
		logger.Warn("解析静态文件目录失败，使用相对路径", zap.String("path", relativePath), zap.Error(err))
		return relativePath
	}
	return abs
}

func main() {
	// === 子命令：Docker 自动创建管理员 ===
	if len(os.Args) >= 2 && os.Args[1] == "setup-admin" {
		runSetupAdmin()
		return
	}

	// === 子命令：重置密码 ===
	if len(os.Args) >= 2 && os.Args[1] == "reset-password" {
		runResetPassword()
		return
	}

	// === 子命令：在线更新辅助程序 ===
	if len(os.Args) >= 2 && os.Args[1] == "update-helper" {
		runUpdateHelper()
		return
	}

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Fatalf("加载时区失败：%v", err)
	}
	time.Local = loc

	// 确保 config 目录存在（.installed 文件需要写入）
	if err := os.MkdirAll("config", 0755); err != nil {
		log.Fatalf("创建 config 目录失败：%v", err)
	}

	// === 首次部署检测（仅二进制启动时生效，go run 跳过） ===
	if isFirstInstall() {
		if isDevMode() {
			fmt.Println("[dev] 开发模式，跳过安装检测")
		} else if err := config.Init(); err == nil {
			// 配置可加载 → 老版本升级，自动补齐 .installed
			fmt.Println("检测到旧版本部署（无 .installed 标记），自动补齐...")
			saveInstalledInfo(&InstalledInfo{
				Version:       Version,
				BuildTime:     BuildTime,
				GitCommit:     GitCommit,
				GoVersion:     runtime.Version(),
				SchemaVersion: CurrentSchemaVersion,
				InstalledAt:   time.Now().Format(time.RFC3339),
			})
		} else {
			// 生产构建 + 配置不可加载 → 真正首次部署，启动安装向导
			fmt.Println("===================================")
			fmt.Printf("  Xt-PMS %s (commit: %s)\n", Version, GitCommit)
			fmt.Printf("  构建时间: %s\n", BuildTime)
			fmt.Println("===================================")
			installer.Run() // 阻塞直到安装完成
			// 安装成功，写入 .installed 标记，继续正常启动
			saveInstalledInfo(&InstalledInfo{
				Version:       Version,
				BuildTime:     BuildTime,
				GitCommit:     GitCommit,
				GoVersion:     runtime.Version(),
				SchemaVersion: CurrentSchemaVersion,
				InstalledAt:   time.Now().Format(time.RFC3339),
			})
		}
	}

	if err := config.Init(); err != nil {
		log.Fatalf("加载配置失败：%v", err)
	}

	if err := logger.Init(); err != nil {
		log.Fatalf("初始化日志系统失败：%v", err)
	}

	// 确保程序退出前刷新日志缓冲区
	defer logger.Sync()

	// 确保必要的目录存在（自动创建，已存在则跳过）
	ensureDirs := []string{
		"./uploads",
		"./logs",
		"./system-assets",
	}
	for _, dir := range ensureDirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			logger.Warn("创建目录失败", zap.String("dir", dir), zap.Error(err))
		}
	}

	if err := dao.InitDB(config.GlobalConfig); err != nil {
		logger.Error("初始化数据库失败", zap.Error(err))
		os.Exit(1)
	}

	// 启动后台性能监控（goroutine 泄漏 + 连接池使用率自动告警）
	monitor.StartWatcher()

	// 初始化安全配置服务（加载动态安全配置到内存缓存）
	service.InitSecurityConfigService()

	// 初始化更新服务（确保 sys_update_log 表存在）
	if err := service.InitUpdateService(); err != nil {
		logger.Warn("初始化更新服务失败", zap.Error(err))
	}

	// 后台异步生成缺失的缩略图（不阻塞启动）
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("[缩略图] 启动扫描 panic 已恢复", zap.Any("recover", r))
			}
		}()
		var urls []string
		if err := dao.DB.Table("production_order_item").
			Where("product_image_url IS NOT NULL AND product_image_url != ''").
			Distinct("product_image_url").
			Pluck("product_image_url", &urls).Error; err != nil {
			logger.Error("[缩略图] 查询产品图片URL失败", zap.Error(err))
			return
		}
		if len(urls) == 0 {
			return
		}
		basePath := config.GlobalConfig.Upload.BasePath
		utils.GenerateMissingThumbnails(basePath, urls, func(msg string, fields ...interface{}) {
			logger.Info("[缩略图] "+msg, zap.Any("fields", fields))
		})
	}()

	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	if config.GlobalConfig.Server.Port <= 0 {
		addr = ":9898"
	}

	// 创建 Gin 引擎（使用 gin.New 替代 gin.Default，手动注册中间件）
	r := gin.New()
	r.Use(middleware.SecurityHeaders())     // 安全响应头（X-Frame-Options、CSP 等）
	r.Use(middleware.BodyLimit())           // 请求体大小限制
	r.Use(monitor.MetricsMiddleware())      // 运行时性能指标收集（必须放在 Logger 之前）
	r.Use(middleware.Logger())              // HTTP 请求日志
	r.Use(middleware.CORS())                // 跨域
	r.Use(middleware.XSSSanitize())         // XSS 输入清理
	r.Use(gin.Recovery())                   // 异常恢复

	// 限制 multipart 表单（文件上传）内存占用为 8MB（单次请求）
	r.MaxMultipartMemory = 8 << 20

	// 配置静态文件服务（使用绝对路径防止路径遍历攻击）
	uploadsDir := safeStaticDir("./uploads")
	r.StaticFS("/uploads", http.Dir(uploadsDir))

	// 系统配置资源（Favicon、登录背景、图标等），与用户业务文件隔离
	systemAssetsDir := safeStaticDir("./system-assets")
	r.StaticFS("/system-assets", http.Dir(systemAssetsDir))

	// 浏览器兜底：请求 /favicon.ico 时返回默认图标
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./system-assets/logo.svg")
	})

	router.SetupRouter(r)

	// 生产构建时启用嵌入式前端 SPA 服务（-tags release）
	serveEmbeddedSPA(r)

	cronJob := cron.New(cron.WithLocation(time.Local))

	// 每天凌晨 3:00 自动清理孤文件
	if _, err := cronJob.AddFunc("0 3 * * *", func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("[文件清理] 任务 panic 已恢复", zap.Any("recover", r))
			}
		}()
		job.CleanupOrphanFilesWithRetention(job.GetRetentionDays())
	}); err != nil {
		logger.Warn("注册文件清理定时任务失败", zap.Error(err))
	} else {
		logger.Info("文件清理定时任务已启动",
			zap.Int("retention_days", job.GetRetentionDays()),
			zap.String("schedule", "0 3 * * *"),
		)
	}

	// 每天凌晨 3:30 清理半年之前的操作日志
	if _, err := cronJob.AddFunc("30 3 * * *", func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("[操作日志清理] 任务 panic 已恢复", zap.Any("recover", r))
			}
		}()
		job.CleanupOperLogs(job.CleanupOperLogRetentionDays)
	}); err != nil {
		logger.Warn("注册操作日志清理定时任务失败", zap.Error(err))
	} else {
		logger.Info("操作日志清理定时任务已启动",
			zap.Int("retention_days", job.CleanupOperLogRetentionDays),
			zap.String("schedule", "30 3 * * *"),
		)
	}

	if len(cronJob.Entries()) > 0 {
		cronJob.Start()
		defer cronJob.Stop()
	}

	// pprof：仅在非 production 模式启用（避免生产环境暴露调试端点）
	if strings.EqualFold(strings.TrimSpace(config.GlobalConfig.App.Mode), "development") {
		go func() {
			logger.Info("pprof 启动于 http://localhost:6060/debug/pprof")
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				logger.Warn("pprof 启动失败", zap.Error(err))
			}
		}()
	}

	logger.Info("服务器启动",
		zap.String("version", Version),
		zap.String("commit", GitCommit),
		zap.String("addr", addr),
	)

	// 刷新 .installed 中的版本信息（首次安装后已在 main 开头写入，升级时更新版本号）
	refreshInstalledInfo()

	// 在独立 goroutine 中启动 HTTP 服务
	srv := &http.Server{
		Addr:           addr,
		Handler:        r,
		MaxHeaderBytes: 1 << 20, // 请求头最大 1MB，防止超大 header 攻击
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP 服务异常退出", zap.Error(err))
			os.Exit(1)
		}
	}()

	// 等待中断信号（SIGINT/SIGTERM），优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("收到关闭信号，正在优雅关闭服务器...")

	// 1. 关闭 WebSocket 管理器（断开所有客户端连接，释放 goroutine）
	controller.WS.Shutdown()

	// 2. 关闭定时任务
	cronJob.Stop()

	// 3. 关闭 HTTP 服务（给正在处理的请求最多 5 秒完成时间）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Warn("HTTP 服务关闭超时", zap.Error(err))
	}

	logger.Info("服务器已完全关闭")
}

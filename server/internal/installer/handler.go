package installer

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================
// 安装向导 HTTP 路由注册
// ============================================

// RegisterRoutes 在 Gin 引擎上注册安装向导路由
func RegisterRoutes(r *gin.Engine) {
	// 安装向导首页（SPA 单页）
	r.GET("/", serveWizardPage)

	// API 端点
	api := r.Group("/api/setup")
	{
		api.POST("/test-db", handleTestDB)
		api.POST("/check-db", handleCheckDB)
		api.POST("/install", handleInstall)
		api.GET("/generate-pepper", handleGeneratePepper)
		api.GET("/generate-jwt", handleGenerateJWT)
	}
}

// ============================================
// 页面处理
// ============================================

func serveWizardPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, wizardHTML)
}

// ============================================
// API 处理
// ============================================

type testDBRequest struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password"`
	DBName   string `json:"db_name" binding:"required"`
}

func handleTestDB(c *gin.Context) {
	var req testDBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数不完整"})
		return
	}

	if err := TestDBConnection(req.Host, req.Port, req.User, req.Password, req.DBName); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "数据库连接成功！"})
}

func handleCheckDB(c *gin.Context) {
	var req testDBRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数不完整"})
		return
	}

	empty, err := CheckDBIsEmpty(req.Host, req.Port, req.User, req.Password, req.DBName)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
		return
	}

	if empty {
		c.JSON(http.StatusOK, gin.H{"success": true, "empty": true, "message": "数据库为空，可以安装"})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": true, "empty": false, "message": "⚠️ 数据库已有表，安装将覆盖现有数据！"})
	}
}

func handleGeneratePepper(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"value": GeneratePepper()})
}

func handleGenerateJWT(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"value": GenerateJWTSecret()})
}

func handleInstall(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "读取请求失败"})
		return
	}

	var cfg InstallConfig
	if err := json.Unmarshal(body, &cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "参数格式错误"})
		return
	}

	// 设置默认值
	if cfg.ServerPort <= 0 {
		cfg.ServerPort = 9898
	}
	if cfg.AppMode == "" {
		cfg.AppMode = "production"
	}
	if cfg.Pepper == "" {
		cfg.Pepper = GeneratePepper()
	}
	if cfg.JWTSecret == "" {
		cfg.JWTSecret = GenerateJWTSecret()
	}

	result := RunInstall(&cfg)
	if !result.Success {
		c.JSON(http.StatusInternalServerError, result)
		return
	}

	c.JSON(http.StatusOK, result)

	// 安装成功 → 关闭向导，主进程继续启动正常服务
	go func() {
		time.Sleep(500 * time.Millisecond)
		Shutdown()
	}()
}

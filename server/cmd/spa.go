//go:build release

package main

import (
	"embed"
	"mime"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/KamyangTech/xt-pms/server/pkg/logger"
)

//go:embed web-dist
var webDist embed.FS

// serveEmbeddedSPA 从嵌入式文件系统提供前端 SPA 静态文件。
// 不使用 Gin FileFromFS 或 http.FileServer（与 embed.FS 存在 301 兼容性问题），
// 直接读取文件内容通过 c.Data 返回，行为等效于 Nginx try_files。
func serveEmbeddedSPA(r *gin.Engine) {
	// 缓存 index.html，所有 SPA 路由回退只读一次
	indexHTML, err := webDist.ReadFile("web-dist/index.html")
	if err != nil {
		logger.Error("嵌入式 frontend 不可用: 缺少 index.html", zap.Error(err))
		return
	}

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 排除健康检查和调试端点，避免被 SPA fallback 拦截返回 index.html
		if path == "/health" ||
			strings.HasPrefix(path, "/api") ||
			strings.HasPrefix(path, "/uploads") ||
			strings.HasPrefix(path, "/system-assets") {
			return
		}

		// 尝试读取请求路径对应文件
		filePath := "web-dist/" + strings.TrimPrefix(path, "/")
		if data, err := webDist.ReadFile(filePath); err == nil {
			c.Data(200, mime.TypeByExtension(filepath.Ext(path)), data)
			return
		}

		// 文件不存在 → SPA fallback (等效 Nginx try_files $uri /index.html)
		c.Data(200, "text/html; charset=utf-8", indexHTML)
	})
}

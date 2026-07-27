package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 设置 HTTP 安全响应头
// 防止常见 Web 攻击：XSS、点击劫持、MIME 嗅探等
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止 MIME 类型嗅探（浏览器不会猜测文件类型）
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止点击劫持（禁止页面被嵌入 iframe）
		c.Header("X-Frame-Options", "DENY")

		// XSS 防护（启用浏览器内置过滤器）
		c.Header("X-XSS-Protection", "1; mode=block")

		// 限制引用来源策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 内容安全策略（限制资源加载来源）
		// 注意：根据实际需求调整，这里采用宽松策略避免影响正常功能
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data: blob:; "+
				"font-src 'self' data:; "+
				"connect-src 'self' ws: wss:")

		// 权限策略（禁用不必要的浏览器 API）
		c.Header("Permissions-Policy",
			"camera=(), microphone=(), geolocation=(")

		// 跨域资源策略
		c.Header("Cross-Origin-Resource-Policy", "same-site")

		c.Next()
	}
}

// NoCache 禁用缓存中间件（用于 API 响应，防止敏感数据被缓存）
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}

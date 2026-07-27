package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/config"
)

// isAllowedOrigin 检查请求 Origin 是否在白名单中
func isAllowedOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	// 通配符 * 不做为有效白名单，防止 Allow-Credentials 被绕过
	for _, allowed := range config.GlobalConfig.CORS.AllowedOrigins {
		trimmed := strings.TrimSpace(allowed)
		if trimmed == "*" || trimmed == "" {
			continue
		}
		if strings.EqualFold(trimmed, origin) {
			return true
		}
	}
	return false
}

func isDevOrigin(origin string) bool {
	devPrefixes := []string{
		"http://localhost",
		"http://127.0.0.1",
		"http://0.0.0.0",
	}
	lowerOrigin := strings.ToLower(origin)
	for _, prefix := range devPrefixes {
		if strings.HasPrefix(lowerOrigin, prefix) {
			return true
		}
	}
	return false
}

// CORS 跨域中间件
//
// 设计原则：
// - 生产模式下，前端和后端通常部署在同一域下（SPA 内嵌在 server 中），
//   浏览器同源请求不会带 Origin 头，不需要任何 CORS 头即可正常工作。
//   仅当用户显式配置了 allowed_origins 时才响应跨域请求（用于分离部署场景）。
// - 开发模式下，仅放行本地地址（localhost/127.0.0.1），配合 Allow-Credentials
//   满足 Vite dev server 的联调需求。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 无 Origin 头 → 同源请求 → 不需要 CORS 头，直接放行
		if origin == "" {
			c.Next()
			return
		}

		allowOrigin := ""
		allowCredentials := false

		if config.GlobalConfig.App.Mode == "development" {
			// 开发模式：仅放行本地地址，允许携带凭证（Cookie/Authorization）
			if isDevOrigin(origin) {
				allowOrigin = origin
				allowCredentials = true
			}
		} else {
			// 生产模式：仅白名单域名允许跨域（用于前后端分离部署场景）
			if isAllowedOrigin(origin) {
				allowOrigin = origin
				allowCredentials = true
			}
		}

		// 未匹配 → 不设任何 CORS 头，浏览器将拒绝跨域请求
		if allowOrigin == "" {
			c.Next()
			return
		}

		c.Header("Access-Control-Allow-Origin", allowOrigin)
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		if allowCredentials {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

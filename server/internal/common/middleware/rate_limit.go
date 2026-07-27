package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
)

// loginLimiter 登录接口专用限流器：每个IP每分钟最多10次尝试
var loginLimiter = utils.NewRateLimiter(10, 60, 5*60)

// captchaLimiter 验证码接口专用限流器：每个IP每分钟最多30次请求
var captchaLimiter = utils.NewRateLimiter(30, 60, 5*60)

// LoginRateLimit 登录接口限流中间件（基于客户端IP）
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// 如果请求经过反向代理，优先取 X-Forwarded-For
		if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
			ip = strings.TrimSpace(ip)
		}

		if !loginLimiter.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "登录请求过于频繁，请稍后再试",
			})
			return
		}

		c.Next()
	}
}

// CaptchaRateLimit 验证码接口限流中间件（基于客户端IP）
func CaptchaRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// 如果请求经过反向代理，优先取 X-Forwarded-For
		if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
			ip = strings.Split(forwarded, ",")[0]
			ip = strings.TrimSpace(ip)
		}

		if !captchaLimiter.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "验证码请求过于频繁，请稍后再试",
			})
			return
		}

		c.Next()
	}
}

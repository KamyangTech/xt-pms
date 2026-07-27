package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
)

// BodyLimit 请求体大小限制中间件
// 防止客户端发送超大请求体导致服务器内存耗尽（DoS 攻击）
func BodyLimit() gin.HandlerFunc {
	maxSize := config.GlobalConfig.Server.MaxBodySize
	if maxSize <= 0 {
		maxSize = 10 * 1024 * 1024 // 默认 10MB
	}

	return func(c *gin.Context) {
		if c.Request.ContentLength > int64(maxSize) {
			response.Error(c, http.StatusRequestEntityTooLarge,
				fmt.Sprintf("请求体超过最大允许大小 (%d MB)", maxSize/(1024*1024)))
			c.Abort()
			return
		}
		c.Next()
	}
}

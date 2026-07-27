package middleware

import (
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"github.com/gin-gonic/gin"
)

// XSSSanitize 自动清理请求中的用户输入，防止 XSS
// 对 URL 查询参数进行统一清理，覆盖搜索关键词等常见输入点
func XSSSanitize() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 清理查询参数
		query := c.Request.URL.Query()
		for key, values := range query {
			for i, v := range values {
				sanitized := utils.SanitizeString(v)
				if sanitized != v {
					query[key][i] = sanitized
				}
			}
		}
		c.Request.URL.RawQuery = query.Encode()
		c.Next()
	}
}

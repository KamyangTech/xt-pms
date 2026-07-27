package monitor

import (
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware 收集所有 HTTP 请求的性能指标
// 放在所有其他中间件之前，以准确测量端到端延迟
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		IncInFlight()
		start := time.Now()

		c.Next()

		DecInFlight()
		latency := time.Since(start)
		RecordRequest(c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
	}
}

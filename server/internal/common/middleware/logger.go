package middleware

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

// sensitiveQueryParams 需要脱敏的 URL 查询参数名
var sensitiveQueryParams = map[string]struct{}{
	"token":       {},
	"access_token": {},
	"code":        {}, // OAuth 授权码
}

func redactQuery(rawQuery string) string {
	if rawQuery == "" || len(sensitiveQueryParams) == 0 {
		return rawQuery
	}
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		return rawQuery
	}
	for param := range values {
		if _, isSensitive := sensitiveQueryParams[strings.ToLower(param)]; isSensitive {
			values[param] = []string{"[REDACTED]"}
		}
	}
	return values.Encode()
}

// Logger 基于 zap 的 Gin HTTP 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		// 跳过 OPTIONS 预检请求日志，避免刷屏
		if c.Request.Method == "OPTIONS" {
			return
		}

		cost := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		ip := c.ClientIP()
		errStr := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// 对查询参数中的敏感字段（token 等）进行脱敏
		safeQuery := redactQuery(query)

		var statusColor string
		switch {
		case status >= 500:
			statusColor = "\033[1;31m" // 粗体红色
		case status >= 400:
			statusColor = "\033[1;33m" // 粗体黄色
		case status >= 300:
			statusColor = "\033[1;36m" // 粗体青色
		default:
			statusColor = "\033[1;32m" // 粗体绿色 (2xx)
		}
		msg := fmt.Sprintf("HTTP %s%d\033[0m %s %s", statusColor, status, method, path)
		fields := []zap.Field{
			zap.String("query", safeQuery),
			zap.String("ip", ip),
			zap.Duration("cost", cost),
		}

		switch {
		case len(errStr) > 0:
			logger.Error(msg, append(fields, zap.String("errors", errStr))...)
		case status >= 500:
			logger.Error(msg, fields...)
		case status >= 400:
			logger.Warn(msg, fields...)
		default:
			logger.Info(msg, fields...)
		}
	}
}

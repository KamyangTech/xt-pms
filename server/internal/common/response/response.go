package response

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

const (
	SUCCESS      = 200
	BAD_REQUEST  = 400
	UNAUTHORIZED = 401
	FORBIDDEN    = 403
	NOT_FOUND    = 404
	ERROR        = 500
)

// HttpError 带 HTTP 状态码的业务错误，用于区分 4xx 和 5xx
type HttpError struct {
	Code    int    // 业务状态码（如 400/401/403/500）
	Message string // 用户端提示信息
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// safeMessage 在生产环境下将内部错误信息替换为通用提示，避免泄露内部细节
// 策略：仅对包含明确敏感关键词（SQL 错误、堆栈信息、文件路径等）的消息进行脱敏，
// 避免误杀 Gin 参数校验错误（如 "Error:Field validation..."）
func safeMessage(message string) string {
	if config.GlobalConfig.App.Mode != "production" {
		return message
	}
	// 仅对包含明确技术细节的消息进行脱敏：
	// - SQL 相关：sqlstate, sql error 等
	// - 文件路径：.go 文件、.sql 文件
	// - 连接错误：connection refused 等
	// - 堆栈/panic：panic、stack trace
	// 注意：不再匹配泛化的 "error:" 和 "/"，避免误伤 Gin 框架参数绑定校验的错误提示
	sensitiveKeywords := []string{
		// PostgreSQL
		"sqlstate",
		"sql error",
		// MySQL 特有错误码
		"error 1054",      // Unknown column
		"error 1062",      // Duplicate entry
		"error 1146",      // Table doesn't exist
		"error 1451",      // Foreign key constraint (DELETE)
		"error 1452",      // Foreign key constraint (INSERT/UPDATE)
		"duplicate entry", // MySQL 唯一约束冲突详情
		"unknown column",  // MySQL 未知列
		"doesn't exist",   // MySQL 表/库不存在
		// 通用
		"db error",
		"panic",
		"stack trace",
		"goroutine",
		".go:",              // Go 文件路径+行号 (Unix)
		".sql:",             // SQL 文件路径
		"runtime.",          // Go runtime 堆栈
		"connection refused",
		"no such host",
	}
	lowerMsg := strings.ToLower(message)
	for _, kw := range sensitiveKeywords {
		if strings.Contains(lowerMsg, kw) {
			return "服务器内部错误，请稍后重试"
		}
	}
	return message
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    SUCCESS,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	httpCode := http.StatusOK
	switch code {
	case SUCCESS:
		httpCode = http.StatusOK
	case BAD_REQUEST:
		httpCode = http.StatusBadRequest
	case UNAUTHORIZED:
		httpCode = http.StatusUnauthorized
	case FORBIDDEN:
		httpCode = http.StatusForbidden
	case NOT_FOUND:
		httpCode = http.StatusNotFound
	case ERROR:
		httpCode = http.StatusInternalServerError
	}
	c.JSON(httpCode, Response{
		Code:    code,
		Message: safeMessage(message),
	})
}

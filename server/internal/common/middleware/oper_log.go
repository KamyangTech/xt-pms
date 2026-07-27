package middleware

import (
	"bytes"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

// 操作日志写入配置
const (
	logChanSize      = 1024            // channel 缓冲区大小
	logBatchSize     = 50              // 每批最大条数
	logFlushInterval = 3 * time.Second // 最长刷新间隔
	maxLogBodySize   = 2048            // 请求体日志最大字节数
)

var operLogDAO = dao.NewSysOperLogDAO()
var logEntryCh = make(chan *entity.SysOperLog, logChanSize)
var logWorkerWg sync.WaitGroup

func init() {
	logWorkerWg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("操作日志 worker panic，进程可能不稳定", zap.Any("panic", r))
			}
			logWorkerWg.Done()
		}()
		logWorker()
	}()
}

// StopLogWorker 优雅关闭日志 worker，等待所有待写入日志刷入 DB
func StopLogWorker() {
	close(logEntryCh)
	logWorkerWg.Wait()
}

// logWorker 后台协程：批量消费日志 channel，定时/定量刷入 DB
func logWorker() {
	batch := make([]*entity.SysOperLog, 0, logBatchSize)
	ticker := time.NewTicker(logFlushInterval)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := operLogDAO.BatchCreate(batch); err != nil {
			logger.Error("操作日志批量写入失败", zap.Int("count", len(batch)), zap.Error(err))
		}
		batch = batch[:0]
	}

	for {
		select {
		case entry, ok := <-logEntryCh:
			if !ok {
				flush()
				return
			}
			batch = append(batch, entry)
			if len(batch) >= logBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

// bodyLogWriter 包装 ResponseWriter，最多捕获前 maxLogBodySize 字节的响应内容
type bodyLogWriter struct {
	gin.ResponseWriter
	body    *bytes.Buffer
	maxSize int
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	if w.body.Len() < w.maxSize {
		remain := w.maxSize - w.body.Len()
		if len(b) > remain {
			w.body.Write(b[:remain])
		} else {
			w.body.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	if w.body.Len() < w.maxSize {
		remain := w.maxSize - w.body.Len()
		if len(s) > remain {
			w.body.WriteString(s[:remain])
		} else {
			w.body.WriteString(s)
		}
	}
	return w.ResponseWriter.WriteString(s)
}

func (w *bodyLogWriter) WriteHeader(code int) {
	// Gin 内部可能会多次调用 WriteHeader，只转发给原始 writer
	w.ResponseWriter.WriteHeader(code)
}

// OperationLog 操作日志中间件
// 控制器通过设置 Context 值来标注日志信息：
//
//	c.Set("oper_log_title", "创建模具")        // 必填，模块标题
//	c.Set("oper_log_type", 1)                 // 必填，操作类型 1=增 2=删 3=改
//	c.Set("oper_log_business_type", "mold")   // 可选，业务类型标识
//	c.Set("oper_log_business_id", int64(123)) // 可选，业务记录ID
//	c.Set("oper_log_error", "错误信息")        // 可选，手动标记异常
func OperationLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只记录非 GET 请求（查询操作不产生操作日志）
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		// 多部件上传不记录 body（文件内容无业务意义且可能很大）
		if strings.HasPrefix(c.ContentType(), "multipart/form-data") {
			c.Next()
			return
		}

		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		logBody := body
		if len(logBody) > maxLogBodySize {
			logBody = logBody[:maxLogBodySize]
		}

		// 包装响应写入器以捕获返回内容（限制最大捕获量）
		blw := &bodyLogWriter{
			body:           bytes.NewBuffer(nil),
			ResponseWriter: c.Writer,
			maxSize:        maxLogBodySize,
		}
		c.Writer = blw

		start := time.Now()
		c.Next()
		cost := time.Since(start)

		rawTitle, hasTitle := c.Get("oper_log_title")
		rawType, hasType := c.Get("oper_log_type")
		if !hasTitle || !hasType {
			return
		}

		status := c.Writer.Status()
		logEntry := entity.SysOperLog{
			Title:         toString(rawTitle),
			OperType:      toInt8(rawType),
			Method:        c.FullPath(),
			RequestURL:    c.Request.URL.Path,
			RequestMethod: c.Request.Method,
			OperName:      GetUsername(c),
			OperIP:        c.ClientIP(),
			RequestParams: truncateStr(maskSensitiveFields(string(logBody)), 2000),
			JSONResult:    truncateStr(maskSensitiveFields(blw.body.String()), 2000),
			Status:        boolToInt8(status >= 400),
			CostTime:      cost.Milliseconds(),
			CreateTime:    time.Now(),
		}

		// 可选业务字段
		if bt, ok := c.Get("oper_log_business_type"); ok {
			logEntry.BusinessType = toString(bt)
		}
		if bid, ok := c.Get("oper_log_business_id"); ok {
			logEntry.BusinessID = toInt64(bid)
		}
		if errMsg, ok := c.Get("oper_log_error"); ok {
			logEntry.ErrorMsg = toString(errMsg)
			logEntry.Status = 1
		}

		// 异步推送到 channel（非阻塞，channel 满了直接丢弃）
		select {
		case logEntryCh <- &logEntry:
		default:
			logger.Warn("操作日志 channel 已满，丢弃日志",
				zap.String("title", logEntry.Title),
				zap.String("oper_name", logEntry.OperName),
			)
		}
	}
}

// --- 敏感数据脱敏 ---

var sensitivePattern = regexp.MustCompile(
	`"(password|old_password|new_password|token|secret|authorization|access_token|refresh_token|api_key|secret_key)"\s*:\s*"[^"]+"`,
)

func maskSensitiveFields(raw string) string {
	// 先通过正则替换 JSON body 中的敏感字段
	result := sensitivePattern.ReplaceAllString(raw, `"$1":"***"`)
	// 再替换 URL 查询参数中的敏感信息
	if idx := strings.Index(result, "?"); idx > 0 {
		before := result[:idx]
		query := result[idx:]
		query = regexp.MustCompile(`(password|token|secret)=[^&]+`).
			ReplaceAllString(query, `$1=***`)
		result = before + query
	}
	return result
}

// --- 辅助函数 ---

func toString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toInt8(v interface{}) int8 {
	switch t := v.(type) {
	case int:
		return int8(t)
	case int8:
		return t
	case int32:
		return int8(t)
	case int64:
		return int8(t)
	case float64:
		return int8(t)
	}
	return 0
}

func toInt64(v interface{}) int64 {
	switch t := v.(type) {
	case int:
		return int64(t)
	case int64:
		return t
	case float64:
		return int64(t)
	}
	return 0
}

func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}

func truncateStr(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return s
}

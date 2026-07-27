package logger

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// sensitiveFields 需要脱敏的字段名（不区分大小写匹配）
var sensitiveFields = []string{
	"password", "passwd", "pwd",
	"secret", "token", "authorization",
	"api_key", "apikey", "access_key",
}

// redactSQL 对 SQL 中敏感字段对应的参数值进行脱敏
func redactSQL(sql string) string {
	for _, field := range sensitiveFields {
		// 匹配 field = 'xxx' 或 field = xxx 或 field=? 后的绑定值
		pattern := regexp.MustCompile(`(?i)(` + field + `)\s*=\s*'?[^',)]*'?`)
		sql = pattern.ReplaceAllString(sql, "${1}='[REDACTED]'")
	}
	return sql
}

func isRecordNotFoundErr(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// GormLogger 将 zap 日志适配为 GORM 的日志接口
type GormLogger struct {
	gormlogger.Config
	zapLogger *zap.Logger
}

// NewGormLogger 创建基于 zap 的 GORM 日志适配器
func NewGormLogger(cfg gormlogger.Config) *GormLogger {
	return &GormLogger{
		Config:    cfg,
		zapLogger: Logger.WithOptions(zap.AddCallerSkip(3)),
	}
}

// LogMode 设置日志级别
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 输出 Info 级别日志
func (l *GormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.zapLogger.Info(fmt.Sprintf(msg, args...))
	}
}

// Warn 输出 Warn 级别日志
func (l *GormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.zapLogger.Warn(fmt.Sprintf(msg, args...))
	}
}

// Error 输出 Error 级别日志
func (l *GormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.zapLogger.Error(fmt.Sprintf(msg, args...))
	}
}

// Trace 记录 SQL 执行日志（关键方法：记录 SQL、行数、耗时）
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// 对 SQL 中的敏感字段值进行脱敏（password、token 等）
	redactedSQL := redactSQL(sql)

	slowThreshold := l.SlowThreshold
	ignoreRecordNotFound := l.IgnoreRecordNotFoundError
	fileLine := utils.FileWithLineNum()

	fields := []zap.Field{
		zap.String("sql", redactedSQL),
		zap.Duration("elapsed", elapsed),
		zap.Int64("rows", rows),
		zap.String("file", fileLine),
	}

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error:
		if ignoreRecordNotFound && isRecordNotFoundErr(err) {
			l.zapLogger.Debug("DB", fields...)
		} else {
			l.zapLogger.Error("DB异常", append(fields, zap.Error(err))...)
		}
	case elapsed > slowThreshold && slowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		l.zapLogger.Warn("DB慢查询", fields...)
	case l.LogLevel >= gormlogger.Info:
		l.zapLogger.Debug("DB", fields...)
	}
}

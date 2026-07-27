package logger

import (
	"os"
	"time"

	"github.com/KamyangTech/xt-pms/server/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

func Init() error {
	// 使用配置文件，如果没有则使用默认值
	logConfig := config.GlobalConfig.Log
	if logConfig.Level == "" {
		logConfig.Level = "info"
	}
	if logConfig.Format == "" {
		logConfig.Format = "json"
	}
	if logConfig.Output == "" {
		logConfig.Output = "both"
	}
	if logConfig.File.Path == "" {
		logConfig.File.Path = "logs/pms.log"
	}
	if logConfig.File.MaxSize == 0 {
		logConfig.File.MaxSize = 100
	}
	if logConfig.File.MaxBackups == 0 {
		logConfig.File.MaxBackups = 5
	}
	if logConfig.File.MaxAge == 0 {
		logConfig.File.MaxAge = 30
	}

	// 配置日志级别
	level := parseLogLevel(logConfig.Level)

	// 配置日志轮转
	lumberJackLogger := &lumberjack.Logger{
		Filename:   logConfig.File.Path,
		MaxSize:    logConfig.File.MaxSize,
		MaxBackups: logConfig.File.MaxBackups,
		MaxAge:     logConfig.File.MaxAge,
		Compress:   logConfig.File.Compress,
	}

	writeSyncer := zapcore.AddSync(lumberJackLogger)

	// 选择编码器
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.CallerKey = "caller"
	encoderConfig.MessageKey = "msg"
	encoderConfig.StacktraceKey = "stacktrace"

	if logConfig.Format == "console" {
		// console 格式：着色 + 简短时间，终端观感更好
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			// 白字黑底(反色)，每条日志起点清晰分隔
			enc.AppendString("\033[7m" + t.Format("01-02 15:04:05.000") + "\033[0m")
		}
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		// JSON 格式：保留完整 ISO8601 时间
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 配置输出目标
	var cores []zapcore.Core

	switch logConfig.Output {
	case "console":
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	case "file":
		cores = append(cores, zapcore.NewCore(encoder, writeSyncer, level))
	case "both":
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
		cores = append(cores, zapcore.NewCore(encoder, writeSyncer, level))
	default:
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
		cores = append(cores, zapcore.NewCore(encoder, writeSyncer, level))
	}

	core := zapcore.NewTee(cores...)

	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return nil
}

func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Sync 刷新日志缓冲区，在程序退出前调用
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

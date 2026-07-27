package utils

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type RetryConfig struct {
	MaxRetries   int           // 最大重试次数
	InitialDelay time.Duration // 初始延迟
	MaxDelay     time.Duration // 最大延迟
	Multiplier   float64       // 延迟倍数（指数退避）
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxRetries:   3,
	InitialDelay: 1 * time.Second,
	MaxDelay:     10 * time.Second,
	Multiplier:   2.0,
}

// Retry 执行带重试的操作
// operation: 要执行的操作函数，返回 error
// config: 重试配置，如果为 nil 则使用默认配置
// operationName: 操作名称，用于日志记录
func Retry(operation func() error, config *RetryConfig, operationName string) error {
	if config == nil {
		config = &DefaultRetryConfig
	}

	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		err := operation()
		if err == nil {
			// 成功
			if attempt > 0 {
				logger.Info("操作重试成功",
					zap.String("operation", operationName),
					zap.Int("attempt", attempt+1))
			}
			return nil
		}

		lastErr = err

		// 如果是最后一次尝试，直接返回错误
		if attempt == config.MaxRetries {
			break
		}

		// 记录重试日志
		logger.Warn("操作失败，准备重试",
			zap.String("operation", operationName),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", config.MaxRetries),
			zap.Error(err),
			zap.Duration("next_delay", delay))

		// 等待后重试
		time.Sleep(delay)

		// 指数退避：delay = delay * multiplier，但不超过 maxDelay
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	// 所有重试都失败了
	logger.Error("操作重试失败",
		zap.String("operation", operationName),
		zap.Int("total_attempts", config.MaxRetries+1),
		zap.Error(lastErr))

	return lastErr
}

// RetryAsync 异步执行带重试的操作（不阻塞调用方）
func RetryAsync(operation func() error, config *RetryConfig, operationName string) {
	go func() {
		if err := Retry(operation, config, operationName); err != nil {
			// 这里可以添加告警通知逻辑
			logger.Error("异步操作最终失败，建议人工介入",
				zap.String("operation", operationName),
				zap.Error(err))
		}
	}()
}

package utils

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type PerformanceMonitor struct {
	startTime time.Time
	operation string
	threshold time.Duration // 慢查询阈值
}

func NewPerformanceMonitor(operation string, threshold ...time.Duration) *PerformanceMonitor {
	t := 1 * time.Second // 默认1秒为慢查询
	if len(threshold) > 0 {
		t = threshold[0]
	}

	return &PerformanceMonitor{
		startTime: time.Now(),
		operation: operation,
		threshold: t,
	}
}

// End 结束监控并记录
func (pm *PerformanceMonitor) End() {
	duration := time.Since(pm.startTime)

	// 记录所有操作的耗时
	logger.Debug("操作耗时",
		zap.String("operation", pm.operation),
		zap.Duration("duration", duration),
	)

	// 如果超过阈值，记录警告
	if duration > pm.threshold {
		logger.Warn("慢操作警告",
			zap.String("operation", pm.operation),
			zap.Duration("duration", duration),
			zap.Duration("threshold", pm.threshold),
		)
	}
}

func RecordDBOperation(operation string, fn func() error) error {
	pm := NewPerformanceMonitor("db:"+operation, 500*time.Millisecond)
	defer pm.End()

	return fn()
}

func RecordServiceOperation(operation string, fn func() error) error {
	pm := NewPerformanceMonitor("service:"+operation, 1*time.Second)
	defer pm.End()

	return fn()
}

func RecordControllerOperation(operation string, fn func()) {
	pm := NewPerformanceMonitor("controller:"+operation, 2*time.Second)
	defer pm.End()

	fn()
}

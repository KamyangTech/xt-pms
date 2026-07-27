package monitor

import (
	"fmt"
	"runtime"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

// goroutine 泄漏检测阈值
const (
	goroutineWarnThreshold  = 5000   // goroutine 数超过此值发出警告
	goroutineAlertThreshold = 50000  // goroutine 数超过此值发出严重告警
	poolWarnThreshold       = 0.85   // 连接池使用率 > 85% 警告
	poolAlertThreshold      = 0.95   // 连接池使用率 > 95% 严重告警
	watchInterval           = 30 * time.Second
)

// StartWatcher 启动后台性能监控协程
// 定期检查 goroutine 数量和 DB 连接池状态
func StartWatcher() {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("性能监控 watcher panic", zap.Any("panic", r))
			}
		}()

		ticker := time.NewTicker(watchInterval)
		defer ticker.Stop()

		// 记录初始 goroutine 数作为基线
		initialGoroutines := runtime.NumGoroutine()
		logger.Info("性能监控 watcher 已启动",
			zap.Int("initial_goroutines", initialGoroutines),
			zap.Duration("interval", watchInterval))

		for range ticker.C {
			checkGoroutines()
			checkConnectionPool()
		}
	}()
}

func checkGoroutines() {
	num := runtime.NumGoroutine()
	if num > goroutineAlertThreshold {
		logger.Error("⚠️ CRITICAL: goroutine 泄漏严重",
			zap.Int("goroutines", num),
			zap.Int("threshold", goroutineAlertThreshold))
	} else if num > goroutineWarnThreshold {
		logger.Warn("⚠️ 可能存在 goroutine 泄漏",
			zap.Int("goroutines", num),
			zap.Int("threshold", goroutineWarnThreshold))
	}
}

func checkConnectionPool() {
	sqlDB, err := dao.DB.DB()
	if err != nil {
		return
	}
	stats := sqlDB.Stats()

	if stats.MaxOpenConnections <= 0 {
		return
	}

	usage := float64(stats.InUse) / float64(stats.MaxOpenConnections)
	waiting := stats.WaitCount

	fields := []zap.Field{
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int("max_open", stats.MaxOpenConnections),
		zap.String("usage", fmt.Sprintf("%.0f%%", usage*100)),
		zap.Int64("waiting", waiting),
	}

	if usage >= poolAlertThreshold {
		logger.Error("⚠️ CRITICAL: DB 连接池即将耗尽", fields...)
	} else if usage >= poolWarnThreshold {
		logger.Warn("⚠️ DB 连接池使用率偏高", fields...)
	}
}

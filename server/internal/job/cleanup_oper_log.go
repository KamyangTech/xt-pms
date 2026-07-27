package job

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

// CleanupOperLogRetentionDays 操作日志保留天数（默认 180 = 半年）
var CleanupOperLogRetentionDays = 180

// CleanupOperLogs 清理超过指定天数的操作日志
// days: 保留天数，<=0 表示清理全部
func CleanupOperLogs(days int) {
	if days <= 0 {
		days = CleanupOperLogRetentionDays
	}

	before := time.Now().AddDate(0, 0, -days)
	logDAO := dao.NewSysOperLogDAO()

	count, err := logDAO.DeleteOlderThan(before)
	if err != nil {
		logger.Error("[操作日志清理] 清理失败", zap.Error(err))
		return
	}

	if count > 0 {
		logger.Info("[操作日志清理] 清理完成",
			zap.Int("deleted_count", int(count)),
			zap.Int("retention_days", days),
			zap.Time("before", before),
		)
	} else {
		logger.Debug("[操作日志清理] 无需清理", zap.Int("retention_days", days))
	}
}

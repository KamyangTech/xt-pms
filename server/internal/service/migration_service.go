// Package service - 数据库增量迁移
//
// 迁移机制：
// 1. 将 SQL 变更放入 migrations/updates/，命名格式 vX.Y.Z.sql（按版本号顺序执行）
// 2. 每次服务启动时自动检查，已执行过的（记录在 sys_migration 表）不会重复执行
// 3. 无需手动执行 SQL，在线更新/手动替换/supervisor 重启都会触发
//
// 示例：
//   migrations/updates/
//   ├── v0.2.0.sql   # 添加系统监控菜单
//   ├── v0.2.1.sql   # 新增 XX 索引
//   └── v0.3.0.sql   # 新增 XX 表
//
// 幂等建议：SQL 内部尽量使用 INSERT IGNORE / ALTER TABLE ADD COLUMN（MySQL 不支持 IF NOT EXISTS 时需要先查 INFORMATION_SCHEMA）
package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

// MigrationService 数据库增量迁移服务
// 在每次启动时自动扫描 migrations/updates/ 目录，执行未应用过的 SQL 迁移
type MigrationService struct{}

func NewMigrationService() *MigrationService {
	return &MigrationService{}
}

// RunPendingMigrations 执行所有待处理的迁移文件
// 迁移文件放在 migrations/updates/ 目录下，命名为 vX.Y.Z.sql
// 执行顺序按版本号排序，已执行过的会记录在 sys_migration 表中，不会重复执行
func (s *MigrationService) RunPendingMigrations() error {
	// 确保追踪表存在
	if err := s.ensureMigrationTable(); err != nil {
		return fmt.Errorf("创建迁移追踪表失败: %w", err)
	}

	// 扫描迁移文件
	files, err := s.scanMigrationFiles()
	if err != nil {
		return fmt.Errorf("扫描迁移文件失败: %w", err)
	}

	if len(files) == 0 {
		return nil
	}

	applied, err := s.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("查询已执行迁移失败: %w", err)
	}

	// 执行未应用的迁移
	executed := 0
	for _, f := range files {
		version := strings.TrimSuffix(f, ".sql")

		if applied[version] {
			continue
		}

		logger.Info("正在执行数据库迁移",
			zap.String("version", version),
			zap.String("file", f))

		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("读取迁移文件 %s 失败: %w", f, err)
		}

		if err := dao.DB.Exec(string(sql)).Error; err != nil {
			return fmt.Errorf("执行迁移 %s 失败: %w", version, err)
		}

		// 记录已执行
		if err := dao.DB.Exec(
			"INSERT INTO sys_migration (version, filename, executed_at) VALUES (?, ?, ?)",
			version, f, time.Now(),
		).Error; err != nil {
			return fmt.Errorf("记录迁移 %s 失败: %w", version, err)
		}

		logger.Info("数据库迁移完成", zap.String("version", version))
		executed++
	}

	if executed > 0 {
		logger.Info("所有数据库迁移已执行", zap.Int("count", executed))
	}
	return nil
}

func (s *MigrationService) ensureMigrationTable() error {
	return dao.DB.Exec(`
		CREATE TABLE IF NOT EXISTS sys_migration (
			id          BIGINT AUTO_INCREMENT PRIMARY KEY,
			version     VARCHAR(32)  NOT NULL,
			filename    VARCHAR(255) NOT NULL,
			executed_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY uk_version (version)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据库迁移版本追踪表'
	`).Error
}

// scanMigrationFiles 扫描所有候选目录中的 .sql 迁移文件
// 查找优先级: migrations/updates/ > migrations/data_seeds/
func (s *MigrationService) scanMigrationFiles() ([]string, error) {
	dirs := []string{
		"migrations/updates",
		"migrations",
		"../migrations/updates",
		"../migrations",
	}

	// 只获取 v*.sql 格式的迁移文件
	var files []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasPrefix(e.Name(), "v") || !strings.HasSuffix(e.Name(), ".sql") {
				continue
			}
			files = append(files, filepath.Join(dir, e.Name()))
		}
		if len(files) > 0 {
			break
		}
	}

	// 按版本号排序
	sort.Slice(files, func(i, j int) bool {
		return compareVersionStrings(
			strings.TrimSuffix(filepath.Base(files[i]), ".sql"),
			strings.TrimSuffix(filepath.Base(files[j]), ".sql"),
		) < 0
	})

	return files, nil
}

func (s *MigrationService) getAppliedMigrations() (map[string]bool, error) {
	rows, err := dao.DB.Raw("SELECT version FROM sys_migration").Rows()
	if err != nil {
		// 表可能不存在，返回空 map
		return make(map[string]bool), nil
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, nil
}

// compareVersionStrings 简单比较 v0.1.0 < v0.2.0 这种格式
func compareVersionStrings(a, b string) int {
	partsA := strings.Split(strings.TrimPrefix(a, "v"), ".")
	partsB := strings.Split(strings.TrimPrefix(b, "v"), ".")

	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := 0; i < maxLen; i++ {
		va, vb := 0, 0
		if i < len(partsA) {
			fmt.Sscanf(partsA[i], "%d", &va)
		}
		if i < len(partsB) {
			fmt.Sscanf(partsB[i], "%d", &vb)
		}
		if va < vb {
			return -1
		}
		if va > vb {
			return 1
		}
	}
	return 0
}

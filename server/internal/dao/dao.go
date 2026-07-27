package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// maskDSN 脱敏 DSN 中的密码，用于日志输出（仅显示前 3 位 + ***）
func maskDSN(dsn string) string {
	// 格式: user:password@tcp(host:port)/dbname?...
	atIdx := strings.Index(dsn, "@")
	if atIdx == -1 {
		return dsn
	}
	colonIdx := strings.Index(dsn, ":")
	if colonIdx == -1 || colonIdx >= atIdx {
		return dsn
	}
	// 保留用户名和冒号，密码部分脱敏
	return dsn[:colonIdx+1] + "***" + dsn[atIdx:]
}

func InitDB(cfg *config.Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	var logLevel gormlogger.LogLevel
	if strings.EqualFold(strings.TrimSpace(cfg.App.Mode), "production") {
		logLevel = gormlogger.Silent
	} else {
		switch strings.ToLower(strings.TrimSpace(cfg.Log.Level)) {
		case "debug":
			logLevel = gormlogger.Info
		case "info":
			logLevel = gormlogger.Warn
		case "warn":
			logLevel = gormlogger.Warn
		case "error":
			logLevel = gormlogger.Error
		default:
			logLevel = gormlogger.Warn
		}
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.NewGormLogger(gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		}),
	})

	if err != nil {
		logger.Error("连接数据库失败",
			zap.String("dsn", maskDSN(dsn)),
			zap.Error(err))
		return fmt.Errorf("连接数据库失败：%w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取 sql.DB 失败：%w", err)
	}

	// 设置连接池参数（从配置读取，含默认值）
	maxOpenConns := cfg.Database.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = 100
	}
	maxIdleConns := cfg.Database.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 25
	}
	connMaxLifetime := cfg.Database.ConnMaxLifetime
	if connMaxLifetime <= 0 {
		connMaxLifetime = 10 * time.Minute
	}
	connMaxIdleTime := cfg.Database.ConnMaxIdleTime
	if connMaxIdleTime <= 0 {
		connMaxIdleTime = 5 * time.Minute
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	logger.Info("数据库连接池配置",
		zap.Int("max_open_conns", maxOpenConns),
		zap.Int("max_idle_conns", maxIdleConns),
		zap.Duration("conn_max_lifetime", connMaxLifetime),
		zap.Duration("conn_max_idle_time", connMaxIdleTime))

	return nil
}

func GetDB() *gorm.DB {
	return DB
}

// FindSuperAdmin 查询超级管理员用户名（用于重置密码）
func FindSuperAdmin() (string, error) {
	type result struct {
		Username string
	}
	var r result
	err := DB.Raw(`
		SELECT u.username
		FROM sys_user u
		JOIN sys_user_role ur ON u.id = ur.user_id
		JOIN sys_role r ON ur.role_id = r.id
		WHERE r.role_name = '超级管理员'
		LIMIT 1
	`).Scan(&r).Error
	if err != nil {
		return "", err
	}
	return r.Username, nil
}

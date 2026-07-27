package monitor

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
)

type PoolStats struct {
	MaxOpen       int `json:"max_open_connections"`
	Open          int `json:"open_connections"`
	Idle          int `json:"idle_connections"`
	InUse         int `json:"in_use_connections"`
	Waiting       int64 `json:"waiting_count"`
	WaitDuration  string `json:"wait_duration"`  // 累计等待时间
	MaxIdleClosed int64 `json:"max_idle_closed"`  // 因超过 MaxIdleConns 关闭的连接数
	MaxLifetimeClosed int64 `json:"max_lifetime_closed"` // 因超过 ConnMaxLifetime 关闭的连接数
}

type HealthStatus struct {
	Status    string    `json:"status"`     // "ok" / "degraded" / "down"
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`

	DB      *DBHealth      `json:"database"`
	Runtime *RuntimeHealth `json:"runtime"`
}

type DBHealth struct {
	Status string    `json:"status"`      // "ok" / "error"
	Error  string    `json:"error,omitempty"`
	Pool   *PoolStats `json:"pool,omitempty"`
}

type RuntimeHealth struct {
	Goroutines  int    `json:"goroutines"`
	HeapAlloc   string `json:"heap_alloc_mb"`
	HeapSys     string `json:"heap_sys_mb"`
	NumGC       uint32 `json:"num_gc"`
	NumCPU      int    `json:"num_cpu"`
	GoVersion   string `json:"go_version"`
}

var healthMu sync.RWMutex
var lastDBErr error
var lastDBCheck time.Time

func CheckHealth() *HealthStatus {
	h := &HealthStatus{
		Timestamp: time.Now(),
		Uptime:    fmt.Sprintf("%.0fs", time.Since(globalMetrics.uptime).Seconds()),
	}

	// 数据库检查
	h.DB = checkDB()

	// 运行时检查
	h.Runtime = checkRuntime()

	// 综合判定
	if h.DB.Status != "ok" {
		h.Status = "down"
	} else if h.Runtime.Goroutines > 10000 {
		h.Status = "degraded"
	} else {
		h.Status = "ok"
	}

	return h
}

func checkDB() *DBHealth {
	dbh := &DBHealth{Status: "ok"}

	sqlDB, err := dao.DB.DB()
	if err != nil {
		dbh.Status = "error"
		dbh.Error = fmt.Sprintf("获取 sql.DB 失败: %v", err)
		return dbh
	}

	// 连接池快照
	stats := sqlDB.Stats()
	dbh.Pool = &PoolStats{
		MaxOpen:            stats.MaxOpenConnections,
		Open:               stats.OpenConnections,
		Idle:               stats.Idle,
		InUse:              stats.InUse,
		Waiting:            stats.WaitCount,
		WaitDuration:       stats.WaitDuration.String(),
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}

	// Ping 测试（带超时）
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	healthMu.Lock()
	defer healthMu.Unlock()

	if err := sqlDB.PingContext(ctx); err != nil {
		dbh.Status = "error"
		dbh.Error = fmt.Sprintf("DB ping 失败: %v", err)
		lastDBErr = err
	} else {
		lastDBErr = nil
	}
	lastDBCheck = time.Now()

	// 连接池使用率预警（>80% 则降级）
	if stats.MaxOpenConnections > 0 {
		usage := float64(stats.InUse) / float64(stats.MaxOpenConnections)
		if usage > 0.9 && dbh.Status == "ok" {
			dbh.Status = "degraded"
			dbh.Error = fmt.Sprintf("连接池使用率 %.0f%%", usage*100)
		}
	}

	return dbh
}

func checkRuntime() *RuntimeHealth {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	return &RuntimeHealth{
		Goroutines: runtime.NumGoroutine(),
		HeapAlloc:  fmt.Sprintf("%.1f", float64(mem.HeapAlloc)/1024/1024),
		HeapSys:    fmt.Sprintf("%.1f", float64(mem.HeapSys)/1024/1024),
		NumGC:      mem.NumGC,
		NumCPU:     runtime.NumCPU(),
		GoVersion:  runtime.Version(),
	}
}

// GetPoolStats 单独获取连接池统计（无需 ping）
func GetPoolStats() *PoolStats {
	sqlDB, err := dao.DB.DB()
	if err != nil {
		return nil
	}
	stats := sqlDB.Stats()
	return &PoolStats{
		MaxOpen:           stats.MaxOpenConnections,
		Open:              stats.OpenConnections,
		Idle:              stats.Idle,
		InUse:             stats.InUse,
		Waiting:           stats.WaitCount,
		WaitDuration:      stats.WaitDuration.String(),
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}
}

package utils

import (
	"sync"
	"time"
)

// RateLimiter 简单的内存限流器（滑动窗口）
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*window
	rate     int           // 窗口内允许的最大请求数
	window   time.Duration // 时间窗口大小
	cleanup  time.Duration // 清理过期条目间隔
}

type window struct {
	entries []time.Time // 请求时间戳
}

func NewRateLimiter(rate int, windowDur, cleanup time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*window),
		rate:     rate,
		window:   windowDur,
		cleanup:  cleanup,
	}
	go rl.cleanupLoop()
	return rl
}

// Allow 检查是否允许通过，true=允许，false=被限流
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	w, exists := rl.visitors[key]
	if !exists {
		rl.visitors[key] = &window{entries: []time.Time{now}}
		return true
	}

	// 移除窗口之前的过期记录
	cutoff := now.Add(-rl.window)
	valid := 0
	for _, t := range w.entries {
		if t.After(cutoff) {
			w.entries[valid] = t
			valid++
		}
	}
	w.entries = w.entries[:valid]

	if len(w.entries) >= rl.rate {
		return false // 限流
	}

	w.entries = append(w.entries, now)
	return true
}

// cleanupLoop 定期清理过期条目，防止内存泄漏
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-rl.window)
		for key, w := range rl.visitors {
			valid := 0
			for _, t := range w.entries {
				if t.After(cutoff) {
					w.entries[valid] = t
					valid++
				}
			}
			if valid == 0 {
				delete(rl.visitors, key)
			} else {
				w.entries = w.entries[:valid]
			}
		}
		rl.mu.Unlock()
	}
}

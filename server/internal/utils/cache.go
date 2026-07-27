package utils

import (
	"fmt"
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Value     interface{}
	ExpiredAt time.Time
}

// CacheService 内存缓存服务（线程安全）
type CacheService struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

var cacheInstance *CacheService
var once sync.Once

// GetCache 获取缓存实例（单例）
func GetCache() *CacheService {
	once.Do(func() {
		cacheInstance = &CacheService{
			items: make(map[string]*CacheItem),
		}
		go cacheInstance.cleanup()
	})
	return cacheInstance
}

// Set 设置缓存
func (c *CacheService) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = &CacheItem{
		Value:     value,
		ExpiredAt: time.Now().Add(ttl),
	}
}

// Get 获取缓存
func (c *CacheService) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiredAt) {
		return nil, false
	}

	return item.Value, true
}

// Delete 删除缓存
func (c *CacheService) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// GetWorkOrderProgress 获取工单进度（带缓存）
func (c *CacheService) GetWorkOrderProgress(workOrderID int64) (int, bool) {
	cacheKey := fmt.Sprintf("work_order_progress:%d", workOrderID)
	value, ok := c.Get(cacheKey)
	if !ok {
		return 0, false
	}

	progress, ok := value.(int)
	return progress, ok
}

// SetWorkOrderProgress 设置工单进度缓存
func (c *CacheService) SetWorkOrderProgress(workOrderID int64, progress int) {
	cacheKey := fmt.Sprintf("work_order_progress:%d", workOrderID)
	c.Set(cacheKey, progress, 5*time.Minute) // 缓存5分钟
}

// InvalidateWorkOrderProgress 使工单进度缓存失效
func (c *CacheService) InvalidateWorkOrderProgress(workOrderID int64) {
	cacheKey := fmt.Sprintf("work_order_progress:%d", workOrderID)
	c.Delete(cacheKey)
}

// GetMoldLock 获取模具锁（简单实现，生产环境建议使用Redis分布式锁）
func (c *CacheService) GetMoldLock(moldID int64, ttl time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("mold_lock:%d", moldID)

	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[lockKey]; exists {
		if time.Now().Before(item.ExpiredAt) {
			return false, nil // 锁已被占用
		}
	}

	c.items[lockKey] = &CacheItem{
		Value:     true,
		ExpiredAt: time.Now().Add(ttl),
	}

	return true, nil
}

// ReleaseMoldLock 释放模具锁
func (c *CacheService) ReleaseMoldLock(moldID int64) {
	lockKey := fmt.Sprintf("mold_lock:%d", moldID)
	c.Delete(lockKey)
}

// cleanup 定期清理过期缓存项
func (c *CacheService) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiredAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

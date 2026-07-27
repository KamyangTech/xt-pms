package middleware

import (
	"sync"
	"time"
)

// pwdVerCacheEntry 密码版本号缓存条目
type pwdVerCacheEntry struct {
	Version int
	ExpireAt time.Time
}

// pwdVerCache 用户密码版本号内存缓存（TTL 滑动过期）
var pwdVerCache sync.Map

// pwdVerCacheTTL 缓存有效期：5 分钟（密码修改后最多 5 分钟内生效，可接受）
const pwdVerCacheTTL = 5 * time.Minute

// GetPwdVersionFromCache 从缓存获取密码版本号，缓存命中返回 (version, true)，未命中返回 (0, false)
func GetPwdVersionFromCache(userID int64) (int, bool) {
	if val, ok := pwdVerCache.Load(userID); ok {
		entry := val.(pwdVerCacheEntry)
		if time.Now().Before(entry.ExpireAt) {
			return entry.Version, true // 缓存有效
		}
		pwdVerCache.Delete(userID) // 过期删除
	}
	return 0, false
}

// SetPwdVersionCache 写入密码版本号到缓存
func SetPwdVersionCache(userID int64, version int) {
	pwdVerCache.Store(userID, pwdVerCacheEntry{
		Version:  version,
		ExpireAt: time.Now().Add(pwdVerCacheTTL),
	})
}

// InvalidatePwdVersionCache 清除指定用户的密码版本号缓存（密码修改时调用）
func InvalidatePwdVersionCache(userID int64) {
	pwdVerCache.Delete(userID)
}

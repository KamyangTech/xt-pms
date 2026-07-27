package service

import (
	"sync"

	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

// SecurityConfigService 安全配置服务（带内存缓存）
type SecurityConfigService struct {
	dao       *dao.SysSecurityConfigDAO
	mu        sync.RWMutex
	cache     *entity.SysSecurityConfig // 内存缓存，避免频繁读DB
}

// 全局单例
var globalSecurityConfigService *SecurityConfigService

// InitSecurityConfigService 初始化安全配置服务（应用启动时调用）
func InitSecurityConfigService() *SecurityConfigService {
	if globalSecurityConfigService == nil {
		globalSecurityConfigService = &SecurityConfigService{
			dao: dao.NewSysSecurityConfigDAO(),
		}
		// 启动时加载一次缓存
		globalSecurityConfigService.loadCache()
	}
	return globalSecurityConfigService
}

func GetSecurityConfigService() *SecurityConfigService {
	return globalSecurityConfigService
}

// loadCache 从数据库加载到内存
func (s *SecurityConfigService) loadCache() {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg, err := s.dao.Get()
	if err != nil {
		// 表可能还不存在，使用默认值
		s.cache = &entity.SysSecurityConfig{
			TokenExpire:       config.GlobalConfig.JWT.Expire,
			MaxLoginAttempts:  5,
			LockDuration:      900,
			PasswordMinLen:    6,
			CookieSecure:      config.GlobalConfig.Server.CookieSecure,
			WebSocketSecure:   config.GlobalConfig.Server.CookieSecure, // 默认与 CookieSecure 一致
		}
		return
	}
	s.cache = cfg
	// 同步到 config 包的全局覆盖值（供 middleware 使用）
	if cfg.TokenExpire > 0 {
		config.SetJWTExpireOverride(cfg.TokenExpire)
	}
}

// GetTokenExpire 获取 Token 过期时间（秒），优先从缓存读取，fallback 到 config.yaml
func (s *SecurityConfigService) GetTokenExpire() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cache != nil && s.cache.TokenExpire > 0 {
		return s.cache.TokenExpire
	}
	return config.GlobalConfig.JWT.Expire
}

// GetCookieSecure 获取 Cookie Secure 标志
func (s *SecurityConfigService) GetCookieSecure() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cache != nil {
		return s.cache.CookieSecure
	}
	return config.GlobalConfig.Server.CookieSecure
}

// GetWebSocketSecure 获取 WebSocket Secure 标志
func (s *SecurityConfigService) GetWebSocketSecure() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.cache != nil {
		return s.cache.WebSocketSecure
	}
	return config.GlobalConfig.Server.CookieSecure
}

// Get 获取完整的安全配置
func (s *SecurityConfigService) Get() (*entity.SysSecurityConfig, error) {
	s.mu.RLock()
	if s.cache != nil {
		cfg := *s.cache // 值拷贝
		s.mu.RUnlock()
		return &cfg, nil
	}
	s.mu.RUnlock()

	// 缓存未命中，从数据库加载并更新缓存
	cfg, err := s.dao.Get()
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.cache = cfg
	s.mu.Unlock()
	return cfg, nil
}

// Update 更新安全配置（同时刷新内存缓存）
func (s *SecurityConfigService) Update(req *entity.SysSecurityConfig) error {
	// 参数校验
	if req.TokenExpire < 300 {
		req.TokenExpire = 300 // 最少5分钟
	}
	if req.TokenExpire > 86400 * 30 {
		req.TokenExpire = 86400 * 30 // 最长30天
	}
	if req.MaxLoginAttempts < 1 {
		req.MaxLoginAttempts = 1
	}
	if req.MaxLoginAttempts > 20 {
		req.MaxLoginAttempts = 20
	}
	if req.PasswordMinLen < 4 {
		req.PasswordMinLen = 4
	}
	if req.PasswordMinLen > 32 {
		req.PasswordMinLen = 32
	}

	existing, _ := s.dao.Get()
	if existing == nil || existing.ID == 0 {
		req.ID = 0
	} else {
		req.ID = existing.ID
		req.CreateTime = existing.CreateTime
	}

	if err := s.dao.Save(req); err != nil {
		return err
	}

	// 更新缓存
	s.mu.Lock()
	// WebSocketSecure 变更时需要同步到全局（供后端 WebSocket 使用）
	if req.WebSocketSecure != s.cache.WebSocketSecure {
		s.cache.WebSocketSecure = req.WebSocketSecure
	}
	s.cache = req
	s.mu.Unlock()

	// 同步到 config 包的全局覆盖值（供 middleware 使用）
	if req.TokenExpire > 0 {
		config.SetJWTExpireOverride(req.TokenExpire)
	}

	return nil
}

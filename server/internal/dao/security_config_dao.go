package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SysSecurityConfigDAO struct{}

// NewSysSecurityConfigDAO 创建安全配置DAO实例
func NewSysSecurityConfigDAO() *SysSecurityConfigDAO {
	return &SysSecurityConfigDAO{}
}

// Get 获取安全配置（通常只有一条记录）
func (d *SysSecurityConfigDAO) Get() (*entity.SysSecurityConfig, error) {
	var cfg entity.SysSecurityConfig
	err := DB.First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save 保存安全配置
func (d *SysSecurityConfigDAO) Save(cfg *entity.SysSecurityConfig) error {
	return DB.Save(cfg).Error
}

package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SysSystemInfoDAO struct{}

// NewSysSystemInfoDAO 创建系统信息DAO实例
func NewSysSystemInfoDAO() *SysSystemInfoDAO {
	return &SysSystemInfoDAO{}
}

// Get 获取系统信息（通常只有一条记录）
func (d *SysSystemInfoDAO) Get() (*entity.SysSystemInfo, error) {
	var info entity.SysSystemInfo
	err := DB.First(&info).Error
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// Save 保存系统信息
func (d *SysSystemInfoDAO) Save(info *entity.SysSystemInfo) error {
	return DB.Save(info).Error
}

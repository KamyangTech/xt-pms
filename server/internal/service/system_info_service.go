package service

import (
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type SysSystemInfoService struct {
	systemInfoDAO *dao.SysSystemInfoDAO
}

// NewSysSystemInfoService 创建系统信息Service实例
func NewSysSystemInfoService() *SysSystemInfoService {
	return &SysSystemInfoService{
		systemInfoDAO: dao.NewSysSystemInfoDAO(),
	}
}

// Get 获取系统信息
func (s *SysSystemInfoService) Get() (*entity.SysSystemInfo, error) {
	return s.systemInfoDAO.Get()
}

// Save 保存系统信息（如果不存在则创建，存在则更新）
func (s *SysSystemInfoService) Save(info *entity.SysSystemInfo) error {
	existing, err := s.systemInfoDAO.Get()
	if err != nil || existing == nil {
		// 不存在，设置ID为0让GORM自增创建
		info.ID = 0
		return s.systemInfoDAO.Save(info)
	}
	// 存在，更新 - 保留原有ID
	info.ID = existing.ID
	info.CreateTime = existing.CreateTime
	return s.systemInfoDAO.Save(info)
}

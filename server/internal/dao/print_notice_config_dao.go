package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type PrintNoticeConfigDAO struct{}

func NewPrintNoticeConfigDAO() *PrintNoticeConfigDAO {
	return &PrintNoticeConfigDAO{}
}

// GetConfigByUserAndType 根据用户ID和配置类型获取配置
func (d *PrintNoticeConfigDAO) GetConfigByUserAndType(userID int64, configType string) (*entity.PrintNoticeConfig, error) {
	var config entity.PrintNoticeConfig
	err := DB.Where("user_id = ? AND config_type = ?", userID, configType).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetDefaultConfig 获取默认配置（user_id = 0）
func (d *PrintNoticeConfigDAO) GetDefaultConfig(configType string) (*entity.PrintNoticeConfig, error) {
	var config entity.PrintNoticeConfig
	err := DB.Where("user_id = 0 AND config_type = ? AND is_default = 1", configType).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// CreateConfig 创建配置
func (d *PrintNoticeConfigDAO) CreateConfig(config *entity.PrintNoticeConfig) error {
	return DB.Create(config).Error
}

// UpdateConfig 更新配置
func (d *PrintNoticeConfigDAO) UpdateConfig(config *entity.PrintNoticeConfig) error {
	return DB.Save(config).Error
}

// DeleteConfig 删除配置
func (d *PrintNoticeConfigDAO) DeleteConfig(id int64) error {
	return DB.Delete(&entity.PrintNoticeConfig{}, id).Error
}

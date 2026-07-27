package service

import (
	"encoding/json"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type PrintNoticeConfigService struct {
	configDAO *dao.PrintNoticeConfigDAO
}

func NewPrintNoticeConfigService() *PrintNoticeConfigService {
	return &PrintNoticeConfigService{
		configDAO: dao.NewPrintNoticeConfigDAO(),
	}
}

// GetConfig 获取用户的打印配置，如果没有则返回默认配置
func (s *PrintNoticeConfigService) GetConfig(userID int64, configType string) (interface{}, error) {
	logger.Debug("获取打印配置",
		zap.Int64("user_id", userID),
		zap.String("config_type", configType))

	// 先尝试获取用户配置
	config, err := s.configDAO.GetConfigByUserAndType(userID, configType)
	if err == nil && config != nil {
		logger.Debug("找到用户配置", zap.Int64("config_id", config.ID))
		result := parseConfigByType(config.ConfigData, configType)
		if result != nil {
			return result, nil
		}
	}

	// 如果用户没有配置，尝试获取默认配置
	logger.Debug("未找到用户配置，尝试获取默认配置")
	defaultConfig, err := s.configDAO.GetDefaultConfig(configType)
	if err == nil && defaultConfig != nil {
		logger.Debug("找到默认配置", zap.Int64("config_id", defaultConfig.ID))
		result := parseConfigByType(defaultConfig.ConfigData, configType)
		if result != nil {
			return result, nil
		}
	}

	// 如果都没有，返回硬编码的默认配置
	logger.Debug("未找到任何配置，返回硬编码默认值")
	return getDefaultConfigByType(configType), nil
}

// parseConfigByType 根据配置类型解析JSON配置
func parseConfigByType(configData map[string]interface{}, configType string) interface{} {
	jsonBytes, _ := json.Marshal(configData)
	switch configType {
	case "delivery_note":
		var data entity.DeliveryNoteConfigData
		if err := json.Unmarshal(jsonBytes, &data); err == nil {
			return &data
		}
	default:
		var data entity.PrintNoticeConfigData
		if err := json.Unmarshal(jsonBytes, &data); err == nil {
			return &data
		}
	}
	return nil
}

// getDefaultConfigByType 根据配置类型获取硬编码默认值
func getDefaultConfigByType(configType string) interface{} {
	switch configType {
	case "delivery_note":
		defaultData := entity.GetDefaultDeliveryNoteConfig()
		return &defaultData
	default:
		defaultData := entity.GetDefaultConfig()
		return &defaultData
	}
}

// SaveConfig 保存用户的打印配置
func (s *PrintNoticeConfigService) SaveConfig(userID int64, configType string, configData map[string]interface{}) error {
	logger.Debug("保存打印配置",
		zap.Int64("user_id", userID),
		zap.String("config_type", configType))

	existingConfig, err := s.configDAO.GetConfigByUserAndType(userID, configType)
	if err == nil && existingConfig != nil {
		logger.Debug("更新现有配置", zap.Int64("config_id", existingConfig.ID))
		existingConfig.ConfigData = configData
		return s.configDAO.UpdateConfig(existingConfig)
	}

	logger.Debug("创建新配置", zap.Int64("user_id", userID))
	newConfig := &entity.PrintNoticeConfig{
		UserID:     userID,
		ConfigType: configType,
		ConfigData: configData,
		IsDefault:  0,
	}
	return s.configDAO.CreateConfig(newConfig)
}

// ResetToDefault 重置为默认配置
func (s *PrintNoticeConfigService) ResetToDefault(userID int64, configType string) error {
	config, err := s.configDAO.GetConfigByUserAndType(userID, configType)
	if err == nil && config != nil {
		return s.configDAO.DeleteConfig(config.ID)
	}
	return nil
}

package utils

import "github.com/KamyangTech/xt-pms/server/config"

// IsProductionMode 判断当前是否为生产模式
func IsProductionMode() bool {
	if config.GlobalConfig == nil {
		return false
	}
	return config.GlobalConfig.App.Mode == "production"
}

// IsDevelopmentMode 判断当前是否为开发模式
func IsDevelopmentMode() bool {
	return !IsProductionMode()
}

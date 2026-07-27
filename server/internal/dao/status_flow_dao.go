package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

// StatusFlowConfigDAO 状态同步配置 DAO
type StatusFlowConfigDAO struct{}

func NewStatusFlowConfigDAO() *StatusFlowConfigDAO {
	return &StatusFlowConfigDAO{}
}

// FindByWorkOrderStatus 根据工单状态查询配置
func (d *StatusFlowConfigDAO) FindByWorkOrderStatus(workOrderStatus string) ([]*entity.StatusFlowConfig, error) {
	var configs []*entity.StatusFlowConfig
	err := DB.Where("work_order_status = ? AND is_active = ?", workOrderStatus, 1).
		Order("priority ASC").
		Find(&configs).Error
	return configs, err
}

// FindByMoldOrderStatus 根据订单状态查询配置
func (d *StatusFlowConfigDAO) FindByMoldOrderStatus(moldOrderStatus string) ([]*entity.StatusFlowConfig, error) {
	var configs []*entity.StatusFlowConfig
	err := DB.Where("mold_order_status = ? AND is_active = ?", moldOrderStatus, 1).
		Order("priority ASC").
		Find(&configs).Error
	return configs, err
}

// FindByMoldStatus 根据模具状态查询配置
func (d *StatusFlowConfigDAO) FindByMoldStatus(moldStatus string) ([]*entity.StatusFlowConfig, error) {
	var configs []*entity.StatusFlowConfig
	err := DB.Where("mold_status = ? AND is_active = ?", moldStatus, 1).
		Order("priority ASC").
		Find(&configs).Error
	return configs, err
}

// FindAllActive 查询所有启用的配置
func (d *StatusFlowConfigDAO) FindAllActive() ([]*entity.StatusFlowConfig, error) {
	var configs []*entity.StatusFlowConfig
	err := DB.Where("is_active = ?", 1).Order("priority ASC").Find(&configs).Error
	return configs, err
}

// FindByID 根据ID查询配置
func (d *StatusFlowConfigDAO) FindByID(id int64) (*entity.StatusFlowConfig, error) {
	var config entity.StatusFlowConfig
	err := DB.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

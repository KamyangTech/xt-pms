package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

// ProductionStatusFlowDAO 生产订单状态流转规则 DAO
type ProductionStatusFlowDAO struct{}

// NewProductionStatusFlowDAO 创建状态流转规则 DAO 实例
func NewProductionStatusFlowDAO() *ProductionStatusFlowDAO {
	return &ProductionStatusFlowDAO{}
}

// FindByFromStatusValue 根据起始状态值查询可用的流转规则
func (d *ProductionStatusFlowDAO) FindByFromStatusValue(fromStatusValue string) ([]*entity.ProductionStatusFlow, error) {
	var flows []*entity.ProductionStatusFlow
	err := DB.Where("from_status_value = ? AND is_enabled = ?", fromStatusValue, 1).
		Order("sort ASC").
		Find(&flows).Error
	return flows, err
}

// FindByFromStatusValues 批量查询多个起始状态值对应的可用流转规则（优化 N+1 问题）
func (d *ProductionStatusFlowDAO) FindByFromStatusValues(fromStatusValues []string) ([]*entity.ProductionStatusFlow, error) {
	if len(fromStatusValues) == 0 {
		return nil, nil
	}
	var flows []*entity.ProductionStatusFlow
	err := DB.Where("from_status_value IN ? AND is_enabled = ?", fromStatusValues, 1).
		Order("sort ASC").
		Find(&flows).Error
	return flows, err
}

// FindByActionCode 根据操作编码查询流转规则
func (d *ProductionStatusFlowDAO) FindByActionCode(actionCode string) (*entity.ProductionStatusFlow, error) {
	var flow entity.ProductionStatusFlow
	err := DB.Where("action_code = ? AND is_enabled = ?", actionCode, 1).First(&flow).Error
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

// ValidateFlow 验证状态流转是否合法
func (d *ProductionStatusFlowDAO) ValidateFlow(fromStatusValue, actionCode string) (*entity.ProductionStatusFlow, error) {
	var flow entity.ProductionStatusFlow
	err := DB.Where("from_status_value = ? AND action_code = ? AND is_enabled = ?",
		fromStatusValue, actionCode, 1).First(&flow).Error
	if err != nil {
		return nil, err
	}
	return &flow, nil
}

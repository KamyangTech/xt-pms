package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ProductionStatusFlowService struct {
	flowDAO *dao.ProductionStatusFlowDAO
	dictDAO *dao.SysDictDataDAO
}

func NewProductionStatusFlowService() *ProductionStatusFlowService {
	return &ProductionStatusFlowService{
		flowDAO: dao.NewProductionStatusFlowDAO(),
		dictDAO: dao.NewSysDictDataDAO(),
	}
}

// GetAvailableActions 获取当前状态下可用的操作列表
func (s *ProductionStatusFlowService) GetAvailableActions(currentStatusValue string) ([]*entity.ProductionStatusFlow, error) {
	return s.flowDAO.FindByFromStatusValue(currentStatusValue)
}

// ValidateAction 验证操作是否合法
func (s *ProductionStatusFlowService) ValidateAction(currentStatusValue, actionCode string) (*entity.ProductionStatusFlow, error) {
	flow, err := s.flowDAO.ValidateFlow(currentStatusValue, actionCode)
	if err != nil {
		return nil, errors.New("当前状态不允许执行此操作")
	}
	return flow, nil
}

// GetTargetStatusValue 根据操作编码获取目标状态值
func (s *ProductionStatusFlowService) GetTargetStatusValue(currentStatusValue, actionCode string) (string, error) {
	flow, err := s.ValidateAction(currentStatusValue, actionCode)
	if err != nil {
		return "", err
	}
	return flow.ToStatusValue, nil
}

// GetDictValue 根据字典编码和状态值获取字典项（用于验证状态是否存在）
func (s *ProductionStatusFlowService) GetDictValue(dictCode, statusValue string) (string, error) {
	dict, err := s.dictDAO.FindByTypeCodeAndValue(dictCode, statusValue)
	if err != nil {
		return "", errors.New("状态配置不存在")
	}
	if dict == nil {
		return "", errors.New("状态配置不存在")
	}
	return dict.DictValue, nil
}

package service

import (
	"errors"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type MoldProductRelationService struct {
	relationDAO *dao.MoldProductRelationDAO
}

func NewMoldProductRelationService() *MoldProductRelationService {
	return &MoldProductRelationService{
		relationDAO: dao.NewMoldProductRelationDAO(),
	}
}

// CreateRelation 创建模具产品关联
func (s *MoldProductRelationService) CreateRelation(tx *gorm.DB, relation *entity.MoldProductRelation) error {
	if tx == nil {
		tx = dao.GetDB()
	}

	// 检查是否已存在相同关联（如果 production_order_item_id 存在）
	if relation.ProductionOrderItemID != nil {
		existing, err := s.relationDAO.FindByMoldAndItem(relation.MoldID, *relation.ProductionOrderItemID)
		if existing != nil && err == nil {
			return errors.New("该模具与生产明细的关联已存在")
		}
	}

	return s.relationDAO.Create(tx, relation)
}

// CreateRelationsBatch 批量创建模具产品关联
func (s *MoldProductRelationService) CreateRelationsBatch(tx *gorm.DB, relations []*entity.MoldProductRelation) error {
	if tx == nil {
		tx = dao.GetDB()
	}

	if len(relations) == 0 {
		return nil
	}

	return s.relationDAO.CreateBatch(tx, relations)
}

// GetRelationsByMoldID 获取模具关联的所有产品
func (s *MoldProductRelationService) GetRelationsByMoldID(moldID int64) ([]entity.MoldProductRelation, error) {
	return s.relationDAO.FindByMoldID(moldID)
}

// GetRelationsByProductID 获取产品关联的所有模具
func (s *MoldProductRelationService) GetRelationsByProductID(productID int64) ([]entity.MoldProductRelation, error) {
	return s.relationDAO.FindByProductID(productID)
}

// GetRelationsByItemID 获取生产订单明细关联的模具
func (s *MoldProductRelationService) GetRelationsByItemID(itemID int64) ([]entity.MoldProductRelation, error) {
	return s.relationDAO.FindByProductionOrderItemID(itemID)
}

// DeleteRelationsByItemID 删除生产订单明细的所有关联
func (s *MoldProductRelationService) DeleteRelationsByItemID(tx *gorm.DB, itemID int64) error {
	if tx == nil {
		tx = dao.GetDB()
	}
	return s.relationDAO.DeleteByProductionOrderItemID(itemID)
}

// DeleteRelationByID 根据ID删除模具产品关联
func (s *MoldProductRelationService) DeleteRelationByID(id int64) error {
	return s.relationDAO.DeleteByID(id)
}

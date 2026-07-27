package dao

import (
	"gorm.io/gorm"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type MoldProductRelationDAO struct{}

// NewMoldProductRelationDAO 创建模具产品关联DAO实例
func NewMoldProductRelationDAO() *MoldProductRelationDAO {
	return &MoldProductRelationDAO{}
}

// Create 创建模具产品关联
func (d *MoldProductRelationDAO) Create(tx *gorm.DB, relation *entity.MoldProductRelation) error {
	if tx == nil {
		tx = GetDB()
	}
	return tx.Create(relation).Error
}

// CreateBatch 批量创建模具产品关联
func (d *MoldProductRelationDAO) CreateBatch(tx *gorm.DB, relations []*entity.MoldProductRelation) error {
	if tx == nil {
		tx = GetDB()
	}
	return tx.CreateInBatches(relations, 100).Error
}

// FindByID 根据ID查询关联
func (d *MoldProductRelationDAO) FindByID(id int64) (*entity.MoldProductRelation, error) {
	var relation entity.MoldProductRelation
	err := GetDB().Preload("Mold").Preload("Product").Preload("ProductionOrderItem").First(&relation, id).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// FindByMoldID 根据模具ID查询所有关联的产品
func (d *MoldProductRelationDAO) FindByMoldID(moldID int64) ([]entity.MoldProductRelation, error) {
	var relations []entity.MoldProductRelation
	err := GetDB().Preload("Product").Preload("ProductionOrderItem").Where("mold_id = ?", moldID).Find(&relations).Error
	return relations, err
}

// FindByProductID 根据产品ID查询所有关联的模具
func (d *MoldProductRelationDAO) FindByProductID(productID int64) ([]entity.MoldProductRelation, error) {
	var relations []entity.MoldProductRelation
	err := GetDB().Preload("Mold").Where("product_id = ?", productID).Find(&relations).Error
	return relations, err
}

// FindByProductionOrderItemID 根据生产订单明细ID查询关联
func (d *MoldProductRelationDAO) FindByProductionOrderItemID(itemID int64) ([]entity.MoldProductRelation, error) {
	var relations []entity.MoldProductRelation
	err := GetDB().Preload("Mold").Preload("Product").Where("production_order_item_id = ?", itemID).Find(&relations).Error
	return relations, err
}

// FindByMoldAndItem 根据模具ID和生产订单明细ID查询关联
func (d *MoldProductRelationDAO) FindByMoldAndItem(moldID, itemID int64) (*entity.MoldProductRelation, error) {
	var relation entity.MoldProductRelation
	err := GetDB().Where("mold_id = ? AND production_order_item_id = ?", moldID, itemID).First(&relation).Error
	if err != nil {
		return nil, err
	}
	return &relation, nil
}

// Update 更新关联
func (d *MoldProductRelationDAO) Update(relation *entity.MoldProductRelation) error {
	return GetDB().Save(relation).Error
}

// Delete 删除关联
func (d *MoldProductRelationDAO) Delete(id int64) error {
	return GetDB().Delete(&entity.MoldProductRelation{}, id).Error
}

// DeleteByMoldID 删除模具的所有关联
func (d *MoldProductRelationDAO) DeleteByMoldID(moldID int64) error {
	return GetDB().Where("mold_id = ?", moldID).Delete(&entity.MoldProductRelation{}).Error
}

// DeleteByProductionOrderItemID 删除生产订单明细的所有关联
func (d *MoldProductRelationDAO) DeleteByProductionOrderItemID(itemID int64) error {
	return GetDB().Where("production_order_item_id = ?", itemID).Delete(&entity.MoldProductRelation{}).Error
}

// DeleteByID 根据ID删除模具产品关联
func (d *MoldProductRelationDAO) DeleteByID(id int64) error {
	return GetDB().Delete(&entity.MoldProductRelation{}, id).Error
}

// GetProductCountByMoldID 获取模具关联的产品数量
func (d *MoldProductRelationDAO) GetProductCountByMoldID(moldID int64) (int64, error) {
	var count int64
	err := GetDB().Model(&entity.MoldProductRelation{}).Where("mold_id = ?", moldID).Count(&count).Error
	return count, err
}

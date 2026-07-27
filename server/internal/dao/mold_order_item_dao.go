package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type MoldOrderItemDAO struct{}

// NewMoldOrderItemDAO 创建模具订单明细 DAO 实例
func NewMoldOrderItemDAO() *MoldOrderItemDAO {
	return &MoldOrderItemDAO{}
}

// Create 创建订单明细
func (d *MoldOrderItemDAO) Create(item *entity.MoldOrderItem) error {
	return DB.Create(item).Error
}

// CreateInTx 在事务中创建订单明细
func (d *MoldOrderItemDAO) CreateInTx(tx *gorm.DB, item *entity.MoldOrderItem) error {
	return tx.Create(item).Error
}

// FindByOrderID 根据订单ID查询所有明细
func (d *MoldOrderItemDAO) FindByOrderID(orderID int64) ([]*entity.MoldOrderItem, error) {
	var items []*entity.MoldOrderItem
	err := DB.Preload("Mold").Where("order_id = ?", orderID).Order("sort ASC").Find(&items).Error
	return items, err
}

// FindByOrderIDs 批量查询多个订单的明细（避免N+1）
func (d *MoldOrderItemDAO) FindByOrderIDs(orderIDs []int64) ([]*entity.MoldOrderItem, error) {
	if len(orderIDs) == 0 {
		return []*entity.MoldOrderItem{}, nil
	}
	var items []*entity.MoldOrderItem
	err := DB.Preload("Mold").Where("order_id IN ?", orderIDs).Order("order_id ASC, sort ASC").Find(&items).Error
	return items, err
}

// DeleteByOrderID 删除订单的所有明细
func (d *MoldOrderItemDAO) DeleteByOrderID(orderID int64) error {
	return DB.Where("order_id = ?", orderID).Delete(&entity.MoldOrderItem{}).Error
}

// Update 更新订单明细
func (d *MoldOrderItemDAO) Update(item *entity.MoldOrderItem) error {
	return DB.Save(item).Error
}

// Delete 删除订单明细
func (d *MoldOrderItemDAO) Delete(id int64) error {
	return DB.Delete(&entity.MoldOrderItem{}, id).Error
}

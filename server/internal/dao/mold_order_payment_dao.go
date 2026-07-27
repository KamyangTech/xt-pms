package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type MoldOrderPaymentDAO struct{}

// NewMoldOrderPaymentDAO 创建模具订单款项 DAO 实例
func NewMoldOrderPaymentDAO() *MoldOrderPaymentDAO {
	return &MoldOrderPaymentDAO{}
}

// Create 创建款项记录
func (d *MoldOrderPaymentDAO) Create(payment *entity.MoldOrderPayment) error {
	return DB.Create(payment).Error
}

// FindByID 根据 ID 查询款项
func (d *MoldOrderPaymentDAO) FindByID(id int64) (*entity.MoldOrderPayment, error) {
	var payment entity.MoldOrderPayment
	err := DB.First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// FindByOrderID 根据订单 ID 查询所有款项
func (d *MoldOrderPaymentDAO) FindByOrderID(orderID int64) ([]*entity.MoldOrderPayment, error) {
	var payments []*entity.MoldOrderPayment
	err := DB.Where("order_id = ?", orderID).Order("id ASC").Find(&payments).Error
	return payments, err
}

// FindByWorkOrderID 根据工单 ID 查询所有款项
func (d *MoldOrderPaymentDAO) FindByWorkOrderID(workOrderID int64) ([]*entity.MoldOrderPayment, error) {
	var payments []*entity.MoldOrderPayment
	err := DB.Where("work_order_id = ?", workOrderID).Order("id ASC").Find(&payments).Error
	return payments, err
}

// Update 更新款项记录
func (d *MoldOrderPaymentDAO) Update(payment *entity.MoldOrderPayment) error {
	return DB.Save(payment).Error
}

// Delete 删除款项记录
func (d *MoldOrderPaymentDAO) Delete(id int64) error {
	return DB.Delete(&entity.MoldOrderPayment{}, id).Error
}

// DeleteByOrderID 删除订单的所有款项记录
func (d *MoldOrderPaymentDAO) DeleteByOrderID(orderID int64) error {
	return DB.Where("order_id = ?", orderID).Delete(&entity.MoldOrderPayment{}).Error
}

// DeleteByWorkOrderID 删除工单的所有款项记录
func (d *MoldOrderPaymentDAO) DeleteByWorkOrderID(workOrderID int64) error {
	return DB.Where("work_order_id = ?", workOrderID).Delete(&entity.MoldOrderPayment{}).Error
}

// DeleteByWorkOrderIDWithTx 在事务中删除工单的所有款项记录
func (d *MoldOrderPaymentDAO) DeleteByWorkOrderIDWithTx(tx *gorm.DB, workOrderID int64) error {
	return tx.Where("work_order_id = ?", workOrderID).Delete(&entity.MoldOrderPayment{}).Error
}

// CreateWithTx 在事务中创建款项记录
func (d *MoldOrderPaymentDAO) CreateWithTx(tx *gorm.DB, payment *entity.MoldOrderPayment) error {
	return tx.Create(payment).Error
}

// GetPaymentStats 获取订单款项统计
func (d *MoldOrderPaymentDAO) GetPaymentStats(orderID int64) (map[string]interface{}, error) {
	var stats struct {
		ContractTotal float64 `gorm:"column:contract_total"`
		ActualTotal   float64 `gorm:"column:actual_total"`
		ShouldTotal   float64 `gorm:"column:should_total"`
	}

	err := DB.Model(&entity.MoldOrderPayment{}).
		Select("COALESCE(SUM(contract_amount), 0) as contract_total, COALESCE(SUM(actual_amount), 0) as actual_total, COALESCE(SUM(should_amount), 0) as should_total").
		Where("order_id = ?", orderID).
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"contract_total": stats.ContractTotal,
		"actual_total":   stats.ActualTotal,
		"should_total":   stats.ShouldTotal,
		"pending_total":  stats.ShouldTotal - stats.ActualTotal,
	}

	return result, nil
}

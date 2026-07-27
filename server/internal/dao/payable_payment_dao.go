package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type PayablePaymentDAO struct{}

// NewPayablePaymentDAO 创建付款记录DAO实例
func NewPayablePaymentDAO() *PayablePaymentDAO {
	return &PayablePaymentDAO{}
}

// Create 创建付款记录
func (d *PayablePaymentDAO) Create(payment *entity.PayablePayment) error {
	return DB.Create(payment).Error
}

// FindByPayableID 查询应付账款的所有付款记录
func (d *PayablePaymentDAO) FindByPayableID(payableID int64) ([]*entity.PayablePayment, error) {
	var payments []*entity.PayablePayment
	err := DB.Where("payable_id = ?", payableID).
		Order("payment_date DESC, id DESC").
		Find(&payments).Error
	return payments, err
}

// GetTotalPaid 获取应付账款的已付总额
func (d *PayablePaymentDAO) GetTotalPaid(payableID int64) (float64, error) {
	var total float64
	err := DB.Model(&entity.PayablePayment{}).
		Where("payable_id = ?", payableID).
		Select("COALESCE(SUM(payment_amount), 0)").
		Scan(&total).Error
	return total, err
}

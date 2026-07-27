package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ReceivablePaymentDAO struct{}

// NewReceivablePaymentDAO 创建收款记录DAO实例
func NewReceivablePaymentDAO() *ReceivablePaymentDAO {
	return &ReceivablePaymentDAO{}
}

// Create 创建收款记录
func (d *ReceivablePaymentDAO) Create(payment *entity.ReceivablePayment) error {
	return DB.Create(payment).Error
}

// FindByID 根据ID查询收款记录
func (d *ReceivablePaymentDAO) FindByID(id int64) (*entity.ReceivablePayment, error) {
	var payment entity.ReceivablePayment
	err := DB.First(&payment, id).Error
	return &payment, err
}

// FindByReceivableID 查询应收账款的所有收款记录
func (d *ReceivablePaymentDAO) FindByReceivableID(receivableID int64) ([]*entity.ReceivablePayment, error) {
	var payments []*entity.ReceivablePayment
	err := DB.Where("receivable_id = ?", receivableID).
		Order("payment_date DESC, id DESC").
		Find(&payments).Error
	return payments, err
}

// Update 更新收款记录
func (d *ReceivablePaymentDAO) Update(payment *entity.ReceivablePayment) error {
	return DB.Save(payment).Error
}

// GetTotalReceived 获取应收账款的已收总额
func (d *ReceivablePaymentDAO) GetTotalReceived(receivableID int64) (float64, error) {
	var total float64
	err := DB.Model(&entity.ReceivablePayment{}).
		Where("receivable_id = ?", receivableID).
		Select("COALESCE(SUM(payment_amount), 0)").
		Scan(&total).Error
	return total, err
}

package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type ExpensePaymentDAO struct{}

// NewExpensePaymentDAO 创建费用支付记录DAO实例
func NewExpensePaymentDAO() *ExpensePaymentDAO {
	return &ExpensePaymentDAO{}
}

// Create 创建支付记录
func (d *ExpensePaymentDAO) Create(payment *entity.ExpensePayment) error {
	return DB.Create(payment).Error
}

// FindByExpenseID 查询费用的所有支付记录
func (d *ExpensePaymentDAO) FindByExpenseID(expenseID int64) ([]*entity.ExpensePayment, error) {
	var payments []*entity.ExpensePayment
	err := DB.Where("expense_id = ?", expenseID).
		Order("payment_date DESC, id DESC").
		Find(&payments).Error
	return payments, err
}

// GetTotalPaid 获取费用的已付总额
func (d *ExpensePaymentDAO) GetTotalPaid(expenseID int64) (float64, error) {
	var total float64
	err := DB.Model(&entity.ExpensePayment{}).
		Where("expense_id = ?", expenseID).
		Select("COALESCE(SUM(payment_amount), 0)").
		Scan(&total).Error
	return total, err
}

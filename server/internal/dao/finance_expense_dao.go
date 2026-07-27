package dao

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type FinanceExpenseDAO struct{}

// NewFinanceExpenseDAO 创建财务费用DAO实例
func NewFinanceExpenseDAO() *FinanceExpenseDAO {
	return &FinanceExpenseDAO{}
}

// Create 创建费用记录
func (d *FinanceExpenseDAO) Create(expense *entity.FinanceExpense) error {
	return DB.Create(expense).Error
}

// FindByID 根据ID查询费用
func (d *FinanceExpenseDAO) FindByID(id int64) (*entity.FinanceExpense, error) {
	var expense entity.FinanceExpense
	err := DB.First(&expense, id).Error
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

// FindList 分页查询费用列表
func (d *FinanceExpenseDAO) FindList(page, pageSize int, expenseType, relatedOrderType string, relatedOrderID *int64, startDate, endDate string) ([]*entity.FinanceExpense, int64, error) {
	var expenses []*entity.FinanceExpense
	var total int64

	query := DB.Model(&entity.FinanceExpense{})

	// 应用过滤条件
	if expenseType != "" {
		query = query.Where("expense_type = ?", expenseType)
	}
	if relatedOrderID != nil && *relatedOrderID > 0 {
		query = query.Where("related_order_id = ?", *relatedOrderID)
		if relatedOrderType != "" {
			query = query.Where("related_order_type = ?", relatedOrderType)
		}
	}
	if startDate != "" {
		query = query.Where("create_time >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("create_time <= ?", endDate+" 23:59:59")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("create_time DESC").Find(&expenses).Error

	return expenses, total, err
}

// FindByDeliveryID 根据交付记录ID查询费用
func (d *FinanceExpenseDAO) FindByDeliveryID(deliveryID int64) ([]*entity.FinanceExpense, error) {
	var expenses []*entity.FinanceExpense
	err := DB.Where("delivery_id = ?", deliveryID).Find(&expenses).Error
	return expenses, err
}

// Update 更新费用
func (d *FinanceExpenseDAO) Update(expense *entity.FinanceExpense) error {
	return DB.Save(expense).Error
}

// Delete 删除费用
func (d *FinanceExpenseDAO) Delete(id int64) error {
	return DB.Delete(&entity.FinanceExpense{}, id).Error
}

// PayExpense 支付费用（旧方法，仅更新金额）
func (d *FinanceExpenseDAO) PayExpense(id int64, amount float64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var expense entity.FinanceExpense
		if err := tx.First(&expense, id).Error; err != nil {
			return err
		}

		expense.PaidAmount += amount
		return tx.Save(&expense).Error
	})
}

// PayWithDetails 支付费用并创建详细支付记录
func (d *FinanceExpenseDAO) PayWithDetails(id int64, amount float64, paymentDate time.Time, paymentMethod, transactionNo, paymentAccount, receiverAccount, remark string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var expense entity.FinanceExpense
		if err := tx.First(&expense, id).Error; err != nil {
			return err
		}

		// 1. 创建支付记录
		payment := &entity.ExpensePayment{
			ExpenseID:       id,
			PaymentAmount:   amount,
			PaymentDate:     paymentDate,
			PaymentMethod:   paymentMethod,
			TransactionNo:   transactionNo,
			PaymentAccount:  paymentAccount,
			ReceiverAccount: receiverAccount,
			Remark:          remark,
		}
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		// 2. 更新费用已付金额
		expense.PaidAmount += amount
		return tx.Save(&expense).Error
	})
}

// GetSummaryStats 获取汇总统计
func (d *FinanceExpenseDAO) GetSummaryStats() (totalAmount, totalPaid, totalBalance float64, err error) {
	var result struct {
		TotalAmount     float64
		TotalPaidAmount float64
	}
	err = DB.Model(&entity.FinanceExpense{}).
		Select("COALESCE(SUM(amount), 0) AS total_amount, COALESCE(SUM(paid_amount), 0) AS total_paid_amount").
		Scan(&result).Error
	if err != nil {
		return 0, 0, 0, err
	}
	return result.TotalAmount, result.TotalPaidAmount, result.TotalAmount - result.TotalPaidAmount, nil
}

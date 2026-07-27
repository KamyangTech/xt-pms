package service

import (
	"fmt"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type FinanceExpenseService struct {
	expenseDAO  *dao.FinanceExpenseDAO
	paymentDAO  *dao.ExpensePaymentDAO
}

func NewFinanceExpenseService() *FinanceExpenseService {
	return &FinanceExpenseService{
		expenseDAO: dao.NewFinanceExpenseDAO(),
		paymentDAO: dao.NewExpensePaymentDAO(),
	}
}

// CreateExpense 创建费用记录
func (s *FinanceExpenseService) CreateExpense(expense *entity.FinanceExpense) error {
	if expense.Amount <= 0 {
		return fmt.Errorf("费用金额必须大于0")
	}
	return s.expenseDAO.Create(expense)
}

// enrichExpense 填充费用计算字段（余额、状态）
func enrichExpense(e *entity.FinanceExpense) {
	e.Balance = e.Amount - e.PaidAmount
	switch {
	case e.PaidAmount <= 0:
		e.Status = "pending"
	case e.PaidAmount >= e.Amount:
		e.Status = "paid"
	default:
		e.Status = "partial"
	}
}

// GetExpenseByID 根据ID获取费用
func (s *FinanceExpenseService) GetExpenseByID(id int64) (*entity.FinanceExpense, error) {
	expense, err := s.expenseDAO.FindByID(id)
	if err != nil || expense == nil {
		return expense, err
	}
	enrichExpense(expense)
	return expense, nil
}

// GetExpenseList 获取费用列表
func (s *FinanceExpenseService) GetExpenseList(page, pageSize int, expenseType string, relatedOrderID *int64, relatedOrderType, startDate, endDate string) ([]*entity.FinanceExpense, int64, error) {
	expenses, total, err := s.expenseDAO.FindList(page, pageSize, expenseType, relatedOrderType, relatedOrderID, startDate, endDate)
	if err != nil {
		return nil, 0, err
	}
	for _, e := range expenses {
		enrichExpense(e)
	}
	return expenses, total, nil
}

// GetExpensesByDelivery 根据交付记录ID获取费用列表
func (s *FinanceExpenseService) GetExpensesByDelivery(deliveryID int64) ([]*entity.FinanceExpense, error) {
	expenses, err := s.expenseDAO.FindByDeliveryID(deliveryID)
	if err != nil {
		return nil, err
	}
	for _, e := range expenses {
		enrichExpense(e)
	}
	return expenses, nil
}

// UpdateExpense 更新费用
func (s *FinanceExpenseService) UpdateExpense(expense *entity.FinanceExpense) error {
	if expense.ID == 0 {
		return fmt.Errorf("费用ID不能为空")
	}
	return s.expenseDAO.Update(expense)
}

// DeleteExpense 删除费用
func (s *FinanceExpenseService) DeleteExpense(id int64) error {
	return s.expenseDAO.Delete(id)
}

// PayExpense 支付费用（含详细记录）
func (s *FinanceExpenseService) PayExpense(id int64, amount float64, paymentDate time.Time, paymentMethod, transactionNo, paymentAccount, receiverAccount, remark string) error {
	if amount <= 0 {
		return fmt.Errorf("支付金额必须大于0")
	}

	expense, err := s.expenseDAO.FindByID(id)
	if err != nil {
		return fmt.Errorf("费用记录不存在")
	}
	balance := expense.Amount - expense.PaidAmount
	if amount > balance {
		return fmt.Errorf("支付金额不能超过余额 %.2f", balance)
	}

	return s.expenseDAO.PayWithDetails(id, amount, paymentDate, paymentMethod, transactionNo, paymentAccount, receiverAccount, remark)
}

// GetPaymentsByExpenseID 获取费用的支付记录列表
func (s *FinanceExpenseService) GetPaymentsByExpenseID(expenseID int64) ([]*entity.ExpensePayment, error) {
	return s.paymentDAO.FindByExpenseID(expenseID)
}

// GetSummaryStats 获取汇总统计
func (s *FinanceExpenseService) GetSummaryStats() (totalAmount, totalPaid, totalBalance float64, err error) {
	return s.expenseDAO.GetSummaryStats()
}

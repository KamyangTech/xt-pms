package service

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
)

type AccountsPayableService struct {
	apDAO *dao.AccountsPayableDAO
}

func NewAccountsPayableService() *AccountsPayableService {
	return &AccountsPayableService{
		apDAO: dao.NewAccountsPayableDAO(),
	}
}

// CreatePayable 创建应付账款
func (s *AccountsPayableService) CreatePayable(ap *entity.AccountsPayable) error {
	// 参数验证
	if ap.SupplierID == 0 {
		return fmt.Errorf("供应商ID不能为空")
	}
	if ap.Amount <= 0 {
		return fmt.Errorf("应付金额必须大于0")
	}

	return s.apDAO.Create(ap)
}

// GetPayableByID 根据ID获取应付账款
func (s *AccountsPayableService) GetPayableByID(id int64) (*entity.AccountsPayable, error) {
	return s.apDAO.FindByID(id)
}

// GetPayableList 获取应付账款列表
func (s *AccountsPayableService) GetPayableList(page, pageSize int, supplierID *int64, status string) ([]*entity.AccountsPayable, int64, error) {
	return s.apDAO.FindList(page, pageSize, supplierID, status)
}

// GetPayableByOrder 根据订单获取应付账款
func (s *AccountsPayableService) GetPayableByOrder(orderID int64, orderType string) (*entity.AccountsPayable, error) {
	return s.apDAO.FindByOrderID(orderID, orderType)
}

// UpdatePayable 更新应付账款
func (s *AccountsPayableService) UpdatePayable(ap *entity.AccountsPayable) error {
	if ap.ID == 0 {
		return fmt.Errorf("应付账款ID不能为空")
	}

	// 只有待付款状态的应付账款允许修改
	existing, err := s.apDAO.FindByID(ap.ID)
	if err != nil {
		return err
	}
	if existing.Status != "PENDING" {
		return fmt.Errorf("当前状态(%s)不允许修改，仅待付款状态可修改", existing.Status)
	}

	return s.apDAO.Update(ap)
}

// DeletePayable 删除应付账款
func (s *AccountsPayableService) DeletePayable(id int64) error {
	return s.apDAO.Delete(id)
}

// GetPaymentsByPayableID 获取应付账款的付款记录列表
func (s *AccountsPayableService) GetPaymentsByPayableID(payableID int64) ([]*entity.PayablePayment, error) {
	paymentDAO := dao.NewPayablePaymentDAO()

	payments, err := paymentDAO.FindByPayableID(payableID)
	if err != nil {
		return nil, err
	}

	for _, p := range payments {
		if p.ReceiptFileID != "" {
			p.ReceiptFiles = loadPayableFiles(p.ReceiptFileID)
		}
		if p.InvoiceFileID != "" {
			p.InvoiceFiles = loadPayableFiles(p.InvoiceFileID)
		}
	}

	return payments, nil
}

func loadPayableFiles(fileIDs string) []*entity.File {
	fileDAO := dao.NewFileDAO()
	ids := strings.Split(fileIDs, ",")
	var files []*entity.File
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if file, err := fileDAO.FindByID(id); err == nil {
			files = append(files, file)
		}
	}
	return files
}

// Pay 付款
func (s *AccountsPayableService) Pay(id int64, amount float64, discountAmount float64, paymentDate time.Time, paymentMethod, transactionNo, paymentAccount, receiverAccount, remark, receiptFileID, invoiceFileID string) error {
	if amount <= 0 && discountAmount <= 0 {
		return fmt.Errorf("付款金额和优惠金额不能同时为空")
	}
	if amount < 0 || discountAmount < 0 {
		return fmt.Errorf("付款金额和优惠金额不能为负数")
	}

	ap, err := s.apDAO.FindByID(id)
	if err != nil {
		return err
	}

	balance := ap.Amount - ap.PaidAmount - ap.DiscountAmount
	if amount+discountAmount > balance {
		return fmt.Errorf("付款金额+优惠金额(%.2f)不能超过余额 %.2f", amount+discountAmount, balance)
	}

	return s.apDAO.PayWithDetails(id, amount, discountAmount, paymentDate, paymentMethod, transactionNo, paymentAccount, receiverAccount, remark, receiptFileID, invoiceFileID)
}

// GetOverdueList 获取逾期应付账款列表
func (s *AccountsPayableService) GetOverdueList() ([]*entity.AccountsPayable, error) {
	return s.apDAO.GetOverdueList()
}

// ProcessOverduePayables 处理逾期应付账款（可定时任务调用）
func (s *AccountsPayableService) ProcessOverduePayables() error {
	overdueList, err := s.apDAO.GetOverdueList()
	if err != nil {
		return err
	}

	now := time.Now()
	updatedCount := 0
	for _, ap := range overdueList {
		if ap.DueDate != nil && ap.DueDate.Before(now) && ap.Status != "PAID" {
			days := int(now.Sub(ap.DueDate.Time).Hours() / 24)

			if ap.Status != "OVERDUE" {
				oldStatus := ap.Status
				ap.Status = "OVERDUE"

				if err := s.apDAO.Update(ap); err != nil {
					logger.Error("更新应付账款逾期状态失败",
						zap.Int64("id", ap.ID),
						zap.String("old_status", oldStatus),
						zap.Error(err),
					)
					continue
				}

				logger.Info("应付账款逾期状态更新",
					zap.Int64("id", ap.ID),
					zap.String("old_status", oldStatus),
					zap.Int("days_overdue", days),
				)
				updatedCount++
			}
		}
	}

	logger.Info("逾期应付账款处理完成",
		zap.Int("updated_count", updatedCount),
	)
	return nil
}

// CheckAndUpdateOverdue 检查并更新逾期状态（可定时任务调用）
func (s *AccountsPayableService) CheckAndUpdateOverdue() error {
	return s.ProcessOverduePayables()
}

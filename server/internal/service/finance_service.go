package service

import (
	"fmt"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type FinanceService struct {
	arDAO        *dao.AccountsReceivableDAO
	apDAO        *dao.AccountsPayableDAO
	outsourceDAO *dao.OutsourcingRecordDAO
}

func NewFinanceService() *FinanceService {
	return &FinanceService{
		arDAO:        dao.NewAccountsReceivableDAO(),
		apDAO:        dao.NewAccountsPayableDAO(),
		outsourceDAO: dao.NewOutsourcingRecordDAO(),
	}
}

type PaymentItem struct {
	PaymentType string
	Amount      float64
	DueDate     *dto.Date
	Remark      string
	SourceType  string // mold, production, outsource
}

// CreateReceivablesForOrder 为订单创建应收账款（支持防重复）
func (s *FinanceService) CreateReceivablesForOrder(customerID int64, orderID *int64, orderType string,
	paymentItems []PaymentItem) error {

	if len(paymentItems) == 0 {
		return nil
	}

	// 检查是否已存在应收账款（防止重复创建）
	if orderID != nil {
		existing, err := s.arDAO.FindByOrderID(*orderID, orderType)
		if err == nil && len(existing) > 0 {
			logger.Info("订单已存在应收账款，跳过创建",
				zap.Int64("order_id", *orderID),
				zap.String("order_type", orderType),
				zap.Int("count", len(existing)),
			)
			return nil
		}
	}

	for _, item := range paymentItems {
		if item.Amount <= 0 {
			continue
		}

		ar := &entity.AccountsReceivable{
			CustomerID:  customerID,
			OrderID:     orderID,
			OrderType:   orderType,
			PaymentType: item.PaymentType,
			Amount:      item.Amount,
			DueDate:     item.DueDate,
			Status:      "PENDING",
			SourceType:  item.SourceType,
			Remark:      item.Remark,
		}

		if err := s.arDAO.Create(ar); err != nil {
			logger.Error("创建应收账款失败", zap.Error(err))
			return fmt.Errorf("创建应收账款失败: %w", err)
		}

		logger.Info("创建应收账款成功",
			zap.Int64("customer_id", customerID),
			zap.Float64("amount", item.Amount),
			zap.String("payment_type", item.PaymentType),
			zap.String("order_type", orderType),
		)
	}

	return nil
}

// CreatePayablesForOutsourcing 为外发记录创建应付账款（支持防重复）
func (s *FinanceService) CreatePayablesForOutsourcing(processorID int64, outsourcingRecordID int64,
	paymentItems []PaymentItem, sourceType string) error {

	if len(paymentItems) == 0 {
		return nil
	}

	// 检查是否已存在应付账款（防止重复创建）
	existing, err := s.apDAO.FindByOrderID(outsourcingRecordID, "outsourcing")
	if err == nil && existing != nil {
		logger.Info("外发记录已存在应付账款，跳过创建",
			zap.Int64("outsourcing_record_id", outsourcingRecordID),
		)
		return nil
	}

	for _, item := range paymentItems {
		if item.Amount <= 0 {
			continue
		}

		ap := &entity.AccountsPayable{
			SupplierID:     processorID,
			OrderID:        &outsourcingRecordID,
			OrderType:      "outsourcing",
			PaymentType:    item.PaymentType,
			SourceType:     sourceType,
			Amount:         item.Amount,
			DueDate:        item.DueDate,
			Status:  "PENDING",
			Remark:  item.Remark,
		}

		if err := s.apDAO.Create(ap); err != nil {
			logger.Error("创建应付账款失败", zap.Error(err))
			return fmt.Errorf("创建应付账款失败: %w", err)
		}

		logger.Info("创建应付账款成功",
			zap.Int64("processor_id", processorID),
			zap.Float64("amount", item.Amount),
			zap.String("payment_type", item.PaymentType),
			zap.String("source_type", sourceType),
		)
	}

	return nil
}

// GetOutsourcingRecordByID 获取外发记录（辅助方法）
func (s *FinanceService) GetOutsourcingRecordByID(id int64) (*entity.OutsourcingRecord, error) {
	return s.outsourceDAO.FindByID(id)
}

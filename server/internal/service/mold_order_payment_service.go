package service

import (
	"fmt"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"gorm.io/gorm"
)

type MoldOrderPaymentService struct {
	paymentDAO *dao.MoldOrderPaymentDAO
}

func NewMoldOrderPaymentService() *MoldOrderPaymentService {
	return &MoldOrderPaymentService{
		paymentDAO: dao.NewMoldOrderPaymentDAO(),
	}
}

// CreatePayment 创建款项记录
func (s *MoldOrderPaymentService) CreatePayment(payment *entity.MoldOrderPayment) error {
	if payment.OrderID == 0 {
		return fmt.Errorf("订单 ID 不能为空")
	}
	if payment.PaymentType == "" {
		return fmt.Errorf("款项类型不能为空")
	}

	return s.paymentDAO.Create(payment)
}

// GetPaymentsByOrderID 获取订单的所有款项
func (s *MoldOrderPaymentService) GetPaymentsByOrderID(orderID int64) ([]*entity.MoldOrderPayment, error) {
	return s.paymentDAO.FindByOrderID(orderID)
}

// GetPaymentsByWorkOrderID 根据工单ID获取款项列表
func (s *MoldOrderPaymentService) GetPaymentsByWorkOrderID(workOrderID int64) ([]*entity.MoldOrderPayment, error) {
	return s.paymentDAO.FindByWorkOrderID(workOrderID)
}

// GetPaymentByID 获取单个款项
func (s *MoldOrderPaymentService) GetPaymentByID(id int64) (*entity.MoldOrderPayment, error) {
	return s.paymentDAO.FindByID(id)
}

// UpdatePayment 更新款项记录
func (s *MoldOrderPaymentService) UpdatePayment(payment *entity.MoldOrderPayment) error {
	if payment.PaymentType == "" {
		return fmt.Errorf("款项类型不能为空")
	}

	return s.paymentDAO.Update(payment)
}

// DeletePayment 删除款项记录
func (s *MoldOrderPaymentService) DeletePayment(id int64) error {
	return s.paymentDAO.Delete(id)
}

// DeletePaymentsByOrderID 删除订单的所有款项记录
func (s *MoldOrderPaymentService) DeletePaymentsByOrderID(orderID int64) error {
	return s.paymentDAO.DeleteByOrderID(orderID)
}

// GetPaymentStats 获取订单款项统计
func (s *MoldOrderPaymentService) GetPaymentStats(orderID int64) (map[string]interface{}, error) {
	return s.paymentDAO.GetPaymentStats(orderID)
}

// ReplacePaymentsByWorkOrder 在事务中替换工单的所有款项记录（先删后建）
func (s *MoldOrderPaymentService) ReplacePaymentsByWorkOrder(orderID int64, workOrderID int64, items []entity.MoldOrderPayment) error {
	if workOrderID == 0 {
		return fmt.Errorf("工单ID不能为空")
	}

	db := dao.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 删除该工单的所有旧款项
		if err := s.paymentDAO.DeleteByWorkOrderIDWithTx(tx, workOrderID); err != nil {
			return err
		}

		// 2. 逐条创建新款项
		for i := range items {
			items[i].OrderID = orderID
			items[i].WorkOrderID = &workOrderID
			if err := s.paymentDAO.CreateWithTx(tx, &items[i]); err != nil {
				return err
			}
		}

		return nil
	})
}

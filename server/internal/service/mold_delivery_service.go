package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type MoldDeliveryService struct {
	moldDeliveryDAO  *dao.MoldDeliveryDAO
	moldWorkOrderDAO *dao.MoldWorkOrderDAO
}

func NewMoldDeliveryService() *MoldDeliveryService {
	return &MoldDeliveryService{
		moldDeliveryDAO:  dao.NewMoldDeliveryDAO(),
		moldWorkOrderDAO: dao.NewMoldWorkOrderDAO(),
	}
}

func (s *MoldDeliveryService) CreateMoldDelivery(delivery *entity.MoldDelivery) (*entity.MoldDelivery, error) {
	if delivery.DeliveryNo == "" {
		delivery.DeliveryNo = fmt.Sprintf("DLV-%d", time.Now().Unix())
	}

	err := s.moldDeliveryDAO.Create(delivery)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	err = s.moldWorkOrderDAO.Updates(delivery.MoldWorkID, map[string]any{
		"status":              constant.WorkOrderStatusDelivered,
		"actual_deliver_time": now,
		"update_time":         now,
	})
	if err != nil {
		return nil, fmt.Errorf("创建交付记录成功但更新工单状态失败: %v", err)
	}

	// 交付完成后，尝试解锁关联综合订单的尾款应收
	if delivery.MoldWorkID > 0 {
		if workOrder, err := s.moldWorkOrderDAO.FindByID(delivery.MoldWorkID); err == nil && workOrder != nil {
			if workOrder.MoldOrderID != nil {
				arService := NewAccountsReceivableService()
				if err := arService.UnlockFinalReceivablesByMasterOrder(*workOrder.MoldOrderID); err != nil {
					return nil, fmt.Errorf("创建交付记录成功但解锁尾款失败: %v", err)
				}
			}
		}
	}

	// ❌ 已删除：不再在交付时创建应收账款（改为订单创建时创建）
	// 应收账款现在在订单创建时自动生成

	return delivery, nil
}

func (s *MoldDeliveryService) GetMoldDeliveryByID(id int64) (*entity.MoldDelivery, error) {
	return s.moldDeliveryDAO.FindByID(id)
}

func (s *MoldDeliveryService) GetMoldDeliveriesByWorkOrderID(workOrderID int64) ([]*entity.MoldDelivery, error) {
	return s.moldDeliveryDAO.FindByWorkOrderID(workOrderID)
}

func (s *MoldDeliveryService) UpdateMoldDelivery(delivery *entity.MoldDelivery) error {
	original, err := s.moldDeliveryDAO.FindByID(delivery.ID)
	if err != nil {
		return errors.New("交付记录不存在")
	}
	delivery.CreateTime = original.CreateTime
	return s.moldDeliveryDAO.Update(delivery)
}

func (s *MoldDeliveryService) DeleteMoldDelivery(id int64) error {
	return s.moldDeliveryDAO.Delete(id)
}

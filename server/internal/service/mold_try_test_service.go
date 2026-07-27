package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
)

type MoldTryTestService struct {
	moldTryTestDAO    *dao.MoldTryTestDAO
	moldWorkOrderDAO  *dao.MoldWorkOrderDAO
	moldOrderDAO      *dao.MoldOrderDAO
	statusSyncService *MoldStatusSyncService
}

func NewMoldTryTestService() *MoldTryTestService {
	return &MoldTryTestService{
		moldTryTestDAO:    dao.NewMoldTryTestDAO(),
		moldWorkOrderDAO:  dao.NewMoldWorkOrderDAO(),
		moldOrderDAO:      dao.NewMoldOrderDAO(),
		statusSyncService: NewMoldStatusSyncService(),
	}
}

func (s *MoldTryTestService) CreateMoldTryTest(test *entity.MoldTryTest) (*entity.MoldTryTest, error) {
	// 🔒 强约束：互斥检查 - 获取当前工单信息
	workOrder, err := s.moldWorkOrderDAO.FindByID(test.MoldWorkID)
	if err != nil {
		return nil, fmt.Errorf("查找工单失败: %w", err)
	}

	// 注意：只有修模/改造工单（order_type=REPAIR或MODIFICATION）才会锁定其他工单
	// 如果当前工单本身就是修模/改造工单，则不锁定
	if workOrder.OrderType == constant.MoldOrderTypeRepair || workOrder.OrderType == constant.MoldOrderTypeModification {
		// 修模/改造工单不被锁定，继续执行后续逻辑
	} else {
		var activeWorkOrderCount int64
		dao.DB.Model(&entity.MoldWorkOrder{}).
			Where("mold_id = ? AND id != ? AND status >= ? AND status < ? AND is_cancelled = 0 AND (order_type = ? OR order_type = ?)",
				workOrder.MoldID, workOrder.ID, entity.WorkOrderStatusNotStarted, entity.WorkOrderStatusDelivered).
			Count(&activeWorkOrderCount)
		if activeWorkOrderCount > 0 {
			return nil, fmt.Errorf("该模具当前有修模/改造工单正在进行中，试模记录已锁定")
		}
	}

	// 先创建试模记录
	err = s.moldTryTestDAO.Create(test)
	if err != nil {
		return nil, err
	}

	// 自动更新工单状态为"已试模"
	// 只允许工单在"加工中"、"已完成"状态时进入试模阶段
	// 不允许在"未开始"状态添加试模记录（工序尚未开始）
	canEnterTestPhase := workOrder.Status == constant.WorkOrderStatusProcessing ||
		workOrder.Status == constant.WorkOrderStatusCompleted

	if !canEnterTestPhase {
		return nil, fmt.Errorf("当前工单状态不允许添加试模记录，请先开始工序")
	}

	// 使用语义标签获取"已试模"状态，而非硬编码
	testedStatus := s.statusSyncService.StatusFlowService.GetTestedWorkOrderStatus()
	if testedStatus != nil {
		workOrder.Status = *testedStatus
		if err := s.moldWorkOrderDAO.Update(workOrder); err != nil {
			return test, fmt.Errorf("更新工单状态失败: %w", err)
		}

		// 触发状态同步（同步到订单和模具）
		// 注意：先刷新缓存，确保使用最新的状态同步配置
		s.statusSyncService.StatusFlowService.RefreshCache()
		if err := s.statusSyncService.OnWorkOrderStatusChanged(workOrder); err != nil {
			return test, fmt.Errorf("同步状态失败: %w", err)
		}
	}

	return test, nil
}

func (s *MoldTryTestService) GetMoldTryTestByID(id int64) (*entity.MoldTryTest, error) {
	return s.moldTryTestDAO.FindByID(id)
}

func (s *MoldTryTestService) GetMoldTryTestsByWorkOrderID(workOrderID int64) ([]*entity.MoldTryTest, error) {
	return s.moldTryTestDAO.FindByWorkOrderID(workOrderID)
}

// GetMoldTryTestsByMoldID 根据模具ID获取试模记录列表（通过模具的最新工单查询）
func (s *MoldTryTestService) GetMoldTryTestsByMoldID(moldID int64) ([]*entity.MoldTryTest, error) {
	// 先查询模具获取最新工单ID
	moldDAO := dao.NewMoldDAO()
	mold, err := moldDAO.FindByID(moldID)
	if err != nil {
		return nil, fmt.Errorf("查找模具失败: %w", err)
	}

	if mold.LatestWorkOrderID == nil {
		// 模具没有关联工单，返回空列表
		return []*entity.MoldTryTest{}, nil
	}

	// 通过工单ID查询试模记录
	return s.moldTryTestDAO.FindByWorkOrderID(*mold.LatestWorkOrderID)
}

func (s *MoldTryTestService) UpdateMoldTryTest(test *entity.MoldTryTest) error {
	// 🔒 强约束：互斥检查 - 获取当前试模记录对应的工单
	original, err := s.moldTryTestDAO.FindByID(test.ID)
	if err != nil {
		return errors.New("试模记录不存在")
	}

	workOrder, err := s.moldWorkOrderDAO.FindByID(original.MoldWorkID)
	if err != nil {
		return fmt.Errorf("查找工单失败: %w", err)
	}

	// 注意：只有修模/改造工单（order_type=2或3）才会锁定
	// 如果当前工单本身就是修模/改造工单，则不锁定
	if workOrder.OrderType == constant.MoldOrderTypeRepair || workOrder.OrderType == constant.MoldOrderTypeModification {
		// 修模/改造工单不被锁定，继续执行后续逻辑
	} else {
		var activeWorkOrderCount int64
		dao.DB.Model(&entity.MoldWorkOrder{}).
			Where("mold_id = ? AND id != ? AND status >= ? AND status < ? AND is_cancelled = 0 AND (order_type = ? OR order_type = ?)",
				workOrder.MoldID, workOrder.ID, entity.WorkOrderStatusNotStarted, entity.WorkOrderStatusDelivered).
			Count(&activeWorkOrderCount)
		if activeWorkOrderCount > 0 {
			return fmt.Errorf("该模具当前有修模/改造工单正在进行中，试模记录已锁定")
		}
	}

	// 保留原始的 create_time，不覆盖为默认值
	test.CreateTime = original.CreateTime
	return s.moldTryTestDAO.Update(test)
}

func (s *MoldTryTestService) DeleteMoldTryTest(id int64) error {
	// 🔒 强约束：互斥检查 - 先获取试模记录，找到对应的工单ID
	test, err := s.moldTryTestDAO.FindByID(id)
	if err != nil {
		return errors.New("试模记录不存在")
	}

	workOrder, err := s.moldWorkOrderDAO.FindByID(test.MoldWorkID)
	if err != nil {
		return fmt.Errorf("查找工单失败: %w", err)
	}

	// 注意：只有修模/改造工单（order_type=2或3）才会锁定
	// 如果当前工单本身就是修模/改造工单，则不锁定
	if workOrder.OrderType == constant.MoldOrderTypeRepair || workOrder.OrderType == constant.MoldOrderTypeModification {
		// 修模/改造工单不被锁定，继续执行后续逻辑
	} else {
		var activeWorkOrderCount int64
		dao.DB.Model(&entity.MoldWorkOrder{}).
			Where("mold_id = ? AND id != ? AND status >= ? AND status < ? AND is_cancelled = 0 AND (order_type = ? OR order_type = ?)",
				workOrder.MoldID, workOrder.ID, entity.WorkOrderStatusNotStarted, entity.WorkOrderStatusDelivered).
			Count(&activeWorkOrderCount)
		if activeWorkOrderCount > 0 {
			return fmt.Errorf("该模具当前有修模/改造工单正在进行中，试模记录已锁定")
		}
	}

	if err := s.moldTryTestDAO.Delete(id); err != nil {
		return err
	}

	remainingTests, err := s.moldTryTestDAO.FindByWorkOrderID(test.MoldWorkID)
	if err != nil {
		return fmt.Errorf("查询工单试模记录失败: %w", err)
	}

	// 如果没有其他试模记录，且工单状态为"已试模"，则回滚到"已完成"
	if len(remainingTests) == 0 {
		workOrder, err := s.moldWorkOrderDAO.FindByID(test.MoldWorkID)
		if err != nil {
			return fmt.Errorf("查询工单失败: %w", err)
		}

		testedStatus := s.statusSyncService.StatusFlowService.GetTestedWorkOrderStatus()
		testedStatusValue := constant.WorkOrderStatusTested // 默认值
		if testedStatus != nil {
			testedStatusValue = *testedStatus
		}

		if workOrder.Status == testedStatusValue {
			// 回滚工单状态为"已完成"
			now := time.Now()

			completedStatus := s.statusSyncService.StatusFlowService.GetCompletedWorkOrderStatus()
			completedStatusValue := constant.WorkOrderStatusCompleted // 默认值
			if completedStatus != nil {
				completedStatusValue = *completedStatus
			}

			// 使用Select明确指定要更新的字段，包括nil值
			err = dao.DB.Model(&entity.MoldWorkOrder{}).
				Where("id = ?", test.MoldWorkID).
				Select("status", "update_time").
				Updates(map[string]interface{}{
					"status":      completedStatusValue,
					"update_time": now,
				}).Error
			if err != nil {
				return fmt.Errorf("删除成功但回滚工单状态失败: %w", err)
			}

			updatedWorkOrder, err := s.moldWorkOrderDAO.FindByID(test.MoldWorkID)
			if err != nil {
				return fmt.Errorf("删除成功但查询工单失败: %v", err)
			}

			// 触发状态同步（同步到订单和模具）
			if err := s.statusSyncService.OnWorkOrderStatusChanged(updatedWorkOrder); err != nil {
				return fmt.Errorf("删除成功但同步状态失败: %w", err)
			}
		}
	}

	return nil
}

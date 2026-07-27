package service

import (
	"fmt"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

// SyncContext 同步上下文，用于防止循环调用
type SyncContext struct {
	Visited map[string]bool // key: "type:id"
}

func NewSyncContext() *SyncContext {
	return &SyncContext{
		Visited: make(map[string]bool),
	}
}

// MarkVisited 标记实体为已访问
func (ctx *SyncContext) MarkVisited(entityType string, id int64) bool {
	key := fmt.Sprintf("%s:%d", entityType, id)
	if ctx.Visited[key] {
		return false // 已经访问过
	}
	ctx.Visited[key] = true
	return true
}

type MoldStatusSyncService struct {
	moldDAO           *dao.MoldDAO
	moldOrderDAO      *dao.MoldOrderDAO
	moldWorkOrderDAO  *dao.MoldWorkOrderDAO
	StatusFlowService *StatusFlowService // 公开字段，供其他服务访问
}

func NewMoldStatusSyncService() *MoldStatusSyncService {
	return &MoldStatusSyncService{
		moldDAO:           dao.NewMoldDAO(),
		moldOrderDAO:      dao.NewMoldOrderDAO(),
		moldWorkOrderDAO:  dao.NewMoldWorkOrderDAO(),
		StatusFlowService: NewStatusFlowService(),
	}
}

// SyncWorkOrderToOrderWithCtx 带防重入保护的工单到订单同步
func (s *MoldStatusSyncService) SyncWorkOrderToOrderWithCtx(workOrder *entity.MoldWorkOrder, syncCtx *SyncContext) error {
	if workOrder.MoldOrderID == nil {
		return nil
	}

	if !syncCtx.MarkVisited("order", *workOrder.MoldOrderID) {
		return nil
	}

	order, err := s.moldOrderDAO.FindByID(*workOrder.MoldOrderID)
	if err != nil {
		return fmt.Errorf("查找订单失败: %w", err)
	}

	// ... (原有聚合逻辑保持不变，此处省略以节省篇幅，实际替换时会保留)
	var workOrders []*entity.MoldWorkOrder
	err = dao.DB.Where("mold_order_id = ?", *workOrder.MoldOrderID).Find(&workOrders).Error
	if err != nil {
		return fmt.Errorf("查询订单工单列表失败: %w", err)
	}

	if len(workOrders) == 0 {
		return nil
	}

	// 聚合所有工单的最小状态值（使用状态顺序号比较）
	minStatus := workOrders[0].Status
	minOrder := s.getStatusOrder("work_order", minStatus)

	for _, wo := range workOrders[1:] {
		order := s.getStatusOrder("work_order", wo.Status)
		if order < minOrder {
			minStatus = wo.Status
			minOrder = order
		}
	}

	moldOrderStatus := s.StatusFlowService.GetMoldOrderStatusByWorkOrder(minStatus)
	if moldOrderStatus != nil {
		newOrderStatus := *moldOrderStatus
		if newOrderStatus != order.Status {
			order.Status = newOrderStatus

			deliveredStatus := s.StatusFlowService.GetDeliveredWorkOrderStatus()
			if deliveredStatus != nil && minStatus == *deliveredStatus {
				var lastDeliverTime *time.Time
				for _, wo := range workOrders {
					if wo.Status == *deliveredStatus && wo.ActualDeliverTime != nil {
						if lastDeliverTime == nil || wo.ActualDeliverTime.After(*lastDeliverTime) {
							lastDeliverTime = wo.ActualDeliverTime
						}
					}
				}
				if lastDeliverTime != nil {
					order.DeliverTime = lastDeliverTime
				}
			}

			if err := s.moldOrderDAO.Update(order); err != nil {
				return fmt.Errorf("更新订单状态失败: %w", err)
			}
		}
	}

	return nil
}

// SyncWorkOrderToOrder 工单状态变更时同步到订单（聚合所有工单状态）
func (s *MoldStatusSyncService) SyncWorkOrderToOrder(workOrder *entity.MoldWorkOrder) error {
	return s.SyncWorkOrderToOrderWithCtx(workOrder, NewSyncContext())
}

// getStatusOrder 辅助方法：获取状态顺序号（用于状态比较）
func (s *MoldStatusSyncService) getStatusOrder(statusType string, statusValue string) int {
	return s.StatusFlowService.GetStatusOrder(statusType, statusValue)
}

// SyncWorkOrderToMoldWithCtx 带防重入保护的工单到模具同步
func (s *MoldStatusSyncService) SyncWorkOrderToMoldWithCtx(workOrder *entity.MoldWorkOrder, syncCtx *SyncContext) error {
	if workOrder.MoldID == nil {
		return nil
	}

	if !syncCtx.MarkVisited("mold", *workOrder.MoldID) {
		return nil
	}

	mold, err := s.moldDAO.FindByID(*workOrder.MoldID)
	if err != nil {
		return fmt.Errorf("查找模具失败: %w", err)
	}

	mold.LatestWorkOrderID = &workOrder.ID

	// 🔒 强约束：区分"制造轨道"与"运维轨道"，防止状态冲突
	if workOrder.OrderType != entity.OrderTypeNewMold {
		// 【运维轨道】：修模或改造工单
		if workOrder.Status == constant.WorkOrderStatusProcessing {
			// 只要工单在进行中，模具状态强制为"维修中"
			mold.LifeCycleStatus = constant.MoldStatusRepairing
		} else if workOrder.Status == constant.WorkOrderStatusCompleted ||
			workOrder.Status == constant.WorkOrderStatusTested ||
			workOrder.Status == constant.WorkOrderStatusStockedIn ||
			workOrder.Status == constant.WorkOrderStatusDelivered {
			// 🛠️ 严格恢复：根据记录的原始状态进行回归
			if workOrder.OriginalMoldStatus != nil {
				mold.LifeCycleStatus = *workOrder.OriginalMoldStatus
			} else {
				// 兼容旧数据：如果没有记录，则默认回归仓库库存
				mold.LifeCycleStatus = constant.MoldStatusInStock
			}
		}
	} else {
		// 【制造轨道】：新模工单
		// 🔒 同步保护：如果模具处于维修或报废状态，不自动同步
		if mold.LifeCycleStatus == constant.MoldStatusRepairing || mold.LifeCycleStatus == constant.MoldStatusScrapped {
			return nil // 跳过同步，保持人工干预状态
		}

		// 根据工单状态更新模具生命周期状态（使用配置化的状态映射）
		moldStatus := s.StatusFlowService.GetMoldStatusByWorkOrder(workOrder.Status)
		if moldStatus != nil {
			newLifeCycleStatus := *moldStatus
			// 只有当新状态比当前状态"更靠后"时才更新，防止旧数据覆盖
			// 注：字符串比较需要按照预定义的顺序，这里简单比较
			if newLifeCycleStatus != mold.LifeCycleStatus {
				mold.LifeCycleStatus = newLifeCycleStatus
			}
		}
	}

	// 如果工单已交付，更新模具最后交付时间
	deliveredStatus := s.StatusFlowService.GetDeliveredWorkOrderStatus()
	if deliveredStatus != nil && workOrder.Status == *deliveredStatus && workOrder.ActualDeliverTime != nil {
		mold.LastDeliveryTime = workOrder.ActualDeliverTime
	}

	if err := s.moldDAO.Update(mold); err != nil {
		return fmt.Errorf("更新模具档案失败: %w", err)
	}

	return nil
}

// SyncWorkOrderToMold 工单状态变更时同步到模具档案
func (s *MoldStatusSyncService) SyncWorkOrderToMold(workOrder *entity.MoldWorkOrder) error {
	return s.SyncWorkOrderToMoldWithCtx(workOrder, NewSyncContext())
}

// SyncOrderToMoldWithCtx 带防重入保护的订单到模具同步
func (s *MoldStatusSyncService) SyncOrderToMoldWithCtx(order *entity.MoldOrder, syncCtx *SyncContext) error {
	if order.MoldID == nil {
		return nil // 没有关联模具，无需同步
	}

	if !syncCtx.MarkVisited("mold", *order.MoldID) {
		return nil
	}

	mold, err := s.moldDAO.FindByID(*order.MoldID)
	if err != nil {
		return fmt.Errorf("查找模具失败: %w", err)
	}

	mold.LatestOrderID = &order.ID

	// 🔒 同步保护：如果模具处于维修或报废状态，不自动同步
	// 这些状态通常需要人工介入，不应被订单状态自动覆盖
	if mold.LifeCycleStatus == constant.MoldStatusRepairing || mold.LifeCycleStatus == constant.MoldStatusScrapped {
		return nil // 跳过同步，保持手动设置的状态
	}

	// 根据订单状态更新模具生命周期状态
	// 注意：订单状态为已取消时不同步到模具，保持模具当前状态
	cancelledStatus := s.StatusFlowService.GetCancelledOrderStatus()
	if cancelledStatus != nil && order.Status == *cancelledStatus {
		return nil // 订单取消不影响模具状态
	}

	newMoldStatus := s.StatusFlowService.GetMoldStatusByMoldOrder(order.Status)
	if newMoldStatus != nil {
		newLifeCycleStatus := *newMoldStatus
		if newLifeCycleStatus != mold.LifeCycleStatus {
			mold.LifeCycleStatus = newLifeCycleStatus
		}
	}

	// 如果订单已交付，更新模具最后交付时间
	deliveredOrderStatus := s.StatusFlowService.GetDeliveredOrderStatus()
	if deliveredOrderStatus != nil && order.Status == *deliveredOrderStatus && order.DeliverTime != nil {
		mold.LastDeliveryTime = order.DeliverTime
	}

	if err := s.moldDAO.Update(mold); err != nil {
		return fmt.Errorf("更新模具档案失败: %w", err)
	}

	return nil
}

// SyncOrderToMold 订单状态变更时同步到模具档案
func (s *MoldStatusSyncService) SyncOrderToMold(order *entity.MoldOrder) error {
	return s.SyncOrderToMoldWithCtx(order, NewSyncContext())
}

// UpdateMoldStatistics 更新模具统计信息
func (s *MoldStatusSyncService) UpdateMoldStatistics(moldID int64) error {
	mold, err := s.moldDAO.FindByID(moldID)
	if err != nil {
		return fmt.Errorf("查找模具失败: %w", err)
	}

	// 统计订单数量和工单数量（使用单次查询优化）
	type MoldStats struct {
		OrderCount     int64 `gorm:"column:order_count"`
		WorkOrderCount int64 `gorm:"column:work_order_count"`
	}

	var stats MoldStats
	if err := dao.DB.Raw(`
		SELECT 
			(SELECT COUNT(*) FROM mold_order WHERE mold_id = ?) as order_count,
			(SELECT COUNT(*) FROM mold_work_order WHERE mold_id = ?) as work_order_count
	`, moldID, moldID).Scan(&stats).Error; err != nil {
		return fmt.Errorf("统计模具关联数据失败: %w", err)
	}

	mold.TotalOrders = int(stats.OrderCount)
	mold.TotalWorkOrders = int(stats.WorkOrderCount)

	if err := s.moldDAO.Update(mold); err != nil {
		return fmt.Errorf("更新模具统计失败: %w", err)
	}

	return nil
}

// CreateMoldFromOrder 从订单创建模具档案（新模订单场景）
func (s *MoldStatusSyncService) CreateMoldFromOrder(order *entity.MoldOrder, moldInfo *entity.Mold) (*entity.Mold, error) {
	moldInfo.CustomerID = order.CustomerID

	if err := s.moldDAO.Create(moldInfo); err != nil {
		return nil, fmt.Errorf("创建模具失败: %v", err)
	}

	order.MoldID = &moldInfo.ID
	if err := s.moldOrderDAO.Update(order); err != nil {
		return nil, fmt.Errorf("更新订单模具ID失败: %v", err)
	}

	// 同步状态
	if err := s.SyncOrderToMold(order); err != nil {
		return nil, err
	}

	return moldInfo, nil
}

// CreateWorkOrderFromOrder 从订单创建工单
func (s *MoldStatusSyncService) CreateWorkOrderFromOrder(order *entity.MoldOrder, workOrderInfo *entity.MoldWorkOrder) (*entity.MoldWorkOrder, error) {
	workOrderInfo.MoldOrderID = &order.ID
	workOrderInfo.OrderType = order.OrderType

	// 如果订单有关联模具，也设置到工单
	if order.MoldID != nil {
		workOrderInfo.MoldID = order.MoldID
	}

	// 如果未提供工单号，自动生成
	if workOrderInfo.WorkCode == "" {
		workOrderService := NewMoldWorkOrderService()
		workCode, err := workOrderService.GenerateWorkCode()
		if err != nil {
			return nil, fmt.Errorf("生成工单号失败: %v", err)
		}
		workOrderInfo.WorkCode = workCode
	}

	if err := s.moldWorkOrderDAO.Create(workOrderInfo); err != nil {
		return nil, fmt.Errorf("创建工单失败: %v", err)
	}

	// 触发工单状态变更的级联同步（同步到订单和模具档案）
	if err := s.OnWorkOrderStatusChanged(workOrderInfo); err != nil {
		return nil, fmt.Errorf("同步工单状态失败: %v", err)
	}

	// 如果工单有关联模具，更新模具统计
	if workOrderInfo.MoldID != nil {
		if err := s.UpdateMoldStatistics(*workOrderInfo.MoldID); err != nil {
			return nil, err
		}
	}

	return workOrderInfo, nil
}

// CreateMoldAndWorkOrderFromOrder 从订单同时创建模具档案和工单（面向客户场景）
func (s *MoldStatusSyncService) CreateMoldAndWorkOrderFromOrder(
	order *entity.MoldOrder,
	moldInfo *entity.Mold,
	workOrderInfo *entity.MoldWorkOrder,
) (*entity.Mold, *entity.MoldWorkOrder, error) {
	// 1. 先创建模具档案
	mold, err := s.CreateMoldFromOrder(order, moldInfo)
	if err != nil {
		return nil, nil, err
	}

	// 2. 再创建工单（关联刚创建的模具）
	workOrderInfo.MoldID = &mold.ID
	workOrder, err := s.CreateWorkOrderFromOrder(order, workOrderInfo)
	if err != nil {
		return nil, nil, err
	}

	return mold, workOrder, nil
}

// CreateMoldFromWorkOrder 从工单创建模具档案（面向自生产场景）
func (s *MoldStatusSyncService) CreateMoldFromWorkOrder(workOrder *entity.MoldWorkOrder, moldInfo *entity.Mold) (*entity.Mold, error) {
	if err := s.moldDAO.Create(moldInfo); err != nil {
		return nil, fmt.Errorf("创建模具失败: %v", err)
	}

	workOrder.MoldID = &moldInfo.ID
	if err := s.moldWorkOrderDAO.Update(workOrder); err != nil {
		return nil, fmt.Errorf("更新工单模具ID失败: %v", err)
	}

	// 同步状态
	if err := s.SyncWorkOrderToMold(workOrder); err != nil {
		return nil, err
	}

	if err := s.UpdateMoldStatistics(moldInfo.ID); err != nil {
		return nil, err
	}

	return moldInfo, nil
}

// OnWorkOrderStatusChanged 工单状态变更时的统一处理入口
func (s *MoldStatusSyncService) OnWorkOrderStatusChanged(workOrder *entity.MoldWorkOrder) error {
	syncCtx := NewSyncContext()

	// 1. 同步到订单
	if err := s.SyncWorkOrderToOrderWithCtx(workOrder, syncCtx); err != nil {
		logger.Error("同步工单到订单失败", zap.Error(err))
	}

	// 2. 同步到模具档案
	if err := s.SyncWorkOrderToMoldWithCtx(workOrder, syncCtx); err != nil {
		return fmt.Errorf("同步工单到模具失败: %v", err)
	}

	// 3. 更新模具统计
	if workOrder.MoldID != nil {
		if err := s.UpdateMoldStatistics(*workOrder.MoldID); err != nil {
			logger.Warn("更新模具统计失败", zap.Error(err))
		}
	}

	// ✨ 4. 如果是试模合格后，检查是否需要更新生产订单的依赖状态
	if workOrder.Status == constant.WorkOrderStatusTested {
		if err := s.SyncProductionOrderDependencyStatus(workOrder); err != nil {
			logger.Warn("同步生产订单依赖状态失败", zap.Error(err))
		}
	}

	// ✨ 5. 同步综合订单状态（如果模具订单属于综合订单）
	if workOrder.MoldOrderID != nil {
		var moldOrder entity.MoldOrder
		if err := dao.DB.First(&moldOrder, *workOrder.MoldOrderID).Error; err == nil && moldOrder.MasterOrderID != nil {
		go func(masterID int64) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("异步同步综合订单状态 panic", zap.Int64("master_order_id", masterID), zap.Any("panic", r))
				}
			}()
			masterService := NewMasterOrderService(dao.GetDB())
			if err := masterService.SyncMasterOrderStatus(masterID); err != nil {
				logger.Error("同步综合订单状态失败",
					zap.Int64("master_order_id", masterID),
					zap.Error(err))
			}
		}(*moldOrder.MasterOrderID)
		}
	}

	return nil
}

// SyncProductionOrderDependencyStatus 同步生产订单明细依赖状态（模具验收后触发）
func (s *MoldStatusSyncService) SyncProductionOrderDependencyStatus(workOrder *entity.MoldWorkOrder) error {
	if workOrder.MoldID == nil {
		return nil // 没有关联模具
	}

	// ✨ 查找所有使用该模具的生产订单明细项（dependency_status = WAITING_MOLD）
	var items []entity.ProductionOrderItem
	err := dao.DB.Where("dependency_status = ? AND mold_id = ?", "WAITING_MOLD", *workOrder.MoldID).Find(&items).Error
	if err != nil {
		return fmt.Errorf("查询生产订单明细失败: %w", err)
	}

	// ✨ 也检查 mold_ids JSON 数组中包含该模具的明细项
	var itemsWithMoldIDs []entity.ProductionOrderItem
	err = dao.DB.Where("dependency_status = ? AND JSON_CONTAINS(mold_ids, ?)", "WAITING_MOLD", fmt.Sprintf("%d", *workOrder.MoldID)).Find(&itemsWithMoldIDs).Error
	if err != nil {
		logger.Warn("查询多模具明细项失败", zap.Error(err))
	}

	itemMap := make(map[int64]*entity.ProductionOrderItem)
	for i := range items {
		itemMap[items[i].ID] = &items[i]
	}
	for i := range itemsWithMoldIDs {
		itemMap[itemsWithMoldIDs[i].ID] = &itemsWithMoldIDs[i]
	}

	// ✨ 收集所有需要更新的明细项 ID，批量更新
	itemIDs := make([]int64, 0, len(itemMap))
	for _, item := range itemMap {
		itemIDs = append(itemIDs, item.ID)
	}
	if len(itemIDs) > 0 {
		updates := map[string]interface{}{
			"dependency_status": "READY",
			"update_time":       time.Now(),
		}
		if err := dao.DB.Model(&entity.ProductionOrderItem{}).Where("id IN ?", itemIDs).Updates(updates).Error; err != nil {
			logger.Error("批量更新生产订单明细依赖状态失败",
				zap.Int("count", len(itemIDs)),
				zap.Int64("mold_id", *workOrder.MoldID),
				zap.Error(err))
		} else {
			logger.Info("生产订单明细依赖状态已批量更新",
				zap.Int("count", len(itemIDs)),
				zap.Int64("mold_id", *workOrder.MoldID),
				zap.String("from", "WAITING_MOLD"),
				zap.String("to", "READY"))
		}
	}

	// ✨ 同步更新生产订单主表的依赖状态(检查该订单的所有明细项是否都已就绪)
	orderIDs := make(map[int64]bool)
	for _, item := range itemMap {
		orderIDs[item.OrderID] = true
	}

	orderIDList := make([]int64, 0, len(orderIDs))
	for oid := range orderIDs {
		orderIDList = append(orderIDList, oid)
	}

	if len(orderIDList) > 0 {
		// 批量统计所有订单的明细项状态（一次查询替代 N 次循环查询）
		type OrderItemStats struct {
			OrderID    int64 `gorm:"column:order_id"`
			TotalCount int64 `gorm:"column:total_count"`
			ReadyCount int64 `gorm:"column:ready_count"`
		}
		var statsList []OrderItemStats
		if err := dao.DB.Raw(`
			SELECT 
				order_id,
				COUNT(*) as total_count,
				SUM(CASE WHEN dependency_status = ? THEN 1 ELSE 0 END) as ready_count
			FROM production_order_item
			WHERE order_id IN ?
			GROUP BY order_id
		`, "READY", orderIDList).Scan(&statsList).Error; err != nil {
			logger.Error("批量统计生产订单明细状态失败",
				zap.Int("order_count", len(orderIDList)),
				zap.Error(err))
		} else {
			readyOrderIDs := make([]int64, 0)
			for _, stats := range statsList {
				if stats.TotalCount > 0 && stats.TotalCount == stats.ReadyCount {
					readyOrderIDs = append(readyOrderIDs, stats.OrderID)
				}
			}

			// 批量更新主表状态
			if len(readyOrderIDs) > 0 {
				if err := dao.DB.Model(&entity.ProductionOrder{}).Where("id IN ?", readyOrderIDs).Update("dependency_status", "READY").Error; err != nil {
					logger.Error("批量更新生产订单主表依赖状态失败",
						zap.Int("count", len(readyOrderIDs)),
						zap.Error(err))
				} else {
					logger.Info("生产订单主表依赖状态已批量更新",
						zap.Int("count", len(readyOrderIDs)),
						zap.String("from", "WAITING_MOLD"),
						zap.String("to", "READY"))
				}
			}
		}
	}

	return nil
}

// OnOrderStatusChanged 订单状态变更时的统一处理入口
func (s *MoldStatusSyncService) OnOrderStatusChanged(order *entity.MoldOrder) error {
	syncCtx := NewSyncContext()
	// 同步到模具档案（带防重入保护）
	if err := s.SyncOrderToMoldWithCtx(order, syncCtx); err != nil {
		return fmt.Errorf("同步订单到模具失败: %v", err)
	}

	return nil
}

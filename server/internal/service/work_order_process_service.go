package service

import (
	"fmt"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WorkOrderProcessService struct {
	workOrderProcessDAO *dao.WorkOrderProcessDAO
	processTemplateDAO  *dao.ProcessTemplateDAO
	moldWorkOrderDAO    *dao.MoldWorkOrderDAO
	moldWorkOrderSvc    *MoldWorkOrderService
	statusFlowService   *StatusFlowService
}

func NewWorkOrderProcessService() *WorkOrderProcessService {
	return &WorkOrderProcessService{
		workOrderProcessDAO: dao.NewWorkOrderProcessDAO(),
		processTemplateDAO:  dao.NewProcessTemplateDAO(),
		moldWorkOrderDAO:    dao.NewMoldWorkOrderDAO(),
		moldWorkOrderSvc:    NewMoldWorkOrderService(),
		statusFlowService:   NewStatusFlowService(),
	}
}

// GetWorkOrderProcesses 获取工单的所有工序
func (s *WorkOrderProcessService) GetWorkOrderProcesses(workOrderID int64) ([]*entity.WorkOrderProcess, error) {
	return s.workOrderProcessDAO.FindByWorkOrderID(workOrderID)
}

// ToggleProcess 切换工序完成状态
func (s *WorkOrderProcessService) ToggleProcess(workOrderID, processTemplateID int64, isCompleted int8) error {
	workOrder, err := s.moldWorkOrderDAO.FindByID(workOrderID)
	if err != nil {
		return fmt.Errorf("工单不存在: %v", err)
	}

	if workOrder.IsCancelled == 1 {
		return fmt.Errorf("工单已取消，无法操作工序")
	}

	// 🔒 强约束：工单互斥检查 - 优化：单次查询替代 Count+First
	if workOrder.MoldID != nil && workOrder.OrderType != constant.MoldOrderTypeRepair && workOrder.OrderType != constant.MoldOrderTypeModification {
		var activeWorkOrder entity.MoldWorkOrder
		err := dao.DB.Where("mold_id = ? AND id != ? AND status >= ? AND status < ? AND is_cancelled = 0 AND (order_type = ? OR order_type = ?)",
			*workOrder.MoldID, workOrderID, entity.WorkOrderStatusNotStarted, entity.WorkOrderStatusDelivered,
			constant.MoldOrderTypeRepair, constant.MoldOrderTypeModification).
			First(&activeWorkOrder).Error
		if err == nil {
			// 找到了进行中的修模/改造工单
			orderTypeText := "修模"
			if activeWorkOrder.OrderType == constant.MoldOrderTypeModification {
				orderTypeText = "改造"
			}
			return fmt.Errorf("该模具已有进行中的%s工单（%s），当前工单已被锁定，无法操作工序",
				orderTypeText, activeWorkOrder.WorkCode)
		}
	}

	// 🔒 强约束：禁止对已完成的旧工单进行工序回退（撤销操作）
	if workOrder.Status != constant.WorkOrderStatusNotStarted && workOrder.Status != constant.WorkOrderStatusProcessing && isCompleted < 1 {
		return fmt.Errorf("工单已进入后期阶段，禁止回退工序操作")
	}

	// 🔒 强约束：验证工序阶段依赖 - 完成/跳过操作需要检查前置工序
	if isCompleted >= 1 {
		if err := s.validateProcessDependency(workOrderID, processTemplateID); err != nil {
			return err
		}
	}

	// 查找或创建关联记录
	wop, err := s.workOrderProcessDAO.FindOrCreate(workOrderID, processTemplateID)
	if err != nil {
		return err
	}

	wop.IsCompleted = isCompleted
	if isCompleted == 1 {
		now := time.Now()
		wop.CompletedAt = &now
	} else {
		wop.CompletedAt = nil
	}

	if err := s.workOrderProcessDAO.Update(wop); err != nil {
		return err
	}

	// ✅ 性能优化：传入已查询的工单对象，避免重复查询
	if err := s.updateWorkOrderProgressAndStatusWithWorkOrder(workOrderID, workOrder); err != nil {
		logger.Warn("更新工单进度失败", zap.Error(err))
	}

	return nil
}

// CreateWorkOrderProcesses 为工单创建所有工序关联
func (s *WorkOrderProcessService) CreateWorkOrderProcesses(workOrderID int64) error {
	return s.CreateWorkOrderProcessesInTx(dao.DB, workOrderID)
}

// CreateWorkOrderProcessesInTx 在事务中为工单创建所有工序关联
func (s *WorkOrderProcessService) CreateWorkOrderProcessesInTx(tx *gorm.DB, workOrderID int64) error {
	templates, err := s.processTemplateDAO.FindActive()
	if err != nil {
		return err
	}

	if len(templates) == 0 {
		return nil
	}

	var wops []*entity.WorkOrderProcess
	for _, template := range templates {
		wops = append(wops, &entity.WorkOrderProcess{
			WorkOrderID:       workOrderID,
			ProcessTemplateID: template.ID,
			IsCompleted:       0,
		})
	}

	// 批量插入：GORM 会自动生成批量 SQL（INSERT INTO ... VALUES (...), (...), ...）
	if err := tx.Create(&wops).Error; err != nil {
		return err
	}

	return nil
}

// updateWorkOrderProgressAndStatusWithWorkOrder 计算并更新工单总进度和状态（传入已查询的工单对象，避免重复查询）
func (s *WorkOrderProcessService) updateWorkOrderProgressAndStatusWithWorkOrder(workOrderID int64, workOrder *entity.MoldWorkOrder) error {
	// ✅ 性能优化：只查询必要字段，避免 JOIN process_template 表
	var wops []*entity.WorkOrderProcess
	if err := dao.DB.Model(&entity.WorkOrderProcess{}).
		Select("id", "work_order_id", "process_template_id", "is_completed").
		Where("work_order_id = ?", workOrderID).
		Find(&wops).Error; err != nil {
		return err
	}

	if len(wops) == 0 {
		return nil
	}

	validCount := 0
	completedCount := 0
	for _, wop := range wops {
		if wop.IsCompleted != 2 { // 跳过状态的不计入总数
			validCount++
			if wop.IsCompleted == 1 {
				completedCount++
			}
		}
	}

	progress := 0
	if validCount > 0 {
		progress = (completedCount * 100) / validCount
	}

	// ✅ 使用传入的 workOrder 对象，避免重复查询
	workOrder.Progress = progress

	// 根据完成比例更新工单状态（根据实际进度动态校准）
	// 工单状态流转：未开始 → 加工中 → 已完成 → 已试模 → 已验收 → 已入库 → 已交付
	// 注意：状态4及以上是独立流程（试模/验收/入库/交付），不受工序进度影响
	// 但状态3及以下是根据工序进度决定的，需要动态校准
	newStatus := workOrder.Status // 默认保持当前状态

	// 获取“已完成”状态的阈值（用于判断是否进入试模阶段）
	completedStatus := s.statusFlowService.GetCompletedWorkOrderStatus()

	if workOrder.Status == constant.WorkOrderStatusTested ||
		workOrder.Status == constant.WorkOrderStatusStockedIn ||
		workOrder.Status == constant.WorkOrderStatusDelivered ||
		(completedStatus != nil && workOrder.Status == *completedStatus) {
		// 状态已经是试模及以上，不受工序进度影响，保持原状态
		newStatus = workOrder.Status
	} else {
		// 状态在已完成及以下，根据进度动态校准
		// 使用语义标签动态获取状态值，避免硬编码
		newStatusValue := s.statusFlowService.GetWorkOrderStatusByProgress(progress)
		if newStatusValue != nil {
			newStatus = *newStatusValue
		}
	}

	// 只有状态发生变化时才更新
	if workOrder.Status != newStatus {
		workOrder.Status = newStatus
		if err := s.moldWorkOrderDAO.Update(workOrder); err != nil {
			return err
		}

		// ✅ 性能优化：异步触发工单状态变更的级联同步（避免阻塞响应）
		go func(wo *entity.MoldWorkOrder) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("异步同步工单状态 panic", zap.Int64("work_order_id", wo.ID), zap.Any("panic", r))
				}
			}()
			if err := s.moldWorkOrderSvc.statusSyncService.OnWorkOrderStatusChanged(wo); err != nil {
				logger.Error("异步同步工单状态失败", zap.Error(err), zap.Int64("work_order_id", wo.ID))
			}
		}(workOrder)
	} else {
		// 状态不变，只更新进度
		if err := s.moldWorkOrderDAO.Update(workOrder); err != nil {
			return err
		}
	}

	return nil
}

// DeleteWorkOrderProcesses 删除工单的所有工序关联
func (s *WorkOrderProcessService) DeleteWorkOrderProcesses(workOrderID int64) error {
	return s.workOrderProcessDAO.DeleteByWorkOrderID(workOrderID)
}

// validateProcessDependency 验证工序阶段依赖 - 检查前置工序是否已全部完成
func (s *WorkOrderProcessService) validateProcessDependency(workOrderID, processTemplateID int64) error {
	// 查询当前工序模板信息
	var processTemplate entity.ProcessTemplate
	if err := dao.DB.First(&processTemplate, processTemplateID).Error; err != nil {
		return fmt.Errorf("工序模板不存在: %v", err)
	}

	var workOrderProcesses []*entity.WorkOrderProcess
	if err := dao.DB.Preload("ProcessTemplate").Where("work_order_id = ?", workOrderID).Find(&workOrderProcesses).Error; err != nil {
		return fmt.Errorf("查询工序失败: %v", err)
	}

	// 按 group_id 分组
	type GroupInfo struct {
		GroupID   int
		GroupName string
		Processes []*entity.WorkOrderProcess
	}
	groupsMap := make(map[int]*GroupInfo)
	var groups []*GroupInfo

	for _, wop := range workOrderProcesses {
		groupID := wop.ProcessTemplate.GroupID
		if _, exists := groupsMap[groupID]; !exists {
			groupInfo := &GroupInfo{
				GroupID:   groupID,
				GroupName: wop.ProcessTemplate.GroupName,
				Processes: []*entity.WorkOrderProcess{},
			}
			groupsMap[groupID] = groupInfo
			groups = append(groups, groupInfo)
		}
		groupsMap[groupID].Processes = append(groupsMap[groupID].Processes, wop)
	}

	// 按 group_id 排序
	for i := 0; i < len(groups); i++ {
		for j := i + 1; j < len(groups); j++ {
			if groups[i].GroupID > groups[j].GroupID {
				groups[i], groups[j] = groups[j], groups[i]
			}
		}
	}

	// 找到当前工序所在的组
	currentGroupID := processTemplate.GroupID
	var currentGroupIndex int = -1
	for i, group := range groups {
		if group.GroupID == currentGroupID {
			currentGroupIndex = i
			break
		}
	}

	if currentGroupIndex == -1 {
		return fmt.Errorf("当前工序不属于任何分组")
	}

	// 如果是第一个分组，无需检查前置依赖
	if currentGroupIndex == 0 {
		return nil
	}

	for i := 0; i < currentGroupIndex; i++ {
		prevGroup := groups[i]
		for _, prevProcess := range prevGroup.Processes {
			// 前置工序必须完成（is_completed = 1）或跳过（is_completed = 2）
			if prevProcess.IsCompleted != 1 && prevProcess.IsCompleted != 2 {
				return fmt.Errorf("需等待上一组【%s】完成后才能操作，当前工序【%s】未完成",
					prevGroup.GroupName, prevProcess.ProcessTemplate.ProcessName)
			}
		}
	}

	return nil
}

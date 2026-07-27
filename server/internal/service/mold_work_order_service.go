package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MoldWorkOrderService struct {
	moldWorkOrderDAO  *dao.MoldWorkOrderDAO
	statusSyncService *MoldStatusSyncService
	moldService       *MoldService
	moldStockService  *MoldStockService
}

func NewMoldWorkOrderService() *MoldWorkOrderService {
	return &MoldWorkOrderService{
		moldWorkOrderDAO:  dao.NewMoldWorkOrderDAO(),
		statusSyncService: NewMoldStatusSyncService(),
		moldService:       NewMoldService(),
		moldStockService:  NewMoldStockService(),
	}
}

// GenerateWorkCode 生成工单号
// 格式：MW-YYMMDD-01
// MW: Mold Work 模具工单固定前缀
// 如果提供了事务(tx)，则在事务内查询以确保并发安全
func (s *MoldWorkOrderService) GenerateWorkCode() (string, error) {
	return s.GenerateWorkCodeWithTx(nil)
}

// GenerateWorkCodeWithTx 在指定事务中生成工单号（带重试机制）
func (s *MoldWorkOrderService) GenerateWorkCodeWithTx(tx *gorm.DB) (string, error) {
	const maxRetries = 3 // 最大重试次数

	for attempt := 1; attempt <= maxRetries; attempt++ {
		now := time.Now()
		dateStr := now.Format("060102") // YYMMDD 6位日期

		// 固定前缀 MW (Mold Work)
		prefix := "MW"

		// 构造查询前缀：MW-日期-
		queryPrefix := fmt.Sprintf("%s-%s-", prefix, dateStr)

		// 查找该日期下的最大序列号
		var maxSeq int
		var err error
		if tx != nil {
			// 在事务内查询，使用 SELECT ... FOR UPDATE 锁定
			err = tx.Model(&entity.MoldWorkOrder{}).
				Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("work_code LIKE ?", queryPrefix+"%").
				Select("COALESCE(MAX(CAST(SUBSTRING_INDEX(work_code, '-', -1) AS UNSIGNED)), 0)").
				Scan(&maxSeq).Error
		} else {
			maxSeq, err = s.moldWorkOrderDAO.FindMaxSequenceByPrefix(queryPrefix)
		}
		if err != nil {
			return "", fmt.Errorf("查询最大序列号失败: %w", err)
		}

		newSeq := maxSeq + 1
		workCode := fmt.Sprintf("%s%02d", queryPrefix, newSeq)

		var existingCount int64
		if tx != nil {
			tx.Model(&entity.MoldWorkOrder{}).Where("work_code = ?", workCode).Count(&existingCount)
		} else {
			dao.DB.Model(&entity.MoldWorkOrder{}).Where("work_code = ?", workCode).Count(&existingCount)
		}

		if existingCount == 0 {
			// 工单号可用，返回
			return workCode, nil
		}

		// 工单号已存在，记录日志并重试
		logger.Warn("工单号冲突，重试",
			zap.String("work_code", workCode),
			zap.Int("attempt", attempt),
			zap.Int("max_retries", maxRetries))

		// 如果不是最后一次尝试，等待一小段时间后重试
		if attempt < maxRetries {
			time.Sleep(10 * time.Millisecond)
		}
	}

	// 重试多次后仍然失败，返回错误
	return "", fmt.Errorf("生成工单号失败：多次重试后仍发现冲突，请稍后重试")
}

func (s *MoldWorkOrderService) CreateMoldWorkOrder(workOrderDTO dto.MoldWorkOrderDTO, userID *int64) (*entity.MoldWorkOrder, error) {
	// 如果未提供工单号，自动生成
	if workOrderDTO.WorkCode == "" {
		var err error
		workOrderDTO.WorkCode, err = s.GenerateWorkCode()
		if err != nil {
			return nil, fmt.Errorf("生成工单号失败: %v", err)
		}
	} else {
		// 如果手动提供了工单号，检查是否已存在
		existing, _ := s.moldWorkOrderDAO.FindByCode(workOrderDTO.WorkCode)
		if existing != nil {
			return nil, errors.New("工单号已存在")
		}
	}

	var planStart *time.Time
	if workOrderDTO.PlanStart != nil && !workOrderDTO.PlanStart.Time.IsZero() {
		t := workOrderDTO.PlanStart.Time
		planStart = &t
	}

	var planEnd *time.Time
	if workOrderDTO.PlanEnd != nil && !workOrderDTO.PlanEnd.Time.IsZero() {
		t := workOrderDTO.PlanEnd.Time
		planEnd = &t
	}

	var deliverTime *time.Time
	if workOrderDTO.DeliverTime != nil && !workOrderDTO.DeliverTime.Time.IsZero() {
		t := workOrderDTO.DeliverTime.Time
		deliverTime = &t
	}

	workOrder := &entity.MoldWorkOrder{
		WorkCode:    workOrderDTO.WorkCode,
		MoldOrderID: workOrderDTO.MoldOrderID,
		OrderType:   workOrderDTO.OrderType,
		MoldID:      workOrderDTO.MoldID,
		PlanStart:   planStart,
		PlanEnd:     planEnd,
		Progress:    workOrderDTO.Progress,
		// 🔒 状态由系统自动管理，不从 DTO 接收
		Status:      constant.WorkOrderStatusNotStarted, // 强制初始化为"未开始"
		WarehouseID: workOrderDTO.WarehouseID,
		DeliverTime: deliverTime,
		Remark:      workOrderDTO.Remark,
	}
	if workOrder.OrderType == "" {
		workOrder.OrderType = constant.MoldOrderTypeNew // 默认为新模
	}

	// 拷贝外发字段
	workOrder.WorkType = workOrderDTO.WorkType
	workOrder.ProcessorID = workOrderDTO.ProcessorID
	if workOrderDTO.OutsourcePrice != nil {
		workOrder.OutsourcePrice = *workOrderDTO.OutsourcePrice
	}
	if workOrderDTO.OutsourceTotalAmount != nil {
		workOrder.OutsourceTotalAmount = *workOrderDTO.OutsourceTotalAmount
	}

	var createdWorkOrder *entity.MoldWorkOrder

	// 开启事务确保数据一致性
	err := dao.DB.Transaction(func(tx *gorm.DB) error {
		// 根据订单类型进行不同处理
		if workOrder.OrderType == constant.MoldOrderTypeNew {
			// 新模工单：必须提供新模具信息
			if workOrderDTO.NewMoldCode == "" {
				return &response.HttpError{Code: response.BAD_REQUEST, Message: "新模工单必须填写模具编号"}
			}
			if workOrderDTO.NewMoldName == "" {
				return &response.HttpError{Code: response.BAD_REQUEST, Message: "新模工单必须填写产品名称"}
			}
			if workOrderDTO.CustomerID == 0 {
				return &response.HttpError{Code: response.BAD_REQUEST, Message: "新模工单必须选择客户"}
			}
			// 新模工单不应该关联已有模具
			if workOrderDTO.MoldID != nil {
				return &response.HttpError{Code: response.BAD_REQUEST, Message: "新模工单不应选择已有模具，系统将自动创建模具档案"}
			}

			// 🔒 强约束：如果提供了模具订单ID，检查该订单关联的模具是否有修模/改造工单在进行中
			if workOrderDTO.MoldOrderID != nil {
				var order entity.MoldOrder
				if err := tx.First(&order, *workOrderDTO.MoldOrderID).Error; err == nil && order.MoldID != nil {
					var activeRepairWorkOrderCount int64
					tx.Model(&entity.MoldWorkOrder{}).
					Where("mold_id = ? AND status IN (?, ?) AND is_cancelled = 0 AND (order_type = ? OR order_type = ?)",
						*order.MoldID, constant.WorkOrderStatusNotStarted, constant.WorkOrderStatusProcessing,
						constant.MoldOrderTypeRepair, constant.MoldOrderTypeModification).
						Count(&activeRepairWorkOrderCount)
				if activeRepairWorkOrderCount > 0 {
					return &response.HttpError{Code: response.BAD_REQUEST, Message: "该模具已有修模/改造工单在进行中，无法创建新的新模工单。请等待修模完成后再创建"}
				}
				}
			}

			// 1. 创建工单（在事务中）
			if err := tx.Create(workOrder).Error; err != nil {
				return err
			}

			// 2. 准备模具信息并创建
			moldInfo := &entity.Mold{
				MoldCode:        workOrderDTO.NewMoldCode,
				MoldName:        workOrderDTO.NewMoldName,
				CustomerID:      workOrderDTO.CustomerID,
				LifeCycleStatus: constant.MoldStatusDesigning, // 初始状态
				MoldType:        workOrderDTO.NewMoldType,
				MaterialType:    workOrderDTO.NewMaterialType,
				Remark:          workOrderDTO.NewMoldRemark,
			}

			if err := tx.Create(moldInfo).Error; err != nil {
				return fmt.Errorf("创建模具失败: %v", err)
			}

			// 3. 更新工单的模具ID
			workOrder.MoldID = &moldInfo.ID
			if err := tx.Save(workOrder).Error; err != nil {
				return fmt.Errorf("更新工单模具ID失败: %v", err)
			}

			// 4. 为工单创建工序关联（根据启用的工序模板）
			workOrderProcessService := NewWorkOrderProcessService()
			if err := workOrderProcessService.CreateWorkOrderProcessesInTx(tx, workOrder.ID); err != nil {
				return fmt.Errorf("创建工单工序失败: %v", err)
			}

			createdWorkOrder = workOrder
			return nil
		} else {
			// 修模/改模工单：必须选择已有模具
			if workOrder.MoldID == nil {
				return &response.HttpError{Code: response.BAD_REQUEST, Message: "修模/改模工单必须选择已有模具"}
			}

			// 🔒 强约束：检查模具是否处于不可修改的状态
			existingMold, err := s.moldService.GetMoldByID(*workOrder.MoldID)
			if err != nil {
				return &response.HttpError{Code: response.BAD_REQUEST, Message: "关联的模具不存在"}
			}
			if existingMold.LifeCycleStatus == constant.MoldStatusScrapped {
				return &response.HttpError{Code: response.BAD_REQUEST, Message: "已报废的模具无法创建工单"}
			}

			// 🔒 强约束：模具必须先完成新模制造才能创建修模工单（NOT_STARTED 或 PROCESSING 即为未完成）
			var unfinishedNewMoldWorkOrderCount int64
			tx.Model(&entity.MoldWorkOrder{}).
				Where("mold_id = ? AND order_type = ? AND status IN (?, ?)", *workOrder.MoldID, constant.MoldOrderTypeNew,
					constant.WorkOrderStatusNotStarted, constant.WorkOrderStatusProcessing).
				Count(&unfinishedNewMoldWorkOrderCount)
			if unfinishedNewMoldWorkOrderCount > 0 {
				return &response.HttpError{Code: response.BAD_REQUEST, Message: "该模具的新模制造工单尚未完成，无法创建修模/改造工单。请等待模具制造完成后再创建"}
			}

			// 🔒 强约束：工单互斥检查 - 同一模具不允许同时进行多个修模/改模工单
			// "进行中" = NOT_STARTED 或 PROCESSING（字符串精确匹配，不用 > < 因为 varchar 排序不等于状态顺序）
			var activeWorkOrderCount int64
			tx.Model(&entity.MoldWorkOrder{}).
				Where("mold_id = ? AND id != ? AND status IN (?, ?) AND is_cancelled = 0 AND (order_type = ? OR order_type = ?)",
					*workOrder.MoldID, 0, constant.WorkOrderStatusNotStarted, constant.WorkOrderStatusProcessing,
					constant.MoldOrderTypeRepair, constant.MoldOrderTypeModification).
				Count(&activeWorkOrderCount)
			if activeWorkOrderCount > 0 {
				// 查询正在进行中的工单信息，用于错误提示
				var activeWorkOrder entity.MoldWorkOrder
				tx.Where("mold_id = ? AND id != ? AND status IN (?, ?) AND is_cancelled = 0 AND (order_type = ? OR order_type = ?)",
					*workOrder.MoldID, 0, constant.WorkOrderStatusNotStarted, constant.WorkOrderStatusProcessing,
					constant.MoldOrderTypeRepair, constant.MoldOrderTypeModification).
				First(&activeWorkOrder)

				orderTypeText := "修模"
				if activeWorkOrder.OrderType == constant.MoldOrderTypeModification {
					orderTypeText = "改造"
				}

				return &response.HttpError{Code: response.BAD_REQUEST, Message: fmt.Sprintf("该模具已有进行中的%s工单（%s），请等待该工单完成后再创建新的修模工单",
					orderTypeText, activeWorkOrder.WorkCode)}
			}

			// 📝 记录修模前的原始状态，用于后续精确恢复
			originalStatus := existingMold.LifeCycleStatus
			workOrder.OriginalMoldStatus = &originalStatus

			// 🔒 校验：如果关联了模具订单，订单必须包含该模具
			if workOrder.MoldOrderID != nil {
				var itemCount int64
				tx.Model(&entity.MoldOrderItem{}).
					Where("order_id = ? AND mold_id = ?", *workOrder.MoldOrderID, *workOrder.MoldID).
					Count(&itemCount)
				if itemCount == 0 {
					// 也检查工单表中是否有该订单+模具的组合（兼容旧版无明细的订单）
					var woCount int64
					tx.Model(&entity.MoldWorkOrder{}).
						Where("mold_order_id = ? AND mold_id = ?", *workOrder.MoldOrderID, *workOrder.MoldID).
						Count(&woCount)
					if woCount == 0 {
						return &response.HttpError{Code: response.BAD_REQUEST, Message: "所选订单不包含该模具，请选择正确的订单"}
					}
				}
			}

			// 1. 直接创建工单（在事务中）
			if err := tx.Create(workOrder).Error; err != nil {
				return err
			}

			// 2. 为工单创建工序关联（根据启用的工序模板）
			workOrderProcessService := NewWorkOrderProcessService()
			if err := workOrderProcessService.CreateWorkOrderProcessesInTx(tx, workOrder.ID); err != nil {
				return fmt.Errorf("创建工单工序失败: %v", err)
			}

			createdWorkOrder = workOrder
			return nil
		}
	})

	if err != nil {
		return nil, err
	}

	// 事务提交后，执行状态同步和统计更新（这些操作失败不影响主流程）
	if createdWorkOrder.MoldID != nil {
		if err := s.statusSyncService.UpdateMoldStatistics(*createdWorkOrder.MoldID); err != nil {
			logger.Warn("更新模具统计失败", zap.Error(err))
		}

		// 触发状态同步
		if err := s.statusSyncService.OnWorkOrderStatusChanged(createdWorkOrder); err != nil {
			logger.Error("状态同步失败", zap.Error(err))
		}
	}

	// 如果是外发工单，自动创建外发记录
	if createdWorkOrder.WorkType == "OUTSOURCED" && createdWorkOrder.ProcessorID != nil {
		outsourcingService := NewOutsourcingService()
		recordCode, err := outsourcingService.GenerateRecordCode()
		if err == nil {
			outsourcingRecord := &entity.OutsourcingRecord{
				RecordCode:         recordCode,
				OrderID:            nil, // ✨ 模具外发不需要关联生产订单
				WorkOrderID:        &createdWorkOrder.ID,
				OutsourceType:      "MOLD",
				ProcessorID:        *createdWorkOrder.ProcessorID,
				MoldID:             createdWorkOrder.MoldID,
				Quantity:           1, // 模具外发数量为1
				UnitPrice:          createdWorkOrder.OutsourcePrice,
				TotalAmount:        createdWorkOrder.OutsourceTotalAmount,
				OutsourceDate:      createdWorkOrder.PlanStart,
				ExpectedReturnDate: createdWorkOrder.PlanEnd,
				Status:             "PENDING",
				CreateUserID:       userID,
			}
			if err := outsourcingService.outsourcingDAO.Create(outsourcingRecord); err == nil {
				logger.Info("模具外发工单创建成功，已自动生成外发记录",
					zap.Int64("work_order_id", createdWorkOrder.ID),
					zap.String("record_code", recordCode))
			} else {
				logger.Error("创建外发记录失败",
					zap.Int64("work_order_id", createdWorkOrder.ID),
					zap.Error(err))
			}
		}
	}

	// 重新查询更新后的工单（包含关联数据）
	return s.GetMoldWorkOrderByID(createdWorkOrder.ID)
}

func (s *MoldWorkOrderService) GetMoldWorkOrderByID(id int64) (*entity.MoldWorkOrder, error) {
	workOrder, err := s.moldWorkOrderDAO.FindByID(id)
	if err != nil {
		return nil, err
	}

	s.batchProcessWorkOrders([]*entity.MoldWorkOrder{workOrder})

	return workOrder, nil
}

func (s *MoldWorkOrderService) GetAllMoldWorkOrders() ([]*entity.MoldWorkOrder, error) {
	workOrders, err := s.moldWorkOrderDAO.FindAll()
	if err != nil {
		return nil, err
	}

	// 批量处理工单数据（避免N+1查询）
	s.batchProcessWorkOrders(workOrders)

	return workOrders, nil
}

func (s *MoldWorkOrderService) GetMoldWorkOrdersByPage(page, pageSize int) ([]*entity.MoldWorkOrder, int64, error) {
	result, err := s.moldWorkOrderDAO.FindPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	workOrders := result.List.([]*entity.MoldWorkOrder)
	// 批量处理工单数据（避免N+1查询）
	s.batchProcessWorkOrders(workOrders)

	return workOrders, result.Total, nil
}

// GetMoldWorkOrdersByPageWithFilter 带筛选条件的分页获取模具工单
func (s *MoldWorkOrderService) GetMoldWorkOrdersByPageWithFilter(page, pageSize int, workCode string, moldCode string, customerID *int64, status *int8, orderType string, processStage *string) ([]*entity.MoldWorkOrder, int64, error) {
	offset := (page - 1) * pageSize
	workOrders, total, err := s.moldWorkOrderDAO.FindPageWithFilter(offset, pageSize, workCode, moldCode, customerID, status, orderType, processStage)
	if err != nil {
		return nil, 0, err
	}

	// 批量处理工单数据（避免N+1查询）
	s.batchProcessWorkOrders(workOrders)

	return workOrders, total, nil
}

// GetMoldWorkOrdersByMoldOrderID 根据订单ID获取所有工单
func (s *MoldWorkOrderService) GetMoldWorkOrdersByMoldOrderID(moldOrderID int64) ([]*entity.MoldWorkOrder, error) {
	workOrders, err := s.moldWorkOrderDAO.FindByMoldOrderID(moldOrderID)
	if err != nil {
		return nil, err
	}

	// 批量处理工单数据（避免N+1查询）
	s.batchProcessWorkOrders(workOrders)

	return workOrders, nil
}

// batchProcessWorkOrders 批量处理工单数据（进度计算 + 工序附加 + 锁定状态）
// 通过批量查询避免N+1问题
func (s *MoldWorkOrderService) batchProcessWorkOrders(workOrders []*entity.MoldWorkOrder) {
	if len(workOrders) == 0 {
		return
	}

	// 1. 收集所有工单ID
	workOrderIDs := make([]int64, 0, len(workOrders))
	for _, wo := range workOrders {
		workOrderIDs = append(workOrderIDs, wo.ID)
	}

	// 2. 批量查询所有工单的工序数据（一次查询）
	wopDAO := dao.NewWorkOrderProcessDAO()
	allWops, err := wopDAO.FindByWorkOrderIDs(workOrderIDs)
	if err != nil {
		return
	}

	// 3. 按工单ID分组
	wopsByWorkOrderID := make(map[int64][]*entity.WorkOrderProcess)
	for _, wop := range allWops {
		wopsByWorkOrderID[wop.WorkOrderID] = append(wopsByWorkOrderID[wop.WorkOrderID], wop)
	}

	// 4. 查询一次工序模板总数（所有工单共享）
	ptDAO := dao.NewProcessTemplateDAO()
	allTemplates, err := ptDAO.FindActive()
	if err != nil {
		return
	}
	totalProcesses := len(allTemplates)
	if totalProcesses == 0 {
		return
	}

	// 5. 批量查询所有模具的活跃修模/改模工单（一次查询，只查询必要字段）
	moldIDs := make([]int64, 0, len(workOrders))
	moldIDSet := make(map[int64]bool)
	for _, wo := range workOrders {
		if wo.MoldID != nil && !moldIDSet[*wo.MoldID] {
			moldIDs = append(moldIDs, *wo.MoldID)
			moldIDSet[*wo.MoldID] = true
		}
	}

	// 查询这些模具的所有活跃修模/改模工单（只查 mold_id，避免加载关联数据）
	var activeMoldIDs []int64
	if len(moldIDs) > 0 {
		dao.DB.Model(&entity.MoldWorkOrder{}).
			Select("mold_id").
			Where("mold_id IN ? AND status >= ? AND status < ? AND is_cancelled = 0 AND (order_type = ? OR order_type = ?)",
				moldIDs, constant.WorkOrderStatusNotStarted, constant.WorkOrderStatusDelivered,
				constant.MoldOrderTypeRepair, constant.MoldOrderTypeModification).
			Find(&activeMoldIDs)
	}

	// 转为 Set，用于快速查找
	activeOrdersByMoldID := make(map[int64]bool)
	for _, moldID := range activeMoldIDs {
		activeOrdersByMoldID[moldID] = true
	}

	// 6. 为每个工单计算进度和附加工序
	// 提前获取状态配置（避免循环内重复调用）
	completedStatus := s.statusSyncService.StatusFlowService.GetCompletedWorkOrderStatus()

	for _, workOrder := range workOrders {
		wops := wopsByWorkOrderID[workOrder.ID]

		completedCount := 0
		processes := make([]entity.WorkOrderProcessInfo, 0, len(wops))

		for _, wop := range wops {
			if wop.IsCompleted == 1 {
				completedCount++
			}

			processName := ""
			if wop.ProcessTemplate != nil {
				processName = wop.ProcessTemplate.ProcessName
			}
			processes = append(processes, entity.WorkOrderProcessInfo{
				ProcessTemplateID: wop.ProcessTemplateID,
				ProcessName:       processName,
				IsCompleted:       wop.IsCompleted,
				CompletedAt:       wop.CompletedAt,
			})
		}

		progress := (completedCount * 100) / totalProcesses

		// 附加工序数据
		workOrder.Processes = processes
		workOrder.Progress = progress

		// 根据进度重新计算工单状态（状态必须与进度匹配）
		// 注意：如果工单已经进入试模及以后阶段，保持原状态不变
		// 但对于前期阶段的工单，状态必须严格匹配进度
		// 使用语义标签获取状态，而非硬编码
		canUpdateByProgress := false
		if completedStatus != nil {
			// 如果工单状态不等于"已完成"，说明还在工序阶段
			canUpdateByProgress = workOrder.Status == constant.WorkOrderStatusNotStarted || workOrder.Status == constant.WorkOrderStatusProcessing || workOrder.Status == constant.WorkOrderStatusCompleted
		} else {
			// 如果没有配置，使用默认逻辑（状态<=已完成）
			canUpdateByProgress = workOrder.Status == constant.WorkOrderStatusNotStarted || workOrder.Status == constant.WorkOrderStatusProcessing || workOrder.Status == constant.WorkOrderStatusCompleted
		}

		if canUpdateByProgress {
			newStatus := s.statusSyncService.StatusFlowService.GetWorkOrderStatusByProgress(progress)
			if newStatus != nil {
				workOrder.Status = *newStatus
			}
		}

		// 只有修模/改模工单会锁定其他工单
		if workOrder.MoldID != nil && (workOrder.OrderType == constant.MoldOrderTypeRepair || workOrder.OrderType == constant.MoldOrderTypeModification) {
			// 修模/改模工单本身不锁定
			workOrder.IsLocked = false
		} else if workOrder.MoldID != nil && activeOrdersByMoldID[*workOrder.MoldID] {
			// 如果该模具存在活跃修模/改模工单，则当前工单被锁定
			workOrder.IsLocked = true
		} else {
			workOrder.IsLocked = false
		}
	}
}

func (s *MoldWorkOrderService) UpdateMoldWorkOrder(workOrder *entity.MoldWorkOrder) error {
	oldWorkOrder, err := s.moldWorkOrderDAO.FindByID(workOrder.ID)
	if err := utils.EnsureExists(oldWorkOrder, err, "工单"); err != nil {
		return err
	}

	// 🔒 使用事务确保数据一致性：工单更新 + 自动入库 + 状态同步
	err = dao.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 更新工单
		if err := tx.Model(&entity.MoldWorkOrder{}).Where("id = ?", workOrder.ID).Updates(workOrder).Error; err != nil {
			return fmt.Errorf("更新工单失败: %w", err)
		}

		// 2. 如果工单状态变更为"已入库"，自动创建入库记录
		if workOrder.Status == constant.WorkOrderStatusStockedIn && oldWorkOrder.Status != constant.WorkOrderStatusStockedIn && workOrder.MoldID != nil {
			stock := &entity.MoldStock{
				MoldID:      *workOrder.MoldID,
				WarehouseID: workOrder.WarehouseID,
				StockType:   1, // 1=入库
				OperateTime: workOrder.ActualEnd,
				Quantity:    1,
				Remark:      fmt.Sprintf("工单 %s 入库", workOrder.WorkCode),
			}

			// 在事务中创建库存记录
			if err := tx.Create(stock).Error; err != nil {
				return fmt.Errorf("自动创建入库记录失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	// 3. 事务提交后，执行状态同步（异步执行，避免阻塞）
	// 注意：这里不再返回错误，因为主流程已成功
	workOrderCopy := *workOrder // 复制一份，避免闭包引用问题
	utils.RetryAsync(func() error {
		return s.statusSyncService.OnWorkOrderStatusChanged(&workOrderCopy)
	}, nil, "工单状态同步")

	return nil
}

func (s *MoldWorkOrderService) DeleteMoldWorkOrder(id int64) error {
	workOrder, err := s.moldWorkOrderDAO.FindByID(id)
	if err := utils.EnsureExists(workOrder, err, "工单"); err != nil {
		return err
	}

	// 已开始的工单（状态 != 未开始 或已取消）不允许删除，即使是超级管理员也不行
	if workOrder.Status != constant.WorkOrderStatusNotStarted {
		return errors.New("工单已进入生产阶段，禁止删除，请使用取消功能")
	}

	// 已取消的工单也不允许删除
	if workOrder.IsCancelled == 1 {
		return errors.New("工单已取消，禁止删除")
	}

	// 🔒 强约束：如果该模具已有修模/改造工单在进行中，禁止删除（即使未开始）
	// 如果当前工单本身就是修模/改造工单，则不锁定
	if workOrder.MoldID != nil && workOrder.OrderType != constant.MoldOrderTypeRepair && workOrder.OrderType != constant.MoldOrderTypeModification {
		var activeWorkOrderCount int64
		dao.DB.Model(&entity.MoldWorkOrder{}).
			Where("mold_id = ? AND id != ? AND status >= ? AND status < ? AND is_cancelled = 0 AND (order_type = ? OR order_type = ?)",
				*workOrder.MoldID, workOrder.ID, constant.WorkOrderStatusNotStarted, constant.WorkOrderStatusDelivered,
				constant.MoldOrderTypeRepair, constant.MoldOrderTypeModification).
			Count(&activeWorkOrderCount)
		if activeWorkOrderCount > 0 {
			return errors.New("该模具已有修模/改造工单在进行中，禁止删除此工单")
		}
	}

	type RelatedCounts struct {
		ProcessCount    int64 `gorm:"column:process_count"`
		TryTestCount    int64 `gorm:"column:try_test_count"`
		AcceptanceCount int64 `gorm:"column:acceptance_count"`
	}

	var counts RelatedCounts
	if err := dao.DB.Raw(`
		SELECT 
			(SELECT COUNT(*) FROM work_order_process WHERE work_order_id = ?) as process_count,
			(SELECT COUNT(*) FROM mold_try_test WHERE mold_work_id = ?) as try_test_count,
			(SELECT COUNT(*) FROM mold_acceptance WHERE mold_work_id = ?) as acceptance_count
	`, id, id, id).Scan(&counts).Error; err != nil {
		return errors.New("检查关联记录失败")
	}

	if counts.ProcessCount > 0 || counts.TryTestCount > 0 || counts.AcceptanceCount > 0 {
		return errors.New("工单已关联工序记录、试模记录或验收记录，禁止删除")
	}

	return s.moldWorkOrderDAO.Delete(id)
}

// CancelMoldWorkOrder 取消工单（软删除，标记取消状态并锁定）
func (s *MoldWorkOrderService) CancelMoldWorkOrder(id int64, reason string, userID *int64) error {
	workOrder, err := s.moldWorkOrderDAO.FindByID(id)
	if err := utils.EnsureExists(workOrder, err, "工单"); err != nil {
		return err
	}

	// 已取消的工单不允许重复取消
	if workOrder.IsCancelled == 1 {
		return errors.New("工单已取消，无需重复操作")
	}

	// 已完成、已试模及之后的工单不允许取消
	if workOrder.Status == constant.WorkOrderStatusCompleted ||
		workOrder.Status == constant.WorkOrderStatusTested ||
		workOrder.Status == constant.WorkOrderStatusStockedIn ||
		workOrder.Status == constant.WorkOrderStatusDelivered {
		return errors.New("工单已进入后期阶段，不允许取消")
	}

	// 🔒 使用事务确保数据一致性：更新工单 + 状态同步
	now := time.Now()
	workOrder.Status = constant.WorkOrderStatusCancelled
	workOrder.IsCancelled = 1
	workOrder.CancelTime = &now
	workOrder.CancelReason = reason
	workOrder.CancelUserID = userID

	// 在事务中更新工单
	err = dao.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entity.MoldWorkOrder{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":         workOrder.Status,
			"is_cancelled":   workOrder.IsCancelled,
			"cancel_time":    workOrder.CancelTime,
			"cancel_reason":  workOrder.CancelReason,
			"cancel_user_id": workOrder.CancelUserID,
		}).Error; err != nil {
			return fmt.Errorf("取消工单失败: %w", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	// 事务提交后，异步执行状态同步（避免阻塞）
	workOrderCopy := *workOrder // 复制一份，避免闭包引用问题
	utils.RetryAsync(func() error {
		return s.statusSyncService.OnWorkOrderStatusChanged(&workOrderCopy)
	}, nil, "取消工单后状态同步")

	return nil
}

func (s *MoldWorkOrderService) UpdateMoldWorkOrderStatus(id int64, status int8) error {
	return s.moldWorkOrderDAO.UpdateStatus(id, status)
}

// GetMoldWorkOrdersByIDs 根据ID列表批量查询工单
func (s *MoldWorkOrderService) GetMoldWorkOrdersByIDs(ids []int64) ([]*entity.MoldWorkOrder, error) {
	return s.moldWorkOrderDAO.FindByIDs(ids)
}

// CheckOutsourcingStatus 检查工单是否有未完成的外发记录
func (s *MoldWorkOrderService) CheckOutsourcingStatus(workOrderID int64) (bool, *entity.OutsourcingRecord, error) {
	outsourcingDAO := dao.NewOutsourcingRecordDAO()
	record, err := outsourcingDAO.FindByWorkOrderID(workOrderID)
	if err != nil {
		// 没有外发记录
		return false, nil, nil
	}

	if record.Status == "PENDING" || record.Status == "SENT" || record.Status == "PRODUCING" {
		return true, record, nil
	}

	// 外发已完成(RETURNED=已返回，ACCEPTED=已验收，CANCELLED=已取消)
	return false, nil, nil
}

// CanCancelWorkOrder 检查工单是否可以取消
func (s *MoldWorkOrderService) CanCancelWorkOrder(workOrderID int64) (bool, string, error) {
	workOrder, err := s.moldWorkOrderDAO.FindByID(workOrderID)
	if err != nil {
		return false, "", errors.New("工单不存在")
	}

	// 只有未开始和加工中的工单可以取消
	if workOrder.Status != constant.WorkOrderStatusNotStarted &&
		workOrder.Status != constant.WorkOrderStatusProcessing {
		return false, "工单已进入后期阶段，不能取消", nil
	}

	hasUnfinishedOutsourcing, record, _ := s.CheckOutsourcingStatus(workOrderID)
	if hasUnfinishedOutsourcing {
		statusText := map[string]string{
			"PENDING":   "待外发",
			"SENT":      "已外发",
			"PRODUCING": "生产中",
		}
		return false, fmt.Sprintf("工单有%s状态的外发记录(%s)，请先处理外发", statusText[record.Status], record.RecordCode), nil
	}

	return true, "", nil
}

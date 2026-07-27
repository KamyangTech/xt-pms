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
)

type MoldOrderService struct {
	moldOrderDAO      *dao.MoldOrderDAO
	moldOrderItemDAO  *dao.MoldOrderItemDAO
	moldWorkOrderDAO  *dao.MoldWorkOrderDAO
	statusSyncService *MoldStatusSyncService
	moldService       *MoldService
	workOrderService  *MoldWorkOrderService
}

func NewMoldOrderService() *MoldOrderService {
	return &MoldOrderService{
		moldOrderDAO:      dao.NewMoldOrderDAO(),
		moldOrderItemDAO:  dao.NewMoldOrderItemDAO(),
		moldWorkOrderDAO:  dao.NewMoldWorkOrderDAO(),
		statusSyncService: NewMoldStatusSyncService(),
		moldService:       NewMoldService(),
		workOrderService:  NewMoldWorkOrderService(),
	}
}

// GenerateOrderCode 生成订单号
// 格式：D-YYYYMMDD-NNN
// D: Order订单固定前缀
func (s *MoldOrderService) GenerateOrderCode() (string, error) {
	now := time.Now()
	dateStr := now.Format("060102") // YYMMDD 6位日期

	// 固定前缀 MO (Mold Order)
	prefix := "MO"

	// 构造查询前缀：MO-日期-
	queryPrefix := fmt.Sprintf("%s-%s-", prefix, dateStr)

	// 查找该日期下的最大序列号
	maxSeq, err := s.moldOrderDAO.FindMaxOrderSequenceByPrefix(queryPrefix)
	if err != nil {
		return "", fmt.Errorf("查询最大序列号失败: %v", err)
	}

	newSeq := maxSeq + 1
	orderCode := fmt.Sprintf("%s%02d", queryPrefix, newSeq)

	// 确保唯一性（极端情况下并发创建）
	existing, _ := s.moldOrderDAO.FindByCode(orderCode)
	if existing != nil {
		// 如果已存在，继续尝试下一个
		newSeq++
		orderCode = fmt.Sprintf("%s%02d", queryPrefix, newSeq)
	}

	return orderCode, nil
}

// CreateMoldOrder 创建模具订单（支持一单多款）
func (s *MoldOrderService) CreateMoldOrder(orderDTO dto.MoldOrderDTO) (*entity.MoldOrder, error) {
	// 1. 生成或校验订单号
	var orderCode string
	if orderDTO.OrderCode == "" {
		var err error
		orderCode, err = s.GenerateOrderCode()
		if err != nil {
			return nil, fmt.Errorf("生成订单号失败: %v", err)
		}
	} else {
		orderCode = orderDTO.OrderCode
		existing, _ := s.moldOrderDAO.FindByCode(orderCode)
		if existing != nil {
			return nil, errors.New("订单编号已存在")
		}
	}

	// 2. 准备订单基础信息
	var deliverTime *time.Time
	if orderDTO.DeliverTime != nil && !orderDTO.DeliverTime.Time.IsZero() {
		t := orderDTO.DeliverTime.Time
		deliverTime = &t
	}

	var createUserID *int64
	if orderDTO.CreateUserID > 0 {
		createUserID = &orderDTO.CreateUserID
	}

	order := &entity.MoldOrder{
		OrderCode:        orderCode,
		CustomerID:       orderDTO.CustomerID,
		OrderType:        orderDTO.OrderType,
		CreateUserID:     createUserID,
		Priority:         orderDTO.Priority,
		PlanStartDate:    orderDTO.PlanStartDate, // 直接赋值（都是 *dto.Date 类型）
		PlanEndDate:      orderDTO.PlanEndDate,   // 直接赋值（都是 *dto.Date 类型）
		TotalAmount:      orderDTO.TotalAmount,
		AdditionalAmount: orderDTO.AdditionalAmount,
		Status:           orderDTO.Status,
		DeliverTime:      deliverTime,
		Remark:           orderDTO.Remark,
	}
	if order.Status == "" {
		order.Status = constant.MoldOrderStatusNotStarted
	}
	if order.OrderType == "" {
		order.OrderType = constant.MoldOrderTypeNew
	}

	// 3. 开启事务处理多模具逻辑
	var createdOrder *entity.MoldOrder
	err := dao.DB.Transaction(func(tx *gorm.DB) error {
		// 3.1 创建订单主表
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		createdOrder = order

		// 3.2 处理明细项（Items）
		if len(orderDTO.Items) > 0 {
			for _, item := range orderDTO.Items {
				if err := s.createOrderItemInTx(tx, createdOrder, item); err != nil {
					return err
				}
			}
		} else if orderDTO.AutoCreateWorkOrder && orderDTO.WorkOrderInfo != nil && orderDTO.MoldInfo != nil {
			// 兼容旧版单个工单创建逻辑
			if err := s.createLegacyOrderItemInTx(tx, createdOrder, orderDTO); err != nil {
				return err
			}
		}

		// ✨ 自动创建应收账款/应付账款（按款项归属分流）- 使用统一方法
		if createdOrder.TotalAmount > 0 {
			if orderDTO.PaymentPlan != nil && orderDTO.PaymentPlan.Enabled && len(orderDTO.PaymentPlan.Items) > 0 {
				// 按款项归属分流：模具款→应收，外发应付款→应付
				var receivableItems []dto.PaymentItemDTO
				for _, item := range orderDTO.PaymentPlan.Items {
					if item.SourceType == "outsource" {
						// 外发应付款 → 创建应付账款
						// 从模具工单的外发记录中查找加工商ID
						var processorID int64
						var outsourceRecordCode string
						var outsourceRecordID *int64
						var workOrders []entity.MoldWorkOrder
						if err := tx.Where("mold_order_id = ? AND work_type = 'OUTSOURCED'", createdOrder.ID).Find(&workOrders).Error; err == nil && len(workOrders) > 0 {
							for _, wo := range workOrders {
								var outsRecord entity.OutsourcingRecord
								if err := tx.Where("work_order_id = ?", wo.ID).First(&outsRecord).Error; err == nil && wo.ProcessorID != nil {
									processorID = *wo.ProcessorID
									outsourceRecordCode = outsRecord.RecordCode
									outsourceRecordID = &outsRecord.ID
									break
								}
							}
						}

						if processorID > 0 {
							dueDate := parseDueDate(item.DueDate)
							ap := &entity.AccountsPayable{
								SupplierID: processorID, OrderID: outsourceRecordID,
								OrderType: "outsourcing", PaymentType: item.PaymentType,
								SourceType: "mold", OrderCode: outsourceRecordCode,
								Amount: item.CalculatedAmount, DueDate: dueDate,
								Status: "PENDING",
								Remark: fmt.Sprintf("模具订单 %s 外发款项：%s", createdOrder.OrderCode, item.PaymentType),
							}
							if err := tx.Create(ap).Error; err != nil {
								return fmt.Errorf("创建外发应付账款失败: %w", err)
							}
						}
					} else {
						receivableItems = append(receivableItems, item)
					}
				}

				// 统一调用公共方法创建应收账款
				if err := SyncReceivablesFromPaymentPlan(tx, createdOrder.CustomerID, createdOrder.ID, "mold_order", createdOrder.OrderCode, receivableItems); err != nil {
					return err
				}
			} else {
				// 使用默认的单笔应收
				if err := SyncReceivablesFromPaymentPlan(tx, createdOrder.CustomerID, createdOrder.ID, "mold_order", createdOrder.OrderCode,
					[]dto.PaymentItemDTO{{PaymentType: "total", CalculatedAmount: createdOrder.TotalAmount, SourceType: "mold"}},
				); err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdOrder, nil
}

// createOrderItemInTx 在事务中创建单个订单明细（模具+工单）
func (s *MoldOrderService) createOrderItemInTx(tx *gorm.DB, order *entity.MoldOrder, item dto.MoldOrderItemDTO) error {
	var moldID int64

	// 根据订单类型处理模具档案
	if order.OrderType == constant.MoldOrderTypeRepair || order.OrderType == constant.MoldOrderTypeModification {
		// 修模/改模订单：关联已有模具档案
		if item.MoldID == nil || *item.MoldID == 0 {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: "修模/改模订单必须指定已有模具ID"}
		}
		moldID = *item.MoldID
		var mold entity.Mold
		if err := tx.First(&mold, moldID).Error; err != nil {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: "指定的模具档案不存在"}
		}

		// "进行中"仅指未开始或加工中，已试模/已完成/已入库/已交付/已取消均允许创建新订单
		var activeWorkOrderCount int64
		tx.Model(&entity.MoldWorkOrder{}).
			Where("mold_id = ? AND status IN (?, ?)",
				moldID,
				constant.WorkOrderStatusNotStarted,
				constant.WorkOrderStatusProcessing).
			Count(&activeWorkOrderCount)
		if activeWorkOrderCount > 0 {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: "该模具已有进行中的工单，无法创建新的修模/改模订单。请等待当前工单完成后再创建"}
		}

		// 校验：修模/改模必须有已完成的新模工单（仅第一次修/改需要新模完成）
		var completedNewMoldWorkOrderCount int64
		tx.Model(&entity.MoldWorkOrder{}).
			Where("mold_id = ? AND order_type = ? AND status IN (?, ?, ?, ?)", moldID, constant.MoldOrderTypeNew,
				constant.WorkOrderStatusCompleted, constant.WorkOrderStatusTested, constant.WorkOrderStatusStockedIn, constant.WorkOrderStatusDelivered).
			Count(&completedNewMoldWorkOrderCount)
		if completedNewMoldWorkOrderCount == 0 {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: "该模具的新模制造工单尚未完成，无法创建修模/改模订单。请等待新模制造完成后再创建修模/改模订单"}
		}
	} else {
		// 新模订单：创建新模具档案
		// ✨ 校验模具编号必填
		if item.MoldCode == "" {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: "模具编号不能为空"}
		}

		// ✨ 校验模具编号唯一性（防止数据库已有重复）
		var existingMold entity.Mold
		if err := tx.Where("mold_code = ?", item.MoldCode).First(&existingMold).Error; err == nil {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: fmt.Sprintf("模具编号 [%s] 已存在，请使用其他编号", item.MoldCode)}
		}

		// ✨ 校验模具名称必填
		if item.MoldName == "" {
			return &response.HttpError{Code: response.BAD_REQUEST, Message: "模具名称不能为空"}
		}

		mold := &entity.Mold{
			MoldCode:        item.MoldCode,
			MoldName:        item.MoldName,
			CustomerID:      order.CustomerID,
			LifeCycleStatus: constant.MoldStatusDesigning,
			MoldType:        item.MoldType,
			MaterialType:    item.MaterialType,
			CavityCount:     item.CavityCount,
			Remark:          item.Remark,
		}
		if err := tx.Create(mold).Error; err != nil {
			return fmt.Errorf("创建模具档案失败: %v", err)
		}
		moldID = mold.ID
	}

	// 2. 创建工单
	var planStart, planEnd *time.Time
	if item.PlanStart != nil && !item.PlanStart.Time.IsZero() {
		t := item.PlanStart.Time
		planStart = &t
	}
	if item.PlanEnd != nil && !item.PlanEnd.Time.IsZero() {
		t := item.PlanEnd.Time
		planEnd = &t
	}

	workOrder := &entity.MoldWorkOrder{
		MoldOrderID:          &order.ID,
		OrderType:            order.OrderType,
		MoldID:               &moldID,
		PlanStart:            planStart,
		PlanEnd:              planEnd,
		Status:               constant.WorkOrderStatusNotStarted,
		WorkType:             item.WorkType,    // ✨ 设置生产方式
		ProcessorID:          item.ProcessorID, // ✨ 设置加工商ID
		OutsourcePrice:       0,                // 默认0
		OutsourceTotalAmount: 0,                // 默认0
	}
	// ✨ 如果是外发，设置外发价格
	if item.WorkType == "OUTSOURCED" && item.OutsourcePrice != nil {
		workOrder.OutsourcePrice = *item.OutsourcePrice
		workOrder.OutsourceTotalAmount = *item.OutsourcePrice
	}
	// 生成工单号（在事务中生成以避免并发冲突）
	woService := NewMoldWorkOrderService()
	code, err := woService.GenerateWorkCodeWithTx(tx)
	if err != nil {
		return err
	}
	workOrder.WorkCode = code

	if err := tx.Create(workOrder).Error; err != nil {
		return fmt.Errorf("创建工单失败: %v", err)
	}

	// 3. 初始化该模具的款项记录（如果指定了金额）
	if item.Amount > 0 {
		payment := &entity.MoldOrderPayment{
			OrderID:        order.ID,
			WorkOrderID:    &workOrder.ID,
			ContractAmount: item.Amount,
			ShouldAmount:   item.Amount,
			PaymentType:    "首款",
			Status:         1,
		}
		tx.Create(payment)
	}

	// 4. 为工单创建工序关联
	wopService := NewWorkOrderProcessService()
	if err := wopService.CreateWorkOrderProcessesInTx(tx, workOrder.ID); err != nil {
		return fmt.Errorf("创建工单工序失败: %v", err)
	}

	// 5. 更新模具档案的关联字段（latest_order_id 和 latest_work_order_id）
	if err := tx.Model(&entity.Mold{}).Where("id = ?", moldID).Updates(map[string]interface{}{
		"latest_order_id":      order.ID,
		"latest_work_order_id": workOrder.ID,
	}).Error; err != nil {
		return fmt.Errorf("更新模具关联字段失败: %v", err)
	}

	// 6. 如果是外发工单，自动创建外发记录
	if workOrder.WorkType == "OUTSOURCED" && workOrder.ProcessorID != nil {
		logger.Info("检测到外发工单，开始创建外发记录",
			zap.Int64("work_order_id", workOrder.ID),
			zap.Int64("processor_id", *workOrder.ProcessorID))

		outsourcingService := NewOutsourcingService()
		recordCode, err := outsourcingService.GenerateRecordCodeWithDB(tx)
		if err == nil {
			outsourcingRecord := &entity.OutsourcingRecord{
				RecordCode:         recordCode,
				OrderID:            nil, // ✨ 模具外发不需要关联生产订单
				WorkOrderID:        &workOrder.ID,
				OutsourceType:      "MOLD",
				ProcessorID:        *workOrder.ProcessorID,
				MoldID:             workOrder.MoldID,
				Quantity:           1, // 模具外发数量为1
				UnitPrice:          workOrder.OutsourcePrice,
				TotalAmount:        workOrder.OutsourceTotalAmount,
				OutsourceDate:      workOrder.PlanStart,
				ExpectedReturnDate: workOrder.PlanEnd,
				Status:             "PENDING",
				CreateUserID:       nil,
			}
			if err := tx.Create(outsourcingRecord).Error; err == nil {
				logger.Info("模具外发记录创建成功",
					zap.Int64("work_order_id", workOrder.ID),
					zap.String("record_code", recordCode))

				// ✨ 自动创建应付账款（在事务提交后执行）
				if outsourcingRecord.TotalAmount > 0 {
					apService := NewAccountsPayableService()
					apDAO := dao.NewAccountsPayableDAO()

					existingAP, _ := apDAO.FindByOrderID(outsourcingRecord.ID, "outsourcing")
				if existingAP == nil {
					now := time.Now()
					dueDate := &dto.Date{Time: now.AddDate(0, 0, 30)}

					ap := &entity.AccountsPayable{
						SupplierID:     *workOrder.ProcessorID,
						OrderID:        &outsourcingRecord.ID,
						OrderType:      "outsourcing",
						Amount:         outsourcingRecord.TotalAmount,
						DueDate:        dueDate,
							Status:  "PENDING",
							Remark:  fmt.Sprintf("模具外发记录 %s 创建时自动生成", recordCode),
						}

						if err := apService.CreatePayable(ap); err != nil {
							logger.Error("创建应付账款失败", zap.Error(err))
						} else {
							logger.Info("应付账款创建成功",
								zap.Int64("ap_id", ap.ID),
								zap.Int64("record_id", outsourcingRecord.ID),
								zap.Float64("amount", outsourcingRecord.TotalAmount))
						}
					}
				}
			} else {
				logger.Error("创建模具外发记录失败", zap.Error(err))
				// 不阻断流程，仅记录日志
			}
		} else {
			logger.Error("生成外发单号失败", zap.Error(err))
		}
	}

	return nil
}

// createLegacyOrderItemInTx 兼容旧版单个工单创建逻辑
func (s *MoldOrderService) createLegacyOrderItemInTx(tx *gorm.DB, order *entity.MoldOrder, orderDTO dto.MoldOrderDTO) error {
	// 复用原有逻辑，但放入事务中执行
	moldInfo := &entity.Mold{
		MoldCode:        orderDTO.MoldInfo.MoldCode,
		MoldName:        orderDTO.MoldInfo.MoldName,
		CustomerID:      orderDTO.CustomerID,
		LifeCycleStatus: constant.MoldStatusDesigning,
		MoldType:        orderDTO.MoldInfo.MoldType,
		MaterialType:    orderDTO.MoldInfo.MaterialType,
		CavityCount:     orderDTO.MoldInfo.CavityCount,
		Remark:          orderDTO.MoldInfo.Remark,
	}
	if err := tx.Create(moldInfo).Error; err != nil {
		return err
	}

	var workPlanStart, workPlanEnd *time.Time
	if orderDTO.WorkOrderInfo.PlanStart != nil && !orderDTO.WorkOrderInfo.PlanStart.Time.IsZero() {
		t := orderDTO.WorkOrderInfo.PlanStart.Time
		workPlanStart = &t
	}
	if orderDTO.WorkOrderInfo.PlanEnd != nil && !orderDTO.WorkOrderInfo.PlanEnd.Time.IsZero() {
		t := orderDTO.WorkOrderInfo.PlanEnd.Time
		workPlanEnd = &t
	}

	workOrderInfo := &entity.MoldWorkOrder{
		MoldOrderID: &order.ID,
		OrderType:   order.OrderType,
		MoldID:      &moldInfo.ID,
		PlanStart:   workPlanStart,
		PlanEnd:     workPlanEnd,
		Status:      constant.WorkOrderStatusNotStarted,
		Remark:      orderDTO.WorkOrderInfo.Remark,
	}

	woService := NewMoldWorkOrderService()
	code, err := woService.GenerateWorkCodeWithTx(tx)
	if err != nil {
		return err
	}
	workOrderInfo.WorkCode = code

	if err := tx.Create(workOrderInfo).Error; err != nil {
		return err
	}

	wopService := NewWorkOrderProcessService()
	if err := wopService.CreateWorkOrderProcessesInTx(tx, workOrderInfo.ID); err != nil {
		return err
	}

	return tx.Model(&entity.Mold{}).Where("id = ?", moldInfo.ID).Updates(map[string]interface{}{
		"latest_order_id":      order.ID,
		"latest_work_order_id": workOrderInfo.ID,
	}).Error
}

// GetMoldOrderByID 根据 ID 获取模具订单
func (s *MoldOrderService) GetMoldOrderByID(id int64) (*entity.MoldOrder, error) {
	order, err := s.moldOrderDAO.FindByID(id)
	if err != nil {
		return nil, err
	}

	items, err := s.moldOrderItemDAO.FindByOrderID(id)
	if err != nil {
		// 记录错误但不中断流程
		logger.Warn("加载订单明细失败", zap.Error(err), zap.Int64("order_id", id))
	} else {
		// 如果没有明细，但订单有金额（旧版订单），记录警告
		if len(items) == 0 && order.TotalAmount > 0 {
			logger.Warn("订单有金额但无明细数据（旧版订单）",
				zap.Int64("order_id", id),
				zap.Float64("total_amount", order.TotalAmount))
		}

		// 【优化】批量查询该订单的所有工单，避免 N+1 循环查询
		workOrders, err := s.moldWorkOrderDAO.FindByMoldOrderID(id)
		if err == nil {
			workOrderMap := make(map[int64]*entity.MoldWorkOrder, len(workOrders))
			for _, wo := range workOrders {
				if wo.MoldID != nil {
					workOrderMap[*wo.MoldID] = wo
				}
			}

			for _, item := range items {
				if wo, ok := workOrderMap[item.MoldID]; ok {
					item.WorkType = wo.WorkType
					item.ProcessorID = wo.ProcessorID
					item.OutsourcePrice = wo.OutsourcePrice
				}
			}
		}

		order.Items = items
	}

	// 如果是旧版订单（没有明细但有总金额），尝试从工单中获取模具信息
	if len(order.Items) == 0 && order.TotalAmount > 0 {
		// 通过订单ID查询所有工单（优化：只查询必要的字段）
		var workOrders []*entity.MoldWorkOrder
		if err := dao.DB.Select("id, mold_id").Where("mold_order_id = ?", id).Order("id ASC").Find(&workOrders).Error; err == nil && len(workOrders) > 0 {
			moldIDs := make([]int64, 0)
			for _, wo := range workOrders {
				if wo.MoldID != nil && *wo.MoldID > 0 {
					moldIDs = append(moldIDs, *wo.MoldID)
				}
			}

			// 批量查询模具信息
			var molds []*entity.Mold
			if len(moldIDs) > 0 {
				dao.DB.Where("id IN ?", moldIDs).Find(&molds)
			}

			moldMap := make(map[int64]*entity.Mold)
			for _, m := range molds {
				moldMap[m.ID] = m
			}

			// 为每个工单构造一个虚拟的明细项
			var virtualItems []*entity.MoldOrderItem
			for i, wo := range workOrders {
				if wo.MoldID != nil {
					if mold, ok := moldMap[*wo.MoldID]; ok {
						virtualItem := &entity.MoldOrderItem{
							OrderID: order.ID,
							MoldID:  mold.ID,
							Amount:  order.TotalAmount, // 注意：旧版订单可能所有模具共用一个总金额
							Sort:    i,
							Mold:    mold,
						}
						virtualItems = append(virtualItems, virtualItem)
					}
				}
			}
			if len(virtualItems) > 0 {
				order.Items = virtualItems
			}
		}
	}

	return order, nil
}

// GetAllMoldOrders 获取所有模具订单
func (s *MoldOrderService) GetAllMoldOrders() ([]*entity.MoldOrder, error) {
	return s.moldOrderDAO.FindAll()
}

// GetMoldOrdersByPage 分页获取模具订单
func (s *MoldOrderService) GetMoldOrdersByPage(page, pageSize int) ([]*entity.MoldOrder, int64, error) {
	result, err := s.moldOrderDAO.FindPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	orders := result.List.([]*entity.MoldOrder)
	// 填充 MasterOrderCode
	for _, order := range orders {
		if order.MasterOrder != nil {
			order.MasterOrderCode = order.MasterOrder.OrderCode
		}
	}
	return orders, result.Total, nil
}

// GetMoldOrdersByPageWithFilter 带筛选条件的分页获取模具订单
func (s *MoldOrderService) GetMoldOrdersByPageWithFilter(page, pageSize int, orderCode string, moldCode string, customerID *int64, status *string, createTimeStart, createTimeEnd, planEndDateStart, planEndDateEnd string) ([]*entity.MoldOrder, int64, error) {
	offset := (page - 1) * pageSize
	return s.moldOrderDAO.FindPageWithFilter(offset, pageSize, orderCode, moldCode, customerID, status, createTimeStart, createTimeEnd, planEndDateStart, planEndDateEnd)
}

// UpdateMoldOrder 更新模具订单
func (s *MoldOrderService) UpdateMoldOrder(order *entity.MoldOrder) error {
	return s.UpdateMoldOrderWithItems(order, nil, 0)
}

// UpdateMoldOrderWithItems 更新模具订单及工单信息
func (s *MoldOrderService) UpdateMoldOrderWithItems(order *entity.MoldOrder, items []dto.MoldOrderItemDTO, userID int64) error {
	existingOrder, err := s.moldOrderDAO.FindByID(order.ID)
	if err := utils.EnsureExists(existingOrder, err, "订单"); err != nil {
		return err
	}

	// 开启事务
	tx := dao.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Model(&entity.MoldOrder{}).Where("id = ?", order.ID).Updates(order).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 处理工单更新（如果有items）
	if items != nil && len(items) > 0 {
		logger.Info("开始处理工单更新", zap.Int("items_count", len(items)), zap.Int64("order_id", order.ID))
		outsourcingService := NewOutsourcingService()

		// 【优化】批量查询该订单的所有工单，避免 N+1 循环查询
		allWorkOrders, err := s.moldWorkOrderDAO.FindByMoldOrderID(order.ID)
		workOrderMap := make(map[int64]*entity.MoldWorkOrder, len(allWorkOrders))
		if err == nil {
			for _, wo := range allWorkOrders {
				if wo.MoldID != nil {
					workOrderMap[*wo.MoldID] = wo
				}
			}
		}

		for idx, item := range items {
			var moldID int64
			if item.MoldID != nil {
				moldID = *item.MoldID
			}
			logger.Info("处理工单项",
				zap.Int("index", idx),
				zap.Int64("mold_id", moldID),
				zap.String("work_type", item.WorkType),
				zap.Any("processor_id", item.ProcessorID))

			// 从 map 中查找对应工单（替代 N 次独立查询）
			workOrder, ok := workOrderMap[moldID]
			if !ok {
				logger.Warn("未找到对应工单，跳过",
					zap.Int64("order_id", order.ID),
					zap.Int64("mold_id", moldID))
				continue
			}

			logger.Info("找到工单，开始更新",
				zap.Int64("work_order_id", workOrder.ID),
				zap.String("old_work_type", workOrder.WorkType),
				zap.String("new_work_type", item.WorkType))

			if item.WorkType != "" {
				workOrder.WorkType = item.WorkType
			}
			if item.ProcessorID != nil {
				workOrder.ProcessorID = item.ProcessorID
			}
			if item.OutsourcePrice != nil {
				workOrder.OutsourcePrice = *item.OutsourcePrice
				workOrder.OutsourceTotalAmount = *item.OutsourcePrice
			}

			if err := tx.Save(workOrder).Error; err != nil {
				tx.Rollback()
				return err
			}

			if workOrder.WorkType == "OUTSOURCED" && workOrder.ProcessorID != nil {
				outsourcePricePtr := &workOrder.OutsourcePrice
				if err := outsourcingService.CreateOutsourcingRecordForWorkOrder(tx, workOrder, workOrder.ProcessorID, outsourcePricePtr, &userID); err != nil {
					logger.Error("创建外发记录失败", zap.Error(err), zap.Int64("work_order_id", workOrder.ID))
					// 不阻断流程，仅记录日志
				}
			} else {
				// 切换回本厂模式，撤回外发记录（如果存在）
				reason := "生产方式变更为：本厂生产"
				if err := outsourcingService.CancelOutsourcingRecordByWorkOrderID(workOrder.ID, reason); err != nil {
					logger.Error("撤回外发记录失败", zap.Error(err), zap.Int64("work_order_id", workOrder.ID))
					// 不阻断流程，仅记录日志
				}
			}
		}
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 状态变更时同步到模具档案
	if err := s.statusSyncService.OnOrderStatusChanged(order); err != nil {
		return err
	}

	return nil
}

// DeleteMoldOrder 删除模具订单
func (s *MoldOrderService) DeleteMoldOrder(id int64, isAdmin bool) error {
	order, err := s.moldOrderDAO.FindByID(id)
	if err := utils.EnsureExists(order, err, "订单"); err != nil {
		return err
	}

	// 非超级管理员：已开始的订单不允许删除
	if !isAdmin && order.Status != constant.MoldOrderStatusNotStarted {
		return errors.New("订单已开始生产，普通用户无法删除，请使用取消功能或联系超级管理员")
	}

	// 超级管理员允许删除任何状态的订单

	if !isAdmin {
		var workOrders []*entity.MoldWorkOrder
		dao.DB.Where("mold_order_id = ?", id).Find(&workOrders)
		for _, wo := range workOrders {
			if wo.Status != constant.WorkOrderStatusNotStarted { // 工单已开始加工
				return errors.New("订单关联的工单已开始加工，普通用户无法删除，请联系超级管理员")
			}
		}
	}

	return s.moldOrderDAO.Delete(id)
}

// UpdateMoldOrderStatus 更新模具订单状态
func (s *MoldOrderService) UpdateMoldOrderStatus(id int64, status int8) error {
	return s.moldOrderDAO.UpdateStatus(id, status)
}

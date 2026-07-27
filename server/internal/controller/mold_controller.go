package controller

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type MoldController struct {
	moldService          *service.MoldService
	moldOrderService     *service.MoldOrderService
	moldBOMService       *service.MoldBOMService
	moldWorkOrderService *service.MoldWorkOrderService
	moldTryTestService   *service.MoldTryTestService
	moldStockService     *service.MoldStockService
	moldDeliveryService  *service.MoldDeliveryService
	statusFlowService    *service.StatusFlowService
}

func NewMoldController() *MoldController {
	return &MoldController{
		moldService:          service.NewMoldService(),
		moldOrderService:     service.NewMoldOrderService(),
		moldBOMService:       service.NewMoldBOMService(),
		moldWorkOrderService: service.NewMoldWorkOrderService(),
		moldTryTestService:   service.NewMoldTryTestService(),
		moldStockService:     service.NewMoldStockService(),
		moldDeliveryService:  service.NewMoldDeliveryService(),
		statusFlowService:    service.NewStatusFlowService(),
	}
}

// GetMolds 获取模具列表
func (c *MoldController) GetMolds(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	moldCode := ctx.Query("mold_code")
	moldName := ctx.Query("mold_name")
	customerIDStr := ctx.Query("customer_id")
	lifeCycleStatusStr := ctx.Query("life_cycle_status")
	moldType := ctx.Query("mold_type")
	moldProcess := ctx.Query("mold_process")
	materialType := ctx.Query("material_type")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	var customerID *int64
	if customerIDStr != "" {
		id, err := strconv.ParseInt(customerIDStr, 10, 64)
		if err == nil {
			customerID = &id
		}
	}

	molds, total, err := c.moldService.GetMoldsByPageWithFilter(page, pageSize, moldCode, moldName, customerID, lifeCycleStatusStr, moldType, moldProcess, materialType, startDate, endDate)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取模具列表失败")
		return
	}

	// 批量查询主图（避免N+1查询）
	moldIDs := make([]int64, 0, len(molds))
	for _, mold := range molds {
		moldIDs = append(moldIDs, mold.ID)
	}
	primaryImageMap, _ := c.moldService.GetPrimaryImagesByMoldIDs(moldIDs)

	// 为每个模具添加当前工序信息和主图（批量查询优化）
	enrichedList := c.enrichMoldListWithCurrentProcess(molds, primaryImageMap)

	response.Success(ctx, gin.H{
		"list":      enrichedList,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetMoldByID 根据 ID 获取模具
func (c *MoldController) GetMoldByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "参数错误")
		return
	}
	mold, err := c.moldService.GetMoldByID(id)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "模具不存在")
		return
	}

	// 添加当前正在进行的工序信息
	// 为单个模具使用旧的单个查询方法（兼容旧API）
	moldList := []*entity.Mold{mold}
	enrichedList := c.enrichMoldListWithCurrentProcess(moldList, map[int64]string{})
	if len(enrichedList) > 0 {
		result := enrichedList[0]
		timeline := c.buildManufacturingTimeline(mold)
		result["manufacturing_timeline"] = timeline
		response.Success(ctx, result)
	} else {
		response.Error(ctx, response.ERROR, "数据处理失败")
	}
}

// buildManufacturingTimeline 构建制造进度时间线
func (c *MoldController) buildManufacturingTimeline(mold *entity.Mold) gin.H {
	timeline := gin.H{
		"order_create":           mold.CreateTime, // 步骤1: 订单创建时间
		"start_manufacturing":    nil,             // 步骤2: 开始制造时间
		"manufacturing_complete": nil,             // 步骤3: 制造完成时间
		"try_test":               nil,             // 步骤4: 试模时间
		"storage":                nil,             // 步骤5: 入库时间
	}

	// 从工单获取实际开始时间和结束时间
	if mold.LatestWorkOrderID != nil {
		workOrderDAO := dao.NewMoldWorkOrderDAO()
		workOrder, err := workOrderDAO.FindByID(*mold.LatestWorkOrderID)
		if err == nil && workOrder != nil {
			// 步骤2: 开始制造 - 使用工单的实际开始时间或计划开始时间
			if workOrder.ActualStart != nil {
				timeline["start_manufacturing"] = workOrder.ActualStart
			} else if workOrder.PlanStart != nil {
				timeline["start_manufacturing"] = workOrder.PlanStart
			}

			// 步骤5: 入库 - 使用工单的实际结束时间
			if workOrder.ActualEnd != nil {
				timeline["storage"] = workOrder.ActualEnd
			}

			// 步骤4: 试模 - 查询工单关联的试模记录
			if mold.LifeCycleStatus == constant.MoldStatusTesting ||
				mold.LifeCycleStatus == constant.MoldStatusTested ||
				mold.LifeCycleStatus == constant.MoldStatusProducing ||
				mold.LifeCycleStatus == constant.MoldStatusRepairing ||
				mold.LifeCycleStatus == constant.MoldStatusInStock ||
				mold.LifeCycleStatus == constant.MoldStatusDelivered {
				tryTestDAO := dao.NewMoldTryTestDAO()
				tryTests, err := tryTestDAO.FindByWorkOrderID(workOrder.ID)
				if err == nil && len(tryTests) > 0 {
					// 使用最后一次试模的 test_time
					lastTest := tryTests[len(tryTests)-1]
					if lastTest.TestTime != nil {
						timeline["try_test"] = lastTest.TestTime
					} else {
						timeline["try_test"] = lastTest.CreateTime
					}
				} else {
					// 如果没有试模记录，用状态变更时间近似
					timeline["try_test"] = mold.UpdateTime
				}
			}
		}
	}

	// 步骤3: 制造完成 - 当状态为已完成或更后时，如果没有具体的完成时间，用状态变更的时间近似
	if mold.LifeCycleStatus == constant.MoldStatusTesting ||
		mold.LifeCycleStatus == constant.MoldStatusTested ||
		mold.LifeCycleStatus == constant.MoldStatusProducing ||
		mold.LifeCycleStatus == constant.MoldStatusRepairing ||
		mold.LifeCycleStatus == constant.MoldStatusInStock ||
		mold.LifeCycleStatus == constant.MoldStatusDelivered {
		timeline["manufacturing_complete"] = mold.UpdateTime
	}

	return timeline
}

// CreateMold 创建模具
func (c *MoldController) CreateMold(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建模具")
	ctx.Set("oper_log_type", 1) // 1=新增
	ctx.Set("oper_log_business_type", "mold")

	var moldDTO dto.MoldDTO
	if err := ctx.ShouldBindJSON(&moldDTO); err != nil {
		logger.Warn("模具创建参数绑定失败", zap.Error(err))
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}
	mold, err := c.moldService.CreateMold(moldDTO)
	if err != nil {
		// 业务校验错误(BadRequest) vs 系统错误(InternalServerError)
		var httpErr *response.HttpError
		if errors.As(err, &httpErr) {
			logger.Warn("模具创建业务校验失败", zap.Error(err))
		} else {
			logger.Error("创建模具失败", zap.Error(err))
		}
		handleServiceError(ctx, err)
		return
	}
	response.Success(ctx, mold)
}

// UpdateMold 更新模具
func (c *MoldController) UpdateMold(ctx *gin.Context) {
	ctx.Set("oper_log_title", "更新模具")
	ctx.Set("oper_log_type", 3) // 3=修改
	ctx.Set("oper_log_business_type", "mold")

	var moldDTO dto.MoldUpdateDTO
	if err := ctx.ShouldBindJSON(&moldDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}
	err := c.moldService.UpdateMold(moldDTO)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// DeleteMold 删除模具
func (c *MoldController) DeleteMold(ctx *gin.Context) {
	ctx.Set("oper_log_title", "删除模具")
	ctx.Set("oper_log_type", 2) // 2=删除
	ctx.Set("oper_log_business_type", "mold")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "参数错误")
		return
	}

	err = c.moldService.DeleteMold(id, false)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// enrichMoldListWithCurrentProcess 批量为模具列表添加当前工序信息（避免N+1查询）
func (c *MoldController) enrichMoldListWithCurrentProcess(molds []*entity.Mold, primaryImageMap map[int64]string) []gin.H {
	if len(molds) == 0 {
		return []gin.H{}
	}

	result := make([]gin.H, 0, len(molds))
	for _, mold := range molds {
		moldData := c.enrichMoldSimple(mold, primaryImageMap)
		result = append(result, moldData)
	}

	return result
}

// enrichMoldSimple 为模具添加基本扩展字段（不含旧工序数据）
func (c *MoldController) enrichMoldSimple(mold *entity.Mold, primaryImageMap map[int64]string) gin.H {
	result := gin.H{
		"id":                   mold.ID,
		"mold_code":            mold.MoldCode,
		"mold_name":            mold.MoldName,
		"customer_id":          mold.CustomerID,
		"latest_order_id":      mold.LatestOrderID,
		"latest_work_order_id": mold.LatestWorkOrderID,
		"total_orders":         mold.TotalOrders,
		"total_work_orders":    mold.TotalWorkOrders,
		"last_delivery_time":   mold.LastDeliveryTime,
		"production_date":      mold.ProductionDate,
		"life_cycle_status":    mold.LifeCycleStatus,
		"mold_type":            mold.MoldType,
		"mold_process":         mold.MoldProcess,
		"material_type":        mold.MaterialType,
		"cavity_count":         mold.CavityCount,
		"remark":               mold.Remark,
		"create_time":          mold.CreateTime,
		"update_time":          mold.UpdateTime,
		"customer":             mold.Customer,
		"images":               mold.Images,
		"primary_image_url":    primaryImageMap[mold.ID],
	}
	return result
}

// GetMoldOrders 获取模具订单列表
func (c *MoldController) GetMoldOrders(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	orderCode := ctx.Query("order_code")
	moldCode := ctx.Query("mold_code")
	customerIDStr := ctx.Query("customer_id")
	statusStr := ctx.Query("status")
	createTimeStart := ctx.Query("create_time_start")
	createTimeEnd := ctx.Query("create_time_end")
	dueSoon := ctx.Query("due_soon")

	var customerID *int64
	if customerIDStr != "" {
		id, err := strconv.ParseInt(customerIDStr, 10, 64)
		if err == nil {
			customerID = &id
		}
	}

	var status *string
	if statusStr != "" {
		// 尝试解析为数字（前端传递数字状态码），转换为数据库字符串状态码
		if s, err := strconv.ParseInt(statusStr, 10, 8); err == nil {
			statusCode := constant.MoldOrderStatusNumericToCode(int8(s))
			if statusCode != "" {
				status = &statusCode
			}
		} else {
			// 直接使用字符串状态码
			status = &statusStr
		}
	}

	// 到期预警：计算 plan_end_date 范围
	planEndDateStart := ""
	planEndDateEnd := ""
	if dueSoon != "" {
		if days, err := strconv.Atoi(dueSoon); err == nil && days > 0 {
			planEndDateStart = time.Now().Format("2006-01-02")
			planEndDateEnd = time.Now().AddDate(0, 0, days).Format("2006-01-02")
		}
	}

	orders, total, err := c.moldOrderService.GetMoldOrdersByPageWithFilter(page, pageSize, orderCode, moldCode, customerID, status, createTimeStart, createTimeEnd, planEndDateStart, planEndDateEnd)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取订单列表失败")
		return
	}

	// ✅ 批量查询所有订单的模具编号（避免N+1）
	orderIDs := make([]int64, 0, len(orders))
	for _, order := range orders {
		orderIDs = append(orderIDs, order.ID)
	}

	// 一次性查询所有订单的工单和模具信息
	if len(orderIDs) > 0 {
		var workOrders []*entity.MoldWorkOrder
		dao.DB.Preload("Mold").Where("mold_order_id IN ?", orderIDs).Find(&workOrders)

		// 按订单ID分组
		moldCodesMap := make(map[int64][]string)
		for _, wo := range workOrders {
			if wo.Mold != nil && wo.Mold.MoldCode != "" {
				moldCodesMap[*wo.MoldOrderID] = append(moldCodesMap[*wo.MoldOrderID], wo.Mold.MoldCode)
			}
		}

		// 填充到订单对象
		for _, order := range orders {
			order.MoldCodes = moldCodesMap[order.ID]
		}
	}

	// 注意：不在查询时自动同步状态，避免覆盖用户手动设置的状态
	// 状态同步应该在工单状态变更时触发（如试模、入库等操作）
	// 如需批量修复历史数据，请使用 FixOrderStatus 接口

	response.Success(ctx, gin.H{
		"list":      orders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// safeStatusPattern 仅允许大写字母和下划线，匹配所有业务状态常量格式
var safeStatusPattern = regexp.MustCompile(`^[A-Z][A-Z_]*$`)

// safeSQLStatus 校验 status 字符串是否合法，防止二阶 SQL 注入
// 合法的 status 格式示例：PROCESSING、IN_STOCK、PENDING_SCHEDULE
func safeSQLStatus(status string) bool {
	return safeStatusPattern.MatchString(status)
}

// batchSyncOrderStatus 批量同步订单状态（根据工单聚合，批量查询避免 N+1）
func (c *MoldController) batchSyncOrderStatus(orders []*entity.MoldOrder) {
	if len(orders) == 0 {
		return
	}

	orderIDs := make([]int64, 0, len(orders))
	orderMap := make(map[int64]*entity.MoldOrder)
	for _, order := range orders {
		if order.ID == 0 {
			continue
		}
		orderIDs = append(orderIDs, order.ID)
		orderMap[order.ID] = order
	}

	if len(orderIDs) == 0 {
		return
	}

	// 一次性批量查询所有订单的工单（避免 N+1）
	var allWorkOrders []*entity.MoldWorkOrder
	dao.DB.Where("mold_order_id IN ?", orderIDs).Find(&allWorkOrders)

	// 按订单 ID 分组
	workOrdersMap := make(map[int64][]*entity.MoldWorkOrder)
	for _, wo := range allWorkOrders {
		if wo.MoldOrderID != nil {
			workOrdersMap[*wo.MoldOrderID] = append(workOrdersMap[*wo.MoldOrderID], wo)
		}
	}

	type orderUpdate struct {
		ID     int64
		Status string
	}
	var updates []orderUpdate

	for _, orderID := range orderIDs {
		workOrders := workOrdersMap[orderID]
		if len(workOrders) == 0 {
			continue
		}

		// 聚合所有工单状态，取最未完成的状态（最小值）
		minStatus := workOrders[0].Status
		minOrder := c.statusFlowService.GetStatusOrder("work_order", minStatus)

		for _, wo := range workOrders[1:] {
			order := c.statusFlowService.GetStatusOrder("work_order", wo.Status)
			if order < minOrder {
				minStatus = wo.Status
				minOrder = order
			}
		}

		// 根据最未完成的工单状态计算订单状态
		mappedStatus := c.statusFlowService.GetMoldOrderStatusByWorkOrder(minStatus)
		if mappedStatus == nil {
			continue
		}
		newOrderStatus := *mappedStatus

		order := orderMap[orderID]
		if newOrderStatus != order.Status {
			// 更新内存中的对象
			order.Status = newOrderStatus
			updates = append(updates, orderUpdate{ID: orderID, Status: newOrderStatus})
		}
	}

	// 批量更新数据库（CASE WHEN 单条 SQL 替代 N+1 逐条 UPDATE）
	if len(updates) > 0 {
		var ids []int64
		var caseSQL strings.Builder
		caseSQL.WriteString("UPDATE mold_order SET status = CASE id ")
		for _, u := range updates {
			if !safeSQLStatus(u.Status) {
				logger.Warn("batchSyncOrderStatus 跳过非法状态值", zap.String("status", u.Status), zap.Int64("id", u.ID))
				continue
			}
			caseSQL.WriteString(fmt.Sprintf("WHEN %d THEN '%s' ", u.ID, u.Status))
			ids = append(ids, u.ID)
		}
		caseSQL.WriteString("END WHERE id IN (")
		for i, id := range ids {
			if i > 0 {
				caseSQL.WriteString(",")
			}
			caseSQL.WriteString(strconv.FormatInt(id, 10))
		}
		caseSQL.WriteString(")")
		if err := dao.DB.Exec(caseSQL.String()).Error; err != nil {
			logger.Error("批量更新订单状态失败", zap.Int("count", len(updates)), zap.Error(err))
		}
	}
}

// batchSyncWorkOrderStatusAsync 异步批量同步工单状态（根据工序进度）
func (c *MoldController) batchSyncWorkOrderStatusAsync(workOrders []*entity.MoldWorkOrder) {
	if len(workOrders) == 0 {
		return
	}

	workOrderIDs := make([]int64, 0, len(workOrders))
	for _, wo := range workOrders {
		if wo.ID > 0 {
			workOrderIDs = append(workOrderIDs, wo.ID)
		}
	}

	// 一次性查询所有工单的工序（避免 N+1 查询）
	var allProcesses []*entity.WorkOrderProcess
	dao.DB.Where("work_order_id IN ?", workOrderIDs).Find(&allProcesses)

	// 按工单 ID 分组
	processesByWorkOrder := make(map[int64][]*entity.WorkOrderProcess)
	for _, p := range allProcesses {
		processesByWorkOrder[p.WorkOrderID] = append(processesByWorkOrder[p.WorkOrderID], p)
	}

	// 批量更新
	updates := make([]map[string]interface{}, 0)
	for _, wo := range workOrders {
		if wo.ID == 0 {
			continue
		}

		workOrderProcesses := processesByWorkOrder[wo.ID]
		if len(workOrderProcesses) == 0 {
			continue
		}

		completedCount := 0
		for _, wop := range workOrderProcesses {
			if wop.IsCompleted == 1 {
				completedCount++
			}
		}

		newProgress := (completedCount * 100) / len(workOrderProcesses)

		// 根据进度计算新的工单状态
		var newStatus string
		if wo.Status == constant.WorkOrderStatusTested ||
			wo.Status == constant.WorkOrderStatusStockedIn ||
			wo.Status == constant.WorkOrderStatusDelivered {
			// 状态已经是试模及以上，不受工序进度影响
			newStatus = wo.Status
		} else {
			// 状态在已完成及以下，根据进度动态校准
			if newProgress == 0 {
				newStatus = constant.WorkOrderStatusNotStarted // 未开始
			} else if newProgress > 0 && newProgress < 100 {
				newStatus = constant.WorkOrderStatusProcessing // 加工中
			} else if newProgress == 100 {
				newStatus = constant.WorkOrderStatusCompleted // 已完成
			}
		}

		if newProgress != wo.Progress || newStatus != wo.Status {
			updates = append(updates, map[string]interface{}{
				"id":       wo.ID,
				"progress": newProgress,
				"status":   newStatus,
			})
		}
	}

	// 批量更新数据库（使用 CASE WHEN 单条 SQL 替代 N+1 逐条 UPDATE）
	if len(updates) > 0 {
		tx := dao.DB.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				logger.Error("batchSyncWorkOrderStatusAsync panic 回滚",
					zap.Any("panic", r))
			}
		}()

		var (
			ids              []int64
			progressCaseSQL  strings.Builder
			statusCaseSQL    strings.Builder
		)
		progressCaseSQL.WriteString("UPDATE mold_work_order SET progress = CASE id ")
		statusCaseSQL.WriteString("status = CASE id ")

		for _, update := range updates {
			id := update["id"].(int64)
			progress := update["progress"].(int)
			status := update["status"].(string)
			if !safeSQLStatus(status) {
				logger.Warn("batchSyncWorkOrderStatusAsync 跳过非法状态值", zap.String("status", status), zap.Int64("id", id))
				continue
			}
			progressCaseSQL.WriteString(fmt.Sprintf("WHEN %d THEN %d ", id, progress))
			statusCaseSQL.WriteString(fmt.Sprintf("WHEN %d THEN '%s' ", id, status))
			ids = append(ids, id)
		}

		var idStrs []string
		for _, id := range ids {
			idStrs = append(idStrs, strconv.FormatInt(id, 10))
		}

		sql := progressCaseSQL.String() + statusCaseSQL.String() + "END, update_time = NOW() WHERE id IN (" + strings.Join(idStrs, ",") + ")"
		if err := tx.Exec(sql).Error; err != nil {
			logger.Error("批量更新工单状态失败", zap.Int("count", len(updates)), zap.Error(err))
			tx.Rollback()
			return
		}

		if err := tx.Commit().Error; err != nil {
			logger.Error("提交工单状态更新事务失败", zap.Error(err))
		}
	}
}

// batchSyncWorkOrderStatus 同步批量同步工单状态（保留用于手动触发）
func (c *MoldController) batchSyncWorkOrderStatus(workOrders []*entity.MoldWorkOrder) {
	if len(workOrders) == 0 {
		return
	}

	workOrderIDs := make([]int64, 0, len(workOrders))
	for _, wo := range workOrders {
		if wo.ID > 0 {
			workOrderIDs = append(workOrderIDs, wo.ID)
		}
	}

	// 一次性查询所有工单的工序（避免 N+1 查询）
	var allProcesses []*entity.WorkOrderProcess
	dao.DB.Where("work_order_id IN ?", workOrderIDs).Find(&allProcesses)

	// 按工单 ID 分组
	processesByWorkOrder := make(map[int64][]*entity.WorkOrderProcess)
	for _, p := range allProcesses {
		processesByWorkOrder[p.WorkOrderID] = append(processesByWorkOrder[p.WorkOrderID], p)
	}

	// 批量更新
	updates := make([]map[string]interface{}, 0)
	for _, wo := range workOrders {
		if wo.ID == 0 {
			continue
		}

		workOrderProcesses := processesByWorkOrder[wo.ID]
		if len(workOrderProcesses) == 0 {
			continue
		}

		completedCount := 0
		for _, wop := range workOrderProcesses {
			if wop.IsCompleted == 1 {
				completedCount++
			}
		}

		newProgress := (completedCount * 100) / len(workOrderProcesses)

		// 根据进度计算新的工单状态
		var newStatus string
		if wo.Status == constant.WorkOrderStatusTested ||
			wo.Status == constant.WorkOrderStatusStockedIn ||
			wo.Status == constant.WorkOrderStatusDelivered {
			// 状态已经是试模及以上，不受工序进度影响
			newStatus = wo.Status
		} else {
			// 状态在已完成及以下，根据进度动态校准
			if newProgress == 0 {
				newStatus = constant.WorkOrderStatusNotStarted // 未开始
			} else if newProgress > 0 && newProgress < 100 {
				newStatus = constant.WorkOrderStatusProcessing // 加工中
			} else if newProgress == 100 {
				newStatus = constant.WorkOrderStatusCompleted // 已完成
			}
		}

		if newProgress != wo.Progress || newStatus != wo.Status {
			updates = append(updates, map[string]interface{}{
				"id":       wo.ID,
				"progress": newProgress,
				"status":   newStatus,
			})
		}
	}

	// 批量更新数据库（使用 CASE WHEN 单条 SQL 替代 N+1 逐条 UPDATE）
	if len(updates) > 0 {
		tx := dao.DB.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				logger.Error("batchSyncWorkOrderStatus panic 回滚",
					zap.Any("panic", r))
			}
		}()

		var (
			ids             []int64
			progressCaseSQL strings.Builder
			statusCaseSQL   strings.Builder
		)
		progressCaseSQL.WriteString("UPDATE mold_work_order SET progress = CASE id ")
		statusCaseSQL.WriteString("status = CASE id ")

		for _, update := range updates {
			id := update["id"].(int64)
			progress := update["progress"].(int)
			status := update["status"].(string)
			if !safeSQLStatus(status) {
				logger.Warn("batchSyncWorkOrderStatus 跳过非法状态值", zap.String("status", status), zap.Int64("id", id))
				continue
			}
			progressCaseSQL.WriteString(fmt.Sprintf("WHEN %d THEN %d ", id, progress))
			statusCaseSQL.WriteString(fmt.Sprintf("WHEN %d THEN '%s' ", id, status))
			ids = append(ids, id)
		}

		var idStrs []string
		for _, id := range ids {
			idStrs = append(idStrs, strconv.FormatInt(id, 10))
		}

		sql := progressCaseSQL.String() + statusCaseSQL.String() + "END, update_time = NOW() WHERE id IN (" + strings.Join(idStrs, ",") + ")"
		if err := tx.Exec(sql).Error; err != nil {
			logger.Error("批量更新工单状态失败", zap.Int("count", len(updates)), zap.Error(err))
			tx.Rollback()
			return
		}

		if err := tx.Commit().Error; err != nil {
			logger.Error("提交工单状态更新事务失败", zap.Error(err))
		}
	}
}

// FixOrderStatus 修复所有订单状态（手动触发）
func (c *MoldController) FixOrderStatus(ctx *gin.Context) {
	orders, _, err := c.moldOrderService.GetMoldOrdersByPageWithFilter(1, 10000, "", "", nil, nil, "", "", "", "")
	if err != nil {
		response.Error(ctx, response.ERROR, "获取订单列表失败")
		return
	}

	// 批量修复订单状态
	c.batchSyncOrderStatus(orders)

	response.Success(ctx, gin.H{
		"message": fmt.Sprintf("成功修复 %d 个订单状态", len(orders)),
	})
}

// CreateMoldOrder 创建模具订单
func (c *MoldController) CreateMoldOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建模具订单")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "mold_order")

	var orderDTO dto.MoldOrderDTO
	if err := ctx.ShouldBindJSON(&orderDTO); err != nil {
		logger.Error("模具订单创建参数绑定失败", zap.Error(err))
		errMsg := utils.FormatValidationError(err)
		response.Error(ctx, response.BAD_REQUEST, errMsg)
		return
	}
	userID := middleware.GetUserID(ctx)
	orderDTO.CreateUserID = userID

	if orderDTO.PlanStartDate != nil && orderDTO.PlanEndDate != nil {
		if err := validatePlanDates(orderDTO.PlanStartDate.Time, orderDTO.PlanEndDate.Time); err != nil {
			response.Error(ctx, response.BAD_REQUEST, err.Error())
			return
		}
	}

	order, err := c.moldOrderService.CreateMoldOrder(orderDTO)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	response.Success(ctx, order)

	// 通知：创建模具订单，查询订单明细获取所有模具编号（批量查询避免 N+1）
	items, _ := dao.NewMoldOrderItemDAO().FindByOrderID(order.ID)
	moldCount := len(items)
	moldCodes := make([]string, 0, moldCount)
	moldIDsToQuery := make([]int64, 0)
	for _, item := range items {
		if item.Mold != nil {
			moldCodes = append(moldCodes, item.Mold.MoldCode)
		} else if item.MoldID > 0 {
			moldIDsToQuery = append(moldIDsToQuery, item.MoldID)
			moldCodes = append(moldCodes, "") // 占位
		}
	}
	// 批量查询缺失的模具编号
	if len(moldIDsToQuery) > 0 {
		moldCodeMap := getMoldCodesByMoldIDs(moldIDsToQuery)
		for i, item := range items {
			if item.Mold == nil && item.MoldID > 0 {
				if code, ok := moldCodeMap[item.MoldID]; ok {
					moldCodes[i] = code
				} else {
					moldCodes[i] = fmt.Sprintf("#%d", item.MoldID)
				}
			}
		}
	}
	// 去除空字符串
	filteredCodes := make([]string, 0, len(moldCodes))
	for _, c := range moldCodes {
		if c != "" {
			filteredCodes = append(filteredCodes, c)
		}
	}
	moldCodesStr := strings.Join(filteredCodes, "、")
	NotifyOperatorAndAdmins(userID, "模具订单",
		fmt.Sprintf("订单 %s 已创建，包含 %d 套模具，编号 %s", order.OrderCode, moldCount, moldCodesStr),
		"mold:order_created", "low",
		fmt.Sprintf("/mold/order/detail/%d", order.ID), order.ID)
}
func (c *MoldController) UpdateMoldOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "更新模具订单")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "mold_order")

	var orderDTO dto.MoldOrderDTO
	if err := ctx.ShouldBindJSON(&orderDTO); err != nil {
		logger.Warn("模具订单更新参数绑定失败", zap.Error(err))
		response.Error(ctx, response.BAD_REQUEST, fmt.Sprintf("请求参数错误: %v", err))
		return
	}

	// 将 DTO 转换为 entity
	order := &entity.MoldOrder{
		ID:          orderDTO.ID,
		OrderCode:   orderDTO.OrderCode,
		CustomerID:  orderDTO.CustomerID,
		OrderType:   orderDTO.OrderType,
		MoldID:      orderDTO.MoldID,
		Priority:    orderDTO.Priority,
		TotalAmount: orderDTO.TotalAmount,
		Status:      orderDTO.Status,
		Remark:      orderDTO.Remark,
	}

	// 直接赋值（实体和 DTO 都使用 *dto.Date 类型）
	order.PlanStartDate = orderDTO.PlanStartDate
	order.PlanEndDate = orderDTO.PlanEndDate
	if orderDTO.DeliverTime != nil && !orderDTO.DeliverTime.Time.IsZero() {
		t := orderDTO.DeliverTime.Time
		order.DeliverTime = &t
	}

	userIDPtr := middleware.GetCurrentUserID(ctx)
	if userIDPtr == nil {
		logger.Warn("获取用户ID失败")
		response.Error(ctx, response.UNAUTHORIZED, "用户未登录")
		return
	}
	userID := *userIDPtr

	// 查询更新前的旧订单，用于生成变更内容
	oldOrder, _ := dao.NewMoldOrderDAO().FindByID(orderDTO.ID)

	// ✨ 使用新方法更新订单和工单
	err := c.moldOrderService.UpdateMoldOrderWithItems(order, orderDTO.Items, userID)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	response.Success(ctx, nil)

	// 通知：更新模具订单
	orderCode := orderDTO.OrderCode
	if orderCode == "" {
		orderCode = fmt.Sprintf("#%d", orderDTO.ID)
	}

	var changeParts []string
	if oldOrder != nil {
		if oldOrder.TotalAmount != orderDTO.TotalAmount {
			changeParts = append(changeParts, fmt.Sprintf("金额 %.2f→%.2f", oldOrder.TotalAmount, orderDTO.TotalAmount))
		}
		if oldOrder.Status != orderDTO.Status && orderDTO.Status != "" {
			changeParts = append(changeParts, fmt.Sprintf("状态 %s→%s", oldOrder.Status, orderDTO.Status))
		}
		oldStart := ""
		if oldOrder.PlanStartDate != nil && !oldOrder.PlanStartDate.Time.IsZero() {
			oldStart = oldOrder.PlanStartDate.Time.Format("2006-01-02")
		}
		newStart := ""
		if orderDTO.PlanStartDate != nil && !orderDTO.PlanStartDate.Time.IsZero() {
			newStart = orderDTO.PlanStartDate.Time.Format("2006-01-02")
		}
		if oldStart != newStart && (newStart != "" || oldStart != "") {
			if newStart != "" {
				changeParts = append(changeParts, fmt.Sprintf("计划开始日期→%s", newStart))
			} else {
				changeParts = append(changeParts, "计划开始日期已清除")
			}
		}
		oldEnd := ""
		if oldOrder.PlanEndDate != nil && !oldOrder.PlanEndDate.Time.IsZero() {
			oldEnd = oldOrder.PlanEndDate.Time.Format("2006-01-02")
		}
		newEnd := ""
		if orderDTO.PlanEndDate != nil && !orderDTO.PlanEndDate.Time.IsZero() {
			newEnd = orderDTO.PlanEndDate.Time.Format("2006-01-02")
		}
		if oldEnd != newEnd && (newEnd != "" || oldEnd != "") {
			if newEnd != "" {
				changeParts = append(changeParts, fmt.Sprintf("交期→%s", newEnd))
			} else {
				changeParts = append(changeParts, "交期已清除")
			}
		}
		if orderDTO.Remark != "" {
			changeParts = append(changeParts, fmt.Sprintf("备注：%s", orderDTO.Remark))
		}
	}
	changeContent := "基本信息"
	if len(changeParts) > 0 {
		changeContent = strings.Join(changeParts, "；")
	}

	// 查询订单明细获取所有模具编号（批量查询避免 N+1）
	items2, _ := dao.NewMoldOrderItemDAO().FindByOrderID(orderDTO.ID)
	var itemCodes []string
	moldIDsToQuery2 := make([]int64, 0)
	for _, item := range items2 {
		if item.Mold != nil {
			itemCodes = append(itemCodes, item.Mold.MoldCode)
		} else if item.MoldID > 0 {
			moldIDsToQuery2 = append(moldIDsToQuery2, item.MoldID)
			itemCodes = append(itemCodes, "") // 占位
		}
	}
	if len(moldIDsToQuery2) > 0 {
		moldCodeMap2 := getMoldCodesByMoldIDs(moldIDsToQuery2)
		for i, item := range items2 {
			if item.Mold == nil && item.MoldID > 0 {
				if code, ok := moldCodeMap2[item.MoldID]; ok {
					itemCodes[i] = code
				} else {
					itemCodes[i] = fmt.Sprintf("#%d", item.MoldID)
				}
			}
		}
	}
	// 去除空字符串
	filteredCodes2 := make([]string, 0, len(itemCodes))
	for _, c := range itemCodes {
		if c != "" {
			filteredCodes2 = append(filteredCodes2, c)
		}
	}
	moldCodesStr2 := strings.Join(filteredCodes2, "、")

	NotifyOperatorAndAdmins(userID, "模具订单",
		fmt.Sprintf("订单 %s 已更新，对应模具 %s ，更新内容为 %s", orderCode, moldCodesStr2, changeContent),
		"mold:order_updated", "low",
		"", orderDTO.ID)
}

// DeleteMoldOrder 删除模具订单
func (c *MoldController) DeleteMoldOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "删除模具订单")
	ctx.Set("oper_log_type", 2)
	ctx.Set("oper_log_business_type", "mold_order")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的订单ID")
		return
	}

	if err := c.moldOrderService.DeleteMoldOrder(id, false); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// GetMoldOrderByID 根据 ID 获取模具订单详情
func (c *MoldController) GetMoldOrderByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的订单ID")
		return
	}

	order, err := c.moldOrderService.GetMoldOrderByID(id)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "订单不存在")
		return
	}

	response.Success(ctx, order)
}

// GetMoldWorkOrderByID 根据ID获取模具工单详情
func (c *MoldController) GetMoldWorkOrderByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "参数错误")
		return
	}
	workOrder, err := c.moldWorkOrderService.GetMoldWorkOrderByID(id)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "工单不存在")
		return
	}
	response.Success(ctx, workOrder)
}

// GetMoldWorkOrders 获取模具工单列表
func (c *MoldController) GetMoldWorkOrders(ctx *gin.Context) {
	// ✅ 性能监控：记录接口耗时
	pm := utils.NewPerformanceMonitor("GetMoldWorkOrders", 1*time.Second)
	defer pm.End()

	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	workCode := ctx.Query("workCode")
	moldCode := ctx.Query("moldCode")
	customerIDStr := ctx.Query("customerId")
	statusStr := ctx.Query("status")
	orderTypeStr := ctx.Query("orderType")
	processStage := ctx.Query("processStage")

	var customerID *int64
	if customerIDStr != "" {
		id, err := strconv.ParseInt(customerIDStr, 10, 64)
		if err == nil {
			customerID = &id
		}
	}

	var status *int8
	if statusStr != "" {
		s, err := strconv.ParseInt(statusStr, 10, 8)
		if err == nil {
			statusVal := int8(s)
			status = &statusVal
		}
	}

	orderType := orderTypeStr

	moldOrderIDStr := ctx.Query("mold_order_id")
	if moldOrderIDStr != "" {
		moldOrderID, err := strconv.ParseInt(moldOrderIDStr, 10, 64)
		if err == nil {
			// 直接查询该订单的所有工单（不分页）
			workOrders, err := c.moldWorkOrderService.GetMoldWorkOrdersByMoldOrderID(moldOrderID)
			if err != nil {
				response.Error(ctx, response.ERROR, fmt.Sprintf("获取工单列表失败: %v", err))
				return
			}
			response.Success(ctx, gin.H{
				"list":  workOrders,
				"total": len(workOrders),
			})
			return
		}
	}

	var processStagePtr *string
	if processStage != "" {
		processStagePtr = &processStage
	}

	workOrders, total, err := c.moldWorkOrderService.GetMoldWorkOrdersByPageWithFilter(page, pageSize, workCode, moldCode, customerID, status, orderType, processStagePtr)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取工单列表失败")
		return
	}

	// ✅ 性能优化：异步批量同步工单状态，避免阻塞响应
	// 🔒 修复：深拷贝ID列表，避免数据竞争和访问已释放内存
	workOrderIDs := make([]int64, 0, len(workOrders))
	for _, wo := range workOrders {
		if wo.ID > 0 {
			workOrderIDs = append(workOrderIDs, wo.ID)
		}
	}

	go func(ids []int64) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("异步同步工单状态 panic", zap.Any("panic", r))
			}
		}()
		// 延迟执行，避免影响主流程
		time.Sleep(100 * time.Millisecond)

		// 在 goroutine 中重新查询工单数据，确保数据安全
		if len(ids) == 0 {
			return
		}

		var freshWorkOrders []*entity.MoldWorkOrder
		if err := dao.DB.Where("id IN ?", ids).Find(&freshWorkOrders).Error; err != nil {
			logger.Error("异步同步工单状态失败：查询工单列表错误", zap.Error(err))
			return
		}

		c.batchSyncWorkOrderStatusAsync(freshWorkOrders)
	}(workOrderIDs)

	response.Success(ctx, gin.H{
		"list":      workOrders,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateMoldWorkOrder 创建模具工单
func (c *MoldController) CreateMoldWorkOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建模具工单")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "mold_work_order")

	var workOrderDTO dto.MoldWorkOrderDTO
	if err := ctx.ShouldBindJSON(&workOrderDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, fmt.Sprintf("请求参数错误: %v", err))
		return
	}

	// ✅ 安全加固：使用验证器进行深度验证
	validator := utils.NewWorkOrderValidator()
	if err := validator.ValidateCreateMoldWorkOrder(workOrderDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, err.Error())
		return
	}

	// 安全加固：清理字符串输入，防止 XSS
	if workOrderDTO.Remark != "" {
		workOrderDTO.Remark = utils.SanitizeString(workOrderDTO.Remark)
	}
	if workOrderDTO.NewMoldCode != "" {
		workOrderDTO.NewMoldCode = utils.SanitizeString(workOrderDTO.NewMoldCode)
	}
	if workOrderDTO.NewMoldName != "" {
		workOrderDTO.NewMoldName = utils.SanitizeString(workOrderDTO.NewMoldName)
	}

	createUserID := middleware.GetUserID(ctx)
	workOrder, err := c.moldWorkOrderService.CreateMoldWorkOrder(workOrderDTO, &createUserID)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	response.Success(ctx, workOrder)

	// 通知：创建模具工单
	userID := middleware.GetUserID(ctx)
	workCode := workOrder.WorkCode
	moldCode := ""
	if workOrderDTO.MoldID != nil {
		moldCode = getMoldCodeByMoldID(*workOrderDTO.MoldID)
	}
	NotifyOperatorAndAdmins(userID, "模具工单",
		fmt.Sprintf("工单 %s 已创建，对应模具编号 %s", workCode, moldCode),
		"mold:work_order_created", "low",
		fmt.Sprintf("/mold/work-order/%d", workOrder.ID), workOrder.ID)
}

// UpdateMoldWorkOrder 更新模具工单
func (c *MoldController) UpdateMoldWorkOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "更新模具工单")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "mold_work_order")

	var req struct {
		ID          int64     `json:"id" binding:"required"`
		MoldOrderID *int64    `json:"mold_order_id"`
		MoldID      *int64    `json:"mold_id"`
		PlanStart   *dto.Date `json:"plan_start"`
		PlanEnd     *dto.Date `json:"plan_end"`
		Progress    int       `json:"progress"`
		Status      string    `json:"status"`
		WarehouseID *int64    `json:"warehouse_id"`
		DeliverTime *dto.Date `json:"deliver_time"`
		Remark      string    `json:"remark"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, fmt.Sprintf("请求参数错误: %v", err))
		return
	}

	existing, err := c.moldWorkOrderService.GetMoldWorkOrderByID(req.ID)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "工单不存在")
		return
	}

	// 🔒 校验：如果同时修改了订单和模具（或只修改其中一个），确保订单包含该模具
	moldIDToCheck := existing.MoldID
	if req.MoldID != nil {
		moldIDToCheck = req.MoldID
	}
	moldOrderIDToCheck := existing.MoldOrderID
	if req.MoldOrderID != nil {
		moldOrderIDToCheck = req.MoldOrderID
	}
	if moldOrderIDToCheck != nil && moldIDToCheck != nil {
		var itemCount int64
		dao.DB.Model(&entity.MoldOrderItem{}).
			Where("order_id = ? AND mold_id = ?", *moldOrderIDToCheck, *moldIDToCheck).
			Count(&itemCount)
		if itemCount == 0 {
			var woCount int64
			dao.DB.Model(&entity.MoldWorkOrder{}).
				Where("mold_order_id = ? AND mold_id = ? AND id != ?", *moldOrderIDToCheck, *moldIDToCheck, req.ID).
				Count(&woCount)
			if woCount == 0 {
				response.Error(ctx, response.BAD_REQUEST, "所选订单不包含该模具，请选择正确的订单")
				return
			}
		}
	}

	if req.MoldOrderID != nil {
		existing.MoldOrderID = req.MoldOrderID
	}
	if req.MoldID != nil {
		existing.MoldID = req.MoldID
	}
	if req.PlanStart != nil {
		t := req.PlanStart.Time
		existing.PlanStart = &t
	}
	if req.PlanEnd != nil {
		t := req.PlanEnd.Time
		existing.PlanEnd = &t
	}
	if req.Progress != 0 {
		existing.Progress = req.Progress
	}
	// ✨ 恢复状态更新：允许在入库等操作时更新工单状态
	if req.Status != "" {
		existing.Status = req.Status
	}
	if req.WarehouseID != nil {
		existing.WarehouseID = req.WarehouseID
	}
	if req.DeliverTime != nil {
		t := req.DeliverTime.Time
		existing.DeliverTime = &t
	}
	if req.Remark != "" {
		existing.Remark = req.Remark
	}

	if err := c.moldWorkOrderService.UpdateMoldWorkOrder(existing); err != nil {
		handleServiceError(ctx, err)
		return
	}

	// 通知：工单状态变更（通知操作人 + 管理员）
	userID := middleware.GetUserID(ctx)
	statusName := getWorkOrderStatusName(existing.Status)
	moldCode := ""
	if existing.MoldID != nil {
		moldCode = getMoldCodeByMoldID(*existing.MoldID)
	}
	NotifyOperatorAndAdmins(userID, "模具工单",
		fmt.Sprintf("工单 %s ，模具编号 %s 状态已变更为 %s", existing.WorkCode, moldCode, statusName),
		"mold:work_order_status", getStatusQuality(existing.Status),
		fmt.Sprintf("/mold/work-order/%d", existing.ID), existing.ID)

	response.Success(ctx, existing)
}

// DeleteMoldWorkOrder 删除模具工单
func (c *MoldController) DeleteMoldWorkOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "删除模具工单")
	ctx.Set("oper_log_type", 2)
	ctx.Set("oper_log_business_type", "mold_work_order")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的工单ID")
		return
	}

	if err := c.moldWorkOrderService.DeleteMoldWorkOrder(id); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// CancelMoldWorkOrder 取消模具工单（软删除，标记取消状态并锁定）
func (c *MoldController) CancelMoldWorkOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "取消模具工单")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "mold_work_order")

	var req struct {
		ID     int64  `json:"id" binding:"required"`
		Reason string `json:"reason" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "参数错误")
		return
	}

	userID := middleware.GetCurrentUserID(ctx)

	if err := c.moldWorkOrderService.CancelMoldWorkOrder(req.ID, req.Reason, userID); err != nil {
		handleServiceError(ctx, err)
		return
	}

	// 通知：工单已取消（通知操作人 + 管理员）
	uid := int64(0)
	if userID != nil {
		uid = *userID
	}
	cancelWorkOrder, _ := dao.NewMoldWorkOrderDAO().FindByID(req.ID)
	cancelWorkCode := getWorkCodeByWorkOrderID(req.ID)
	cancelMoldCode := ""
	if cancelWorkOrder != nil {
		cancelWorkCode = cancelWorkOrder.WorkCode
		if cancelWorkOrder.MoldID != nil {
			cancelMoldCode = getMoldCodeByMoldID(*cancelWorkOrder.MoldID)
		}
	}
	NotifyOperatorAndAdmins(uid, "模具工单",
		fmt.Sprintf("工单 %s ，模具编号 %s 已取消，原因：%s", cancelWorkCode, cancelMoldCode, req.Reason),
		"mold:work_order_status", "high",
		"", req.ID)

	response.Success(ctx, nil)
}

// GetMoldTryTests 获取模具试模记录列表
func (c *MoldController) GetMoldTryTests(ctx *gin.Context) {
	// 支持两种方式：通过模具ID查询或通过工单ID查询
	moldIDStr := ctx.Query("mold_id")
	workOrderIDStr := ctx.Query("work_order_id")

	var tests []*entity.MoldTryTest

	if moldIDStr != "" {
		// 通过模具ID查询：先获取模具，再查询其关联工单，再查询试模记录
		moldID, err := strconv.ParseInt(moldIDStr, 10, 64)
		if err != nil {
			response.Error(ctx, response.BAD_REQUEST, "模具 ID 参数错误")
			return
		}
		tests, err = c.moldTryTestService.GetMoldTryTestsByMoldID(moldID)
		if err != nil {
			response.Error(ctx, response.ERROR, "获取试模记录失败")
			return
		}
	} else if workOrderIDStr != "" {
		// 通过工单ID查询（原有逻辑）
		workOrderID, err := strconv.ParseInt(workOrderIDStr, 10, 64)
		if err != nil {
			response.Error(ctx, response.BAD_REQUEST, "工单 ID 参数错误")
			return
		}
		tests, err = c.moldTryTestService.GetMoldTryTestsByWorkOrderID(workOrderID)
		if err != nil {
			response.Error(ctx, response.ERROR, "获取试模记录失败")
			return
		}
	} else {
		response.Error(ctx, response.ERROR, "请提供模具ID或工单ID")
		return
	}

	response.Success(ctx, gin.H{"list": tests, "total": len(tests)})
}

// CreateMoldTryTest 创建模具试模记录
func (c *MoldController) CreateMoldTryTest(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建试模记录")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "mold_try_test")

	var testDTO dto.MoldTryTestDTO
	if err := ctx.ShouldBindJSON(&testDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, fmt.Sprintf("请求参数错误: %v", err))
		return
	}
	test := &entity.MoldTryTest{
		MoldWorkID: testDTO.MoldWorkID,
		TestTime:   testDTO.TestTime.Time, // CustomTime 转换为 *time.Time
		OkQty:      testDTO.OkQty,
		BadQty:     testDTO.BadQty,
		Result:     testDTO.Result,
		TestMode:   testDTO.TestMode,
		UserID:     testDTO.UserID,
		Remark:     testDTO.Remark,
	}
	created, err := c.moldTryTestService.CreateMoldTryTest(test)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	response.Success(ctx, created)

	// 通知：试模完成（含工单状态变更为"已试模"）
	userID := middleware.GetUserID(ctx)
	workCode := getWorkCodeByWorkOrderID(testDTO.MoldWorkID)
	resultText := testDTO.Result
	if resultText == "" {
		resultText = "未填写"
	}

	tryTestWorkOrder, _ := dao.NewMoldWorkOrderDAO().FindByID(testDTO.MoldWorkID)
	tryTestMoldCode := ""
	if tryTestWorkOrder != nil && tryTestWorkOrder.MoldID != nil {
		tryTestMoldCode = getMoldCodeByMoldID(*tryTestWorkOrder.MoldID)
	}

	var title string
	notifQuality := "middle"
	if resultText == "合格" {
		title = fmt.Sprintf("模具编号 %s 试模已合格", tryTestMoldCode)
		notifQuality = "low"
	} else {
		title = fmt.Sprintf("工单 %s ，模具编号 %s 试模结果：%s", workCode, tryTestMoldCode, resultText)
	}
	NotifyOperatorAndAdmins(userID, "模具工单", title,
		"mold:try_test", notifQuality,
		fmt.Sprintf("/mold/work-order/%d", testDTO.MoldWorkID), testDTO.MoldWorkID)
}
func (c *MoldController) UpdateMoldTryTest(ctx *gin.Context) {
	ctx.Set("oper_log_title", "更新试模记录")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "mold_try_test")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "参数错误")
		return
	}

	var testDTO dto.MoldTryTestDTO
	if err := ctx.ShouldBindJSON(&testDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, fmt.Sprintf("请求参数错误: %v", err))
		return
	}

	test := &entity.MoldTryTest{
		ID:         id,
		MoldWorkID: testDTO.MoldWorkID,
		TestTime:   testDTO.TestTime.Time, // CustomTime 转换为 *time.Time
		OkQty:      testDTO.OkQty,
		BadQty:     testDTO.BadQty,
		Result:     testDTO.Result,
		TestMode:   testDTO.TestMode,
		UserID:     testDTO.UserID,
		Remark:     testDTO.Remark,
	}

	err = c.moldTryTestService.UpdateMoldTryTest(test)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	response.Success(ctx, nil)

	// 通知：更新试模记录
	userID := middleware.GetUserID(ctx)
	resultText := testDTO.Result
	if resultText == "" {
		resultText = "未填写"
	}
	updatedWorkOrder, _ := dao.NewMoldWorkOrderDAO().FindByID(testDTO.MoldWorkID)
	updatedMoldCode := ""
	if updatedWorkOrder != nil && updatedWorkOrder.MoldID != nil {
		updatedMoldCode = getMoldCodeByMoldID(*updatedWorkOrder.MoldID)
	}
	title := fmt.Sprintf("模具编号 %s 试模记录已更新为 %s", updatedMoldCode, resultText)
	if resultText == "合格" {
		title = fmt.Sprintf("模具编号 %s 试模已合格", updatedMoldCode)
	}
	NotifyOperatorAndAdmins(userID, "模具工单", title,
		"mold:try_test_updated", "low",
		fmt.Sprintf("/mold/work-order/%d", testDTO.MoldWorkID), testDTO.MoldWorkID)
}

// DeleteMoldTryTest 删除模具试模记录
func (c *MoldController) DeleteMoldTryTest(ctx *gin.Context) {
	ctx.Set("oper_log_title", "删除试模记录")
	ctx.Set("oper_log_type", 2)
	ctx.Set("oper_log_business_type", "mold_try_test")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "参数错误")
		return
	}

	err = c.moldTryTestService.DeleteMoldTryTest(id)
	if err != nil {
		response.Error(ctx, response.ERROR, "删除失败")
		return
	}
	response.Success(ctx, nil)
}

// GetMoldDeliveries 获取模具交付记录列表
func (c *MoldController) GetMoldDeliveries(ctx *gin.Context) {
	workOrderIDStr := ctx.Query("work_order_id")
	workOrderID, err := strconv.ParseInt(workOrderIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "工单 ID 参数错误")
		return
	}
	deliveries, err := c.moldDeliveryService.GetMoldDeliveriesByWorkOrderID(workOrderID)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取交付记录失败")
		return
	}
	response.Success(ctx, gin.H{"list": deliveries, "total": len(deliveries)})
}

// CreateMoldDelivery 创建模具交付记录
func (c *MoldController) CreateMoldDelivery(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建交付记录")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "mold_delivery")

	var deliveryDTO dto.MoldDeliveryDTO
	if err := ctx.ShouldBindJSON(&deliveryDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}
	delivery := &entity.MoldDelivery{
		MoldWorkID:        deliveryDTO.MoldWorkID,
		CustomerID:        deliveryDTO.CustomerID,
		DeliveryTime:      deliveryTimeToPtr(deliveryDTO.DeliveryTime),
		DelivererID:       deliveryDTO.DelivererID,
		ReceiverName:      deliveryDTO.ReceiverName,
		ReceiverPhone:     deliveryDTO.ReceiverPhone,
		DeliveryAddress:   deliveryDTO.DeliveryAddress,
		LogisticsCompany:  deliveryDTO.LogisticsCompany,
		TrackingNo:        deliveryDTO.TrackingNo,
		DeliveryMethod:    deliveryDTO.DeliveryMethod,
		DocumentsAttached: deliveryDTO.DocumentsAttached,
		Remark:            deliveryDTO.Remark,
	}
	created, err := c.moldDeliveryService.CreateMoldDelivery(delivery)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	response.Success(ctx, created)

	// 通知：交付完成
	userID := middleware.GetUserID(ctx)
	methodText := deliveryDTO.DeliveryMethod
	if methodText == "" {
		methodText = "未知方式"
	}
	deliveryWorkOrder, _ := dao.NewMoldWorkOrderDAO().FindByID(deliveryDTO.MoldWorkID)
	deliveryMoldCode := ""
	if deliveryWorkOrder != nil && deliveryWorkOrder.MoldID != nil {
		deliveryMoldCode = getMoldCodeByMoldID(*deliveryWorkOrder.MoldID)
	}
	NotifyOperatorAndAdmins(userID, "模具工单",
		fmt.Sprintf("模具编号 %s 已交付，交付方式：%s", deliveryMoldCode, methodText),
		"mold:delivery", "low",
		"", deliveryDTO.MoldWorkID)
}

// GetMoldStocks 获取模具库存记录列表
func (c *MoldController) GetMoldStocks(ctx *gin.Context) {
	moldIDStr := ctx.Query("mold_id")
	moldID, err := strconv.ParseInt(moldIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "模具 ID 参数错误")
		return
	}
	stocks, err := c.moldStockService.GetMoldStocksByMoldID(moldID)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取库存记录失败")
		return
	}
	response.Success(ctx, gin.H{"list": stocks, "total": len(stocks)})
}

// CreateMoldStock 创建模具库存记录
func (c *MoldController) CreateMoldStock(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建模具库存记录")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "mold_stock")

	var stockDTO dto.MoldStockDTO
	if err := ctx.ShouldBindJSON(&stockDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	var operateTime *time.Time
	if stockDTO.OperateTime != "" {
		parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", stockDTO.OperateTime, time.Local)
		if err != nil {
			response.Error(ctx, response.BAD_REQUEST, "操作时间格式错误，请使用 YYYY-MM-DD HH:mm:ss 格式")
			return
		}
		operateTime = &parsedTime
	}

	stock := &entity.MoldStock{
		MoldID:      stockDTO.MoldID,
		WarehouseID: stockDTO.WarehouseID,
		StockType:   stockDTO.StockType,
		OperateTime: operateTime,
		OperatorID:  stockDTO.OperatorID,
		CustomerID:  stockDTO.CustomerID,
		Quantity:    stockDTO.Quantity,
		Remark:      stockDTO.Remark,
	}
	created, err := c.moldStockService.CreateMoldStock(stock)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}
	response.Success(ctx, created)

	// 通知：模具库存变动
	userID := middleware.GetUserID(ctx)
	moldCode := getMoldCodeByMoldID(stockDTO.MoldID)
	quantity := stockDTO.Quantity
	stockTypeText := "入库"
	extraInfo := ""
	if stockDTO.StockType == 2 {
		stockTypeText = "出库"
	} else if stockDTO.StockType == 3 {
		stockTypeText = "交付客户"
	}
	// 入库时查询仓库位置
	if stockDTO.StockType == 1 && stockDTO.WarehouseID != nil {
		warehouse, _ := dao.NewWarehouseDAO().FindByID(*stockDTO.WarehouseID)
		if warehouse != nil {
			extraInfo = fmt.Sprintf("，当前存放在 %s", warehouse.WhName)
		}
	}
	notifyContent := fmt.Sprintf("模具 %s 已%s %d 套%s", moldCode, stockTypeText, quantity, extraInfo)
	NotifyOperatorAndAdmins(userID, "模具管理", notifyContent,
		"mold:stock_changed", "low",
		"", stockDTO.MoldID)
}

// GetMoldStatistics 获取模具统计信息
func (c *MoldController) GetMoldStatistics(ctx *gin.Context) {
	statistics, err := c.moldService.GetMoldStatistics()
	if err != nil {
		response.Error(ctx, response.ERROR, "获取统计信息失败")
		return
	}
	response.Success(ctx, statistics)
}

// ========== 通知辅助函数 ==========

// getMoldCodeByMoldID 根据模具ID查询模具编号
func getMoldCodeByMoldID(moldID int64) string {
	mold, err := dao.NewMoldDAO().FindByID(moldID)
	if err != nil {
		return fmt.Sprintf("#%d", moldID)
	}
	return mold.MoldCode
}

// getMoldCodesByMoldIDs 批量查询模具编号（避免 N+1）
func getMoldCodesByMoldIDs(moldIDs []int64) map[int64]string {
	result := make(map[int64]string, len(moldIDs))
	if len(moldIDs) == 0 {
		return result
	}
	var molds []entity.Mold
	dao.DB.Where("id IN ?", moldIDs).Select("id, mold_code").Find(&molds)
	for _, m := range molds {
		result[m.ID] = m.MoldCode
	}
	return result
}

// getWorkCodeByWorkOrderID 根据工单ID查询工单号
func getWorkCodeByWorkOrderID(workOrderID int64) string {
	wo, err := dao.NewMoldWorkOrderDAO().FindByID(workOrderID)
	if err != nil {
		return fmt.Sprintf("#%d", workOrderID)
	}
	return wo.WorkCode
}
func getWorkOrderStatusName(status string) string {
	switch status {
	case constant.WorkOrderStatusNotStarted:
		return "未开始"
	case constant.WorkOrderStatusProcessing:
		return "加工中"
	case constant.WorkOrderStatusCompleted:
		return "已完成"
	case constant.WorkOrderStatusTested:
		return "已试模"
	case constant.WorkOrderStatusStockedIn:
		return "已入库"
	case constant.WorkOrderStatusDelivered:
		return "已交付"
	case constant.WorkOrderStatusCancelled:
		return "已取消"
	default:
		return status
	}
}

func getStatusQuality(status string) string {
	switch status {
	case constant.WorkOrderStatusDelivered:
		return "low"
	case constant.WorkOrderStatusProcessing:
		return "middle"
	case constant.WorkOrderStatusCancelled:
		return "high"
	default:
		return "low"
	}
}

// deliveryTimeToPtr 将 *dto.FlexibleTime 转换为 *time.Time（用于赋值给 entity）
func deliveryTimeToPtr(ft *dto.FlexibleTime) *time.Time {
	if ft == nil {
		return nil
	}
	return &ft.Time
}

package controller

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type ProductionController struct {
	orderService       *service.ProductionOrderService
	reportService      *service.ProductionReportService
	outsourcingService *service.OutsourcingService
	stockInDAO         *dao.ProductionStockInDAO
	stockOutDAO        *dao.ProductionStockOutDAO
}

func NewProductionController() *ProductionController {
	return &ProductionController{
		orderService:       service.NewProductionOrderService(),
		reportService:      service.NewProductionReportService(),
		outsourcingService: service.NewOutsourcingService(),
		stockInDAO:         dao.NewProductionStockInDAO(),
		stockOutDAO:        dao.NewProductionStockOutDAO(),
	}
}

// GetOrders 获取生产订单列表
func (c *ProductionController) GetOrders(ctx *gin.Context) {
	var query dto.ProductionOrderQueryDTO
	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}
	if query.DueSoon != nil && *query.DueSoon != "" {
		if days, err := strconv.Atoi(*query.DueSoon); err == nil && days > 0 {
			today := time.Now().Format("2006-01-02")
			end := time.Now().AddDate(0, 0, days).Format("2006-01-02")
			query.PlanEndDateStart = &today
			query.PlanEndDateEnd = &end
		}
	}

	orders, total, err := c.orderService.GetOrdersByPage(query)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	// 批量查询已出货数量（优化：由 N 次查询变为 1 次）
	if len(orders) > 0 {
		orderIDs := make([]int64, len(orders))
		for i, order := range orders {
			orderIDs[i] = order.ID
		}
		if shippedQtyMap, err := c.orderService.BatchGetShippedQty(orderIDs); err == nil {
			for i := range orders {
				if qty, ok := shippedQtyMap[orders[i].ID]; ok {
					orders[i].ShippedQty = qty
				}
			}
		}
	}

	// 批量获取可用操作编码（优化：由 N 次 API 调用变为 1 次数据库查询）
	if len(orders) > 0 {
		c.orderService.BatchGetActionCodes(orders)
	}

	response.Success(ctx, gin.H{
		"list":      orders,
		"total":     total,
		"page":      query.Page,
		"page_size": query.PageSize,
	})
}

// GetOrderByID 获取订单详情
func (c *ProductionController) GetOrderByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的订单ID")
		return
	}

	order, err := c.orderService.GetOrderByID(id)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "订单不存在")
		return
	}

	response.Success(ctx, order)
}

// GetProductionItemsByMoldID 获取模具关联的生产订单明细
func (c *ProductionController) GetProductionItemsByMoldID(ctx *gin.Context) {
	moldIDStr := ctx.Param("mold_id")
	moldID, err := strconv.ParseInt(moldIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的模具ID")
		return
	}

	items, err := c.orderService.GetProductionItemsByMoldID(moldID)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, items)
}

// CreateOrder 创建生产订单
func (c *ProductionController) CreateOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建生产订单")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "production_order")

	var orderDTO dto.ProductionOrderDTO
	if err := ctx.ShouldBindJSON(&orderDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	// 从 JWT token 中获取用户ID
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "未登录或会话已过期")
		return
	}

	if err := validatePlanDateStrings(orderDTO.PlanStartDate, orderDTO.PlanEndDate); err != nil {
		response.Error(ctx, response.BAD_REQUEST, err.Error())
		return
	}

	order, err := c.orderService.CreateProductionOrder(orderDTO, userID)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, order)

	// 通知：创建生产订单（批量查询模具编号避免 N+1）
	moldCodes := make([]string, 0)
	productCount := 0
	if orderDTO.Items != nil {
		productCount = len(orderDTO.Items)
		moldIDsToQuery := make([]int64, 0)
		for _, item := range orderDTO.Items {
			if len(item.MoldIDs) > 0 {
				moldIDsToQuery = append(moldIDsToQuery, item.MoldIDs...)
			} else if item.MoldID != nil && *item.MoldID > 0 {
				moldIDsToQuery = append(moldIDsToQuery, *item.MoldID)
			}
		}
		// 批量查询模具编号
		moldCodeMap := getMoldCodesByMoldIDs(moldIDsToQuery)
		for _, item := range orderDTO.Items {
			if len(item.MoldIDs) > 0 {
				for _, mid := range item.MoldIDs {
					if code, ok := moldCodeMap[mid]; ok {
						moldCodes = append(moldCodes, code)
					} else {
						moldCodes = append(moldCodes, fmt.Sprintf("#%d", mid))
					}
				}
			} else if item.MoldID != nil {
				if code, ok := moldCodeMap[*item.MoldID]; ok {
					moldCodes = append(moldCodes, code)
				} else {
					moldCodes = append(moldCodes, fmt.Sprintf("#%d", *item.MoldID))
				}
			} else if item.ExternalMoldCode != nil && *item.ExternalMoldCode != "" {
				moldCodes = append(moldCodes, *item.ExternalMoldCode)
			}
		}
	}
	// 去重
	moldCodeMap := make(map[string]bool)
	uniqueMoldCodes := make([]string, 0)
	for _, c := range moldCodes {
		if !moldCodeMap[c] {
			moldCodeMap[c] = true
			uniqueMoldCodes = append(uniqueMoldCodes, c)
		}
	}
	moldCodesStr := strings.Join(uniqueMoldCodes, "、")
	NotifyOperatorAndAdmins(userID, "生产订单",
		fmt.Sprintf("订单 %s 已创建，模具 %s，包含%d款产品", order.OrderCode, moldCodesStr, productCount),
		"production:order_created", "low",
		"", order.ID)
}

// UpdateOrder 更新生产订单
func (c *ProductionController) UpdateOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "更新生产订单")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "production_order")

	var orderDTO dto.ProductionOrderDTO
	if err := ctx.ShouldBindJSON(&orderDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(ctx)
	if err := c.orderService.UpdateProductionOrder(orderDTO, userID); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// CancelOrder 取消订单
func (c *ProductionController) CancelOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "取消生产订单")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "production_order")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的订单ID")
		return
	}

	var body struct {
		Reason string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		body.Reason = ""
	}

	if err := c.orderService.CancelOrder(id, body.Reason); err != nil {
		handleServiceError(ctx, err)
		return
	}

	// 通知：订单已取消
	userID := middleware.GetUserID(ctx)
	orderCode := getOrderCodeOrID(id, c)
	productName := getOrderProductName(id, c)
	reason := body.Reason
	if reason == "" {
		reason = "无原因"
	}
	NotifyOperatorAndAdmins(userID, "生产订单",
		fmt.Sprintf("订单 %s ，产品名称 %s 已取消，原因：%s", orderCode, productName, reason),
		"production:order_status", "high",
		"", id)

	response.Success(ctx, nil)
}

// StartProduction 开始生产
func (c *ProductionController) StartProduction(ctx *gin.Context) {
	ctx.Set("oper_log_title", "开始生产")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "production_order")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的订单ID")
		return
	}

	if err := c.orderService.StartProduction(id); err != nil {
		handleServiceError(ctx, err)
		return
	}

	// 通知：订单开始生产（通知操作人 + 管理员）
	userID := middleware.GetUserID(ctx)
	productName := getOrderProductName(id, c)
	NotifyOperatorAndAdmins(userID, "生产订单",
		fmt.Sprintf("产品名称 %s 开始生产", productName),
		"production:order_status", "middle",
		"", id)

	response.Success(ctx, nil)
}

// CompleteProduction 完成生产（直接标记为已入库，报工时已自动入库）
func (c *ProductionController) CompleteProduction(ctx *gin.Context) {
	ctx.Set("oper_log_title", "完成生产")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "production_order")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的订单ID")
		return
	}

	if err := c.orderService.CompleteProduction(id); err != nil {
		handleServiceError(ctx, err)
		return
	}

	// 通知：生产已完成
	userID := middleware.GetUserID(ctx)
	productName := getOrderProductName(id, c)
	NotifyOperatorAndAdmins(userID, "生产订单",
		fmt.Sprintf("产品名称 %s 已完成生产", productName),
		"production:order_status", "middle",
		"", id)

	response.Success(ctx, nil)
}

// StockInOrder 生产入库
func (c *ProductionController) StockInOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "生产入库")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "production_stock_in")

	var stockInDTO dto.ProductionStockInDTO
	if err := ctx.ShouldBindJSON(&stockInDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	// 从 JWT token 中获取用户ID
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "未登录或会话已过期")
		return
	}

	if err := c.orderService.StockInOrder(stockInDTO, userID); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)

	// 通知：生产入库
	orderCode := getOrderCodeOrID(stockInDTO.OrderID, c)
	productName := getOrderProductName(stockInDTO.OrderID, c)
	NotifyOperatorAndAdmins(userID, "生产订单",
		fmt.Sprintf("订单 %s ，产品 %s 已入库", orderCode, productName),
		"production:stock_in", "middle",
		"", stockInDTO.OrderID)
}

// ShipmentOrder 出货
func (c *ProductionController) ShipmentOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "生产出货")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "production_order")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的订单ID")
		return
	}

	var body struct {
		LogisticsCompany string `json:"logistics_company"`
		TrackingNo       string `json:"tracking_no"`
		Remark           string `json:"remark"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		body.Remark = ""
	}

	if err := c.orderService.ShipmentOrder(id, body.Remark); err != nil {
		handleServiceError(ctx, err)
		return
	}

	// 通知：已出货（通知操作人 + 管理员）
	userID := middleware.GetUserID(ctx)
	orderCode := getOrderCodeOrID(id, c)
	productName := getOrderProductName(id, c)
	NotifyOperatorAndAdmins(userID, "生产订单",
		fmt.Sprintf("订单 %s ,产品名称 %s 已出货", orderCode, productName),
		"production:order_status", "low",
		"", id)

	response.Success(ctx, nil)
}

// FinishOrder 完成订单
func (c *ProductionController) FinishOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "完成生产订单")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "production_order")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的订单ID")
		return
	}

	if err := c.orderService.CompleteOrder(id); err != nil {
		handleServiceError(ctx, err)
		return
	}

	// 通知：订单已完结（通知操作人 + 管理员）
	userID := middleware.GetUserID(ctx)
	orderCode := getOrderCodeOrID(id, c)
	productName := getOrderProductName(id, c)
	NotifyOperatorAndAdmins(userID, "生产订单",
		fmt.Sprintf("订单 %s ,产品名称 %s 已完结", orderCode, productName),
		"production:order_status", "low",
		"", id)

	response.Success(ctx, nil)
}

// ReplenishShipment 补发（已完成订单的再次出货）
func (c *ProductionController) ReplenishShipment(ctx *gin.Context) {
	ctx.Set("oper_log_title", "补发出货")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "production_order")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的订单ID")
		return
	}

	var req struct {
		Remark string `json:"remark"`
	}
	ctx.ShouldBindJSON(&req)

	if err := c.orderService.ReplenishShipment(id, req.Remark); err != nil {
		handleServiceError(ctx, err)
		return
	}
	response.Success(ctx, nil)
}

// FinishOrderItem 完成单个生产明细项
func (c *ProductionController) FinishOrderItem(ctx *gin.Context) {
	itemIdStr := ctx.Param("item_id")
	itemId, err := strconv.ParseInt(itemIdStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的明细项ID")
		return
	}

	if err := c.orderService.CompleteProductionItem(itemId); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// GetAvailableActions 获取订单可用的操作列表
func (c *ProductionController) GetAvailableActions(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的订单ID")
		return
	}

	orderDAO := dao.NewProductionOrderDAO()
	order, err := orderDAO.FindByID(id)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "订单不存在")
		return
	}

	flowService := service.NewProductionStatusFlowService()

	//  检查依赖状态：如果是 WAITING_MOLD，不返回任何操作
	if order.DependencyStatus == "WAITING_MOLD" {
		response.Success(ctx, []*entity.ProductionStatusFlow{})
		return
	}

	// ✨ 优先使用明细项状态，如果没有明细项则使用订单主表状态
	var currentStatusValue string
	if len(order.Items) > 0 {
		firstItemStatus := order.Items[0].Status
		allSameStatus := true
		for _, item := range order.Items {
			if item.Status != firstItemStatus {
				allSameStatus = false
				break
			}
		}

		if !allSameStatus {
			// 明细项状态不一致，返回空操作列表
			response.Success(ctx, []*entity.ProductionStatusFlow{})
			return
		}

		// 使用明细项的状态
		currentStatusValue, err = flowService.GetDictValue("production-order-status", firstItemStatus)
		if err != nil {
			response.Error(ctx, response.ERROR, "订单明细状态配置异常")
			return
		}
	} else {
		// 没有明细项，使用订单主表状态（兼容旧数据）
		currentStatusValue, err = flowService.GetDictValue("production-order-status", order.Status)
		if err != nil {
			response.Error(ctx, response.ERROR, "订单状态配置异常")
			return
		}
	}

	actions, err := flowService.GetAvailableActions(currentStatusValue)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取可用操作失败")
		return
	}

	response.Success(ctx, actions)
}

// SubmitReport 提交报工
func (c *ProductionController) SubmitReport(ctx *gin.Context) {
	ctx.Set("oper_log_title", "提交报工")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "production_report")

	var reportDTO dto.ProductionReportDTO
	if err := ctx.ShouldBindJSON(&reportDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	// 从 JWT token 中获取用户ID
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "未登录或会话已过期")
		return
	}

	if err := c.reportService.SubmitReport(reportDTO, userID); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)

	// 通知：提交报工
	productName := getOrderProductName(reportDTO.OrderID, c)
	NotifyOperatorAndAdmins(userID, "生产订单",
		fmt.Sprintf("产品 %s 报工：合格 %d，不良 %d，报废 %d", productName, reportDTO.OkQty, reportDTO.BadQty, reportDTO.ScrapQty),
		"production:report_submitted", "middle",
		"", reportDTO.OrderID)
}

// GetReports 获取报工记录
func (c *ProductionController) GetReports(ctx *gin.Context) {
	orderIDStr := ctx.Query("order_id")
	if orderIDStr == "" {
		response.Error(ctx, response.BAD_REQUEST, "请提供订单ID")
		return
	}

	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的订单ID")
		return
	}

	reports, err := c.reportService.GetReportsByOrder(orderID)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, reports)
}

// CreateOutsourcingRecord 创建外发记录
func (c *ProductionController) CreateOutsourcingRecord(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建外发记录")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "outsourcing")

	var recordDTO dto.OutsourcingRecordDTO
	if err := ctx.ShouldBindJSON(&recordDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	// 从 JWT token 中获取用户ID
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "未登录或会话已过期")
		return
	}

	if err := c.outsourcingService.CreateOutsourcingRecord(recordDTO, userID); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// SendOutsourcing 确认外发
func (c *ProductionController) SendOutsourcing(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的外发记录ID")
		return
	}

	if err := c.outsourcingService.SendOutsourcing(id); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)

	// 通知：确认外发
	userID := middleware.GetUserID(ctx)
	outRecord, _ := dao.NewOutsourcingRecordDAO().FindByID(id)
	outMoldCode := ""
	outProdName := ""
	outTypeName := "外发"
	if outRecord != nil {
		if outRecord.WorkOrderID != nil {
			// 模具外发
			outTypeName = "模具外发"
			if outRecord.MoldID != nil {
				outMoldCode = getMoldCodeByMoldID(*outRecord.MoldID)
			} else {
				// 通过工单查模具
				if wo, _ := dao.NewMoldWorkOrderDAO().FindByID(*outRecord.WorkOrderID); wo != nil && wo.MoldID != nil {
					outMoldCode = getMoldCodeByMoldID(*wo.MoldID)
				}
			}
			NotifyOperatorAndAdmins(userID, "外发管理",
				fmt.Sprintf("%s ，模具编号 %s 已发出，请注意提醒加工商跟进", outTypeName, outMoldCode),
				"outsourcing:sent", "middle",
				"", id)
		} else if outRecord.OrderID != nil {
			// 生产外发
			outTypeName = "生产外发"
			if outRecord.MoldID != nil {
				outMoldCode = getMoldCodeByMoldID(*outRecord.MoldID)
			}
			if po, _ := dao.NewProductionOrderDAO().FindByID(*outRecord.OrderID); po != nil {
				outProdName = po.ProductName
			}
			NotifyOperatorAndAdmins(userID, "外发管理",
				fmt.Sprintf("%s ，模具编号 %s ，产品 %s 已发出，请注意提醒加工商跟进", outTypeName, outMoldCode, outProdName),
				"outsourcing:sent", "middle",
				"", id)
		} else {
			NotifyOperatorAndAdmins(userID, "外发管理",
				fmt.Sprintf("外发任务 #%d 已发出，请注意提醒加工商跟进", id),
				"outsourcing:sent", "middle",
				"", id)
		}
	} else {
		NotifyOperatorAndAdmins(userID, "外发管理",
			fmt.Sprintf("外发任务 #%d 已发出，请注意提醒加工商跟进", id),
			"outsourcing:sent", "middle",
			"", id)
	}
}

// ReceiveOutsourcing 回收确认
func (c *ProductionController) ReceiveOutsourcing(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的外发记录ID")
		return
	}

	if err := c.outsourcingService.ReceiveOutsourcing(id); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)

	// 通知：外发回收
	userID := middleware.GetUserID(ctx)
	retRecord, _ := dao.NewOutsourcingRecordDAO().FindByID(id)
	retMoldCode := ""
	retProdName := ""
	if retRecord != nil {
		if retRecord.WorkOrderID != nil {
			// 模具外发回收
			if retRecord.MoldID != nil {
				retMoldCode = getMoldCodeByMoldID(*retRecord.MoldID)
			} else {
				if wo, _ := dao.NewMoldWorkOrderDAO().FindByID(*retRecord.WorkOrderID); wo != nil && wo.MoldID != nil {
					retMoldCode = getMoldCodeByMoldID(*wo.MoldID)
				}
			}
			NotifyOperatorAndAdmins(userID, "外发管理",
				fmt.Sprintf("模具外发 ，模具编号 %s 已回收，请注意检查模具情况", retMoldCode),
				"outsourcing:returned", "middle",
				"", id)
		} else if retRecord.OrderID != nil {
			// 生产外发回收
			if retRecord.MoldID != nil {
				retMoldCode = getMoldCodeByMoldID(*retRecord.MoldID)
			}
			if po, _ := dao.NewProductionOrderDAO().FindByID(*retRecord.OrderID); po != nil {
				retProdName = po.ProductName
			}
			NotifyOperatorAndAdmins(userID, "外发管理",
				fmt.Sprintf("生产外发 ，模具编号 %s ，产品 %s 已回收，请注意检查模具和产品情况", retMoldCode, retProdName),
				"outsourcing:returned", "middle",
				"", id)
		} else {
			NotifyOperatorAndAdmins(userID, "外发管理",
				fmt.Sprintf("外发任务 #%d 已回收，请注意检查模具情况", id),
				"outsourcing:returned", "middle",
				"", id)
		}
	} else {
		NotifyOperatorAndAdmins(userID, "外发管理",
			fmt.Sprintf("外发任务 #%d 已回收，请注意检查模具情况", id),
			"outsourcing:returned", "middle",
			"", id)
	}
}

// GenerateShareToken 生成分享链接token
func (c *ProductionController) GenerateShareToken(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的外发记录ID")
		return
	}

	token, err := c.outsourcingService.GenerateShareToken(id)
	if err != nil {
		response.Error(ctx, response.ERROR, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"share_token": token,
	})
}

// RegenerateShareToken 重新生成分享链接token（旧链接失效）
func (c *ProductionController) RegenerateShareToken(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的外发记录ID")
		return
	}

	token, err := c.outsourcingService.RegenerateShareToken(id)
	if err != nil {
		response.Error(ctx, response.ERROR, err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"share_token": token,
	})
}

// GetOutsourcingReports 获取外发记录的所有报工记录
func (c *ProductionController) GetOutsourcingReports(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的外发记录ID")
		return
	}

	reportDAO := dao.NewOutsourcingReportDAO()
	reports, err := reportDAO.FindByOutsourcingID(id)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询报工记录失败")
		return
	}

	if reports == nil {
		reports = []entity.OutsourcingReport{}
	}
	response.Success(ctx, reports)
}

// AddOutsourcingReport 添加外发报工
func (c *ProductionController) AddOutsourcingReport(ctx *gin.Context) {
	var reportDTO dto.OutsourcingReportDTO
	if err := ctx.ShouldBindJSON(&reportDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	if err := c.outsourcingService.AddOutsourcingReport(reportDTO); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// GetOutsourcingRecords 获取外发记录列表
func (c *ProductionController) GetOutsourcingRecords(ctx *gin.Context) {
	statusStr := ctx.Query("status")
	var status *int8
	if statusStr != "" {
		s, err := strconv.ParseInt(statusStr, 10, 8)
		if err == nil {
			s8 := int8(s)
			status = &s8
		}
	}

	records, err := c.outsourcingService.GetOutsourcingRecords(status)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, records)
}

// GetStatistics 获取订单统计
func (c *ProductionController) GetStatistics(ctx *gin.Context) {
	stats, err := c.orderService.GetStatistics()
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, stats)
}

// CreateMoldUsageRecord 创建模具使用记录
func (c *ProductionController) CreateMoldUsageRecord(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建模具使用记录")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "mold_usage")

	var body struct {
		OrderID int64 `json:"order_id" binding:"required"`
		Shots   int   `json:"shots" binding:"required"`
		OkQty   int   `json:"ok_qty"`
		BadQty  int   `json:"bad_qty"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	if err := c.orderService.CreateMoldUsageRecord(body.OrderID, body.Shots, body.OkQty, body.BadQty); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// GetMoldUsageHistory 获取模具使用历史
func (c *ProductionController) GetMoldUsageHistory(ctx *gin.Context) {
	// 从 URL 参数获取 mold_id（可选）
	moldIDStr := ctx.Param("mold_id")
	var moldID int64 = 0

	if moldIDStr != "" && moldIDStr != "0" {
		var err error
		moldID, err = strconv.ParseInt(moldIDStr, 10, 64)
		if err != nil {
			response.Error(ctx, response.BAD_REQUEST, "无效的模具ID")
			return
		}
	}

	// 支持按模具编号模糊查询（当未指定 mold_id 时使用）
	moldCode := ctx.Query("mold_code")
	page, pageSize := parsePageParams(ctx)

	records, total, err := c.orderService.GetMoldUsageHistory(moldID, page, pageSize, moldCode)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetMoldUsageStats 获取模具使用统计
func (c *ProductionController) GetMoldUsageStats(ctx *gin.Context) {
	moldIDStr := ctx.Param("mold_id")
	moldID, err := strconv.ParseInt(moldIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的模具ID")
		return
	}

	stats, err := c.orderService.GetMoldUsageStats(moldID)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, stats)
}

// GetOutsourcingRecordsByType 根据外发类型分页获取外发记录列表
func (c *ProductionController) GetOutsourcingRecordsByType(ctx *gin.Context) {
	outsourceTypeStr := ctx.Query("outsource_type")
	statusStr := ctx.Query("status")
	recordCodeStr := ctx.Query("record_code")
	processorIDStr := ctx.Query("processor_id")
	page, pageSize := parsePageParams(ctx)

	var processorID *int64
	if processorIDStr != "" {
		if id, err := strconv.ParseInt(processorIDStr, 10, 64); err == nil {
			processorID = &id
		}
	}

	daoInstance := dao.NewOutsourcingRecordDAO()
	records, total, err := daoInstance.FindByTypeWithPage(outsourceTypeStr, statusStr, recordCodeStr, processorID, page, pageSize)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	// 批量查询进度
	reportDAO := dao.NewOutsourcingReportDAO()
	ids := make([]int64, len(records))
	for i, r := range records {
		ids[i] = r.ID
	}
	progressMap, _ := reportDAO.GetOutsourcingTotalQtyBatch(ids)

	// 批量查询模具外发的工单工序状态
	workOrderProcessMap := make(map[int64][]entity.WorkOrderProcess)
	if outsourceTypeStr == "MOLD" {
		workOrderIDs := make([]int64, 0, len(records))
		for _, r := range records {
			if r.WorkOrderID != nil {
				workOrderIDs = append(workOrderIDs, *r.WorkOrderID)
			}
		}
		if len(workOrderIDs) > 0 {
			processDAO := dao.NewWorkOrderProcessDAO()
			processes, _ := processDAO.FindByWorkOrderIDs(workOrderIDs)
			for _, p := range processes {
				workOrderProcessMap[p.WorkOrderID] = append(workOrderProcessMap[p.WorkOrderID], *p)
			}
		}
	}

	// 丰富列表数据
	enriched := make([]gin.H, len(records))
	// ✨ 生产外发：批量收集订单关联的模具ID，查询模具编号映射
	moldCodeMap := make(map[int64]string)
	if outsourceTypeStr == "PRODUCTION" {
		moldIDSet := make(map[int64]struct{})
		for _, r := range records {
			if r.Order == nil {
				continue
			}
			// 兜底：从明细聚合产品名称（兼容历史数据 prod_order.product_name 为空）
			if r.Order.ProductName == "" && len(r.Order.Items) > 0 {
				names := make([]string, 0, len(r.Order.Items))
				for _, it := range r.Order.Items {
					if it.ProductName != "" {
						names = append(names, it.ProductName)
					}
				}
				switch len(names) {
				case 1:
					r.Order.ProductName = names[0]
				default:
					r.Order.ProductName = fmt.Sprintf("%s 等 %d 项产品", names[0], len(names))
				}
			}
			// 兜底：外发总金额 = 外发单价 × 计划数量（兼容历史数据 total_amount=0）
			if r.TotalAmount == 0 && r.Order != nil && r.Order.OutsourcePrice > 0 && r.Order.PlanQty > 0 {
				r.TotalAmount = r.Order.OutsourcePrice * float64(r.Order.PlanQty)
			}
			// 订单主表的模具
			if r.Order.MoldID != nil && *r.Order.MoldID > 0 {
				moldIDSet[*r.Order.MoldID] = struct{}{}
			}
			// 订单明细中的模具
			for _, item := range r.Order.Items {
				if item.MoldID != nil && *item.MoldID > 0 {
					moldIDSet[*item.MoldID] = struct{}{}
				}
				for _, mid := range item.MoldIDs {
					if mid > 0 {
						moldIDSet[mid] = struct{}{}
					}
				}
			}
		}
		if len(moldIDSet) > 0 {
			ids := make([]int64, 0, len(moldIDSet))
			for id := range moldIDSet {
				ids = append(ids, id)
			}
			moldDAO := dao.NewMoldDAO()
			if m, err := moldDAO.FindCodesByIDs(ids); err == nil {
				moldCodeMap = m
			} else {
				logger.Warn("批量查询模具编号失败", zap.Error(err))
			}
		}
	}

	for i, r := range records {
		completed := progressMap[r.ID]
		percent := 0
		if r.Quantity > 0 {
			percent = completed * 100 / r.Quantity
			if percent > 100 {
				percent = 100
			}
		}

		// 模具外发：附带工单工序状态
		var processList []entity.WorkOrderProcess
		if r.WorkOrderID != nil {
			processList = workOrderProcessMap[*r.WorkOrderID]
		}

		// ✨ 生产外发：收集关联生产订单的模具编号（去重保序）
		var moldCodes []string
		var moldItems []gin.H
		if outsourceTypeStr == "PRODUCTION" && r.Order != nil {
			seen := make(map[string]struct{})
			addCode := func(id int64, code string, isExternal bool) {
				if code == "" {
					return
				}
				if _, ok := seen[code]; ok {
					return
				}
				seen[code] = struct{}{}
				moldCodes = append(moldCodes, code)
				// 外部模具（ExternalMoldCode）没有对应 mold.id，不返回 id
				moldID := int64(0)
				if !isExternal {
					moldID = id
				}
				moldItems = append(moldItems, gin.H{
					"id":         moldID,
					"mold_code":  code,
					"is_external": isExternal,
				})
			}
			// 1) 订单主表 MoldID
			if r.Order.MoldID != nil && *r.Order.MoldID > 0 {
				if code, ok := moldCodeMap[*r.Order.MoldID]; ok {
					addCode(*r.Order.MoldID, code, false)
				}
			}
			// 2) 订单主表 ExternalMoldCode（外来模具）
			if r.Order.ExternalMoldCode != nil && *r.Order.ExternalMoldCode != "" {
				addCode(0, *r.Order.ExternalMoldCode, true)
			}
			// 3) 订单明细 MoldID / MoldIDs / ExternalMoldCode
			for _, item := range r.Order.Items {
				if item.MoldID != nil && *item.MoldID > 0 {
					if code, ok := moldCodeMap[*item.MoldID]; ok {
						addCode(*item.MoldID, code, false)
					}
				}
				for _, mid := range item.MoldIDs {
					if mid > 0 {
						if code, ok := moldCodeMap[mid]; ok {
							addCode(mid, code, false)
						}
					}
				}
				if item.ExternalMoldCode != nil && *item.ExternalMoldCode != "" {
					addCode(0, *item.ExternalMoldCode, true)
				}
			}
		}

		enriched[i] = gin.H{
			"id":                 r.ID,
			"record_code":        r.RecordCode,
			"order_id":           r.OrderID,
			"work_order_id":      r.WorkOrderID,
			"outsource_type":     r.OutsourceType,
			"processor_id":       r.ProcessorID,
			"mold_id":            r.MoldID,
			"external_mold_info": r.ExternalMoldInfo,
			"outsource_date":     r.OutsourceDate,
			"expected_return_date": r.ExpectedReturnDate,
			"actual_return_date": r.ActualReturnDate,
			"quantity":           r.Quantity,
			"unit_price":         r.UnitPrice,
			"total_amount":       r.TotalAmount,
			"paid_amount":        r.PaidAmount,
			"logistics_company":  r.LogisticsCompany,
			"tracking_no":        r.TrackingNo,
			"delivery_address":   r.DeliveryAddress,
			"receiver_name":      r.ReceiverName,
			"receiver_phone":     r.ReceiverPhone,
			"status":             r.Status,
			"remark":             r.Remark,
			"share_token":        r.ShareToken,
			"create_user_id":     r.CreateUserID,
			"create_time":        r.CreateTime,
			"update_time":        r.UpdateTime,
			"order":              r.Order,
			"processor":          r.Processor,
			"mold":               r.Mold,
			"work_order":         r.WorkOrder,
			"progress_qty":       completed,
			"progress_percent":   percent,
			"work_order_processes": processList,
			"mold_codes":           moldCodes,
			"mold_items":           moldItems,
		}
	}

	response.Success(ctx, gin.H{
		"list":      enriched,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetStockInList 获取生产入库列表
func (c *ProductionController) GetStockInList(ctx *gin.Context) {
	orderCode := ctx.Query("order_code")
	productName := ctx.Query("product_name")
	stockCode := ctx.Query("stock_code")
	page, pageSize := parsePageParams(ctx)

	list, total, err := c.orderService.GetStockInList(page, pageSize, orderCode, productName, stockCode)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetStockInByID 获取入库记录详情
func (c *ProductionController) GetStockInByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的入库记录ID")
		return
	}

	stockIn, err := c.orderService.GetStockInByID(id)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "入库记录不存在")
		return
	}

	response.Success(ctx, stockIn)
}

// UpdateStockIn 更新入库记录
func (c *ProductionController) UpdateStockIn(ctx *gin.Context) {
	ctx.Set("oper_log_title", "更新入库记录")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "production_stock_in")

	var req struct {
		ID          int64  `json:"id"`
		Quantity    int    `json:"quantity"`
		StockInDate string `json:"stock_in_date"`
		BatchNo     string `json:"batch_no"`
		Remark      string `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	var stockInDate *time.Time
	if req.StockInDate != "" {
		// 使用中国时区（东八区）解析时间，避免时区偏移
		loc, _ := time.LoadLocation("Asia/Shanghai")
		parsedDate, err := time.ParseInLocation("2006-01-02 15:04:05", req.StockInDate, loc)
		if err != nil {
		parsedDate, err = time.ParseInLocation("2006-01-02", req.StockInDate, loc)
		}
		if err != nil {
			response.Error(ctx, response.BAD_REQUEST, "入库日期格式错误，请使用 YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD 格式")
			return
		}
		stockInDate = &parsedDate
	}

	stockIn := &entity.ProductionStockIn{
		ID:       req.ID,
		Quantity: req.Quantity,
		BatchNo:  req.BatchNo,
		Remark:   req.Remark,
	}
	if stockInDate != nil {
		stockIn.StockInDate = stockInDate
	}

	if err := c.orderService.UpdateStockIn(stockIn); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// GetStockStatistics 获取库存统计数据
func (c *ProductionController) GetStockStatistics(ctx *gin.Context) {
	stats, err := c.orderService.GetStockStatistics()
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, stats)
}

// GetStockTrend 获取库存趋势数据
func (c *ProductionController) GetStockTrend(ctx *gin.Context) {
	daysStr := ctx.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)

	trend, err := c.orderService.GetStockTrend(days)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, trend)
}

// GetProductStockList 获取产品库存汇总列表
func (c *ProductionController) GetProductStockList(ctx *gin.Context) {
	moldCode := ctx.Query("mold_code")
	productName := ctx.Query("product_name")
	page, pageSize := parsePageParams(ctx)

	list, total, err := c.orderService.GetProductStockList(page, pageSize, moldCode, productName)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetStockRecordsByProductID 获取指定产品的出入库记录
func (c *ProductionController) GetStockRecordsByProductID(ctx *gin.Context) {
	moldIDStr := ctx.Param("mold_id")
	moldID, err := strconv.ParseInt(moldIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的模具ID")
		return
	}

	// 从 query 参数中获取 product_id（可选）
	var productID *int64
	if productIDStr := ctx.Query("product_id"); productIDStr != "" {
		if pid, parseErr := strconv.ParseInt(productIDStr, 10, 64); parseErr == nil {
			productID = &pid
		}
	}

	sortOrder := ctx.DefaultQuery("sort_order", "desc")     // 默认降序
	statusFilters := ctx.DefaultQuery("status_filters", "") // 多选状态筛选，逗号分隔

	page, pageSize := parsePageParams(ctx)

	records, total, summary, err := c.orderService.GetStockRecordsByMoldIDAndProduct(moldID, productID, page, pageSize, sortOrder, statusFilters)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":      records,
		"total":     total,
		"summary":   summary,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetStockRecordsByExternalMoldCode 获取指定外来模具的出入库记录
func (c *ProductionController) GetStockRecordsByExternalMoldCode(ctx *gin.Context) {
	externalMoldCode := ctx.Param("external_mold_code")
	if externalMoldCode == "" {
		response.Error(ctx, response.BAD_REQUEST, "外来模具编号不能为空")
		return
	}

	// 从 query 参数中获取 product_id（可选）
	var productID *int64
	if productIDStr := ctx.Query("product_id"); productIDStr != "" {
		if pid, parseErr := strconv.ParseInt(productIDStr, 10, 64); parseErr == nil {
			productID = &pid
		}
	}

	sortOrder := ctx.DefaultQuery("sort_order", "desc")     // 默认降序
	statusFilters := ctx.DefaultQuery("status_filters", "") // 多选状态筛选，逗号分隔

	page, pageSize := parsePageParams(ctx)

	records, total, summary, err := c.orderService.GetStockRecordsByExternalMoldCodeAndProduct(externalMoldCode, productID, page, pageSize, sortOrder, statusFilters)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":      records,
		"total":     total,
		"summary":   summary,
		"page":      page,
		"page_size": pageSize,
	})
}

// CreateStockDirectly 直接创建库存（用于历史库存导入或手动调整）
func (c *ProductionController) CreateStockDirectly(ctx *gin.Context) {
	ctx.Set("oper_log_title", "直接创建库存")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "stock")

	var req struct {
		MoldID           *int64  `json:"mold_id"`            // 自有模具 ID
		ExternalMoldCode *string `json:"external_mold_code"` // 外来模具编号
		ProductID        *int64  `json:"product_id"`         // 产品ID（可选）
		WarehouseID      int64   `json:"warehouse_id" binding:"required"`
		Quantity         float64 `json:"quantity" binding:"required,min=0"`
		SafeStock        float64 `json:"safe_stock"`
		Remark           string  `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	// 验证：必须提供 mold_id 或 external_mold_code 之一
	if req.MoldID == nil && (req.ExternalMoldCode == nil || *req.ExternalMoldCode == "") {
		response.Error(ctx, response.BAD_REQUEST, "请提供模具ID或外来模具编号")
		return
	}

	var materialID int64
	var err error
	var productName string

	if req.MoldID != nil && *req.MoldID > 0 {
		// 自有模具：获取或创建对应的物料记录
		materialID, err = c.orderService.EnsureMoldMaterial(*req.MoldID, req.ProductID)
		if err != nil {
			response.Error(ctx, response.ERROR, "获取物料信息失败: "+err.Error())
			return
		}

		moldDAO := dao.NewMoldDAO()
		mold, err := moldDAO.FindByID(*req.MoldID)
		if err == nil && mold != nil {
			productName = mold.MoldName
		}

		// 如果传入了ProductID，从Product表中获取产品名称
		if req.ProductID != nil && *req.ProductID > 0 {
			productDAO := dao.NewProductDAO()
			product, err := productDAO.FindByID(*req.ProductID)
			if err == nil && product != nil {
				productName = product.ProName
			}
		}
	} else if req.ExternalMoldCode != nil && *req.ExternalMoldCode != "" {
		// 外来模具：获取或创建对应的物料记录
		materialID, err = c.orderService.EnsureExternalMoldMaterial(*req.ExternalMoldCode, productName)
		if err != nil {
			response.Error(ctx, response.ERROR, "获取外来模具物料信息失败: "+err.Error())
			return
		}

		// 外来模具使用编号作为产品名称
		productName = *req.ExternalMoldCode

		// 如果传入了ProductID，从Product表中获取产品名称
		if req.ProductID != nil && *req.ProductID > 0 {
			productDAO := dao.NewProductDAO()
			product, err := productDAO.FindByID(*req.ProductID)
			if err == nil && product != nil {
				productName = product.ProName
			}
		}
	}

	// 增加实际库存（stock 表）
	if err := c.orderService.InStockDirectly(materialID, req.WarehouseID, req.Quantity); err != nil {
		response.Error(ctx, response.ERROR, "更新库存失败: "+err.Error())
		return
	}

	// 【关键】同时创建入库记录（production_stock_in 表），以便库存列表能够显示
	stockCode, err := c.orderService.GenerateStockInCode()
	if err != nil {
		response.Error(ctx, response.ERROR, "生成入库单号失败: "+err.Error())
		return
	}

	// 自动生成批次号
	batchNo := c.orderService.GenerateBatchNo(req.MoldID, req.ExternalMoldCode, req.ProductID)

	now := time.Now()
	// 从 JWT token 中获取用户ID
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "未登录或会话已过期")
		return
	}
	quantityInt := int(req.Quantity)

	stockIn := &entity.ProductionStockIn{
		StockCode:        stockCode,
		OrderID:          nil, // 直接入库没有关联订单
		WarehouseID:      req.WarehouseID,
		MoldID:           req.MoldID,
		ExternalMoldCode: req.ExternalMoldCode,
		ProductID:        req.ProductID, // 保存产品ID
		ProductName:      productName,
		Quantity:         quantityInt,
		BatchNo:          batchNo, // 使用自动生成的批次号
		StockInDate:      &now,
		OperatorID:       &userID,
		Remark:           req.Remark,
	}

	stockInDAO := dao.NewProductionStockInDAO()
	if err := stockInDAO.Create(stockIn); err != nil {
		response.Error(ctx, response.ERROR, "创建入库记录失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message":     "库存创建成功",
		"stock_in_id": stockIn.ID,
	})
}

// CheckStock 盘点库存
func (c *ProductionController) CheckStock(ctx *gin.Context) {
	ctx.Set("oper_log_title", "盘点库存")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "stock")

	var req struct {
		MoldID           *int64  `json:"mold_id"`
		ExternalMoldCode *string `json:"external_mold_code"`
		ProductID        *int64  `json:"product_id"`
		WarehouseID      int64   `json:"warehouse_id" binding:"required"`
		ActualQuantity   float64 `json:"actual_quantity" binding:"required,min=0"`
		OutReason        string  `json:"out_reason"`
		Remark           string  `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	// 验证：必须提供 mold_id 或 external_mold_code 之一
	if req.MoldID == nil && (req.ExternalMoldCode == nil || *req.ExternalMoldCode == "") {
		response.Error(ctx, response.BAD_REQUEST, "请提供模具ID或外来模具编号")
		return
	}

	var materialID int64
	var err error
	var productName string

	if req.MoldID != nil && *req.MoldID > 0 {
		// 自有模具：获取或创建对应的物料记录
		materialID, err = c.orderService.EnsureMoldMaterial(*req.MoldID, req.ProductID)
		if err != nil {
			response.Error(ctx, response.ERROR, "获取物料信息失败: "+err.Error())
			return
		}

		moldDAO := dao.NewMoldDAO()
		mold, err := moldDAO.FindByID(*req.MoldID)
		if err == nil && mold != nil {
			productName = mold.MoldName
		}

		if req.ProductID != nil && *req.ProductID > 0 {
			productDAO := dao.NewProductDAO()
			product, err := productDAO.FindByID(*req.ProductID)
			if err == nil && product != nil {
				productName = product.ProName
			}
		}
	} else if req.ExternalMoldCode != nil && *req.ExternalMoldCode != "" {
		materialID, err = c.orderService.EnsureExternalMoldMaterial(*req.ExternalMoldCode, productName)
		if err != nil {
			response.Error(ctx, response.ERROR, "获取外来模具物料信息失败: "+err.Error())
			return
		}
		productName = *req.ExternalMoldCode

		if req.ProductID != nil && *req.ProductID > 0 {
			productDAO := dao.NewProductDAO()
			product, err := productDAO.FindByID(*req.ProductID)
			if err == nil && product != nil {
				productName = product.ProName
			}
		}
	}

	// 获取当前库存用于日志
	currentStock, _ := c.orderService.GetStockByMaterialProductAndWarehouse(materialID, req.ProductID, req.WarehouseID)
	var currentQty float64
	if currentStock != nil {
		currentQty = currentStock.StockQty
	}

	// 盘点：直接设置库存为实际盘点数量
	diff, err := c.orderService.CheckStock(materialID, req.ProductID, req.WarehouseID, req.ActualQuantity)
	if err != nil {
		response.Error(ctx, response.ERROR, "盘点失败: "+err.Error())
		return
	}

	// 记录盘点调整日志（创建入库或出库记录）
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "未登录或会话已过期")
		return
	}

	now := time.Now()

	if diff != 0 {
		var checkRemark string
		if diff > 0 {
			checkRemark = "盘点调整(盘盈): 原库存 " + fmt.Sprintf("%.2f", currentQty) + ", 实际 " + fmt.Sprintf("%.2f", req.ActualQuantity)
		} else {
			checkRemark = "盘点调整(盘亏): 原库存 " + fmt.Sprintf("%.2f", currentQty) + ", 实际 " + fmt.Sprintf("%.2f", req.ActualQuantity)
		}
		if req.Remark != "" {
			checkRemark += "; " + req.Remark
		}

		if diff > 0 {
			// 盘盈：创建入库记录
			stockCode, err := c.orderService.GenerateStockInCode()
			if err != nil {
				response.Error(ctx, response.ERROR, "生成入库单号失败: "+err.Error())
				return
			}
			batchNo := c.orderService.GenerateBatchNo(req.MoldID, req.ExternalMoldCode, req.ProductID)
			quantityInt := int(diff)

			stockIn := &entity.ProductionStockIn{
				StockCode:        stockCode,
				OrderID:          nil,
				WarehouseID:      req.WarehouseID,
				MoldID:           req.MoldID,
				ExternalMoldCode: req.ExternalMoldCode,
				ProductID:        req.ProductID,
				ProductName:      productName,
				Quantity:         quantityInt,
				BatchNo:          batchNo,
				StockInDate:      &now,
				OperatorID:       &userID,
				Remark:           checkRemark,
			}
			stockInDAO := dao.NewProductionStockInDAO()
			if err := stockInDAO.Create(stockIn); err != nil {
				response.Error(ctx, response.ERROR, "创建盘点入库记录失败: "+err.Error())
				return
			}
		} else if diff < 0 {
			// 盘亏：创建出库记录
			stockCode, err := c.orderService.GenerateStockInCode()
			if err != nil {
				response.Error(ctx, response.ERROR, "生成出库单号失败: "+err.Error())
				return
			}
			quantityInt := int(-diff)
			outType := "other"
			outReason := req.OutReason
			if outReason == "" {
				outReason = "check"
			}

		stockOutRecord := &entity.ProductionStockOut{
			StockCode:        stockCode,
			OrderID:          nil,
			WarehouseID:      req.WarehouseID,
			MoldID:           req.MoldID,
			ExternalMoldCode: req.ExternalMoldCode,
			ProductID:        req.ProductID,
			ProductName:      productName,
			Quantity:         quantityInt,
			StockOutDate:     &now,
			OperatorID:       &userID,
			Remark:           checkRemark,
			OutType:          outType,
			OutReason:        &outReason,
		}
		stockOutDAO := dao.NewProductionStockOutDAO()
		if err := stockOutDAO.Create(stockOutRecord); err != nil {
				response.Error(ctx, response.ERROR, "创建盘点出库记录失败: "+err.Error())
				return
			}
		}
	}

	response.Success(ctx, gin.H{
		"message":       "盘点成功",
		"diff":          diff,
		"current_qty":   currentQty,
		"actual_qty":    req.ActualQuantity,
	})
}

// DeleteStock 删除库存记录
func (c *ProductionController) DeleteStock(ctx *gin.Context) {
	ctx.Set("oper_log_title", "删除库存记录")
	ctx.Set("oper_log_type", 2)
	ctx.Set("oper_log_business_type", "stock")

	var req struct {
		MoldID           *int64  `json:"mold_id"`
		ExternalMoldCode *string `json:"external_mold_code"`
		ProductID        *int64  `json:"product_id"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	if err := c.orderService.DeleteStockByMoldAndProduct(req.MoldID, req.ExternalMoldCode, req.ProductID); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"message": "删除成功",
	})
}

// StockOutOrder 生产订单出库
func (c *ProductionController) StockOutOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "生产出库")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "production_stock_out")

	var req struct {
		OrderID          int64    `json:"order_id"`
		WarehouseID      int64    `json:"warehouse_id" binding:"required"`
		MoldID           *int64   `json:"mold_id"`
		ExternalMoldCode *string  `json:"external_mold_code"`
		ProductID        *int64   `json:"product_id"`
		ProductName      string   `json:"product_name" binding:"required"`
		Quantity         int      `json:"quantity" binding:"required,min=1"`
		BatchNo          string   `json:"batch_no"`
		Remark           string   `json:"remark"`
		CustomerName     *string  `json:"customer_name"`
		DeliveryDate     *string  `json:"delivery_date"`
		DeliveryMethod   *string  `json:"delivery_method"`
		ExpressCompany   *string  `json:"express_company"`
		ExpressNo        *string  `json:"express_no"`
		Photos           *string  `json:"photos"`
		TotalWeight      *float64 `json:"total_weight"`
		DriverPhone      *string  `json:"driver_phone"`
		DriverPlate      *string  `json:"driver_plate"`
		OutType          string   `json:"out_type"`          // 出库类型(normal=正常发货, other=其他原因出库)
		OutReason        *string  `json:"out_reason"`        // 出库原因
		OutReasonRemark  *string  `json:"out_reason_remark"` // 出库原因备注
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	if req.ProductName == "" {
		response.Error(ctx, response.BAD_REQUEST, "产品名称不能为空")
		return
	}
	if req.CustomerName == nil || *req.CustomerName == "" {
		response.Error(ctx, response.BAD_REQUEST, "接收客户不能为空")
		return
	}
	if req.DeliveryDate == nil || *req.DeliveryDate == "" {
		response.Error(ctx, response.BAD_REQUEST, "发货日期不能为空")
		return
	}
	if req.DeliveryMethod == nil || *req.DeliveryMethod == "" {
		response.Error(ctx, response.BAD_REQUEST, "发货方式不能为空")
		return
	}

	var deliveryDate *time.Time
	if req.DeliveryDate != nil && *req.DeliveryDate != "" {
		// 使用中国时区（东八区）解析时间，避免时区偏移
		loc, _ := time.LoadLocation("Asia/Shanghai")
		parsedDate, err := time.ParseInLocation("2006-01-02 15:04:05", *req.DeliveryDate, loc)
		if err != nil {
			parsedDate, err = time.ParseInLocation("2006-01-02", *req.DeliveryDate, loc)
		}
		if err != nil {
			response.Error(ctx, response.BAD_REQUEST, "发货日期格式错误，请使用 YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD 格式")
			return
		}
		deliveryDate = &parsedDate
	}

	// 设置默认出库类型
	if req.OutType == "" {
		req.OutType = "normal"
	}

	if req.OutType != "normal" && req.OutType != "other" {
		response.Error(ctx, response.BAD_REQUEST, "出库类型错误，必须为 normal 或 other")
		return
	}

	// 如果是其他原因出库，必须填写出库原因
	if req.OutType == "other" && (req.OutReason == nil || *req.OutReason == "") {
		response.Error(ctx, response.BAD_REQUEST, "请选择出库原因")
		return
	}

	// 从 JWT token 中获取用户ID
	userID := middleware.GetUserID(ctx)
	if userID == 0 {
		response.Error(ctx, response.UNAUTHORIZED, "未登录或会话已过期")
		return
	}

	if err := c.orderService.StockOutOrder(req.OrderID, req.WarehouseID, req.MoldID, req.ExternalMoldCode, req.ProductID, req.ProductName, req.Quantity, req.BatchNo, req.Remark, req.CustomerName, deliveryDate, req.DeliveryMethod, req.ExpressCompany, req.ExpressNo, req.Photos, req.TotalWeight, req.DriverPhone, req.DriverPlate, req.OutType, req.OutReason, req.OutReasonRemark, userID); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"message": "出库成功",
	})

	// 通知：生产出库
	orderCode := getOrderCodeOrID(req.OrderID, c)
	productName := getOrderProductName(req.OrderID, c)
	NotifyOperatorAndAdmins(userID, "生产订单",
		fmt.Sprintf("订单 %s ，产品 %s 已出库，数量：%d", orderCode, productName, req.Quantity),
		"production:stock_out", "middle",
		"", req.OrderID)
}

// GetStockOutList 获取出库记录列表
func (c *ProductionController) GetStockOutList(ctx *gin.Context) {
	orderCode := ctx.Query("order_code")
	productName := ctx.Query("product_name")
	stockCode := ctx.Query("stock_code")
	page, pageSize := parsePageParams(ctx)

	list, total, err := c.orderService.GetStockOutList(page, pageSize, orderCode, productName, stockCode)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetStockOutByID 获取出库记录详情
func (c *ProductionController) GetStockOutByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的出库记录ID")
		return
	}

	stockOut, err := c.orderService.GetStockOutByID(id)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "出库记录不存在")
		return
	}

	response.Success(ctx, stockOut)
}

// GetStockOutRecordsByMoldID 获取指定模具的出库记录
func (c *ProductionController) GetStockOutRecordsByMoldID(ctx *gin.Context) {
	moldIDStr := ctx.Param("mold_id")
	moldID, err := strconv.ParseInt(moldIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的模具ID")
		return
	}

	page, pageSize := parsePageParams(ctx)

	records, total, err := c.orderService.GetStockOutRecordsByMoldID(moldID, page, pageSize)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetCurrentStock 获取当前库存数量（用于出库验证）
// 支持 product_id 或 product_name 查询（兼容历史数据 product_id 为 NULL 的场景）
func (c *ProductionController) GetCurrentStock(ctx *gin.Context) {
	var req struct {
		MoldID           *int64  `json:"mold_id"`
		ExternalMoldCode string  `json:"external_mold_code"`
		ProductID        *int64  `json:"product_id"`         // 可选，优先使用
		ProductName      string  `json:"product_name"`       // 可选，当 product_id 为空时使用
		WarehouseID      int64   `json:"warehouse_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	// 至少需要 product_id 或 product_name 其中一个
	if (req.ProductID == nil || *req.ProductID == 0) && req.ProductName == "" {
		response.Error(ctx, response.BAD_REQUEST, "产品信息不能为空")
		return
	}

	// 根据 production_stock_in 和 production_stock_out 动态计算库存
	// 以产品 + 仓库为主键，模具为可选过滤条件

	var totalStockIn int
	var totalStockOut int
	var err error

	if req.ProductID != nil && *req.ProductID > 0 {
		// 按 product_id 查询（标准路径）
		stockInQuery := dao.GetDB().Table("production_stock_in").
			Select("COALESCE(SUM(quantity), 0)").
			Where("product_id = ?", *req.ProductID).
			Where("warehouse_id = ?", req.WarehouseID)

		stockOutQuery := dao.GetDB().Model(&entity.ProductionStockOut{}).
			Select("COALESCE(SUM(quantity), 0)").
			Where("product_id = ?", *req.ProductID).
			Where("warehouse_id = ?", req.WarehouseID)

		if req.MoldID != nil && *req.MoldID > 0 {
			stockInQuery = stockInQuery.Where("mold_id = ?", *req.MoldID)
			stockOutQuery = stockOutQuery.Where("mold_id = ?", *req.MoldID)
		} else if req.ExternalMoldCode != "" {
			stockInQuery = stockInQuery.Where("external_mold_code = ?", req.ExternalMoldCode)
			stockOutQuery = stockOutQuery.Where("external_mold_code = ?", req.ExternalMoldCode)
		}

		err = stockInQuery.Scan(&totalStockIn).Error
		if err != nil {
			response.Error(ctx, response.ERROR, "查询入库数量失败: "+err.Error())
			return
		}
		err = stockOutQuery.Scan(&totalStockOut).Error
		if err != nil {
			response.Error(ctx, response.ERROR, "查询出库数量失败: "+err.Error())
			return
		}
	} else {
		// 按 product_name 查询（兼容历史数据 product_id 为 NULL）
		stockInQuery := dao.GetDB().Table("production_stock_in").
			Select("COALESCE(SUM(quantity), 0)").
			Where("product_name = ?", req.ProductName).
			Where("warehouse_id = ?", req.WarehouseID)

		stockOutQuery := dao.GetDB().Model(&entity.ProductionStockOut{}).
			Select("COALESCE(SUM(quantity), 0)").
			Where("product_name = ?", req.ProductName).
			Where("warehouse_id = ?", req.WarehouseID)

		if req.MoldID != nil && *req.MoldID > 0 {
			stockInQuery = stockInQuery.Where("mold_id = ?", *req.MoldID)
			stockOutQuery = stockOutQuery.Where("mold_id = ?", *req.MoldID)
		} else if req.ExternalMoldCode != "" {
			stockInQuery = stockInQuery.Where("external_mold_code = ?", req.ExternalMoldCode)
			stockOutQuery = stockOutQuery.Where("external_mold_code = ?", req.ExternalMoldCode)
		}

		err = stockInQuery.Scan(&totalStockIn).Error
		if err != nil {
			response.Error(ctx, response.ERROR, "查询入库数量失败: "+err.Error())
			return
		}
		err = stockOutQuery.Scan(&totalStockOut).Error
		if err != nil {
			response.Error(ctx, response.ERROR, "查询出库数量失败: "+err.Error())
			return
		}
	}

	currentStock := totalStockIn - totalStockOut
	if currentStock < 0 {
		currentStock = 0
	}

	response.Success(ctx, gin.H{
		"current_stock":   currentStock,
		"available_qty":   currentStock,
		"total_stock_in":  totalStockIn,
		"total_stock_out": totalStockOut,
	})
}

// GetStockOutRecordsByExternalMoldCode 获取指定外来模具的出库记录
func (c *ProductionController) GetStockOutRecordsByExternalMoldCode(ctx *gin.Context) {
	externalMoldCode := ctx.Param("external_mold_code")
	if externalMoldCode == "" {
		response.Error(ctx, response.BAD_REQUEST, "外来模具编号不能为空")
		return
	}

	page, pageSize := parsePageParams(ctx)

	records, total, err := c.orderService.GetStockOutRecordsByExternalMoldCode(externalMoldCode, page, pageSize)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// RecognizeExpressCompany 识别快递公司
// 基于淘宝开放平台（2024.09.27更新）的快递单号正则规则进行匹配
func (c *ProductionController) RecognizeExpressCompany(ctx *gin.Context) {
	expressNo := ctx.Query("express_no")
	if expressNo == "" {
		response.Error(ctx, response.BAD_REQUEST, "快递单号不能为空")
		return
	}

	companyCode, companyName := recognizeExpressByRegex(expressNo)
	response.Success(ctx, gin.H{
		"express_company":      companyCode,
		"express_company_name": companyName,
	})
}

// UpdateStockOut 更新出库记录
func (c *ProductionController) UpdateStockOut(ctx *gin.Context) {
	ctx.Set("oper_log_title", "更新出库记录")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "production_stock_out")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的出库记录ID")
		return
	}

	var req struct {
		StockOutDate   *string  `json:"stock_out_date"`
		CustomerName   *string  `json:"customer_name"`
		DeliveryDate   *string  `json:"delivery_date"`
		DeliveryMethod *string  `json:"delivery_method"`
		ExpressCompany *string  `json:"express_company"`
		ExpressNo      *string  `json:"express_no"`
		Photos         *string  `json:"photos"`
		TotalWeight    *float64 `json:"total_weight"`
		DriverPhone    *string  `json:"driver_phone"`
		DriverPlate    *string  `json:"driver_plate"`
		Remark         string   `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	var stockOutDate *time.Time
	if req.StockOutDate != nil && *req.StockOutDate != "" {
		// 使用中国时区（东八区）解析时间，避免时区偏移
		loc, _ := time.LoadLocation("Asia/Shanghai")
		parsedDate, err := time.ParseInLocation("2006-01-02 15:04:05", *req.StockOutDate, loc)
		if err != nil {
		parsedDate, err = time.ParseInLocation("2006-01-02", *req.StockOutDate, loc)
		}
		if err != nil {
			response.Error(ctx, response.BAD_REQUEST, "出库日期格式错误，请使用 YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD 格式")
			return
		}
		stockOutDate = &parsedDate
	}

	var deliveryDate *time.Time
	if req.DeliveryDate != nil && *req.DeliveryDate != "" {
		// 使用中国时区（东八区）解析时间，避免时区偏移
		loc, _ := time.LoadLocation("Asia/Shanghai")
		parsedDate, err := time.ParseInLocation("2006-01-02 15:04:05", *req.DeliveryDate, loc)
		if err != nil {
			parsedDate, err = time.ParseInLocation("2006-01-02", *req.DeliveryDate, loc)
		}
		if err != nil {
			response.Error(ctx, response.BAD_REQUEST, "发货日期格式错误，请使用 YYYY-MM-DD HH:mm:ss 或 YYYY-MM-DD 格式")
			return
		}
		deliveryDate = &parsedDate
	}

	if err := c.orderService.UpdateStockOut(id, stockOutDate, req.CustomerName, deliveryDate, req.DeliveryMethod, req.ExpressCompany, req.ExpressNo, req.Photos, req.TotalWeight, req.DriverPhone, req.DriverPlate, req.Remark); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"message": "更新成功",
	})
}

// GetStockOutListForPrint 获取用于打印的出库记录列表
func (c *ProductionController) GetStockOutListForPrint(ctx *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required,min=1"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	records, err := c.orderService.GetStockOutListForPrint(req.IDs)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	// 验证是否为同一客户（如果有多个记录）
	if len(records) > 1 {
		firstCustomer := ""
		for _, record := range records {
			if record.CustomerName != nil {
				if firstCustomer == "" {
					firstCustomer = *record.CustomerName
				} else if *record.CustomerName != firstCustomer {
					response.Error(ctx, response.BAD_REQUEST, "合并发货的记录必须属于同一客户")
					return
				}
			}
		}
	}

	customerName := ""
	if len(records) > 0 && records[0].CustomerName != nil {
		customerName = *records[0].CustomerName
	}

	response.Success(ctx, gin.H{
		"records":       records,
		"customer_name": customerName,
	})
}

// ========== 通知辅助函数 ==========

// getOrderCodeOrID 获取订单编号（优先从数据库查询）
func getOrderCodeOrID(orderID int64, c *ProductionController) string {
	order, err := c.orderService.GetOrderByID(orderID)
	if err != nil || order == nil {
		return fmt.Sprintf("#%d", orderID)
	}
	if order.OrderCode != "" {
		return order.OrderCode
	}
	return fmt.Sprintf("#%d", orderID)
}

func getOrderProductName(orderID int64, c *ProductionController) string {
	order, err := c.orderService.GetOrderByID(orderID)
	if err != nil || order == nil {
		return fmt.Sprintf("#%d", orderID)
	}
	if order.ProductName != "" {
		return order.ProductName
	}
	// 主订单产品名为空时，从订单明细获取产品名称
	if len(order.Items) > 0 && order.Items[0].ProductName != "" {
		return order.Items[0].ProductName
	}
	return fmt.Sprintf("#%d", orderID)
}

// parsePageParams 安全解析分页参数，解析失败时使用默认值
func parsePageParams(ctx *gin.Context) (page int, pageSize int) {
	page, _ = strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(ctx.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	return
}

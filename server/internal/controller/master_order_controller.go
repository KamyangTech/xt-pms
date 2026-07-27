package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
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

type MasterOrderController struct {
	service *service.MasterOrderService
}

func NewMasterOrderController() *MasterOrderController {
	return &MasterOrderController{
		service: service.NewMasterOrderService(dao.GetDB()),
	}
}

// CreateCombinedOrder 创建订单（支持综合订单、仅模具、仅生产）
func (c *MasterOrderController) CreateCombinedOrder(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建综合订单")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "master_order")

	var req struct {
		Master      entity.MasterOrder      `json:"master"`
		MoldOrder   *entity.MoldOrder       `json:"mold_order,omitempty"`   // 可选
		ProdOrder   *entity.ProductionOrder `json:"prod_order,omitempty"`   // 可选
		PaymentPlan *dto.PaymentPlanDTO     `json:"payment_plan,omitempty"` // ✨ 款项配置（可选）
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		// ✨ 输出详细的绑定错误信息
		logger.Warn("综合订单创建参数绑定失败", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Debug("接收到的综合订单数据",
		zap.Any("master", req.Master),
		zap.Any("mold_order", req.MoldOrder),
		zap.Any("prod_order", req.ProdOrder),
		zap.Any("payment_plan", req.PaymentPlan))

	if req.MoldOrder != nil && req.MoldOrder.PlanStartDate != nil && req.MoldOrder.PlanEndDate != nil {
		if err := validatePlanDates(req.MoldOrder.PlanStartDate.Time, req.MoldOrder.PlanEndDate.Time); err != nil {
			response.Error(ctx, response.BAD_REQUEST, err.Error())
			return
		}
	}
	if req.ProdOrder != nil && req.ProdOrder.PlanStartDate != nil && req.ProdOrder.PlanEndDate != nil {
		if err := validatePlanDates(*req.ProdOrder.PlanStartDate, *req.ProdOrder.PlanEndDate); err != nil {
			response.Error(ctx, response.BAD_REQUEST, err.Error())
			return
		}
	}

	// ✨ 校验生产订单明细是否为空
	if req.ProdOrder != nil && len(req.ProdOrder.Items) == 0 {
		response.Error(ctx, response.BAD_REQUEST, "生产订单至少需要一条产品明细")
		return
	}

	// ✨ 校验生产订单明细的必填字段
	if req.ProdOrder != nil {
		for i, item := range req.ProdOrder.Items {
			itemLabel := fmt.Sprintf("第%d行产品", i+1)
			if item.ProductName == "" {
				response.Error(ctx, response.BAD_REQUEST, itemLabel+"的产品名称不能为空")
				return
			}
			if item.PlanQty <= 0 {
				response.Error(ctx, response.BAD_REQUEST, itemLabel+"的计划数量必须大于 0")
				return
			}
		}

		// ✨ 外发生产时校验加工商
		if req.ProdOrder.ProductionType == "OUTSOURCED" && req.ProdOrder.ProcessorID == nil {
			response.Error(ctx, response.BAD_REQUEST, "外发生产时，请选择加工商")
			return
		}
	}

	// 将指针类型转换为值类型（如果为 nil 则传 nil）
	var moldOrder entity.MoldOrder
	var prodOrder entity.ProductionOrder
	var moldOrderPtr *entity.MoldOrder
	var prodOrderPtr *entity.ProductionOrder

	if req.MoldOrder != nil {
		moldOrder = *req.MoldOrder
		moldOrderPtr = &moldOrder
	}

	if req.ProdOrder != nil {
		prodOrder = *req.ProdOrder
		prodOrderPtr = &prodOrder
	}

	if err := c.service.CreateCombinedOrder(&req.Master, moldOrderPtr, prodOrderPtr, req.PaymentPlan); err != nil {
		var httpErr *response.HttpError
		if errors.As(err, &httpErr) {
			logger.Warn("综合订单创建业务校验失败", zap.Error(err))
		} else {
			logger.Error("创建综合订单失败", zap.Error(err))
		}
		handleServiceError(ctx, err)
		return
	}

	logger.Info("综合订单创建成功", zap.String("order_code", req.Master.OrderCode))
	response.Success(ctx, gin.H{
		"message":    "订单创建成功",
		"order_code": req.Master.OrderCode,
	})

	// 通知：创建综合订单
	userID := middleware.GetUserID(ctx)
	orderTypeText := "综合订单"
	if req.MoldOrder != nil && req.ProdOrder == nil {
		orderTypeText = "模具订单"
	} else if req.MoldOrder == nil && req.ProdOrder != nil {
		orderTypeText = "生产订单"
	}
	// 统计模具数量和产品数量
	moldCount := 0
	prodCount := 0
	if req.MoldOrder != nil && req.MoldOrder.Items != nil {
		moldCount = len(req.MoldOrder.Items)
	}
	if req.ProdOrder != nil && req.ProdOrder.Items != nil {
		prodCount = len(req.ProdOrder.Items)
	}
	if orderTypeText == "模具订单" {
		NotifyOperatorAndAdmins(userID, "综合订单",
			fmt.Sprintf("%s %s 已创建，包含 %d 套模具", orderTypeText, req.Master.OrderCode, moldCount),
			"master_order:created", "low",
			"", req.Master.ID)
	} else {
		NotifyOperatorAndAdmins(userID, "综合订单",
			fmt.Sprintf("%s %s 已创建，包含 %d 套模具，%d 个产品", orderTypeText, req.Master.OrderCode, moldCount, prodCount),
			"master_order:created", "low",
			"", req.Master.ID)
	}
}

// GetDetail 获取综合订单详情
func (c *MasterOrderController) GetDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的订单ID")
		return
	}

	order, err := c.service.GetMasterOrderByID(id)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "订单不存在")
		return
	}

	response.Success(ctx, order)
}

// List 综合订单列表
func (c *MasterOrderController) List(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	orderCode := ctx.Query("order_code")
	customerName := ctx.Query("customer_name")
	createTimeStart := ctx.Query("create_time_start")
	createTimeEnd := ctx.Query("create_time_end")
	dueSoon := ctx.Query("due_soon")

	orders, total, err := c.service.List(page, pageSize, orderCode, customerName, createTimeStart, createTimeEnd, dueSoon)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{
		"list":  orders,
		"total": total,
	})
}

// Edit 编辑综合订单基本信息（仅允许修改交货日期和备注）
func (c *MasterOrderController) Edit(ctx *gin.Context) {
	ctx.Set("oper_log_title", "编辑综合订单")
	ctx.Set("oper_log_type", 2)
	ctx.Set("oper_log_business_type", "master_order")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的订单ID")
		return
	}

	var req struct {
		PlanEndDate *string `json:"plan_end_date"`
		Remark     string  `json:"remark"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "参数错误: "+err.Error())
		return
	}

	var planEnd *time.Time
	if req.PlanEndDate != nil && *req.PlanEndDate != "" {
		t, err := time.Parse("2006-01-02", *req.PlanEndDate)
		if err != nil {
			response.Error(ctx, response.BAD_REQUEST, "日期格式错误，应为 YYYY-MM-DD")
			return
		}
		planEnd = &t
	}

	if err := c.service.EditMasterOrder(id, planEnd, req.Remark); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{"message": "编辑成功"})
}

// Cancel 取消综合订单
func (c *MasterOrderController) Cancel(ctx *gin.Context) {
	ctx.Set("oper_log_title", "取消综合订单")
	ctx.Set("oper_log_type", 2)
	ctx.Set("oper_log_business_type", "master_order")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的订单ID")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "参数错误: "+err.Error())
		return
	}
	if req.Reason == "" {
		response.Error(ctx, response.ERROR, "请填写取消原因")
		return
	}

	userID := middleware.GetUserID(ctx)
	if err := c.service.CancelMasterOrder(id, req.Reason, userID); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{"message": "订单已取消"})
}

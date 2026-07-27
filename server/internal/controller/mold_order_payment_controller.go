package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type MoldOrderPaymentController struct {
	paymentService *service.MoldOrderPaymentService
}

func NewMoldOrderPaymentController() *MoldOrderPaymentController {
	return &MoldOrderPaymentController{
		paymentService: service.NewMoldOrderPaymentService(),
	}
}

// GetOrderPayments 获取订单的所有款项
func (ctrl *MoldOrderPaymentController) GetOrderPayments(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "订单 ID 格式错误")
		return
	}

	payments, err := ctrl.paymentService.GetPaymentsByOrderID(orderID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("获取款项列表失败: %v", err))
		return
	}

	response.Success(c, payments)
}

// GetPaymentsByWorkOrderID 根据工单ID获取款项列表
func (ctrl *MoldOrderPaymentController) GetPaymentsByWorkOrderID(c *gin.Context) {
	workOrderIDStr := c.Param("workOrderId")
	workOrderID, err := strconv.ParseInt(workOrderIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "工单 ID 格式错误")
		return
	}

	payments, err := ctrl.paymentService.GetPaymentsByWorkOrderID(workOrderID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("获取款项列表失败: %v", err))
		return
	}

	response.Success(c, payments)
}

// GetPaymentByID 根据 ID 获取单个款项记录
func (ctrl *MoldOrderPaymentController) GetPaymentByID(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseInt(paymentIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "款项 ID 格式错误")
		return
	}

	payment, err := ctrl.paymentService.GetPaymentByID(paymentID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "款项记录不存在")
		return
	}

	response.Success(c, payment)
}

// CreatePayment 新增款项记录
func (ctrl *MoldOrderPaymentController) CreatePayment(c *gin.Context) {
	var payment entity.MoldOrderPayment
	if err := c.ShouldBindJSON(&payment); err != nil {
		response.Error(c, http.StatusBadRequest, "参数格式错误")
		return
	}

	if err := ctrl.paymentService.CreatePayment(&payment); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("创建款项记录失败: %v", err))
		return
	}

	response.Success(c, payment)
}

// UpdatePayment 更新款项记录
func (ctrl *MoldOrderPaymentController) UpdatePayment(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseInt(paymentIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "款项 ID 格式错误")
		return
	}

	var payment entity.MoldOrderPayment
	if err := c.ShouldBindJSON(&payment); err != nil {
		response.Error(c, http.StatusBadRequest, "参数格式错误")
		return
	}

	payment.ID = paymentID
	if err := ctrl.paymentService.UpdatePayment(&payment); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("更新款项记录失败: %v", err))
		return
	}

	response.Success(c, payment)
}

// DeletePayment 删除款项记录
func (ctrl *MoldOrderPaymentController) DeletePayment(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseInt(paymentIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "款项 ID 格式错误")
		return
	}

	if err := ctrl.paymentService.DeletePayment(paymentID); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("删除款项记录失败: %v", err))
		return
	}

	response.Success(c, nil)
}

// ReplacePaymentsByWorkOrder 批量替换工单的所有款项记录（事务保护）
func (ctrl *MoldOrderPaymentController) ReplacePaymentsByWorkOrder(c *gin.Context) {
	var req struct {
		OrderID     int64                       `json:"order_id" binding:"required"`
		WorkOrderID int64                       `json:"work_order_id" binding:"required"`
		Items       []entity.MoldOrderPayment   `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数格式错误: "+err.Error())
		return
	}

	if err := ctrl.paymentService.ReplacePaymentsByWorkOrder(req.OrderID, req.WorkOrderID, req.Items); err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("替换款项记录失败: %v", err))
		return
	}

	response.Success(c, nil)
}

// GetPaymentStats 获取订单款项统计
func (ctrl *MoldOrderPaymentController) GetPaymentStats(c *gin.Context) {
	orderIDStr := c.Param("orderId")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "订单 ID 格式错误")
		return
	}

	stats, err := ctrl.paymentService.GetPaymentStats(orderID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("获取款项统计失败: %v", err))
		return
	}

	response.Success(c, stats)
}

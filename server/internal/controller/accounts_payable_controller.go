package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type AccountsPayableController struct {
	apService       *service.AccountsPayableService
	dictDataService *service.SysDictDataService
}

func NewAccountsPayableController() *AccountsPayableController {
	return &AccountsPayableController{
		apService:       service.NewAccountsPayableService(),
		dictDataService: service.NewSysDictDataService(),
	}
}

// List 获取应付账款列表
func (ctrl *AccountsPayableController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var supplierID *int64
	if sid := c.Query("supplier_id"); sid != "" {
		id, err := strconv.ParseInt(sid, 10, 64)
		if err == nil {
			supplierID = &id
		}
	}

	status := c.Query("status")

	list, total, err := ctrl.apService.GetPayableList(page, pageSize, supplierID, status)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// Detail 获取应付账款详情
func (ctrl *AccountsPayableController) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	ap, err := ctrl.apService.GetPayableByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "应付账款不存在")
		return
	}

	response.Success(c, ap)
}

// Create 创建应付账款
func (ctrl *AccountsPayableController) Create(c *gin.Context) {
	c.Set("oper_log_title", "创建应付账款")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "accounts_payable")

	var ap entity.AccountsPayable
	if err := c.ShouldBindJSON(&ap); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := ctrl.apService.CreatePayable(&ap); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, ap)

	// 通知：创建应付账款
	userID := middleware.GetUserID(c)
	payOrderCode := ""
	if ap.OrderID != nil && ap.OrderType != "" {
		if ap.OrderType == "outsourcing" {
			if or, err := dao.NewOutsourcingRecordDAO().FindByID(*ap.OrderID); err == nil {
				payOrderCode = or.RecordCode
			}
		}
	}
	if payOrderCode == "" && ap.OrderID != nil {
		payOrderCode = fmt.Sprintf("#%d", *ap.OrderID)
	}
	// 付款类型映射
	payTypeName := ""
	switch ap.PaymentType {
	case "total":
		payTypeName = "全款"
	case "first":
		payTypeName = "首款"
	case "final":
		payTypeName = "尾款"
	case "installment":
		payTypeName = "分期款"
	default:
		payTypeName = ap.PaymentType
	}
	NotifyOperatorAndAdmins(userID, "应付账款",
		fmt.Sprintf("来自订单 %s 的供应商 %s 应付 ¥%.2f 已创建", payOrderCode, payTypeName, ap.Amount),
		"finance:payable_created", "low",
		"", ap.ID)
}

// Update 更新应付账款
func (ctrl *AccountsPayableController) Update(c *gin.Context) {
	c.Set("oper_log_title", "更新应付账款")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "accounts_payable")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var ap entity.AccountsPayable
	if err := c.ShouldBindJSON(&ap); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	ap.ID = id
	if err := ctrl.apService.UpdatePayable(&ap); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, ap)
}

// Delete 删除应付账款
func (ctrl *AccountsPayableController) Delete(c *gin.Context) {
	c.Set("oper_log_title", "删除应付账款")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "accounts_payable")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := ctrl.apService.DeletePayable(id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// Pay 付款
func (ctrl *AccountsPayableController) Pay(c *gin.Context) {
	c.Set("oper_log_title", "应付账款付款")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "accounts_payable")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req struct {
		Amount          float64 `json:"amount" binding:"required"`
		DiscountAmount  float64 `json:"discount_amount"`
		PaymentDate     string  `json:"payment_date"` // YYYY-MM-DD
		PaymentMethod   string  `json:"payment_method"`
		TransactionNo   string  `json:"transaction_no"`
		PaymentAccount  string  `json:"payment_account"`
		ReceiverAccount string  `json:"receiver_account"`
		Remark          string  `json:"remark"`
		ReceiptFileID   string  `json:"receipt_file_id"`
		InvoiceFileID   string  `json:"invoice_file_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	var paymentDate time.Time
	if req.PaymentDate != "" {
		paymentDate, err = time.Parse("2006-01-02", req.PaymentDate)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "日期格式错误，应为 YYYY-MM-DD")
			return
		}
	} else {
		paymentDate = time.Now()
	}

	if err := ctrl.apService.Pay(id, req.Amount, req.DiscountAmount, paymentDate, req.PaymentMethod, req.TransactionNo, req.PaymentAccount, req.ReceiverAccount, req.Remark, req.ReceiptFileID, req.InvoiceFileID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)

	// 通知：应付账款付款完成
	userID := middleware.GetUserID(c)
	method := req.PaymentMethod
	if method == "" {
		method = "未知方式"
	} else {
		if label, _ := ctrl.dictDataService.GetLabelByCodeAndValue("payment-method", method); label != "" {
			method = label
		}
	}
	paidAP, _ := dao.NewAccountsPayableDAO().FindByID(id)
	paidOrderCode, paidPayType := getPayableOrderInfo(paidAP)
	NotifyOperatorAndAdmins(userID, "应付账款",
		fmt.Sprintf("来自订单 %s 的应付款 %s 已付款 ¥%.2f ，付款方式：%s", paidOrderCode, paidPayType, req.Amount, method),
		"finance:payable_paid", "middle",
		"", id)
}

// GetPayments 获取应付账款的付款记录列表
func (ctrl *AccountsPayableController) GetPayments(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	payments, err := ctrl.apService.GetPaymentsByPayableID(id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "获取付款记录失败")
		return
	}

	response.Success(c, payments)
}

// Overdue 获取逾期应付账款列表
func (ctrl *AccountsPayableController) Overdue(c *gin.Context) {
	list, err := ctrl.apService.GetOverdueList()
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, list)
}

// ========== 通知辅助函数 ==========

func getPayableOrderInfo(ap *entity.AccountsPayable) (orderCode, payTypeName string) {
	if ap == nil {
		return "", ""
	}
	// 订单编号
	if ap.OrderID != nil && ap.OrderType != "" {
		if ap.OrderType == "outsourcing" {
			if or, err := dao.NewOutsourcingRecordDAO().FindByID(*ap.OrderID); err == nil {
				orderCode = or.RecordCode
			}
		}
	}
	if orderCode == "" && ap.OrderID != nil {
		orderCode = fmt.Sprintf("#%d", *ap.OrderID)
	}
	// 付款类型映射
	switch ap.PaymentType {
	case "total":
		payTypeName = "全款"
	case "first":
		payTypeName = "首款"
	case "final":
		payTypeName = "尾款"
	case "installment":
		payTypeName = "分期款"
	default:
		payTypeName = ap.PaymentType
	}
	return
}

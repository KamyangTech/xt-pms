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

type FinanceExpenseController struct {
	expenseService  *service.FinanceExpenseService
	dictDataService *service.SysDictDataService
}

func NewFinanceExpenseController() *FinanceExpenseController {
	return &FinanceExpenseController{
		expenseService:  service.NewFinanceExpenseService(),
		dictDataService: service.NewSysDictDataService(),
	}
}

// List 获取费用列表
func (ctrl *FinanceExpenseController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	expenseType := c.Query("expense_type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var relatedOrderID *int64
	if oid := c.Query("related_order_id"); oid != "" {
		id, err := strconv.ParseInt(oid, 10, 64)
		if err == nil {
			relatedOrderID = &id
		}
	}
	relatedOrderType := c.Query("related_order_type")

	list, total, err := ctrl.expenseService.GetExpenseList(page, pageSize, expenseType, relatedOrderID, relatedOrderType, startDate, endDate)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// Detail 获取费用详情
func (ctrl *FinanceExpenseController) Detail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	expense, err := ctrl.expenseService.GetExpenseByID(id)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	response.Success(c, expense)
}

// Create 创建费用记录
func (ctrl *FinanceExpenseController) Create(c *gin.Context) {
	c.Set("oper_log_title", "创建费用记录")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "finance_expense")

	var expense entity.FinanceExpense
	if err := c.ShouldBindJSON(&expense); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	if err := ctrl.expenseService.CreateExpense(&expense); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, expense)

	// 通知：创建费用记录
	userID := middleware.GetUserID(c)
	expenseType := expense.ExpenseType
	if expenseType == "" {
		expenseType = "未分类"
	}
	// 费用类型中文映射
	expenseTypeCN := expenseType
	switch expenseType {
	case "freight":
		expenseTypeCN = "运费"
	case "packaging":
		expenseTypeCN = "包装费"
	case "insurance":
		expenseTypeCN = "保险费"
	case "other":
		expenseTypeCN = "其他"
	}
	expOrderCode := ""
	if expense.RelatedOrderID != nil && expense.RelatedOrderType != "" {
		if expense.RelatedOrderType == "mold_order" {
			if mo, err := dao.NewMoldOrderDAO().FindByID(*expense.RelatedOrderID); err == nil {
				expOrderCode = mo.OrderCode
			}
		} else if expense.RelatedOrderType == "sale_order" {
			expOrderCode = fmt.Sprintf("#%d", *expense.RelatedOrderID)
		}
	}
	if expOrderCode == "" && expense.RelatedOrderID != nil {
		expOrderCode = fmt.Sprintf("#%d", *expense.RelatedOrderID)
	}
	NotifyOperatorAndAdmins(userID, "财务费用",
		fmt.Sprintf("来自订单 %s 的 %s 费用 ¥%.2f 已创建", expOrderCode, expenseTypeCN, expense.Amount),
		"finance:expense_created", "low",
		"", expense.ID)
}

// Update 更新费用
func (ctrl *FinanceExpenseController) Update(c *gin.Context) {
	c.Set("oper_log_title", "更新费用记录")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "finance_expense")

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var expense entity.FinanceExpense
	if err := c.ShouldBindJSON(&expense); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	expense.ID = id
	if err := ctrl.expenseService.UpdateExpense(&expense); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除费用
func (ctrl *FinanceExpenseController) Delete(c *gin.Context) {
	c.Set("oper_log_title", "删除费用记录")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "finance_expense")

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := ctrl.expenseService.DeleteExpense(id); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)
}

// Pay 支付费用（含详细记录）
func (ctrl *FinanceExpenseController) Pay(c *gin.Context) {
	c.Set("oper_log_title", "支付费用")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "finance_expense")

	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Amount          float64 `json:"amount" binding:"required"`
		PaymentDate     string  `json:"payment_date"` // YYYY-MM-DD
		PaymentMethod   string  `json:"payment_method"`
		TransactionNo   string  `json:"transaction_no"`
		PaymentAccount  string  `json:"payment_account"`
		ReceiverAccount string  `json:"receiver_account"`
		Remark          string  `json:"remark"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "请求参数错误")
		return
	}

	var paymentDate time.Time
	if req.PaymentDate != "" {
		var err error
		paymentDate, err = time.Parse("2006-01-02", req.PaymentDate)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "日期格式错误，应为 YYYY-MM-DD")
			return
		}
	} else {
		paymentDate = time.Now()
	}

	if err := ctrl.expenseService.PayExpense(id, req.Amount, paymentDate, req.PaymentMethod, req.TransactionNo, req.PaymentAccount, req.ReceiverAccount, req.Remark); err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}

	response.Success(c, nil)

	// 通知：费用支付完成
	userID := middleware.GetUserID(c)
	method := req.PaymentMethod
	if method == "" {
		method = "未知方式"
	} else {
		if label, _ := ctrl.dictDataService.GetLabelByCodeAndValue("payment-method", method); label != "" {
			method = label
		}
	}
	paidExpense, _ := dao.NewFinanceExpenseDAO().FindByID(id)
	paidExpOrderCode := ""
	paidExpTypeCN := ""
	paidExpBalance := 0.0
	if paidExpense != nil {
		// 订单编号
		if paidExpense.RelatedOrderID != nil && paidExpense.RelatedOrderType != "" {
			if paidExpense.RelatedOrderType == "mold_order" {
				if mo, err := dao.NewMoldOrderDAO().FindByID(*paidExpense.RelatedOrderID); err == nil {
					paidExpOrderCode = mo.OrderCode
				}
			} else {
				paidExpOrderCode = fmt.Sprintf("#%d", *paidExpense.RelatedOrderID)
			}
		}
		if paidExpOrderCode == "" && paidExpense.RelatedOrderID != nil {
			paidExpOrderCode = fmt.Sprintf("#%d", *paidExpense.RelatedOrderID)
		}
		// 费用类型中文
		switch paidExpense.ExpenseType {
		case "freight":
			paidExpTypeCN = "运费"
		case "packaging":
			paidExpTypeCN = "包装费"
		case "insurance":
			paidExpTypeCN = "保险费"
		case "other":
			paidExpTypeCN = "其他"
		default:
			paidExpTypeCN = paidExpense.ExpenseType
		}
		// 剩余未支付
		paidExpBalance = paidExpense.Amount - paidExpense.PaidAmount
	}
	NotifyOperatorAndAdmins(userID, "财务费用",
		fmt.Sprintf("来自订单 %s 的 %s 费用已支付 ¥%.2f ，支付方式为 %s ，剩余未支付 ¥%.2f", paidExpOrderCode, paidExpTypeCN, req.Amount, method, paidExpBalance),
		"finance:expense_paid", "middle",
		"", id)
}

// GetPayments 获取费用的支付记录列表
func (ctrl *FinanceExpenseController) GetPayments(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	payments, err := ctrl.expenseService.GetPaymentsByExpenseID(id)
	if err != nil {
		response.Error(c, response.ERROR, "获取支付记录失败")
		return
	}
	response.Success(c, payments)
}

// GetByDelivery 根据交付记录ID获取费用列表
func (ctrl *FinanceExpenseController) GetByDelivery(c *gin.Context) {
	deliveryID, _ := strconv.ParseInt(c.Param("delivery_id"), 10, 64)
	expenses, err := ctrl.expenseService.GetExpensesByDelivery(deliveryID)
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	response.Success(c, expenses)
}

// GetSummaryStats 获取费用汇总统计
func (ctrl *FinanceExpenseController) GetSummaryStats(c *gin.Context) {
	totalAmount, totalPaid, totalBalance, err := ctrl.expenseService.GetSummaryStats()
	if err != nil {
		response.Error(c, response.ERROR, err.Error())
		return
	}
	response.Success(c, gin.H{
		"total_amount":   totalAmount,
		"total_paid":     totalPaid,
		"total_balance":  totalBalance,
	})
}

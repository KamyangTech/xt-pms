package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type AccountsReceivableController struct {
	arService       *service.AccountsReceivableService
	dictDataService *service.SysDictDataService
	permService     *service.PermissionService
}

func NewAccountsReceivableController() *AccountsReceivableController {
	return &AccountsReceivableController{
		arService:       service.NewAccountsReceivableService(),
		dictDataService: service.NewSysDictDataService(),
		permService:     service.NewPermissionService(),
	}
}

// parseMultiQuery 解析多值查询参数，支持三种形式：
// 1) ?key=a&key=b  2) ?key=a,b  3) ?key[]=a&key[]=b
func parseMultiQuery(c *gin.Context, paramName string) []string {
	raw := append(c.QueryArray(paramName), c.QueryArray(paramName+"[]")...)
	if len(raw) == 0 {
		return nil
	}
	seen := make(map[string]struct{})
	var result []string
	for _, item := range raw {
		for _, part := range strings.Split(item, ",") {
			v := strings.TrimSpace(part)
			if v == "" {
				continue
			}
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// List 获取应收账款列表
func (ctrl *AccountsReceivableController) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var customerID *int64
	if cid := c.Query("customer_id"); cid != "" {
		id, err := strconv.ParseInt(cid, 10, 64)
		if err == nil {
			customerID = &id
		}
	}

	status := c.Query("status")
	sourceTypes := parseMultiQuery(c, "source_types")
	paymentTypes := parseMultiQuery(c, "payment_types")

	list, total, err := ctrl.arService.GetReceivableList(page, pageSize, customerID, status, sourceTypes, paymentTypes)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// Detail 获取应收账款详情
func (ctrl *AccountsReceivableController) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	ar, err := ctrl.arService.GetReceivableByID(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "应收账款不存在")
		return
	}

	response.Success(c, ar)
}

// Create 创建应收账款
func (ctrl *AccountsReceivableController) Create(c *gin.Context) {
	c.Set("oper_log_title", "创建应收账款")
	c.Set("oper_log_type", 1)
	c.Set("oper_log_business_type", "accounts_receivable")

	var ar entity.AccountsReceivable
	if err := c.ShouldBindJSON(&ar); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	if err := ctrl.arService.CreateReceivable(&ar); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, ar)

	// 通知：创建应收账款
	userID := middleware.GetUserID(c)
	arCust, _ := dao.NewCustomerDAO().FindByID(ar.CustomerID)
	arCustName := ""
	if arCust != nil {
		arCustName = arCust.CusName
		if arCustName == "" {
			arCustName = arCust.CusCode
		}
	} else {
		arCustName = fmt.Sprintf("#%d", ar.CustomerID)
	}
	arOrderCode := ""
	if ar.OrderID != nil && ar.OrderType != "" {
		switch ar.OrderType {
		case "mold_order":
			if mo, err := dao.NewMoldOrderDAO().FindByID(*ar.OrderID); err == nil {
				arOrderCode = mo.OrderCode
			}
		case "production_order":
			if po, err := dao.NewProductionOrderDAO().FindByID(*ar.OrderID); err == nil {
				arOrderCode = po.OrderCode
			}
		case "master_order":
			// master_order table
		}
	}
	if arOrderCode == "" && ar.OrderID != nil {
		arOrderCode = fmt.Sprintf("#%d", *ar.OrderID)
	}
	NotifyOperatorAndAdmins(userID, "应收账款",
		fmt.Sprintf("客户 %s 应收 ¥%.2f 已创建，来自订单 %s", arCustName, ar.Amount, arOrderCode),
		"finance:receivable_created", "low",
		"", ar.ID)
}

// Update 更新应收账款
func (ctrl *AccountsReceivableController) Update(c *gin.Context) {
	c.Set("oper_log_title", "更新应收账款")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "accounts_receivable")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var ar entity.AccountsReceivable
	if err := c.ShouldBindJSON(&ar); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	ar.ID = id
	if err := ctrl.arService.UpdateReceivable(&ar); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, ar)
}

// Delete 删除应收账款
func (ctrl *AccountsReceivableController) Delete(c *gin.Context) {
	c.Set("oper_log_title", "删除应收账款")
	c.Set("oper_log_type", 2)
	c.Set("oper_log_business_type", "accounts_receivable")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := ctrl.arService.DeleteReceivable(id); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// Receive 收款
func (ctrl *AccountsReceivableController) Receive(c *gin.Context) {
	c.Set("oper_log_title", "应收账款收款")
	c.Set("oper_log_type", 3)
	c.Set("oper_log_business_type", "accounts_receivable")

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	var req struct {
		Amount           float64 `json:"amount" binding:"required"`
		DiscountAmount   float64 `json:"discount_amount"`
		PaymentDate      string  `json:"payment_date"` // YYYY-MM-DD
		PaymentMethod    string  `json:"payment_method"`
		TransactionNo    string  `json:"transaction_no"`
		ReceivingAccount string  `json:"receiving_account"`
		Remark           string  `json:"remark"`
		ReceiptFileID    string  `json:"receipt_file_id"`
		InvoiceFileID    string  `json:"invoice_file_id"`
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

	if err := ctrl.arService.ReceivePayment(id, req.Amount, req.DiscountAmount, paymentDate, req.PaymentMethod, req.TransactionNo, req.ReceivingAccount, req.Remark, req.ReceiptFileID, req.InvoiceFileID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)

	// 通知：收款完成
	userID := middleware.GetUserID(c)
	method := req.PaymentMethod
	if method == "" {
		method = "未知方式"
	} else {
		if label, _ := ctrl.dictDataService.GetLabelByCodeAndValue("payment-method", method); label != "" {
			method = label
		}
	}
	recAR, _ := dao.NewAccountsReceivableDAO().FindByID(id)
	recOrderCode := ""
	recPayTypeName := ""
	if recAR != nil {
		// 订单编号
		if recAR.OrderID != nil && recAR.OrderType != "" {
			switch recAR.OrderType {
			case "mold_order":
				if mo, err := dao.NewMoldOrderDAO().FindByID(*recAR.OrderID); err == nil {
					recOrderCode = mo.OrderCode
				}
			case "production_order":
				if po, err := dao.NewProductionOrderDAO().FindByID(*recAR.OrderID); err == nil {
					recOrderCode = po.OrderCode
				}
			}
		}
		if recOrderCode == "" && recAR.OrderID != nil {
			recOrderCode = fmt.Sprintf("#%d", *recAR.OrderID)
		}
		// 收款类型映射
		switch recAR.PaymentType {
		case "deposit":
			recPayTypeName = "首款"
		case "interim":
			recPayTypeName = "中期款"
		case "final":
			recPayTypeName = "尾款"
		case "total":
			recPayTypeName = "全款"
		default:
			recPayTypeName = recAR.PaymentType
		}
	}
	NotifyOperatorAndAdmins(userID, "应收账款",
		fmt.Sprintf("来自订单 %s 的应收款 %s 已收款 ¥%.2f ，收款方式：%s", recOrderCode, recPayTypeName, req.Amount, method),
		"finance:receivable_received", "middle",
		"", id)
}

// GetPayments 获取应收账款的收款记录列表
func (ctrl *AccountsReceivableController) GetPayments(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的ID")
		return
	}

	payments, err := ctrl.arService.GetPaymentsByReceivableID(id)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "获取收款记录失败")
		return
	}

	response.Success(c, payments)
}

// UpdatePayment 更新收款记录
func (ctrl *AccountsReceivableController) UpdatePayment(c *gin.Context) {
	paymentID, err := strconv.ParseInt(c.Param("paymentId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的收款记录ID")
		return
	}

	var req struct {
		PaymentDate      string `json:"payment_date"`
		PaymentMethod    string `json:"payment_method"`
		TransactionNo    string `json:"transaction_no"`
		ReceivingAccount string `json:"receiving_account"`
		Remark           string `json:"remark"`
		ReceiptFileID    string `json:"receipt_file_id"`
		InvoiceFileID    string `json:"invoice_file_id"`
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
	}

	payment := &entity.ReceivablePayment{
		ID:               paymentID,
		PaymentDate:      paymentDate,
		PaymentMethod:    req.PaymentMethod,
		TransactionNo:    req.TransactionNo,
		ReceivingAccount: req.ReceivingAccount,
		Remark:           req.Remark,
		ReceiptFileID:    req.ReceiptFileID,
		InvoiceFileID:    req.InvoiceFileID,
	}

	if err := ctrl.arService.UpdateReceivablePayment(payment); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, nil)
}

// Overdue 获取逾期应收账款列表
func (ctrl *AccountsReceivableController) Overdue(c *gin.Context) {
	list, err := ctrl.arService.GetOverdueList()
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, list)
}

// GetByOrder 按订单ID和类型查询应收账款
//
// 该接口是订单详情页的附属数据接口，需要根据 orderType 走对应的订单权限
// 而不是独立的财务权限：
//   - mold_order       → mold:order:list
//   - master_order     → order-management:list
//   - production_order → production:order:list
//   - 其他             → finance:receivable:list（兜底）
func (ctrl *AccountsReceivableController) GetByOrder(c *gin.Context) {
	orderID, err := strconv.ParseInt(c.Param("orderId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "无效的订单ID")
		return
	}

	orderType := c.Param("orderType")
	if orderType == "" {
		response.Error(c, http.StatusBadRequest, "订单类型不能为空")
		return
	}

	// 根据订单类型动态校验权限（应收是订单的附属数据，应跟随订单权限）
	// master_order 可以来自综合订单详情或模具订单详情，因此允许 mold:order:list 作为替代权限
	requiredPerm := "finance:receivable:list"
	var alternativePerm string
	switch orderType {
	case "mold_order":
		requiredPerm = "mold:order:list"
	case "master_order":
		requiredPerm = "order-management:list"
		alternativePerm = "mold:order:list" // 模具订单详情页回退查询 master 应收时可用
	case "production_order":
		requiredPerm = "production:order:list"
	}

	if userID := middleware.GetCurrentUserID(c); userID != nil {
		hasPerm := ctrl.permService.HasPermission(*userID, requiredPerm)
		if !hasPerm && alternativePerm != "" {
			hasPerm = ctrl.permService.HasPermission(*userID, alternativePerm)
		}
		if !hasPerm {
			response.Error(c, response.FORBIDDEN, "无权限访问")
			return
		}
	}

	list, err := ctrl.arService.GetReceivableByOrder(orderID, orderType)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, list)
}

// AgingStatistics 获取账龄统计
func (ctrl *AccountsReceivableController) AgingStatistics(c *gin.Context) {
	stats, err := ctrl.arService.GetAgingStatistics()
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, stats)
}

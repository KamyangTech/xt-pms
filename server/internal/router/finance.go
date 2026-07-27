package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

func RegisterFinanceRoutes(r *gin.RouterGroup) {
	permService := service.NewPermissionService()
	finance := r.Group("/finance")
	{
		// 应收账款
		arCtrl := controller.NewAccountsReceivableController()
		ar := finance.Group("/receivables")
		{
			ar.GET("", middleware.RequirePerm("finance:receivable:list", permService.HasPermission), arCtrl.List)
			// 按订单查询应收是订单详情页的附属数据，权限在 controller 中按 orderType 动态判断
			ar.GET("/by-order/:orderId/:orderType", arCtrl.GetByOrder)
			ar.GET("/:id", middleware.RequirePerm("finance:receivable:list", permService.HasPermission), arCtrl.Detail)
			ar.POST("", middleware.RequirePerm("finance:receivable:list", permService.HasPermission), arCtrl.Create)
			ar.PUT("/:id", middleware.RequirePerm("finance:receivable:list", permService.HasPermission), arCtrl.Update)
			ar.DELETE("/:id", middleware.RequirePerm("finance:receivable:list", permService.HasPermission), arCtrl.Delete)
			ar.POST("/:id/receive", middleware.RequirePerm("finance:receivable:list", permService.HasPermission), arCtrl.Receive)
			ar.GET("/:id/payments", middleware.RequirePerm("finance:receivable:list", permService.HasPermission), arCtrl.GetPayments)
			ar.PUT("/payments/:paymentId", middleware.RequirePerm("finance:receivable:list", permService.HasPermission), arCtrl.UpdatePayment)
			ar.GET("/overdue", middleware.RequirePerm("finance:receivable:list", permService.HasPermission), arCtrl.Overdue)
			ar.GET("/aging-stats", middleware.RequirePerm("finance:receivable:list", permService.HasPermission), arCtrl.AgingStatistics)
		}

		// 应付账款
		apCtrl := controller.NewAccountsPayableController()
		ap := finance.Group("/payables")
		{
			ap.GET("", middleware.RequirePerm("finance:payable:list", permService.HasPermission), apCtrl.List)
			ap.GET("/:id", middleware.RequirePerm("finance:payable:list", permService.HasPermission), apCtrl.Detail)
			ap.POST("", middleware.RequirePerm("finance:payable:list", permService.HasPermission), apCtrl.Create)
			ap.PUT("/:id", middleware.RequirePerm("finance:payable:list", permService.HasPermission), apCtrl.Update)
			ap.DELETE("/:id", middleware.RequirePerm("finance:payable:list", permService.HasPermission), apCtrl.Delete)
			ap.POST("/:id/pay", middleware.RequirePerm("finance:payable:list", permService.HasPermission), apCtrl.Pay)
			ap.GET("/:id/payments", middleware.RequirePerm("finance:payable:list", permService.HasPermission), apCtrl.GetPayments)
			ap.GET("/overdue", middleware.RequirePerm("finance:payable:list", permService.HasPermission), apCtrl.Overdue)
		}

		// 费用管理
		expenseCtrl := controller.NewFinanceExpenseController()
		expense := finance.Group("/expenses")
		{
			expense.GET("", middleware.RequirePerm("finance:expense:list", permService.HasPermission), expenseCtrl.List)
			expense.GET("/stats", middleware.RequirePerm("finance:expense:list", permService.HasPermission), expenseCtrl.GetSummaryStats)
			expense.GET("/:id", middleware.RequirePerm("finance:expense:list", permService.HasPermission), expenseCtrl.Detail)
			expense.POST("", middleware.RequirePerm("finance:expense:list", permService.HasPermission), expenseCtrl.Create)
			expense.PUT("/:id", middleware.RequirePerm("finance:expense:list", permService.HasPermission), expenseCtrl.Update)
			expense.DELETE("/:id", middleware.RequirePerm("finance:expense:list", permService.HasPermission), expenseCtrl.Delete)
			expense.POST("/:id/pay", middleware.RequirePerm("finance:expense:list", permService.HasPermission), expenseCtrl.Pay)
			expense.GET("/:id/payments", middleware.RequirePerm("finance:expense:list", permService.HasPermission), expenseCtrl.GetPayments)
			expense.GET("/delivery/:delivery_id", middleware.RequirePerm("finance:expense:list", permService.HasPermission), expenseCtrl.GetByDelivery)
		}
	}
}

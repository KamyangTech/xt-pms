package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

func SetupMoldRoutes(api *gin.RouterGroup) {
	moldController := controller.NewMoldController()
	processTemplateController := controller.NewProcessTemplateController()
	workOrderProcessController := controller.NewWorkOrderProcessController()
	moldProductRelationController := controller.NewMoldProductRelationController()
	permService := service.NewPermissionService()

	mold := api.Group("/mold")
	{
		// 模具档案管理（读操作用 list，写操作同时接受 list 权限）
		mold.GET("", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.GetMolds)
		mold.GET("/statistics", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.GetMoldStatistics)
		mold.GET("/:id", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.GetMoldByID)
		mold.POST("", middleware.RequireAnyPerm([]string{"mold:mold:create", "mold:mold:list"}, permService.HasPermission), moldController.CreateMold)
		mold.PUT("", middleware.RequireAnyPerm([]string{"mold:mold:update", "mold:mold:list"}, permService.HasPermission), moldController.UpdateMold)
		mold.DELETE("/:id", middleware.RequireAnyPerm([]string{"mold:mold:delete", "mold:mold:list"}, permService.HasPermission), moldController.DeleteMold)

		// 模具订单管理
		mold.GET("/orders", middleware.RequirePerm("mold:order:list", permService.HasPermission), moldController.GetMoldOrders)
		mold.GET("/order/:id", middleware.RequirePerm("mold:order:list", permService.HasPermission), moldController.GetMoldOrderByID)
		mold.POST("/orders/fix-status", middleware.RequireAnyPerm([]string{"mold:order:update", "mold:order:list"}, permService.HasPermission), moldController.FixOrderStatus)
		mold.POST("/order", middleware.RequireAnyPerm([]string{"mold:order:create", "mold:order:list"}, permService.HasPermission), moldController.CreateMoldOrder)
		mold.PUT("/order", middleware.RequireAnyPerm([]string{"mold:order:update", "mold:order:list"}, permService.HasPermission), moldController.UpdateMoldOrder)
		mold.DELETE("/order/:id", middleware.RequireAnyPerm([]string{"mold:order:delete", "mold:order:list"}, permService.HasPermission), moldController.DeleteMoldOrder)

		// 模具工单管理
		mold.GET("/work-orders", middleware.RequirePerm("mold:work-order:list", permService.HasPermission), moldController.GetMoldWorkOrders)
		mold.GET("/work-order/:id", middleware.RequirePerm("mold:work-order:list", permService.HasPermission), moldController.GetMoldWorkOrderByID)
		mold.POST("/work-order", middleware.RequireAnyPerm([]string{"mold:work-order:create", "mold:work-order:list"}, permService.HasPermission), moldController.CreateMoldWorkOrder)
		mold.PUT("/work-order", middleware.RequireAnyPerm([]string{"mold:work-order:update", "mold:work-order:list"}, permService.HasPermission), moldController.UpdateMoldWorkOrder)
		mold.DELETE("/work-order/:id", middleware.RequireAnyPerm([]string{"mold:work-order:delete", "mold:work-order:list"}, permService.HasPermission), moldController.DeleteMoldWorkOrder)
		mold.POST("/work-order/cancel", middleware.RequireAnyPerm([]string{"mold:work-order:update", "mold:work-order:list"}, permService.HasPermission), moldController.CancelMoldWorkOrder)

		// 工单工序管理
		mold.GET("/work-order/:id/processes", middleware.RequirePerm("mold:work-order:list", permService.HasPermission), workOrderProcessController.GetWorkOrderProcesses)
		mold.POST("/work-order/process/toggle", middleware.RequirePerm("mold:work-order:list", permService.HasPermission), workOrderProcessController.ToggleProcess)

		// 模具试模管理
		mold.GET("/try-tests", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.GetMoldTryTests)
		mold.POST("/try-test", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.CreateMoldTryTest)
		mold.PUT("/try-test/:id", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.UpdateMoldTryTest)
		mold.DELETE("/try-test/:id", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.DeleteMoldTryTest)

		// 模具交付管理
		mold.GET("/deliveries", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.GetMoldDeliveries)
		mold.POST("/delivery", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.CreateMoldDelivery)

		// 模具库存管理
		mold.GET("/stocks", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.GetMoldStocks)
		mold.POST("/stock", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldController.CreateMoldStock)

		// 模具图片管理
		imageController := controller.NewMoldImageController()
		mold.GET("/:id/images", middleware.RequirePerm("mold:mold:list", permService.HasPermission), imageController.GetMoldImages)
		mold.POST("/:id/images/attach", middleware.RequirePerm("mold:mold:list", permService.HasPermission), imageController.AttachFiles)
		mold.DELETE("/images/:imageId", middleware.RequirePerm("mold:mold:list", permService.HasPermission), imageController.DeleteMoldImage)
		mold.PUT("/:id/images/:imageId/primary", middleware.RequirePerm("mold:mold:list", permService.HasPermission), imageController.SetPrimaryImage)

		// 模具产品关联管理
		relationGroup := api.Group("/mold-product-relation")
		{
			relationGroup.GET("/by-mold", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldProductRelationController.GetRelationsByMoldID)
			relationGroup.GET("/by-product", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldProductRelationController.GetRelationsByProductID)
			relationGroup.GET("/by-item", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldProductRelationController.GetRelationsByItemID)
			relationGroup.POST("", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldProductRelationController.CreateRelation)
			relationGroup.DELETE("/:id", middleware.RequirePerm("mold:mold:list", permService.HasPermission), moldProductRelationController.DeleteRelation)
		}

		// 模具订单款项管理
		paymentController := controller.NewMoldOrderPaymentController()
		orderGroup := api.Group("/mold-order")
		{
			orderGroup.GET("/:orderId/payments", middleware.RequirePerm("mold:order:list", permService.HasPermission), paymentController.GetOrderPayments)
	orderGroup.GET("/payments/by-work-order/:workOrderId", middleware.RequirePerm("mold:order:list", permService.HasPermission), paymentController.GetPaymentsByWorkOrderID)
		orderGroup.GET("/payments/:id", middleware.RequirePerm("mold:order:list", permService.HasPermission), paymentController.GetPaymentByID)
		orderGroup.POST("/payments", middleware.RequirePerm("mold:order:list", permService.HasPermission), paymentController.CreatePayment)
		orderGroup.PUT("/payments/:id", middleware.RequirePerm("mold:order:list", permService.HasPermission), paymentController.UpdatePayment)
		orderGroup.DELETE("/payments/:id", middleware.RequirePerm("mold:order:list", permService.HasPermission), paymentController.DeletePayment)
			orderGroup.GET("/:orderId/payment-stats", middleware.RequirePerm("mold:order:list", permService.HasPermission), paymentController.GetPaymentStats)
			orderGroup.POST("/payments/replace-by-work-order", middleware.RequirePerm("mold:order:list", permService.HasPermission), paymentController.ReplacePaymentsByWorkOrder)
		}
	}

	// 工序模板管理路由
	processTemplate := api.Group("/process-template")
	{
		processTemplate.GET("", middleware.RequirePerm("system:process-template:list", permService.HasPermission), processTemplateController.GetProcessTemplateList)
		processTemplate.GET("/active", middleware.RequirePerm("system:process-template:list", permService.HasPermission), processTemplateController.GetActiveProcessTemplates)
		processTemplate.POST("", middleware.RequirePerm("system:process-template:list", permService.HasPermission), processTemplateController.CreateProcessTemplate)
		processTemplate.PUT("", middleware.RequirePerm("system:process-template:list", permService.HasPermission), processTemplateController.UpdateProcessTemplate)
		processTemplate.DELETE("/:id", middleware.RequirePerm("system:process-template:list", permService.HasPermission), processTemplateController.DeleteProcessTemplate)
	}
}

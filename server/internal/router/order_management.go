package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

func SetupOrderManagementRoutes(r *gin.RouterGroup) {
	orderCtrl := controller.NewMasterOrderController()
	permService := service.NewPermissionService()

	orders := r.Group("/master-orders")
	{
		orders.POST("", middleware.RequirePerm("order-management:list", permService.HasPermission), orderCtrl.CreateCombinedOrder)
		orders.GET("", middleware.RequirePerm("order-management:list", permService.HasPermission), orderCtrl.List)
		orders.GET("/:id", middleware.RequirePerm("order-management:list", permService.HasPermission), orderCtrl.GetDetail)
		orders.PUT("/:id", middleware.RequirePerm("order-management:list", permService.HasPermission), orderCtrl.Edit)
		orders.POST("/:id/cancel", middleware.RequirePerm("order-management:list", permService.HasPermission), orderCtrl.Cancel)
	}
}

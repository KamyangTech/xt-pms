package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

func SetupDashboardRoutes(api *gin.RouterGroup) {
	permService := service.NewPermissionService()
	dashboard := api.Group("/dashboard")
	{
		dc := controller.NewDashboardController()

		dashboard.GET("/overview", middleware.RequirePerm("dashboard:base", permService.HasPermission), dc.GetOverview)
		dashboard.GET("/production-trend", middleware.RequirePerm("dashboard:base", permService.HasPermission), dc.GetProductionTrend)
		dashboard.GET("/mold-status", middleware.RequirePerm("dashboard:base", permService.HasPermission), dc.GetMoldStatus)
		dashboard.GET("/order-ranking", middleware.RequirePerm("dashboard:base", permService.HasPermission), dc.GetOrderRanking)
		dashboard.GET("/in-production-molds", middleware.RequirePerm("dashboard:base", permService.HasPermission), dc.GetInProductionMolds)
		dashboard.GET("/in-progress-mold-orders", middleware.RequirePerm("dashboard:base", permService.HasPermission), dc.GetInProgressMoldOrders)
		dashboard.GET("/in-progress-production-orders", middleware.RequirePerm("dashboard:base", permService.HasPermission), dc.GetInProgressProductionOrders)
		dashboard.GET("/finance-overview", middleware.RequirePerm("dashboard:detail", permService.HasPermission), dc.GetFinanceOverview)
		dashboard.GET("/detail-overview", middleware.RequirePerm("dashboard:detail", permService.HasPermission), dc.GetDetailOverview)
		dashboard.GET("/detail-production-trend", middleware.RequirePerm("dashboard:detail", permService.HasPermission), dc.GetDetailProductionTrend)
		dashboard.GET("/delivery-punctuality", middleware.RequirePerm("dashboard:detail", permService.HasPermission), dc.GetDeliveryPunctuality)
	}
}

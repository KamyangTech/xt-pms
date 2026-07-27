package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

func SetupProductionRoutes(api *gin.RouterGroup) {
	productionController := controller.NewProductionController()
	externalMoldArchiveController := controller.NewExternalMoldArchiveController()
	permService := service.NewPermissionService()

	production := api.Group("/production")
	{
		// 生产订单管理（写操作同时接受 list 权限）
		production.GET("/orders", middleware.RequirePerm("production:order:list", permService.HasPermission), productionController.GetOrders)
		production.GET("/orders/:id", middleware.RequirePerm("production:order:list", permService.HasPermission), productionController.GetOrderByID)
		production.GET("/mold/:mold_id/production-items", middleware.RequirePerm("production:order:list", permService.HasPermission), productionController.GetProductionItemsByMoldID)
		production.POST("/order", middleware.RequireAnyPerm([]string{"production:order:create", "production:order:list"}, permService.HasPermission), productionController.CreateOrder)
		production.PUT("/order", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.UpdateOrder)
		production.DELETE("/order/:id", middleware.RequireAnyPerm([]string{"production:order:delete", "production:order:list"}, permService.HasPermission), productionController.CancelOrder)
		production.POST("/order/:id/start", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.StartProduction)
		production.POST("/order/:id/complete", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.CompleteProduction)
		production.POST("/order/stock-in", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.StockInOrder)
		production.POST("/order/:id/shipment", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.ShipmentOrder)
		production.POST("/order/:id/finish", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.FinishOrder)
		production.POST("/order/:id/replenish", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.ReplenishShipment)
		production.POST("/order/item/:item_id/finish", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.FinishOrderItem)
		production.GET("/order/:id/actions", middleware.RequirePerm("production:order:list", permService.HasPermission), productionController.GetAvailableActions)

		// 生产报工
		production.POST("/report", middleware.RequireAnyPerm([]string{"production:work-report:submit", "production:work-report:list"}, permService.HasPermission), productionController.SubmitReport)
		production.GET("/reports", middleware.RequirePerm("production:work-report:list", permService.HasPermission), productionController.GetReports)

		// 外发管理
		production.GET("/outsourcing", middleware.RequirePerm("production:order:list", permService.HasPermission), productionController.GetOutsourcingRecords)
		production.GET("/outsourcing/list", middleware.RequirePerm("production:order:list", permService.HasPermission), productionController.GetOutsourcingRecordsByType)
		production.POST("/outsourcing", middleware.RequireAnyPerm([]string{"production:order:create", "production:order:list"}, permService.HasPermission), productionController.CreateOutsourcingRecord)
		production.POST("/outsourcing/:id/send", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.SendOutsourcing)
		production.POST("/outsourcing/:id/receive", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.ReceiveOutsourcing)
		production.POST("/outsourcing/:id/report", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.AddOutsourcingReport)
		production.POST("/outsourcing/:id/share-token", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.GenerateShareToken)
		production.POST("/outsourcing/:id/regenerate-token", middleware.RequireAnyPerm([]string{"production:order:update", "production:order:list"}, permService.HasPermission), productionController.RegenerateShareToken)
		production.GET("/outsourcing/:id/reports", middleware.RequirePerm("production:order:list", permService.HasPermission), productionController.GetOutsourcingReports)

		// 模具使用管理
		production.POST("/mold-usage", middleware.RequireAnyPerm([]string{"production:order:create", "production:order:list"}, permService.HasPermission), productionController.CreateMoldUsageRecord)
		production.GET("/mold/:mold_id/usage-history", middleware.RequirePerm("mold:usage-history:list", permService.HasPermission), productionController.GetMoldUsageHistory)
		production.GET("/mold/usage-history", middleware.RequirePerm("mold:usage-history:list", permService.HasPermission), productionController.GetMoldUsageHistory)
		production.GET("/mold/:mold_id/usage-stats", middleware.RequirePerm("mold:usage-history:list", permService.HasPermission), productionController.GetMoldUsageStats)

		// 统计分析
		production.GET("/statistics", middleware.RequirePerm("production:order:list", permService.HasPermission), productionController.GetStatistics)

		// 库存管理（写操作同时接受 list 权限）
		production.GET("/stock/product-list", middleware.RequirePerm("production:stock:list", permService.HasPermission), productionController.GetProductStockList)
		production.POST("/stock/delete", middleware.RequireAnyPerm([]string{"production:stock:out", "production:stock:list"}, permService.HasPermission), productionController.DeleteStock)
		production.GET("/stock/ext/:external_mold_code/records", middleware.RequirePerm("production:stock:list", permService.HasPermission), productionController.GetStockRecordsByExternalMoldCode)
		production.GET("/stock/:mold_id/records", middleware.RequirePerm("production:stock:list", permService.HasPermission), productionController.GetStockRecordsByProductID)
		production.GET("/stock-in/list", middleware.RequirePerm("production:stock:list", permService.HasPermission), productionController.GetStockInList)
		production.GET("/stock-in/:id", middleware.RequirePerm("production:stock:list", permService.HasPermission), productionController.GetStockInByID)
		production.PUT("/stock-in", middleware.RequireAnyPerm([]string{"production:stock:in", "production:stock:list"}, permService.HasPermission), productionController.UpdateStockIn)
		production.POST("/stock/direct", middleware.RequireAnyPerm([]string{"production:stock:in", "production:stock:list"}, permService.HasPermission), productionController.CreateStockDirectly)
		production.POST("/stock/check", middleware.RequireAnyPerm([]string{"production:stock:in", "production:stock:list"}, permService.HasPermission), productionController.CheckStock)
		production.GET("/stock/statistics", middleware.RequirePerm("production:stock:list", permService.HasPermission), productionController.GetStockStatistics)
		production.GET("/stock/trend", middleware.RequirePerm("production:stock:list", permService.HasPermission), productionController.GetStockTrend)

		// 出库管理
		production.POST("/stock-out", middleware.RequireAnyPerm([]string{"production:stock-out:create", "production:stock-out:list"}, permService.HasPermission), productionController.StockOutOrder)
		production.GET("/stock-out/list", middleware.RequirePerm("production:stock-out:list", permService.HasPermission), productionController.GetStockOutList)
		production.GET("/stock-out/:id", middleware.RequirePerm("production:stock-out:list", permService.HasPermission), productionController.GetStockOutByID)
		production.PUT("/stock-out/:id", middleware.RequireAnyPerm([]string{"production:stock-out:update", "production:stock-out:list"}, permService.HasPermission), productionController.UpdateStockOut)
		production.GET("/stock-out/mold/:mold_id/records", middleware.RequirePerm("production:stock-out:list", permService.HasPermission), productionController.GetStockOutRecordsByMoldID)
		production.GET("/stock-out/ext/:external_mold_code/records", middleware.RequirePerm("production:stock-out:list", permService.HasPermission), productionController.GetStockOutRecordsByExternalMoldCode)
		production.GET("/stock-out/recognize-express", middleware.RequirePerm("production:stock-out:list", permService.HasPermission), productionController.RecognizeExpressCompany)
		production.POST("/stock-out/list-for-print", middleware.RequirePerm("production:stock-out:list", permService.HasPermission), productionController.GetStockOutListForPrint)
		production.POST("/stock/current", middleware.RequirePerm("production:stock-out:list", permService.HasPermission), productionController.GetCurrentStock)

		// 外来模具档案管理
		production.GET("/external-mold-archive/query", middleware.RequirePerm("production:order:list", permService.HasPermission), externalMoldArchiveController.GetByMoldCode)
		production.GET("/external-mold-archive/list", middleware.RequirePerm("production:order:list", permService.HasPermission), externalMoldArchiveController.ListArchives)
		production.PUT("/external-mold-archive/:id", middleware.RequirePerm("production:order:list", permService.HasPermission), externalMoldArchiveController.UpdateArchive)
		production.DELETE("/external-mold-archive/:id", middleware.RequirePerm("production:order:list", permService.HasPermission), externalMoldArchiveController.DeleteArchive)
	}
}

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/monitor"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
)

func SetupRouter(r *gin.Engine) {
	// 健康检查端点（公开，供负载均衡器和监控系统使用）
	r.GET("/health", monitor.HealthHandler)

	// API 路由组
	api := r.Group("/api")

	{
		// 测试路由（公开）
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		// WebSocket 实时通知（公开路由，内部自行验证 token）
		systemController := controller.NewSystemController()
		api.GET("/ws", systemController.HandleWebSocket)

		// 系统管理路由（包含登录等公开路由）
		// 注意：SetupSystemRoutes 内部会区分公开路由和需要认证的路由
		SetupSystemRoutes(api)

		// 外发公开路由（加工商免登录报工）
		outsourcingPublicController := controller.NewOutsourcingPublicController()
		api.GET("/public/outsourcing/:token", outsourcingPublicController.GetByToken)
		api.POST("/public/outsourcing/:token/report", outsourcingPublicController.AddReport)
		api.POST("/public/outsourcing/:token/process/toggle", outsourcingPublicController.ToggleProcess)

		// 需要认证的路由组
		auth := api.Group("")
		auth.Use(middleware.JWT())
		auth.Use(middleware.OperationLog())
		{
			// 性能监控调试端点（需登录）
			auth.GET("/debug/metrics", monitor.MetricsHandler)
			auth.GET("/debug/pool", monitor.PoolHandler)
			auth.GET("/debug/health", monitor.FullHealthHandler)

			// 全局搜索（无需独立权限，登录即可使用）
			auth.GET("/search", controller.NewSearchController().Search)

			// 公共引用数据路由（仅需 JWT，用于跨模块下拉/选择器数据）
			SetupReferenceRoutes(auth)

			// 基础资料路由（包含字典管理）
			SetupBasicRoutes(auth)

			// 模具管理路由
			SetupMoldRoutes(auth)

			// 产品管理路由
			SetupProductRoutes(auth)

			// 生产管理路由
			SetupProductionRoutes(auth)

			// 数据看板路由
			SetupDashboardRoutes(auth)

			// 通用上传路由
			SetupUploadRoutes(auth)

			// 订单管理路由
			SetupOrderManagementRoutes(auth)

			// 财务管理路由
			RegisterFinanceRoutes(auth)
		}
	}
}

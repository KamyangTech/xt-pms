package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

func SetupSystemRoutes(api *gin.RouterGroup) {
	system := api.Group("/system")
	{
		userController := controller.NewUserController()
		systemController := controller.NewSystemController()
		menuController := controller.NewMenuController()
		permService := service.NewPermissionService()

		// 公开路由（不需要认证）
		system.GET("/captcha", middleware.CaptchaRateLimit(), userController.GetCaptcha)

		// 登录接口应用 IP 限流：每 IP 每分钟最多 10 次尝试
		system.POST("/login", middleware.LoginRateLimit(), userController.Login)

		// 字典版本查询（用于前端版本轮询，无需 JWT 认证）
		system.GET("/dict/version", systemController.GetDictVersion)

		// 系统信息配置（获取无需认证）
		system.GET("/system-info", systemController.GetSystemInfo)

		// ===== 需要认证的路由 =====
		auth := system.Group("")
		auth.Use(middleware.JWT(), middleware.OperationLog())
		{
			// 个人中心（无独立权限，只要登录即可）
			auth.GET("/profile", userController.GetProfile)
			auth.PUT("/profile", userController.UpdateProfile)
			auth.POST("/profile/password", userController.ChangePassword)
			auth.GET("/profile/statistics", userController.GetUserStatistics)

			// 用户管理
			auth.GET("/users", middleware.RequirePerm("system:user:list", permService.HasPermission), userController.GetUsers)
			auth.GET("/users/:id", middleware.RequirePerm("system:user:list", permService.HasPermission), userController.GetUserByID)
			auth.POST("/users", middleware.RequirePerm("system:user:create", permService.HasPermission), userController.CreateUser)
			auth.PUT("/users", middleware.RequirePerm("system:user:update", permService.HasPermission), userController.UpdateUser)
			auth.DELETE("/users/:id", middleware.RequirePerm("system:user:delete", permService.HasPermission), userController.DeleteUser)
			auth.PUT("/users/status", middleware.RequirePerm("system:user:update", permService.HasPermission), userController.BatchUpdateStatus)
			auth.PUT("/users/:id/reset-pwd", middleware.RequirePerm("system:user:update", permService.HasPermission), userController.ResetPassword)
			auth.PUT("/users/:id/unlock", middleware.RequirePerm("system:user:update", permService.HasPermission), userController.UnlockUser)

			// 角色管理
			auth.GET("/roles", middleware.RequirePerm("system:role:list", permService.HasPermission), systemController.GetRoles)
			auth.GET("/roles/:id", middleware.RequirePerm("system:role:list", permService.HasPermission), systemController.GetRoleByID)
			auth.POST("/roles", middleware.RequirePerm("system:role:create", permService.HasPermission), systemController.CreateRole)
			auth.PUT("/roles", middleware.RequirePerm("system:role:update", permService.HasPermission), systemController.UpdateRole)
			auth.DELETE("/roles/:id", middleware.RequirePerm("system:role:delete", permService.HasPermission), systemController.DeleteRole)

			// 部门管理
			auth.GET("/depts", middleware.RequirePerm("system:dept:list", permService.HasPermission), systemController.GetDepts)
			auth.GET("/depts/:id", middleware.RequirePerm("system:dept:list", permService.HasPermission), systemController.GetDeptByID)
			auth.POST("/depts", middleware.RequirePerm("system:dept:list", permService.HasPermission), systemController.CreateDept)
			auth.PUT("/depts", middleware.RequirePerm("system:dept:list", permService.HasPermission), systemController.UpdateDept)
			auth.DELETE("/depts/:id", middleware.RequirePerm("system:dept:list", permService.HasPermission), systemController.DeleteDept)
			auth.PUT("/depts/:id/status", middleware.RequirePerm("system:dept:list", permService.HasPermission), systemController.UpdateDeptStatus)

			// 公司信息管理（读取无需额外权限，登录即可；修改需权限）
			companyController := controller.NewCompanyController()
			auth.GET("/company", companyController.GetCompany)
			auth.PUT("/company", middleware.RequirePerm("system:company:list", permService.HasPermission), companyController.UpdateCompany)
			auth.POST("/company/logo", middleware.RequirePerm("system:company:list", permService.HasPermission), systemController.UploadCompanyLogo)
			auth.DELETE("/company/logo", middleware.RequirePerm("system:company:list", permService.HasPermission), systemController.DeleteCompanyLogo)

			// 菜单管理（读操作需登录即可，写操作需对应权限）
			auth.GET("/menus", menuController.GetMenuTree)  // 登录用户获取自身菜单树，无需额外权限
			auth.GET("/menus/:id", middleware.RequirePerm("system:menu:list", permService.HasPermission), menuController.GetMenuByID)
			auth.POST("/menus", middleware.RequirePerm("system:menu:create", permService.HasPermission), menuController.CreateMenu)
			auth.PUT("/menus", middleware.RequirePerm("system:menu:update", permService.HasPermission), menuController.UpdateMenu)
			auth.DELETE("/menus/:id", middleware.RequirePerm("system:menu:delete", permService.HasPermission), menuController.DeleteMenu)

			// 角色分配菜单
			auth.GET("/role-menus/tree", middleware.RequirePerm("system:role:assign-perm", permService.HasPermission), menuController.GetMenuTree)
			auth.GET("/roles/:id/menus", middleware.RequirePerm("system:role:assign-perm", permService.HasPermission), systemController.GetRoleMenus)
			auth.POST("/role-menus", middleware.RequirePerm("system:role:assign-perm", permService.HasPermission), systemController.AssignRoleMenus)

			// 用户权限（登录用户必调，不设独立权限）
			auth.GET("/user/permissions", systemController.GetUserPermissions)

			// 字典管理
			auth.GET("/dict-types", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.GetDictTypes)
			auth.GET("/dict-types/:id", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.GetDictTypeByID)
			auth.POST("/dict-types", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.CreateDictType)
			auth.PUT("/dict-types", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.UpdateDictType)
			auth.DELETE("/dict-types/:id", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.DeleteDictType)

			// 字典数据管理（系统内置状态：只读+可编辑颜色和备注）
			auth.GET("/dict-data", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.GetDictDataByType)
			auth.GET("/dict-data-by-id", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.GetDictDataByTypeID)
			auth.GET("/dict-data-page", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.GetDictDataPage)
			auth.PUT("/dict-data", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.UpdateDictData)
			auth.POST("/dict-data/clear-cache", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.ClearDictCache)
			auth.POST("/dict-data/restore-seed", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.RestoreDictSeed)
			auth.POST("/dict-data/restore-production-seed", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.RestoreProductionDictSeed)

			// 业务配置数据管理（支持全字段增删改查）
			auth.POST("/business-dict-data", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.CreateDictData)
			auth.PUT("/business-dict-data", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.UpdateBusinessDictData)
			auth.DELETE("/business-dict-data/:id", middleware.RequirePerm("system:dict:list", permService.HasPermission), systemController.DeleteDictData)

			// 状态流转配置管理（只读）
			statusFlowController := controller.NewStatusFlowController()
			auth.GET("/status-flow-configs", middleware.RequirePerm("system:process-template:list", permService.HasPermission), statusFlowController.GetStatusFlowConfigs)
			auth.GET("/status-flow-configs/:id", middleware.RequirePerm("system:process-template:list", permService.HasPermission), statusFlowController.GetStatusFlowConfigByID)

			auth.GET("/print-notice-config", middleware.RequirePerm("system:print-config:list", permService.HasPermission), systemController.GetPrintNoticeConfig)
			auth.POST("/print-notice-config", middleware.RequirePerm("system:print-config:list", permService.HasPermission), systemController.SavePrintNoticeConfig)
			auth.POST("/print-notice-config/reset", middleware.RequirePerm("system:print-config:list", permService.HasPermission), systemController.ResetPrintNoticeConfig)

			// 系统信息配置
			auth.PUT("/system-info", middleware.RequirePerm("system:settings:list", permService.HasPermission), systemController.UpdateSystemInfo)
			auth.POST("/favicon", middleware.RequirePerm("system:settings:list", permService.HasPermission), systemController.UploadFavicon)
			auth.DELETE("/favicon", middleware.RequirePerm("system:settings:list", permService.HasPermission), systemController.DeleteFavicon)

			// 系统设置 - 敏感操作
			auth.POST("/cleanup-orphan-files", middleware.RequirePerm("system:settings:list", permService.HasPermission), systemController.CleanupOrphanFiles)
			auth.GET("/cleanup-orphan-files/count", middleware.RequirePerm("system:settings:list", permService.HasPermission), systemController.GetOrphanFileCount)
			auth.GET("/cleanup-config", middleware.RequirePerm("system:settings:list", permService.HasPermission), systemController.GetCleanupConfig)
			auth.PUT("/cleanup-config", middleware.RequirePerm("system:settings:list", permService.HasPermission), systemController.UpdateCleanupConfig)

			// 登录与安全配置
			auth.GET("/security-config", middleware.RequirePerm("system:settings:list", permService.HasPermission), systemController.GetSecurityConfig)
			auth.PUT("/security-config", middleware.RequirePerm("system:settings:list", permService.HasPermission), systemController.UpdateSecurityConfig)

			// 通知管理（登录即可，不设独立权限）
			notificationController := controller.NewNotificationController()
			auth.GET("/notifications", notificationController.GetPage)
			auth.GET("/notifications/types", notificationController.GetTypes)
			auth.GET("/notifications/unread-count", notificationController.GetUnreadCount)
			auth.PUT("/notifications/:id/read", notificationController.MarkAsRead)
			auth.PUT("/notifications/read-all", notificationController.MarkAllAsRead)
			auth.DELETE("/notifications/:id", notificationController.Delete)
			auth.POST("/notifications/test", notificationController.TestSend) // 测试用

			// 通知配置管理（登录即可，权限控制可后续通过系统权限配置添加）
			auth.GET("/notification-configs", notificationController.GetConfigs)
			auth.GET("/notification-configs/types", notificationController.GetAllTypes)
			auth.POST("/notification-configs", notificationController.CreateConfig)
			auth.PUT("/notification-configs", notificationController.UpdateConfig)
			auth.DELETE("/notification-configs/:id", notificationController.DeleteConfig)

			// 操作日志管理
			operLogController := controller.NewOperLogController()
			auth.GET("/oper-logs", middleware.RequirePerm("system:oper-log:list", permService.HasPermission), operLogController.GetList)
			auth.GET("/oper-logs/:id", middleware.RequirePerm("system:oper-log:list", permService.HasPermission), operLogController.GetDetail)
			auth.GET("/oper-log/export", middleware.RequirePerm("system:oper-log:list", permService.HasPermission), operLogController.Export)
			auth.GET("/oper-log/oper-types", middleware.RequirePerm("system:oper-log:list", permService.HasPermission), operLogController.GetOperTypeOptions)

			// 收款账户管理
			auth.GET("/receiving-accounts", middleware.RequirePerm("system:receiving-account:list", permService.HasPermission), systemController.GetReceivingAccounts)
			auth.GET("/receiving-accounts/:id", middleware.RequirePerm("system:receiving-account:list", permService.HasPermission), systemController.GetReceivingAccountByID)
			auth.POST("/receiving-accounts", middleware.RequirePerm("system:receiving-account:list", permService.HasPermission), systemController.CreateReceivingAccount)
			auth.PUT("/receiving-accounts", middleware.RequirePerm("system:receiving-account:list", permService.HasPermission), systemController.UpdateReceivingAccount)
			auth.DELETE("/receiving-accounts/:id", middleware.RequirePerm("system:receiving-account:list", permService.HasPermission), systemController.DeleteReceivingAccount)
			auth.POST("/receiving-accounts/batch-delete", middleware.RequirePerm("system:receiving-account:list", permService.HasPermission), systemController.BatchDeleteReceivingAccounts)
			auth.PUT("/receiving-accounts/:id/status", middleware.RequirePerm("system:receiving-account:list", permService.HasPermission), systemController.UpdateReceivingAccountStatus)

			// 系统在线更新
			updateController := controller.NewUpdateController()
			auth.GET("/update/check", middleware.RequirePerm("system:settings:list", permService.HasPermission), updateController.CheckUpdate)
			auth.POST("/update/download", middleware.RequirePerm("system:settings:list", permService.HasPermission), updateController.DownloadUpdate)
			auth.POST("/update/apply", middleware.RequirePerm("system:settings:list", permService.HasPermission), updateController.ApplyUpdate)
			auth.POST("/update/restart", middleware.RequirePerm("system:settings:list", permService.HasPermission), updateController.Restart)
			auth.GET("/update/install-info", middleware.RequirePerm("system:settings:list", permService.HasPermission), updateController.GetInstallInfo)
		}
	}
}

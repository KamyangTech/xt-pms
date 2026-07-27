package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

func SetupBasicRoutes(api *gin.RouterGroup) {
	basicController := controller.NewBasicController()
	permService := service.NewPermissionService()

	basic := api.Group("/basic")
	{
		// 客户管理
		customer := basic.Group("/customer")
		{
			customer.GET("", middleware.RequirePerm("system:customer:list", permService.HasPermission), basicController.GetCustomers)
			customer.GET("/:id", middleware.RequirePerm("system:customer:list", permService.HasPermission), basicController.GetCustomerByID)
			customer.POST("", middleware.RequirePerm("system:customer:list", permService.HasPermission), basicController.CreateCustomer)
			customer.PUT("", middleware.RequirePerm("system:customer:list", permService.HasPermission), basicController.UpdateCustomer)
			customer.DELETE("/:id", middleware.RequirePerm("system:customer:list", permService.HasPermission), basicController.DeleteCustomer)
			customer.POST("/batch-delete", middleware.RequirePerm("system:customer:list", permService.HasPermission), basicController.BatchDeleteCustomers)
		}

		// 仓库管理
		warehouse := basic.Group("/warehouse")
		{
			warehouse.GET("", middleware.RequirePerm("system:warehouse:list", permService.HasPermission), basicController.GetWarehouses)
			warehouse.GET("/:id", middleware.RequirePerm("system:warehouse:list", permService.HasPermission), basicController.GetWarehouseByID)
			warehouse.POST("", middleware.RequirePerm("system:warehouse:list", permService.HasPermission), basicController.CreateWarehouse)
			warehouse.PUT("", middleware.RequirePerm("system:warehouse:list", permService.HasPermission), basicController.UpdateWarehouse)
			warehouse.DELETE("/:id", middleware.RequirePerm("system:warehouse:list", permService.HasPermission), basicController.DeleteWarehouse)
			warehouse.POST("/batch-delete", middleware.RequirePerm("system:warehouse:list", permService.HasPermission), basicController.BatchDeleteWarehouses)
		}

		// 工序管理
		process := basic.Group("/process")
		{
			process.GET("", middleware.RequirePerm("system:process-template:list", permService.HasPermission), basicController.GetProcesses)
			process.GET("/:id", middleware.RequirePerm("system:process-template:list", permService.HasPermission), basicController.GetProcessByID)
			process.POST("", middleware.RequirePerm("system:process-template:list", permService.HasPermission), basicController.CreateProcess)
			process.PUT("", middleware.RequirePerm("system:process-template:list", permService.HasPermission), basicController.UpdateProcess)
			process.DELETE("/:id", middleware.RequirePerm("system:process-template:list", permService.HasPermission), basicController.DeleteProcess)
		}

		// 供应商管理
		supplier := basic.Group("/supplier")
		{
			supplier.GET("", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.GetSuppliers)
			supplier.GET("/:id", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.GetSupplierByID)
			supplier.POST("", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.CreateSupplier)
			supplier.PUT("", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.UpdateSupplier)
			supplier.DELETE("/:id", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.DeleteSupplier)
			supplier.POST("/batch-delete", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.BatchDeleteSuppliers)
		}

		// 加工商管理
		processor := basic.Group("/processor")
		{
			processor.GET("", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.GetProcessors)
			processor.GET("/:id", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.GetProcessorByID)
			processor.POST("", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.CreateProcessor)
			processor.PUT("", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.UpdateProcessor)
			processor.DELETE("/:id", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.DeleteProcessor)
			processor.POST("/batch-delete", middleware.RequirePerm("system:processor-supplier:list", permService.HasPermission), basicController.BatchDeleteProcessors)
		}
	}
}

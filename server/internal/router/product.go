package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

func SetupProductRoutes(api *gin.RouterGroup) {
	productController := controller.NewProductController()
	permService := service.NewPermissionService()

	product := api.Group("/product")
	{
		// 产品档案管理
		product.GET("", middleware.RequirePerm("product:list", permService.HasPermission), productController.GetProducts)
		product.POST("/sync", middleware.RequirePerm("product:list", permService.HasPermission), productController.SyncProducts)
		product.GET("/:id", middleware.RequirePerm("product:list", permService.HasPermission), productController.GetProductByID)
		product.POST("", middleware.RequirePerm("product:list", permService.HasPermission), productController.CreateProduct)
		product.PUT("", middleware.RequirePerm("product:list", permService.HasPermission), productController.UpdateProduct)
		product.DELETE("/:id", middleware.RequirePerm("product:list", permService.HasPermission), productController.DeleteProduct)
		product.GET("/generate-code", middleware.RequirePerm("product:list", permService.HasPermission), productController.GenerateProductCode)

		// 产品 BOM 管理
		product.GET("/boms", middleware.RequirePerm("product:list", permService.HasPermission), productController.GetProductBOMs)
		product.POST("/bom", middleware.RequirePerm("product:list", permService.HasPermission), productController.CreateProductBOM)
		product.PUT("/bom", middleware.RequirePerm("product:list", permService.HasPermission), productController.UpdateProductBOM)
		product.DELETE("/bom/:id", middleware.RequirePerm("product:list", permService.HasPermission), productController.DeleteProductBOM)

		// 库存管理（复用 production:stock:list 权限）
		product.GET("/stocks", middleware.RequirePerm("production:stock:list", permService.HasPermission), productController.GetStocks)
		product.GET("/stocks/:id", middleware.RequirePerm("production:stock:list", permService.HasPermission), productController.GetStockByID)
		product.POST("/stock", middleware.RequirePerm("production:stock:list", permService.HasPermission), productController.CreateStock)
		product.PUT("/stock", middleware.RequirePerm("production:stock:list", permService.HasPermission), productController.UpdateStock)
		product.POST("/stock/in", middleware.RequirePerm("production:stock:list", permService.HasPermission), productController.InStock)
		product.POST("/stock/out", middleware.RequirePerm("production:stock:list", permService.HasPermission), productController.OutStock)
		product.GET("/stock/safe-check", middleware.RequirePerm("production:stock:list", permService.HasPermission), productController.CheckSafeStock)
	}
}

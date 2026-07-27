package router

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/controller"
)

// SetupReferenceRoutes 注册公共引用数据路由
// 这些路由仅需 JWT 认证，无需额外菜单权限
// 专门给各模块页面提供下拉/选择器数据，避免跨模块的 403 权限问题
func SetupReferenceRoutes(api *gin.RouterGroup) {
	basicController := controller.NewBasicController()
	systemController := controller.NewSystemController()
	processTemplateController := controller.NewProcessTemplateController()
	statusFlowController := controller.NewStatusFlowController()
	moldController := controller.NewMoldController()
	productController := controller.NewProductController()
	companyController := controller.NewCompanyController()

	ref := api.Group("/reference")
	{
		// 客户下拉
		ref.GET("/customers", basicController.GetCustomers)
		ref.GET("/customers/:id", basicController.GetCustomerByID)

		// 仓库下拉
		ref.GET("/warehouses", basicController.GetWarehouses)

		// 工序下拉
		ref.GET("/processes", basicController.GetProcesses)

		// 供应商下拉
		ref.GET("/suppliers", basicController.GetSuppliers)

		// 加工商下拉
		ref.GET("/processors", basicController.GetProcessors)

		// 工序模板下拉
		ref.GET("/process-templates", processTemplateController.GetProcessTemplateList)
		ref.GET("/process-templates/active", processTemplateController.GetActiveProcessTemplates)

		// 字典数据下拉
		ref.GET("/dict-data", systemController.GetDictDataByType)
		ref.GET("/dict-data-by-id", systemController.GetDictDataByTypeID)

		// 字典类型下拉
		ref.GET("/dict-types", systemController.GetDictTypes)

		// 状态流转配置下拉
		ref.GET("/status-flow-configs", statusFlowController.GetStatusFlowConfigs)
		ref.GET("/status-flow-configs/:id", statusFlowController.GetStatusFlowConfigByID)

		// 公司信息（打印等场景需要，所有登录用户可读取）
		ref.GET("/company", companyController.GetCompany)

		ref.GET("/print-notice-config", systemController.GetPrintNoticeConfig)

		// 收款账户下拉
		ref.GET("/receiving-accounts", systemController.GetReceivingAccounts)
		ref.GET("/receiving-accounts/:id", systemController.GetReceivingAccountByID)

		// 模具档案下拉
		ref.GET("/molds", moldController.GetMolds)
		ref.GET("/molds/:id", moldController.GetMoldByID)

		// 产品下拉
		ref.GET("/products", productController.GetProducts)
		ref.GET("/products/:id", productController.GetProductByID)

		// 模具产品关联下拉（用于发货管理等需要双向联动的场景）
		moldProductRelationController := controller.NewMoldProductRelationController()
		ref.GET("/mold-product-relations/by-mold", moldProductRelationController.GetRelationsByMoldID)
		ref.GET("/mold-product-relations/by-product", moldProductRelationController.GetRelationsByProductID)
	}
}

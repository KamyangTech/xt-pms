package controller

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type ProductController struct {
	productService    *service.ProductService
	productBOMService *service.ProductBOMService
	stockService      *service.StockService
}

func NewProductController() *ProductController {
	return &ProductController{
		productService:    service.NewProductService(),
		productBOMService: service.NewProductBOMService(),
		stockService:      service.NewStockService(),
	}
}

// ==================== 产品档案管理 ====================

// GetProducts 获取产品列表
func (c *ProductController) GetProducts(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	products, total, err := c.productService.GetProductsWithStockByPage(page, pageSize)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取产品列表失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":     products,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

// SyncProducts 从生产记录同步产品
func (c *ProductController) SyncProducts(ctx *gin.Context) {
	err := c.productService.SyncProductsFromProductionRecords()
	if err != nil {
		response.Error(ctx, response.ERROR, "同步失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{
		"message": "同步成功",
	})
}

// GetProductByID 获取产品详情
func (c *ProductController) GetProductByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的产品 ID")
		return
	}

	product, err := c.productService.GetProductByID(id)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "产品不存在")
		return
	}

	response.Success(ctx, product)
}

// CreateProduct 创建产品
func (c *ProductController) CreateProduct(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建产品")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "product")

	var productDTO dto.ProductDTO
	if err := ctx.ShouldBindJSON(&productDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误："+err.Error())
		return
	}

	product, err := c.productService.CreateProduct(productDTO)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, product)

	// 通知：创建产品
	userID := middleware.GetUserID(ctx)
	productName := product.ProName
	if productName == "" {
		productName = product.ProCode
	}
	NotifyOperatorAndAdmins(userID, "产品管理",
		fmt.Sprintf("产品 %s 已创建", productName),
		"product:created", "low",
		"", product.ID)
}

// GenerateProductCode 生成产品编码
func (c *ProductController) GenerateProductCode(ctx *gin.Context) {
	code, err := c.productService.GenerateProductCode()
	if err != nil {
		response.Error(ctx, response.ERROR, "生成编码失败")
		return
	}

	response.Success(ctx, gin.H{
		"product_code": code,
	})
}

// UpdateProduct 更新产品
func (c *ProductController) UpdateProduct(ctx *gin.Context) {
	ctx.Set("oper_log_title", "更新产品")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "product")

	var productDTO dto.ProductUpdateDTO
	if err := ctx.ShouldBindJSON(&productDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	err := c.productService.UpdateProduct(productDTO)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)

	// 通知：更新产品
	userID := middleware.GetUserID(ctx)
	prod, _ := dao.NewProductDAO().FindByID(productDTO.ID)
	prodName := ""
	if prod != nil {
		prodName = prod.ProName
		if prodName == "" {
			prodName = prod.ProCode
		}
	} else {
		prodName = fmt.Sprintf("#%d", productDTO.ID)
	}
	NotifyOperatorAndAdmins(userID, "产品管理",
		fmt.Sprintf("产品 %s 信息已更新", prodName),
		"product:updated", "low",
		"", productDTO.ID)
}

// DeleteProduct 删除产品
func (c *ProductController) DeleteProduct(ctx *gin.Context) {
	ctx.Set("oper_log_title", "删除产品")
	ctx.Set("oper_log_type", 2)
	ctx.Set("oper_log_business_type", "product")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的产品 ID")
		return
	}

	deleteProd, _ := dao.NewProductDAO().FindByID(id)
	deleteProdName := ""
	if deleteProd != nil {
		deleteProdName = deleteProd.ProName
		if deleteProdName == "" {
			deleteProdName = deleteProd.ProCode
		}
	} else {
		deleteProdName = fmt.Sprintf("#%d", id)
	}

	err = c.productService.DeleteProduct(id)
	if err != nil {
		response.Error(ctx, response.ERROR, "删除失败")
		return
	}

	response.Success(ctx, nil)

	// 通知：删除产品
	userID := middleware.GetUserID(ctx)
	NotifyOperatorAndAdmins(userID, "产品管理",
		fmt.Sprintf("产品 %s 已删除", deleteProdName),
		"product:deleted", "low",
		"", id)
}

// ==================== 产品 BOM 管理 ====================

// GetProductBOMs 获取产品的 BOM 清单
func (c *ProductController) GetProductBOMs(ctx *gin.Context) {
	productIDStr := ctx.Query("product_id")
	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的产品 ID")
		return
	}

	boms, err := c.productBOMService.GetProductBOMByProductID(productID)
	if err != nil {
		response.Error(ctx, response.ERROR, "获取 BOM 清单失败")
		return
	}

	totalCost := 0.0
	for _, bom := range boms {
		totalCost += bom.TotalPrice
	}

	response.Success(ctx, gin.H{
		"list":      boms,
		"total":     len(boms),
		"totalCost": totalCost,
	})
}

// CreateProductBOM 创建产品 BOM
func (c *ProductController) CreateProductBOM(ctx *gin.Context) {
	ctx.Set("oper_log_title", "创建产品BOM")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "product_bom")

	var bom entity.ProductBOM
	if err := ctx.ShouldBindJSON(&bom); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	createdBOM, err := c.productBOMService.CreateProductBOM(&bom)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, createdBOM)

	// 通知：创建产品BOM
	userID := middleware.GetUserID(ctx)
	bomProd, _ := dao.NewProductDAO().FindByID(bom.ProductID)
	bomProdName := ""
	if bomProd != nil {
		bomProdName = bomProd.ProName
		if bomProdName == "" {
			bomProdName = bomProd.ProCode
		}
	} else {
		bomProdName = fmt.Sprintf("#%d", bom.ProductID)
	}
	NotifyOperatorAndAdmins(userID, "产品管理",
		fmt.Sprintf("产品 %s BOM 已创建", bomProdName),
		"product:bom_created", "low",
		"", createdBOM.ID)
}

// UpdateProductBOM 更新产品 BOM
func (c *ProductController) UpdateProductBOM(ctx *gin.Context) {
	ctx.Set("oper_log_title", "更新产品BOM")
	ctx.Set("oper_log_type", 3)
	ctx.Set("oper_log_business_type", "product_bom")

	var bom entity.ProductBOM
	if err := ctx.ShouldBindJSON(&bom); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	err := c.productBOMService.UpdateProductBOM(&bom)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// DeleteProductBOM 删除产品 BOM
func (c *ProductController) DeleteProductBOM(ctx *gin.Context) {
	ctx.Set("oper_log_title", "删除产品BOM")
	ctx.Set("oper_log_type", 2)
	ctx.Set("oper_log_business_type", "product_bom")

	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的 BOM ID")
		return
	}

	err = c.productBOMService.DeleteProductBOM(id)
	if err != nil {
		response.Error(ctx, response.ERROR, "删除失败")
		return
	}

	response.Success(ctx, nil)
}

// ==================== 库存管理 ====================

// GetStocks 获取库存列表
func (c *ProductController) GetStocks(ctx *gin.Context) {
	warehouseIDStr := ctx.Query("warehouse_id")

	var stocks []*entity.Stock
	var err error

	if warehouseIDStr != "" {
		warehouseID, err := strconv.ParseInt(warehouseIDStr, 10, 64)
		if err != nil {
			response.Error(ctx, response.ERROR, "无效的仓库 ID")
			return
		}
		stocks, err = c.stockService.GetStocksByWarehouse(warehouseID)
	} else {
		allStocks, _ := c.stockService.GetStocksByWarehouse(0)
		stocks = allStocks
	}

	if err != nil {
		response.Error(ctx, response.ERROR, "获取库存列表失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":  stocks,
		"total": len(stocks),
	})
}

// GetStockByID 获取库存详情
func (c *ProductController) GetStockByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.ERROR, "无效的库存 ID")
		return
	}

	stock, err := c.stockService.GetStockByID(id)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "库存记录不存在")
		return
	}

	response.Success(ctx, stock)
}

// CreateStock 创建库存记录
func (c *ProductController) CreateStock(ctx *gin.Context) {
	var stock entity.Stock
	if err := ctx.ShouldBindJSON(&stock); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	createdStock, err := c.stockService.CreateStock(&stock)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, createdStock)
}

// UpdateStock 更新库存
func (c *ProductController) UpdateStock(ctx *gin.Context) {
	var stock entity.Stock
	if err := ctx.ShouldBindJSON(&stock); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	err := c.stockService.UpdateStock(&stock)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)
}

// InStock 入库操作
func (c *ProductController) InStock(ctx *gin.Context) {
	ctx.Set("oper_log_title", "产品入库")
	ctx.Set("oper_log_type", 1)
	ctx.Set("oper_log_business_type", "stock")

	var req struct {
		MaterialID  int64   `json:"material_id" binding:"required"`
		WarehouseID int64   `json:"warehouse_id" binding:"required"`
		Qty         float64 `json:"qty" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	err := c.stockService.InStock(req.MaterialID, req.WarehouseID, req.Qty)
	if err != nil {
		response.Error(ctx, response.ERROR, "入库失败")
		return
	}

	response.Success(ctx, nil)

	// 通知：物料入库
	userID := middleware.GetUserID(ctx)
	stockInMat, _ := dao.NewMaterialDAO().FindByID(req.MaterialID)
	stockInMatName := ""
	if stockInMat != nil {
		stockInMatName = stockInMat.MatName
		if stockInMatName == "" {
			stockInMatName = stockInMat.MatCode
		}
	} else {
		stockInMatName = fmt.Sprintf("#%d", req.MaterialID)
	}
	NotifyOperatorAndAdmins(userID, "库存管理",
		fmt.Sprintf("物料 %s 入库，数量：%.2f", stockInMatName, req.Qty),
		"stock:in", "middle",
		"", req.MaterialID)
}

// OutStock 出库操作
func (c *ProductController) OutStock(ctx *gin.Context) {
	ctx.Set("oper_log_title", "产品出库")
	ctx.Set("oper_log_type", 2)
	ctx.Set("oper_log_business_type", "stock")

	var req struct {
		MaterialID  int64   `json:"material_id" binding:"required"`
		WarehouseID int64   `json:"warehouse_id" binding:"required"`
		Qty         float64 `json:"qty" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	err := c.stockService.OutStock(req.MaterialID, req.WarehouseID, req.Qty)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, nil)

	// 通知：物料出库
	userID := middleware.GetUserID(ctx)
	stockOutMat, _ := dao.NewMaterialDAO().FindByID(req.MaterialID)
	stockOutMatName := ""
	if stockOutMat != nil {
		stockOutMatName = stockOutMat.MatName
		if stockOutMatName == "" {
			stockOutMatName = stockOutMat.MatCode
		}
	} else {
		stockOutMatName = fmt.Sprintf("#%d", req.MaterialID)
	}
	NotifyOperatorAndAdmins(userID, "库存管理",
		fmt.Sprintf("物料 %s 出库，数量：%.2f", stockOutMatName, req.Qty),
		"stock:out", "middle",
		"", req.MaterialID)
}

// CheckSafeStock 检查安全库存预警
func (c *ProductController) CheckSafeStock(ctx *gin.Context) {
	lowStocks, err := c.stockService.CheckSafeStock()
	if err != nil {
		response.Error(ctx, response.ERROR, "检查安全库存失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":  lowStocks,
		"total": len(lowStocks),
	})
}

package controller

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type ExternalMoldArchiveController struct {
	archiveService *service.ExternalMoldArchiveService
}

func NewExternalMoldArchiveController() *ExternalMoldArchiveController {
	return &ExternalMoldArchiveController{
		archiveService: service.NewExternalMoldArchiveService(),
	}
}

// GetByMoldCode 根据模具编号查询档案（用于自动填充）
func (c *ExternalMoldArchiveController) GetByMoldCode(ctx *gin.Context) {
	moldCode := strings.TrimSpace(ctx.Query("mold_code"))
	if moldCode == "" {
		response.Error(ctx, response.BAD_REQUEST, "模具编号不能为空")
		return
	}

	// 获取排除的订单ID（编辑模式下传入当前订单ID）
	var excludeOrderID *int64
	excludeOrderIDStr := ctx.Query("exclude_order_id")
	if excludeOrderIDStr != "" {
		id, err := strconv.ParseInt(excludeOrderIDStr, 10, 64)
		if err == nil && id > 0 {
			excludeOrderID = &id
		}
	}

	archive, err := c.archiveService.GetByMoldCode(moldCode)
	if err != nil {
		handleServiceError(ctx, err)
		return
	}

	if archive == nil {
		response.Success(ctx, gin.H{
			"exists":         false,
			"data":           nil,
			"pending_orders": nil,
		})
		return
	}

	// 查询该模具的未完成订单（排除当前编辑的订单）
	pendingOrders, err := c.archiveService.GetPendingOrdersByMoldCode(moldCode, excludeOrderID)
	if err != nil {
		pendingOrders = nil
	}

	response.Success(ctx, gin.H{
		"exists":         true,
		"data":           archive,
		"pending_orders": pendingOrders,
	})
}

// ListArchives 查询档案列表
func (c *ExternalMoldArchiveController) ListArchives(ctx *gin.Context) {
	moldCode := ctx.Query("mold_code")
	customerIDStr := ctx.Query("customer_id")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "10")

	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	var customerID *int64
	if customerIDStr != "" {
		id, err := strconv.ParseInt(customerIDStr, 10, 64)
		if err == nil && id > 0 {
			customerID = &id
		}
	}

	archives, total, err := c.archiveService.ListArchives(moldCode, customerID, page, pageSize)
	if err != nil {
		response.Error(ctx, response.ERROR, "查询失败")
		return
	}

	response.Success(ctx, gin.H{
		"list":      archives,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateArchive 更新档案信息
func (c *ExternalMoldArchiveController) UpdateArchive(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的档案ID")
		return
	}

	var req struct {
		ProductName string `json:"product_name" binding:"required"`
		Remark      string `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	if err := c.archiveService.UpdateArchive(id, req.ProductName, req.Remark); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{"message": "更新成功"})
}

// DeleteArchive 删除档案（软删除）
func (c *ExternalMoldArchiveController) DeleteArchive(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的档案ID")
		return
	}

	if err := c.archiveService.DeleteArchive(id); err != nil {
		handleServiceError(ctx, err)
		return
	}

	response.Success(ctx, gin.H{"message": "删除成功"})
}

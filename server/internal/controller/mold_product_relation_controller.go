package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type MoldProductRelationController struct {
	relationService *service.MoldProductRelationService
}

func NewMoldProductRelationController() *MoldProductRelationController {
	return &MoldProductRelationController{
		relationService: service.NewMoldProductRelationService(),
	}
}

// GetRelationsByMoldID 获取模具关联的所有产品
func (c *MoldProductRelationController) GetRelationsByMoldID(ctx *gin.Context) {
	moldIDStr := ctx.Query("mold_id")
	if moldIDStr == "" {
		response.Error(ctx, http.StatusBadRequest, "模具ID不能为空")
		return
	}

	moldID, err := strconv.ParseInt(moldIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的模具ID")
		return
	}

	relations, err := c.relationService.GetRelationsByMoldID(moldID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "查询失败: "+err.Error())
		return
	}

	response.Success(ctx, relations)
}

// GetRelationsByProductID 获取产品关联的所有模具
func (c *MoldProductRelationController) GetRelationsByProductID(ctx *gin.Context) {
	productIDStr := ctx.Query("product_id")
	if productIDStr == "" {
		response.Error(ctx, http.StatusBadRequest, "产品ID不能为空")
		return
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的产品ID")
		return
	}

	relations, err := c.relationService.GetRelationsByProductID(productID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "查询失败: "+err.Error())
		return
	}

	response.Success(ctx, relations)
}

// GetRelationsByItemID 获取生产订单明细关联的模具
func (c *MoldProductRelationController) GetRelationsByItemID(ctx *gin.Context) {
	itemIDStr := ctx.Query("item_id")
	if itemIDStr == "" {
		response.Error(ctx, http.StatusBadRequest, "订单明细ID不能为空")
		return
	}

	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的订单明细ID")
		return
	}

	relations, err := c.relationService.GetRelationsByItemID(itemID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "查询失败: "+err.Error())
		return
	}

	response.Success(ctx, relations)
}

// CreateRelation 创建模具产品关联
func (c *MoldProductRelationController) CreateRelation(ctx *gin.Context) {
	var req struct {
		MoldID                int64  `json:"mold_id" binding:"required"`
		ProductID             *int64 `json:"product_id"`
		ProductionOrderItemID *int64 `json:"production_order_item_id"`
		ProductName           string `json:"product_name"`
		ProductImageURL       string `json:"product_image_url"`
		PlanQty               int    `json:"plan_qty"`
		RelationType          string `json:"relation_type"`
		Remark                string `json:"remark"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误: "+err.Error())
		return
	}

	relation := &entity.MoldProductRelation{
		MoldID:                req.MoldID,
		ProductionOrderItemID: req.ProductionOrderItemID,
		ProductName:           req.ProductName,
		ProductImageURL:       req.ProductImageURL,
		PlanQty:               req.PlanQty,
		Remark:                req.Remark,
	}

	if err := c.relationService.CreateRelation(nil, relation); err != nil {
		response.Error(ctx, http.StatusBadRequest, "创建失败: "+err.Error())
		return
	}

	response.Success(ctx, relation)
}

// DeleteRelation 删除模具产品关联
func (c *MoldProductRelationController) DeleteRelation(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "无效的ID")
		return
	}

	if err := c.relationService.DeleteRelationByID(id); err != nil {
		response.Error(ctx, http.StatusBadRequest, "删除失败: "+err.Error())
		return
	}

	response.Success(ctx, gin.H{"message": "删除成功"})
}

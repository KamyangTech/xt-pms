package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type StatusFlowController struct {
	statusFlowService *service.StatusFlowService
}

func NewStatusFlowController() *StatusFlowController {
	return &StatusFlowController{
		statusFlowService: service.NewStatusFlowService(),
	}
}

// GetStatusFlowConfigs 获取所有状态同步配置
func (sfc *StatusFlowController) GetStatusFlowConfigs(c *gin.Context) {
	configs, err := sfc.statusFlowService.GetAll()
	if err != nil {
		response.Error(c, response.ERROR, "获取状态同步配置失败")
		return
	}

	response.Success(c, gin.H{
		"list":  configs,
		"total": len(configs),
	})
}

// GetStatusFlowConfigByID 根据 ID 获取状态同步配置
func (sfc *StatusFlowController) GetStatusFlowConfigByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}

	config, err := sfc.statusFlowService.GetByID(id)
	if err != nil {
		response.Error(c, response.NOT_FOUND, "配置不存在")
		return
	}

	response.Success(c, config)
}

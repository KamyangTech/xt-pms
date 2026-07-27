package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type ProcessTemplateController struct {
	processTemplateService *service.ProcessTemplateService
}

func NewProcessTemplateController() *ProcessTemplateController {
	return &ProcessTemplateController{
		processTemplateService: service.NewProcessTemplateService(),
	}
}

// CreateProcessTemplate 创建工序模板
func (ctrl *ProcessTemplateController) CreateProcessTemplate(c *gin.Context) {
	var template entity.ProcessTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	result, err := ctrl.processTemplateService.CreateProcessTemplate(&template)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "创建失败")
		return
	}

	response.Success(c, result)
}

// GetProcessTemplateList 获取工序模板列表
func (ctrl *ProcessTemplateController) GetProcessTemplateList(c *gin.Context) {
	templates, err := ctrl.processTemplateService.GetAllProcessTemplates()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "查询失败")
		return
	}

	response.Success(c, templates)
}

// GetActiveProcessTemplates 获取启用的工序模板列表
func (ctrl *ProcessTemplateController) GetActiveProcessTemplates(c *gin.Context) {
	templates, err := ctrl.processTemplateService.GetActiveProcessTemplates()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "查询失败")
		return
	}

	response.Success(c, templates)
}

// UpdateProcessTemplate 更新工序模板
func (ctrl *ProcessTemplateController) UpdateProcessTemplate(c *gin.Context) {
	var template entity.ProcessTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := ctrl.processTemplateService.UpdateProcessTemplate(&template); err != nil {
		response.Error(c, http.StatusBadRequest, "更新失败")
		return
	}

	response.Success(c, nil)
}

// DeleteProcessTemplate 删除工序模板
func (ctrl *ProcessTemplateController) DeleteProcessTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	if err := ctrl.processTemplateService.DeleteProcessTemplate(id); err != nil {
		response.Error(c, http.StatusBadRequest, "删除失败")
		return
	}

	response.Success(c, nil)
}

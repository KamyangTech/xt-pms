package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type CompanyController struct {
	companyService *service.SysCompanyService
}

func NewCompanyController() *CompanyController {
	return &CompanyController{
		companyService: service.NewSysCompanyService(),
	}
}

// GetCompany 获取公司信息
func (c *CompanyController) GetCompany(ctx *gin.Context) {
	company, err := c.companyService.GetCompany()
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "获取公司信息失败")
		return
	}
	response.Success(ctx, company)
}

// UpdateCompany 更新公司信息
func (c *CompanyController) UpdateCompany(ctx *gin.Context) {
	var req dto.CompanyDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, "参数错误")
		return
	}

	company := &entity.SysCompany{
		CompanyName:      req.CompanyName,
		ShortName:        req.ShortName,
		EnglishName:      req.EnglishName,
		EnglishShortName: req.EnglishShortName,
		Logo:             req.Logo,
	}

	if err := c.companyService.SaveCompany(company); err != nil {
		response.Error(ctx, http.StatusBadRequest, "保存公司信息失败")
		return
	}

	response.Success(ctx, nil)
}

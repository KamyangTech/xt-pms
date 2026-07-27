package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type SearchController struct {
	searchService *service.SearchService
}

func NewSearchController() *SearchController {
	return &SearchController{
		searchService: service.NewSearchService(),
	}
}

// Search 全局搜索接口
// GET /api/search?keyword=xxx&page=1&page_size=20
func (c *SearchController) Search(ctx *gin.Context) {
	var params dto.SearchParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误")
		return
	}

	if params.Keyword == "" {
		response.Error(ctx, response.BAD_REQUEST, "搜索关键词不能为空")
		return
	}

	result, err := c.searchService.GlobalSearch(params.Keyword, params.Page, params.PageSize)
	if err != nil {
		response.Error(ctx, response.ERROR, "搜索失败: "+err.Error())
		return
	}

	response.Success(ctx, result)
}

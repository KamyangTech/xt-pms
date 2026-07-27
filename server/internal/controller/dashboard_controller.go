package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type DashboardController struct{}

func NewDashboardController() *DashboardController {
	return &DashboardController{}
}

// getDashboardService 获取 DashboardService 实例（延迟获取，确保 DB 已初始化）
func getDashboardService() *service.DashboardService {
	return service.NewDashboardService()
}

// GetOverview 获取概览统计数据
func (dc *DashboardController) GetOverview(c *gin.Context) {
	svc := getDashboardService()
	overview, err := svc.GetOverview()
	if err != nil {
		response.Error(c, response.ERROR, "获取数据失败："+err.Error())
		return
	}

	response.Success(c, overview)
}

// GetProductionTrend 获取生产趋势数据
func (dc *DashboardController) GetProductionTrend(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	svc := getDashboardService()
	trendData, err := svc.GetProductionTrend(period, startDate, endDate)
	if err != nil {
		response.Error(c, response.ERROR, "获取趋势数据失败："+err.Error())
		return
	}

	response.Success(c, trendData)
}

// GetMoldStatus 获取模具状态统计
func (dc *DashboardController) GetMoldStatus(c *gin.Context) {
	svc := getDashboardService()
	status, err := svc.GetMoldStatus()
	if err != nil {
		response.Error(c, response.ERROR, "获取模具状态失败："+err.Error())
		return
	}

	response.Success(c, status)
}

// GetInProductionMolds 获取正在制造的模具
func (dc *DashboardController) GetInProductionMolds(c *gin.Context) {
	svc := getDashboardService()
	molds, err := svc.GetInProductionMoldList(10)
	if err != nil {
		response.Error(c, response.ERROR, "获取数据失败："+err.Error())
		return
	}
	response.Success(c, molds)
}

// GetInProgressProductionOrders 获取进行中的生产订单（固定8条最近到期）
func (dc *DashboardController) GetInProgressProductionOrders(c *gin.Context) {
	svc := getDashboardService()
	result, err := svc.GetInProgressProductionOrders(8)
	if err != nil {
		response.Error(c, response.ERROR, "获取数据失败："+err.Error())
		return
	}
	response.Success(c, result)
}

// GetInProgressMoldOrders 获取进行中的模具订单（固定8条最近到期）
func (dc *DashboardController) GetInProgressMoldOrders(c *gin.Context) {
	svc := getDashboardService()
	result, err := svc.GetInProgressMoldOrders(8)
	if err != nil {
		response.Error(c, response.ERROR, "获取数据失败："+err.Error())
		return
	}
	response.Success(c, result)
}

// GetOrderRanking 获取订单排名
func (dc *DashboardController) GetOrderRanking(c *gin.Context) {
	rankingType := c.DefaultQuery("type", "customer") // customer or mold
	limit := 6                                        // 默认返回前6条

	svc := getDashboardService()
	rankingData, err := svc.GetOrderRanking(rankingType, limit)
	if err != nil {
		response.Error(c, response.ERROR, "获取排名数据失败："+err.Error())
		return
	}

	response.Success(c, rankingData)
}

// GetFinanceOverview 获取财务概览数据
func (dc *DashboardController) GetFinanceOverview(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	if startDate == "" || endDate == "" {
		response.Error(c, response.ERROR, "缺少日期参数")
		return
	}

	svc := getDashboardService()
	data, err := svc.GetFinanceOverview(startDate, endDate)
	if err != nil {
		response.Error(c, response.ERROR, "获取财务概览失败："+err.Error())
		return
	}

	response.Success(c, data)
}

// GetDetailOverview 获取统计报表页概览数据（4张卡片）
func (dc *DashboardController) GetDetailOverview(c *gin.Context) {
	svc := getDashboardService()
	data, err := svc.GetDetailOverview()
	if err != nil {
		response.Error(c, response.ERROR, "获取数据失败："+err.Error())
		return
	}
	response.Success(c, data)
}

// GetDetailProductionTrend 获取统计报表页趋势图数据（近12个月月度趋势）
func (dc *DashboardController) GetDetailProductionTrend(c *gin.Context) {
	svc := getDashboardService()
	data, err := svc.GetDetailProductionTrend()
	if err != nil {
		response.Error(c, response.ERROR, "获取趋势数据失败："+err.Error())
		return
	}
	response.Success(c, data)
}

// GetDeliveryPunctuality 获取交付准时率数据（近12个月）
func (dc *DashboardController) GetDeliveryPunctuality(c *gin.Context) {
	svc := getDashboardService()
	data, err := svc.GetDeliveryPunctuality()
	if err != nil {
		response.Error(c, response.ERROR, "获取交付准时率数据失败："+err.Error())
		return
	}
	response.Success(c, data)
}

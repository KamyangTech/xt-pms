package monitor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
)

// SimpleHealth 公开健康检查专用结构体，不暴露任何系统内部指标
type SimpleHealth struct {
	Status string `json:"status"`
}

// HealthHandler GET /health — 公开，仅返回 ok/down
// 供负载均衡器、K8s livenessProbe 使用，不泄露系统内部信息
// 注意：必须使用标准响应包装，前端才能解析
func HealthHandler(c *gin.Context) {
	status := CheckHealth()
	if status.Status == "down" {
		c.JSON(http.StatusServiceUnavailable, response.Response{
			Code:    response.ERROR,
			Message: "service unavailable",
			Data:    SimpleHealth{Status: "down"},
		})
		return
	}
	c.JSON(http.StatusOK, response.Response{
		Code:    response.SUCCESS,
		Message: "success",
		Data:    SimpleHealth{Status: status.Status},
	})
}

// FullHealthHandler GET /api/debug/health — 需要认证，返回完整健康详情
func FullHealthHandler(c *gin.Context) {
	status := CheckHealth()
	resp := response.Response{
		Code:    response.SUCCESS,
		Message: "success",
		Data:    status,
	}
	httpCode := http.StatusOK
	if status.Status == "down" {
		resp.Code = response.ERROR
		resp.Message = "service unavailable"
		httpCode = http.StatusServiceUnavailable
	}
	c.JSON(httpCode, resp)
}

// MetricsHandler GET /api/debug/metrics — 需要认证，返回运行时指标
func MetricsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, response.Response{
		Code:    response.SUCCESS,
		Message: "success",
		Data:    Snapshot(),
	})
}

// PoolHandler GET /api/debug/pool — 需要认证，返回 DB 连接池详情
func PoolHandler(c *gin.Context) {
	stats := GetPoolStats()
	if stats == nil {
		c.JSON(http.StatusInternalServerError, response.Response{
			Code:    response.ERROR,
			Message: "无法获取连接池状态",
		})
		return
	}
	c.JSON(http.StatusOK, response.Response{
		Code:    response.SUCCESS,
		Message: "success",
		Data:    stats,
	})
}

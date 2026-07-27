package controller

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

type OperLogController struct {
	service *service.SysOperLogService
}

func NewOperLogController() *OperLogController {
	return &OperLogController{
		service: service.NewSysOperLogService(),
	}
}

// GetList 分页获取操作日志
func (oc *OperLogController) GetList(c *gin.Context) {
	HandlePageQuery(c, func(page, pageSize int) (interface{}, int64, error) {
		title := c.Query("title")
		operName := c.Query("oper_name")
		operTypeStr := c.DefaultQuery("oper_type", "0")
		statusStr := c.DefaultQuery("status", "-1")
		businessType := c.Query("business_type")
		startTime := c.Query("start_time")
		endTime := c.Query("end_time")

		operType, _ := strconv.Atoi(operTypeStr)
		status, _ := strconv.Atoi(statusStr)

		return oc.service.GetListByPage(page, pageSize, title, operName, int8(operType), int8(status), businessType, startTime, endTime)
	})
}

// GetDetail 获取操作日志详情
func (oc *OperLogController) GetDetail(c *gin.Context) {
	HandlePathID(c, func(id int64) (interface{}, error) {
		return oc.service.GetByID(id)
	})
}

// Export 导出操作日志为 CSV（直接流式写入 HTTP 响应）
func (oc *OperLogController) Export(c *gin.Context) {
	title := c.Query("title")
	operName := c.Query("oper_name")
	operTypeStr := c.DefaultQuery("oper_type", "0")
	statusStr := c.DefaultQuery("status", "-1")
	businessType := c.Query("business_type")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	operType, _ := strconv.Atoi(operTypeStr)
	status, _ := strconv.Atoi(statusStr)

	filename := fmt.Sprintf("操作日志_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Header("Content-Type", "text/csv; charset=utf-8")

	if err := oc.service.ExportCSV(c.Writer, title, operName, int8(operType), int8(status), businessType, startTime, endTime); err != nil {
		// 此时 header 已写入，无法返回 JSON 错误，直接 panic 由 Recovery 中间件处理
		panic(err)
	}
}

// GetOperTypeOptions 获取操作类型选项（供前端下拉筛选）
func (oc *OperLogController) GetOperTypeOptions(c *gin.Context) {
	options := []gin.H{
		{"label": entity.OperTypeName(1), "value": 1},
		{"label": entity.OperTypeName(2), "value": 2},
		{"label": entity.OperTypeName(3), "value": 3},
		{"label": entity.OperTypeName(4), "value": 4},
		{"label": entity.OperTypeName(5), "value": 5},
		{"label": entity.OperTypeName(6), "value": 6},
		{"label": entity.OperTypeName(7), "value": 7},
		{"label": entity.OperTypeName(8), "value": 8},
	}
	response.Success(c, options)
}

package controller

import (
	"sort"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/service"
)

// OutsourcingPublicController 外发公开控制器（无需JWT认证，供加工商使用）
type OutsourcingPublicController struct {
	outsourcingService *service.OutsourcingService
}

func NewOutsourcingPublicController() *OutsourcingPublicController {
	return &OutsourcingPublicController{
		outsourcingService: service.NewOutsourcingService(),
	}
}

// GetByToken 通过分享token获取外发记录信息（公开API）
func (c *OutsourcingPublicController) GetByToken(ctx *gin.Context) {
	token := ctx.Param("token")
	if token == "" {
		response.Error(ctx, response.BAD_REQUEST, "无效的分享链接")
		return
	}

	record, err := c.outsourcingService.GetOutsourcingByShareToken(token)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, err.Error())
		return
	}

	processorName := ""
	processorID := int64(0)
	if record.Processor != nil {
		processorName = record.Processor.ProcessorName
		processorID = record.Processor.ID
	}

	// 顺带返回公司名称（加工商汇报页面展示用）
	companyService := service.NewSysCompanyService()
	companyName := ""
	if company, cerr := companyService.GetCompany(); cerr == nil && company != nil {
		companyName = company.CompanyName
	}

	// 查询历史报工记录
	reportDAO := dao.NewOutsourcingReportDAO()
	reports, _ := reportDAO.FindByOutsourcingID(record.ID)
	reportList := make([]gin.H, 0, len(reports))
	for _, r := range reports {
		reportList = append(reportList, gin.H{
			"id":           r.ID,
			"report_date":  r.ReportDate,
			"ok_qty":       r.OkQty,
			"bad_qty":      r.BadQty,
			"defect_reason": r.DefectReason,
			"remark":       r.Remark,
		})
	}

	resp := gin.H{
		"id":            record.ID,
		"record_code":   record.RecordCode,
		"outsource_type": record.OutsourceType,
		"processor": gin.H{
			"id":   processorID,
			"name": processorName,
		},
		"quantity":             record.Quantity,
		"status":               record.Status,
		"outsource_date":       record.OutsourceDate,
		"expected_return_date": record.ExpectedReturnDate,
		"remark":               record.Remark,
		"company_name":         companyName,
		"mold":                 record.Mold,
		"work_order":           record.WorkOrder,
		"reports":              reportList,
	}

	// 模具外发：附带工单工序信息
	if record.OutsourceType == "MOLD" && record.WorkOrderID != nil {
		processDAO := dao.NewWorkOrderProcessDAO()
		processes, _ := processDAO.FindByWorkOrderID(*record.WorkOrderID)

		processTemplateDAO := dao.NewProcessTemplateDAO()
		templates, _ := processTemplateDAO.FindActive()

		processList := make([]gin.H, 0, len(templates))
		for _, t := range templates {
			status := int8(0)
			var completedAt interface{}
			for _, p := range processes {
				if p.ProcessTemplateID == t.ID {
					status = p.IsCompleted
					completedAt = p.CompletedAt
					break
				}
			}
			processList = append(processList, gin.H{
				"template_id":   t.ID,
				"process_name":  t.ProcessName,
				"group_id":      t.GroupID,
				"group_name":    t.GroupName,
				"sort":          t.Sort,
				"is_completed":  status,
				"completed_at":  completedAt,
			})
		}

		// 按分组排序
		sort.Slice(processList, func(i, j int) bool {
			gi := processList[i]["group_id"].(int)
			gj := processList[j]["group_id"].(int)
			if gi != gj {
				return gi < gj
			}
			return processList[i]["sort"].(int) < processList[j]["sort"].(int)
		})

		resp["processes"] = processList
	}

	response.Success(ctx, resp)
}

// AddReport 通过分享token添加外发报工（公开API，加工商提交进度）
func (c *OutsourcingPublicController) AddReport(ctx *gin.Context) {
	token := ctx.Param("token")
	if token == "" {
		response.Error(ctx, response.BAD_REQUEST, "无效的分享链接")
		return
	}

	var reportDTO dto.OutsourcingPublicReportDTO
	if err := ctx.ShouldBindJSON(&reportDTO); err != nil {
		response.Error(ctx, response.BAD_REQUEST, "请求参数错误: "+err.Error())
		return
	}

	if err := c.outsourcingService.AddPublicReport(token, reportDTO); err != nil {
		response.Error(ctx, response.ERROR, err.Error())
		return
	}

	response.Success(ctx, nil)
}

// ToggleProcess 公开 API：加工商切换工序完成状态
func (c *OutsourcingPublicController) ToggleProcess(ctx *gin.Context) {
	token := ctx.Param("token")
	if token == "" {
		response.Error(ctx, response.BAD_REQUEST, "无效的分享链接")
		return
	}

	templateIDStr := ctx.Query("template_id")
	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		response.Error(ctx, response.BAD_REQUEST, "无效的工序ID")
		return
	}

	record, err := c.outsourcingService.GetOutsourcingByShareToken(token)
	if err != nil {
		response.Error(ctx, response.NOT_FOUND, "分享链接无效或已过期")
		return
	}

	if record.WorkOrderID == nil {
		response.Error(ctx, response.BAD_REQUEST, "该外发记录没有关联工单")
		return
	}

	newStatus := int8(1)
	if statusStr := ctx.Query("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil && (s == 0 || s == 1 || s == 2) {
			newStatus = int8(s)
		}
	} else {
		// 自动 toggle：取当前状态反转
		processDAO := dao.NewWorkOrderProcessDAO()
		existing, _ := processDAO.FindByWorkOrderIDAndTemplateID(*record.WorkOrderID, templateID)
		if existing != nil && existing.IsCompleted == 1 {
			newStatus = 0
		}
	}

	processDAO := dao.NewWorkOrderProcessDAO()
	if err := processDAO.UpsertProcess(*record.WorkOrderID, templateID, newStatus); err != nil {
		response.Error(ctx, response.ERROR, "更新工序状态失败")
		return
	}

	response.Success(ctx, gin.H{"is_completed": newStatus})
}

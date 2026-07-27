package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/service"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type WorkOrderProcessController struct {
	workOrderProcessService *service.WorkOrderProcessService
}

func NewWorkOrderProcessController() *WorkOrderProcessController {
	return &WorkOrderProcessController{
		workOrderProcessService: service.NewWorkOrderProcessService(),
	}
}

// GetWorkOrderProcesses 获取工单的工序列表
func (ctrl *WorkOrderProcessController) GetWorkOrderProcesses(c *gin.Context) {
	idStr := c.Param("id")

	if idStr == "" {
		response.Error(c, http.StatusBadRequest, "工单ID不能为空")
		return
	}

	workOrderID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "参数错误")
		return
	}

	processes, err := ctrl.workOrderProcessService.GetWorkOrderProcesses(workOrderID)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "查询失败")
		return
	}

	response.Success(c, processes)
}

// ToggleProcess 切换工序完成状态
func (ctrl *WorkOrderProcessController) ToggleProcess(c *gin.Context) {
	var req dto.ToggleProcessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("工序切换参数绑定失败", zap.Error(err))
		response.Error(c, http.StatusBadRequest, fmt.Sprintf("参数错误: %v", err))
		return
	}

	logger.Debug("收到工序切换请求",
		zap.Int64("work_order_id", req.WorkOrderID),
		zap.Int64("process_template_id", req.ProcessTemplateID),
		zap.Int8("is_completed", int8(req.IsCompleted)))

	if err := ctrl.workOrderProcessService.ToggleProcess(req.WorkOrderID, req.ProcessTemplateID, int8(req.IsCompleted)); err != nil {
		// 输出详细错误信息用于调试
		logger.Error("切换工序状态失败",
			zap.Int64("work_order_id", req.WorkOrderID),
			zap.Int64("process_template_id", req.ProcessTemplateID),
			zap.Error(err))
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// 通知：工序状态变更（通知操作人 + 管理员）
	userID := middleware.GetUserID(c)
	go func() {
		// 文案使用"工序"前缀，避免与工单/模具整体状态混淆
		var actionText string
		switch int8(req.IsCompleted) {
		case 1:
			actionText = "已完成"
		case 2:
			actionText = "已跳过"
		case 0:
			actionText = "已撤销完成"
		default:
			actionText = "状态已更新"
		}

		moldCode := fmt.Sprintf("#%d", req.WorkOrderID)
		processName := fmt.Sprintf("#%d", req.ProcessTemplateID)

		if workOrder, err := dao.NewMoldWorkOrderDAO().FindByID(req.WorkOrderID); err == nil && workOrder.MoldID != nil {
			if mold, err := dao.NewMoldDAO().FindByID(*workOrder.MoldID); err == nil {
				moldCode = mold.MoldCode
			}
		}
		if pt, err := dao.NewProcessTemplateDAO().FindByID(req.ProcessTemplateID); err == nil {
			processName = pt.ProcessName
		}

		NotifyOperatorAndAdmins(userID, "模具工单",
			fmt.Sprintf("模具 %s 的「%s」工序%s", moldCode, processName, actionText),
			"mold:work_order_status", "middle",
			fmt.Sprintf("/mold/work-order/%d", req.WorkOrderID), req.WorkOrderID)
	}()

	response.Success(c, nil)
}

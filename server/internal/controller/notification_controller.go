package controller

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/service"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type NotificationController struct {
	notificationService       *service.NotificationService
	notificationConfigService *service.NotificationConfigService
}

func NewNotificationController() *NotificationController {
	return &NotificationController{
		notificationService:       service.NewNotificationService(),
		notificationConfigService: service.NewNotificationConfigService(),
	}
}

// GetPage 获取当前用户通知分页
func (nc *NotificationController) GetPage(c *gin.Context) {
	var query dto.NotificationPageQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	result, err := nc.notificationService.GetPage(userID, &query)
	if err != nil {
		response.Error(c, response.ERROR, "获取通知列表失败："+err.Error())
		return
	}

	response.Success(c, result)
}

// GetUnreadCount 获取未读通知数量
func (nc *NotificationController) GetUnreadCount(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count, err := nc.notificationService.GetUnreadCount(userID)
	if err != nil {
		response.Error(c, response.ERROR, "获取未读数量失败："+err.Error())
		return
	}

	response.Success(c, map[string]int64{"count": count})
}

// GetTypes 获取当前用户通知的去重类型列表
func (nc *NotificationController) GetTypes(c *gin.Context) {
	userID := middleware.GetUserID(c)
	types, err := nc.notificationService.GetDistinctTypes(userID)
	if err != nil {
		response.Error(c, response.ERROR, "获取通知类型失败："+err.Error())
		return
	}
	response.Success(c, types)
}

// GetAllTypes 获取所有通知的去重类型列表（全局，用于配置页面的下拉选项）
func (nc *NotificationController) GetAllTypes(c *gin.Context) {
	types, err := nc.notificationService.GetAllDistinctTypes()
	if err != nil {
		response.Error(c, response.ERROR, "获取通知类型失败："+err.Error())
		return
	}
	response.Success(c, types)
}

// MarkAsRead 标记单条通知为已读
func (nc *NotificationController) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	if err := nc.notificationService.MarkAsRead(id, userID); err != nil {
		response.Error(c, response.ERROR, "操作失败："+err.Error())
		return
	}

	response.Success(c, nil)
}

// MarkAllAsRead 全部标记已读
func (nc *NotificationController) MarkAllAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := nc.notificationService.MarkAllAsRead(userID); err != nil {
		response.Error(c, response.ERROR, "操作失败："+err.Error())
		return
	}

	response.Success(c, nil)
}

// Delete 删除通知
func (nc *NotificationController) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}

	userID := middleware.GetUserID(c)
	if err := nc.notificationService.Delete(id, userID); err != nil {
		response.Error(c, response.ERROR, "删除失败："+err.Error())
		return
	}

	response.Success(c, nil)
}

// TestSend 测试发送通知（仅开发环境使用）
func (nc *NotificationController) TestSend(c *gin.Context) {
	userID := middleware.GetUserID(c)
	now := time.Now().Format("15:04:05")

	notifs := []struct {
		title, content, ntype, typeLabel, quality string
	}{
		{"[模具工单] 工单 MO-2024-001 已完成加工", "工单 MO-2024-001 已完成全部加工工序，请安排试模。", "mold:work_order_status", "模具工单", "high"},
		{"[生产订单] PO-2024-001 已完工入库", "生产订单 PO-2024-001 已完成生产并入库，请安排发货。", "production:order_status", "生产订单", "middle"},
		{"[品质检验] 产品 QC-2024-001 检验不合格", "产品 QC-2024-001 抽检不合格率 15%，超过标准 5%，请处理。", "production:quality_fail", "品质检验", "high"},
		{"[应收账款] 客户 XX 公司款项即将到期", "应收款 AR-2024-001 距离到期日还有 3 天，请及时跟进。", "finance:receivable_due", "应收账款", "low"},
		{"[外发管理] 外发任务已返回", "外发加工任务 OS-2024-001 已完成并返回，请安排入库检验。", "outsourcing:returned", "外发管理", "middle"},
	}

	for _, n := range notifs {
		nc.SendNotification(userID, n.title+" ("+now+")", n.content, n.ntype, n.typeLabel, n.quality, "", 0)
	}

	response.Success(c, gin.H{"message": fmt.Sprintf("已发送 %d 条测试通知给您", len(notifs))})
}

// ---------- 通知配置管理 ----------

// GetConfigs 获取通知配置分页列表
func (nc *NotificationController) GetConfigs(c *gin.Context) {
	query := struct {
		Page     int `form:"page" json:"page"`
		PageSize int `form:"page_size" json:"page_size"`
	}{Page: 1, PageSize: 20}
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}
	result, err := nc.notificationConfigService.GetPage(query.Page, query.PageSize)
	if err != nil {
		response.Error(c, response.ERROR, "获取配置列表失败："+err.Error())
		return
	}
	response.Success(c, result)
}

// CreateConfig 创建通知配置
func (nc *NotificationController) CreateConfig(c *gin.Context) {
	var config entity.SysNotificationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误："+err.Error())
		return
	}
	if err := nc.notificationConfigService.Create(&config); err != nil {
		response.Error(c, response.ERROR, "创建失败："+err.Error())
		return
	}
	response.Success(c, config)
}

// UpdateConfig 更新通知配置
func (nc *NotificationController) UpdateConfig(c *gin.Context) {
	var config entity.SysNotificationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Error(c, response.BAD_REQUEST, "请求参数错误："+err.Error())
		return
	}
	if err := nc.notificationConfigService.Update(&config); err != nil {
		response.Error(c, response.ERROR, "更新失败："+err.Error())
		return
	}
	response.Success(c, nil)
}

// DeleteConfig 删除通知配置
func (nc *NotificationController) DeleteConfig(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Error(c, response.BAD_REQUEST, "参数错误")
		return
	}
	if err := nc.notificationConfigService.Delete(id); err != nil {
		response.Error(c, response.ERROR, "删除失败："+err.Error())
		return
	}
	response.Success(c, nil)
}

// ---------- 以下方法供其他业务模块调用 ----------

// SendNotification 发送通知给单个用户（入库 + WebSocket 推送）
func (nc *NotificationController) SendNotification(userID int64, title, content, notifType, typeLabel, quality, refURL string, refID int64) error {
	// 1. 入库
	notif, err := nc.notificationService.CreateNotification(userID, title, content, notifType, typeLabel, quality, refURL, refID)
	if err != nil {
		return err
	}

	// 2. WebSocket 推送
	WS.SendToUser(userID, "notification", map[string]interface{}{
		"id":          notif.ID,
		"title":       title,
		"content":     content,
		"type":        notifType,
		"type_label":  typeLabel,
		"quality":     quality,
		"ref_id":      refID,
		"ref_url":     refURL,
		"is_read":     0,
		"create_time": notif.CreateTime.Format("2006-01-02 15:04:05"),
	})

	return nil
}

// SendNotificationToUsers 发送通知给多个用户
func (nc *NotificationController) SendNotificationToUsers(userIDs []int64, title, content, notifType, typeLabel, quality, refURL string, refID int64) error {
	for _, uid := range userIDs {
		if err := nc.SendNotification(uid, title, content, notifType, typeLabel, quality, refURL, refID); err != nil {
			return err
		}
	}
	return nil
}

// ---------- 全局快捷函数（供其他 controller 调用） ----------

var _globalNotifController *NotificationController

func getGlobalNotifController() *NotificationController {
	if _globalNotifController == nil {
		_globalNotifController = NewNotificationController()
	}
	return _globalNotifController
}

// SendNotify 全局快捷函数：发送通知给单个用户
func SendNotify(userID int64, title, content, notifType, typeLabel, quality, refURL string, refID int64) error {
	return getGlobalNotifController().SendNotification(userID, title, content, notifType, typeLabel, quality, refURL, refID)
}

// SendNotifyWrapper 发送格式化通知（供业务模块调用）
// usage: controller.SendNotifyWrapper(userID, "模具工单", fmt.Sprintf("工单 %s 状态变更为 %s", orderNo, status), "mold:work_order_status", "high", "/mold/work-order/123", 123)
func SendNotifyWrapper(userID int64, typeLabel, content, notifType, quality, refURL string, refID int64) error {
	title := fmt.Sprintf("[%s] %s", typeLabel, content)

	// 截断过长标题（最多 200 字符）
	titleRunes := []rune(title)
	if len(titleRunes) > 200 {
		title = string(titleRunes[:197]) + "..."
	}
	contentRunes := []rune(content)
	if len(contentRunes) > 500 {
		content = string(contentRunes[:497]) + "..."
	}

	return getGlobalNotifController().SendNotification(userID, title, content, notifType, typeLabel, quality, refURL, refID)
}

// ---------- 管理员通知辅助函数 ----------

// NotifyOperatorAndAdmins 发送通知给操作人 + 根据通知配置自动路由给对应角色/用户
// 操作人如果也匹配配置规则，不会重复发送
func NotifyOperatorAndAdmins(operatorID int64, typeLabel, content, notifType, quality, refURL string, refID int64) error {
	// 1. 先发送给操作人
	if err := SendNotifyWrapper(operatorID, typeLabel, content, notifType, quality, refURL, refID); err != nil {
		logger.Error("通知操作人失败", zap.Error(err))
	}

	// 2. 查询通知配置，获取其他应接收的用户
	configSvc := service.NewNotificationConfigService()
	recipientIDs, err := configSvc.GetRecipientIDs(notifType)
	if err != nil {
		logger.Error("查询通知配置失败", zap.Error(err))
		return nil // 配置查询失败不影响主流程
	}

	// 3. 发送给所有配置的用户（排除操作人自己，避免重复）
	for _, uid := range recipientIDs {
		if uid == operatorID {
			continue
		}
		if err := SendNotifyWrapper(uid, typeLabel, content, notifType, quality, refURL, refID); err != nil {
			logger.Error("通知接收者失败", zap.Int64("user_id", uid), zap.Error(err))
		}
	}

	return nil
}

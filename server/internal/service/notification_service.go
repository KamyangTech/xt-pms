package service

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

type NotificationService struct {
	notificationDAO *dao.NotificationDAO
}

func NewNotificationService() *NotificationService {
	return &NotificationService{
		notificationDAO: dao.NewNotificationDAO(),
	}
}

// CreateNotification 创建通知（仅入库，不推送）
func (s *NotificationService) CreateNotification(userID int64, title, content, notifType, typeLabel, quality, refURL string, refID int64) (*entity.SysNotification, error) {
	notif := &entity.SysNotification{
		UserID:     userID,
		Title:      title,
		Content:    content,
		Type:       notifType,
		TypeLabel:  typeLabel,
		Quality:    quality,
		RefID:      refID,
		RefURL:     refURL,
		IsRead:     0,
		CreateTime: time.Now(),
	}

	if err := s.notificationDAO.Create(notif); err != nil {
		logger.Error("保存通知失败", zap.Error(err))
		return nil, err
	}

	return notif, nil
}

// GetPage 获取用户通知分页
func (s *NotificationService) GetPage(userID int64, query *dto.NotificationPageQuery) (*utils.PageResult, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	result, err := s.notificationDAO.FindPageByUserID(userID, query.Page, query.PageSize, query.IsRead, query.Type)
	if err != nil {
		return nil, err
	}

	var dtoList []dto.NotificationListDTO
	if notifications, ok := result.List.(*[]*entity.SysNotification); ok && notifications != nil {
		for _, n := range *notifications {
			dtoList = append(dtoList, dto.NotificationListDTO{
				ID:         n.ID,
				Title:      n.Title,
				Content:    n.Content,
				Type:       n.Type,
				TypeLabel:  n.TypeLabel,
				Quality:    n.Quality,
				RefID:      n.RefID,
				RefURL:     n.RefURL,
				IsRead:     n.IsRead,
				CreateTime: n.CreateTime.Format("2006-01-02 15:04:05"),
			})
		}
	}
	result.List = dtoList
	return result, nil
}

// GetUnreadCount 获取用户未读通知数量
func (s *NotificationService) GetUnreadCount(userID int64) (int64, error) {
	return s.notificationDAO.CountUnreadByUserID(userID)
}

// GetDistinctTypes 获取用户通知的去重类型列表
func (s *NotificationService) GetDistinctTypes(userID int64) ([]string, error) {
	return s.notificationDAO.FindDistinctTypesByUserID(userID)
}

// GetAllDistinctTypes 获取所有通知的去重类型列表（全局，不限制用户）
func (s *NotificationService) GetAllDistinctTypes() ([]string, error) {
	return s.notificationDAO.FindAllDistinctTypes()
}

// MarkAsRead 标记已读
func (s *NotificationService) MarkAsRead(id int64, userID int64) error {
	return s.notificationDAO.MarkAsRead(id, userID)
}

// MarkAllAsRead 全部标记已读
func (s *NotificationService) MarkAllAsRead(userID int64) error {
	return s.notificationDAO.MarkAllAsRead(userID)
}

// Delete 删除通知
func (s *NotificationService) Delete(id int64, userID int64) error {
	return s.notificationDAO.Delete(id, userID)
}

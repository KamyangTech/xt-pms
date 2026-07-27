package service

import (
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
)

type NotificationConfigService struct {
	configDAO *dao.NotificationConfigDAO
	userDAO   *dao.SysUserDAO
}

func NewNotificationConfigService() *NotificationConfigService {
	return &NotificationConfigService{
		configDAO: dao.NewNotificationConfigDAO(),
		userDAO:   dao.NewSysUserDAO(),
	}
}

// Create 创建配置
func (s *NotificationConfigService) Create(config *entity.SysNotificationConfig) error {
	return s.configDAO.Create(config)
}

// Update 更新配置
func (s *NotificationConfigService) Update(config *entity.SysNotificationConfig) error {
	return s.configDAO.Update(config)
}

// Delete 删除配置
func (s *NotificationConfigService) Delete(id int64) error {
	return s.configDAO.Delete(id)
}

// GetByID 获取配置详情
func (s *NotificationConfigService) GetByID(id int64) (*entity.SysNotificationConfig, error) {
	return s.configDAO.FindByID(id)
}

// GetPage 分页查询
func (s *NotificationConfigService) GetPage(page, pageSize int) (*utils.PageResult, error) {
	return s.configDAO.FindAll(page, pageSize)
}

// GetRecipientIDs 根据通知类型获取所有应接收的用户 ID（去重）
func (s *NotificationConfigService) GetRecipientIDs(notifType string) ([]int64, error) {
	rules, err := s.configDAO.FindMatchedRules(notifType)
	if err != nil {
		return nil, err
	}

	// 去重
	seen := make(map[int64]bool)
	var ids []int64

	for _, rule := range rules {
		if rule.TargetType == "user" {
			if !seen[rule.TargetID] {
				seen[rule.TargetID] = true
				ids = append(ids, rule.TargetID)
			}
		} else if rule.TargetType == "role" {
			userIDs, err := s.userDAO.GetUserIDsByRoleID(rule.TargetID)
			if err != nil {
				continue
			}
			for _, uid := range userIDs {
				if !seen[uid] {
					seen[uid] = true
					ids = append(ids, uid)
				}
			}
		}
	}

	return ids, nil
}

package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type NotificationDAO struct{}

// NewNotificationDAO 创建通知 DAO 实例
func NewNotificationDAO() *NotificationDAO {
	return &NotificationDAO{}
}

// Create 创建通知
func (d *NotificationDAO) Create(notification *entity.SysNotification) error {
	return DB.Create(notification).Error
}

// BatchCreate 批量创建通知
func (d *NotificationDAO) BatchCreate(notifications []*entity.SysNotification) error {
	if len(notifications) == 0 {
		return nil
	}
	return DB.Create(notifications).Error
}

// FindByID 根据 ID 查询通知
func (d *NotificationDAO) FindByID(id int64) (*entity.SysNotification, error) {
	var n entity.SysNotification
	err := DB.First(&n, id).Error
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// FindPageByUserID 分页查询用户通知
func (d *NotificationDAO) FindPageByUserID(userID int64, page, pageSize int, isRead *int8, notifType string) (*utils.PageResult, error) {
	var list []*entity.SysNotification
	return utils.Paginate(DB, page, pageSize, &list, func(db *gorm.DB) *gorm.DB {
		query := db.Where("user_id = ?", userID).Order("create_time DESC")
		if isRead != nil {
			query = query.Where("is_read = ?", *isRead)
		}
		if notifType != "" {
			// 如果 type 以 % 结尾，使用 LIKE 前缀匹配
			// 例如 "mold:%" 匹配所有 mold: 开头的类型
			if notifType[len(notifType)-1] == '%' {
				query = query.Where("type LIKE ?", notifType)
			} else {
				query = query.Where("type = ?", notifType)
			}
		}
		return query
	})
}

// CountUnreadByUserID 查询用户未读通知数量
func (d *NotificationDAO) CountUnreadByUserID(userID int64) (int64, error) {
	var count int64
	err := DB.Model(&entity.SysNotification{}).
		Where("user_id = ? AND is_read = 0", userID).
		Count(&count).Error
	return count, err
}

// MarkAsRead 标记单条通知为已读
func (d *NotificationDAO) MarkAsRead(id int64, userID int64) error {
	return DB.Model(&entity.SysNotification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", 1).Error
}

// BatchMarkAsRead 批量标记已读
func (d *NotificationDAO) BatchMarkAsRead(ids []int64, userID int64) error {
	return DB.Model(&entity.SysNotification{}).
		Where("id IN ? AND user_id = ?", ids, userID).
		Update("is_read", 1).Error
}

// MarkAllAsRead 标记用户所有通知为已读
func (d *NotificationDAO) MarkAllAsRead(userID int64) error {
	return DB.Model(&entity.SysNotification{}).
		Where("user_id = ? AND is_read = 0", userID).
		Update("is_read", 1).Error
}

// Delete 删除通知
func (d *NotificationDAO) Delete(id int64, userID int64) error {
	return DB.Where("id = ? AND user_id = ?", id, userID).Delete(&entity.SysNotification{}).Error
}

// FindDistinctTypesByUserID 查询用户通知的去重类型列表
func (d *NotificationDAO) FindDistinctTypesByUserID(userID int64) ([]string, error) {
	var types []string
	err := DB.Model(&entity.SysNotification{}).
		Where("user_id = ?", userID).
		Select("DISTINCT type").
		Order("type ASC").
		Pluck("type", &types).Error
	return types, err
}

// FindAllDistinctTypes 查询所有通知的去重类型列表（全局，不分用户）
func (d *NotificationDAO) FindAllDistinctTypes() ([]string, error) {
	var types []string
	err := DB.Model(&entity.SysNotification{}).
		Select("DISTINCT type").
		Order("type ASC").
		Pluck("type", &types).Error
	return types, err
}

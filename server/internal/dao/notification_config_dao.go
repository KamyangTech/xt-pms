package dao

import (
	"github.com/KamyangTech/xt-pms/server/internal/model/entity"
	"github.com/KamyangTech/xt-pms/server/internal/utils"
	"gorm.io/gorm"
)

type NotificationConfigDAO struct{}

// NewNotificationConfigDAO 创建通知配置 DAO 实例
func NewNotificationConfigDAO() *NotificationConfigDAO {
	return &NotificationConfigDAO{}
}

// Create 创建配置
func (d *NotificationConfigDAO) Create(config *entity.SysNotificationConfig) error {
	return DB.Create(config).Error
}

// Update 更新配置（仅更新非零字段）
func (d *NotificationConfigDAO) Update(config *entity.SysNotificationConfig) error {
	return DB.Model(config).Updates(config).Error
}

// Delete 删除配置
func (d *NotificationConfigDAO) Delete(id int64) error {
	return DB.Delete(&entity.SysNotificationConfig{}, id).Error
}

// FindByID 根据 ID 查询
func (d *NotificationConfigDAO) FindByID(id int64) (*entity.SysNotificationConfig, error) {
	var config entity.SysNotificationConfig
	err := DB.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// FindAll 查询所有配置（支持分页）
func (d *NotificationConfigDAO) FindAll(page, pageSize int) (*utils.PageResult, error) {
	var list []*entity.SysNotificationConfig
	return utils.Paginate(DB, page, pageSize, &list, func(db *gorm.DB) *gorm.DB {
		return db.Order("create_time DESC")
	})
}

// FindAllEnabled 查询所有启用的配置
func (d *NotificationConfigDAO) FindAllEnabled() ([]*entity.SysNotificationConfig, error) {
	var list []*entity.SysNotificationConfig
	err := DB.Where("enabled = 1").Find(&list).Error
	return list, err
}

// FindMatchedRules 根据通知类型查找所有匹配的启用规则
// 匹配逻辑：通知类型 LIKE type_pattern
// 例如 type_pattern='mold:%' 可以匹配 mold:work_order_created
func (d *NotificationConfigDAO) FindMatchedRules(notifType string) ([]*entity.SysNotificationConfig, error) {
	var list []*entity.SysNotificationConfig
	err := DB.Where("enabled = 1 AND ? LIKE type_pattern", notifType).Find(&list).Error
	return list, err
}

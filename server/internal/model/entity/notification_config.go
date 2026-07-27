package entity

import "time"

// SysNotificationConfig 通知配置：定义某种通知类型发送给哪些角色/用户
type SysNotificationConfig struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	TypePattern string    `gorm:"size:100;not null;comment:通知类型匹配模式，如 mold:% 或 mold:work_order_created" json:"type_pattern"`
	TargetType  string    `gorm:"size:20;not null;comment:目标类型：role=角色 user=用户" json:"target_type"`
	TargetID    int64     `gorm:"not null;comment:目标ID（角色ID或用户ID）" json:"target_id"`
	Enabled     int8      `gorm:"not null;default:1;comment:是否启用 1=启用 0=禁用" json:"enabled"`
	CreateTime  time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime  time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (SysNotificationConfig) TableName() string {
	return "sys_notification_config"
}

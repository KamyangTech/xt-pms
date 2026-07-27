package entity

import "time"

type SysNotification struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:主键 ID" json:"id"`
	UserID     int64     `gorm:"not null;index;comment:接收用户 ID" json:"user_id"`
	Title      string    `gorm:"size:200;not null;default:'';comment:通知标题" json:"title"`
	Content    string    `gorm:"size:500;not null;default:'';comment:通知内容" json:"content"`
	Type       string    `gorm:"size:50;not null;default:'';comment:通知类型" json:"type"`
	TypeLabel  string    `gorm:"size:50;not null;default:'';comment:类型显示名称" json:"type_label"`
	Quality    string    `gorm:"size:20;not null;default:'low';comment:优先级 low/middle/high" json:"quality"`
	RefID      int64     `gorm:"default:0;comment:关联业务记录 ID" json:"ref_id"`
	RefURL     string    `gorm:"size:255;default:'';comment:关联跳转链接" json:"ref_url"`
	IsRead     int8      `gorm:"not null;default:0;comment:是否已读 0=未读 1=已读" json:"is_read"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
}

// TableName 指定表名
func (SysNotification) TableName() string {
	return "sys_notification"
}

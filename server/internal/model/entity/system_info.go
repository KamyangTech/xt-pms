package entity

import "time"

type SysSystemInfo struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SystemName        string    `gorm:"column:system_name;type:varchar(100);not null;default:'';comment:系统名称" json:"system_name"`
	SystemVersion     string    `gorm:"column:system_version;type:varchar(50);not null;default:'';comment:系统版本" json:"system_version"`
	SystemDescription string    `gorm:"column:system_description;type:varchar(500);not null;default:'';comment:系统简介" json:"system_description"`
	TechStack         string    `gorm:"column:tech_stack;type:varchar(500);not null;default:'';comment:技术栈（逗号分隔）" json:"tech_stack"`
	FaviconURL        string    `gorm:"column:favicon_url;type:varchar(500);not null;default:'';comment:Favicon图标URL" json:"favicon_url"`
	CreateTime        time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime        time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

// TableName 指定表名
func (SysSystemInfo) TableName() string {
	return "sys_system_info"
}

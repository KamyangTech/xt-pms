package entity

import "time"

type SysMenu struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:菜单ID" json:"id"`
	ParentID   int64     `gorm:"default:0;comment:父菜单ID" json:"parent_id"`
	MenuName   string    `gorm:"size:50;not null;comment:菜单名称" json:"menu_name"`
	MenuType   int8      `gorm:"default:1;comment:1=目录 2=菜单 3=按钮 4=页面元素" json:"menu_type"`
	Permission string    `gorm:"size:100;comment:权限标识" json:"permission"`
	Path       string    `gorm:"size:200;comment:路由路径" json:"path"`
	Component  string    `gorm:"size:255;comment:组件路径" json:"component"`
	Icon       string    `gorm:"size:50;comment:图标" json:"icon"`
	Sort       int       `gorm:"default:0;comment:排序" json:"sort"`
	Status     int8      `gorm:"default:1;comment:0=禁用 1=启用" json:"status"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
	// 非数据库字段
	Children []*SysMenu `gorm:"-" json:"children,omitempty"`
}

func (SysMenu) TableName() string {
	return "sys_menu"
}

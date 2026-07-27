package entity

import "time"

type ProductionStatusFlow struct {
	ID              int64     `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	FromStatusValue string    `gorm:"size:10;not null;comment:起始状态值" json:"from_status_value"`
	ToStatusValue   string    `gorm:"size:10;not null;comment:目标状态值" json:"to_status_value"`
	ActionName      string    `gorm:"size:50;not null;comment:操作名称" json:"action_name"`
	ActionCode      string    `gorm:"size:50;not null;comment:操作编码" json:"action_code"`
	IsEnabled       int8      `gorm:"default:1;comment:是否启用：1=启用，0=禁用" json:"is_enabled"`
	Sort            int       `gorm:"default:0;comment:排序" json:"sort"`
	Remark          string    `gorm:"size:255;comment:备注" json:"remark"`
	CreateTime      time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime      time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (ProductionStatusFlow) TableName() string {
	return "production_status_flow"
}

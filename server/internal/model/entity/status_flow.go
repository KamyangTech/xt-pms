package entity

import "time"

// StatusFlowConfig 状态同步配置表
// 用于配置工单、订单、模具档案三者之间的状态同步关系
type StatusFlowConfig struct {
	ID int64 `gorm:"primaryKey;autoIncrement;comment:配置ID" json:"id"`

	// 三个状态字段（至少填写一个）- 使用语义化字符串
	WorkOrderStatus *string `gorm:"size:50;comment:工单状态值（语义化字符串）" json:"work_order_status"`
	MoldOrderStatus *string `gorm:"size:50;comment:订单状态值（语义化字符串）" json:"mold_order_status"`
	MoldStatus      *string `gorm:"size:50;comment:模具档案状态值（语义化字符串）" json:"mold_status"`

	// 配置说明
	Description string `gorm:"size:500;comment:配置说明" json:"description"`
	Priority    int    `gorm:"default:0;comment:优先级（数字越小优先级越高）" json:"priority"`
	IsActive    int8   `gorm:"default:1;comment:是否启用：1=启用，0=禁用" json:"is_active"`
	StatusOrder int    `gorm:"default:0;comment:状态顺序号（用于状态比较）" json:"status_order"`

	// 审计字段
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (StatusFlowConfig) TableName() string {
	return "status_flow_config"
}

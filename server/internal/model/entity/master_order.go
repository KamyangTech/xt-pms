package entity

import "time"

type ProgressTimeline struct {
	MoldManufacturing    *string `json:"mold_manufacturing,omitempty"`     // 模具制造开始时间 (YYYY-MM-DD)
	MoldManufacturingEnd *string `json:"mold_manufacturing_end,omitempty"` // 模具制造结束时间 (YYYY-MM-DD)
	ProductProduction    *string `json:"product_production,omitempty"`     // 产品生产开始时间 (YYYY-MM-DD)
	ProductProductionEnd *string `json:"product_production_end,omitempty"` // 产品生产结束时间 (YYYY-MM-DD)
	StockIn              *string `json:"stock_in,omitempty"`               // 产品入库时间 (YYYY-MM-DD)
}

// MasterOrder 综合订单表（用于统一管理模具和生产订单）
type MasterOrder struct {
	ID               int64     `gorm:"primaryKey;autoIncrement;comment:综合订单 ID" json:"id"`
	OrderCode        string    `gorm:"size:50;not null;uniqueIndex;comment:综合订单编号（唯一）" json:"order_code"`
	CustomerID       int64     `gorm:"not null;comment:客户 ID" json:"customer_id"`
	OrderType        string    `gorm:"size:50;not null;default:'COMBINED';comment:下单类型: COMBINED=模具+产品, MOLD_ONLY=仅模具, PROD_ONLY=仅产品" json:"order_type"`
	Status           string    `gorm:"size:50;not null;default:'PENDING';comment:整体状态" json:"status"`
	TotalAmount      float64   `gorm:"default:0.00;comment:订单总金额" json:"total_amount"`
	Remark           string    `gorm:"type:text;comment:备注" json:"remark"`
	CancelReason     string    `gorm:"type:text;comment:取消原因" json:"cancel_reason"`
	CreateUserID     *int64    `gorm:"comment:创建人 ID" json:"create_user_id"`
	CreateTime       time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime       time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 首付款比例（5% 一档，默认 30%）
	DepositRatio int        `gorm:"default:30;comment:首付款比例（5%一档，默认30%）" json:"deposit_ratio"`
	QuickPlanEnd  *time.Time `gorm:"type:date;comment:订单交期（综合级别，快捷编辑用）" json:"quick_plan_end"`

	// 关联关系
	Customer         *Customer         `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"customer,omitempty"`
	MoldOrders       []MoldOrder       `gorm:"foreignKey:MasterOrderID" json:"mold_orders,omitempty"`
	ProductionOrders []ProductionOrder `gorm:"foreignKey:MasterOrderID" json:"production_orders,omitempty"`

	// 非数据库字段：统计信息（用于列表展示）
	MoldOrderCount       int     `gorm:"-" json:"mold_order_count,omitempty"`       // 模具订单数量
	ProductionOrderCount int     `gorm:"-" json:"production_order_count,omitempty"` // 生产订单数量
	PlanStartDate        *string `gorm:"-" json:"plan_start_date,omitempty"`        // 最早计划开始日期
	PlanEndDate          *string `gorm:"-" json:"plan_end_date,omitempty"`          // 最晚计划结束日期
	RemarkSummary        string  `gorm:"-" json:"remark_summary,omitempty"`         // 备注摘要（前50个字符）

	// 非数据库字段：订单进度时间线
	ProgressTimeline *ProgressTimeline `gorm:"-" json:"progress_timeline,omitempty"` // 订单进度时间线
}

func (MasterOrder) TableName() string {
	return "master_orders"
}

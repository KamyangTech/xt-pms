package entity

import "time"

type ExternalMoldArchive struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;comment:档案ID" json:"id"`
	MoldCode     string     `gorm:"size:100;not null;uniqueIndex;comment:外来模具编号（唯一）" json:"mold_code"`
	ProductName  string     `gorm:"size:200;not null;comment:产品名称" json:"product_name"`
	CustomerID   *int64     `gorm:"comment:客户ID（可选）" json:"customer_id"`
	CustomerName *string    `gorm:"size:100;comment:客户名称（冗余字段，便于查询）" json:"customer_name"`
	FirstOrderID *int64     `gorm:"comment:首次使用的订单ID" json:"first_order_id"`
	FirstUseDate *time.Time `gorm:"comment:首次使用日期" json:"first_use_date"`
	LastOrderID  *int64     `gorm:"comment:最近使用的订单ID" json:"last_order_id"`
	LastUseDate  *time.Time `gorm:"comment:最近使用日期" json:"last_use_date"`
	UseCount     int        `gorm:"default:0;comment:使用次数" json:"use_count"`
	Status       int8       `gorm:"default:1;comment:状态：1=启用，0=禁用" json:"status"`
	Remark       string     `gorm:"size:500;comment:备注" json:"remark"`
	CreateTime   time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime   time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (ExternalMoldArchive) TableName() string {
	return "external_mold_archive"
}

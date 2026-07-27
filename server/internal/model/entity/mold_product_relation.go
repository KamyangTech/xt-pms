package entity

import "time"

// MoldProductRelation 模具产品关联表
// 用于记录模具与产品的生产关系，支持一套模具对应多个产品
type MoldProductRelation struct {
	ID                    int64     `gorm:"primaryKey;autoIncrement;comment:关联ID" json:"id"`
	MoldID                int64     `gorm:"not null;comment:模具ID (mold.id)" json:"mold_id"`
	ProductID             *int64    `gorm:"comment:产品ID (product.id)，可选" json:"product_id"`
	ProductionOrderItemID *int64    `gorm:"comment:生产订单明细ID (production_order_item.id)" json:"production_order_item_id"`
	ProductName           string    `gorm:"size:200;comment:产品名称快照（生产时填写）" json:"product_name"`
	ProductImageURL       string    `gorm:"size:500;comment:产品图片URL（生产时上传）" json:"product_image_url"`
	PlanQty               int       `gorm:"default:0;comment:计划生产数量" json:"plan_qty"`
	RelationType          string    `gorm:"size:50;default:'PRODUCTION';comment:关联类型: PRODUCTION=生产关联" json:"relation_type"`
	Remark                string    `gorm:"size:500;comment:备注" json:"remark"`
	CreateTime            time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime            time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联对象
	Mold                *Mold                `gorm:"foreignKey:MoldID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"mold,omitempty"`
	Product             *Product             `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"product,omitempty"`
	ProductionOrderItem *ProductionOrderItem `gorm:"foreignKey:ProductionOrderItemID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"production_order_item,omitempty"`
}

func (MoldProductRelation) TableName() string {
	return "mold_product_relation"
}

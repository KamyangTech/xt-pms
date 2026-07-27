package entity

import "time"

type Product struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:产品 ID" json:"id"`
	ProCode    string    `gorm:"size:50;not null;uniqueIndex;comment:产品编码（唯一）" json:"pro_code"`
	ProName    string    `gorm:"size:100;not null;uniqueIndex;comment:产品名称（唯一）" json:"pro_name"`
	Spec       string    `gorm:"size:255;comment:产品规格" json:"spec"`
	SpecDetail string    `gorm:"type:text;comment:产品详细规格参数(JSON格式)" json:"spec_detail"`
	MoldID     *int64    `gorm:"comment:默认使用模具 ID（mold.id）" json:"mold_id"`
	Unit       string    `gorm:"size:20;comment:产品单位" json:"unit"`
	Price      float64   `gorm:"default:0.00;comment:产品销售单价" json:"price"`
	Weight     float64   `gorm:"default:0.00;comment:产品重量" json:"weight"`
	Status     int8      `gorm:"default:1;comment:产品状态：0=禁用，1=启用" json:"status"`
	Remark     string    `gorm:"size:255;comment:备注" json:"remark"`
	ImageURL   string    `gorm:"size:500;comment:产品图片URL（手动上传）" json:"image_url"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (Product) TableName() string {
	return "product"
}

// ProductBOM 产品 BOM 表
type ProductBOM struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:BOM ID" json:"id"`
	ProductID  int64     `gorm:"not null;comment:关联产品 ID（product.id）" json:"product_id"`
	MaterialID int64     `gorm:"not null;comment:关联物料 ID（material.id）" json:"material_id"`
	Qty        float64   `gorm:"default:1.00;comment:单台产品物料用量" json:"qty"`
	LossRate   float64   `gorm:"default:0.00;comment:物料损耗率（%）" json:"loss_rate"`
	UnitPrice  float64   `gorm:"default:0.00;comment:物料单价" json:"unit_price"`
	TotalPrice float64   `gorm:"default:0.00;comment:总价" json:"total_price"`
	Remark     string    `gorm:"size:255;comment:备注" json:"remark"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (ProductBOM) TableName() string {
	return "product_bom"
}

// Stock 库存表（结余粒度：物料+仓库；按产品耗用在出库/领料流水体现，见 product_id 备注）
type Stock struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:库存 ID" json:"id"`
	MaterialID int64     `gorm:"not null;uniqueIndex:uk_mat_wh;comment:关联物料 ID（material.id）" json:"material_id"`
	ProductID  *int64    `gorm:"comment:可选备注：追溯/展示用，不参与结余唯一键" json:"product_id"`
	WhID       int64     `gorm:"not null;uniqueIndex:uk_mat_wh;comment:关联仓库 ID（warehouse.id）" json:"wh_id"`
	StockQty   float64   `gorm:"default:0.00;comment:当前库存数量" json:"stock_qty"`
	SafeStock  float64   `gorm:"default:0.00;comment:安全库存数量" json:"safe_stock"`
	Remark     string    `gorm:"size:255;comment:备注" json:"remark"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Product *Product `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"product,omitempty"`
}

func (Stock) TableName() string {
	return "stock"
}

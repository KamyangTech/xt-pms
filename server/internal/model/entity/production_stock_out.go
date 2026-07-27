package entity

import (
	"time"
)

type ProductionStockOut struct {
	ID               int64      `gorm:"primaryKey;autoIncrement;comment:出库ID" json:"id"`
	StockCode        string     `gorm:"size:50;not null;uniqueIndex;comment:出库单号(唯一)" json:"stock_code"`
	OrderID          *int64     `gorm:"comment:订单ID(直接出库时为空)" json:"order_id"`
	WarehouseID      int64      `gorm:"not null;comment:仓库ID" json:"warehouse_id"`
	MoldID           *int64     `gorm:"comment:模具ID(自有模具)" json:"mold_id"`
	ExternalMoldCode *string    `gorm:"size:100;comment:外来模具编号(外来模具)" json:"external_mold_code"`
	ProductID        *int64     `gorm:"comment:产品ID（product.id），用于区分同一模具不同产品的出库" json:"product_id"`
	ProductName      string     `gorm:"size:200;comment:产品名称(出库时的产品名称)" json:"product_name"`
	ProductImageURL  *string    `gorm:"-" json:"product_image_url"` // 关联查询字段，不存储到数据库
	CustomerName     *string    `gorm:"size:200;comment:接收客户名称" json:"customer_name"`
	Quantity         int        `gorm:"not null;comment:出库数量" json:"quantity"`
	BatchNo          string     `gorm:"size:50;comment:批次号" json:"batch_no"`
	StockOutDate     *time.Time `gorm:"comment:出库日期" json:"stock_out_date"`
	OutType          string     `gorm:"size:20;default:normal;comment:出库类型(normal=正常发货, other=其他原因出库)" json:"out_type"`
	OutReason        *string    `gorm:"size:50;comment:出库原因(expired=过期丢弃, damaged=损坏, unqualified=不达标, other=其他原因)" json:"out_reason"`
	OutReasonRemark  *string    `gorm:"size:500;comment:出库原因备注" json:"out_reason_remark"`
	DeliveryDate     *time.Time `gorm:"comment:发货日期" json:"delivery_date"`
	DeliveryMethod   *string    `gorm:"size:50;comment:发货方式(字典值)" json:"delivery_method"`
	ExpressCompany   *string    `gorm:"size:100;comment:快递公司(字典值)" json:"express_company"`
	ExpressNo        *string    `gorm:"size:100;comment:快递单号" json:"express_no"`
	Photos           *string    `gorm:"type:json;comment:发货照片(JSON数组)" json:"photos"`
	TotalWeight      *float64   `gorm:"type:decimal(10,2);comment:总重量(kg)" json:"total_weight"`
	DriverPhone      *string    `gorm:"size:50;comment:司机电话(货拉拉用)" json:"driver_phone"`
	DriverPlate      *string    `gorm:"size:50;comment:车牌号(货拉拉用)" json:"driver_plate"`
	OperatorID       *int64     `gorm:"comment:操作人ID" json:"operator_id"`
	Remark           string     `gorm:"type:text;comment:备注" json:"remark"`
	CreateTime       time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime       time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Order     *ProductionOrder `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"order,omitempty"`
	Warehouse *Warehouse       `gorm:"foreignKey:WarehouseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"warehouse,omitempty"`
	Mold      *Mold            `gorm:"foreignKey:MoldID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"mold,omitempty"`
	Product   *Product         `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"product,omitempty"`
}

func (ProductionStockOut) TableName() string {
	return "production_stock_out"
}

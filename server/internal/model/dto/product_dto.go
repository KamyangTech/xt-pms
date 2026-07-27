package dto

import "time"

// ProductDTO 产品档案请求 DTO
type ProductDTO struct {
	ProCode    string  `json:"pro_code" binding:"required"`
	ProName    string  `json:"pro_name" binding:"required"`
	Spec       string  `json:"spec"`
	SpecDetail string  `json:"spec_detail"`
	MoldID     *int64  `json:"mold_id"`
	Unit       string  `json:"unit"`
	Price      float64 `json:"price"`
	Weight     float64 `json:"weight"`
	Status     int8    `json:"status"`
	Remark     string  `json:"remark"`
	ImageURL   string  `json:"image_url"`
}

// ProductUpdateDTO 产品更新请求 DTO
type ProductUpdateDTO struct {
	ID         int64   `json:"id" binding:"required"`
	ProCode    string  `json:"pro_code"`
	ProName    string  `json:"pro_name"`
	Spec       string  `json:"spec"`
	SpecDetail string  `json:"spec_detail"`
	MoldID     *int64  `json:"mold_id"`
	Unit       string  `json:"unit"`
	Price      float64 `json:"price"`
	Weight     float64 `json:"weight"`
	Status     int8    `json:"status"`
	Remark     string  `json:"remark"`
	ImageURL   string  `json:"image_url"`
}

// ProductBOMDTO 产品 BOM 请求 DTO
type ProductBOMDTO struct {
	ProductID  int64   `json:"product_id" binding:"required"`
	MaterialID int64   `json:"material_id" binding:"required"`
	Qty        float64 `json:"qty"`
	LossRate   float64 `json:"loss_rate"`
	UnitPrice  float64 `json:"unit_price"`
	TotalPrice float64 `json:"total_price"`
	Remark     string  `json:"remark"`
}

// StockDTO 库存请求 DTO
type StockDTO struct {
	MaterialID int64   `json:"material_id" binding:"required"`
	WhID       int64   `json:"wh_id" binding:"required"`
	StockQty   float64 `json:"stock_qty"`
	SafeStock  float64 `json:"safe_stock"`
	Remark     string  `json:"remark"`
}

// ProductWithStockVO 产品带库存信息响应 VO
type ProductWithStockVO struct {
	ID         int64     `json:"id"`
	ProCode    string    `json:"pro_code"`
	ProName    string    `json:"pro_name"`
	Spec       string    `json:"spec"`
	SpecDetail string    `json:"spec_detail"`
	MoldID     *int64    `json:"mold_id"`
	MoldCode   string    `json:"mold_code"`
	Unit       string    `json:"unit"`
	Price      float64   `json:"price"`
	Weight     float64   `json:"weight"`
	Status     int8      `json:"status"`
	Remark     string    `json:"remark"`
	ImageURL   string    `json:"image_url"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
	// 库存信息
	TotalIn      int    `json:"total_in"`      // 总入库
	TotalOut     int    `json:"total_out"`     // 总出库
	CurrentStock int    `json:"current_stock"` // 当前库存
	// 产品图片（从生产记录回退获取，优先使用 ImageURL）
	ProductImageURL string `json:"product_image_url"`
}

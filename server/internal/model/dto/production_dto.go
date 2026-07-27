package dto

import (
	"encoding/json"
	"time"
)

// ProductionOrderDTO 生产订单DTO
type ProductionOrderDTO struct {
	ID               int64           `json:"id"`
	OrderCode        string          `json:"order_code"`
	CustomerID       int64           `json:"customer_id" binding:"required"`
	ProductName      string          `json:"product_name"`
	ProductionType   *string         `json:"production_type"`
	MoldSourceType   *string         `json:"mold_source_type"`
	MoldID           *int64          `json:"mold_id"`
	ExternalMoldCode *string         `json:"external_mold_code"`
	ExternalMoldInfo json.RawMessage `json:"external_mold_info"`
	PlanStartDate    *string         `json:"plan_start_date"`
	PlanEndDate      *string         `json:"plan_end_date"`

	// 生产工艺信息
	ProcessType   *string `json:"process_type"`
	ProcessRemark string  `json:"process_remark"`

	// 原料信息
	MaterialType  *string  `json:"material_type"`
	ProductWeight *float64 `json:"product_weight"`
	HasMetalParts bool     `json:"has_metal_parts"`

	// 后处理流程
	ManualProcess   string `json:"manual_process"`
	AssemblyProcess string `json:"assembly_process"`

	DeviceID       *string                  `json:"device_id"`
	WorkerID       *string                  `json:"worker_id"`
	ProcessorID    *int64                   `json:"processor_id"`
	OutsourcePrice float64                  `json:"outsource_price"`
	MaterialCost   float64                  `json:"material_cost"`
	ProcessingCost float64                  `json:"processing_cost"`
	UnitPrice      float64                  `json:"unit_price"`
	Priority       string                   `json:"priority"`
	Remark         string                   `json:"remark"`
	Items          []ProductionOrderItemDTO `json:"items"`
}

// ProductionOrderItemDTO 生产订单明细DTO
type ProductionOrderItemDTO struct {
	ID                int64   `json:"id"`
	ProductID         *int64  `json:"product_id"` // 可选，手动输入产品名称时为空
	ProductName       *string `json:"product_name"`
	ProductImageURL   *string `json:"product_image_url"`
	MoldID            *int64  `json:"mold_id"`            // 单模具（兼容旧数据）
	MoldIDs           []int64 `json:"mold_ids"`           // 多模具列表
	ExternalMoldCode  *string `json:"external_mold_code"` // 外来模具编号
	DeviceID          *int64  `json:"device_id"`
	PlasticMaterialID *int64  `json:"plastic_material_id"` // 塑胶原料字典值
	Weight            *float64 `json:"weight"`            // 产品重量（克）
	PlanQty           int     `json:"plan_qty" binding:"gt=0"`
	Sort              int     `json:"sort"`
	Remark            string  `json:"remark"`
	DependencyStatus  string  `json:"dependency_status"`   // ✨ 新增：明细级别的依赖状态
	// ✨ 款项配置字段（金额计算）
	UnitPrice        float64 `json:"unit_price"`          // 单价（元）
	PackagingFee     float64 `json:"packaging_fee"`       // 包装费（元）
	FreightFee       float64 `json:"freight_fee"`         // 运费（元）
	OutsourcePrice   float64 `json:"outsource_price"`     // 外发单价（元）
}

// ProductionOrderQueryDTO 生产订单查询DTO
type ProductionOrderQueryDTO struct {
	Page             int     `form:"page" binding:"min=1"`
	PageSize         int     `form:"page_size" binding:"min=1,max=100"`
	Keyword          string  `form:"keyword"`                     // 关键词搜索（订单号、客户名称、产品名称）
	OrderCode        string  `form:"order_code"`
	CustomerID       *int64  `form:"customer_id"`
	ProductName      string  `form:"product_name"`
	ProductionType   *string `form:"production_type"`
	MoldSourceType   *string `form:"mold_source_type"`
	Status           *string `form:"status"`
	ItemStatus       *string `form:"item_status"`                // 按明细项状态过滤（报工页使用，与卡片展示一致）
	Priority         *string `form:"priority"`
	PlanEndDateStart *string `form:"plan_end_date_start"`
	PlanEndDateEnd   *string `form:"plan_end_date_end"`
	CreateTimeStart  *string `form:"create_time_start"`
	CreateTimeEnd    *string `form:"create_time_end"`
	DueSoon          *string `form:"due_soon"`
}

// ProductionScheduleDTO 生产排程DTO
type ProductionScheduleDTO struct {
	ID           int64     `json:"id"`
	OrderID      int64     `json:"order_id" binding:"required"`
	ScheduleDate time.Time `json:"schedule_date" binding:"required"`
	TimeSlot     *string   `json:"time_slot"`
	DeviceID     *string   `json:"device_id"`
	WorkerID     *string   `json:"worker_id"`
	PlanQty      int       `json:"plan_qty" binding:"gt=0"`
	Status       string    `json:"status"`
	Remark       string    `json:"remark"`
}

// ProductionReportDTO 生产报工DTO
type ProductionReportDTO struct {
	ID           int64   `json:"id"`
	OrderID      int64   `json:"order_id" binding:"required"`
	ItemID       *int64  `json:"item_id"` // ✨ 新增：订单明细项ID（支持按明细报工）
	ScheduleID   *int64  `json:"schedule_id"`
	ReportDate   string  `json:"report_date" binding:"required"`
	Shift        *string `json:"shift"`
	OkQty        int     `json:"ok_qty" binding:"gte=0"`
	BadQty       int     `json:"bad_qty" binding:"gte=0"`
	ScrapQty     int     `json:"scrap_qty" binding:"gte=0"`
	DeviceID     *string `json:"device_id"`
	WorkerID     *string `json:"worker_id"`
	WorkHours    float64 `json:"work_hours"`
	DefectReason string  `json:"defect_reason"`
	Remark       string  `json:"remark"`
}

// OutsourcingRecordDTO 外发记录DTO
type OutsourcingRecordDTO struct {
	ID                 int64           `json:"id"`
	RecordCode         string          `json:"record_code"`
	OrderID            int64           `json:"order_id"`
	WorkOrderID        *int64          `json:"work_order_id"`
	OutsourceType      string          `json:"outsource_type"` // PRODUCTION=生产外发, MOLD=模具外发
	ProcessorID        int64           `json:"processor_id" binding:"required"`
	MoldID             *int64          `json:"mold_id"`
	ExternalMoldInfo   json.RawMessage `json:"external_mold_info"`
	OutsourceDate      *FlexibleTime  `json:"outsource_date"`
	ExpectedReturnDate *FlexibleTime  `json:"expected_return_date"`
	ActualReturnDate   *FlexibleTime  `json:"actual_return_date"`
	Quantity           int             `json:"quantity" binding:"gt=0"`
	UnitPrice          float64         `json:"unit_price"`
	LogisticsCompany   string          `json:"logistics_company"`
	TrackingNo         string          `json:"tracking_no"`
	DeliveryAddress    string          `json:"delivery_address"`
	ReceiverName       string          `json:"receiver_name"`
	ReceiverPhone      string          `json:"receiver_phone"`
	Status             string          `json:"status"`
	Remark             string          `json:"remark"`
}

// OutsourcingReportDTO 外发报工DTO
type OutsourcingReportDTO struct {
	ID            int64           `json:"id"`
	OutsourcingID int64           `json:"outsourcing_id" binding:"required"`
	ReportDate    *FlexibleTime  `json:"report_date"`
	OkQty         int             `json:"ok_qty" binding:"gte=0"`
	BadQty        int             `json:"bad_qty" binding:"gte=0"`
	DefectReason  string          `json:"defect_reason"`
	Photos        json.RawMessage `json:"photos"`
	Remark        string          `json:"remark"`
}

// OutsourcingPublicReportDTO 外发公开报工DTO（加工商通过分享链接提交，无需outsourcing_id）
type OutsourcingPublicReportDTO struct {
	ReportDate   *FlexibleTime  `json:"report_date"`
	OkQty        *int            `json:"ok_qty" binding:"omitempty,gte=0"`
	BadQty       *int            `json:"bad_qty" binding:"omitempty,gte=0"`
	DefectReason string          `json:"defect_reason"`
	Photos       json.RawMessage `json:"photos"`
	Remark       string          `json:"remark"`
}

// ProductionStockInDTO 生产入库DTO
type ProductionStockInDTO struct {
	ID               int64      `json:"id"`
	StockCode        string     `json:"stock_code"`
	OrderID          int64      `json:"order_id" binding:"required"`
	ItemID           *int64     `json:"item_id"` // ✨ 新增：明细项ID（支持按明细项入库）
	WarehouseID      int64      `json:"warehouse_id" binding:"required"`
	MoldID           *int64     `json:"mold_id"`
	ExternalMoldCode *string    `json:"external_mold_code"`
	ProductID        *int64     `json:"product_id"`   // 产品ID，用于区分同一模具不同产品的库存
	ProductName      string     `json:"product_name"` // 入库时的产品名称
	Quantity         int        `json:"quantity" binding:"gt=0"`
	BatchNo          string     `json:"batch_no"`
	StockInDate      *FlexibleTime `json:"stock_in_date"`
	OperatorID       *int64     `json:"operator_id"`
	Remark           string     `json:"remark"`
}

type ExternalMoldInfo struct {
	MoldCode     string `json:"mold_code"`
	MoldName     string `json:"mold_name"`
	CustomerName string `json:"customer_name"`
	CavityCount  *int   `json:"cavity_count"`
	Spec         string `json:"spec"`
	Material     string `json:"material"`
	Weight       string `json:"weight"`
	Remark       string `json:"remark"`
}

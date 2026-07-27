package entity

import (
	"encoding/json"
	"time"
)

type ProductionOrder struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;comment:生产订单ID" json:"id"`
	OrderCode   string `gorm:"size:50;not null;uniqueIndex;comment:订单编号（唯一）" json:"order_code"`
	CustomerID  int64  `gorm:"not null;comment:客户ID（customer.id）" json:"customer_id"`
	ProductName string `gorm:"size:200;not null;comment:产品名称（手动输入）" json:"product_name"`

	// 生产配置（字典：production-type, mold-source-type）
	ProductionType   string          `gorm:"size:50;not null;default:IN_HOUSE;comment:生产方式" json:"production_type"`
	MoldSourceType   string          `gorm:"size:50;not null;default:SELF_MADE;comment:模具来源" json:"mold_source_type"`
	MoldID           *int64          `gorm:"comment:模具ID（mold.id），自有模具时必填" json:"mold_id"`
	ExternalMoldCode *string         `gorm:"size:100;comment:外来模具编号（外来模具时填写）" json:"external_mold_code"`
	ExternalMoldInfo json.RawMessage `gorm:"type:json;comment:外来模具信息JSON（模具来源=外来时必填）" json:"external_mold_info"`

	// 生产计划
	PlanQty         int        `gorm:"not null;default:0;comment:计划生产数量" json:"plan_qty"`
	CompletedQty    int        `gorm:"default:0;comment:已完成数量" json:"completed_qty"`
	OkQty           int        `gorm:"default:0;comment:良品数量" json:"ok_qty"`
	BadQty          int        `gorm:"default:0;comment:不良数量" json:"bad_qty"`
	MoldShots       int        `gorm:"default:0;comment:模具使用次数(模次)" json:"mold_shots"`
	PlanStartDate   *time.Time `gorm:"type:date;comment:计划开始日期" json:"plan_start_date"`
	PlanEndDate     *time.Time `gorm:"type:date;comment:计划交期" json:"plan_end_date"`
	ActualStartDate *time.Time `gorm:"type:date;comment:实际开始日期" json:"actual_start_date"`
	ActualEndDate   *time.Time `gorm:"type:date;comment:实际完成日期" json:"actual_end_date"`

	// 生产工艺信息
	ProcessType   *string `gorm:"type:varchar(50);comment:生产工艺类型：injection=注塑, blow_molding=吹塑, stamping=冲压" json:"process_type"`
	ProcessRemark string  `gorm:"type:text;comment:工艺备注" json:"process_remark"`

	// 原料信息
	MaterialType  *string  `gorm:"type:varchar(100);comment:原料类型：PP, PE, TPE等" json:"material_type"`
	ProductWeight *float64 `gorm:"type:decimal(10,2);comment:产品重量(克)" json:"product_weight"`
	HasMetalParts bool     `gorm:"default:false;comment:是否包含金属配件" json:"has_metal_parts"`

	// 后处理流程
	ManualProcess   string `gorm:"type:text;comment:手工处理流程" json:"manual_process"`
	AssemblyProcess string `gorm:"type:text;comment:组装流程" json:"assembly_process"`

	// 资源配置（内部生产）
	DeviceID *string `gorm:"size:50;comment:生产设备字典值" json:"device_id"`
	WorkerID *string `gorm:"size:50;comment:负责人字典值" json:"worker_id"`

	// 外发配置（外发生产）
	ProcessorID          *int64  `gorm:"comment:加工商ID（processor.id）" json:"processor_id"`
	OutsourcePrice       float64 `gorm:"default:0.00;comment:外发单价" json:"outsource_price"`
	OutsourceTotalAmount float64 `gorm:"default:0.00;comment:外发总金额" json:"outsource_total_amount"`

	// 成本核算
	MaterialCost   float64 `gorm:"default:0.00;comment:材料成本" json:"material_cost"`
	ProcessingCost float64 `gorm:"default:0.00;comment:加工成本" json:"processing_cost"`
	TotalCost      float64 `gorm:"default:0.00;comment:总成本" json:"total_cost"`
	UnitPrice      float64 `gorm:"default:0.00;comment:产品单价" json:"unit_price"`
	TotalAmount    float64 `gorm:"default:0.00;comment:订单总金额" json:"total_amount"`

	// 状态管理（字典：production-order-status, order-priority）
	Status   string `gorm:"size:50;not null;default:PENDING_SCHEDULE;comment:订单状态" json:"status"`
	Priority string `gorm:"size:50;default:NORMAL;comment:优先级（字典：order-priority）" json:"priority"`

	// 其他
	Remark           string    `gorm:"type:text;comment:备注" json:"remark"`
	CreateUserID     *int64    `gorm:"comment:创建人ID" json:"create_user_id"`
	CreateTime       time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime       time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
	MasterOrderID    *int64    `gorm:"comment:关联综合订单 ID (master_orders.id)" json:"master_order_id"`
	MasterCancelled  int8      `gorm:"default:0;comment:父综合订单已取消标记：0=否，1=是" json:"master_cancelled"`
	DependencyStatus string    `gorm:"size:50;default:'NONE';comment:依赖状态: NONE=无依赖, WAITING_MOLD=等待模具, READY=就绪" json:"dependency_status"`

	// 关联关系
	MasterOrder *MasterOrder          `gorm:"foreignKey:MasterOrderID" json:"master_order,omitempty"`
	Customer    *Customer             `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"customer,omitempty"`
	Mold        *Mold                 `gorm:"foreignKey:MoldID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"mold,omitempty"`
	Processor   *Processor            `gorm:"foreignKey:ProcessorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"processor,omitempty"`
	Items       []ProductionOrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`

	// 非数据库字段 - 查询时计算填充
	ShippedQty          int      `gorm:"-" json:"shipped_qty"`
	AvailableActionCodes []string `gorm:"-" json:"available_action_codes"` // 可用操作编码列表 // 已出货数量（从出库记录统计）
}

func (ProductionOrder) TableName() string {
	return "production_order"
}

type ProductionOrderItem struct {
	ID                int64    `gorm:"primaryKey;autoIncrement;comment:明细ID" json:"id"`
	OrderID           int64    `gorm:"not null;comment:订单ID" json:"order_id"`
	ProductID         *int64   `gorm:"comment:产品ID（可选，手动输入产品名称时为空）" json:"product_id"`
	ProductName       string   `gorm:"size:200;not null;comment:产品名称（独立于模具，如：手机壳A）" json:"product_name"`
	ProductImageURL   string   `gorm:"size:500;comment:产品图片URL" json:"product_image_url"`
	MoldID            *int64   `gorm:"comment:模具ID（自有模具）" json:"mold_id"`
	MoldIDs           []int64  `gorm:"type:json;serializer:json;comment:关联模具ID列表（JSON数组，支持多模具）" json:"mold_ids"`
	ExternalMoldCode  *string  `gorm:"size:100;comment:外来模具编号（外来模具时填写）" json:"external_mold_code"`
	DeviceID          *int64   `gorm:"comment:设备ID" json:"device_id"`
	PlasticMaterialID *int64   `gorm:"comment:塑胶原料字典值（关联sys_dict_data.id）" json:"plastic_material_id"`
	Weight            *float64 `gorm:"type:decimal(10,2);comment:产品重量（克）" json:"weight"`
	PlanQty           int      `gorm:"not null;default:0;comment:计划数量" json:"plan_qty"`
	ProcessRemark     string   `gorm:"type:text;comment:工艺参数/备注" json:"process_remark"` // ✨ 添加工艺参数字段
	CompletedQty      int      `gorm:"default:0;comment:完成数量" json:"completed_qty"`
	// ✨ 新增：明细项独立状态
	Status               string     `gorm:"size:50;default:'PRODUCING';comment:明细项状态：PENDING_SCHEDULE=待排产, PRODUCING=生产中, PAUSED=已暂停, STOCKED_IN=已入库, COMPLETED=已完成" json:"status"`
	ActualCompletionTime *time.Time `gorm:"comment:实际完成时间（该明细项完成生产的时刻）" json:"actual_completion_time"`
	OkQty                int        `gorm:"default:0;comment:良品数" json:"ok_qty"`
	BadQty               int        `gorm:"default:0;comment:不良数" json:"bad_qty"`
	Sort                 int        `gorm:"default:0;comment:排序" json:"sort"`
	Remark               string     `gorm:"size:500;comment:备注" json:"remark"`
	DependencyStatus     string    `gorm:"size:50;default:'NONE';comment:依赖状态: NONE=无依赖, WAITING_MOLD=等待模具, READY=就绪" json:"dependency_status"` // ✨ 新增：明细级别的依赖状态

	// ✨ 款项配置字段（金额计算）
	UnitPrice          float64 `gorm:"default:0.00;comment:单价（元）" json:"unit_price"`
	PackagingFee       float64 `gorm:"default:0.00;comment:包装费（元）" json:"packaging_fee"`
	FreightFee         float64 `gorm:"default:0.00;comment:运费（元）" json:"freight_fee"`
	OutsourcePrice     float64 `gorm:"default:0.00;comment:外发单价（元）" json:"outsource_price"`

	CreateTime           time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime           time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Order   *ProductionOrder `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"order,omitempty"`
	Product *Product         `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"product,omitempty"`
	Mold    *Mold            `gorm:"foreignKey:MoldID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"mold,omitempty"`

	// 多模具关联（非数据库字段）
	Molds []*Mold `gorm:"-" json:"molds,omitempty"`
}

func (ProductionOrderItem) TableName() string {
	return "production_order_item"
}

type ProductionSchedule struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;comment:排程ID" json:"id"`
	OrderID      int64     `gorm:"not null;comment:订单ID" json:"order_id"`
	ScheduleDate time.Time `gorm:"type:date;not null;comment:排产日期" json:"schedule_date"`
	TimeSlot     *string   `gorm:"size:50;comment:时间段（字典：time-slot）" json:"time_slot"`
	DeviceID     *string   `gorm:"size:50;comment:分配设备" json:"device_id"`
	WorkerID     *string   `gorm:"size:50;comment:分配人员" json:"worker_id"`
	PlanQty      int       `gorm:"default:0;comment:计划产量" json:"plan_qty"`
	Status       string    `gorm:"size:50;default:PENDING;comment:状态（字典：schedule-status）" json:"status"`
	Remark       string    `gorm:"type:text;comment:备注" json:"remark"`
	CreateTime   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime   time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Order  *ProductionOrder `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"order,omitempty"`
}

func (ProductionSchedule) TableName() string {
	return "production_schedule"
}

type ProductionReport struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;comment:报工ID" json:"id"`
	OrderID      int64     `gorm:"not null;comment:订单ID" json:"order_id"`
	ItemID       *int64    `gorm:"comment:订单明细项ID（支持按明细报工）" json:"item_id"` // ✨ 新增
	ScheduleID   *int64    `gorm:"comment:排程ID" json:"schedule_id"`
	ReportDate   time.Time `gorm:"not null;comment:报工时间" json:"report_date"`
	Shift        *string   `gorm:"size:50;comment:班次（字典：shift）" json:"shift"`
	OkQty        int       `gorm:"default:0;comment:良品数量" json:"ok_qty"`
	BadQty       int       `gorm:"default:0;comment:不良数量" json:"bad_qty"`
	ScrapQty     int       `gorm:"default:0;comment:报废数量" json:"scrap_qty"`
	DeviceID     *string   `gorm:"size:50;comment:使用设备" json:"device_id"`
	WorkerID     *string   `gorm:"size:50;comment:操作工" json:"worker_id"`
	WorkHours    float64   `gorm:"default:0.00;comment:工时（小时）" json:"work_hours"`
	DefectReason string    `gorm:"type:text;comment:不良原因" json:"defect_reason"`
	Remark       string    `gorm:"type:text;comment:备注" json:"remark"`
	CreateUserID *int64    `gorm:"comment:报工人ID" json:"create_user_id"`
	CreateTime   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime   time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Order    *ProductionOrder    `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"order,omitempty"`
	Schedule *ProductionSchedule `gorm:"foreignKey:ScheduleID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"schedule,omitempty"`
}

func (ProductionReport) TableName() string {
	return "production_report"
}

// ProductionStockIn 生产入库表
type OutsourcingRecord struct {
	ID               int64           `gorm:"primaryKey;autoIncrement;comment:外发记录ID" json:"id"`
	RecordCode       string          `gorm:"size:50;not null;uniqueIndex;comment:外发单号（唯一）" json:"record_code"`
	OrderID          *int64          `gorm:"comment:关联生产订单ID(production_order.id),生产外发时填写" json:"order_id"`
	WorkOrderID      *int64          `gorm:"comment:关联模具工单ID(mold_work_order.id),模具外发时填写" json:"work_order_id"`
	OutsourceType    string          `gorm:"size:50;default:PRODUCTION;comment:外发类型（字典：outsource-type）" json:"outsource_type"`
	ProcessorID      int64           `gorm:"not null;comment:加工商ID" json:"processor_id"`
	MoldID           *int64          `gorm:"comment:外发模具ID（自有模具）" json:"mold_id"`
	ExternalMoldInfo json.RawMessage `gorm:"type:json;comment:外来模具信息" json:"external_mold_info"`

	// 外发信息
	OutsourceDate      *time.Time `gorm:"comment:外发日期" json:"outsource_date"`
	ExpectedReturnDate *time.Time `gorm:"comment:预计返回日期" json:"expected_return_date"`
	ActualReturnDate   *time.Time `gorm:"comment:实际返回日期" json:"actual_return_date"`
	Quantity           int        `gorm:"default:0;comment:外发数量" json:"quantity"`
	UnitPrice          float64    `gorm:"default:0.00;comment:单价" json:"unit_price"`
	TotalAmount        float64    `gorm:"default:0.00;comment:总金额" json:"total_amount"`
	PaidAmount         float64    `gorm:"default:0.00;comment:已付金额" json:"paid_amount"`

	// 物流信息
	LogisticsCompany string `gorm:"size:100;comment:物流公司" json:"logistics_company"`
	TrackingNo       string `gorm:"size:50;comment:物流单号" json:"tracking_no"`
	DeliveryAddress  string `gorm:"type:text;comment:交付地址" json:"delivery_address"`
	ReceiverName     string `gorm:"size:50;comment:收货人" json:"receiver_name"`
	ReceiverPhone    string `gorm:"size:20;comment:收货电话" json:"receiver_phone"`

	// 状态
	Status string `gorm:"size:50;default:PENDING;comment:状态（字典：outsourcing-status）" json:"status"`

	Remark       string    `gorm:"type:text;comment:备注" json:"remark"`
	ShareToken   *string   `gorm:"size:32;uniqueIndex;comment:分享token（8位随机字符，用于加工商免登录报工）" json:"share_token"`
	CreateUserID *int64    `gorm:"comment:创建人ID" json:"create_user_id"`
	CreateTime   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime   time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Order     *ProductionOrder `gorm:"foreignKey:OrderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"order,omitempty"`
	Processor *Processor       `gorm:"foreignKey:ProcessorID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"processor,omitempty"`
	Mold      *Mold            `gorm:"foreignKey:MoldID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"mold,omitempty"`
	WorkOrder *MoldWorkOrder   `gorm:"foreignKey:WorkOrderID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"work_order,omitempty"`
}

func (OutsourcingRecord) TableName() string {
	return "outsourcing_record"
}

type OutsourcingReport struct {
	ID            int64           `gorm:"primaryKey;autoIncrement;comment:外发报工ID" json:"id"`
	OutsourcingID int64           `gorm:"not null;comment:外发记录ID" json:"outsourcing_id"`
	ReportDate    *time.Time      `gorm:"comment:报工日期" json:"report_date"`
	OkQty         int             `gorm:"default:0;comment:良品数量" json:"ok_qty"`
	BadQty        int             `gorm:"default:0;comment:不良数量" json:"bad_qty"`
	DefectReason  string          `gorm:"type:text;comment:不良原因" json:"defect_reason"`
	Photos        json.RawMessage `gorm:"type:json;comment:现场照片URL列表" json:"photos"`
	Remark        string          `gorm:"type:text;comment:备注" json:"remark"`
	CreateTime    time.Time       `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime    time.Time       `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Outsourcing *OutsourcingRecord `gorm:"foreignKey:OutsourcingID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"outsourcing,omitempty"`
}

func (OutsourcingReport) TableName() string {
	return "outsourcing_report"
}

type ProductionStockIn struct {
	ID               int64      `gorm:"primaryKey;autoIncrement;comment:入库ID" json:"id"`
	StockCode        string     `gorm:"size:50;not null;uniqueIndex;comment:入库单号（唯一）" json:"stock_code"`
	OrderID          *int64     `gorm:"comment:订单ID（直接入库时为空）" json:"order_id"`
	WarehouseID      int64      `gorm:"not null;comment:仓库ID" json:"warehouse_id"`
	MoldID           *int64     `gorm:"comment:模具ID（自有模具）" json:"mold_id"`
	ExternalMoldCode *string    `gorm:"size:100;comment:外来模具编号（外来模具）" json:"external_mold_code"`
	ProductID        *int64     `gorm:"comment:产品ID（product.id），用于区分同一模具不同产品的入库" json:"product_id"`
	ProductName      string     `gorm:"size:200;comment:产品名称（入库时的产品名称）" json:"product_name"`
	Quantity         int        `gorm:"not null;comment:入库数量" json:"quantity"`
	BatchNo          string     `gorm:"size:50;comment:批次号" json:"batch_no"`
	StockInDate      *time.Time `gorm:"comment:入库日期" json:"stock_in_date"`
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

func (ProductionStockIn) TableName() string {
	return "production_stock_in"
}

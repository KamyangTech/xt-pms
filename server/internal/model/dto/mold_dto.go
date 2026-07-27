package dto

// MoldDTO 模具档案请求 DTO
type MoldDTO struct {
	MoldCode        string `json:"mold_code" binding:"required"`
	MoldName        string `json:"mold_name" binding:"required"`
	CustomerID      int64  `json:"customer_id" binding:"required"`
	ProductionDate  string `json:"production_date" binding:"required"`
	LifeCycleStatus string `json:"life_cycle_status" binding:"required"` // 生命周期状态：语义化字符串
	MoldType        string `json:"mold_type" binding:"required"`         // 模胚情况：NEW=全新，USED=二手
	MoldProcess     string `json:"mold_process" binding:"required"`      // 模具工艺
	MaterialType    string `json:"material_type" binding:"required"`     // 材质：STEEL=钢材，HARDENED_STEEL=加硬钢材
	CavityCount     int    `json:"cavity_count" binding:"required,gt=0"` // 模腔数（一模几腔）
	Remark          string `json:"remark"`
}

// MoldUpdateDTO 模具更新请求 DTO
type MoldUpdateDTO struct {
	ID              int64  `json:"id" binding:"required"`
	MoldCode        string `json:"mold_code"`
	MoldName        string `json:"mold_name"`
	CustomerID      int64  `json:"customer_id"`
	ProductionDate  string `json:"production_date" binding:"required"`
	LifeCycleStatus string `json:"life_cycle_status"` // 生命周期状态：语义化字符串
	MoldType        string `json:"mold_type"`
	MoldProcess     string `json:"mold_process"`
	MaterialType    string `json:"material_type"`
	CavityCount     int    `json:"cavity_count"` // 模腔数（一模几腔）
	Remark          string `json:"remark"`
}

// MoldOrderDTO 模具订单请求 DTO
type MoldOrderDTO struct {
	ID               int64   `json:"id"`
	OrderCode        string  `json:"order_code"`
	CustomerID       int64   `json:"customer_id" binding:"required"`
	OrderType        string  `json:"order_type"`
	CreateUserID     int64   `json:"create_user_id"`
	Priority         string  `json:"priority"` // 优先级（字典：order-priority）
	MoldID           *int64  `json:"mold_id"`
	PlanStartDate    *Date   `json:"plan_start_date"` // 使用自定义 Date 类型，支持 "YYYY-MM-DD" 格式
	PlanEndDate      *Date   `json:"plan_end_date"`   // 使用自定义 Date 类型，支持 "YYYY-MM-DD" 格式
	TotalAmount      float64 `json:"total_amount"`
	AdditionalAmount float64 `json:"additional_amount"`
	Status           string  `json:"status"`       // 语义化字符串状态
	DeliverTime      *Date   `json:"deliver_time"` // 使用自定义 Date 类型
	Remark           string  `json:"remark"`
	// 是否同时创建工单和模具档案（面向客户场景）
	AutoCreateWorkOrder bool               `json:"auto_create_work_order"`
	Items               []MoldOrderItemDTO `json:"items"`           // 一单多款明细
	WorkOrderInfo       *MoldWorkOrderDTO  `json:"work_order_info"` // 兼容旧版单个工单
	MoldInfo            *MoldDTO           `json:"mold_info"`       // 兼容旧版单个模具
	// ✨ 款项配置（可选）
	PaymentPlan *PaymentPlanDTO `json:"payment_plan"`
}

// PaymentPlanDTO 款项配置 DTO
type PaymentPlanDTO struct {
	Enabled bool             `json:"enabled"`
	Items   []PaymentItemDTO `json:"items"`
}

// PaymentItemDTO 款项明细 DTO
type PaymentItemDTO struct {
	Index            int     `json:"index"`
	PaymentType      string  `json:"payment_type" binding:"required"`
	AmountType       string  `json:"amount_type"` // amount=金额, ratio=比例
	AmountValue      float64 `json:"amount_value"`
	CalculatedAmount float64 `json:"calculated_amount"`
	DueDate          string  `json:"due_date"`
	SourceType       string  `json:"source_type"` // 款项来源：mold=模具款, production=生产款
}

// MoldOrderItemDTO 模具订单明细 DTO
type MoldOrderItemDTO struct {
	ID           *int64  `json:"id"`                           // 明细ID
	MoldID       *int64  `json:"mold_id"`                      // 模具ID（修模订单时必填，新模订单时为空）
	MoldCode     string  `json:"mold_code" binding:"required"` // 模具编号
	MoldName     string  `json:"mold_name" binding:"required"` // 产品名称
	Amount       float64 `json:"amount"`                       // 该模具对应金额
	PlanStart    *Date   `json:"plan_start"`                   // 工单计划开始
	PlanEnd      *Date   `json:"plan_end"`                     // 工单计划结束
	MoldType     string  `json:"mold_type"`                    // 模胚情况：NEW=全新，USED=二手
	MaterialType string  `json:"material_type"`                // 材质：STEEL=钢材，HARDENED_STEEL=加硬钢材
	MoldProcess  string  `json:"mold_process"`                 // 模具工艺
	CavityCount  int     `json:"cavity_count"`                 // 模腔数（一模几腔）
	Remark       string  `json:"remark"`                       // 备注
	// ✨ 工单相关信息（生产方式）
	WorkType       string   `json:"work_type"`       // 工单类型：IN_HOUSE=内部生产，OUTSOURCED=外发加工
	ProcessorID    *int64   `json:"processor_id"`    // 加工商ID（外发时填写）
	OutsourcePrice *float64 `json:"outsource_price"` // 外发单价
}

// MoldWorkOrderDTO 模具工单请求 DTO
type MoldWorkOrderDTO struct {
	ID          int64  `json:"id"`
	WorkCode    string `json:"work_code"` // 工单号（可选，为空时系统自动生成）
	MoldOrderID *int64 `json:"mold_order_id"`
	OrderType   string `json:"order_type" binding:"required"` // 必填：NEW_MOLD=新模，REPAIR=修模，MODIFICATION=改造
	MoldID      *int64 `json:"mold_id"`                       // 修模/改模时必填，新模时为空
	PlanStart   *Date  `json:"plan_start"`                    // 使用自定义 Date 类型
	PlanEnd     *Date  `json:"plan_end"`                      // 使用自定义 Date 类型
	Progress    int    `json:"progress"`
	// 🔒 Status 字段已移除：工单状态由系统自动管理，不允许前端传入
	WarehouseID *int64 `json:"warehouse_id"`
	DeliverTime *Date  `json:"deliver_time"` // 使用自定义 Date 类型
	Remark      string `json:"remark"`

	// 新模工单时必填字段（用于自动创建模具档案）
	NewMoldCode     string `json:"new_mold_code"`     // 新模具编号（新模工单必填）
	NewMoldName     string `json:"new_mold_name"`     // 新模具名称（新模工单必填）
	CustomerID      int64  `json:"customer_id"`       // 客户ID（新模工单必填）
	NewMoldType     string `json:"new_mold_type"`     // 模具类型（可选）：NEW=全新，USED=二手
	NewMaterialType string `json:"new_material_type"` // 材质类型（可选）：STEEL=钢材，HARDENED_STEEL=加硬钢材
	NewMoldRemark   string  `json:"new_mold_remark"`   // 模具备注（可选）

	// 外发工单字段
	WorkType              string   `json:"work_type"`                // 工单类型：IN_HOUSE=内部生产，OUTSOURCED=外发加工
	ProcessorID           *int64   `json:"processor_id"`             // 加工商ID（外发时填写）
	OutsourcePrice        *float64 `json:"outsource_price"`          // 外发单价
	OutsourceTotalAmount  *float64 `json:"outsource_total_amount"`   // 外发总金额
}

// MoldTryTestDTO 模具试模请求 DTO
type MoldTryTestDTO struct {
	MoldWorkID int64      `json:"mold_work_id" binding:"required"`
	TestTime   CustomTime `json:"test_time"`
	OkQty      int        `json:"ok_qty"`
	BadQty     int        `json:"bad_qty"`
	Result     string     `json:"result"`
	TestMode   string     `json:"test_mode"`
	UserID     *int64     `json:"user_id"`
	Remark     string     `json:"remark"`
}

// MoldDeliveryDTO 模具交付请求 DTO
type MoldDeliveryDTO struct {
	MoldWorkID        int64      `json:"mold_work_id" binding:"required"`
	CustomerID        int64      `json:"customer_id" binding:"required"`
	DeliveryTime      *FlexibleTime `json:"delivery_time"`
	DelivererID       *int64     `json:"deliverer_id"`
	ReceiverName      string     `json:"receiver_name"`
	ReceiverPhone     string     `json:"receiver_phone"`
	DeliveryAddress   string     `json:"delivery_address"`
	LogisticsCompany  string     `json:"logistics_company"`
	TrackingNo        string     `json:"tracking_no"`
	DeliveryMethod    string     `json:"delivery_method"`
	DocumentsAttached string     `json:"documents_attached"`
	Remark            string     `json:"remark"`
}

// MoldStockDTO 模具库存请求 DTO
type MoldStockDTO struct {
	MoldID      int64  `json:"mold_id" binding:"required"`
	WarehouseID *int64 `json:"warehouse_id"`
	StockType   int8   `json:"stock_type" binding:"required"`
	OperateTime string `json:"operate_time"` // 字符串格式："2006-01-02 15:04:05"
	OperatorID  *int64 `json:"operator_id"`
	CustomerID  *int64 `json:"customer_id"`
	Quantity    int    `json:"quantity"`
	Remark      string `json:"remark"`
}

// ProcessTemplateDTO 工序模板请求 DTO
type ProcessTemplateDTO struct {
	ID          int64  `json:"id"`
	ProcessName string `json:"process_name" binding:"required"`
	Sort        int    `json:"sort"`
	Status      int8   `json:"status"`
	Remark      string `json:"remark"`
}

type ToggleProcessRequest struct {
	WorkOrderID       int64 `json:"work_order_id" binding:"required"`
	ProcessTemplateID int64 `json:"process_template_id" binding:"required"`
	IsCompleted       int   `json:"is_completed" binding:"oneof=0 1 2"` // 0=未完成，1=已完成，2=跳过/不需要
}

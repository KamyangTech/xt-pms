package entity

import (
	"time"

	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
)

type Mold struct {
	ID                int64      `gorm:"primaryKey;autoIncrement;comment:模具 ID" json:"id"`
	MoldCode          string     `gorm:"size:50;not null;uniqueIndex;comment:模具编号（唯一）" json:"mold_code"`
	MoldName          string     `gorm:"size:100;comment:模具名称（用户自定义，如：塑料多产品模具、手机壳专用模具等）" json:"mold_name"`
	CustomerID        int64      `gorm:"not null;comment:客户 ID（customer.id）" json:"customer_id"`
	OwnershipType     int8       `gorm:"default:1;comment:权属类型: 1=自有, 2=客户资产, 3=第三方" json:"ownership_type"`
	LatestOrderID     *int64     `gorm:"comment:最新关联订单 ID" json:"latest_order_id"`
	LatestWorkOrderID *int64     `gorm:"comment:最新关联工单 ID" json:"latest_work_order_id"`
	TotalOrders       int        `gorm:"default:0;comment:累计订单数量" json:"total_orders"`
	TotalWorkOrders   int        `gorm:"default:0;comment:累计工单数量" json:"total_work_orders"`
	LastDeliveryTime  *time.Time `gorm:"comment:最后交付时间" json:"last_delivery_time"`
	LastUsageDate     *time.Time `gorm:"comment:最后一次使用日期" json:"last_usage_date"`
	TotalShots        int        `gorm:"default:0;comment:累计使用模次" json:"total_shots"`
	ProductionDate    *time.Time `gorm:"comment:开始生产日期" json:"production_date"`
	// 模具状态（字典：mold-process-status, mold-location, mold-life-cycle, mold-type, mold-process, material-type）
	ProcessStatus      string    `gorm:"size:50;default:PROCESSING;comment:工序状态" json:"process_status"`
	CurrentLocation    string    `gorm:"size:50;default:IN_FACTORY;comment:当前位置" json:"current_location"`
	LifeCycleStatus    string    `gorm:"default:'DESIGNING';comment:生命周期状态" json:"life_cycle_status"`
	MoldType           string    `gorm:"size:50;default:NEW;comment:模胚情况" json:"mold_type"`
	MoldProcess        string    `gorm:"size:50;default:'NORMAL';comment:模具工艺" json:"mold_process"`
	MaterialType       string    `gorm:"size:50;default:'STEEL';comment:材质" json:"material_type"`
	CavityCount        int       `gorm:"default:1;comment:模腔数（一模几腔）" json:"cavity_count"`
	ProductionCapacity *float64  `gorm:"type:decimal(10,2);comment:日产能（件/天）" json:"production_capacity"`
	Remark             string    `gorm:"type:text;comment:备注" json:"remark"`
	CreateTime         time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime         time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Customer *Customer `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"customer,omitempty"`

	// 模具图片（非数据库字段，通过 files 表查询）
	Images []File `gorm:"-" json:"images,omitempty"`

	// 产品关联（非数据库字段）
	ProductRelations []MoldProductRelation `gorm:"-" json:"product_relations,omitempty"`
}

func (Mold) TableName() string {
	return "mold"
}

type MoldOrder struct {
	ID         int64  `gorm:"primaryKey;autoIncrement;comment:模具订单 ID" json:"id"`
	OrderCode  string `gorm:"size:50;not null;uniqueIndex;comment:订单编号（唯一）" json:"order_code"`
	CustomerID int64  `gorm:"not null;comment:关联客户 ID（customer.id）" json:"customer_id"`
	// 订单信息（字典：mold-order-type, mold-order-status）
	OrderType        string     `gorm:"default:'NEW_MOLD';comment:订单类型" json:"order_type"`
	CreateUserID     *int64     `gorm:"comment:创建人ID" json:"create_user_id"`
	Priority         string     `gorm:"default:'NORMAL';comment:优先级（字典：order-priority）" json:"priority"`
	MoldID           *int64     `gorm:"comment:关联模具 ID（mold.id），新模订单可为空" json:"mold_id"`
	PlanStartDate    *dto.Date  `gorm:"comment:计划开始日期" json:"plan_start_date"`
	PlanEndDate      *dto.Date  `gorm:"comment:订单交期" json:"plan_end_date"`
	TotalAmount      float64    `gorm:"default:0.00;comment:订单总金额" json:"total_amount"`
	AdditionalAmount float64    `gorm:"default:0.00;comment:附加金额" json:"additional_amount"`
	Status           string     `gorm:"default:'NOT_STARTED';comment:订单状态" json:"status"`
	DeliverTime      *time.Time `gorm:"comment:交付时间" json:"deliver_time"`
	Remark           string     `gorm:"type:text;comment:备注" json:"remark"`
	CreateTime       time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime       time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
	MasterOrderID    *int64     `gorm:"comment:关联综合订单 ID (master_orders.id)" json:"master_order_id"`
	MasterCancelled  int8       `gorm:"default:0;comment:父综合订单已取消标记：0=否，1=是" json:"master_cancelled"`

	// 关联关系
	MasterOrder *MasterOrder     `gorm:"foreignKey:MasterOrderID" json:"master_order,omitempty"`
	Customer    *Customer        `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"customer,omitempty"`
	Mold        *Mold            `gorm:"foreignKey:MoldID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"mold,omitempty"`
	Items       []*MoldOrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"` // 一单多款明细

	// 非数据库字段：模具编号列表（用于列表显示）
	MoldCodes []string `gorm:"-" json:"mold_codes,omitempty"`

	// 非数据库字段：综合订单编号（用于列表显示）
	MasterOrderCode string `gorm:"-" json:"master_order_code,omitempty"`
}

func (MoldOrder) TableName() string {
	return "mold_order"
}

// MoldOrderItem 模具订单明细表（一单多款）
type MoldOrderItem struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:明细ID" json:"id"`
	OrderID    int64     `gorm:"not null;uniqueIndex:uk_order_mold;comment:订单ID（mold_order.id）" json:"order_id"`
	MoldID     int64     `gorm:"not null;uniqueIndex:uk_order_mold;comment:模具ID（mold.id）" json:"mold_id"`
	Amount     float64   `gorm:"default:0.00;comment:该模具的合同金额" json:"amount"`
	Sort       int       `gorm:"default:0;comment:排序号" json:"sort"`
	Remark     string    `gorm:"size:500;comment:备注" json:"remark"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	Order *MoldOrder `gorm:"foreignKey:OrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"order,omitempty"`
	Mold  *Mold      `gorm:"foreignKey:MoldID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"mold,omitempty"`

	// 临时字段（用于创建新模具时传递信息，不存储到数据库）
	MoldCode       string  `gorm:"-" json:"mold_code,omitempty"`       // 模具编号（新模时使用）
	MoldName       string  `gorm:"-" json:"mold_name,omitempty"`       // 模具名称（新模时使用）
	MoldType       string  `gorm:"-" json:"mold_type,omitempty"`       // 模胚情况
	MoldProcess    string  `gorm:"-" json:"mold_process,omitempty"`    // 模具工序
	MaterialType   string  `gorm:"-" json:"material_type,omitempty"`   // 材质类型
	WorkType       string  `gorm:"-" json:"work_type,omitempty"`       // 工单类型：IN_HOUSE=内部生产，OUTSOURCED=外发加工
	ProcessorID    *int64  `gorm:"-" json:"processor_id,omitempty"`    // 加工商ID（外发时填写）
	OutsourcePrice float64 `gorm:"-" json:"outsource_price,omitempty"` // 外发单价
}

func (MoldOrderItem) TableName() string {
	return "mold_order_item"
}

// MoldBOM 模具 BOM 表
type MoldBOM struct {
	ID         int64     `gorm:"primaryKey;autoIncrement;comment:BOM ID" json:"id"`
	MoldID     int64     `gorm:"not null;comment:关联模具 ID（mold.id）" json:"mold_id"`
	MaterialID int64     `gorm:"not null;comment:关联物料 ID（material.id）" json:"material_id"`
	Qty        float64   `gorm:"default:1.00;comment:物料用量" json:"qty"`
	UnitPrice  float64   `gorm:"default:0.00;comment:物料单价" json:"unit_price"`
	TotalPrice float64   `gorm:"default:0.00;comment:总价" json:"total_price"`
	Remark     string    `gorm:"size:255;comment:备注（如：模具配件规格）" json:"remark"`
	CreateTime time.Time `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (MoldBOM) TableName() string {
	return "mold_bom"
}

type MoldWorkOrder struct {
	ID          int64  `gorm:"primaryKey;autoIncrement;comment:模具工单 ID" json:"id"`
	WorkCode    string `gorm:"size:50;not null;uniqueIndex;comment:工单号（唯一）" json:"work_code"`
	MoldOrderID *int64 `gorm:"comment:关联模具订单 ID（mold_order.id）" json:"mold_order_id"`
	// 工单信息（字典：mold-work-order-type, mold-work-order-status）
	OrderType            string     `gorm:"default:'NEW_MOLD';comment:订单类型" json:"order_type"`
	WorkType             string     `gorm:"default:'IN_HOUSE';comment:工单类型（字典：work-type）：IN_HOUSE=内部生产，OUTSOURCED=外发加工" json:"work_type"`
	ProcessorID          *int64     `gorm:"comment:加工商ID（外发时填写）" json:"processor_id"`
	OutsourcePrice       float64    `gorm:"default:0.00;comment:外发单价" json:"outsource_price"`
	OutsourceTotalAmount float64    `gorm:"default:0.00;comment:外发总金额" json:"outsource_total_amount"`
	MoldID               *int64     `gorm:"comment:关联模具 ID（mold.id）" json:"mold_id"`
	PlanStart            *time.Time `gorm:"comment:计划开始时间" json:"plan_start"`
	PlanEnd              *time.Time `gorm:"comment:计划结束时间" json:"plan_end"`
	ActualStart          *time.Time `gorm:"comment:实际开始时间" json:"actual_start"`
	ActualEnd            *time.Time `gorm:"comment:实际结束时间" json:"actual_end"`
	Progress             int        `gorm:"default:0;comment:工单总进度（0-100）" json:"progress"`
	Status               string     `gorm:"default:'NOT_STARTED';comment:工单状态" json:"status"`
	WarehouseID          *int64     `gorm:"comment:入库仓库 ID（warehouse.id）" json:"warehouse_id"`
	DeliverTime          *time.Time `gorm:"comment:交付时间" json:"deliver_time"`
	ActualDeliverTime    *time.Time `gorm:"comment:实际交付时间" json:"actual_deliver_time"`
	OriginalMoldStatus   *string    `gorm:"comment:修模/改造前模具的原始生命周期状态" json:"original_mold_status"`
	Remark               string     `gorm:"type:text;comment:备注" json:"remark"`
	IsCancelled          int8       `gorm:"default:0;comment:是否已取消：0=否，1=是" json:"is_cancelled"`
	MasterCancelled      int8       `gorm:"default:0;comment:父综合订单已取消标记：0=否，1=是" json:"master_cancelled"`
	CancelTime           *time.Time `gorm:"comment:取消时间" json:"cancel_time"`
	CancelReason         string     `gorm:"type:text;comment:取消原因" json:"cancel_reason"`
	CancelUserID         *int64     `gorm:"comment:取消操作人 ID（sys_user.id）" json:"cancel_user_id"`
	CreateTime           time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime           time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 关联关系
	MoldOrder *MoldOrder `gorm:"foreignKey:MoldOrderID" json:"mold_order,omitempty"`
	Mold      *Mold      `gorm:"foreignKey:MoldID" json:"mold,omitempty"`

	// 非数据库字段：工序列表（仅用于返回前端）
	Processes []WorkOrderProcessInfo `gorm:"-" json:"processes,omitempty"`

	// 非数据库字段：是否被锁定（用于前端禁用按钮，修模/改模工单会锁定同模具的其他工单）
	IsLocked bool `gorm:"-" json:"is_locked,omitempty"`
}

// WorkOrderProcessInfo 工单工序信息（用于前端展示）
type WorkOrderProcessInfo struct {
	ProcessTemplateID int64      `json:"process_template_id"`
	ProcessName       string     `json:"process_name"`
	IsCompleted       int8       `json:"is_completed"`
	CompletedAt       *time.Time `json:"completed_at"`
}

func (MoldWorkOrder) TableName() string {
	return "mold_work_order"
}

type MoldTryTest struct {
	ID         int64      `gorm:"primaryKey;autoIncrement;comment:试模 ID" json:"id"`
	MoldWorkID int64      `gorm:"not null;comment:关联模具工单 ID（mold_work_order.id）" json:"mold_work_id"`
	TestTime   *time.Time `gorm:"comment:试模时间" json:"test_time"`
	OkQty      int        `gorm:"default:0;comment:试模合格数量" json:"ok_qty"`
	BadQty     int        `gorm:"default:0;comment:试模不良数量" json:"bad_qty"`
	Result     string     `gorm:"type:text;comment:试模结果（如：合格、需返修）" json:"result"`
	TestMode   string     `gorm:"type:varchar(20);comment:试模方式（自家试模、外协试模、免试）" json:"test_mode"`
	UserID     *int64     `gorm:"comment:操作人 ID（sys_user.id）" json:"user_id"`
	Remark     string     `gorm:"type:text;comment:备注" json:"remark"`
	CreateTime time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (MoldTryTest) TableName() string {
	return "mold_try_test"
}

type MoldStock struct {
	ID          int64      `gorm:"primaryKey;autoIncrement;comment:库存记录 ID" json:"id"`
	MoldID      int64      `gorm:"not null;comment:关联模具 ID（mold.id）" json:"mold_id"`
	WarehouseID *int64     `gorm:"comment:关联仓库 ID（warehouse.id）" json:"warehouse_id"`
	StockType   int8       `gorm:"not null;comment:库存类型：1=入库，2=出库，3=交付客户" json:"stock_type"`
	OperateTime *time.Time `gorm:"comment:操作时间" json:"operate_time"`
	OperatorID  *int64     `gorm:"comment:操作人 ID（sys_user.id）" json:"operator_id"`
	CustomerID  *int64     `gorm:"comment:交付客户 ID（stock_type=3 时必填，customer.id）" json:"customer_id"`
	Quantity    int        `gorm:"default:1;comment:数量" json:"quantity"`
	Remark      string     `gorm:"size:255;comment:备注（如：入库原因、交付单号）" json:"remark"`
	CreateTime  time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime  time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (MoldStock) TableName() string {
	return "mold_stock"
}

type MoldOrderPayment struct {
	ID             int64      `gorm:"primaryKey;autoIncrement;comment:款项 ID" json:"id"`
	OrderID        int64      `gorm:"not null;comment:关联模具订单 ID" json:"order_id"`
	WorkOrderID    *int64     `gorm:"comment:关联工单 ID（mold_work_order.id），用于按模具核算" json:"work_order_id"`
	PaymentType    string     `gorm:"size:50;not null;comment:款项类型（自定义，如：首款、二期款、尾款等）" json:"payment_type"`
	ContractAmount float64    `gorm:"default:0.00;comment:合同约定金额" json:"contract_amount"`
	ActualAmount   float64    `gorm:"default:0.00;comment:实际到账金额" json:"actual_amount"`
	ShouldAmount   float64    `gorm:"default:0.00;comment:应收尾款金额" json:"should_amount"`
	PaymentDate    *time.Time `gorm:"comment:收款日期" json:"payment_date"`
	Status         int8       `gorm:"default:1;comment:状态：1=未收，2=部分收款，3=已收款" json:"status"`
	Remark         string     `gorm:"size:500;comment:备注" json:"remark"`
	CreateTime     time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime     time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (MoldOrderPayment) TableName() string {
	return "mold_order_payment"
}

// ProcessTemplate 工序模板表（固定工序列表）
type ProcessTemplate struct {
	ID          int64     `gorm:"primaryKey;autoIncrement;comment:模板ID" json:"id"`
	GroupID     int       `gorm:"default:0;comment:分组ID（同组工序可并行）" json:"group_id"`
	GroupName   string    `gorm:"size:50;comment:分组名称" json:"group_name"`
	ProcessName string    `gorm:"size:100;not null;comment:工序名称" json:"process_name"`
	Sort        int       `gorm:"not null;default:0;comment:排序号" json:"sort"`
	Status      int8      `gorm:"not null;default:1;comment:状态：1=启用，2=禁用" json:"status"`
	Remark      string    `gorm:"size:500;comment:备注" json:"remark"`
	CreatedAt   time.Time `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
}

func (ProcessTemplate) TableName() string {
	return "process_template"
}

// WorkOrderProcess 工单工序关联表（记录每个工单的工序完成状态）
type WorkOrderProcess struct {
	ID                int64      `gorm:"primaryKey;autoIncrement;comment:关联ID" json:"id"`
	WorkOrderID       int64      `gorm:"not null;comment:工单ID" json:"work_order_id"`
	ProcessTemplateID int64      `gorm:"not null;comment:工序模板ID" json:"process_template_id"`
	IsCompleted       int8       `gorm:"not null;default:0;comment:工序状态：0=未完成，1=已完成，2=已跳过" json:"is_completed"`
	CompletedAt       *time.Time `gorm:"comment:完成时间" json:"completed_at"`
	Remark            string     `gorm:"size:500;comment:备注" json:"remark"`
	CreatedAt         time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`

	// 关联关系
	WorkOrder       *MoldWorkOrder   `gorm:"foreignKey:WorkOrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"work_order,omitempty"`
	ProcessTemplate *ProcessTemplate `gorm:"foreignKey:ProcessTemplateID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"process_template,omitempty"`
}

func (WorkOrderProcess) TableName() string {
	return "work_order_process"
}

// MoldAcceptance 模具验收记录表（已废弃，保留结构体用于兼容历史数据）
// 简化流程后取消验收步骤，试模合格后直接入库
type MoldAcceptance struct {
	ID                    int64      `gorm:"primaryKey;autoIncrement;comment:验收 ID" json:"id"`
	MoldWorkID            int64      `gorm:"not null;comment:关联模具工单 ID" json:"mold_work_id"`
	AcceptanceTime        *time.Time `gorm:"comment:验收时间" json:"acceptance_time"`
	AcceptorID            *int64     `gorm:"comment:验收人 ID" json:"acceptor_id"`
	AcceptanceResult      string     `gorm:"type:varchar(20);comment:验收结果：合格/不合格/让步接收" json:"acceptance_result"`
	QualityScore          *int       `gorm:"comment:质量评分（1-100）" json:"quality_score"`
	DimensionCheck        string     `gorm:"type:text;comment:尺寸检测记录" json:"dimension_check"`
	AppearanceCheck       string     `gorm:"type:text;comment:外观检查记录" json:"appearance_check"`
	FunctionCheck         string     `gorm:"type:text;comment:功能测试记录" json:"function_check"`
	IssuesFound           string     `gorm:"type:text;comment:发现的问题" json:"issues_found"`
	RectificationRequired int8       `gorm:"default:0;comment:是否需要整改：0=否，1=是" json:"rectification_required"`
	RectificationDeadline *time.Time `gorm:"comment:整改截止日期" json:"rectification_deadline"`
	AttachmentUrls        string     `gorm:"type:text;comment:附件URL列表（JSON格式）" json:"attachment_urls"`
	Remark                string     `gorm:"type:text;comment:备注" json:"remark"`
	CreateTime            time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime            time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (MoldAcceptance) TableName() string {
	return "mold_acceptance"
}

type MoldDelivery struct {
	ID                 int64      `gorm:"primaryKey;autoIncrement;comment:交付 ID" json:"id"`
	MoldWorkID         int64      `gorm:"not null;comment:关联模具工单 ID" json:"mold_work_id"`
	DeliveryNo         string     `gorm:"size:50;uniqueIndex;comment:交付单号" json:"delivery_no"`
	CustomerID         int64      `gorm:"not null;comment:客户 ID" json:"customer_id"`
	DeliveryTime       *time.Time `gorm:"comment:交付时间" json:"delivery_time"`
	DelivererID        *int64     `gorm:"comment:交付人 ID" json:"deliverer_id"`
	ReceiverName       string     `gorm:"size:50;comment:收货人姓名" json:"receiver_name"`
	ReceiverPhone      string     `gorm:"size:20;comment:收货人电话" json:"receiver_phone"`
	DeliveryAddress    string     `gorm:"type:text;comment:交付地址" json:"delivery_address"`
	LogisticsCompany   string     `gorm:"size:100;comment:物流公司" json:"logistics_company"`
	TrackingNo         string     `gorm:"size:50;comment:物流单号" json:"tracking_no"`
	DeliveryMethod     string     `gorm:"size:20;comment:交付方式：自提/快递/物流" json:"delivery_method"`
	DocumentsAttached  string     `gorm:"type:text;comment:随附文档（JSON格式）" json:"documents_attached"`
	ConfirmationStatus int8       `gorm:"default:1;comment:确认状态：1=待确认，2=已确认，3=拒收" json:"confirmation_status"`
	ConfirmationTime   *time.Time `gorm:"comment:确认时间" json:"confirmation_time"`
	RejectionReason    string     `gorm:"type:text;comment:拒收原因" json:"rejection_reason"`
	Remark             string     `gorm:"type:text;comment:备注" json:"remark"`
	CreateTime         time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime         time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`
}

func (MoldDelivery) TableName() string {
	return "mold_delivery"
}

type MoldUsageRecord struct {
	ID                    int64      `gorm:"primaryKey;autoIncrement;comment:使用记录ID" json:"id"`
	MoldID                int64      `gorm:"not null;comment:模具ID" json:"mold_id"`
	ProductionOrderID     int64      `gorm:"not null;comment:生产订单ID" json:"production_order_id"`
	ProductionOrderItemID *int64     `gorm:"comment:生产订单明细ID（用于细分到具体产品）" json:"production_order_item_id"`
	ProductID             *int64     `gorm:"comment:产品ID（用于区分同一模具不同产品）" json:"product_id"`
	ProductName           *string    `gorm:"size:200;comment:产品名称快照" json:"product_name"`
	UsageDate             *time.Time `gorm:"type:date;comment:使用日期" json:"usage_date"`
	TotalShots            int        `gorm:"default:0;comment:累计模次(本次使用)" json:"total_shots"`
	OkQty                 int        `gorm:"default:0;comment:良品数量" json:"ok_qty"`
	BadQty                int        `gorm:"default:0;comment:不良数量" json:"bad_qty"`
	StartDate             *time.Time `gorm:"type:date;comment:开始使用日期" json:"start_date"`
	EndDate               *time.Time `gorm:"type:date;comment:结束使用日期" json:"end_date"`
	Remark                string     `gorm:"type:text;comment:备注" json:"remark"`
	CreateTime            time.Time  `gorm:"autoCreateTime;comment:创建时间" json:"create_time"`
	UpdateTime            time.Time  `gorm:"autoUpdateTime;comment:更新时间" json:"update_time"`

	// 非数据库字段（用于查询结果）
	OrderCode  string     `gorm:"-" json:"order_code,omitempty"`  // 订单编号
	ReportTime *time.Time `gorm:"-" json:"report_time,omitempty"` // 报工时间（使用时间）
	MoldCode   string     `gorm:"-" json:"mold_code,omitempty"`   // 模具编号（用于列表显示）
	MoldName   string     `gorm:"-" json:"mold_name,omitempty"`   // 模具名称（用于列表显示）

	// 关联关系
	Mold            *Mold            `gorm:"foreignKey:MoldID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"mold,omitempty"`
	ProductionOrder *ProductionOrder `gorm:"foreignKey:ProductionOrderID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"production_order,omitempty"`
	Product         *Product         `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL" json:"product,omitempty"`
}

func (MoldUsageRecord) TableName() string {
	return "mold_usage_record"
}

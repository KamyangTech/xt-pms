package dto

type DashboardOverview struct {
	// ======== 顶部四张卡片 ========

	// 卡片1：本月新增模具订单数量
	MonthlyNewMoldOrders      int     `json:"monthlyNewMoldOrders"`
	MonthlyNewMoldOrdersGrowth float64 `json:"monthlyNewMoldOrdersGrowth"`

	// 卡片2：本月新增生产订单数量
	MonthlyNewProductionOrders      int     `json:"monthlyNewProductionOrders"`
	MonthlyNewProductionOrdersGrowth float64 `json:"monthlyNewProductionOrdersGrowth"`

	// 卡片3：正在生产流程的模具数量
	InProductionMolds      int     `json:"inProductionMolds"`
	InProductionMoldsGrowth float64 `json:"inProductionMoldsGrowth"`

	// 卡片4：本月新增应收款（营业额）
	MonthlyNewReceivable      float64 `json:"monthlyNewReceivable"`
	MonthlyNewReceivableGrowth float64 `json:"monthlyNewReceivableGrowth"`

	// ======== 其他概览数据（保持兼容） ========
	TodayOutput       int     `json:"todayOutput"`
	TodayOutputGrowth float64 `json:"todayOutputGrowth"`
	TotalMolds        int     `json:"totalMolds"`
	MoldsGrowth       float64 `json:"moldsGrowth"`
	StockWarning       int     `json:"stockWarning"`
	StockWarningChange float64 `json:"stockWarningChange"`
	RawMaterialStock   int     `json:"rawMaterialStock"`
	RawMaterialGrowth  float64 `json:"rawMaterialGrowth"`
	PendingWorkOrders  int     `json:"pendingWorkOrders"`
	WorkOrdersGrowth   float64 `json:"workOrdersGrowth"`
	TotalCustomers     int     `json:"totalCustomers"`
	CustomersGrowth    float64 `json:"customersGrowth"`
	TotalSuppliers     int     `json:"totalSuppliers"`
	SuppliersGrowth    float64 `json:"suppliersGrowth"`
	QualificationRate   float64 `json:"qualificationRate"`
	QualificationGrowth float64 `json:"qualificationGrowth"`
}

type ProductionTrend struct {
	Categories []string       `json:"categories"` // 分类（时间轴）
	Datasets   []TrendDataset `json:"datasets"`   // 数据集
}

type TrendDataset struct {
	Name string    `json:"name"` // 数据集名称
	Data []float64 `json:"data"` // 数据
}

type MoldStatus struct {
	Total       int `json:"total"`       // 总数
	InUse       int `json:"inUse"`       // 使用中
	Maintenance int `json:"maintenance"` // 维护中
	Idle        int `json:"idle"`        // 闲置
	Scrap       int `json:"scrap"`       // 报废
}

type OrderRanking struct {
	Items []RankingItem `json:"items"` // 排名项
}

type RankingItem struct {
	Rank       int    `json:"rank"`       // 排名
	Name       string `json:"name"`       // 名称
	Change     int    `json:"change"`     // 变化（正数上升，负数下降）
	OrderCount int    `json:"orderCount"` // 订单数量
}

type InProductionMold struct {
	MoldID            int64  `json:"mold_id"`
	MoldCode          string `json:"mold_code"`
	MoldName          string `json:"mold_name"`
	WorkOrderID       int64  `json:"work_order_id"`
	Progress          int    `json:"progress"`
	WorkOrderStatus   string `json:"work_order_status"`
	LastCompletedStep string `json:"last_completed_step"`
}

type InProductionMolds struct {
	Items []InProductionMold `json:"items"`
}

type InProgressMoldOrderItem struct {
	OrderID      int64  `json:"order_id"`
	OrderCode    string `json:"order_code"`
	MoldCodes    string `json:"mold_codes"`    // 模具编号列表（逗号分隔）
	CustomerName string `json:"customer_name"`
	Status       string `json:"status"`
	StatusLabel  string `json:"status_label"`
	PlanEndDate  string `json:"plan_end_date"`
}

// InProgressMoldOrders 进行中的模具订单列表（分页）
type InProgressMoldOrders struct {
	Items    []InProgressMoldOrderItem `json:"items"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
}

type InProgressProductionOrderItem struct {
	OrderID      int64  `json:"order_id"`
	OrderCode    string `json:"order_code"`
	CustomerName string `json:"customer_name"`
	Status       string `json:"status"`
	StatusLabel  string `json:"status_label"`
	PlanEndDate  string `json:"plan_end_date"`
}

type InProgressProductionOrders struct {
	Items    []InProgressProductionOrderItem `json:"items"`
	Total    int64                           `json:"total"`
}

type FinanceOverviewDashboard struct {
	// 图表数据（按日期分组的应收/应付金额）
	ChartData []FinanceChartItem `json:"chartData"`

	// 本月应收合计
	ThisMonthReceivable      float64 `json:"thisMonthReceivable"`
	ThisMonthReceivableGrowth float64 `json:"thisMonthReceivableGrowth"`

	// 本月应付合计
	ThisMonthPayable      float64 `json:"thisMonthPayable"`
	ThisMonthPayableGrowth float64 `json:"thisMonthPayableGrowth"`
}

type FinanceChartItem struct {
	Date             string  `json:"date"`
	ReceivableAmount float64 `json:"receivableAmount"`
	PayableAmount    float64 `json:"payableAmount"`
}

// ======== 统计报表页概览 ========

// DashboardDetailOverview 统计报表页概览（4张卡片）
type DashboardDetailOverview struct {
	// 卡片1：本月完成生产订单数
	CompletedProductionOrders         int     `json:"completedProductionOrders"`
	CompletedProductionOrdersGrowth   float64 `json:"completedProductionOrdersGrowth"`

	// 卡片2：本月完成生产订单数 → 进行中生产订单数
	InProgressProductionOrders         int     `json:"inProgressProductionOrders"`
	InProgressProductionOrdersGrowth   float64 `json:"inProgressProductionOrdersGrowth"`

	// 卡片5：进行中模具工单数
	InProgressMoldWorkOrders         int     `json:"inProgressMoldWorkOrders"`
	InProgressMoldWorkOrdersGrowth   float64 `json:"inProgressMoldWorkOrdersGrowth"`

	// 卡片6：本月模具完工数
	CompletedMoldWorkOrders         int     `json:"completedMoldWorkOrders"`
	CompletedMoldWorkOrdersGrowth   float64 `json:"completedMoldWorkOrdersGrowth"`
}

// ======== 统计报表页趋势图 ========

type DetailProductionTrendItem struct {
	Date                      string `json:"date"`
	CompletedProductionOrders int    `json:"completedProductionOrders"`
	InProgressProductionOrders int   `json:"inProgressProductionOrders"`
	CompletedMoldWorkOrders   int    `json:"completedMoldWorkOrders"`
	InProgressMoldWorkOrders  int    `json:"inProgressMoldWorkOrders"`
}

type DetailProductionTrend struct {
	Categories []string `json:"categories"`
	Datasets   []TrendDataset `json:"datasets"`
}

// ======== 交付准时率 ========

type DeliveryPunctualityItem struct {
	Month      string `json:"month"`
	Early      int    `json:"early"`      // 提前交付
	OnTime     int    `json:"onTime"`     // 准时交付
	Delayed    int    `json:"delayed"`    // 延迟交付
}

type DeliveryPunctuality struct {
	Categories []string        `json:"categories"`
	Datasets   []TrendDataset  `json:"datasets"`
}

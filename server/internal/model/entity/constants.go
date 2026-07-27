package entity

// ============================================
// 模具工单状态常量 (已废弃)
// 请使用 internal/model/constant/dict_constant.go 中的 WorkOrderStatus* 常量
// ============================================
const (
	// 未开始
	WorkOrderStatusNotStarted int8 = 1 // 未开始
	// 加工中
	WorkOrderStatusProcessing int8 = 2 // 加工中
	// 已完成（待试模）
	WorkOrderStatusCompleted int8 = 3 // 已完成（待试模）
	// 已试模
	WorkOrderStatusTested int8 = 4 // 已试模
	// 已入库
	WorkOrderStatusInWarehouse int8 = 5 // 已入库
	// 已交付
	WorkOrderStatusDelivered int8 = 6 // 已交付
	// 已取消
	WorkOrderStatusCancelled int8 = 7 // 已取消
)

func WorkOrderStatusName(status int8) string {
	switch status {
	case WorkOrderStatusNotStarted:
		return "未开始"
	case WorkOrderStatusProcessing:
		return "加工中"
	case WorkOrderStatusCompleted:
		return "已完成"
	case WorkOrderStatusTested:
		return "已试模"
	case WorkOrderStatusInWarehouse:
		return "已入库"
	case WorkOrderStatusDelivered:
		return "已交付"
	case WorkOrderStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

// ============================================
// 模具订单状态常量 (已废弃)
// 请使用 internal/model/constant/dict_constant.go 中的 MoldOrderStatus* 常量
// ============================================
const (
	MoldOrderStatusNotStarted int8 = 1 // 待生产
	MoldOrderStatusProcessing int8 = 2 // 加工中
	MoldOrderStatusTesting    int8 = 3 // 试模
	MoldOrderStatusCompleted  int8 = 4 // 已完成
	MoldOrderStatusCancelled  int8 = 5 // 已取消
	MoldOrderStatusDelivered  int8 = 6 // 已交付
)

func MoldOrderStatusName(status int8) string {
	switch status {
	case MoldOrderStatusNotStarted:
		return "待生产"
	case MoldOrderStatusProcessing:
		return "加工中"
	case MoldOrderStatusTesting:
		return "试模"
	case MoldOrderStatusCompleted:
		return "已完成"
	case MoldOrderStatusCancelled:
		return "已取消"
	case MoldOrderStatusDelivered:
		return "已交付"
	default:
		return "未知"
	}
}

// ============================================
// 模具生命周期状态常量 (已废弃)
// 请使用 internal/model/constant/dict_constant.go 中的 MoldStatus* 常量
// ============================================
const (
	MoldLifeCycleNotStarted  int8 = 1 // 未开始
	MoldLifeCycleProcessing  int8 = 2 // 加工中
	MoldLifeCyclePendingTest int8 = 3 // 待试模
	MoldLifeCycleTested      int8 = 4 // 已试模
	MoldLifeCycleCompleted   int8 = 5 // 已完成
	MoldLifeCycleRepairing   int8 = 6 // 维修中
	MoldLifeCycleInWarehouse int8 = 7 // 仓库库存
	MoldLifeCycleDelivered   int8 = 8 // 已交付客户
	MoldLifeCycleScrapped    int8 = 9 // 报废
)

func MoldLifeCycleStatusName(status int8) string {
	switch status {
	case MoldLifeCycleNotStarted:
		return "未开始"
	case MoldLifeCycleProcessing:
		return "加工中"
	case MoldLifeCyclePendingTest:
		return "待试模"
	case MoldLifeCycleTested:
		return "已试模"
	case MoldLifeCycleCompleted:
		return "已完成"
	case MoldLifeCycleRepairing:
		return "维修中"
	case MoldLifeCycleInWarehouse:
		return "仓库库存"
	case MoldLifeCycleDelivered:
		return "已交付客户"
	case MoldLifeCycleScrapped:
		return "报废"
	default:
		return "未知"
	}
}

// ============================================
// 工序完成状态常量
// ============================================
const (
	ProcessStatusPending   int8 = 0 // 未完成
	ProcessStatusCompleted int8 = 1 // 已完成
	ProcessStatusSkipped   int8 = 2 // 已跳过
)

func ProcessStatusName(status int8) string {
	switch status {
	case ProcessStatusPending:
		return "未完成"
	case ProcessStatusCompleted:
		return "已完成"
	case ProcessStatusSkipped:
		return "已跳过"
	default:
		return "未知"
	}
}

// ============================================
// 订单类型常量
// ============================================
const (
	OrderTypeNewMold    string = "NEW_MOLD"     // 新模订单
	OrderTypeRepairMold string = "REPAIR"       // 修模订单
	OrderTypeModifyMold string = "MODIFICATION" // 改造订单
)

func OrderTypeName(orderType string) string {
	switch orderType {
	case OrderTypeNewMold:
		return "新模订单"
	case OrderTypeRepairMold:
		return "修模订单"
	case OrderTypeModifyMold:
		return "改造订单"
	default:
		return "未知"
	}
}

// ============================================
// 用户状态常量
// ============================================
const (
	UserStatusInactive int8 = 0 // 禁用
	UserStatusActive   int8 = 1 // 启用
)

// ============================================
// 产品状态常量
// ============================================
const (
	ProductStatusInactive int8 = 0 // 停用
	ProductStatusActive   int8 = 1 // 启用
)

// ============================================
// 操作日志类型常量
// ============================================
const (
	OperTypeCreate  int8 = 1 // 新增
	OperTypeDelete  int8 = 2 // 删除
	OperTypeUpdate  int8 = 3 // 修改
	OperTypeQuery   int8 = 4 // 查询
	OperTypeImport  int8 = 5 // 导入
	OperTypeExport  int8 = 6 // 导出
	OperTypeLogin   int8 = 7 // 登录
	OperTypeLogout  int8 = 8 // 登出
)

func OperTypeName(operType int8) string {
	switch operType {
	case OperTypeCreate:
		return "新增"
	case OperTypeDelete:
		return "删除"
	case OperTypeUpdate:
		return "修改"
	case OperTypeQuery:
		return "查询"
	case OperTypeImport:
		return "导入"
	case OperTypeExport:
		return "导出"
	case OperTypeLogin:
		return "登录"
	case OperTypeLogout:
		return "登出"
	default:
		return "未知"
	}
}



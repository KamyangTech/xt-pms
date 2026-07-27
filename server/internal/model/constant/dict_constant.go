package constant

// ============================================
// 模具状态常量 (mold-status)
// ============================================
const (
	MoldStatusDesigning      = "DESIGNING"       // 设计中
	MoldStatusProcessing     = "PROCESSING"      // 加工中
	MoldStatusTesting        = "TESTING"         // 试模
	MoldStatusTested         = "TESTED"          // 已试模
	MoldStatusProducing      = "PRODUCING"       // 生产中
	MoldStatusRepairing      = "REPAIRING"       // 维修
	MoldStatusInStock        = "IN_STOCK"        // 仓库库存
	MoldStatusDelivered      = "DELIVERED"       // 已交付客户
	MoldStatusScrapped       = "SCRAPPED"        // 报废
	MoldStatusManufacturingInterrupted = "MANUFACTURING_INTERRUPTED" // 制造中断
)

// ============================================
// 模具工单状态常量 (work-order-status)
// ============================================
const (
	WorkOrderStatusNotStarted = "NOT_STARTED" // 未开始
	WorkOrderStatusProcessing = "PROCESSING"  // 加工中
	WorkOrderStatusCompleted  = "COMPLETED"   // 已完成
	WorkOrderStatusTested     = "TESTED"      // 已试模
	WorkOrderStatusStockedIn  = "STOCKED_IN"  // 已入库
	WorkOrderStatusDelivered  = "DELIVERED"   // 已交付
	WorkOrderStatusCancelled  = "CANCELLED"   // 已取消
)

// ============================================
// 模具订单状态常量 (mold-order-status)
// ============================================
const (
	MoldOrderStatusNotStarted  = "NOT_STARTED"  // 未开始
	MoldOrderStatusProcessing  = "PROCESSING"   // 加工中
	MoldOrderStatusPendingTest = "PENDING_TEST" // 待试模
	MoldOrderStatusTested      = "TESTED"       // 已试模
	MoldOrderStatusCompleted   = "COMPLETED"    // 已完成
	MoldOrderStatusCancelled   = "CANCELLED"    // 已取消
	MoldOrderStatusDelivered   = "DELIVERED"    // 已交付
)

// MoldOrderStatusNumericToCode 将前端数字状态码转换为数据库字符串状态码
// 1→NOT_STARTED, 2→PROCESSING, 3→PENDING_TEST, 4→TESTED, 5→COMPLETED, 6→CANCELLED, 7→DELIVERED
func MoldOrderStatusNumericToCode(n int8) string {
	switch n {
	case 1:
		return MoldOrderStatusNotStarted
	case 2:
		return MoldOrderStatusProcessing
	case 3:
		return MoldOrderStatusPendingTest
	case 4:
		return MoldOrderStatusTested
	case 5:
		return MoldOrderStatusCompleted
	case 6:
		return MoldOrderStatusCancelled
	case 7:
		return MoldOrderStatusDelivered
	default:
		return ""
	}
}

// ============================================
// 综合订单状态常量 (master-order-status)
// ============================================
const (
	MasterOrderStatusPENDING       = "PENDING"         // 待开始
	MasterOrderStatusMoldProducing = "MOLD_PRODUCING"  // 模具制造中
	MasterOrderStatusProducing     = "PRODUCING"       // 生产中
	MasterOrderStatusPendingShip   = "PENDING_SHIP"    // 待发货
	MasterOrderStatusCompleted     = "COMPLETED"       // 已完成
	MasterOrderStatusCancelled     = "CANCELLED"       // 已取消
)

// ============================================
// 模具工艺常量 (mold-process)
// ============================================
const (
	MoldProcessNormal    = "NORMAL"     // 普通模具
	MoldProcessHotRunner = "HOT_RUNNER" // 热流道
)

// ============================================
// 材质类型常量 (material-type)
// ============================================
const (
	MaterialTypeSteel         = "STEEL"          // 钢材
	MaterialTypeHardenedSteel = "HARDENED_STEEL" // 加硬钢材
)

// ============================================
// 模胚情况常量 (mold-type)
// ============================================
const (
	MoldTypeNew  = "NEW"  // 全新
	MoldTypeUsed = "USED" // 二手
)

// ============================================
// 订单优先级常量 (order-priority)
// ============================================
const (
	OrderPriorityNormal     = "NORMAL"      // 普通
	OrderPriorityImportant  = "IMPORTANT"   // 重要
	OrderPriorityUrgent     = "URGENT"      // 紧急
	OrderPriorityVeryUrgent = "VERY_URGENT" // 非常紧急
)

// ============================================
// 生产类型常量 (production-type)
// ============================================
const (
	ProductionTypeInHouse    = "IN_HOUSE"   // 本厂生产
	ProductionTypeOutsourced = "OUTSOURCED" // 外发生产
)

// ============================================
// 模具来源常量 (mold-source)
// ============================================
const (
	MoldSourceSelfMade         = "SELF_MADE"         // 本厂制造模具
	MoldSourceCustomerProvided = "CUSTOMER_PROVIDED" // 客户提供模具
)

// ============================================
// 模具位置常量 (mold-location)
// ============================================
const (
	MoldLocationInFactory   = "IN_FACTORY"  // 厂内
	MoldLocationOutsourcing = "OUTSOURCING" // 外发中
	MoldLocationAtCustomer  = "AT_CUSTOMER" // 客户处
)

// ============================================
// 订单/工单类型常量 (mold_order_type)
// ============================================
const (
	MoldOrderTypeNew          = "NEW_MOLD"     // 新模
	MoldOrderTypeRepair       = "REPAIR"       // 修模
	MoldOrderTypeModification = "MODIFICATION" // 改造
)

// ============================================
// 塑胶原料常量 (plastic-material)
// ============================================
const (
	PlasticMaterialPP   = "PP"   // PP（聚丙烯）
	PlasticMaterialPE   = "PE"   // PE（聚乙烯）
	PlasticMaterialABS  = "ABS"  // ABS
	PlasticMaterialPS   = "PS"   // PS（聚苯乙烯）
	PlasticMaterialPC   = "PC"   // PC（聚碳酸酯）
	PlasticMaterialPVC  = "PVC"  // PVC（聚氯乙烯）
	PlasticMaterialPMMA = "PMMA" // PMMA（亚克力）
	PlasticMaterialPA   = "PA"   // PA（尼龙）
	PlasticMaterialPOM  = "POM"  // POM（聚甲醛）
	PlasticMaterialPET  = "PET"  // PET（聚酯）
	PlasticMaterialTPE  = "TPE"  // TPE（热塑性弹性体）
	PlasticMaterialTPU  = "TPU"  // TPU（热塑性聚氨酯）
)

// ============================================
// 班次管理常量 (shift-type)
// ============================================
const (
	ShiftTypeDay   = "DAY_SHIFT"   // 白班
	ShiftTypeNight = "NIGHT_SHIFT" // 夜班
)

// ============================================
// 款项类型常量 (payment-type)
// 注意：款项类型已改为字典驱动，在 sys_dict_data(dict_code=payment-type) 中配置。
// 行为标志 needs_delivery_unlock 控制该类型是否需要交付后解锁。
// 业务代码中使用 PaymentTypeNeedsDeliveryUnlock() 函数读取该标志。
// 这些常量仅作为代码引用，实际值以字典为准。
// ============================================
const (
	PaymentTypeFirst  = "FIRST_PAYMENT"  // 首款（可自定义）
	PaymentTypeMiddle = "MIDDLE_PAYMENT" // 中期款（可自定义）
	PaymentTypeFinal  = "FINAL_PAYMENT"  // 尾款（可自定义）
)

// ============================================
// 试模方式常量 (test-mode)
// ============================================
const (
	TestModeInternal = "SELF_TEST"     // 自家试模
	TestModeExternal = "EXTERNAL_TEST" // 外协试模
	TestModeExempt   = "EXEMPTED"      // 免试
)

// ============================================
// 生产设备常量 (production-equipment)
// ============================================
const (
	ProductionEquipment960Ton = "TON_960" // 960吨机
	ProductionEquipment600Ton = "TON_600" // 600吨机
	ProductionEquipment300Ton = "TON_300" // 300吨机
)

// ============================================
// 发货方式常量 (delivery_method) - 已使用字符串，保持不变
// ============================================
const (
	DeliveryMethodExpress   = "EXPRESS"   // 快递
	DeliveryMethodLalamove  = "LALAMOVE"  // 货拉拉
	DeliveryMethodPickup    = "PICKUP"    // 自提
	DeliveryMethodLogistics = "LOGISTICS" // 物流配送
)

// ============================================
// 快递公司常量 (express_company)
// ============================================
const (
	ExpressCompanySF   = "SF"   // 顺丰速运
	ExpressCompanyYT   = "YT"   // 圆通速递
	ExpressCompanyZT   = "ZT"   // 中通快递
	ExpressCompanyST   = "ST"   // 申通快递
	ExpressCompanyYD   = "YD"   // 韵达快递
	ExpressCompanyYZ   = "YZ"   // 邮政快递
	ExpressCompanyJT   = "JT"   // 极兔速递
	ExpressCompanyDB   = "DB"   // 德邦快递
	ExpressCompanyBS   = "BS"   // 百世快递
	ExpressCompanyJD   = "JD"   // 京东物流
	ExpressCompanyTT   = "TT"   // 天天快递
)

// ============================================
// 出库类型常量 (out_type) - 已使用字符串，保持不变
// ============================================
const (
	OutTypeNormal = "normal" // 发货
	OutTypeOther  = "other"  // 其他
)

// ============================================
// 出库原因常量 (out_reason)
// ============================================
const (
	OutReasonExpired     = "EXPIRED"     // 过期丢弃
	OutReasonDamaged     = "DAMAGED"     // 损坏
	OutReasonUnqualified = "UNQUALIFIED" // 不达标
	OutReasonOther       = "OTHER"       // 其他原因
)

// ============================================
// 外发状态常量 (outsourcing-status)
// ============================================
const (
	OutsourcingStatusPending   = "PENDING"   // 待发出
	OutsourcingStatusSent      = "SENT"      // 已发出
	OutsourcingStatusProducing = "PRODUCING" // 生产中
	OutsourcingStatusReturned  = "RETURNED"  // 已返回
	OutsourcingStatusCancelled = "CANCELLED" // 已撤回
)

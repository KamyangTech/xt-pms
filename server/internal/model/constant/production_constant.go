package constant

// ============================================
// 生产订单状态常量 (production-order-status)
// ============================================
const (
	ProductionOrderStatusPendingSchedule = "PENDING_SCHEDULE" // 待排产
	ProductionOrderStatusProducing       = "PRODUCING"        // 生产中
	ProductionOrderStatusStockedIn       = "STOCKED_IN"       // 已入库（报工时自动入库，良品即合格品）
	ProductionOrderStatusShipped           = "SHIPPED"            // 已出货
	ProductionOrderStatusCompleted         = "COMPLETED"          // 已完成
	ProductionOrderStatusCancelled         = "CANCELLED"          // 已取消
)

// ============================================
// 生产订单状态流转操作常量
// ============================================
const (
	ProductionActionStartProduction    = "start_production"    // 开始生产
	ProductionActionCompleteProduction = "complete_production" // 完成生产
	ProductionActionStockIn            = "stock_in"            // 入库
	ProductionActionShipment           = "shipment"            // 出货
	ProductionActionFinishOrder        = "finish_order"        // 完成订单
	ProductionActionCancelOrder        = "cancel_order"        // 取消订单
)

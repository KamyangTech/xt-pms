/**
 * 字典值枚举定义
 * 用于替代原有的数字ID，使用语义化字符串
 */

// ============================================
// 模具状态 (mold-status)
// ============================================
export enum MoldStatus {
  DESIGNING = 'DESIGNING',           // 设计中
  PROCESSING = 'PROCESSING',         // 加工中
  TESTING = 'TESTING',               // 试模
  TESTED = 'TESTED',                 // 已试模
  PRODUCING = 'PRODUCING',           // 生产中
  REPAIRING = 'REPAIRING',           // 维修
  IN_STOCK = 'IN_STOCK',             // 仓库库存
  DELIVERED = 'DELIVERED',           // 已交付客户
  SCRAPPED = 'SCRAPPED',             // 报废
}

// ============================================
// 模具工单状态 (work-order-status)
// ============================================
export enum WorkOrderStatus {
  NOT_STARTED = 'NOT_STARTED',  // 未开始
  PROCESSING = 'PROCESSING',    // 加工中
  COMPLETED = 'COMPLETED',      // 已完成
  TESTED = 'TESTED',            // 已试模
  STOCKED_IN = 'STOCKED_IN',    // 已入库
  DELIVERED = 'DELIVERED',      // 已交付
}

// ============================================
// 模具订单状态 (mold-order-status)
// ============================================
export enum MoldOrderStatus {
  NOT_STARTED = 'NOT_STARTED',   // 未开始
  PROCESSING = 'PROCESSING',     // 加工中
  PENDING_TEST = 'PENDING_TEST', // 待试模
  TESTED = 'TESTED',             // 已试模
  COMPLETED = 'COMPLETED',       // 已完成
  CANCELLED = 'CANCELLED',       // 已取消
  DELIVERED = 'DELIVERED',       // 已交付
}

// ============================================
// 生产订单状态 (production-order-status)
// ============================================
export enum ProductionOrderStatus {
  PENDING_SCHEDULE = 'PENDING_SCHEDULE',     // 待排产
  PRODUCING = 'PRODUCING',                   // 生产中
  STOCKED_IN = 'STOCKED_IN',                 // 已入库（报工时自动入库）
  SHIPPED = 'SHIPPED',                       // 已出货
  COMPLETED = 'COMPLETED',                   // 已完成
  CANCELLED = 'CANCELLED',                   // 已取消
}

// ============================================
// 模具工艺 (mold-process)
// ============================================
export enum MoldProcess {
  NORMAL = 'NORMAL',         // 普通模具
  HOT_RUNNER = 'HOT_RUNNER', // 热流道
}

// ============================================
// 材质类型 (material-type)
// ============================================
export enum MaterialType {
  STEEL = 'STEEL',               // 钢材
  HARDENED_STEEL = 'HARDENED_STEEL', // 加硬钢材
}

// ============================================
// 模胚情况 (mold-type)
// ============================================
export enum MoldType {
  NEW = 'NEW',   // 全新
  USED = 'USED', // 二手
}

// ============================================
// 订单优先级 (order-priority)
// ============================================
export enum OrderPriority {
  NORMAL = 'NORMAL',         // 普通
  IMPORTANT = 'IMPORTANT',   // 重要
  URGENT = 'URGENT',         // 紧急
  VERY_URGENT = 'VERY_URGENT', // 非常紧急
}

// ============================================
// 生产类型 (production-type)
// ============================================
export enum ProductionType {
  IN_HOUSE = 'IN_HOUSE',     // 本厂生产
  OUTSOURCED = 'OUTSOURCED', // 外发生产
}

// ============================================
// 模具来源 (mold-source)
// ============================================
export enum MoldSource {
  SELF_MADE = 'SELF_MADE',           // 本厂制造模具
  CUSTOMER_PROVIDED = 'CUSTOMER_PROVIDED', // 客户提供模具
}

// ============================================
// 模具位置 (mold-location)
// ============================================
export enum MoldLocation {
  IN_FACTORY = 'IN_FACTORY',     // 厂内
  OUTSOURCING = 'OUTSOURCING',   // 外发中
  AT_CUSTOMER = 'AT_CUSTOMER',   // 客户处
}

// ============================================
// 订单/工单类型 (mold_order_type)
// ============================================
export enum MoldOrderType {
  NEW_MOLD = 'NEW_MOLD',         // 新模
  REPAIR = 'REPAIR',             // 修模
  MODIFICATION = 'MODIFICATION', // 改造
}

// ============================================
// 塑胶原料 (plastic-material)
// ============================================
export enum PlasticMaterial {
  PP = 'PP',     // PP（聚丙烯）
  PE = 'PE',     // PE（聚乙烯）
  ABS = 'ABS',   // ABS
  PS = 'PS',     // PS（聚苯乙烯）
  PC = 'PC',     // PC（聚碳酸酯）
  PVC = 'PVC',   // PVC（聚氯乙烯）
  PMMA = 'PMMA', // PMMA（亚克力）
  PA = 'PA',     // PA（尼龙）
  POM = 'POM',   // POM（聚甲醛）
  PET = 'PET',   // PET（聚酯）
  TPE = 'TPE',   // TPE（热塑性弹性体）
  TPU = 'TPU',   // TPU（热塑性聚氨酯）
}

// ============================================
// 班次管理 (shift-type)
// ============================================
export enum ShiftType {
  DAY_SHIFT = 'DAY_SHIFT',     // 白班
  NIGHT_SHIFT = 'NIGHT_SHIFT', // 夜班
}

// ============================================
// 状态显示标签映射
// ============================================
export const MOLD_STATUS_LABELS: Record<MoldStatus, string> = {
  [MoldStatus.DESIGNING]: '设计中',
  [MoldStatus.PROCESSING]: '加工中',
  [MoldStatus.TESTING]: '试模',
  [MoldStatus.TESTED]: '已试模',
  [MoldStatus.PRODUCING]: '生产中',
  [MoldStatus.REPAIRING]: '维修',
  [MoldStatus.IN_STOCK]: '仓库库存',
  [MoldStatus.DELIVERED]: '已交付客户',
  [MoldStatus.SCRAPPED]: '报废',
};

export const WORK_ORDER_STATUS_LABELS: Record<WorkOrderStatus, string> = {
  [WorkOrderStatus.NOT_STARTED]: '未开始',
  [WorkOrderStatus.PROCESSING]: '加工中',
  [WorkOrderStatus.COMPLETED]: '已完成',
  [WorkOrderStatus.TESTED]: '已试模',
  [WorkOrderStatus.STOCKED_IN]: '已入库',
  [WorkOrderStatus.DELIVERED]: '已交付',
};

export const MOLD_ORDER_STATUS_LABELS: Record<MoldOrderStatus, string> = {
  [MoldOrderStatus.NOT_STARTED]: '未开始',
  [MoldOrderStatus.PROCESSING]: '加工中',
  [MoldOrderStatus.PENDING_TEST]: '待试模',
  [MoldOrderStatus.TESTED]: '已试模',
  [MoldOrderStatus.COMPLETED]: '已完成',
  [MoldOrderStatus.CANCELLED]: '已取消',
  [MoldOrderStatus.DELIVERED]: '已交付',
};

export const PRODUCTION_ORDER_STATUS_LABELS: Record<ProductionOrderStatus, string> = {
  [ProductionOrderStatus.PENDING_SCHEDULE]: '待排产',
  [ProductionOrderStatus.PRODUCING]: '生产中',
  [ProductionOrderStatus.STOCKED_IN]: '已入库',
  [ProductionOrderStatus.SHIPPED]: '已出货',
  [ProductionOrderStatus.COMPLETED]: '已完成',
  [ProductionOrderStatus.CANCELLED]: '已取消',
};

export const MOLD_PROCESS_LABELS: Record<MoldProcess, string> = {
  [MoldProcess.NORMAL]: '普通模具',
  [MoldProcess.HOT_RUNNER]: '热流道',
};

export const MATERIAL_TYPE_LABELS: Record<MaterialType, string> = {
  [MaterialType.STEEL]: '钢材',
  [MaterialType.HARDENED_STEEL]: '加硬钢材',
};

export const MOLD_TYPE_LABELS: Record<MoldType, string> = {
  [MoldType.NEW]: '全新',
  [MoldType.USED]: '二手',
};

export const ORDER_PRIORITY_LABELS: Record<OrderPriority, string> = {
  [OrderPriority.NORMAL]: '普通',
  [OrderPriority.IMPORTANT]: '重要',
  [OrderPriority.URGENT]: '紧急',
  [OrderPriority.VERY_URGENT]: '非常紧急',
};

export const PRODUCTION_TYPE_LABELS: Record<ProductionType, string> = {
  [ProductionType.IN_HOUSE]: '本厂生产',
  [ProductionType.OUTSOURCED]: '外发生产',
};

export const MOLD_SOURCE_LABELS: Record<MoldSource, string> = {
  [MoldSource.SELF_MADE]: '本厂制造模具',
  [MoldSource.CUSTOMER_PROVIDED]: '客户提供模具',
};

export const MOLD_LOCATION_LABELS: Record<MoldLocation, string> = {
  [MoldLocation.IN_FACTORY]: '厂内',
  [MoldLocation.OUTSOURCING]: '外发中',
  [MoldLocation.AT_CUSTOMER]: '客户处',
};

export const MOLD_ORDER_TYPE_LABELS: Record<MoldOrderType, string> = {
  [MoldOrderType.NEW_MOLD]: '新模',
  [MoldOrderType.REPAIR]: '修模',
  [MoldOrderType.MODIFICATION]: '改造',
};

export const PLASTIC_MATERIAL_LABELS: Record<PlasticMaterial, string> = {
  [PlasticMaterial.PP]: 'PP（聚丙烯）',
  [PlasticMaterial.PE]: 'PE（聚乙烯）',
  [PlasticMaterial.ABS]: 'ABS',
  [PlasticMaterial.PS]: 'PS（聚苯乙烯）',
  [PlasticMaterial.PC]: 'PC（聚碳酸酯）',
  [PlasticMaterial.PVC]: 'PVC（聚氯乙烯）',
  [PlasticMaterial.PMMA]: 'PMMA（亚克力）',
  [PlasticMaterial.PA]: 'PA（尼龙）',
  [PlasticMaterial.POM]: 'POM（聚甲醛）',
  [PlasticMaterial.PET]: 'PET（聚酯）',
  [PlasticMaterial.TPE]: 'TPE（热塑性弹性体）',
  [PlasticMaterial.TPU]: 'TPU（热塑性聚氨酯）',
};

export const SHIFT_TYPE_LABELS: Record<ShiftType, string> = {
  [ShiftType.DAY_SHIFT]: '白班',
  [ShiftType.NIGHT_SHIFT]: '夜班',
};

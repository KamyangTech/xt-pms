// 合同状态枚举
export const CONTRACT_STATUS = {
  FAIL: 0,
  AUDIT_PENDING: 1,
  EXEC_PENDING: 2,
  EXECUTING: 3,
  FINISH: 4,
} as const;

// 合同类型枚举
export const CONTRACT_TYPES = {
  MAIN: 0,
  SUB: 1,
  SUPPLEMENT: 2,
} as const;

// 合同收付类型枚举
export const CONTRACT_PAYMENT_TYPES = {
  PAYMENT: 0,
  RECEIPT: 1,
} as const;

// 标签类型
type TagTheme = 'default' | 'success' | 'primary' | 'warning' | 'danger';
// 通知的优先级对应的标签类型
export const NOTIFICATION_TYPES: Map<string, TagTheme> = new Map([
  ['low', 'primary'],
  ['middle', 'warning'],
  ['high', 'danger'],
]);

/** 通知类型 → 中文标签映射 */
export const NOTIFICATION_TYPE_LABELS: Record<string, string> = {
  // 模具
  'mold:work_order_created': '模具工单',
  'mold:work_order_status': '模具工单',
  'mold:order_created': '模具订单',
  'mold:order_updated': '模具订单',
  'mold:try_test': '模具工单',
  'mold:try_test_updated': '模具工单',
  // 'mold:acceptance': 已废弃 - 取消验收步骤
  // 'mold:acceptance_updated': 已废弃
  'mold:delivery': '模具工单',
  'mold:stock_changed': '模具管理',
  // 生产
  'production:order_created': '生产订单',
  'production:order_status': '生产订单',
  'production:stock_in': '生产订单',
  'production:stock_out': '生产订单',
  'production:report_submitted': '生产订单',
  // 外发
  'outsourcing:sent': '外发管理',
  'outsourcing:returned': '外发管理',
  // 综合订单
  'master_order:created': '综合订单',
  // 财务
  'finance:receivable_due': '应收账款',
  'finance:receivable_created': '应收账款',
  'finance:receivable_received': '应收账款',
  'finance:payable_created': '应付账款',
  'finance:payable_paid': '应付账款',
  'finance:expense_created': '财务费用',
  'finance:expense_paid': '财务费用',
  // 产品
  'product:created': '产品管理',
  'product:updated': '产品管理',
  'product:deleted': '产品管理',
  'product:bom_created': '产品管理',
  // 库存
  'stock:in': '库存管理',
  'stock:out': '库存管理',
};

// 通用请求头
export enum ContentTypeEnum {
  Json = 'application/json;charset=UTF-8',
  FormURLEncoded = 'application/x-www-form-urlencoded;charset=UTF-8',
  FormData = 'multipart/form-data;charset=UTF-8',
}

// 工单状态常量
export const WORK_ORDER_STATUS = {
  NOT_STARTED: 1,    // 未开始
  PROCESSING: 2,     // 加工中
  COMPLETED: 3,      // 已完成（待试模）
  TESTED: 4,         // 已试模
  STOCKED_IN: 5,     // 已入库
  DELIVERED: 6,      // 已交付
  CANCELLED: 7,      // 已取消
} as const;

// 订单类型常量
export const ORDER_TYPE = {
  NEW_MOLD: 1,   // 新模
  REPAIR: 2,     // 修模
  MODIFY: 3,     // 改模
} as const;

// 工序状态常量
export const PROCESS_STATUS = {
  PENDING: 0,    // 未完成/待开工
  COMPLETED: 1,  // 已完成
  SKIPPED: 2,    // 已跳过
} as const;

// 外发状态常量
export const OUTSOURCING_STATUS = {
  PENDING: 'PENDING',       // 待发出
  SENT: 'SENT',             // 已发出
  PRODUCING: 'PRODUCING',   // 生产中
  RETURNED: 'RETURNED',     // 已返回
  CANCELLED: 'CANCELLED',   // 已撤回
} as const;

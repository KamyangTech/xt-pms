// 生产订单
export interface ProductionOrder {
  id: number;
  order_code: string;
  customer_id: number;
  customer_name?: string;
  customer?: { cus_name?: string }; // 关联客户对象
  mold?: { mold_code?: string }; // 关联模具对象（GORM Preload加载）
  product_id?: number;           // ✨ 产品ID便捷字段（来自 items[0]）
  product_name: string;

  production_type: string; // IN_HOUSE=本厂生产, OUTSOURCED=外发生产
  mold_source_type: string; // SELF_MADE=本厂制造模具, CUSTOMER_PROVIDED=客户提供模具
  mold_id?: number;
  mold_code?: string;
  external_mold_code?: string; // ✨ 外来模具编号便捷字段（来自 items[0]）
  external_mold_info?: ExternalMoldInfo;

  plan_qty: number;
  completed_qty: number;
  ok_qty: number;
  bad_qty: number;
  mold_shots?: number;
  plan_start_date?: string;
  plan_end_date?: string;
  actual_start_date?: string;
  actual_end_date?: string;

  // 生产工艺信息
  process_type?: 'injection' | 'blow_molding' | 'stamping'; // injection=注塑, blow_molding=吹塑, stamping=冲压
  process_remark?: string;

  // 原料信息
  material_type?: string; // PP, PE, TPE等
  has_metal_parts?: boolean; // 是否包含金属配件

  // 后处理流程
  manual_process?: string; // 手工处理流程
  assembly_process?: string; // 组装流程

  device_id?: string; // 生产设备字典值
  device_name?: string;
  worker_id?: string; // 负责人字典值

  processor_id?: number;
  processor_name?: string;
  outsource_price?: number;
  outsource_total_amount?: number;

  material_cost?: number;
  processing_cost?: number;
  total_cost?: number;
  unit_price?: number;
  total_amount?: number;

  status: string; // PENDING_SCHEDULE, PRODUCING, STOCKED_IN, COMPLETED, CANCELLED
  priority: string; // NORMAL, IMPORTANT, URGENT, VERY_URGENT
  dependency_status?: string; // ✨ 依赖状态：NONE=无依赖, WAITING_MOLD=等待模具, READY=就绪
  master_order_id?: number; // 关联综合订单 ID
  master_order?: { id?: number; order_code?: string }; // 关联综合订单对象（GORM Preload加载）
  shipped_qty?: number; // 已出货数量（从出库记录统计）
  available_action_codes?: string[]; // 可用操作编码列表（由后端计算返回，无需单独请求）

  remark?: string;
  create_time: string;
  update_time: string;
  items?: ProductionOrderItem[];
}

// 生产订单明细
export interface ProductionOrderItem {
  id?: number;
  order_id?: number;
  product_id?: number; // 可选，手动输入产品名称时为空
  product_name?: string;
  mold_id?: number;
  external_mold_code?: string; // 外来模具编号
  mold_code?: string;
  device_id?: number;
  device_name?: string;
  plan_qty: number;
  completed_qty?: number;
  ok_qty?: number;
  bad_qty?: number;
  product_image_url?: string; // 产品图片URL
  plastic_material_id?: number; // 塑胶原料字典值
  plastic_material_name?: string; // 塑胶原料名称
  sort?: number;
  remark?: string;
  dependency_status?: string; // ✨ 新增：明细级别的依赖状态
  product?: { id?: number; product_name?: string; pro_name?: string }; // 关联产品对象
  mold?: { mold_code?: string }; // 关联模具对象
  // ✨ 款项配置字段（金额计算）
  unit_price?: number;      // 单价（元）
  packaging_fee?: number;   // 包装费（元）
  freight_fee?: number;     // 运费（元）
  outsource_price?: number; // 外发单价（元）
}

// 外来模具信息
export interface ExternalMoldInfo {
  mold_code: string;
  mold_name: string;
  customer_name: string;
  cavity_count?: number;
  spec?: string;
  material?: string;
  weight?: string;
  remark?: string;
}

// 生产报工
export interface ProductionReport {
  id: number;
  order_id: number;
  item_id?: number; // ✨ 新增：订单明细项ID（支持按明细报工）
  schedule_id?: number;
  report_date: string;
  shift?: string; // 班次字典值（语义化字符串：DAY_SHIFT, NIGHT_SHIFT）
  ok_qty: number;
  bad_qty: number;
  scrap_qty: number;
  device_id?: string;
  worker_id?: string;
  work_hours?: number;
  defect_reason?: string;
  remark?: string;
}

// 外发记录
export interface OutsourcingRecord {
  id: number;
  record_code: string;
  order_id?: number;
  work_order_id?: number;
  outsource_type?: 1 | 2; // 1=生产外发, 2=模具外发
  processor_id: number;
  processor_name?: string;
  mold_id?: number;
  external_mold_info?: ExternalMoldInfo;
  outsource_date?: string;
  expected_return_date?: string;
  actual_return_date?: string;
  quantity: number;
  unit_price?: number;
  total_amount?: number;
  paid_amount?: number;
  logistics_company?: string;
  tracking_no?: string;
  delivery_address?: string;
  receiver_name?: string;
  receiver_phone?: string;
  status: 1 | 2 | 3 | 4 | 5 | 6;
  remark?: string;
  share_token?: string;
}

// 外发报工
export interface OutsourcingReport {
  id: number;
  outsourcing_id: number;
  report_date?: string;
  ok_qty: number;
  bad_qty: number;
  defect_reason?: string;
  photos?: string[];
  remark?: string;
}

// 模具使用记录
export interface MoldUsageRecord {
  id: number;
  mold_id: number;
  production_order_id: number;
  usage_date?: string;
  total_shots: number;
  ok_qty: number;
  bad_qty: number;
  start_date?: string;
  end_date?: string;
  remark?: string;
  create_time: string;
  update_time: string;
}

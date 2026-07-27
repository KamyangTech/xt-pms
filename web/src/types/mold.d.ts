// 模具档案
export interface Mold {
  id: number;
  mold_code: string;
  mold_name?: string;
  customer_id: number;
  customer_name?: string;
  ownership_type?: 1 | 2 | 3; // 1=自有, 2=客户资产, 3=第三方
  latest_order_id?: number;
  latest_work_order_id?: number;
  total_orders: number;
  total_work_orders: number;
  last_delivery_time?: string;
  last_usage_date?: string;
  total_shots: number;
  cavity_count?: number;
  production_date?: string;
  process_status?: 1 | 2 | 3 | 4; // 1=制造中, 2=啤货, 3=在仓, 4=废弃
  current_location?: 1 | 2 | 3; // 1=厂内, 2=外发中, 3=客户处
  life_cycle_status?: 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9; // 1=未开始, 2=加工中, 3=待试模, 4=已试模, 5=已完成, 6=维修, 7=仓库库存, 8=已交付客户, 9=报废
  mold_type?: 1 | 2; // 1=全新, 2=二手
  mold_process?: 1 | 2; // 1=普通, 2=热流道
  material_type?: 1 | 2; // 1=钢材, 2=加硬钢材
  remark?: string;
  create_time: string;
  update_time: string;
  images?: MoldFile[];
}

// 模具图片（统一 File 类型，前端兼容使用）
export interface MoldFile {
  id: string;
  url: string;
  file_name: string;
  file_size: number;
  file_type: string;
  is_primary: number;
  sort: number;
  create_time: string;
}

// 模具订单
export interface MoldOrder {
  id: number;
  order_code: string;
  customer_id: number;
  customer_name?: string;
  order_type: 1 | 2 | 3; // 1=新模订单, 2=修模订单, 3=改造订单
  mold_id?: number;
  plan_start_date?: string;
  plan_end_date?: string;
  total_amount: number;
  additional_amount?: number;
  status: string; // 字符串状态码："NOT_STARTED" | "PROCESSING" | "PENDING_TEST" | "TESTED" | "COMPLETED" | "CANCELLED" | "DELIVERED"
  deliver_time?: string;
  remark?: string;
  create_time: string;
  update_time: string;
  items?: MoldOrderItem[];
  mold_codes?: string[];
}

// 模具订单明细
export interface MoldOrderItem {
  id: number;
  order_id: number;
  mold_id: number;
  amount: number;
  sort: number;
  remark?: string;
  create_time: string;
  update_time: string;
}

// 模具工单
export interface MoldWorkOrder {
  id: number;
  work_code: string;
  mold_order_id?: number;
  order_type: 1 | 2 | 3; // 1=新模, 2=修模, 3=改造
  mold_id?: number;
  plan_start?: string;
  plan_end?: string;
  actual_start?: string;
  actual_end?: string;
  progress: number;
  status: 1 | 2 | 3 | 4 | 5 | 6 | 7; // 1=未开始, 2=加工中, 3=已完成, 4=已试模, 5=已入库, 6=已交付, 7=已取消
  warehouse_id?: number;
  deliver_time?: string;
  actual_deliver_time?: string;
  original_mold_status?: number;
  remark?: string;
  is_cancelled: 0 | 1;
  cancel_time?: string;
  cancel_reason?: string;
  cancel_user_id?: number;
  create_time: string;
  update_time: string;
  processes?: WorkOrderProcessInfo[];
  is_locked?: boolean;
}

// 工单工序信息
export interface WorkOrderProcessInfo {
  process_template_id: number;
  process_name: string;
  is_completed: 0 | 1 | 2;
  completed_at?: string;
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

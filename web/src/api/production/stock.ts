import { request } from '@/utils/request';

export function getProductStockList(params?: {
  page?: number;
  page_size?: number;
  mold_code?: string;
  product_name?: string;
}) {
  return request.get<{ list: any[]; total: number }>({
    url: '/production/stock/product-list',
    params,
  });
}

export function deleteStock(data: {
  mold_id?: number;
  external_mold_code?: string;
  product_id?: number;
}) {
  return request.post({
    url: '/production/stock/delete',
    data,
  });
}

export function getStockRecordsByMoldId(
  moldId: number,
  params?: {
    page?: number;
    page_size?: number;
    sort_order?: string; // asc 升序, desc 降序
    status_filters?: string; // 多选状态筛选，逗号分隔
  },
) {
  return request.get<{ list: any[]; total: number; summary?: any }>({
    url: `/production/stock/${moldId}/records`,
    params,
  });
}

export function getStockRecordsByExternalMoldCode(
  externalMoldCode: string,
  params?: {
    page?: number;
    page_size?: number;
    sort_order?: string; // asc 升序, desc 降序
    status_filters?: string; // 多选状态筛选，逗号分隔
  },
) {
  return request.get<{ list: any[]; total: number; summary?: any }>({
    url: `/production/stock/ext/${encodeURIComponent(externalMoldCode)}/records`,
    params,
  });
}

export function getStockInList(params?: {
  page?: number;
  page_size?: number;
  order_code?: string;
  product_name?: string;
  stock_code?: string;
}) {
  return request.get<{ list: any[]; total: number }>({
    url: '/production/stock-in/list',
    params,
  });
}

export function getStockInById(id: number) {
  return request.get<any>({
    url: `/production/stock-in/${id}`,
  });
}

export function updateStockIn(data: any) {
  return request.put({
    url: '/production/stock-in',
    data,
  });
}

// 生产订单入库（报工后自动入库）
export function stockInOrder(data: {
  order_id?: number;
  item_id?: number;
  warehouse_id: number;
  mold_id?: number;
  external_mold_code?: string;
  product_id?: number;
  product_name?: string;
  quantity: number;
  batch_no?: string;
  remark?: string;
}) {
  return request.post({
    url: '/production/stock-in/order',
    data,
  });
}

export function getStockStatistics() {
  return request.get<any>({
    url: '/production/stock/statistics',
  });
}

export function getStockTrend(days: number = 30) {
  return request.get<any[]>({
    url: '/production/stock/trend',
    params: { days },
  });
}

// 直接创建库存（用于历史库存导入或手动调整）
export function createStockDirectly(data: {
  mold_id?: number;
  external_mold_code?: string;
  product_id?: number; // 产品ID，用于区分同一模具不同产品的库存
  warehouse_id: number;
  quantity: number;
  safe_stock?: number;
  remark?: string;
}) {
  return request.post({
    url: '/production/stock/direct',
    data,
  });
}

// 盘点库存（矫正实际库存数量）
export function checkStock(data: {
  mold_id?: number;
  external_mold_code?: string;
  product_id?: number;
  warehouse_id: number;
  actual_quantity: number;
  remark?: string;
}) {
  return request.post<any>({
    url: '/production/stock/check',
    data,
  });
}

// 生产订单出库
export function stockOutOrder(data: {
  order_id?: number;
  warehouse_id: number;
  mold_id?: number;
  external_mold_code?: string;
  product_id?: number; // 产品ID，用于产品维度库存管理
  product_name: string;
  quantity: number;
  batch_no?: string;
  remark?: string;
  customer_name?: string; // 接收客户名称
  delivery_date?: string; // 发货日期 YYYY-MM-DD
  delivery_method?: string; // 发货方式字典值
  express_company?: string; // 快递公司字典值
  express_no?: string; // 快递单号
  photos?: string; // 发货照片(JSON数组)
  total_weight?: number; // 总重量(kg)
  driver_phone?: string; // 司机电话(货拉拉用)
  driver_plate?: string; // 车牌号(货拉拉用)
  out_type?: string; // 出库类型(normal=正常发货, other=其他原因出库)
  out_reason?: string; // 出库原因(expired=过期丢弃, damaged=损坏, unqualified=不达标, other=其他原因)
  out_reason_remark?: string; // 出库原因备注
}) {
  return request.post({
    url: '/production/stock-out',
    data,
  });
}

export function getStockOutList(params?: {
  page?: number;
  page_size?: number;
  order_code?: string;
  product_name?: string;
  stock_code?: string;
}) {
  return request.get<{ list: any[]; total: number }>({
    url: '/production/stock-out/list',
    params,
  });
}

export function getStockOutById(id: number) {
  return request.get<any>({
    url: `/production/stock-out/${id}`,
  });
}

export function updateStockOut(
  id: number,
  data: {
    stock_out_date?: string;
    customer_name?: string;
    delivery_date?: string;
    delivery_method?: string;
    express_company?: string;
    express_no?: string;
    photos?: string;
    total_weight?: number;
    driver_phone?: string;
    driver_plate?: string;
    remark?: string;
  },
) {
  return request.put({
    url: `/production/stock-out/${id}`,
    data,
  });
}

export function getStockOutRecordsByMoldId(
  moldId: number,
  params?: {
    page?: number;
    page_size?: number;
  },
) {
  return request.get<{ list: any[]; total: number }>({
    url: `/production/stock-out/mold/${moldId}/records`,
    params,
  });
}

export function getStockOutRecordsByExternalMoldCode(
  externalMoldCode: string,
  params?: {
    page?: number;
    page_size?: number;
  },
) {
  return request.get<{ list: any[]; total: number }>({
    url: `/production/stock-out/ext/${encodeURIComponent(externalMoldCode)}/records`,
    params,
  });
}

// 识别快递公司（基于快递单号）
export function recognizeExpressCompany(expressNo: string) {
  return request.get<any>({
    url: '/production/stock-out/recognize-express',
    params: { express_no: expressNo },
  });
}

// 获取当前库存数量（用于出库验证）
export function getCurrentStock(data: {
  mold_id?: number;
  external_mold_code?: string;
  product_id?: number;
  product_name?: string;
  warehouse_id: number;
}) {
  return request.post<any>({
    url: '/production/stock/current',
    data,
  });
}

// 获取用于打印的出库记录列表（支持多个ID）
export function getStockOutListForPrint(ids: number[]) {
  return request.post<any>({
    url: '/production/stock-out/list-for-print',
    data: { ids },
  });
}

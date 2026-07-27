import { request } from '@/utils/request';
import type {
  ProductionOrder,
  ProductionReport,
  OutsourcingRecord,
  OutsourcingReport,
} from '@/types/production';

export function getProductionOrders(params?: {
  page?: number;
  page_size?: number;
  status?: string;
  customer_id?: number;
  product_name?: string;
  production_type?: string;
  keyword?: string;
}) {
  return request.get<{ list: ProductionOrder[]; total: number }>({
    url: '/production/orders',
    params,
  });
}

export function getProductionOrderById(id: number) {
  return request.get<ProductionOrder>({ url: `/production/orders/${id}` });
}

export function createProductionOrder(
  data: Omit<ProductionOrder, 'id' | 'create_time' | 'update_time'>
) {
  return request.post<{ id: number }>({ url: '/production/order', data });
}

export function updateProductionOrder(data: Partial<ProductionOrder>) {
  return request.put({ url: '/production/order', data });
}

// 取消订单
export function cancelProductionOrder(id: number, reason?: string) {
  return request.delete({ url: `/production/order/${id}`, data: { reason } });
}

// 开始生产
export function startProduction(id: number) {
  return request.post({ url: `/production/order/${id}/start` });
}

// 完成生产
export function completeProduction(id: number) {
  return request.post({ url: `/production/order/${id}/complete` });
}

// 入库
export function stockInOrder(data: { 
  order_id: number; 
  item_id?: number; // ✨ 新增：明细项ID
  warehouse_id: number; 
  mold_id?: number; 
  external_mold_code?: string; 
  product_id?: number; 
  product_name?: string; // 产品名称
  quantity: number; 
  batch_no?: string; 
  remark?: string 
}) {
  return request.post({ url: '/production/order/stock-in', data });
}

// 出货
export function shipmentOrder(id: number, data?: { logistics_company?: string; tracking_no?: string; remark?: string }) {
  return request.post({ url: `/production/order/${id}/shipment`, data });
}

// 完成订单
export function completeOrder(id: number) {
  return request.post({ url: `/production/order/${id}/finish` });
}

// 补发（已完成订单再次出货）
export function replenishShipment(id: number, data?: { remark?: string }) {
  return request.post({ url: `/production/order/${id}/replenish`, data });
}

// 完成单个明细项
export function completeOrderItem(itemId: number) {
  return request.post({ url: `/production/order/item/${itemId}/finish` });
}

export function getAvailableActions(id: number) {
  return request.get<Array<{
    id: number;
    action_name: string;
    action_code: string;
    from_status_value: string;
    to_status_value: string;
    remark?: string;
  }>>({ url: `/production/order/${id}/actions` });
}

// 提交报工
export function submitReport(data: Omit<ProductionReport, 'id'>) {
  return request.post<{ id: number }>({ url: '/production/report', data });
}

export function getReports(orderId: number) {
  return request.get<ProductionReport[]>({
    url: '/production/reports',
    params: { order_id: orderId },
  });
}

export function createOutsourcingRecord(
  data: Omit<OutsourcingRecord, 'id' | 'record_code'>
) {
  return request.post<{ id: number }>({ url: '/production/outsourcing', data });
}

// 确认外发
export function sendOutsourcing(id: number) {
  return request.post({ url: `/production/outsourcing/${id}/send` });
}

// 回收确认
export function receiveOutsourcing(id: number) {
  return request.post({ url: `/production/outsourcing/${id}/receive` });
}

export function addOutsourcingReport(data: Omit<OutsourcingReport, 'id'>) {
  return request.post<{ id: number }>({
    url: '/production/outsourcing/report',
    data,
  });
}

export function getOutsourcingRecords(params?: {
  status?: 1 | 2 | 3 | 4 | 5 | 6;
  processor_id?: number;
}) {
  return request.get<OutsourcingRecord[]>({
    url: '/production/outsourcing',
    params,
  });
}

export function generateShareToken(id: number) {
  return request.post<{ share_token: string }>({ url: `/production/outsourcing/${id}/share-token` });
}

// 重新生成分享链接token
export function regenerateShareToken(id: number) {
  return request.post<{ share_token: string }>({ url: `/production/outsourcing/${id}/regenerate-token` });
}

export function getOrderStatistics() {
  return request.get<{
    total_orders: number;
    pending_count: number;
    in_progress_count: number;
    completed_count: number;
    cancelled_count: number;
  }>({ url: '/production/statistics' });
}

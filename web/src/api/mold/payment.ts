/**
 * 模具订单款项 API
 */
import { request } from '@/utils/request';

export interface PaymentRecord {
  id: number;
  order_id: number;
  work_order_id?: number;
  payment_type: string;
  source_type?: string; // 款项归属：mold=模具款, outsource=外发应付款
  contract_amount: number;
  actual_amount: number;
  should_amount: number;
  payment_date?: string;
  due_date?: string; // 到期日期
  status: number;
  remark?: string;
  create_time?: string;
  update_time?: string;
}

export interface PaymentStats {
  contract_total: number;
  actual_total: number;
  should_total: number;
  pending_total: number;
}

export function getOrderPayments(orderId: number) {
  return request.get({
    url: `/mold-order/${orderId}/payments`,
  });
}

export function getPaymentById(id: number) {
  return request.get({
    url: `/mold-order/payments/${id}`,
  });
}

// 根据工单ID获取款项列表
export function getPaymentsByWorkOrderId(workOrderId: number) {
  return request.get({
    url: `/mold-order/payments/by-work-order/${workOrderId}`,
  });
}

// 新增款项记录
export function createPayment(data: Partial<PaymentRecord>) {
  return request.post({
    url: '/mold-order/payments',
    data,
  });
}

export function updatePayment(id: number, data: Partial<PaymentRecord>) {
  return request.put({
    url: `/mold-order/payments/${id}`,
    data,
  });
}

export function deletePayment(id: number) {
  return request.delete({
    url: `/mold-order/payments/${id}`,
  });
}

export function getPaymentStats(orderId: number) {
  return request.get({
    url: `/mold-order/${orderId}/payment-stats`,
  });
}

// 批量替换工单的所有款项记录（事务保护）
export function replacePaymentsByWorkOrder(data: {
  order_id: number;
  work_order_id: number;
  items: Array<{
    payment_type: string;
    contract_amount: number;
    should_amount: number;
    actual_amount: number;
    status: number;
    remark?: string;
  }>;
}) {
  return request.post({
    url: '/mold-order/payments/replace-by-work-order',
    data,
  });
}

import { request } from '@/utils/request';

export interface ReceivableParams {
  page?: number;
  page_size?: number;
  customer_id?: number;
  status?: string;
  /** 款项来源多选：mold=模具, production=生产；空数组表示不过滤 */
  source_types?: string[];
  /** 款项类型多选：total=全款, deposit=首款, interim=中期款, final=尾款；空数组表示不过滤 */
  payment_types?: string[];
}

export interface ReceivableRecord {
  id: number;
  customer_id: number;
  order_id?: number;
  order_type?: string;
  amount: number;
  received_amount: number;
  discount_amount: number;
  balance: number;
  due_date?: string;
  aging_days?: number;
  aging_category?: string;
  status: string;
  remark?: string;
  create_time?: string;
  update_time?: string;
  customer?: {
    id: number;
    cus_name: string;
  };
  order_code?: string;
  mold_code?: string;
  product_name?: string;
  payment_type?: string;
  is_locked?: boolean;
  unlock_rule?: string;
}

/**
 * 获取应收账款列表
 */
export function getReceivableList(params: ReceivableParams) {
  return request.get({
    url: '/finance/receivables',
    params,
  });
}

/**
 * 按订单ID和类型获取应收账款
 */
export function getReceivablesByOrder(orderId: number, orderType: string) {
  return request.get({
    url: `/finance/receivables/by-order/${orderId}/${orderType}`,
  });
}

/**
 * 获取应收账款详情
 */
export function getReceivableDetail(id: number) {
  return request.get({
    url: `/finance/receivables/${id}`,
  });
}

/**
 * 创建应收账款
 */
export function createReceivable(data: Partial<ReceivableRecord>) {
  return request.post({
    url: '/finance/receivables',
    data,
  });
}

/**
 * 更新应收账款
 */
export function updateReceivable(id: number, data: Partial<ReceivableRecord>) {
  return request.put({
    url: `/finance/receivables/${id}`,
    data,
  });
}

/**
 * 删除应收账款
 */
export function deleteReceivable(id: number) {
  return request.delete({
    url: `/finance/receivables/${id}`,
  });
}

/**
 * 收款
 */
export function receivePayment(id: number, data: {
  amount: number;
  discount_amount?: number;
  payment_date?: string;
  payment_method?: string;
  transaction_no?: string;
  receiving_account?: string;
  remark?: string;
  receipt_file_id?: string;
  invoice_file_id?: string;
}) {
  return request.post({
    url: `/finance/receivables/${id}/receive`,
    data,
  });
}

/**
 * 获取逾期应收账款列表
 */
export function getOverdueReceivables() {
  return request.get({
    url: '/finance/receivables/overdue',
  });
}

/**
 * 获取账龄统计
 */
export function getAgingStatistics() {
  return request.get({
    url: '/finance/receivables/aging-stats',
  });
}

/**
 * 获取收款记录列表
 */
export function getReceivablePayments(id: number) {
  return request.get({
    url: `/finance/receivables/${id}/payments`,
  });
}

import { request } from '@/utils/request';

export interface PayableParams {
  page?: number;
  page_size?: number;
  supplier_id?: number;
  status?: string;
}

export interface PayableRecord {
  id: number;
  supplier_id: number;
  order_id?: number;
  order_type?: string;
  payment_type?: string;
  source_type?: string;
  order_code?: string;
  amount: number;
  paid_amount: number;
  discount_amount: number;
  balance: number;
  due_date?: string;
  status: string;
  remark?: string;
  create_time?: string;
  update_time?: string;
  supplier?: {
    id: number;
    sup_name: string;
  };
  processor?: {
    id: number;
    processor_name: string;
  };
}

/**
 * 获取应付账款列表
 */
export function getPayableList(params: PayableParams) {
  return request.get({
    url: '/finance/payables',
    params,
  });
}

/**
 * 获取应付账款详情
 */
export function getPayableDetail(id: number) {
  return request.get({
    url: `/finance/payables/${id}`,
  });
}

/**
 * 创建应付账款
 */
export function createPayable(data: Partial<PayableRecord>) {
  return request.post({
    url: '/finance/payables',
    data,
  });
}

/**
 * 更新应付账款
 */
export function updatePayable(id: number, data: Partial<PayableRecord>) {
  return request.put({
    url: `/finance/payables/${id}`,
    data,
  });
}

/**
 * 删除应付账款
 */
export function deletePayable(id: number) {
  return request.delete({
    url: `/finance/payables/${id}`,
  });
}

/**
 * 付款
 */
export function payPayable(id: number, data: {
  amount: number;
  discount_amount?: number;
  payment_date?: string;
  payment_method?: string;
  transaction_no?: string;
  payment_account?: string;
  receiver_account?: string;
  remark?: string;
  receipt_file_id?: string;
  invoice_file_id?: string;
}) {
  return request.post({
    url: `/finance/payables/${id}/pay`,
    data,
  });
}

/**
 * 获取逾期应付账款列表
 */
export function getOverduePayables() {
  return request.get({
    url: '/finance/payables/overdue',
  });
}

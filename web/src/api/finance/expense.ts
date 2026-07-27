import { request } from '@/utils/request';

export interface ExpenseParams {
  page?: number;
  page_size?: number;
  expense_type?: string;
  related_order_id?: number;
  related_order_type?: string;
  start_date?: string;
  end_date?: string;
}

export interface ExpenseRecord {
  id: number;
  expense_type: string;
  related_order_id?: number;
  related_order_type?: string;
  delivery_id?: number;
  amount: number;
  paid_amount: number;
  balance: number;
  status: string;
  payment_date?: string;
  payer?: string;
  receiver?: string;
  remark?: string;
  created_by?: number;
  create_time?: string;
  update_time?: string;
}

/**
 * 获取费用列表
 */
export function getExpenseList(params: ExpenseParams) {
  return request.get({
    url: '/finance/expenses',
    params,
  });
}

/**
 * 获取费用汇总统计
 */
export function getExpenseStats() {
  return request.get({
    url: '/finance/expenses/stats',
  });
}

/**
 * 获取费用详情
 */
export function getExpenseDetail(id: number) {
  return request.get({
    url: `/finance/expenses/${id}`,
  });
}

/**
 * 创建费用记录
 */
export function createExpense(data: Partial<ExpenseRecord>) {
  return request.post({
    url: '/finance/expenses',
    data,
  });
}

/**
 * 更新费用
 */
export function updateExpense(id: number, data: Partial<ExpenseRecord>) {
  return request.put({
    url: `/finance/expenses/${id}`,
    data,
  });
}

/**
 * 删除费用
 */
export function deleteExpense(id: number) {
  return request.delete({
    url: `/finance/expenses/${id}`,
  });
}

/**
 * 支付费用（含详细记录）
 */
export function payExpense(id: number, data: {
  amount: number;
  payment_date?: string;
  payment_method?: string;
  transaction_no?: string;
  payment_account?: string;
  receiver_account?: string;
  remark?: string;
}) {
  return request.post({
    url: `/finance/expenses/${id}/pay`,
    data,
  });
}

/**
 * 获取费用的支付记录列表
 */
export function getExpensePayments(expenseId: number) {
  return request.get({
    url: `/finance/expenses/${expenseId}/payments`,
  });
}

/**
 * 按交付记录获取费用列表
 */
export function getExpensesByDelivery(deliveryId: number) {
  return request.get({
    url: `/finance/expenses/delivery/${deliveryId}`,
  });
}

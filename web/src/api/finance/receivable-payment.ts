import { request } from '@/utils/request';

/**
 * 收款记录接口
 */

export function getReceivablePayments(receivableId: number) {
  return request.get({
    url: `/finance/receivables/${receivableId}/payments`,
  });
}

export function updateReceivablePayment(paymentId: number, data: {
  payment_date?: string;
  payment_method?: string;
  transaction_no?: string;
  receiving_account?: string;
  remark?: string;
  receipt_file_id?: string;
  invoice_file_id?: string;
}) {
  return request.put({
    url: `/finance/receivables/payments/${paymentId}`,
    data,
  });
}

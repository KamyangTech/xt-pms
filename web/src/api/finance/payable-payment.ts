import { request } from '@/utils/request';

/**
 * 付款记录接口
 */

export function getPayablePayments(payableId: number) {
  return request.get({
    url: `/finance/payables/${payableId}/payments`,
  });
}

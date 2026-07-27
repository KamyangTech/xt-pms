import { request } from '@/utils/request';
import type { RequestOptions } from '@/types/axios.d';

export function createCombinedOrder(data: any) {
  return request.post({ url: '/master-orders', data });
}

export function getMasterOrderList(params?: any) {
  return request.get({ url: '/master-orders', params });
}

export function getMasterOrderDetail(id: number, options?: RequestOptions) {
  return request.get({
    url: `/master-orders/${id}`,
    requestOptions: options,
  });
}

// 编辑综合订单（交货日期+备注）
export function editMasterOrder(id: number, data: { plan_end_date?: string; remark: string }) {
  return request.put({ url: `/master-orders/${id}`, data });
}

// 取消综合订单
export function cancelMasterOrder(id: number, data: { reason: string }) {
  return request.post({ url: `/master-orders/${id}/cancel`, data });
}

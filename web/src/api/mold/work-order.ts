import { request } from '@/utils/request';

/**
 * 模具工单管理 API
 */

export function getMoldWorkOrderList(params?: any) {
  return request.get({
    url: '/mold/work-orders',
    params,
  });
}

// 根据订单ID获取工单列表
export function getMoldWorkOrdersByOrderId(orderId: number) {
  return request.get({
    url: '/mold/work-orders',
    params: { mold_order_id: orderId },
  });
}

// 根据ID获取工单详情
export function getMoldWorkOrderById(id: number) {
  return request.get({
    url: `/mold/work-order/${id}`,
  });
}

export function createMoldWorkOrder(data: any) {
  return request.post({
    url: '/mold/work-order',
    data,
  });
}

export function updateMoldWorkOrder(data: any) {
  return request.put({
    url: '/mold/work-order',
    data,
  });
}

export function deleteMoldWorkOrder(id: number) {
  return request.delete({
    url: `/mold/work-order/${id}`,
  });
}

// 取消模具工单
export function cancelMoldWorkOrder(data: { id: number; reason?: string }) {
  return request.post({
    url: '/mold/work-order/cancel',
    data,
  });
}

import { request } from '@/utils/request';

/**
 * 模具档案管理 API
 */

export function getMoldList(params?: any) {
  return request.get({
    url: '/mold',
    params,
  });
}

// 根据 ID 获取模具
export function getMoldById(id: number) {
  return request.get({
    url: `/mold/${id}`,
  });
}

export function createMold(data: any) {
  return request.post({
    url: '/mold',
    data,
  });
}

export function updateMold(data: any) {
  return request.put({
    url: '/mold',
    data,
  });
}

/** @deprecated 后端已禁用物理删除；请使用 updateMold 将 life_cycle_status 置为 SCRAPPED */
export function deleteMold(id: number) {
  return request.delete({
    url: `/mold/${id}`,
  });
}

export function getMoldStatistics() {
  return request.get({
    url: '/mold/statistics',
  });
}

/**
 * 模具订单管理 API
 */

export function getMoldOrderList(params?: any) {
  return request.get({
    url: '/mold/orders',
    params,
  });
}

// 根据ID获取模具订单详情
export function getMoldOrderById(id: number) {
  return request.get({
    url: `/mold/order/${id}`,
  });
}

export function createMoldOrder(data: any) {
  return request.post({
    url: '/mold/order',
    data,
  });
}

export function updateMoldOrder(data: any) {
  return request.put({
    url: '/mold/order',
    data,
  });
}

export function deleteMoldOrder(id: number) {
  return request.delete({
    url: `/mold/order/${id}`,
  });
}

/**
 * 模具工单管理 API
 */

export function getMoldWorkOrderList(params?: any) {
  return request.get({
    url: '/mold/work-orders',
    params,
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

/**
 * 工单工序管理 API
 */

export function getWorkOrderProcesses(workOrderId: number) {
  return request.get({
    url: `/mold/work-order/${workOrderId}/processes`,
  });
}

/**
 * 模具试模管理 API
 */

export function getMoldTryTests(params?: any) {
  return request.get({
    url: '/mold/try-tests',
    params,
  });
}

export function createMoldTryTest(data: any) {
  return request.post({
    url: '/mold/try-test',
    data,
  });
}

export function updateMoldTryTest(id: number, data: any) {
  return request.put({
    url: `/mold/try-test/${id}`,
    data,
  });
}

export function deleteMoldTryTest(id: number) {
  return request.delete({
    url: `/mold/try-test/${id}`,
  });
}

/**
 * 模具交付管理 API
 */

export function getMoldProductionItems(moldId: number) {
  return request.get({
    url: `/production/mold/${moldId}/production-items`,
  });
}

/**
 * 模具交付管理 API
 */

export function getMoldDeliveries(params?: any) {
  return request.get({
    url: '/mold/deliveries',
    params,
  });
}

export function createMoldDelivery(data: any) {
  return request.post({
    url: '/mold/delivery',
    data,
  });
}

/**
 * 模具库存管理 API
 */

export function getMoldStocks(params?: any) {
  return request.get({
    url: '/mold/stocks',
    params,
  });
}

export function createMoldStock(data: any) {
  return request.post({
    url: '/mold/stock',
    data,
  });
}

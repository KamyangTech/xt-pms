import { request } from '@/utils/request';

export function getWarehouseList(params?: any) {
  return request.get({ url: '/basic/warehouse', params });
}

export function getWarehouseById(id: number) {
  return request.get({ url: `/basic/warehouse/${id}` });
}

export function createWarehouse(data: any) {
  return request.post({ url: '/basic/warehouse', data });
}

export function updateWarehouse(data: any) {
  return request.put({ url: '/basic/warehouse', data });
}

export function deleteWarehouse(id: number) {
  return request.delete({ url: `/basic/warehouse/${id}` });
}

// 批量删除仓库
export function batchDeleteWarehouses(ids: number[]) {
  return request.post({ url: '/basic/warehouse/batch-delete', data: { ids } });
}

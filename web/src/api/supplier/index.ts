import { request } from '@/utils/request';

export function getSupplierList(params?: any) {
  return request.get({ url: '/basic/supplier', params });
}

export function createSupplier(data: any) {
  return request.post({ url: '/basic/supplier', data });
}

export function updateSupplier(data: any) {
  return request.put({ url: '/basic/supplier', data });
}

export function deleteSupplier(id: number) {
  return request.delete({ url: `/basic/supplier/${id}` });
}

// 批量删除供应商
export function batchDeleteSuppliers(ids: number[]) {
  return request.post({ url: '/basic/supplier/batch-delete', data: { ids } });
}

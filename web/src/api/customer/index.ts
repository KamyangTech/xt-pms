import { request } from '@/utils/request';

export function getCustomerList(params?: any) {
  return request.get({ url: '/basic/customer', params });
}

export function createCustomer(data: any) {
  return request.post({ url: '/basic/customer', data });
}

export function updateCustomer(data: any) {
  return request.put({ url: '/basic/customer', data });
}

export function deleteCustomer(id: number) {
  return request.delete({ url: `/basic/customer/${id}` });
}

// 批量删除客户
export function batchDeleteCustomers(ids: number[]) {
  return request.post({ url: '/basic/customer/batch-delete', data: { ids } });
}

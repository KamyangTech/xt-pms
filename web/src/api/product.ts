import { request } from '@/utils/request';

export function getProductList(params: any) {
  return request.get({ url: '/product', params });
}

export function generateProductCode() {
  return request.get({ url: '/product/generate-code' });
}

export function createProduct(data: any) {
  return request.post({ url: '/product', data });
}

export function updateProduct(data: any) {
  return request.put({ url: '/product', data });
}

export function getProductById(id: number) {
  return request.get({ url: `/product/${id}` });
}

export function deleteProduct(id: number) {
  return request.delete({ url: `/product/${id}` });
}

// 从生产记录同步产品
export function syncProducts() {
  return request.post({ url: '/product/sync' });
}

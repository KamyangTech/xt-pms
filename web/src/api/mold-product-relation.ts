import { request } from '@/utils/request';

// 根据产品ID查询关联的模具
export function getRelationsByProductID(productId: number) {
  return request.get({ url: '/mold-product-relation/by-product', params: { product_id: productId } });
}

// 根据模具ID查询关联的产品
export function getRelationsByMoldID(moldId: number) {
  return request.get({ url: '/mold-product-relation/by-mold', params: { mold_id: moldId } });
}

// 根据订单明细ID查询关联
export function getRelationsByItemID(itemId: number) {
  return request.get({ url: '/mold-product-relation/by-item', params: { item_id: itemId } });
}

export function createRelation(data: any) {
  return request.post({ url: '/mold-product-relation', data });
}

export function deleteRelation(id: number) {
  return request.delete({ url: `/mold-product-relation/${id}` });
}

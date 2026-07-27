import { request } from '@/utils/request';

/**
 * 公共引用数据 API
 * 用于跨模块下拉/选择器数据获取，仅需 JWT 认证
 */

export function getCustomerList(params?: any) {
  return request.get({ url: '/reference/customers', params });
}

export function getCustomerByID(id: number) {
  return request.get({ url: `/reference/customers/${id}` });
}

export function getWarehouseList(params?: any) {
  return request.get({ url: '/reference/warehouses', params });
}

export function getSupplierList(params?: any) {
  return request.get({ url: '/reference/suppliers', params });
}

export function getProcessorList(params?: any) {
  return request.get({ url: '/reference/processors', params });
}

export function getProcessList(params?: any) {
  return request.get({ url: '/reference/processes', params });
}

export function getActiveProcessTemplates(params?: any) {
  return request.get({ url: '/reference/process-templates/active', params });
}

export function getProcessTemplateList(params?: any) {
  return request.get({ url: '/reference/process-templates', params });
}

export function getDictDataByType(params?: any) {
  return request.get({ url: '/reference/dict-data', params });
}

export function getDictDataByTypeID(params?: any) {
  return request.get({ url: '/reference/dict-data-by-id', params });
}

export function getDictTypes(params?: any) {
  return request.get({ url: '/reference/dict-types', params });
}

export function getStatusFlowConfigs(params?: any) {
  return request.get({ url: '/reference/status-flow-configs', params });
}

export function getPrintNoticeConfig(params?: any) {
  return request.get({ url: '/reference/print-notice-config', params });
}

export function getCompanyInfo() {
  return request.get({ url: '/reference/company' } as any);
}

export function getReceivingAccountList(params?: any) {
  return request.get({ url: '/reference/receiving-accounts', params });
}

export function getMoldList(params?: any) {
  return request.get({ url: '/reference/molds', params });
}

export function getMoldById(id: number) {
  return request.get({ url: `/reference/molds/${id}` });
}

export function getProductList(params?: any) {
  return request.get({ url: '/reference/products', params });
}

export function getProductById(id: number) {
  return request.get({ url: `/reference/products/${id}` });
}

export function getMoldProductRelationsByMold(moldId: number) {
  return request.get({ url: '/reference/mold-product-relations/by-mold', params: { mold_id: moldId } });
}

export function getMoldProductRelationsByProduct(productId: number) {
  return request.get({ url: '/reference/mold-product-relations/by-product', params: { product_id: productId } });
}

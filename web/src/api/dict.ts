import { request } from '@/utils/request';

// ==================== 字典类型管理 ====================

export function getDictTypes(params?: any) {
  return request.get({ url: '/system/dict-types', params });
}

// 根据 ID 获取字典类型
export function getDictTypeById(id: number) {
  return request.get({ url: `/system/dict-types/${id}` });
}

export function createDictType(data: any) {
  return request.post({ url: '/system/dict-types', data });
}

export function updateDictType(data: any) {
  return request.put({ url: '/system/dict-types', data });
}

export function deleteDictType(id: number) {
  return request.delete({ url: `/system/dict-types/${id}` });
}

// ==================== 字典数据管理 ====================

// 根据字典类型编码获取字典数据列表
export function getDictDataByType(params: { dict_type: string }) {
  return request.get({ url: '/system/dict-data', params });
}

// 根据字典类型 ID 获取字典数据列表
export function getDictDataByTypeId(params: { dict_type_id: number }) {
  return request.get({ url: '/system/dict-data-by-id', params });
}

// 分页获取字典数据列表
export function getDictDataPage(params?: any) {
  return request.get({ url: '/system/dict-data-page', params });
}

export function createDictData(data: any) {
  return request.post({ url: '/system/business-dict-data', data });
}

export function updateDictData(data: any) {
  return request.put({ url: '/system/dict-data', data });
}

export function updateBusinessDictData(data: any) {
  return request.put({ url: '/system/business-dict-data', data });
}

export function deleteDictData(id: number) {
  return request.delete({ url: `/system/business-dict-data/${id}` });
}

// 强制删除字典数据（需二次验证）
export function forceDeleteDictData(id: number, data: { status_name: string; username: string; password: string }) {
  return request.post({ url: `/system/dict-data/${id}/force-delete`, data });
}

// 清除字典缓存
export function clearDictCache() {
  return request.post({ url: '/system/dict-data/clear-cache' });
}

// 恢复模具状态字典 & 状态流转配置为默认值
export function restoreDictSeed() {
  return request.post({ url: '/system/dict-data/restore-seed' });
}

// 恢复生产订单状态字典 & 流转规则为默认值
export function restoreProductionDictSeed() {
  return request.post({ url: '/system/dict-data/restore-production-seed' });
}

// ==================== 字典版本查询（版本轮询） ====================

// 获取字典版本号（用于前端版本轮询检测变更）
export function getDictVersion() {
  return request.get({ url: '/system/dict/version' });
}

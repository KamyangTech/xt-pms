import { request } from '@/utils/request';

export function getStatusFlowConfigs() {
  return request.get({ url: '/system/status-flow-configs' });
}

// 根据ID获取状态流转配置
export function getStatusFlowConfigById(id: number) {
  return request.get({ url: `/system/status-flow-configs/${id}` });
}

export function createStatusFlowConfig(data: any) {
  return request.post({ url: '/system/status-flow-configs', data });
}

export function updateStatusFlowConfig(data: any) {
  return request.put({ url: '/system/status-flow-configs', data });
}

export function deleteStatusFlowConfig(id: number) {
  return request.delete({ url: `/system/status-flow-configs/${id}` });
}

// 刷新缓存
export function refreshStatusFlowCache() {
  return request.post({ url: '/system/status-flow-configs/refresh-cache' });
}

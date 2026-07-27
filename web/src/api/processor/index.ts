import { request } from '@/utils/request';

export function getProcessorList(params?: any) {
  return request.get({ url: '/basic/processor', params });
}

export function createProcessor(data: any) {
  return request.post({ url: '/basic/processor', data });
}

export function updateProcessor(data: any) {
  return request.put({ url: '/basic/processor', data });
}

export function deleteProcessor(id: number) {
  return request.delete({ url: `/basic/processor/${id}` });
}

// 批量删除加工商
export function batchDeleteProcessors(ids: number[]) {
  return request.post({ url: '/basic/processor/batch-delete', data: { ids } });
}

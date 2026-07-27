import { request } from '@/utils/request';

export function getProcesses(params: any) {
  return request.get({ url: '/basic/process', params });
}

export function getProcessById(id: number) {
  return request.get({ url: `/basic/process/${id}` });
}

export function createProcess(data: any) {
  return request.post({ url: '/basic/process', data });
}

export function updateProcess(data: any) {
  return request.put({ url: '/basic/process', data });
}

export function deleteProcess(id: number) {
  return request.delete({ url: `/basic/process/${id}` });
}

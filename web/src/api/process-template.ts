import { request } from '@/utils/request';

export const getProcessTemplates = () => {
  return request.get({ url: '/process-template' });
};

export const getActiveProcessTemplates = () => {
  return request.get({ url: '/process-template/active' });
};

export const createProcessTemplate = (data: any) => {
  return request.post({ url: '/process-template', data });
};

export const updateProcessTemplate = (data: any) => {
  return request.put({ url: '/process-template', data });
};

export const deleteProcessTemplate = (id: number) => {
  return request.delete({ url: `/process-template/${id}` });
};

export const getWorkOrderProcesses = (workOrderId: number) => {
  return request.get({ url: `/mold/work-order/${workOrderId}/processes` });
};

// 切换工序完成状态
export const toggleProcess = (data: {
  work_order_id: number;
  process_template_id: number;
  is_completed: number;
}) => {
  return request.post({ url: '/mold/work-order/process/toggle', data });
};

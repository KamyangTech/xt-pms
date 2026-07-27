import { request } from '@/utils/request';

export interface NotificationConfigItem {
  id: number;
  type_pattern: string;
  target_type: string;   // role | user
  target_id: number;
  enabled: number;       // 1=启用 0=禁用
  create_time: string;
  update_time: string;
}

export interface NotificationConfigListResult {
  list: NotificationConfigItem[];
  total: number;
  page: number;
  page_size: number;
}

/** 获取通知配置列表 */
export function getNotificationConfigs(params: { page?: number; page_size?: number }) {
  return request.get<NotificationConfigListResult>({ url: '/system/notification-configs', params });
}

/** 创建通知配置 */
export function createNotificationConfig(data: Partial<NotificationConfigItem>) {
  return request.post({ url: '/system/notification-configs', data });
}

/** 更新通知配置 */
export function updateNotificationConfig(data: Partial<NotificationConfigItem>) {
  return request.put({ url: '/system/notification-configs', data });
}

/** 删除通知配置 */
export function deleteNotificationConfig(id: number) {
  return request.delete({ url: `/system/notification-configs/${id}` });
}

/** 获取所有通知类型列表（全局，用于配置页面） */
export function getAllNotificationTypes() {
  return request.get<string[]>({ url: '/system/notification-configs/types' });
}

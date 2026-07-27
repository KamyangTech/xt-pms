import { request } from '@/utils/request';

export interface NotificationItem {
  id: number;
  title: string;
  content: string;
  type: string;
  type_label: string;
  quality: string;
  ref_id: number;
  ref_url: string;
  is_read: number;
  create_time: string;
}

export interface NotificationPageResult {
  list: NotificationItem[];
  total: number;
  page: number;
  page_size: number;
}

/** 获取通知分页列表（支持类型筛选） */
export function getNotifications(params: { page?: number; page_size?: number; is_read?: number; type?: string }) {
  return request.get<NotificationPageResult>({ url: '/system/notifications', params });
}

/** 获取通知类型列表 */
export function getNotificationTypes() {
  return request.get<string[]>({ url: '/system/notifications/types' });
}

/** 获取未读通知数量 */
export function getUnreadCount() {
  return request.get<{ count: number }>({ url: '/system/notifications/unread-count' });
}

/** 标记单条已读 */
export function markAsRead(id: number) {
  return request.put({ url: `/system/notifications/${id}/read` });
}

/** 标记全部已读 */
export function markAllAsRead() {
  return request.put({ url: '/system/notifications/read-all' });
}

/** 删除通知 */
export function deleteNotification(id: number) {
  return request.delete({ url: `/system/notifications/${id}` });
}

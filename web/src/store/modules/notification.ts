import { defineStore } from 'pinia';

import { getNotifications, getNotificationTypes, getUnreadCount, markAsRead, markAllAsRead, deleteNotification, type NotificationItem } from '@/api/notification';
import { NOTIFICATION_TYPE_LABELS } from '@/constants';
import { getWebSocketClient } from '@/utils/websocket';

export const useNotificationStore = defineStore('notification', {
  state: () => ({
    list: [] as NotificationItem[],
    total: 0,
    page: 1,
    pageSize: 20,
    unreadCount: 0,
    loading: false,
    hasMore: true,
    /** 当前筛选条件 */
    filterType: '',       // 通知类型（空=全部）
    filterIsRead: undefined as number | undefined, // undefined=全部, 0=未读, 1=已读
    /** 可用的通知类型列表 */
    typeList: [] as { value: string; label: string }[],
  }),
  getters: {
    /** 兼容旧引用 */
    unreadMsg: (state) => state.list.filter((item) => item.is_read === 0),
    readMsg: (state) => state.list.filter((item) => item.is_read === 1),
    msgData: (state) => state.list,
  },
  actions: {
    /** 从 API 拉取通知列表（重置页码） */
    async fetchList(isRead?: number, type?: string, page?: number, pageSize?: number) {
      this.loading = true;
      this.page = page ?? 1;
      if (pageSize !== undefined) this.pageSize = pageSize;
      this.filterIsRead = isRead;
      this.filterType = type ?? '';
      this.hasMore = true;
      try {
        const params: { page: number; page_size: number; is_read?: number; type?: string } = {
          page: this.page,
          page_size: this.pageSize,
        };
        if (this.filterIsRead !== undefined) params.is_read = this.filterIsRead;
        if (this.filterType) params.type = this.filterType;

        const res = await getNotifications(params);
        this.list = res.list || [];
        this.total = res.total || 0;
        this.hasMore = (res.list?.length || 0) >= this.pageSize;
      } catch (error) {
        console.error('[NotificationStore] 获取通知列表失败:', error);
      } finally {
        this.loading = false;
      }
    },

    /** 加载更多（下一页追加） */
    async loadMore() {
      if (this.loading || !this.hasMore) return;
      this.loading = true;
      const nextPage = this.page + 1;
      try {
        const params: { page: number; page_size: number; is_read?: number; type?: string } = {
          page: nextPage,
          page_size: this.pageSize,
        };
        if (this.filterIsRead !== undefined) params.is_read = this.filterIsRead;
        if (this.filterType) params.type = this.filterType;

        const res = await getNotifications(params);
        if (res.list?.length) {
          this.list.push(...res.list);
          this.page = nextPage;
          this.total = res.total || 0;
          this.hasMore = (res.list?.length || 0) >= this.pageSize;
        } else {
          this.hasMore = false;
        }
      } catch (error) {
        console.error('[NotificationStore] 加载更多通知失败:', error);
      } finally {
        this.loading = false;
      }
    },

    /** 获取通知类型列表（按中文标签去重，选中时用前缀匹配） */
    async fetchTypes() {
      try {
        const types = await getNotificationTypes();
        // 按类型前缀分组：取冒号前的部分作为组 key
        // 例如 mold:work_order_created 和 mold:work_order_status → 前缀 mold → 标签"模具工单"
        const groupMap = new Map<string, { prefix: string; label: string }>();
        for (const t of types) {
          const label = NOTIFICATION_TYPE_LABELS[t];
          if (!label) continue; // 跳过没有映射的未知类型
          const prefix = t.split(':')[0] + ':%'; // "mold:%"
          if (!groupMap.has(prefix)) {
            groupMap.set(prefix, { prefix, label });
          }
        }
        this.typeList = Array.from(groupMap.entries()).map(([, { prefix, label }]) => ({
          value: prefix,
          label,
        }));
      } catch (error) {
        console.error('[NotificationStore] 获取通知类型失败:', error);
      }
    },

    /** 获取未读数量 */
    async fetchUnreadCount() {
      try {
        const res = await getUnreadCount();
        this.unreadCount = res.count;
      } catch (error) {
        console.error('[NotificationStore] 获取未读数量失败:', error);
      }
    },

    /** 标记单条已读 */
    async markAsRead(id: number) {
      try {
        await markAsRead(id);
        const item = this.list.find((n) => n.id === id);
        if (item) {
          item.is_read = 1;
          this.unreadCount = Math.max(0, this.unreadCount - 1);
        }
      } catch (error) {
        console.error('[NotificationStore] 标记已读失败:', error);
      }
    },

    /** 全部标记已读 */
    async markAllAsRead() {
      try {
        await markAllAsRead();
        this.list.forEach((item) => { item.is_read = 1; });
        this.unreadCount = 0;
      } catch (error) {
        console.error('[NotificationStore] 全部标记已读失败:', error);
      }
    },

    /** 删除通知 */
    async deleteItem(id: number) {
      try {
        await deleteNotification(id);
        const idx = this.list.findIndex((n) => n.id === id);
        if (idx !== -1) {
          if (this.list[idx].is_read === 0) {
            this.unreadCount = Math.max(0, this.unreadCount - 1);
          }
          this.list.splice(idx, 1);
          this.total = Math.max(0, this.total - 1);
        }
      } catch (error) {
        console.error('[NotificationStore] 删除通知失败:', error);
      }
    },

    /** 新增一条通知到列表头部（WebSocket 实时推送触发） */
    addNotification(item: NotificationItem) {
      if (this.list.some((n) => n.id === item.id)) return;
      this.list.unshift(item);
      this.total += 1;
      this.unreadCount += 1;
    },

    /** 初始化 WebSocket 通知监听 */
    initWebSocketListener() {
      const ws = getWebSocketClient();
      if (!ws) return;

      ws.on('notification', (data: any) => {
        this.addNotification(data);
      });
    },
  },
  persist: false,
});

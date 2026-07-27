<template>
  <div class="notification-page">
    <t-card :bordered="false">
      <!-- 顶部标签切换 -->
      <div class="tabs-header">
        <t-tabs v-model="activeTab" @change="onTabChange">
          <t-tab-panel value="all" label="全部通知" />
          <t-tab-panel value="unread" label="未读通知" />
          <t-tab-panel value="read" label="已读通知" />
        </t-tabs>
        <t-button
          v-if="store.unreadCount > 0"
          variant="text"
          theme="primary"
          @click="onMarkAllRead"
        >
          全部已读
        </t-button>
      </div>

      <!-- 通知列表 -->
      <div class="notify-list">
        <div
          v-for="item in store.list"
          :key="item.id"
          class="notify-item"
          :class="{ 'is-read': item.is_read === 1 }"
          @click="onItemClick(item)"
        >
          <t-tag :theme="getTagTheme(item.type)" variant="light" size="small">
            {{ item.type_label || item.type }}
          </t-tag>
          <span class="notify-title">{{ stripTypePrefix(item.title, item.type_label) }}</span>
          <span class="notify-time">{{ formatTime(item.create_time) }}</span>
        </div>

        <!-- 空状态 -->
        <div v-if="!store.loading && store.list.length === 0" class="empty-state">
          <t-icon name="inbox" size="64" />
          <p>暂无通知</p>
        </div>
      </div>
    </t-card>

    <!-- 通知详情弹窗 -->
    <t-dialog
      v-model:visible="detailVisible"
      header="通知详情"
      :footer="false"
      width="560px"
      placement="center"
     :close-on-overlay-click="false">
      <div v-if="selectedItem" class="detail-content">
        <div class="detail-header">
          <t-tag :theme="getTagTheme(selectedItem.type)" variant="light" size="small">
            {{ selectedItem.type_label || selectedItem.type }}
          </t-tag>
          <span class="detail-time">{{ formatTime(selectedItem.create_time) }}</span>
        </div>
        <div class="detail-title">{{ stripTypePrefix(selectedItem.title, selectedItem.type_label) }}</div>
        <div class="detail-body">{{ selectedItem.content || '暂无详细内容' }}</div>
        <div class="detail-footer">
          <t-button
            v-if="selectedItem.ref_url"
            variant="outline"
            theme="primary"
            size="small"
            @click="onNavigate"
          >
            查看详情
          </t-button>
        </div>
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import type { NotificationItem } from '@/api/notification';
import { onMounted, ref } from 'vue';
import dayjs from 'dayjs';

import { useNotificationStore } from '@/store';

const store = useNotificationStore();
const activeTab = ref('all');
const detailVisible = ref(false);
const selectedItem = ref<NotificationItem | null>(null);

onMounted(() => {
  store.fetchList();
  store.fetchUnreadCount();
});

// 全部已读
const onMarkAllRead = async () => {
  await store.markAllAsRead();
};

// 切换标签
const onTabChange = () => {
  if (activeTab.value === 'all') store.fetchList(undefined);
  else if (activeTab.value === 'unread') store.fetchList(0);
  else store.fetchList(1);
};

// 点击通知：打开详情弹窗，未读则标记已读
const onItemClick = async (item: NotificationItem) => {
  selectedItem.value = item;
  detailVisible.value = true;
  if (item.is_read === 0) {
    await store.markAsRead(item.id);
  }
};

// 跳转关联页面
const onNavigate = () => {
  if (selectedItem.value?.ref_url) {
    window.open(selectedItem.value.ref_url, '_blank');
  }
};

// 根据类型返回标签主题色
const getTagTheme = (type: string) => {
  if (type?.startsWith('finance')) return 'danger';
  if (type?.startsWith('mold')) return 'primary';
  if (type?.startsWith('production')) return 'success';
  if (type?.startsWith('outsourcing')) return 'warning';
  return 'default';
};

// 去掉标题中重复的类型前缀，如 [模具订单]
const stripTypePrefix = (title: string, typeLabel?: string) => {
  if (!typeLabel || !title) return title;
  return title.replace(new RegExp(`^\\[${typeLabel}\\][\\s]*`), '');
};

// 格式化时间
const formatTime = (time: string) => {
  return dayjs(time).format('YYYY-MM-DD HH:mm');
};
</script>

<style lang="less" scoped>
.notification-page {
  padding: 16px;
}

.tabs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;

  :deep(.t-tabs) {
    flex: 1;
  }
}

.notify-list {
  margin-top: 8px;
}

.notify-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 18px 0;
  border-bottom: 1px solid var(--td-component-border);
  cursor: pointer;
  transition: background-color 0.15s;

  &:hover {
    background-color: var(--td-bg-color-container-hover);
  }

  &:last-child {
    border-bottom: none;
  }

  &.is-read {
    .notify-title {
      color: var(--td-text-color-secondary);
    }
  }
}

.notify-title {
  flex: 1;
  color: var(--td-text-color-primary);
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notify-time {
  color: var(--td-text-color-placeholder);
  font-size: 13px;
  white-space: nowrap;
  flex-shrink: 0;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.detail-time {
  color: var(--td-text-color-placeholder);
  font-size: 13px;
}

.detail-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  line-height: 1.6;
}

.detail-body {
  padding: 16px;
  background: var(--td-bg-color-secondarycontainer);
  border-radius: var(--td-radius-medium);
  color: var(--td-text-color-secondary);
  font-size: 14px;
  line-height: 1.8;
  white-space: pre-wrap;
}

.detail-footer {
  display: flex;
  justify-content: flex-end;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120px 0;
  color: var(--td-text-color-placeholder);

  p {
    margin-top: 16px;
    font-size: 14px;
  }
}
</style>

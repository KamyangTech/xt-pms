<template>
  <div class="usage-history-detail-container">
    <!-- 顶部返回 -->
    <div class="top-bar">
      <t-button theme="default" @click="goBackToList" variant="text">
        <template #icon><chevron-left-icon /></template>
        返回
      </t-button>
    </div>

    <!-- 模具信息卡片 -->
    <t-card class="mold-info-card" :bordered="false">
      <div class="mold-info-content">
        <div class="mold-basic">
          <div class="mold-avatar">
            <t-image
              v-if="moldInfo.primaryImage"
              :src="moldInfo.primaryImage"
              fit="cover"
              style="width: 80px; height: 80px; border-radius: 8px"
            />
            <div v-else class="mold-avatar-placeholder">
              <layers-icon size="40px" />
            </div>
          </div>
          <div class="mold-details">
            <h3>{{ moldInfo.moldCode }}</h3>
            <p>{{ moldInfo.moldName || '-' }}</p>
            <p class="mold-meta">客户: {{ moldInfo.customerName || '-' }}</p>
          </div>
        </div>
        <div class="mold-stats">
          <div class="stat-item">
            <div class="stat-value">{{ usageStats.total_ok_qty || 0 }}</div>
            <div class="stat-label">合格品数</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ usageStats.total_bad_qty || 0 }}</div>
            <div class="stat-label">不良品数</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ usageStats.total_orders || 0 }}</div>
            <div class="stat-label">使用次数</div>
          </div>
        </div>
      </div>
    </t-card>

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">使用历史</span>
        <div class="toolbar-actions">
          <t-button theme="default" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button theme="default" @click="handleExport">
            <template #icon><download-icon /></template>
            导出
          </t-button>
          <t-button theme="default" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <t-table
        :data="tableData"
        :columns="displayColumns"
        :loading="loading"
        row-key="id"
        :pagination="pagination"
        :stripe="false"
        @page-change="rehandlePageChange"
      >
        <template #orderCode="{ row }">
          <t-link @click="handleViewOrder(row)" theme="primary">
            {{ row.order_code }}
          </t-link>
        </template>
        <template #productName="{ row }">
          {{ row.product_name || '-' }}
        </template>
        <template #okQty="{ row }">
          <span style="color: #008858">{{ row.ok_qty || 0 }}</span>
        </template>
        <template #badQty="{ row }">
          <span :style="{ color: row.bad_qty > 0 ? '#ad352f' : '#999' }">
            {{ row.bad_qty || 0 }}
          </span>
        </template>
        <template #yieldRate="{ row }">
          <t-tag v-if="row.yield_rate !== undefined" :theme="getYieldRateTheme(row.yield_rate)">
            {{ row.yield_rate }}%
          </t-tag>
          <span v-else>-</span>
        </template>
        <template #report_time="{ row }">
          {{ formatReportTime(row) }}
        </template>
      </t-table>
    </t-card>

    <!-- 列设置对话框 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="mold_usage_history_detail_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TableProps } from 'tdesign-vue-next';
import { RefreshIcon, DownloadIcon, LayersIcon, SettingIcon, ChevronLeftIcon } from 'tdesign-icons-vue-next';
import { request } from '@/utils/request';
import { useRoute, useRouter } from 'vue-router';
import { formatDateTime } from '@/utils/date';
import { OPERATION_COLUMN_WIDTH } from '@/utils/operation-column-width';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';

const route = useRoute();
const router = useRouter();

defineOptions({
  name: 'MoldUsageHistoryDetail',
});

// 模具信息
const moldInfo = ref<any>({});
const usageStats = ref<any>({});

// 表格相关
const loading = ref(false);
const tableData = ref<any[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

// 表格列配置
const tableColumns: TableProps['columns'] = [
  { colKey: 'order_code', title: '订单编号', width: 180 },
  { colKey: 'product_name', title: '产品名称', width: 160 },
  { colKey: 'ok_qty', title: '合格品数', width: 100 },
  { colKey: 'bad_qty', title: '不良品数', width: 100 },
  { colKey: 'yield_rate', title: '合格率', width: 100 },
  { colKey: 'report_time', title: '使用时间', width: 120 },
];

const allColumns: ColumnOption[] = [
  { colKey: 'order_code', title: '订单编号', pinned: true },
  { colKey: 'product_name', title: '产品名称' },
  { colKey: 'ok_qty', title: '合格品数' },
  { colKey: 'bad_qty', title: '不良品数' },
  { colKey: 'yield_rate', title: '合格率' },
  { colKey: 'report_time', title: '使用时间' },
];

const DEFAULT_COLUMNS = ['order_code', 'product_name', 'ok_qty', 'bad_qty', 'yield_rate', 'report_time'];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

const displayColumns = computed(() => {
  return selectedColumns.value
    .map(colKey => tableColumns.find(col => col.colKey === colKey))
    .filter(col => col !== undefined) as TableProps['columns'];
});

const getMoldInfo = async (moldId: number) => {
  try {
    const res = await request.get({
      url: `/mold/${moldId}`,
    });

    if (res) {
      moldInfo.value = {
        id: res.id,
        moldCode: res.mold_code,
        moldName: res.mold_name,
        customerName: res.customer?.cus_name,
        primaryImage: res.primary_image_url
          ? (res.primary_image_url.startsWith('http')
              ? res.primary_image_url
              : `${import.meta.env.VITE_API_URL || ''}${res.primary_image_url}`)
          : '',
      };
    }
  } catch (error) {
    console.error('获取模具信息失败:', error);
  }
};

const getUsageStats = async (moldId: number) => {
  try {
    const res = await request.get({
      url: `/production/mold/${moldId}/usage-stats`,
    });

    if (res) {
      usageStats.value = res;
    }
  } catch (error) {
    console.error('获取使用统计失败:', error);
  }
};

// 获取使用历史列表
const getUsageHistory = async () => {
  const moldId = Number(route.params.moldId);
  if (!moldId) return;

  loading.value = true;
  try {
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize,
    };

    const res = await request.get({
      url: `/production/mold/${moldId}/usage-history`,
      params,
    });

    const list = (res.list || []).map((item: any) => {
      const totalQty = (item.ok_qty || 0) + (item.bad_qty || 0);
      const yieldRate = totalQty > 0
        ? ((item.ok_qty / totalQty) * 100).toFixed(2)
        : null;

      return {
        ...item,
        yield_rate: yieldRate,
      };
    });

    tableData.value = list;
    pagination.total = res.total || 0;
  } catch (error) {
    MessagePlugin.error('获取使用历史失败');
  } finally {
    loading.value = false;
  }
};

// 刷新
const handleRefresh = () => {
  const moldId = Number(route.params.moldId);
  if (moldId) {
    getMoldInfo(moldId);
    getUsageStats(moldId);
  }
  getUsageHistory();
  MessagePlugin.success('刷新成功');
};

// 导出
const handleExport = () => {
  const exportData: Record<string, any>[] = tableData.value.map(item => ({
    '订单编号': item.order_code,
    '产品名称': item.product_name || '-',
    '合格品数': item.ok_qty || 0,
    '不良品数': item.bad_qty || 0,
    '合格率': item.yield_rate ? `${item.yield_rate}%` : '-',
    '使用时间': formatDateTime(item.report_time),
  }));

  const headers = Object.keys(exportData[0] || {});
  const csvContent = [
    headers.join(','),
    ...exportData.map(row =>
      headers.map(header => `"${row[header] || ''}"`).join(',')
    )
  ].join('\n');

  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  link.setAttribute('href', url);
  link.setAttribute('download', `模具使用历史_${new Date().getTime()}.csv`);
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  MessagePlugin.success('导出成功');
};

// 查看订单详情
const handleViewOrder = (row: any) => {
  if (row.order_id) {
    router.push({ name: 'ProductionOrderDetail', params: { id: String(row.order_id) } });
  }
};

const goBackToList = () => {
  router.push({ name: 'MoldUsageHistory' });
};

// 分页变化
const rehandlePageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  getUsageHistory();
};

// 打开列设置
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 列配置变化
const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
};

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('mold_usage_history_detail_table_columns');
    if (saved) {
      const savedData = JSON.parse(saved);
      if (Array.isArray(savedData)) {
        const validColumns = savedData.filter((colKey: string) =>
          allColumns.some(col => col.colKey === colKey)
        );
        selectedColumns.value = validColumns.length > 0 ? validColumns : [...DEFAULT_COLUMNS];
        return;
      }
    }
  } catch (error) {
    console.error('加载列配置失败:', error);
  }
  selectedColumns.value = [...DEFAULT_COLUMNS];
};

// 格式化使用时间
const formatReportTime = (row: any) => {
  const timeValue = row.report_time || row.usage_date;
  if (!timeValue) return '-';
  return formatDateTime(timeValue);
};

const getYieldRateTheme = (rate: number) => {
  if (rate >= 95) return 'success';
  if (rate >= 90) return 'warning';
  return 'danger';
};

onMounted(() => {
  const moldId = Number(route.params.moldId);
  if (moldId) {
    getMoldInfo(moldId);
    getUsageStats(moldId);
    getUsageHistory();
  }
  loadColumnConfig();
});

watch(
  () => route.params.moldId,
  (newMoldId) => {
    if (route.name !== 'MoldUsageHistoryDetail') return;
    const moldId = Number(newMoldId);
    if (moldId) {
      getMoldInfo(moldId);
      getUsageStats(moldId);
      getUsageHistory();
    }
    loadColumnConfig();
  },
);
</script>

<style scoped>
.usage-history-detail-container {
  padding: 16px;
}

.mold-info-card {
  margin-bottom: 16px;
}

.mold-info-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.mold-basic {
  display: flex;
  gap: 16px;
  align-items: center;
}

.mold-avatar-placeholder {
  width: 80px;
  height: 80px;
  border-radius: 8px;
  background: linear-gradient(135deg, #f5f5f5 0%, #e7e7e7 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
}

.mold-details h3 {
  margin: 0 0 4px 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.mold-details p {
  margin: 4px 0;
  color: var(--td-text-color-secondary);
  font-size: 14px;
}

.mold-meta {
  font-size: 12px !important;
  color: var(--td-text-color-placeholder) !important;
}

.mold-stats {
  display: flex;
  gap: 32px;
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--td-text-color-primary);
  line-height: 1.2;
}

.stat-label {
  font-size: 12px;
  color: var(--td-text-color-secondary);
  margin-top: 4px;
}

.top-bar {
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-card {
  margin-bottom: 16px;
}

.toolbar-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.toolbar-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}
</style>

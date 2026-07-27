<template>
  <div class="usage-history-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">使用历史记录</span>
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
        <template #order_code="{ row }">
          <t-link @click="handleViewOrder(row)" theme="primary">
            {{ row.order_code }}
          </t-link>
        </template>
        <template #mold_code="{ row }">
          {{ row.mold_code || '-' }}
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
        <template #operation="{ row }">
          <div :style="{ width: `${OPERATION_COLUMN_WIDTH.SMALL}px` }">
            <t-link theme="primary" @click="goToMoldUsage(row)">详情</t-link>
          </div>
        </template>
      </t-table>
    </t-card>
    <!-- 列设置对话框 - 使用通用组件 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="mold_usage_history_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TableProps } from 'tdesign-vue-next';
import { RefreshIcon, DownloadIcon, SettingIcon } from 'tdesign-icons-vue-next';
import { request } from '@/utils/request';
import { useRouter } from 'vue-router';
import { formatDate, formatDateTime } from '@/utils/date';
import { OPERATION_COLUMN_WIDTH } from '@/utils/operation-column-width';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';

const router = useRouter();

defineOptions({
  name: 'MoldUsageHistory',
});

// 表格相关
const loading = ref(false);
const tableData = ref<any[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

// 搜索表单
const formData = reactive({
  moldCode: '',
  dateRange: [],
});

// 表格列配置
const tableColumns: TableProps['columns'] = [
  { colKey: 'mold_code', title: '模具编号', width: 160, fixed: 'left' },
  { colKey: 'order_code', title: '订单编号', width: 180 },
  { colKey: 'product_name', title: '产品名称', width: 160 },
  { colKey: 'ok_qty', title: '合格品数', width: 100 },
  { colKey: 'bad_qty', title: '不良品数', width: 100 },
  { colKey: 'yield_rate', title: '合格率', width: 100 },
  { colKey: 'report_time', title: '使用时间', width: 120 },
  { colKey: 'operation', title: '操作', width: OPERATION_COLUMN_WIDTH.SMALL, fixed: 'right' },
];

// 列设置配置（不含操作列）
const allColumns: ColumnOption[] = [
  { colKey: 'mold_code', title: '模具编号', pinned: true },
  { colKey: 'order_code', title: '订单编号' },
  { colKey: 'product_name', title: '产品名称' },
  { colKey: 'ok_qty', title: '合格品数' },
  { colKey: 'bad_qty', title: '不良品数' },
  { colKey: 'yield_rate', title: '合格率' },
  { colKey: 'report_time', title: '使用时间' },
];

const DEFAULT_COLUMNS = ['mold_code', 'order_code', 'product_name', 'ok_qty', 'bad_qty', 'yield_rate', 'report_time'];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

// 计算显示列（根据用户选择，强制操作列在最后）
const displayColumns = computed(() => {
  const orderedColumns = selectedColumns.value
    .map(colKey => tableColumns.find(col => col.colKey === colKey))
    .filter(col => col !== undefined) as TableProps['columns'];
  const operationCol = tableColumns.find(col => col.colKey === 'operation');
  if (operationCol) orderedColumns.push(operationCol);
  return orderedColumns;
});

// 获取使用历史列表
const getUsageHistory = async (moldCodeParam?: string) => {
  // 如果没有传入参数，使用筛选器中的模具编号
  if (moldCodeParam === undefined && formData.moldCode) {
    moldCodeParam = formData.moldCode;
  }

  loading.value = true;
  try {
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize,
    };

    if (formData.dateRange && formData.dateRange.length === 2) {
      params.start_date = formData.dateRange[0];
      params.end_date = formData.dateRange[1];
    }

    if (moldCodeParam) {
      params.mold_code = moldCodeParam;
    }

    const res = await request.get({
      url: '/production/mold/usage-history',
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

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  getUsageHistory(formData.moldCode);
};

const handleReset = () => {
  formData.dateRange = [];
  formData.moldCode = '';
  getUsageHistory();
};

// 刷新
const handleRefresh = () => {
  getUsageHistory(formData.moldCode);
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
  if (row.order_code) {
    router.push({ name: 'ProductionOrderList', query: { order_code: row.order_code } });
  }
};

// 跳转到指定模具的使用历史
const goToMoldUsage = (row: any) => {
  if (row.mold_id && Number(row.mold_id) > 0) {
    router.push({ name: 'MoldUsageHistoryDetail', params: { moldId: String(row.mold_id) } });
  }
};

// 分页变化
const rehandlePageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  getUsageHistory(formData.moldCode);
};

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('mold_usage_history_table_columns');
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

// 打开列设置对话框
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 列配置变化时的回调
const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
};

// 格式化使用时间（只显示日期）
const formatReportTime = (row: any) => {
  const timeValue = row.report_time || row.usage_date;
  if (!timeValue) return '-';
  return formatDate(timeValue);
};

const getYieldRateTheme = (rate: number) => {
  if (rate >= 95) return 'success';
  if (rate >= 90) return 'warning';
  return 'danger';
};

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'moldCode', placeholder: '模具编号' },
  { type: 'dateRangePicker', field: 'dateRange' },
];

onMounted(() => {
  loadColumnConfig();
  getUsageHistory();
});
</script>

<style scoped>
.usage-history-container {
  padding: 16px;
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

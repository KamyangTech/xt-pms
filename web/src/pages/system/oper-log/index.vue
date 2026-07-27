<template>
  <div class="oper-log-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <t-card class="toolbar-card">
      <div class="toolbar-content">
        <span class="toolbar-title">操作日志</span>
        <div class="toolbar-actions">
          <t-button theme="default" variant="outline" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button theme="default" variant="outline" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
          <t-button theme="primary" @click="handleExport">
            <template #icon><download-icon /></template>
            导出 Excel
          </t-button>
        </div>
      </div>
    </t-card>

    <t-card :bordered="false">
      <t-table
        :data="tableData"
        :columns="displayColumns"
        :loading="loading"
        row-key="id"
        hover
        :pagination="pagination"
        @page-change="onPageChange"
      >
        <template #oper_type="{ row }">
          <t-tag :theme="operTypeTheme(row.oper_type)" variant="light">
            {{ operTypeLabel(row.oper_type) }}
          </t-tag>
        </template>
        <template #status="{ row }">
          <t-tag :theme="row.status === 0 ? 'success' : 'danger'" variant="light">
            {{ row.status === 0 ? '正常' : '异常' }}
          </t-tag>
        </template>
        <template #cost_time="{ row }">{{ formatCostTime(row.cost_time) }}</template>
        <template #create_time="{ row }">{{ formatDateTime(row.create_time) }}</template>
        <template #operation="{ row }">
          <t-link theme="primary" @click="handleViewDetail(row)">查看详情</t-link>
        </template>
      </t-table>
    </t-card>

    <!-- 列设置对话框 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="system_oper_log_table_columns"
      @change="handleColumnsChange"
    />

    <!-- 详情对话框 -->
    <t-dialog
      v-model:visible="detailVisible"
      title="操作日志详情"
      width="700px"
      placement="center"
      :close-btn="true"
      :confirm-btn="null"
      :cancel-btn="{ content: '关闭' }"
     :close-on-overlay-click="false">
      <t-descriptions v-if="detailData" layout="vertical" bordered>
        <t-descriptions-item label="操作模块">{{ detailData.title }}</t-descriptions-item>
        <t-descriptions-item label="操作类型">{{ operTypeLabel(detailData.oper_type) }}</t-descriptions-item>
        <t-descriptions-item label="操作人">{{ detailData.oper_name }}</t-descriptions-item>
        <t-descriptions-item label="操作IP">{{ detailData.oper_ip }}</t-descriptions-item>
        <t-descriptions-item label="操作时间">{{ formatDateTime(detailData.create_time) }}</t-descriptions-item>
        <t-descriptions-item label="耗时">{{ detailData.cost_time }}ms</t-descriptions-item>
        <t-descriptions-item label="请求方式">{{ detailData.request_method }}</t-descriptions-item>
        <t-descriptions-item label="请求URL" :span="2">{{ detailData.request_url }}</t-descriptions-item>
        <t-descriptions-item label="请求参数" :span="3">
          <pre class="code-block">{{ detailData.request_params || '-' }}</pre>
        </t-descriptions-item>
        <t-descriptions-item label="返回结果" :span="3">
          <pre class="code-block">{{ detailData.json_result || '-' }}</pre>
        </t-descriptions-item>
        <t-descriptions-item v-if="detailData.error_msg" label="错误信息" :span="3">
          <pre class="code-block error">{{ detailData.error_msg }}</pre>
        </t-descriptions-item>
      </t-descriptions>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { DownloadIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';
import type { PrimaryTableCol } from 'tdesign-vue-next';

import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';

import { getOperLogDetail, getOperLogList, getOperLogExportUrl, type OperLogItem, type OperLogListParams } from '@/api/system/oper-log';
import { formatDateTime } from '@/utils/date';

defineOptions({
  name: 'SystemOperLog',
});

const loading = ref(false);
const detailVisible = ref(false);
const detailData = ref<OperLogItem | null>(null);

const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
});

const tableData = ref<OperLogItem[]>([]);

const formData = reactive({
  title: '',
  oper_name: '',
  oper_type: undefined as number | undefined,
  status: undefined as number | undefined,
  start_time: '',
  end_time: '',
});

// 浏览器的 dayjs 可能没有，使用 Date 做基础格式化
const formatDateTimeLocal = (val: string | undefined) => {
  if (!val) return '-';
  try {
    const d = new Date(val);
    const pad = (n: number) => String(n).padStart(2, '0');
    return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
  } catch {
    return val;
  }
};

const formatCostTime = (ms: number) => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(1)}s`;
  }
  return `${ms}ms`;
};

const operTypeLabel = (type: number): string => {
  const map: Record<number, string> = { 1: '新增', 2: '删除', 3: '修改', 4: '查询', 5: '导入', 6: '导出', 7: '登录', 8: '登出' };
  return map[type] || `未知(${type})`;
};

const operTypeTheme = (type: number): 'default' | 'success' | 'primary' | 'warning' | 'danger' => {
  const map: Record<number, 'default' | 'success' | 'primary' | 'warning' | 'danger'> = { 1: 'success', 2: 'danger', 3: 'warning', 4: 'default', 5: 'primary', 6: 'primary', 7: 'default', 8: 'default' };
  return map[type] || 'default';
};

// 列设置相关
const baseColumns: PrimaryTableCol[] = [
  { colKey: 'create_time', title: '操作时间', width: 180, fixed: 'left' as const },
  { colKey: 'title', title: '操作模块', width: 130, ellipsis: true },
  { colKey: 'oper_type', title: '操作类型', width: 100 },
  { colKey: 'oper_name', title: '操作人', width: 100 },
  { colKey: 'oper_ip', title: '操作IP', width: 140 },
  { colKey: 'request_method', title: '请求方式', width: 100 },
  { colKey: 'request_url', title: '请求URL', width: 200, ellipsis: true },
  { colKey: 'status', title: '状态', width: 80 },
  { colKey: 'cost_time', title: '耗时', width: 80 },
];

const columnsWithOperation: PrimaryTableCol[] = [
  ...baseColumns,
  { colKey: 'operation', title: '操作', width: 100, fixed: 'right' as const },
];

const allColumns: ColumnOption[] = [
  { colKey: 'create_time', title: '操作时间', pinned: true },
  { colKey: 'title', title: '操作模块' },
  { colKey: 'oper_type', title: '操作类型' },
  { colKey: 'oper_name', title: '操作人' },
  { colKey: 'oper_ip', title: '操作IP' },
  { colKey: 'request_method', title: '请求方式' },
  { colKey: 'request_url', title: '请求URL' },
  { colKey: 'status', title: '状态' },
  { colKey: 'cost_time', title: '耗时' },
];

const DEFAULT_COLUMNS = ['create_time', 'title', 'oper_type', 'oper_name', 'oper_ip', 'request_method', 'request_url', 'status', 'cost_time'];

const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

const displayColumns = computed<PrimaryTableCol[]>(() => {
  const orderedColumns = selectedColumns.value
    .map(colKey => columnsWithOperation.find(col => col.colKey === colKey))
    .filter(col => col !== undefined);

  // 操作列始终在最后
  const operationCol = columnsWithOperation.find(col => col.colKey === 'operation');
  if (operationCol) {
    orderedColumns.push(operationCol);
  }

  return orderedColumns;
});

const fetchList = async () => {
  loading.value = true;
  try {
    const params: OperLogListParams = {
      page: pagination.current,
      page_size: pagination.pageSize,
    };
    if (formData.title) params.title = formData.title;
    if (formData.oper_name) params.oper_name = formData.oper_name;
    if (formData.oper_type !== undefined) params.oper_type = formData.oper_type;
    if (formData.status !== undefined) params.status = formData.status;
    if (formData.start_time) params.start_time = formData.start_time;
    if (formData.end_time) params.end_time = formData.end_time;

    const res = await getOperLogList(params);
    tableData.value = res.list || [];
    pagination.total = res.total || 0;
  } catch (error) {
    console.error('获取操作日志失败:', error);
    MessagePlugin.error('获取操作日志失败');
  } finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  pagination.current = 1;
  fetchList();
};

const handleReset = () => {
  formData.title = '';
  formData.oper_name = '';
  formData.oper_type = undefined;
  formData.status = undefined;
  formData.start_time = '';
  formData.end_time = '';
  pagination.current = 1;
  fetchList();
};

const onPageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  fetchList();
};

const handleRefresh = () => {
  fetchList();
  MessagePlugin.success('刷新成功');
};

const handleExport = () => {
  const params: OperLogListParams = {};
  if (formData.title) params.title = formData.title;
  if (formData.oper_name) params.oper_name = formData.oper_name;
  if (formData.oper_type !== undefined) params.oper_type = formData.oper_type;
  if (formData.status !== undefined) params.status = formData.status;
  if (formData.start_time) params.start_time = formData.start_time;
  if (formData.end_time) params.end_time = formData.end_time;

  window.open(getOperLogExportUrl(params), '_blank');
};

const handleViewDetail = async (row: OperLogItem) => {
  try {
    const res = await getOperLogDetail(row.id);
    detailData.value = res;
    detailVisible.value = true;
  } catch {
    MessagePlugin.error('获取日志详情失败');
  }
};

// 列设置
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

const handleColumnsChange = (cols: string[]) => {
  selectedColumns.value = cols;
};

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('system_oper_log_table_columns');
    if (saved) {
      const savedData = JSON.parse(saved);
      if (savedData.visible && Array.isArray(savedData.visible)) {
        selectedColumns.value = savedData.visible;
      } else if (Array.isArray(savedData)) {
        selectedColumns.value = savedData;
      } else {
        selectedColumns.value = [...DEFAULT_COLUMNS];
      }
    } else {
      selectedColumns.value = [...DEFAULT_COLUMNS];
    }
  } catch (error) {
    console.error('加载列配置失败:', error);
    selectedColumns.value = [...DEFAULT_COLUMNS];
  }
};

const filterConfig: FilterItem[] = [
  { type: 'input', field: 'title', placeholder: '操作模块' },
  { type: 'input', field: 'oper_name', placeholder: '操作人' },
  {
    type: 'select',
    field: 'oper_type',
    placeholder: '操作类型',
    options: [
      { label: '新增', value: 1 },
      { label: '删除', value: 2 },
      { label: '修改', value: 3 },
      { label: '查询', value: 4 },
      { label: '导入', value: 5 },
      { label: '导出', value: 6 },
      { label: '登录', value: 7 },
      { label: '登出', value: 8 },
    ],
    showMore: true,
  },
  {
    type: 'select',
    field: 'status',
    placeholder: '状态',
    options: [
      { label: '正常', value: 0 },
      { label: '异常', value: 1 },
    ],
    showMore: true,
  },
];

onMounted(() => {
  loadColumnConfig();
  fetchList();
});
</script>

<style lang="less" scoped>
.oper-log-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;

  .toolbar-card {
    margin-bottom: 0;

    .toolbar-content {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .toolbar-title {
        font-size: 16px;
        font-weight: 500;
        color: var(--td-text-color-primary);
      }

      .toolbar-actions {
        display: flex;
        gap: 8px;
      }
    }
  }
}

.code-block {
  background: var(--td-bg-color-component, #f5f5f5);
  border-radius: 4px;
  padding: 8px 12px;
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  max-height: 200px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;

  &.error {
    color: var(--td-error-color, #e34d59);
  }
}
</style>

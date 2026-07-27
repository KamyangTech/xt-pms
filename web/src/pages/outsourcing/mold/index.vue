<template>
  <div class="outsourcing-mold-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">模具外发管理</span>
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
        <template #recordCode="{ row }">
          <t-link @click="handleView(row)" class="record-code-link">
            {{ row.record_code }}
          </t-link>
        </template>
        <template #relatedInfo="{ row }">
          <div class="related-info">
            <t-link
              v-if="row.work_order?.work_code"
              theme="primary"
              @click="handleViewWorkOrderProcess(row.work_order)"
              class="work-order-link"
            >
              工单: {{ row.work_order.work_code }}
            </t-link>
            <div v-else>工单: -</div>
            <div v-if="row.mold?.mold_code" class="mold-tags">
              <span class="mold-chip" @click.stop="handleViewMold(row.mold?.id)">
                {{ row.mold.mold_code }}
              </span>
            </div>
            <div v-else style="color: #999; font-size: 12px">模具: -</div>
          </div>
        </template>
        <template #processor="{ row }">
          <div v-if="row.processor">
            <div>{{ row.processor.processor_name }}</div>
            <div v-if="row.processor.process_type" class="processor-sub">{{ row.processor.process_type }}</div>
          </div>
          <div v-else>-</div>
        </template>
        <template #status="{ row }">
          <t-tag :theme="getStatusTheme(row.status)">
            {{ getStatusText(row.status) }}
          </t-tag>
        </template>
        <template #progress="{ row }">
          <span
            v-if="currentStepName(row)"
            class="process-capsule process-completed"
            @click="showProcessDetail(row)"
          >
            {{ currentStepName(row) }}
          </span>
          <span
            v-else-if="row.work_order"
            class="process-capsule process-pending"
            @click="showProcessDetail(row)"
          >
            未开始
          </span>
          <span v-else>-</span>
        </template>
        <template #outsourceDate="{ row }">
          {{ row.outsource_date ? formatDate(row.outsource_date) : '-' }}
        </template>
        <template #expectedReturnDate="{ row }">
          {{ row.expected_return_date ? formatDate(row.expected_return_date) : '-' }}
        </template>
        <template #actualReturnDate="{ row }">
          {{ row.actual_return_date ? formatDate(row.actual_return_date) : '-' }}
        </template>
        <template #createTime="{ row }">
          {{ row.create_time ? formatDate(row.create_time) : '-' }}
        </template>
        <template #operation="{ row }">
          <t-space :size="8">
            <t-link theme="default" @click="handleView(row)">详情</t-link>
            <t-link 
              v-if="row.status === 'PENDING'"
              theme="primary" 
              @click="handleSend(row)"
            >
              发出
            </t-link>
            <t-link 
              v-if="row.status === 'SENT' || row.status === 'PRODUCING'"
              theme="success" 
              @click="handleReceive(row)"
            >
              收回
            </t-link>
            <t-link 
              v-if="row.status === 'SENT' || row.status === 'PRODUCING'"
              theme="default" 
              @click="handleShareLink(row)"
            >
              <template #icon><link-icon /></template>
              分享链接
            </t-link>
            <t-popconfirm
              v-if="(row.status === 'SENT' || row.status === 'PRODUCING') && row.share_token"
              content="重置后旧链接将立即失效，确定重置？"
              @confirm="handleResetLink(row)"
            >
              <t-link theme="danger">重置</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 查看详情对话框 -->
    <t-dialog
      v-model:visible="detailVisible"
      title="模具外发详情"
      width="900px"
      :footer="null"
     :close-on-overlay-click="false">
      <t-descriptions bordered :column="2">
        <t-descriptions-item label="外发单号">{{ detailData.record_code }}</t-descriptions-item>
        <t-descriptions-item label="外发类型">
          <t-tag shape="round" variant="outline" theme="warning">
            模具外发
          </t-tag>
        </t-descriptions-item>
        <t-descriptions-item label="工单编号">
          {{ detailData.work_order?.work_code || '-' }}
        </t-descriptions-item>
        <t-descriptions-item label="模具编号">
          {{ detailData.mold?.mold_code || '-' }}
        </t-descriptions-item>
        <t-descriptions-item label="加工商">
          <span v-if="detailData.processor">
            {{ detailData.processor.process_type ? `${detailData.processor.processor_name} - ${detailData.processor.process_type}` : detailData.processor.processor_name }}
          </span>
          <span v-else>-</span>
        </t-descriptions-item>
        <t-descriptions-item label="外发数量">{{ detailData.quantity || '-' }}</t-descriptions-item>
        <t-descriptions-item label="状态">
          <t-tag :theme="getStatusTheme(detailData.status)">
            {{ getStatusText(detailData.status) }}
          </t-tag>
        </t-descriptions-item>
        <t-descriptions-item label="外发日期">{{ detailData.outsource_date ? formatDateTime(detailData.outsource_date) : '-' }}</t-descriptions-item>
        <t-descriptions-item label="预计返回日期">{{ detailData.expected_return_date ? formatDateTime(detailData.expected_return_date) : '-' }}</t-descriptions-item>
        <t-descriptions-item label="实际返回日期">{{ detailData.actual_return_date ? formatDateTime(detailData.actual_return_date) : '-' }}</t-descriptions-item>
        <t-descriptions-item label="备注" :span="2">
          <div style="white-space: pre-wrap;">{{ detailData.remark || '-' }}</div>
        </t-descriptions-item>
      </t-descriptions>
    </t-dialog>

    <!-- 工序详情抽屉 - 使用公共组件 -->
    <ProcessDetailDrawer
      v-model:visible="processDrawerVisible"
      :work-order="currentWorkOrder"
      @close="processDrawerVisible = false"
      @refresh="handleProcessRefresh"
    />

    <!-- 列设置对话框 - 使用通用组件 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="outsourcing_mold_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { onActivated, onMounted, reactive, ref, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TableProps } from 'tdesign-vue-next';
import { RefreshIcon, DownloadIcon, SettingIcon, LinkIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';

import { getProcessorList } from '@/api/reference';
import { generateShareToken, regenerateShareToken } from '@/api/production/order';
import { useDictRaw } from '@/hooks/useDict';
import { request } from '@/utils/request';
import { copyText } from '@/utils/clipboard';
import { formatDate, formatDateTime } from '@/utils/date';
import { OPERATION_COLUMN_WIDTH } from '@/utils/operation-column-width';
import { useRouter } from 'vue-router';
import ProcessDetailDrawer from '@/components/mold/ProcessDetailDrawer.vue';
import ColumnSettingDialog from '@/components/ColumnSettingDialog/index.vue';

// 配置 dayjs


defineOptions({
  name: 'OutsourcingMold',
});

const router = useRouter();

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
  recordCode: '',
  status: undefined,
  processorId: undefined,
});

// 加工商列表
const processorList = ref<any[]>([]);

// 外发状态字典数据（useDictRaw 自动加载 + 全局缓存，按 sort 排序）
const { items: statusDictList } = useDictRaw('outsourcing-status', (items) =>
  [...items].sort((a: any, b: any) => (a.sort || 0) - (b.sort || 0)),
);

const getProcessorListData = async () => {
  try {
    const res = await getProcessorList();
    processorList.value = res.list || [];
  } catch (error) {
    console.error('获取加工商列表失败:', error);
  }
};

// 获取外发状态字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 所有可选列配置（用于列设置对话框）
const allColumns = [
  { colKey: 'record_code', title: '外发单号', pinned: true },  // ✨ 固定列：不可隐藏/移动
  { colKey: 'relatedInfo', title: '关联信息' },
  { colKey: 'processor', title: '加工商' },
  { colKey: 'quantity', title: '数量' },
  { colKey: 'progress', title: '进度' },
  { colKey: 'status', title: '状态' },
  { colKey: 'outsourceDate', title: '外发日期' },
  { colKey: 'expectedReturnDate', title: '预计返回' },
  { colKey: 'actualReturnDate', title: '实际返回' },
  { colKey: 'createTime', title: '创建时间' },
  { colKey: 'operation', title: '操作', pinned: true },  // ✨ 固定列：不可隐藏/移动
];

// 表格列配置（动态生成，支持工序列宽自适应）
const tableColumns = computed<TableProps['columns']>(() => [
  { colKey: 'record_code', title: '外发单号', width: 180, fixed: 'left' },
  { colKey: 'relatedInfo', title: '关联信息', width: 250 },
  { colKey: 'processor', title: '加工商', width: 180 },
  { colKey: 'quantity', title: '数量', width: 80 },
  { colKey: 'progress', title: '当前工序', width: 120 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'outsourceDate', title: '外发日期', width: 120 },
  { colKey: 'expectedReturnDate', title: '预计返回', width: 120 },
  { colKey: 'actualReturnDate', title: '实际返回', width: 120 },
  { colKey: 'createTime', title: '创建时间', width: 120 },
  { colKey: 'operation', title: '操作', width: 240, fixed: 'right' },
]);

// 默认显示的列（不包括外发单号，因为它固定显示）
const DEFAULT_COLUMNS = ['record_code', 'relatedInfo', 'processor', 'quantity', 'progress', 'status', 'outsourceDate', 'expectedReturnDate', 'actualReturnDate', 'createTime', 'operation'];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([...DEFAULT_COLUMNS]);

const displayColumns = computed(() => {
  // 根据 selectedColumns 过滤 tableColumns
  return tableColumns.value.filter(col => selectedColumns.value.includes(col.colKey as string));
});

// 对话框相关
const detailVisible = ref(false);
const detailData = ref<any>({});



// 工序详情相关
const processDrawerVisible = ref(false);
const currentWorkOrder = ref<any>(null);

const getOutsourcingRecords = async () => {
  loading.value = true;
  try {
    const params: any = {
      outsource_type: 'MOLD', // 固定为模具外发
      page: pagination.current,
      page_size: pagination.pageSize,
    };
    
    if (formData.recordCode) {
      params.record_code = formData.recordCode;
    }
    if (formData.status !== undefined && formData.status !== null) {
      params.status = formData.status;
    }
    if (formData.processorId) {
      params.processor_id = formData.processorId;
    }

    const res = await request.get({
      url: '/production/outsourcing/list',
      params,
    });

    tableData.value = res?.list || [];
    pagination.total = res?.total || 0;
  } catch (error) {
    MessagePlugin.error('获取模具外发记录失败');
  } finally {
    loading.value = false;
  }
};

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  getOutsourcingRecords();
};

const handleReset = () => {
  formData.recordCode = '';
  formData.status = undefined;
  formData.processorId = undefined;
  getOutsourcingRecords();
};

// 刷新
const handleRefresh = () => {
  getOutsourcingRecords();
  MessagePlugin.success('刷新成功');
};

// 导出
const handleExport = () => {
  const exportData = tableData.value.map(item => ({
    '外发单号': item.record_code,
    '工单编号': item.work_order?.work_code || '-',
    '模具编号': item.mold?.mold_code || '-',
    '加工商': item.processor ? (item.processor.process_type ? `${item.processor.processor_name} - ${item.processor.process_type}` : item.processor.processor_name) : '-',
    '数量': item.quantity || '-',
    '状态': getStatusText(item.status),
    '外发日期': item.outsource_date ? formatDate(item.outsource_date) : '-',
    '实际返回日期': item.actual_return_date ? formatDate(item.actual_return_date) : '-',
  }));

  const headers = Object.keys(exportData[0] || {});
  const csvContent = [
    headers.join(','),
    ...exportData.map(row => 
      headers.map(header => `"${(row as any)[header] || ''}"`).join(',')
    )
  ].join('\n');

  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  link.setAttribute('href', url);
  link.setAttribute('download', `模具外发记录_${new Date().getTime()}.csv`);
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  MessagePlugin.success('导出成功');
};

// 查看详情
const handleView = async (row: any) => {
  detailData.value = row;
  detailVisible.value = true;
};

// 发出
const handleSend = async (row: any) => {
  try {
    await request.post({
      url: `/production/outsourcing/${row.id}/send`,
      data: {},
    });
    MessagePlugin.success('发出成功');
    getOutsourcingRecords();
  } catch (error: any) {
    MessagePlugin.error(error.message || '发出失败');
  }
};

// 收回
const handleReceive = async (row: any) => {
  try {
    await request.post({
      url: `/production/outsourcing/${row.id}/receive`,
      data: {},
    });
    MessagePlugin.success('收回成功');
    getOutsourcingRecords();
  } catch (error: any) {
    MessagePlugin.error(error.message || '收回失败');
  }
};

// 分享链接：已有 token 则直接复制，否则先创建
const handleShareLink = async (row: any) => {
  try {
    let token = row.share_token;
    if (!token) {
      const res = await generateShareToken(row.id);
      token = res.share_token;
      row.share_token = token;
    }
    const shareUrl = `${window.location.origin}/r/${token}`;
    const ok = await copyText(shareUrl);
    if (ok) {
      MessagePlugin.success('分享链接已复制到剪贴板');
    } else {
      MessagePlugin.warning('复制失败，请手动复制：' + shareUrl);
    }
  } catch (error: any) {
    MessagePlugin.error(error.message || '生成分享链接失败');
  }
};

const handleResetLink = async (row: any) => {
  try {
    const res = await regenerateShareToken(row.id);
    const token = res.share_token;
    row.share_token = token;
    const shareUrl = `${window.location.origin}/r/${token}`;
    const ok = await copyText(shareUrl);
    if (ok) {
      MessagePlugin.success('链接已重置，旧链接已失效，新链接已复制');
    } else {
      MessagePlugin.warning('复制失败，请手动复制：' + shareUrl);
    }
  } catch (error: any) {
    MessagePlugin.error(error.message || '重置链接失败');
  }
};

// 分页变化
const rehandlePageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  getOutsourcingRecords();
};

const getStatusText = (status: string) => {
  const dictItem = statusDictList.value.find(item => item.dict_value === status);
  return dictItem?.dict_label || status;
};

const getStatusTheme = (status: string) => {
  const dictItem = statusDictList.value.find(item => item.dict_value === status);
  return dictItem?.color || 'default';
};

// 查看工单工序
const handleViewWorkOrderProcess = (workOrder: any) => {
  if (!workOrder?.id) {
    MessagePlugin.warning('工单信息不存在');
    return;
  }
  currentWorkOrder.value = workOrder;
  processDrawerVisible.value = true;
};

// 查看模具详情
const handleViewMold = (moldId: number) => {
  if (moldId) {
    router.push(`/mold/detail/${moldId}`);
  }
};

// 当前刚完成的工序步骤
const currentStepName = (row: any): string => {
  const processes = row.work_order_processes || [];
  if (processes.length === 0) return '';
  let lastName = '';
  processes.forEach((p: any) => {
    if (p.is_completed === 1) {
      const name = p.process_template?.process_name || p.process_name || '';
      if (name) lastName = name;
    }
  });
  return lastName;
};

// 点击工序胶囊打开工序详情抽屉（与工单列表行为一致）
const showProcessDetail = (row: any) => {
  if (row.work_order?.id) {
    currentWorkOrder.value = row.work_order;
    processDrawerVisible.value = true;
  }
};

// 工序操作后刷新
const handleProcessRefresh = async (workOrderId: number) => {
  // 重新加载外发记录列表
  await getOutsourcingRecords();
};

// 加工商选项（格式化标签）
const processorOptions = computed(() =>
  processorList.value.map((p: any) => ({
    value: p.id,
    label: p.process_type ? `${p.processor_name} - ${p.process_type}` : p.processor_name,
  })),
);

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'recordCode', placeholder: '请输入外发单号' },
  { type: 'select', field: 'status', placeholder: '请选择状态', options: statusDictList, labelKey: 'dict_label', valueKey: 'dict_value' },
  { type: 'select', field: 'processorId', placeholder: '请选择加工商', options: processorOptions },
];

onMounted(() => {
  getProcessorListData();
  getOutsourcingRecords();

});

// ✨ KeepAlive 缓存下返回页面时自动刷新列表
onActivated(() => {
  getOutsourcingRecords();
});

// 打开列设置对话框
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
};
</script>

<style scoped>
.outsourcing-mold-container {
  padding: 16px;
}

.search-card,
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

.record-code-link {
  font-weight: 600;
  font-size: 14px;
}

/* 进度记录 */
.report-list {
  margin-top: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.report-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.report-time {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--td-text-color-secondary);
  font-size: 13px;
  white-space: nowrap;
  min-width: 160px;
  padding-top: 4px;
}

.report-time .dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--td-brand-color);
}

.report-body {
  flex: 1;
  background: var(--td-bg-color-container-hover);
  border-radius: 6px;
  padding: 12px 16px;
  font-size: 14px;
  line-height: 1.8;
}

.report-defect {
  color: #e34d59;
}

.report-remark {
  color: var(--td-text-color-secondary);
}

.progress-cell {
  min-width: 120px;
}

.progress-text {
  font-size: 12px;
  color: var(--td-text-color-secondary);
}

.processor-sub {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  margin-top: 2px;
}

/* 工序进度 */
.process-summary {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 180px;
  height: 100%;
  padding: 0;
  cursor: pointer;
}

.process-row-capsules {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 4px 8px;
  border-radius: 6px;
  white-space: nowrap;
  transition: background 0.3s ease;
}

.process-summary:hover .process-row-capsules {
  background: var(--td-bg-color-secondary, #f5f5f5);
  cursor: pointer;
}

.group-processes-list {
  display: inline-flex;
  gap: 5px;
}

.group-arrow {
  color: var(--td-text-color-placeholder, #ccc);
  font-size: 14px;
  font-weight: bold;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
}

.process-name-item {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 3px 9px;
  border-radius: 14px;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  line-height: 1.2;
  text-align: center;
  flex-shrink: 0;
  min-width: fit-content;
  transition: all 0.3s ease;
}

.process-completed {
  background: var(--td-success-color-1, #e3f9e9);
  color: var(--td-success-color, #00a870);
  border: 1px solid var(--td-success-color-2, #b3ebd2);
}

.process-pending {
  background: var(--td-bg-color-component, #f0f0f0);
  color: var(--td-text-color-placeholder, #999);
  border: 1px solid var(--td-border-level-2-color, #e0e0e0);
}

.process-skipped {
  background: #fafafa;
  color: #bbb;
  border: 1px dashed #ddd;
}

.process-empty {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
}

/* 工序胶囊（和工单列表一致） */
.process-capsule {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 4px 12px;
  border-radius: 16px;
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
  line-height: 1.2;
  cursor: pointer;
  transition: all 0.3s ease;
}

.process-capsule.process-completed {
  background: var(--td-success-color-1, #e3f9e9);
  color: var(--td-success-color, #00a870);
  border: 1px solid var(--td-success-color-2, #b3ebd2);
}

.process-capsule.process-completed:hover {
  background: var(--td-success-color-2, #b3ebd2);
}

.process-capsule.process-pending {
  background: var(--td-bg-color-component, #f0f0f0);
  color: var(--td-text-color-placeholder, #999);
  border: 1px solid var(--td-border-level-2-color, #e0e0e0);
}

.process-capsule.process-pending:hover {
  background: var(--td-bg-color-secondary, #e7e7e7);
}

.work-order-link {
  cursor: pointer;
  font-weight: 500;
}

/* 关联信息 */
.related-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
}

/* 模具编号芯片 */
.mold-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 2px;
  margin-left: 0;
  padding-left: 0;
}

.mold-chip {
  display: inline-block;
  font-size: 13px;
  font-weight: 600;
  line-height: 1;
  padding: 5px 12px;
  border-radius: 12px;
  text-align: left;
  white-space: nowrap;
  background: #e8f5e9;
  color: #2ba471;
  border: 1px solid #c8e6c9;
  cursor: pointer;
  transition: opacity 0.2s;
}

.mold-chip:hover {
  opacity: 0.75;
}

/* 工序详情抽屉样式 */
.process-drawer-content {
  padding: 16px;
}

.process-groups-drawer {
  margin-bottom: 16px;
}

.process-group-card {
  border: 1px solid var(--td-border-level-2-color);
  border-radius: 8px;
  margin-bottom: 12px;
  overflow: hidden;
}

.group-card-header {
  background: var(--td-bg-color-container);
  padding: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--td-border-level-1-color);
  margin-bottom: 16px;
}

.group-card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.group-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  background: var(--td-brand-color);
  color: #fff;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 600;
}

.group-hint {
  color: var(--td-text-color-secondary);
  font-size: 12px;
}

.group-card-processes {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
  padding: 0 16px 16px 16px;
}

.process-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 12px;
  background: var(--td-bg-color-container-hover);
  border-radius: 6px;
  transition: all 0.3s ease;
  border: 1px solid var(--td-border-level-1-color);
}

.process-item:last-child {
  margin-bottom: 0;
}

/* 已完成工序 - 绿色边框 */
.process-item-completed {
  border-color: var(--td-success-color) !important;
}

/* 跳过工序 - 灰色虚线边框 */
.process-item-skipped {
  border-color: var(--td-gray-color-5) !important;
  border-style: dashed !important;
  opacity: 0.7;
}

/* 进行中工序 - 蓝色边框 */
.process-item-active {
  border-color: var(--td-brand-color) !important;
}

/* 待开工工序 - 灰色边框 */
.process-item-pending {
  border-color: var(--td-gray-color-3) !important;
}

.process-item:hover {
  background: var(--td-bg-color-container-active);
}

.process-item-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.process-item-name {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  flex: 1;
}

.process-item-action {
  margin-top: 12px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.action-buttons {
  display: flex;
  gap: 8px;
  align-items: center;
}

.action-btn-complete,
.action-btn-skip {
  min-width: 80px;
  padding: 4px 16px !important;
  font-size: 13px !important;
  border-radius: 20px !important;
  transition: all 0.2s ease;
}

.action-btn-cancel {
  font-size: 12px !important;
  color: var(--td-text-color-secondary) !important;
  padding: 4px 8px !important;
  transition: all 0.2s ease;
}

.process-item-time {
  font-size: 14px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  white-space: nowrap;
}

.group-arrow-divider {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 8px 0;
  color: var(--td-text-color-placeholder);
}

.process-stats {
  margin-top: 16px;
}

.stats-container {
  display: flex;
  justify-content: space-around;
  text-align: center;
}

.stat-item {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--td-brand-color);
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: var(--td-text-color-secondary);
}
</style>

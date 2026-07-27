<template>
  <div class="outsourcing-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">外发管理</span>
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
        <template #outsourceType="{ row }">
          <t-tag shape="round" variant="outline" :theme="row.outsource_type === 'PRODUCTION' ? 'primary' : 'warning'">
            {{ row.outsource_type === 'PRODUCTION' ? '生产外发' : '模具外发' }}
          </t-tag>
        </template>
        <template #relatedInfo="{ row }">
          <div v-if="row.mold?.mold_code" class="mold-tags">
            <span class="mold-chip" @click.stop="handleViewMold(row.mold?.id)">
              {{ row.mold.mold_code }}
            </span>
          </div>
          <span v-else-if="row.mold_id" style="color: #999; font-size: 13px;">-</span>
          <span v-else style="color: #999; font-size: 13px;">-</span>
        </template>
        <template #status="{ row }">
          <t-tag :theme="getStatusTheme(row.status)">
            {{ getStatusText(row.status) }}
          </t-tag>
        </template>
        <template #sendTime="{ row }">
          {{ row.send_time ? formatDateTime(row.send_time) : '-' }}
        </template>
        <template #receiveTime="{ row }">
          {{ row.receive_time ? formatDateTime(row.receive_time) : '-' }}
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
              v-if="row.status !== 'RETURNED'" 
              theme="warning" 
              @click="handleAddReport(row)"
            >
              进度汇报
            </t-link>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 查看详情对话框 -->
    <t-dialog
      v-model:visible="detailVisible"
      title="外发详情"
      width="900px"
      :footer="null"
     :close-on-overlay-click="false">
      <t-descriptions bordered :column="2">
        <t-descriptions-item label="外发单号">{{ detailData.record_code }}</t-descriptions-item>
        <t-descriptions-item label="外发类型">
          <t-tag shape="round" variant="outline" :theme="detailData.outsource_type === 'PRODUCTION' ? 'primary' : 'warning'">
            {{ detailData.outsource_type === 'PRODUCTION' ? '生产外发' : '模具外发' }}
          </t-tag>
        </t-descriptions-item>
        <t-descriptions-item label="订单编号">
          {{ detailData.order?.order_code || detailData.mold_work_order?.work_order_code || '-' }}
        </t-descriptions-item>
        <t-descriptions-item label="加工商">{{ detailData.processor?.name || '-' }}</t-descriptions-item>
        <t-descriptions-item label="外发数量">{{ detailData.quantity || '-' }}</t-descriptions-item>
        <t-descriptions-item label="状态">
          <t-tag :theme="getStatusTheme(detailData.status)">
            {{ getStatusText(detailData.status) }}
          </t-tag>
        </t-descriptions-item>
        <t-descriptions-item label="发出时间">{{ detailData.send_time ? formatDateTime(detailData.send_time) : '-' }}</t-descriptions-item>
        <t-descriptions-item label="收回时间">{{ detailData.receive_time ? formatDateTime(detailData.receive_time) : '-' }}</t-descriptions-item>
        <t-descriptions-item label="备注" :span="2">
          <div style="white-space: pre-wrap;">{{ detailData.remark || '-' }}</div>
        </t-descriptions-item>
      </t-descriptions>

      <!-- 进度记录 -->
      <div v-if="reportList.length > 0" class="report-list">
        <t-divider>进度记录</t-divider>
        <div
          v-for="(report, index) in reportList"
          :key="index"
          class="report-item"
        >
          <div class="report-time">
            <span class="dot" />
            <span>{{ formatDateTime(report.report_time) }}</span>
          </div>
          <div class="report-body">
            <div><strong>完成进度：</strong>{{ report.progress }}%</div>
            <div v-if="report.completed_qty"><strong>完成数量：</strong>{{ report.completed_qty }}</div>
            <div v-if="report.remark" class="report-remark">{{ report.remark }}</div>
          </div>
        </div>
      </div>
    </t-dialog>

    <!-- 发出确认对话框 -->
    <t-dialog
      v-model:visible="sendDialogVisible"
      title="确认发出"
      width="500px"
      @confirm="confirmSend"
     :close-on-overlay-click="false">
      <t-form :data="sendForm" :label-width="100">
        <t-form-item label="发出时间">
          <t-date-picker
            v-model="sendForm.sendTime"
            enable-time-picker
            format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="sendForm.remark" placeholder="选填" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 收回确认对话框 -->
    <t-dialog
      v-model:visible="receiveDialogVisible"
      title="确认收回"
      width="500px"
      @confirm="confirmReceive"
     :close-on-overlay-click="false">
      <t-form :data="receiveForm" :label-width="100">
        <t-form-item label="收回时间">
          <t-date-picker
            v-model="receiveForm.receiveTime"
            enable-time-picker
            format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item label="合格数量">
          <t-input-number v-model="receiveForm.okQty" :min="0" placeholder="请输入合格数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="不合格数量">
          <t-input-number v-model="receiveForm.badQty" :min="0" placeholder="请输入不合格数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="receiveForm.remark" placeholder="选填" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 进度汇报对话框 -->
    <t-dialog
      v-model:visible="reportDialogVisible"
      title="进度汇报"
      width="500px"
      @confirm="confirmReport"
     :close-on-overlay-click="false">
      <t-form :data="reportForm" :label-width="100">
        <t-form-item label="汇报时间">
          <t-date-picker
            v-model="reportForm.reportTime"
            enable-time-picker
            format="YYYY-MM-DD HH:mm:ss"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item label="完成进度(%)">
          <t-slider v-model="reportForm.progress" :min="0" :max="100" />
          <div style="text-align: center; margin-top: 8px">{{ reportForm.progress }}%</div>
        </t-form-item>
        <t-form-item label="完成数量">
          <t-input-number v-model="reportForm.completedQty" :min="0" placeholder="请输入完成数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="reportForm.remark" placeholder="选填" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>
    <!-- 列设置对话框 - 使用通用组件 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="outsourcing_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { onActivated, onMounted, reactive, ref, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TableProps } from 'tdesign-vue-next';
import { RefreshIcon, DownloadIcon, SettingIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import dayjs from 'dayjs';
import { formatDate, formatDateTime } from '@/utils/date';
import { getProcessorList } from '@/api/reference';
import { request } from '@/utils/request';
import { useDictRaw } from '@/hooks/useDict';
import { useRoute, useRouter } from 'vue-router';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';

const route = useRoute();
const router = useRouter();

defineOptions({
  name: 'OutsourcingManagement',
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
  recordCode: '',
  outsourceType: undefined,
  status: undefined,
  processorId: undefined,
});

// 加工商列表
const processorList = ref<any[]>([]);

// 外发状态字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: statusDictList } = useDictRaw('outsourcing-status');

// 表格列配置
const allColumns: ColumnOption[] = [
  { colKey: 'record_code', title: '外发单号', pinned: true },
  { colKey: 'outsource_type', title: '外发类型' },
  { colKey: 'relatedInfo', title: '关联信息' },
  { colKey: 'processor_id', title: '加工商' },
  { colKey: 'quantity', title: '数量' },
  { colKey: 'status', title: '状态' },
  { colKey: 'send_time', title: '发出时间' },
  { colKey: 'receive_time', title: '收回时间' },
  { colKey: 'create_time', title: '创建时间' },
  { colKey: 'operation', title: '操作', pinned: true },
];

const DEFAULT_COLUMNS = ['record_code', 'outsource_type', 'relatedInfo', 'processor_id', 'quantity', 'status', 'send_time', 'receive_time', 'create_time', 'operation'];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([...DEFAULT_COLUMNS]);

const displayColumns = computed(() => {
  return allColumns.filter(col => selectedColumns.value.includes(col.colKey));
});

// 表格列配置（旧版，保留以兼容模板引用）
const columns: TableProps['columns'] = allColumns;

// 对话框相关
const detailVisible = ref(false);
const detailData = ref<any>({});
const reportList = ref<any[]>([]);

const sendDialogVisible = ref(false);
const sendForm = reactive({
  id: null,
  sendTime: dayjs().format('YYYY-MM-DD HH:mm:ss'),
  remark: '',
});

const receiveDialogVisible = ref(false);
const receiveForm = reactive({
  id: null,
  receiveTime: dayjs().format('YYYY-MM-DD HH:mm:ss'),
  okQty: 0,
  badQty: 0,
  remark: '',
});

const reportDialogVisible = ref(false);
const reportForm = reactive({
  id: null,
  reportTime: dayjs().format('YYYY-MM-DD HH:mm:ss'),
  progress: 0,
  completedQty: 0,
  remark: '',
});

const getOutsourcingRecords = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize,
    };
    
    if (formData.recordCode) {
      params.record_code = formData.recordCode;
    }
    if (formData.outsourceType !== undefined && formData.outsourceType !== null) {
      params.outsource_type = formData.outsourceType;
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
    pagination.total = res?.total ?? 0;
  } catch (error) {
    MessagePlugin.error('获取外发记录失败');
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
  formData.outsourceType = undefined;
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
  const exportData: Record<string, any>[] = tableData.value.map(item => ({
    '外发单号': item.record_code,
    '外发类型': item.outsource_type === 1 ? '生产外发' : '模具外发',
    '加工商': item.processor?.name || '-',
    '数量': item.quantity || '-',
    '状态': getStatusText(item.status),
    '发出时间': item.send_time ? formatDateTime(item.send_time) : '-',
    '收回时间': item.receive_time ? formatDateTime(item.receive_time) : '-',
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
  link.setAttribute('download', `外发记录_${new Date().getTime()}.csv`);
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  MessagePlugin.success('导出成功');
};

// 查看详情
const handleView = async (row: any) => {
  detailData.value = row;
  
  try {
    const res = await request.get({
      url: `/production/outsourcing/${row.id}/reports`,
    });
    reportList.value = res || [];
  } catch (error) {
    console.error('获取进度汇报失败:', error);
    reportList.value = [];
  }
  
  detailVisible.value = true;
};

// 查看模具详情
const handleViewMold = (moldId: number) => {
  if (moldId) {
    router.push(`/mold/detail/${moldId}`);
  }
};

// 发出
const handleSend = (row: any) => {
  sendForm.id = row.id;
  sendForm.sendTime = dayjs().format('YYYY-MM-DD HH:mm:ss');
  sendForm.remark = '';
  sendDialogVisible.value = true;
};

// 确认发出
const confirmSend = async () => {
  try {
    await request.post({
      url: '/production/outsourcing/send',
      data: {
        id: sendForm.id,
        send_time: sendForm.sendTime,
        remark: sendForm.remark,
      },
    });
    MessagePlugin.success('发出成功');
    sendDialogVisible.value = false;
    getOutsourcingRecords();
  } catch (error: any) {
    MessagePlugin.error(error.message || '发出失败');
  }
};

// 收回
const handleReceive = (row: any) => {
  receiveForm.id = row.id;
  receiveForm.receiveTime = dayjs().format('YYYY-MM-DD HH:mm:ss');
  receiveForm.okQty = 0;
  receiveForm.badQty = 0;
  receiveForm.remark = '';
  receiveDialogVisible.value = true;
};

// 确认收回
const confirmReceive = async () => {
  try {
    await request.post({
      url: '/production/outsourcing/receive',
      data: {
        id: receiveForm.id,
        receive_time: receiveForm.receiveTime,
        ok_qty: receiveForm.okQty,
        bad_qty: receiveForm.badQty,
        remark: receiveForm.remark,
      },
    });
    MessagePlugin.success('收回成功');
    receiveDialogVisible.value = false;
    getOutsourcingRecords();
  } catch (error: any) {
    MessagePlugin.error(error.message || '收回失败');
  }
};

const handleAddReport = (row: any) => {
  reportForm.id = row.id;
  reportForm.reportTime = dayjs().format('YYYY-MM-DD HH:mm:ss');
  reportForm.progress = 0;
  reportForm.completedQty = 0;
  reportForm.remark = '';
  reportDialogVisible.value = true;
};

// 确认进度汇报
const confirmReport = async () => {
  try {
    await request.post({
      url: '/production/outsourcing/report',
      data: {
        outsourcing_id: reportForm.id,
        report_time: reportForm.reportTime,
        progress: reportForm.progress,
        completed_qty: reportForm.completedQty,
        remark: reportForm.remark,
      },
    });
    MessagePlugin.success('汇报成功');
    reportDialogVisible.value = false;
    getOutsourcingRecords();
  } catch (error: any) {
    MessagePlugin.error(error.message || '汇报失败');
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

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('outsourcing_table_columns');
    if (saved) {
      const savedData = JSON.parse(saved);
      if (Array.isArray(savedData)) {
        const validColumns = savedData.filter((colKey: string) => 
          allColumns.some(col => col.colKey === colKey)
        );
        selectedColumns.value = validColumns.length > 0 ? validColumns : [...DEFAULT_COLUMNS];
      } else if (savedData.visible) {
        const validColumns = savedData.visible.filter((colKey: string) => 
          allColumns.some(col => col.colKey === colKey)
        );
        selectedColumns.value = validColumns.length > 0 ? validColumns : [...DEFAULT_COLUMNS];
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

// 打开列设置对话框
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 列配置变化时的回调
const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
};

// 外发类型选项
const outsourceTypeOptions = [
  { label: '生产外发', value: 'PRODUCTION' },
  { label: '模具外发', value: 'MOLD' },
];

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'recordCode', placeholder: '请输入外发单号' },
  { type: 'select', field: 'outsourceType', placeholder: '请选择外发类型', options: outsourceTypeOptions },
  { type: 'select', field: 'status', placeholder: '请选择状态', options: statusDictList, labelKey: 'dict_label', valueKey: 'dict_value' },
  { type: 'select', field: 'processorId', placeholder: '请选择加工商', options: processorList, labelKey: 'displayName', valueKey: 'id' },
];

const loadProcessorList = async () => {
  try {
    const res = await getProcessorList({ page: 1, page_size: 200 });
    processorList.value = res.list || [];
  } catch (error) {
    console.error('加载加工商列表失败:', error);
  }
};

onMounted(() => {
  loadProcessorList();
  
  const moldId = route.query.moldId;
  if (moldId) {
    // 可以添加根据模具ID筛选的逻辑
  }
  
  getOutsourcingRecords();
  
  loadColumnConfig();
});

// ✨ KeepAlive 缓存下返回页面时自动刷新列表
onActivated(() => {
  getOutsourcingRecords();
});
</script>

<style scoped>
.outsourcing-container {
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

/* 模具编号芯片（与生产外发页一致） */
.mold-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
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
</style>

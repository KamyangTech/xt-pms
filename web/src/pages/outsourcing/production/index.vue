<template>
  <div class="outsourcing-production-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">生产外发管理</span>
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
          <t-link theme="primary" @click.stop="handleViewOrder(row.order?.id)" class="order-link">
            订单: {{ row.order?.order_code || '-' }}
          </t-link>
        </template>
        <template #productMold="{ row }">
          <div class="product-mold-cell">
            <div class="product-mold-name">{{ row.order?.product_name || '-' }}</div>
            <div v-if="row.mold_items && row.mold_items.length > 0" class="mold-tags">
              <template v-for="(m, idx) in row.mold_items" :key="idx">
                <span
                  v-if="m.id && m.id > 0"
                  class="mold-chip mold-chip-clickable"
                  @click.stop="handleViewMold(m.id)"
                >
                  {{ m.mold_code }}
                </span>
                <span
                  v-else
                  class="mold-chip mold-chip-external"
                >
                  {{ m.mold_code }}
                </span>
              </template>
            </div>
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
          <t-tooltip :content="`${row.progress_qty || 0} / ${row.quantity || 0}`" placement="top">
            <t-progress
              :percentage="row.progress_percent || 0"
              :stroke-width="10"
              :color="row.progress_percent >= 100 ? '#2ba471' : '#0052d9'"
            />
          </t-tooltip>
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
              v-if="row.status !== 'RETURNED'"
              theme="warning" 
              @click="handleAddReport(row)"
            >
              进度汇报
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
      title="生产外发详情"
      width="900px"
      placement="center"
      :footer="null"
     :close-on-overlay-click="false">
      <t-descriptions bordered :column="2">
        <t-descriptions-item label="外发单号">{{ detailData.record_code }}</t-descriptions-item>
        <t-descriptions-item label="外发类型">
          <t-tag shape="round" variant="outline" theme="primary">
            生产外发
          </t-tag>
        </t-descriptions-item>
        <t-descriptions-item label="订单编号">
          {{ detailData.order?.order_code || '-' }}
        </t-descriptions-item>
        <t-descriptions-item label="产品名称">
          {{ detailData.order?.product_name || '-' }}
        </t-descriptions-item>
        <t-descriptions-item v-if="detailData.mold_codes && detailData.mold_codes.length > 0" label="模具编号">
          <div class="mold-tags">
            <template v-for="(m, idx) in detailData.mold_items || []" :key="idx">
              <span
                v-if="m.id && m.id > 0"
                class="mold-chip mold-chip-clickable"
                @click="handleViewMold(m.id)"
              >
                {{ m.mold_code }}
              </span>
              <span
                v-else
                class="mold-chip mold-chip-external"
              >
                {{ m.mold_code }}
              </span>
            </template>
          </div>
        </t-descriptions-item>
        <t-descriptions-item label="加工商">
          <span v-if="detailData.processor">
            {{ detailData.processor.process_type ? `${detailData.processor.processor_name} - ${detailData.processor.process_type}` : detailData.processor.processor_name }}
          </span>
          <span v-else>-</span>
        </t-descriptions-item>
        <t-descriptions-item label="外发数量">{{ detailData.quantity || '-' }}</t-descriptions-item>
        <t-descriptions-item label="外发金额">{{ detailData.total_amount || '-' }}</t-descriptions-item>
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
            <span>{{ formatDateTime(report.report_date) }}</span>
          </div>
          <div class="report-body">
            <div><strong>合格数量：</strong>{{ report.ok_qty }}</div>
            <div><strong>不合格数量：</strong>{{ report.bad_qty }}</div>
            <div v-if="report.defect_reason" class="report-defect">
              <strong>不良原因：</strong>{{ report.defect_reason }}
            </div>
            <div v-if="report.remark" class="report-remark">
              <strong>备注：</strong>{{ report.remark }}
            </div>
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
        <t-form-item label="合格数量">
          <t-input-number v-model="reportForm.okQty" :min="0" placeholder="请输入合格数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="不合格数量">
          <t-input-number v-model="reportForm.badQty" :min="0" placeholder="请输入不合格数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="不良原因">
          <t-input v-model="reportForm.defectReason" placeholder="选填" />
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
      storage-key="outsourcing_production_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { onActivated, onMounted, reactive, ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import type { TableProps } from 'tdesign-vue-next';
import { RefreshIcon, DownloadIcon, SettingIcon, LinkIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import dayjs from 'dayjs';
import { getProcessorList } from '@/api/reference';
import { request } from '@/utils/request';
import { generateShareToken, regenerateShareToken } from '@/api/production/order';
import { formatDate, formatDateTime } from '@/utils/date';
import { OPERATION_COLUMN_WIDTH } from '@/utils/operation-column-width';
import { copyText } from '@/utils/clipboard';
import { useDictRaw } from '@/hooks/useDict';
import ColumnSettingDialog from '@/components/ColumnSettingDialog/index.vue';

defineOptions({
  name: 'OutsourcingProduction',
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

const getProcessorListData = async () => {
  try {
    const res = await getProcessorList();
    processorList.value = res.list || [];
  } catch (error) {
    console.error('获取加工商列表失败:', error);
  }
};

// 所有可选列配置（用于列设置对话框）
const allColumns = [
  { colKey: 'record_code', title: '外发单号', pinned: true },  // ✨ 固定列：不可隐藏/移动
  { colKey: 'relatedInfo', title: '关联信息' },
  { colKey: 'productMold', title: '产品与模具' },
  { colKey: 'processor', title: '加工商' },
  { colKey: 'quantity', title: '数量' },
  { colKey: 'progress', title: '进度' },
  { colKey: 'total_amount', title: '外发金额' },
  { colKey: 'status', title: '状态' },
  { colKey: 'outsourceDate', title: '外发日期' },
  { colKey: 'expectedReturnDate', title: '预计返回' },
  { colKey: 'actualReturnDate', title: '实际返回' },
  { colKey: 'createTime', title: '创建时间' },
  { colKey: 'operation', title: '操作', pinned: true },  //  固定列：不可隐藏/移动
];

// 表格列配置（用于 TDesign 表格）
const tableColumns: TableProps['columns'] = [
  { colKey: 'record_code', title: '外发单号', width: 180, fixed: 'left' as const },
  { colKey: 'relatedInfo', title: '关联信息', width: 200 },
  { colKey: 'productMold', title: '产品与模具', width: 260 },
  { colKey: 'processor', title: '加工商', width: 180 },
  { colKey: 'quantity', title: '数量', width: 80 },
  { colKey: 'progress', title: '进度', width: 120 },
  { colKey: 'total_amount', title: '外发金额', width: 100 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'outsourceDate', title: '外发日期', width: 120 },
  { colKey: 'expectedReturnDate', title: '预计返回', width: 120 },
  { colKey: 'actualReturnDate', title: '实际返回', width: 120 },
  { colKey: 'createTime', title: '创建时间', width: 120 },
  { colKey: 'operation', title: '操作', width: 280, fixed: 'right' as const },
];

// 默认显示的列（不包括外发单号，因为它固定显示）
const DEFAULT_COLUMNS = ['record_code', 'relatedInfo', 'productMold', 'processor', 'quantity', 'progress', 'total_amount', 'status', 'outsourceDate', 'expectedReturnDate', 'actualReturnDate', 'createTime', 'operation'];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([...DEFAULT_COLUMNS]);

const displayColumns = computed(() => {
  // 根据 selectedColumns 过滤 tableColumns
  return tableColumns.filter(col => selectedColumns.value.includes(col.colKey as string));
});

// 对话框相关
const detailVisible = ref(false);
const detailData = ref<any>({});
const reportList = ref<any[]>([]);

const sendDialogVisible = ref(false);
const sendForm = reactive({
  id: null,
  remark: '',
});

const receiveDialogVisible = ref(false);
const receiveForm = reactive({
  id: null,
  remark: '',
});

const reportDialogVisible = ref(false);
const reportForm = reactive({
  id: null,
  okQty: 0,
  badQty: 0,
  defectReason: '',
  remark: '',
});

const getOutsourcingRecords = async () => {
  loading.value = true;
  try {
    const params: any = {
      outsource_type: 'PRODUCTION', // 固定为生产外发
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
    MessagePlugin.error('获取生产外发记录失败');
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
    '订单编号': item.order?.order_code || '-',
    '产品名称': item.order?.product_name || '-',
    '加工商': item.processor ? (item.processor.process_type ? `${item.processor.processor_name} - ${item.processor.process_type}` : item.processor.processor_name) : '-',
    '数量': item.quantity || '-',
    '外发金额': item.total_amount || '-',
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
  link.setAttribute('download', `生产外发记录_${new Date().getTime()}.csv`);
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);

  MessagePlugin.success('导出成功');
};

// 查看订单详情
const handleViewOrder = (orderId: number) => {
  if (orderId) {
    router.push(`/production/order/detail/${orderId}`);
  }
};

// 查看模具详情（仅自有模具可点击跳转；外部模具 ExternalMoldCode 不可点击）
const handleViewMold = (moldId: number) => {
  if (moldId) {
    router.push(`/mold/detail/${moldId}`);
  }
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

// 发出
const handleSend = (row: any) => {
  sendForm.id = row.id;
  sendForm.remark = '';
  sendDialogVisible.value = true;
};

// 确认发出
const confirmSend = async () => {
  try {
    await request.post({
      url: `/production/outsourcing/${sendForm.id}/send`,
      data: {},
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
  receiveForm.remark = '';
  receiveDialogVisible.value = true;
};

// 确认收回
const confirmReceive = async () => {
  try {
    await request.post({
      url: `/production/outsourcing/${receiveForm.id}/receive`,
      data: {},
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
  reportForm.okQty = 0;
  reportForm.badQty = 0;
  reportForm.defectReason = '';
  reportForm.remark = '';
  reportDialogVisible.value = true;
};

// 确认进度汇报
const confirmReport = async () => {
  try {
    await request.post({
      url: `/production/outsourcing/${reportForm.id}/report`,
      data: {
        outsourcing_id: reportForm.id,
        report_date: dayjs().format('YYYY-MM-DD'),
        ok_qty: reportForm.okQty,
        bad_qty: reportForm.badQty,
        defect_reason: reportForm.defectReason,
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

const { items: statusDictList } = useDictRaw('outsourcing-status');

const getStatusText = (status: string) => {
  const item = statusDictList.value.find((d: any) => d.dict_value === status);
  return item ? item.dict_label : status;
};

const getStatusTheme = (status: string) => {
  const themeMap: Record<string, any> = {
    'PENDING': 'warning',
    'SENT': 'primary',
    'PRODUCING': 'primary',
    'RETURNED': 'success',
  };
  return themeMap[status] || 'default';
};

// 外发状态选项
const productionStatusOptions = [
  { label: '待发出', value: 'PENDING' },
  { label: '已发出', value: 'SENT' },
  { label: '生产中', value: 'PRODUCING' },
  { label: '已返回', value: 'RETURNED' },
];

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
  { type: 'select', field: 'status', placeholder: '请选择状态', options: productionStatusOptions },
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
.outsourcing-production-container {
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

/* 关联信息（仅订单） */
.related-info {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
}

/* 产品与模具列 */
.product-mold-cell {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
}

.product-mold-name {
  color: var(--td-text-color-primary);
  font-size: 14px;
  line-height: 22px;
  padding-left: 1em;
}

/* 模具编号标签（自定义 chip，保证与产品名称左对齐） */
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
}

/* 自有模具：绿色实心 */
.mold-chip-clickable {
  background: #e8f5e9;
  color: #2ba471;
  border: 1px solid #c8e6c9;
  cursor: pointer;
  transition: opacity 0.2s;
}

.mold-chip-clickable:hover {
  opacity: 0.75;
}

/* 外部模具：黄色描边 */
.mold-chip-external {
  background: #fff8e6;
  color: #d97a00;
  border: 1px solid #ffd591;
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
</style>

<template>
  <div class="master-order-list">
    <!-- 列表 -->
    <t-table
      :data="tableData"
      :columns="displayColumns"
      :loading="loading"
      row-key="id"
      :stripe="false"
      :pagination="pagination"
      @page-change="handlePageChange"
    >
      <template #orderType="{ row }">
        <t-tag v-if="row.order_type === 'COMBINED'" theme="primary" variant="outline">模具+产品</t-tag>
        <t-tag v-else-if="row.order_type === 'MOLD_ONLY'" theme="warning" variant="outline">仅模具</t-tag>
        <t-tag v-else-if="row.order_type === 'PROD_ONLY'" theme="success" variant="outline">仅产品</t-tag>
        <t-tag v-else variant="outline">{{ row.order_type }}</t-tag>  <!-- ✨ 兜底显示 -->
      </template>
      <template #mold_order_count="{ row }">
        <span v-if="row.mold_order_count > 0" style="color: #0052d9; font-weight: 500">
          {{ row.mold_order_count }}个
        </span>
        <span v-else style="color: #999">-</span>
      </template>
      <template #production_order_count="{ row }">
        <span v-if="row.production_order_count > 0" style="color: #00a870; font-weight: 500">
          {{ row.production_order_count }}个
        </span>
        <span v-else style="color: #999">-</span>
      </template>
      <template #total_amount="{ row }">
        ¥ {{ row.total_amount || 0 }}
      </template>
      <template #plan_start_date="{ row }">
        {{ formatDate(row.plan_start_date) }}
      </template>
      <template #plan_end_date="{ row }">
        {{ formatDate(row.plan_end_date) }}
      </template>
      <template #remark_summary="{ row }">
        <t-tooltip v-if="row.remark_summary" :content="row.remark">
          {{ row.remark_summary }}
        </t-tooltip>
        <span v-else style="color: #999">-</span>
      </template>
      <template #create_time="{ row }">
        {{ formatDateTime(row.create_time) }}
      </template>
      <template #status="{ row }">
        <t-tag shape="round" :theme="getStatusTheme(row.status)">{{ getStatusText(row.status) }}</t-tag>
      </template>
      <template #op="{ row }">
        <div style="width: 200px;">
          <t-space :size="8">
            <t-link theme="primary" @click="handleView(row)">详情</t-link>
            <t-link v-if="row.status !== 'CANCELLED' && row.status !== 'COMPLETED'" theme="primary" @click="handleEdit(row)">编辑</t-link>
            <t-link v-if="row.status !== 'CANCELLED' && row.status !== 'COMPLETED'" theme="danger" @click="handleCancel(row)">取消</t-link>
          </t-space>
        </div>
      </template>
    </t-table>

    <!-- 编辑对话框 -->
    <t-dialog
      v-model:visible="editDialogVisible"
      header="编辑综合订单"
      :width="500"
      placement="center"
      :close-on-overlay-click="false"
      :on-confirm="handleEditConfirm"
      :on-cancel="() => { editDialogVisible = false; }"
    >
      <t-form :data="editForm" label-width="100px">
        <t-form-item label="交货日期" name="quick_plan_end">
          <t-date-picker v-model="editForm.quick_plan_end" placeholder="选择日期" />
        </t-form-item>
        <t-form-item label="备注" name="remark">
          <t-textarea v-model="editForm.remark" placeholder="请输入备注" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 取消对话框 -->
    <t-dialog
      v-model:visible="cancelDialogVisible"
      header="确认取消综合订单"
      :width="500"
      placement="center"
      :close-on-overlay-click="false"
      :on-confirm="handleCancelConfirm"
      :on-cancel="() => { cancelDialogVisible = false; }"
    >
      <p style="margin-bottom: 16px; color: var(--td-text-color-secondary);">
        取消后，未开工的子订单将一并取消；已开工的将被标记并禁止继续操作。已完成的生产记录保留（不可恢复）。
      </p>
      <t-form label-width="100px">
        <t-form-item label="取消原因" name="reason" :rules="[{ required: true, message: '请填写取消原因' }]">
          <t-textarea v-model="cancelReason" placeholder="请填写取消原因" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { formatDate, formatDateTime } from '@/utils/date';
import { getMasterOrderList, editMasterOrder, cancelMasterOrder } from '@/api/order-management';  // ✨ 导入 API
import { OPERATION_COLUMN_WIDTH } from '@/utils/operation-column-width';

const router = useRouter();
const loading = ref(false);
const tableData = ref([]);
const searchForm = ref({ orderCode: '', customerName: '' });
const pagination = ref({ current: 1, pageSize: 10, total: 0 });

// 编辑对话框
const editDialogVisible = ref(false);
const editRow = ref<any>(null);
const editForm = ref({ quick_plan_end: '', remark: '' });

// 取消对话框
const cancelDialogVisible = ref(false);
const cancelRow = ref<any>(null);
const cancelReason = ref('');

// Props - 从父组件接收筛选参数
interface Props {
  orderCode?: string;
  customerName?: string;
  createTimeStart?: string;
  createTimeEnd?: string;
  dueSoon?: string;
}
const props = withDefaults(defineProps<Props>(), {
  orderCode: '',
  customerName: '',
  createTimeStart: '',
  createTimeEnd: '',
  dueSoon: '',
});

const columns: PrimaryTableCol[] = [
  { colKey: 'order_code', title: '综合单号', width: 160, fixed: 'left' },  // ✨ 固定列
  { colKey: 'customer.cus_name', title: '客户名称', width: 120 },  // ✨ 修复：使用 cus_name
  { colKey: 'order_type', title: '下单类型', width: 120, cell: 'orderType' },  // ✨ 添加 cell 指定自定义模板
  { colKey: 'mold_order_count', title: '模具订单', width: 100 },
  { colKey: 'production_order_count', title: '生产订单', width: 100 },
  { colKey: 'plan_start_date', title: '计划开始', width: 120, cell: 'plan_start_date' },
  { colKey: 'plan_end_date', title: '订单交期', width: 120, cell: 'plan_end_date' },
  { colKey: 'total_amount', title: '总金额', width: 120, cell: 'total_amount' },
  { colKey: 'remark_summary', title: '备注', width: 200, ellipsis: true },
  { colKey: 'status', title: '状态', width: 100, cell: 'status' },  // ✨ 添加 cell 指定自定义模板
  { colKey: 'create_time', title: '创建时间', width: 160, cell: 'create_time' },  // ✨ 添加 cell 指定自定义模板
  { colKey: 'op', title: '操作', width: 200, fixed: 'right', cell: 'op' },  // ✨ 添加 cell 指定自定义模板
];

// 动态列配置（根据用户选择）
const displayColumns = computed(() => {
  const orderedColumns = selectedColumns.value
    .map(colKey => columns.find(col => col.colKey === colKey))
    .filter(col => col !== undefined);
  
  // 将操作列添加到最后
  const operationCol = columns.find(col => col.colKey === 'op');
  if (operationCol) {
    orderedColumns.push(operationCol);
  }
  
  return orderedColumns;
});

// 所有可选列配置（不包括操作列）
const allColumns = [
  { colKey: 'order_code', title: '综合单号', pinned: true },  // ✨ 固定列：不可隐藏/移动
  { colKey: 'customer.cus_name', title: '客户名称' },
  { colKey: 'order_type', title: '下单类型' },
  { colKey: 'mold_order_count', title: '模具订单' },
  { colKey: 'production_order_count', title: '生产订单' },
  { colKey: 'plan_start_date', title: '计划开始' },
  { colKey: 'plan_end_date', title: '订单交期' },
  { colKey: 'total_amount', title: '总金额' },
  { colKey: 'remark_summary', title: '备注' },
  { colKey: 'status', title: '状态' },
  { colKey: 'create_time', title: '创建时间' },
];

// 默认显示的列
const DEFAULT_COLUMNS = ['order_code', 'customer.cus_name', 'order_type', 'mold_order_count', 'production_order_count', 'plan_start_date', 'plan_end_date', 'total_amount', 'status', 'create_time'];

// 列设置相关
const selectedColumns = ref<string[]>([...DEFAULT_COLUMNS]);

const updateColumns = (cols: string[]) => {
  selectedColumns.value = cols;
};

const handleSearch = () => {
  pagination.value.current = 1;
  fetchData();
};

const handleReset = () => {
  pagination.value.current = 1;
  fetchData();
};

const handleRefresh = () => {
  fetchData();
};

const handlePageChange = (pageInfo: PageInfo) => {
  pagination.value.current = pageInfo.current;
  pagination.value.pageSize = pageInfo.pageSize;
  fetchData();
};

const handleView = (row: any) => {
  router.push(`/order-management/detail/${row.id}`);
};

// 编辑
const handleEdit = (row: any) => {
  editRow.value = row;
  editForm.value = {
    quick_plan_end: row.quick_plan_end || '',
    remark: row.remark || '',
  };
  editDialogVisible.value = true;
};

// 确认编辑
const handleEditConfirm = async () => {
  try {
    await editMasterOrder(editRow.value.id, {
      plan_end_date: editForm.value.quick_plan_end || undefined,
      remark: editForm.value.remark,
    });
    MessagePlugin.success('编辑成功');
    editDialogVisible.value = false;
    fetchData();
  } catch (error) {
    console.error('编辑失败:', error);
    MessagePlugin.error('编辑失败');
  }
};

// 取消订单
const handleCancel = (row: any) => {
  cancelRow.value = row;
  cancelReason.value = '';
  cancelDialogVisible.value = true;
};

// 确认取消
const handleCancelConfirm = async () => {
  if (!cancelReason.value.trim()) {
    MessagePlugin.warning('请填写取消原因');
    return;
  }
  try {
    await cancelMasterOrder(cancelRow.value.id, { reason: cancelReason.value });
    MessagePlugin.success('订单已取消');
    cancelDialogVisible.value = false;
    fetchData();
  } catch (error) {
    console.error('取消失败:', error);
    MessagePlugin.error('取消失败');
  }
};

const getStatusTheme = (status: string) => {
  const map: Record<string, any> = {
    PENDING: 'default',
    MOLD_PRODUCING: 'warning',
    PRODUCING: 'primary',
    PENDING_SHIP: 'warning',
    COMPLETED: 'success',
    CANCELLED: 'danger',
  };
  return map[status] || 'default';
};

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    PENDING: '待开始',
    MOLD_PRODUCING: '模具制造中',
    PRODUCING: '生产中',
    PENDING_SHIP: '待发货',
    COMPLETED: '已完成',
    CANCELLED: '已取消',
  };
  if (status in map) return map[status];
  // 兼容旧数据中的 PROCESSING 状态（应视为待开始）
  if (status === 'PROCESSING') return '待开始';
  return status;
};

const fetchData = async () => {
  loading.value = true;
  try {
    // ✨ 调用后端 API - 使用 props 传入的筛选条件
    const params = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
      order_code: props.orderCode || undefined,
      customer_name: props.customerName || undefined,
      create_time_start: props.createTimeStart || undefined,
      create_time_end: props.createTimeEnd || undefined,
      due_soon: props.dueSoon || undefined,
    };
    const res = await getMasterOrderList(params);
    // ✨ 注意：响应拦截器已经解包了 data，所以 res 直接是 { list, total }
    tableData.value = res.list || [];
    pagination.value.total = res.total || 0;
  } catch (error) {
    console.error('获取订单列表失败:', error);
  } finally {
    loading.value = false;
  }
};

onMounted(fetchData);

// 暴露方法给父组件
defineExpose({
  fetchData,
  updateColumns,
});
</script>

<style scoped>
.master-order-list {
  /* 移除了 padding，由父容器控制 */
}
</style>

<template>
  <div class="mold-order-list">
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
      <template #order_code="{ row }">
        <t-link theme="default" @click="handleView(row)">
          {{ row.order_code }}
        </t-link>
      </template>
      <template #order_type="{ row }">
        <t-tag shape="round" variant="outline" :theme="getOrderTypeTheme(row.order_type)" :style="getOrderTypeStyle(row.order_type)">
          {{ getOrderTypeText(row.order_type) }}
        </t-tag>
      </template>
      <template #master_order_code="{ row }">
        <span v-if="row.master_order_code" style="color: #0052d9; cursor: pointer" @click="handleViewMasterOrder(row)">
          {{ row.master_order_code }}
        </span>
        <span v-else style="color: #999">-</span>
      </template>
      <template #priority="{ row }">
        <t-tag shape="round" :theme="getPriorityTheme(row.priority)" :style="getPriorityStyle(row.priority)">
          {{ getPriorityText(row.priority) }}
        </t-tag>
      </template>
      <template #customer="{ row }">
        {{ row.customer?.cus_name || '-' }}
      </template>
      <template #mold_codes="{ row }">
        <div v-if="row.mold_codes && row.mold_codes.length > 0" class="mold-codes-list">
          <div v-for="(code, index) in row.mold_codes" :key="index" class="mold-code-item">
            {{ code }}
          </div>
        </div>
        <span v-else>-</span>
      </template>
      <template #status="{ row }">
        <t-tag shape="round" :theme="getStatusTheme(row.status)" :style="getStatusStyle(row.status)">
          {{ getStatusText(row.status) }}
        </t-tag>
      </template>
      <template #plan_start_date="{ row }">
        {{ formatDate(row.plan_start_date) }}
      </template>
      <template #plan_end_date="{ row }">
        {{ formatDate(row.plan_end_date) }}
      </template>
      <template #deliver_time="{ row }">
        {{ formatDate(row.deliver_time) }}
      </template>
      <template #create_time="{ row }">
        {{ formatDate(row.create_time) }}
      </template>
      <template #total_amount="{ row }">
        ¥{{ row.total_amount?.toFixed(2) || '0.00' }}
      </template>
      <template #op="{ row }">
        <div style="width: 220px;">
          <t-space :size="8" break-line>
            <t-link theme="primary" @click="handleView(row)">详情</t-link>
            <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
            <t-link theme="success" @click="handlePaymentManage(row)">款项</t-link>
            <t-link 
              v-if="canDeliver(row)" 
              theme="warning" 
              @click="handleDeliver(row)"
            >
              交付
            </t-link>
            <t-popconfirm content="确认删除该订单？" @confirm="handleDelete(row)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </div>
      </template>
    </t-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { getMoldOrderList, deleteMoldOrder } from '@/api/mold';
import { useDictRaw } from '@/hooks/useDict';
import { getDictColorRenderConfig } from '@/utils/dict-color';
import { OPERATION_COLUMN_WIDTH } from '@/utils/operation-column-width';

const router = useRouter();
const loading = ref(false);
const tableData = ref<any[]>([]);
const pagination = ref({ current: 1, pageSize: 10, total: 0 });

// Props - 从父组件接收筛选参数
interface Props {
  createTimeStart?: string;
  createTimeEnd?: string;
  dueSoon?: string;
}
const props = withDefaults(defineProps<Props>(), {
  createTimeStart: '',
  createTimeEnd: '',
  dueSoon: '',
});

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: orderStatusList } = useDictRaw('mold-order-status', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '')
    .map((item: any) => ({
      value: item.dict_value,
      label: item.dict_label,
      color: item.color || 'default',
    })),
);
const { items: orderTypeList } = useDictRaw('mold_order_type', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '')
    .map((item: any) => ({
      value: item.dict_value,
      label: item.dict_label,
      color: item.color || 'default',
    })),
);
const { items: priorityList } = useDictRaw('order-priority', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '')
    .map((item: any) => ({
      value: item.dict_value,
      label: item.dict_label,
      color: item.color || 'default',
    })),
);

const columns: PrimaryTableCol[] = [
  { colKey: 'order_code', title: '订单编号', width: 180, fixed: 'left', cell: 'order_code' },  // ✨ 固定列
  { colKey: 'order_type', title: '订单类型', width: 100, cell: 'order_type' },
  { colKey: 'master_order_code', title: '来源订单', width: 150, cell: 'master_order_code' },
  { colKey: 'priority', title: '优先级', width: 100, cell: 'priority' },  // ✨ 添加优先级列
  { colKey: 'customer.cus_name', title: '客户名称', width: 120 },
  { colKey: 'mold_codes', title: '模具编号', width: 180, cell: 'mold_codes' },
  { colKey: 'plan_start_date', title: '计划开始', width: 120, cell: 'plan_start_date' },
  { colKey: 'plan_end_date', title: '订单交期', width: 120, cell: 'plan_end_date' },
  { colKey: 'total_amount', title: '总金额', width: 120, cell: 'total_amount' },
  { colKey: 'status', title: '状态', width: 100, cell: 'status' },
  { colKey: 'deliver_time', title: '交付时间', width: 160, cell: 'deliver_time' },
  { colKey: 'create_time', title: '创建时间', width: 160, cell: 'create_time' },
  { colKey: 'op', title: '操作', width: OPERATION_COLUMN_WIDTH.LARGE, fixed: 'right', cell: 'op' },
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
  { colKey: 'order_code', title: '订单编号', pinned: true },  // ✨ 固定列：不可隐藏/移动
  { colKey: 'order_type', title: '订单类型' },
  { colKey: 'master_order_code', title: '来源订单' },
  { colKey: 'priority', title: '优先级' },  // ✨ 添加优先级列配置
  { colKey: 'customer.cus_name', title: '客户名称' },
  { colKey: 'mold_codes', title: '模具编号' },
  { colKey: 'plan_start_date', title: '计划开始' },
  { colKey: 'plan_end_date', title: '订单交期' },
  { colKey: 'total_amount', title: '总金额' },
  { colKey: 'status', title: '状态' },
  { colKey: 'deliver_time', title: '交付时间' },
  { colKey: 'create_time', title: '创建时间' },
];

// 默认显示的列
const DEFAULT_COLUMNS = ['order_code', 'order_type', 'master_order_code', 'priority', 'customer.cus_name', 'mold_codes', 'plan_start_date', 'plan_end_date', 'total_amount', 'status', 'deliver_time', 'create_time'];

// 列设置相关
const selectedColumns = ref<string[]>([...DEFAULT_COLUMNS]);

const updateColumns = (cols: string[]) => {
  selectedColumns.value = cols;
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
  // 跳转到模具订单详情页，传递已有数据避免重复请求
  router.push({
    path: `/mold/order/detail/${row.id}`,
    state: { __orderData: row },
  });
};

const handleViewMasterOrder = (row: any) => {
  // 如果有来源订单，跳转到综合订单详情页
  if (row.master_order_id) {
    router.push(`/order-management/detail/${row.master_order_id}`);
  }
};

const handleEdit = (row: any) => {
  // 跳转到模具订单编辑页
  router.push({ path: '/mold/order/form', query: { id: row.id } });
};

const handlePaymentManage = (row: any) => {
  // 跳转到款项管理页面
  router.push({ path: '/mold/order/payment', query: { orderId: row.id } });
};

const canDeliver = (row: any) => {
  // 只有状态为“已试模”的订单才能交付
  return row.status === 'TRIED';
};

const handleDeliver = (row: any) => {
  // TODO: 实现交付功能
};

const handleDelete = async (row: any) => {
  try {
    await deleteMoldOrder(row.id);
    MessagePlugin.success('删除成功');
    await fetchData();
  } catch (error: any) {
    console.error('删除订单失败:', error);
    MessagePlugin.error(error?.message || '删除失败');
  }
};

// 日期格式化
import { formatDate } from '@/utils/date';

// 状态标签样式（使用统一的颜色配置系统）
const getStatusStyle = (status: string) => {
  if (!status) return getDictColorRenderConfig('default');
  const statusItem = orderStatusList.value.find(s => s.value === status);
  if (!statusItem) return getDictColorRenderConfig('default');
  
  const color = statusItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);
  
  // 如果是 theme 模式，直接返回空样式，由 t-tag 的 theme 属性处理
  if (renderConfig.mode === 'theme') {
    return {};
  }
  
  // style 模式，使用自定义样式
  if (renderConfig.style) {
    return {
      background: renderConfig.style.background + ' !important',
      color: renderConfig.style.color + ' !important',
      borderColor: renderConfig.style.borderColor + ' !important',
    };
  }
  
  return getDictColorRenderConfig('default');
};

// 状态文本转换
const getStatusText = (status: string) => {
  if (!status) return '未知';
  const statusItem = orderStatusList.value.find(s => s.value === status);
  return statusItem ? statusItem.label : '未知';
};

// 获取状态主题（用于 t-tag 的 theme 属性）
const getStatusTheme = (status: string) => {
  if (!status) return 'default';
  const statusItem = orderStatusList.value.find(s => s.value === status);
  if (!statusItem) return 'default';
  
  const color = statusItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);
  
  return renderConfig.mode === 'theme' ? renderConfig.theme : 'default';
};

// 订单类型标签样式（使用统一的颜色配置系统）
const getOrderTypeStyle = (orderType: string) => {
  if (!orderType) return getDictColorRenderConfig('default');
  const typeItem = orderTypeList.value.find(t => t.value === orderType);
  if (!typeItem) return getDictColorRenderConfig('default');
  
  const color = typeItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);
  
  // 如果是 theme 模式，直接返回空样式，由 t-tag 的 theme 属性处理
  if (renderConfig.mode === 'theme') {
    return {};
  }
  
  // style 模式，使用自定义样式
  if (renderConfig.style) {
    return {
      background: renderConfig.style.background + ' !important',
      color: renderConfig.style.color + ' !important',
      borderColor: renderConfig.style.borderColor + ' !important',
    };
  }
  
  return getDictColorRenderConfig('default');
};

const getOrderTypeTheme = (orderType: string) => {
  if (!orderType) return 'default';
  const typeItem = orderTypeList.value.find(t => t.value === orderType);
  if (!typeItem) return 'default';
  
  const color = typeItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);
  
  return renderConfig.mode === 'theme' ? renderConfig.theme : 'default';
};

// 订单类型文本
const getOrderTypeText = (orderType: string) => {
  if (!orderType) return '未知';
  const typeItem = orderTypeList.value.find(t => t.value === orderType);
  return typeItem ? typeItem.label : '未知';
};

// ✨ 优先级标签样式（使用统一的颜色配置系统）
const getPriorityStyle = (priority: string) => {
  if (!priority) return getDictColorRenderConfig('default');
  const priorityItem = priorityList.value.find(p => p.value === priority);
  if (!priorityItem) return getDictColorRenderConfig('default');
  
  const color = priorityItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);
  
  // 如果是 theme 模式，直接返回空样式，由 t-tag 的 theme 属性处理
  if (renderConfig.mode === 'theme') {
    return {};
  }
  
  // style 模式，使用自定义样式
  if (renderConfig.style) {
    return {
      background: renderConfig.style.background + ' !important',
      color: renderConfig.style.color + ' !important',
      borderColor: renderConfig.style.borderColor + ' !important',
    };
  }
  
  return getDictColorRenderConfig('default');
};

// ✨ 获取优先级主题（用于 t-tag 的 theme 属性）
const getPriorityTheme = (priority: string) => {
  if (!priority) return 'default';
  const priorityItem = priorityList.value.find(p => p.value === priority);
  if (!priorityItem) return 'default';
  
  const color = priorityItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);
  
  return renderConfig.mode === 'theme' ? renderConfig.theme : 'default';
};

// ✨ 优先级文本转换
const getPriorityText = (priority: string) => {
  if (!priority) return '未知';
  const priorityItem = priorityList.value.find(p => p.value === priority);
  return priorityItem ? priorityItem.label : '未知';
};

onMounted(async () => {
  await fetchData();
});
const fetchData = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
    };
    if (props.createTimeStart) params.create_time_start = props.createTimeStart;
    if (props.createTimeEnd) params.create_time_end = props.createTimeEnd;
    if (props.dueSoon) params.due_soon = props.dueSoon;

    const res = await getMoldOrderList(params);
    
    // 后端已经返回下划线命名的字段，直接使用
    tableData.value = (res.list || []).map((item: any) => ({
      ...item,  // ✨ 直接展开所有字段
      customer: item.customer || {},  // ✨ 确保 customer 是对象
      mold_codes: item.mold_codes || [],  // ✨ 确保 mold_codes 是数组
    }));
    
    pagination.value.total = res.total || 0;
  } catch (error) {
    console.error('获取模具订单列表失败:', error);
    MessagePlugin.error('获取模具订单列表失败');
  } finally {
    loading.value = false;
  }
};

// 暴露方法给父组件
defineExpose({
  fetchData,
  updateColumns,
});
</script>

<style scoped>
.mold-order-list {
  /* 移除了 padding，由父容器控制 */
}

.mold-codes-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mold-code-item {
  line-height: 1.5;
}
</style>

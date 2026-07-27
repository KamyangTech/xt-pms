<template>
  <div class="mold-order-detail-page">
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <PageLoading />
    </div>

    <template v-else-if="orderDetail">
      <!-- 顶部操作栏 -->
      <div class="top-bar">
        <t-button theme="default" @click="goBack" variant="text">
          <template #icon><t-icon name="chevron-left" /></template>
          返回
        </t-button>
        <t-button theme="primary" @click="handleEdit">编辑</t-button>
      </div>

      <!-- 基本信息 -->
      <t-card :bordered="false" title="订单信息" class="detail-card">
        <t-descriptions :column="3" bordered>
          <t-descriptions-item label="订单编号">{{ orderDetail.order_code || '-' }}</t-descriptions-item>
          <t-descriptions-item label="订单类型">
            <t-tag shape="round" variant="outline" :theme="getOrderTypeTheme(orderDetail.order_type)">
              {{ getOrderTypeText(orderDetail.order_type) }}
            </t-tag>
          </t-descriptions-item>
          <t-descriptions-item label="优先级">
            <t-tag shape="round" :theme="getPriorityTheme(orderDetail.priority)" :style="getPriorityStyle(orderDetail.priority)">
              {{ getPriorityText(orderDetail.priority) }}
            </t-tag>
          </t-descriptions-item>
          <t-descriptions-item label="客户">{{ orderDetail.customer?.cus_name || '-' }}</t-descriptions-item>
          <t-descriptions-item label="计划开始">{{ formatDate(orderDetail.plan_start_date) }}</t-descriptions-item>
          <t-descriptions-item label="订单交期">{{ formatDate(orderDetail.plan_end_date) }}</t-descriptions-item>
          <t-descriptions-item label="订单状态">
            <t-tag shape="round" :theme="getStatusTheme(orderDetail.status)" :style="getStatusStyle(orderDetail.status)">
              {{ getStatusText(orderDetail.status) }}
            </t-tag>
          </t-descriptions-item>
          <t-descriptions-item label="总金额">
            <span class="amount-text">¥{{ orderDetail.total_amount?.toFixed(2) || '0.00' }}</span>
          </t-descriptions-item>
          <t-descriptions-item label="附加金额">
            <span class="amount-text">¥{{ orderDetail.additional_amount?.toFixed(2) || '0.00' }}</span>
          </t-descriptions-item>
          <t-descriptions-item label="交付时间">{{ formatDateTime(orderDetail.deliver_time) }}</t-descriptions-item>
          <t-descriptions-item label="创建时间">{{ formatDateTime(orderDetail.create_time) }}</t-descriptions-item>
          <t-descriptions-item label="备注" :span="3">
            <div style="white-space: pre-wrap;">{{ orderDetail.remark || '-' }}</div>
          </t-descriptions-item>
        </t-descriptions>
      </t-card>

      <!-- 模具明细（一单多款） -->
      <t-card title="模具明细" class="detail-card" :bordered="false" v-if="items.length > 0">
        <t-table
          :data="items"
          :columns="itemColumns"
          row-key="id"
          :stripe="false"
          bordered
        >
          <template #index="{ rowIndex }">
            {{ rowIndex + 1 }}
          </template>
          <template #mold_code="{ row }">
            <t-link @click="handleViewMoldDetail(row.mold_id)">
              {{ row.mold?.mold_code || row.mold_code || '-' }}
            </t-link>
          </template>
          <template #mold_name="{ row }">
            {{ row.mold?.mold_name || row.mold_name || '-' }}
          </template>
          <template #amount="{ row }">
            <span class="amount-text">¥{{ row.amount?.toFixed(2) || '0.00' }}</span>
          </template>
        </t-table>
      </t-card>

      <!-- 款项信息（新样式：与订单列表弹窗一致） -->
      <t-card :bordered="false" title="应收信息" class="detail-card">
        <t-descriptions :column="2" bordered>
          <t-descriptions-item label="客户名称">
            {{ orderDetail.customer?.cus_name || '-' }}
          </t-descriptions-item>
          <t-descriptions-item label="订单编号">
            <span style="white-space: nowrap; font-weight: 500;">{{ orderDetail.order_code || '-' }}</span>
          </t-descriptions-item>
          <t-descriptions-item label="订单金额">
            ¥{{ (orderDetail.total_amount || 0).toFixed(2) }}
          </t-descriptions-item>
          <t-descriptions-item label="订单状态">
            <t-tag shape="round" :theme="getStatusTheme(orderDetail.status)" :style="getStatusStyle(orderDetail.status)">
              {{ getStatusText(orderDetail.status) }}
            </t-tag>
          </t-descriptions-item>
        </t-descriptions>
      </t-card>

      <!-- 资金汇总 -->
      <t-card :bordered="false" class="detail-card">
        <div class="summary-stats">
          <div class="stat-item">
            <div class="stat-label">应收总额</div>
            <div class="stat-value">¥{{ formatAmount(receivableTotal) }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">已收金额</div>
            <div class="stat-value success">¥{{ formatAmount(receivableReceived) }}</div>
          </div>
          <div class="stat-item">
            <div class="stat-label">余额</div>
            <div class="stat-value warning">¥{{ formatAmount(receivableBalance) }}</div>
          </div>
        </div>
      </t-card>

      <!-- 应收明细 -->
      <t-card :bordered="false" title="应收明细" class="detail-card">
        <t-table
          :data="receivableList"
          :columns="receivableColumns"
          row-key="id"
          hover
          stripe
          :loading="receivableLoading"
        >
          <template #payment_type="{ row }">
            <t-tag size="small" :theme="getPaymentTypeTheme(row.payment_type)" :style="getPaymentTypeStyle(row.payment_type)">
              {{ getPaymentTypeLabel(row.payment_type) }}
            </t-tag>
          </template>
          <template #amount="{ row }">
            ¥{{ formatAmount(row.amount) }}
          </template>
          <template #received_amount="{ row }">
            <span class="text-success">¥{{ formatAmount(row.received_amount) }}</span>
          </template>
          <template #discount_amount="{ row }">
            <span v-if="row.discount_amount > 0" style="color: var(--td-warning-color);">-¥{{ formatAmount(row.discount_amount) }}</span>
            <span v-else>-</span>
          </template>
          <template #balance="{ row }">
            <span :class="row.balance > 0 ? 'text-warning' : 'text-success'">¥{{ formatAmount(row.balance) }}</span>
          </template>
          <template #due_date="{ row }">
            {{ row.due_date ? formatDate(row.due_date) : '-' }}
          </template>
          <template #status="{ row }">
            <t-tag size="small" :theme="getReceivableStatusTheme(row.status)" :style="getReceivableStatusStyle(row.status)">
              {{ getReceivableStatusText(row.status) }}
            </t-tag>
          </template>
          <template #op="{ row }">
            <t-link theme="primary" @click="handleViewReceivablePayments(row)">收款明细</t-link>
          </template>
        </t-table>
        <t-empty v-if="!receivableLoading && receivableList.length === 0" description="暂无应收款项记录" />
      </t-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import { getMoldOrderById } from '@/api/mold';
import { useDictRaw } from '@/hooks/useDict';
import { getReceivablesByOrder } from '@/api/finance/receivable';
import { getOrderPayments } from '@/api/mold/payment';
import PageLoading from '@/components/PageLoading/index.vue';

import { formatDate, formatDateTime } from '@/utils/date';
import { getDictColorRenderConfig } from '@/utils/dict-color';

defineOptions({
  name: 'MoldOrderDetail',
});

const route = useRoute();
const router = useRouter();

// 订单详情数据
const orderDetail = ref<any>(null);
const items = ref<any[]>([]);
const loading = ref(false);

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
  (items as any[]).map((item: any) => ({
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

// 应收数据
const receivableList = ref<any[]>([]);
const receivableLoading = ref(false);
// 款项字典（useDictRaw 自动加载 + 全局缓存）
const { items: paymentTypeOptions } = useDictRaw('payment-type');
const { items: receivableStatusOptions } = useDictRaw('receivable-status');

// 应收汇总计算
const receivableTotal = computed(() => {
  return receivableList.value.reduce((sum, r) => sum + (r.amount || 0), 0);
});
const receivableReceived = computed(() => {
  return receivableList.value.reduce((sum, r) => sum + (r.received_amount || 0), 0);
});
const receivableBalance = computed(() => {
  return receivableList.value.reduce((sum, r) => sum + (r.balance || 0), 0);
});

// 应收明细表格列
const receivableColumns = [
  { colKey: 'payment_type', title: '款项类型', width: 110 },
  { colKey: 'amount', title: '应收金额', width: 130 },
  { colKey: 'received_amount', title: '已收金额', width: 130 },
  { colKey: 'discount_amount', title: '优惠金额', width: 100 },
  { colKey: 'balance', title: '余额', width: 120 },
  { colKey: 'due_date', title: '到期日期', width: 140 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'remark', title: '备注', ellipsis: true, minWidth: 80 },
  { colKey: 'op', title: '操作', width: 110, fixed: 'right' as const },
];

// 模具明细表格列配置
const itemColumns = [
  { colKey: 'index', title: '序号', width: 80 },
  { colKey: 'mold_code', title: '模具编号', width: 150 },
  { colKey: 'mold_name', title: '模具名称', width: 150 },
  { colKey: 'amount', title: '金额', width: 120 },
  { colKey: 'remark', title: '备注', ellipsis: true },
];

// 状态样式
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

// 状态主题
const getStatusTheme = (status: string) => {
  if (!status) return 'default';
  const statusItem = orderStatusList.value.find(s => s.value === status);
  if (!statusItem) return 'default';

  const color = statusItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);

  return renderConfig.mode === 'theme' ? renderConfig.theme : 'default';
};

// 状态文本
const getStatusText = (status: string) => {
  if (!status) return '未知';
  const statusItem = orderStatusList.value.find(s => s.value === status);
  return statusItem ? statusItem.label : '未知';
};

// 订单类型主题
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
  const typeItem = orderTypeList.value.find(t => t.value === orderType);
  return typeItem ? typeItem.label : '未知';
};

// 优先级标签样式
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

// 优先级主题
const getPriorityTheme = (priority: string) => {
  if (!priority) return 'default';
  const priorityItem = priorityList.value.find(p => p.value === priority);
  if (!priorityItem) return 'default';

  const color = priorityItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);

  return renderConfig.mode === 'theme' ? renderConfig.theme : 'default';
};

// 优先级文本
const getPriorityText = (priority: string) => {
  const priorityItem = priorityList.value.find(p => p.value === priority);
  return priorityItem ? priorityItem.label : '未知';
};

// 格式化金额
function formatAmount(amount: number) {
  if (!amount && amount !== 0) return '0.00';
  return amount.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

// 款项类型标签
function getPaymentTypeLabel(type: string) {
  if (!type) return '-';
  const option = paymentTypeOptions.value.find((item: any) => item.dict_value === type);
  return option ? option.dict_label : type;
}
function getPaymentTypeTheme(type: string) {
  const option = paymentTypeOptions.value.find((item: any) => item.dict_value === type);
  if (option && option.color) {
    const renderConfig = getDictColorRenderConfig(option.color);
    if (renderConfig.mode === 'theme') {
      return renderConfig.theme || 'default';
    }
  }
  return 'default';
}
function getPaymentTypeStyle(type: string) {
  const option = paymentTypeOptions.value.find((item: any) => item.dict_value === type);
  if (option && option.color) {
    const renderConfig = getDictColorRenderConfig(option.color);
    if (renderConfig.mode === 'style' && renderConfig.style) {
      return renderConfig.style;
    }
    return {};
  }
  return { background: '#f0f0f0', color: '#333', border: '1px solid #dcdcdc' };
}

// 应收状态样式
function getReceivableStatusStyle(status: string) {
  if (status === 'locked') return {};
  const option = receivableStatusOptions.value.find((item: any) => item.dict_value === status) as any;
  if (option?.color) {
    const config = getDictColorRenderConfig(option.color);
    return config.mode === 'style' ? (config.style || {}) : {};
  }
  return {};
}
function getReceivableStatusTheme(status: string) {
  if (status === 'locked') return undefined;
  const option = receivableStatusOptions.value.find((item: any) => item.dict_value === status) as any;
  if (option?.color) {
    return getDictColorRenderConfig(option.color).theme;
  }
  const themeMap: Record<string, string> = { pending: 'default', partial: 'warning', paid: 'success', overdue: 'danger' };
  return themeMap[status?.toLowerCase()] as any;
}
function getReceivableStatusText(status: string) {
  const map: Record<string, string> = {
    pending: '待收款',
    partial: '部分收款',
    paid: '已收款',
    overdue: '逾期',
    locked: '已锁定',
  };
  return map[status?.toLowerCase()] || status || '未知';
}

const goBack = () => {
  router.back();
};

// 编辑
const handleEdit = () => {
  if (orderDetail.value?.id) {
    router.push(`/mold/order/form?id=${orderDetail.value.id}`);
  }
};

// 查看模具详情
const handleViewMoldDetail = (moldId: number) => {
  if (!moldId) {
    MessagePlugin.warning('该模具暂无档案记录');
    return;
  }
  router.push(`/mold/detail/${moldId}`);
};

// 查看收款明细
const handleViewReceivablePayments = (row: any) => {
  router.push(`/finance/receivable/payments/${row.id}`);
};

const loadOrderDetail = async () => {
  const id = route.params.id as string;
  if (!id) {
    MessagePlugin.error('订单ID不能为空');
    goBack();
    return;
  }

  // 从 history.state 获取基本订单信息，避免多余请求
  const passedData = (history.state as any)?.__orderData;
  if (passedData && String(passedData.id) === id) {
    orderDetail.value = passedData;
  }

  // 始终从 API 获取完整数据（含模具明细、预加载的关联对象）
  loading.value = true;
  try {
    const res = await getMoldOrderById(Number(id));
    orderDetail.value = res;
    items.value = res.items || [];
  } catch (error) {
    console.error('加载订单详情失败:', error);
    // 如果有从列表页传入的缓存数据，继续展示
    if (orderDetail.value?.id) {
      loading.value = false;
      return;
    }
    MessagePlugin.error('获取订单详情失败');
    goBack();
  } finally {
    loading.value = false;
  }
};

// 加载应收数据（替代旧款项统计）
const loadReceivableData = async () => {
  if (!orderDetail.value?.id) return;

  receivableLoading.value = true;
  try {
    const orderId = orderDetail.value.id;
    const masterOrderId = orderDetail.value.master_order_id;

    // 1. 应收账款表（mold_order）
    const arList = await getReceivablesByOrder(orderId, 'mold_order');
    let list = Array.isArray(arList) ? arList : (arList?.list || []);

    // 2. 回退：mold_order_payment 表
    if (list.length === 0) {
      const paymentList = await getOrderPayments(orderId);
      if (paymentList && Array.isArray(paymentList) && paymentList.length > 0) {
        list = paymentList.map((p: any) => ({
          id: p.id,
          order_id: orderId,
          payment_type: p.payment_type || '',
          source_type: p.source_type || 'mold',
          amount: p.contract_amount || p.should_amount || 0,
          received_amount: p.actual_amount || 0,
          should_amount: p.should_amount || 0,
          discount_amount: p.discount_amount || 0,
          balance: (p.should_amount || p.contract_amount || 0) - (p.actual_amount || 0),
          due_date: p.due_date || p.payment_date || '',
          status: p.status === 3 ? 'PAID' : (p.status === 2 ? 'PARTIAL' : 'PENDING'),
          remark: p.remark || '',
        }));
      }
    }

    // 3. 回退：综合订单的应收款
    if (list.length === 0 && masterOrderId) {
      try {
        const masterARList = await getReceivablesByOrder(masterOrderId, 'master_order');
        if (masterARList && Array.isArray(masterARList) && masterARList.length > 0) {
          list = masterARList.filter((ar: any) => ar.source_type === 'mold' || !ar.source_type);
        }
      } catch (e) {
        console.warn('[详情页] 查询综合订单应收款失败:', e);
      }
    }

    receivableList.value = list;
  } catch (error) {
    console.error('加载应收数据失败:', error);
  } finally {
    receivableLoading.value = false;
  }
};

// 加载款项相关字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

onMounted(async () => {
  // 字典已通过 useDictRaw 自动加载

  await loadOrderDetail();
  if (orderDetail.value?.id) {
    await loadReceivableData();
  }
});

watch(
  () => route.params.id,
  (newId) => {
    if (newId && route.name === 'MoldOrderDetail') {
      loadOrderDetail().then(() => {
        if (orderDetail.value?.id) {
          loadReceivableData();
        }
      });
    }
  },
);
</script>

<style lang="less" scoped>
.mold-order-detail-page {
  padding: 16px;
  background-color: var(--td-bg-color-page, #f5f5f5);
  min-height: 100vh;
}

.detail-card {
  margin-bottom: 16px;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.amount-text {
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.top-bar {
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.summary-stats {
  display: flex;
  gap: 16px;
  justify-content: space-around;
}
.stat-item {
  text-align: center;
  flex: 1;
}
.stat-label {
  font-size: 14px;
  color: var(--td-text-color-secondary);
  margin-bottom: 8px;
}
.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}
.stat-value.success {
  color: #00a870;
}
.stat-value.warning {
  color: #ed7b2f;
}

.t-descriptions {
  :deep(.t-descriptions__body) {
    background-color: transparent;
  }
}
</style>

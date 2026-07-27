<template>
  <div class="master-order-detail-page">
    <!-- 加载状态 -->
    <div v-if="loading" class="loading-container">
      <PageLoading />
    </div>

    <template v-else-if="orderDetail">
      <!-- 返回按钮 -->
      <div class="back-button-container">
        <t-button theme="default" @click="goBack" variant="text">
          <template #icon><t-icon name="chevron-left" /></template>
          返回
        </t-button>
      </div>

      <!-- 页面头部 -->
      <t-card :bordered="false">
        <div class="page-header">
          <div class="header-left">
            <h2 class="order-code">{{ orderDetail.order_code }}</h2>
            <t-space>
              <OrderTag kind="type" :theme="getOrderTypeTheme(orderDetail.order_type)">
                {{ getOrderTypeText(orderDetail.order_type) }}
              </OrderTag>
              <OrderTag kind="status" :theme="getStatusTheme(orderDetail.status)">
                {{ getStatusText(orderDetail.status) }}
              </OrderTag>
            </t-space>
          </div>
          <div class="header-right">
            <t-space :size="8">
              <t-button v-if="orderDetail.status !== 'CANCELLED' && orderDetail.status !== 'COMPLETED'" theme="primary" @click="handleEdit">编辑</t-button>
              <t-button v-if="orderDetail.status !== 'CANCELLED' && orderDetail.status !== 'COMPLETED'" theme="danger" @click="handleCancel">取消订单</t-button>
            </t-space>
          </div>
        </div>
      </t-card>

      <!-- 订单信息 -->
      <t-card title="订单信息" :bordered="false">
        <t-descriptions :column="3" bordered>
          <t-descriptions-item label="订单编号">{{ orderDetail.order_code || '-' }}</t-descriptions-item>
          <t-descriptions-item label="下单类型">
            <OrderTag kind="type" :theme="getOrderTypeTheme(orderDetail.order_type)">
              {{ getOrderTypeText(orderDetail.order_type) }}
            </OrderTag>
          </t-descriptions-item>
          <t-descriptions-item label="客户">{{ orderDetail.customer?.cus_name || '-' }}</t-descriptions-item>
          <t-descriptions-item label="订单状态">
            <OrderTag kind="status" :theme="getStatusTheme(orderDetail.status)">
              {{ getStatusText(orderDetail.status) }}
            </OrderTag>
          </t-descriptions-item>
          <t-descriptions-item label="总金额">
            <span class="amount-text">¥{{ orderDetail.total_amount?.toFixed(2) || '0.00' }}</span>
          </t-descriptions-item>
          <t-descriptions-item label="创建时间">{{ formatDateTime(orderDetail.create_time) }}</t-descriptions-item>
          <t-descriptions-item label="更新时间">{{ formatDateTime(orderDetail.update_time) }}</t-descriptions-item>
          <t-descriptions-item label="备注" :span="3">
            <div style="white-space: pre-wrap;">{{ orderDetail.remark || '-' }}</div>
          </t-descriptions-item>
        </t-descriptions>
      </t-card>

      <!-- 应收概览 -->
      <t-card title="应收概览" :bordered="false" v-if="receivableSummary.length">
        <t-table :data="receivableSummary" :columns="receivableColumns" row-key="payment_type" bordered>
          <template #payment_type="{ row }">
            <OrderTag kind="status" :theme="row.payment_type === 'deposit' ? 'success' : 'warning'">
              {{ row.payment_type === 'deposit' ? '首付款' : '尾款' }}
            </OrderTag>
          </template>
          <template #amount="{ row }">
            <span class="amount-text">¥{{ Number(row.amount || 0).toFixed(2) }}</span>
          </template>
          <template #received_amount="{ row }">
            <span class="amount-text success">¥{{ Number(row.received_amount || 0).toFixed(2) }}</span>
          </template>
          <template #balance="{ row }">
            <span class="amount-text" :class="{ warning: Number(row.balance || 0) > 0 }">¥{{ Number(row.balance || 0).toFixed(2) }}</span>
          </template>
          <template #status="{ row }">
            <t-space>
              <t-tag :theme="getReceivableStatusTheme(row.status)">{{ getReceivableStatusText(row.status) }}</t-tag>
              <t-tag v-if="row.is_locked" theme="warning">锁定</t-tag>
            </t-space>
          </template>
        </t-table>
      </t-card>

      <!-- 全流程进度条 -->
      <t-card title="订单进度" :bordered="false">
        <t-steps :current="getCurrentStep()" layout="horizontal" theme="dot">
          <t-step-item 
            title="模具制造" 
            :content="getStepContent(0)" 
          />
          <t-step-item 
            title="产品生产" 
            :content="getStepContent(1)" 
          />
          <t-step-item 
            title="产品入库" 
            :content="getStepContent(2)" 
          />
        </t-steps>
      </t-card>

      <!-- 子订单列表 -->
      <t-card title="子订单" :bordered="false" v-if="hasSubOrders">
        <t-table
          :data="subOrders"
          :columns="subOrderColumns"
          row-key="id"
          :stripe="false"
          bordered
        >
          <template #index="{ rowIndex }">
            {{ rowIndex + 1 }}
          </template>
          <template #order_type="{ row }">
            <OrderTag kind="type" :theme="getSubOrderTypeTheme(row.order_type)">
              {{ getSubOrderTypeText(row.order_type) }}
            </OrderTag>
          </template>
          <template #order_code="{ row }">
            <t-link @click="handleViewSubOrderDetail(row)">
              {{ row.order_code }}
            </t-link>
          </template>
          <template #status="{ row }">
            <OrderTag 
              kind="status"
              :style="{
                background: getDictColorHex(getSubOrderStatusColor(row.status, row.order_type)) + ' !important',
                color: (isLightColor(getSubOrderStatusColor(row.status, row.order_type)) ? 'rgba(0, 0, 0, 0.9)' : '#fff') + ' !important',
                borderColor: getDictColorHex(getSubOrderStatusColor(row.status, row.order_type)) + ' !important'
              }"
            >
              {{ getSubOrderStatusText(row.status, row.order_type) }}
            </OrderTag>
          </template>
          <template #total_amount="{ row }">
            <span class="amount-text">¥{{ row.total_amount?.toFixed(2) || '0.00' }}</span>
          </template>
          <template #plan_dates="{ row }">
            <div style="color: var(--td-text-color-secondary); font-size: 12px;">
              {{ row.plan_start_date ? formatDate(row.plan_start_date) : '-' }} 起
            </div>
            <div style="font-weight: 600; color: var(--td-text-color-primary);">
              至 {{ row.plan_end_date ? formatDate(row.plan_end_date) : '-' }}
            </div>
          </template>
          <template #operation="{ row }">
            <t-button size="small" theme="primary" variant="text" @click="handleViewSubOrderDetail(row)">
              查看详情
            </t-button>
          </template>
        </t-table>
      </t-card>

      <!-- 模具订单详情 -->
      <t-card 
        v-if="moldOrders.length > 0" 
        title="模具订单明细" 
        :bordered="false"
      >
        <t-collapse v-model="activeMoldCollapse">
          <t-collapse-panel
            v-for="moldOrder in moldOrders"
            :key="moldOrder.id"
            :header="`${moldOrder.order_code} - ${moldOrder.customer?.cus_name || ''}`"
          >
            <t-descriptions :column="3" bordered>
              <t-descriptions-item label="订单编号">{{ moldOrder.order_code }}</t-descriptions-item>
              <t-descriptions-item label="订单类型">
                <OrderTag kind="type" :theme="getOrderTypeTheme(moldOrder.order_type)">
                  {{ getOrderTypeText(moldOrder.order_type) }}
                </OrderTag>
              </t-descriptions-item>
              <t-descriptions-item label="客户">{{ moldOrder.customer?.cus_name || '-' }}</t-descriptions-item>
              <t-descriptions-item label="计划开始">{{ formatDate(moldOrder.plan_start_date) }}</t-descriptions-item>
              <t-descriptions-item label="订单交期">{{ formatDate(moldOrder.plan_end_date) }}</t-descriptions-item>
              <t-descriptions-item label="订单状态">
                <OrderTag
                  kind="status"
                  :style="{
                    background: getDictColorHex(getSubOrderStatusColor(moldOrder.status, 'MOLD')) + ' !important',
                    color: (isLightColor(getSubOrderStatusColor(moldOrder.status, 'MOLD')) ? 'rgba(0, 0, 0, 0.9)' : '#fff') + ' !important',
                    borderColor: getDictColorHex(getSubOrderStatusColor(moldOrder.status, 'MOLD')) + ' !important'
                  }"
                >
                  {{ getSubOrderStatusText(moldOrder.status, 'MOLD') }}
                </OrderTag>
              </t-descriptions-item>
              <t-descriptions-item label="总金额">
                <span class="amount-text">¥{{ moldOrder.total_amount?.toFixed(2) || '0.00' }}</span>
              </t-descriptions-item>
              <t-descriptions-item label="备注" :span="3">
                <div style="white-space: pre-wrap;">{{ moldOrder.remark || '-' }}</div>
              </t-descriptions-item>
            </t-descriptions>
            <div style="margin-top: 16px;">
              <t-button size="small" theme="primary" @click="handleViewMoldOrderDetail(moldOrder.id)">
                查看完整详情
              </t-button>
            </div>
          </t-collapse-panel>
        </t-collapse>
      </t-card>

      <!-- 生产订单详情 -->
      <t-card 
        v-if="prodOrders.length > 0" 
        title="生产订单明细" 
        :bordered="false"
      >
        <t-collapse v-model="activeProdCollapse">
          <t-collapse-panel
            v-for="prodOrder in prodOrders"
            :key="prodOrder.id"
            :header="`${prodOrder.order_code} - ${prodOrder.customer?.cus_name || ''}`"
          >
            <t-descriptions :column="3" bordered>
              <t-descriptions-item label="订单编号">{{ prodOrder.order_code }}</t-descriptions-item>
              <t-descriptions-item label="生产类型">
                <OrderTag kind="type" :theme="getProductionTypeTheme(prodOrder.production_type)">
                  {{ getProductionTypeText(prodOrder.production_type) }}
                </OrderTag>
              </t-descriptions-item>
              <t-descriptions-item label="客户">{{ prodOrder.customer?.cus_name || '-' }}</t-descriptions-item>
              <t-descriptions-item label="计划开始">{{ formatDate(prodOrder.plan_start_date) }}</t-descriptions-item>
              <t-descriptions-item label="计划结束">{{ formatDate(prodOrder.plan_end_date) }}</t-descriptions-item>
              <t-descriptions-item label="订单状态">
                <OrderTag
                  kind="status"
                  :style="{
                    background: getDictColorHex(getSubOrderStatusColor(prodOrder.status, 'PRODUCTION')) + ' !important',
                    color: (isLightColor(getSubOrderStatusColor(prodOrder.status, 'PRODUCTION')) ? 'rgba(0, 0, 0, 0.9)' : '#fff') + ' !important',
                    borderColor: getDictColorHex(getSubOrderStatusColor(prodOrder.status, 'PRODUCTION')) + ' !important'
                  }"
                >
                  {{ getSubOrderStatusText(prodOrder.status, 'PRODUCTION') }}
                </OrderTag>
              </t-descriptions-item>
              <t-descriptions-item label="依赖状态" v-if="prodOrder.dependency_status">
                <OrderTag kind="status" :theme="getDependencyStatusTheme(prodOrder.dependency_status)">
                  {{ getDependencyStatusText(prodOrder.dependency_status) }}
                </OrderTag>
              </t-descriptions-item>
              <t-descriptions-item label="计划数量">{{ prodOrder.plan_qty || 0 }}</t-descriptions-item>
              <t-descriptions-item label="完成数量">{{ prodOrder.completed_qty || 0 }}</t-descriptions-item>
              <t-descriptions-item label="备注" :span="3">
                <div style="white-space: pre-wrap;">{{ prodOrder.remark || '-' }}</div>
              </t-descriptions-item>
            </t-descriptions>
            <div style="margin-top: 16px;">
              <t-button size="small" theme="primary" @click="handleViewProdOrderDetail(prodOrder.id)">
                查看完整详情
              </t-button>
            </div>
          </t-collapse-panel>
        </t-collapse>
      </t-card>
    </template>
  </div>

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
    :confirm-btn="{ content: '确定取消', loading: cancelLoading }"
  >
    <p style="margin-bottom: 16px; color: var(--td-text-color-secondary);">
      取消后，未开工的子订单将一并取消；已开工的将被标记并禁止继续操作。已完成的生产记录保留（不可恢复）。
    </p>
    <t-form label-width="100px">
      <t-form-item label="取消原因" name="reason">
        <t-textarea v-model="cancelReason" placeholder="请填写取消原因" :rows="3" />
      </t-form-item>
    </t-form>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import { getMasterOrderDetail, editMasterOrder, cancelMasterOrder } from '@/api/order-management';
import { useDictRaw } from '@/hooks/useDict';
import { formatDate, formatDateTime } from '@/utils/date';
import { getDictColorHex, isLightColor } from '@/utils/dict-color';
import { OrderTag } from '@/components/OrderTag';
import PageLoading from '@/components/PageLoading/index.vue';

defineOptions({
  name: 'MasterOrderDetail',
});

const route = useRoute();
const router = useRouter();

// 订单详情数据
const orderDetail = ref<any>(null);
const moldOrders = ref<any[]>([]);
const prodOrders = ref<any[]>([]);
const receivableSummary = ref<any[]>([]);
const loading = ref(false);

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: orderStatusList } = useDictRaw('master-order-status');
const { items: productionOrderStatusList } = useDictRaw('production-order-status');
const { items: moldOrderStatusList } = useDictRaw('mold-order-status');
const { items: orderTypeList } = useDictRaw('master-order-type');
const { items: productionTypeList } = useDictRaw('production-type');
const { items: dependencyStatusList } = useDictRaw('dependency-status');

const receivableColumns = [
  { colKey: 'payment_type', title: '款项类型', width: 120 },
  { colKey: 'amount', title: '应收金额', width: 140 },
  { colKey: 'received_amount', title: '已收金额', width: 140 },
  { colKey: 'balance', title: '余额', width: 140 },
  { colKey: 'status', title: '状态', width: 160 },
];
// dependencyStatusList 已合并到上方 useDictRaw 声明中

// 折叠面板激活项
const activeMoldCollapse = ref<number[]>([]);
const activeProdCollapse = ref<number[]>([]);

// 子订单表格列配置
const subOrderColumns = [
  { colKey: 'index', title: '序号', width: 80 },
  { colKey: 'order_type', title: '订单类型', width: 120 },
  { colKey: 'order_code', title: '订单编号', width: 150 },
  { colKey: 'status', title: '状态', width: 120 },
  { colKey: 'total_amount', title: '金额', width: 120 },
  { colKey: 'plan_dates', title: '计划日期', width: 180 },
  { colKey: 'operation', title: '操作', width: 120, fixed: 'right' as const },
];

// 是否有子订单
const hasSubOrders = computed(() => {
  return moldOrders.value.length > 0 || prodOrders.value.length > 0;
});

// 子订单列表
const subOrders = computed(() => {
  const orders: any[] = [];
  moldOrders.value.forEach(order => {
    orders.push({ ...order, order_type: 'MOLD' });
  });
  prodOrders.value.forEach(order => {
    orders.push({ ...order, order_type: 'PRODUCTION' });
  });
  return orders;
});

const getOrderTypeText = (type: string) => {
  const map: Record<string, string> = {
    'COMBINED': '模具+产品',
    'MOLD_ONLY': '仅模具',
    'PROD_ONLY': '仅产品'
  };
  return map[type] || type;
};

const getOrderTypeTheme = (type: string) => {
  const map: Record<string, any> = {
    'COMBINED': 'primary',
    'MOLD_ONLY': 'warning',
    'PROD_ONLY': 'success'
  };
  return map[type] || 'default';
};

const getReceivableStatusTheme = (status: string) => {
  const map: Record<string, string> = {
    locked: 'warning',
    PENDING: 'default',
    PARTIAL: 'warning',
    PAID: 'success',
    OVERDUE: 'danger',
  };
  return map[status] || 'default';
};

const getReceivableStatusText = (status: string) => {
  const map: Record<string, string> = {
    locked: '锁定',
    PENDING: '待收款',
    PARTIAL: '部分收款',
    PAID: '已收款',
    OVERDUE: '逾期',
  };
  return map[status] || status;
};

const getStatusText = (status: string) => {
  const item = orderStatusList.value.find((item: any) => item.dict_value === status);
  return item ? item.dict_label : status;
};

const getStatusTheme = (status: string) => {
  const item = orderStatusList.value.find((item: any) => item.dict_value === status);
  return item?.color || 'default';
};

const getProductionTypeText = (type: string) => {
  const item = productionTypeList.value.find((item: any) => item.dict_value === type);
  return item ? item.dict_label : type;
};

const getProductionTypeTheme = (type: string) => {
  const item = productionTypeList.value.find((item: any) => item.dict_value === type);
  return item?.color || 'default';
};

const getDependencyStatusText = (status: string) => {
  const item = dependencyStatusList.value.find((item: any) => item.dict_value === status);
  return item ? item.dict_label : status;
};

const getDependencyStatusTheme = (status: string) => {
  const item = dependencyStatusList.value.find((item: any) => item.dict_value === status);
  return item?.color || 'default';
};

const getSubOrderTypeText = (type: string) => {
  const map: Record<string, string> = {
    'MOLD': '模具订单',
    'PRODUCTION': '生产订单'
  };
  return map[type] || type;
};

const getSubOrderTypeTheme = (type: string) => {
  const map: Record<string, any> = {
    'MOLD': 'warning',
    'PRODUCTION': 'success'
  };
  return map[type] || 'default';
};

const getSubOrderStatusColor = (status: string, orderType: string) => {
  const dictList = orderType === 'MOLD' ? moldOrderStatusList.value : productionOrderStatusList.value;
  const item = dictList.find((item: any) => item.dict_value === status);
  return item?.color || 'default';
};

const getSubOrderStatusText = (status: string, orderType: string) => {
  const dictList = orderType === 'MOLD' ? moldOrderStatusList.value : productionOrderStatusList.value;
  const item = dictList.find((item: any) => item.dict_value === status);
  return item ? item.dict_label : status;
};

const getSubOrderStatusTheme = (status: string, orderType: string) => {
  // 根据订单类型选择对应的字典
  const dictList = orderType === 'MOLD' ? moldOrderStatusList.value : productionOrderStatusList.value;
  const item = dictList.find((item: any) => item.dict_value === status);
  return item?.color || 'default';
};

// 获取当前进度步骤
const getCurrentStep = () => {
  if (prodOrders.value.length === 0 && moldOrders.value.length === 0) return 0;
  
  // 如果有生产订单，根据生产订单状态判断
  if (prodOrders.value.length > 0) {
    const prodOrder = prodOrders.value[0];
    if (prodOrder.dependency_status === 'WAITING_MOLD') return 0;
    if (prodOrder.status === 'PENDING_SCHEDULE') return 1;
    if (prodOrder.status === 'PRODUCING') return 1;
    if (prodOrder.status === 'COMPLETED') return 2;
  }
  
  // 如果只有模具订单，根据模具订单状态判断
  if (moldOrders.value.length > 0) {
    const moldOrder = moldOrders.value[0];
    if (moldOrder.status === 'COMPLETED') return 1;
    return 0;
  }
  
  return 0;
};

const getStepContent = (stepIndex: number) => {
  if (!orderDetail.value?.progress_timeline) {
    const defaultContents = ['模具制造', '产品生产', '产品入库'];
    return defaultContents[stepIndex] || '';
  }

  const timeline = orderDetail.value.progress_timeline;
  
  switch (stepIndex) {
    case 0: // 模具制造
      if (timeline.mold_manufacturing) {
        return timeline.mold_manufacturing;
      }
      return '待开始';
    case 1: // 产品生产
      if (timeline.product_production) {
        return timeline.product_production;
      }
      return '待排产';
    case 2: // 产品入库
      if (timeline.stock_in) {
        return timeline.stock_in;
      }
      return '待入库';
    default:
      return '';
  }
};

const goBack = () => {
  router.push('/order-management/list');
};

// 编辑对话框
const editDialogVisible = ref(false);
const editForm = ref({ quick_plan_end: '', remark: '' });

// 取消对话框
const cancelDialogVisible = ref(false);
const cancelReason = ref('');
const cancelLoading = ref(false);

// 编辑
const handleEdit = () => {
  editForm.value = {
    quick_plan_end: orderDetail.value.quick_plan_end || '',
    remark: orderDetail.value.remark || '',
  };
  editDialogVisible.value = true;
};

// 确认编辑
const handleEditConfirm = async () => {
  try {
    await editMasterOrder(orderDetail.value.id, {
      plan_end_date: editForm.value.quick_plan_end || undefined,
      remark: editForm.value.remark,
    });
    MessagePlugin.success('编辑成功');
    editDialogVisible.value = false;
    loadOrderDetail();
  } catch (error) {
    MessagePlugin.error('编辑失败');
  }
};

// 取消订单
const handleCancel = () => {
  cancelReason.value = '';
  cancelDialogVisible.value = true;
};

// 确认取消
const handleCancelConfirm = async () => {
  if (!cancelReason.value.trim()) {
    MessagePlugin.warning('请填写取消原因');
    return;
  }
  cancelLoading.value = true;
  try {
    await cancelMasterOrder(orderDetail.value.id, { reason: cancelReason.value });
    MessagePlugin.success('订单已取消');
    cancelDialogVisible.value = false;
    loadOrderDetail();
  } catch (error) {
    MessagePlugin.error('取消失败');
  } finally {
    cancelLoading.value = false;
  }
};

// 查看子订单详情
const handleViewSubOrderDetail = (row: any) => {
  if (row.order_type === 'MOLD' || row.mold_items) {
    handleViewMoldOrderDetail(row.id);
  } else {
    handleViewProdOrderDetail(row.id);
  }
};

// 查看模具订单详情
const handleViewMoldOrderDetail = (id: number) => {
  router.push(`/mold/order/detail/${id}`);
};

// 查看生产订单详情
const handleViewProdOrderDetail = (id: number) => {
  router.push(`/production/order/detail/${id}`);
};

const loadOrderDetail = async () => {
  const orderId = Number(route.params.id);
  if (!orderId) {
    MessagePlugin.error('订单 ID 不存在');
    return;
  }

  loading.value = true;
  try {
    const data = await getMasterOrderDetail(orderId, { skipErrorToast: true });
    orderDetail.value = data;
    receivableSummary.value = data.receivables || [];

    // 分离模具订单和生产订单
    if (data.mold_orders) {
      moldOrders.value = data.mold_orders;
      // 默认展开所有模具订单
      activeMoldCollapse.value = data.mold_orders.map((order: any) => order.id);
    }
    if (data.production_orders) {
      prodOrders.value = data.production_orders;
      // 默认展开所有生产订单
      activeProdCollapse.value = data.production_orders.map((order: any) => order.id);
    }
  } catch (error: any) {
    console.error('加载订单详情失败:', error);
    // 兼容两种错误形态：拦截器会改写为 { message, code }，但也可能是 axios 原始错误
    if (error?.response?.status === 404 || error?.code === 404) {
      MessagePlugin.error('订单不存在或已删除');
    } else {
      MessagePlugin.error('加载订单详情失败');
    }
  } finally {
    loading.value = false;
  }
};

// 加载字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

onMounted(() => {
  // 字典已通过 useDictRaw 自动加载
  loadOrderDetail();
});

// 监听路由参数变化，切换详情页时重新加载数据
watch(
  () => route.params.id,
  (newId) => {
    // 必须仍是当前页面，避免从 /order-management/detail/100 跳到 /mold/order/detail/125 时误触发
    if (newId && route.name === 'MasterOrderDetail') {
      loadOrderDetail();
    }
  },
);
</script>

<style scoped>
.master-order-detail-page {
  padding: var(--td-comp-paddingTB-l) var(--td-comp-paddingLR-l);
  background-color: var(--td-bg-color-page);
  min-height: 100vh;
}

.back-button-container {
  margin-bottom: var(--td-comp-margin-l);
}

/* 统一卡片间距与样式 */
.master-order-detail-page .t-card {
  margin-bottom: var(--td-comp-margin-l);
  background-color: var(--td-bg-color-container);
  border-radius: var(--td-radius-medium);
}

.master-order-detail-page .t-card:last-child {
  margin-bottom: 0;
}

.master-order-detail-page .t-descriptions {
  background-color: transparent;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--td-comp-paddingTB-l) 0;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: var(--td-comp-margin-m);
}

.header-right {
  display: flex;
  gap: var(--td-comp-margin-s);
}

.order-code {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.amount-text {
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}
</style>

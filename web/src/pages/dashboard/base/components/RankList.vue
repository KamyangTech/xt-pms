<template>
  <t-row :gutter="[16, 16]" class="row-container">
    <t-col :xs="12" :xl="6">
      <t-card title="进行中的模具订单" class="dashboard-rank-card" :bordered="false">
        <template #actions>
          <div style="display: flex; gap: 8px; align-items: center;">
            <span v-if="moldOrderTotal > 8" style="font-size: 14px; color: var(--td-text-color-secondary); font-weight: 500;">
              还有 {{ moldOrderTotal - 8 }} 个订单
            </span>
            <t-button theme="default" size="small" variant="outline" @click="goToMoldOrders">查看更多</t-button>
          </div>
        </template>
        <div class="table-fixed-height">
          <t-table
            :data="moldOrderData"
            :columns="MOLD_ORDER_COLUMNS"
            row-key="order_id"
            :hover="true"
          >
          <template #status="{ row }">
            <t-tag shape="round" :style="getStatusStyle(row.status)">
              {{ row.status_label }}
            </t-tag>
          </template>
        </t-table>
        </div>
      </t-card>
    </t-col>
    <t-col :xs="12" :xl="6">
      <t-card title="进行中的生产订单" class="dashboard-rank-card" :bordered="false">
        <template #actions>
          <div style="display: flex; gap: 8px; align-items: center;">
            <span v-if="prodOrderTotal > 8" style="font-size: 14px; color: var(--td-text-color-secondary); font-weight: 500;">
              还有 {{ prodOrderTotal - 8 }} 个订单
            </span>
            <t-button theme="default" size="small" variant="outline" @click="goToProdOrders">查看更多</t-button>
          </div>
        </template>
        <div class="table-fixed-height">
          <t-table
            :data="prodOrderData"
            :columns="PROD_ORDER_COLUMNS"
            row-key="order_id"
            :hover="true"
          >
            <template #status="{ row }">
              <t-tag shape="round" :style="getProdStatusStyle(row.status)">
                {{ row.status_label }}
              </t-tag>
            </template>
          </t-table>
        </div>
      </t-card>
    </t-col>
  </t-row>
</template>
<script setup lang="ts">
import { ref, onActivated, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import type { TdBaseTableProps } from 'tdesign-vue-next';

import { getInProgressMoldOrders, getInProgressProductionOrders } from '@/api/dashboard';

// 模具订单状态颜色映射
const MOLD_STATUS_COLORS: Record<string, { bg: string; text: string; border: string }> = {
  NOT_STARTED: { bg: '#f3f3f3', text: '#999', border: '#e0e0e0' },
  PROCESSING: { bg: '#e8f4fd', text: '#0052d9', border: '#b8d9f8' },
  PENDING_TEST: { bg: '#fff3e0', text: '#e37318', border: '#ffd9a8' },
  TESTED: { bg: '#e8f8e8', text: '#00a870', border: '#b8e8b8' },
  COMPLETED: { bg: '#e8f8e8', text: '#00a870', border: '#b8e8b8' },
  CANCELLED: { bg: '#f3f3f3', text: '#999', border: '#e0e0e0' },
  DELIVERED: { bg: '#e8f4fd', text: '#0052d9', border: '#b8d9f8' },
};

const getStatusStyle = (status: string) => {
  const color = MOLD_STATUS_COLORS[status] || MOLD_STATUS_COLORS.NOT_STARTED;
  return {
    background: color.bg,
    color: color.text,
    border: `1px solid ${color.border}`,
  };
};

// 生产订单状态颜色映射
const PROD_STATUS_COLORS: Record<string, { bg: string; text: string; border: string }> = {
  PENDING_SCHEDULE: { bg: '#f3f3f3', text: '#999', border: '#e0e0e0' },
  PRODUCING: { bg: '#e8f4fd', text: '#0052d9', border: '#b8d9f8' },
  STOCKED_IN: { bg: '#e8f8e8', text: '#00a870', border: '#b8e8b8' },
  SHIPPED: { bg: '#e8f8e8', text: '#00a870', border: '#b8e8b8' },
  COMPLETED: { bg: '#e8f8e8', text: '#00a870', border: '#b8e8b8' },
  CANCELLED: { bg: '#f3f3f3', text: '#999', border: '#e0e0e0' },
};

const getProdStatusStyle = (status: string) => {
  const color = PROD_STATUS_COLORS[status] || PROD_STATUS_COLORS.PENDING_SCHEDULE;
  return {
    background: color.bg,
    color: color.text,
    border: `1px solid ${color.border}`,
  };
};

// 进行中的模具订单数据
const moldOrderData = ref<any[]>([]);
const moldOrderTotal = ref(0);

// 列定义：进行中的模具订单
const MOLD_ORDER_COLUMNS: TdBaseTableProps['columns'] = [
  {
    align: 'left',
    ellipsis: true,
    colKey: 'mold_codes',
    title: '模具编号',
    minWidth: 150,
  },
  {
    align: 'left',
    ellipsis: true,
    colKey: 'customer_name',
    title: '客户名称',
    minWidth: 120,
  },
  {
    align: 'center',
    colKey: 'status',
    title: '状态',
    width: 100,
  },
  {
    align: 'center',
    colKey: 'plan_end_date',
    title: '到期日期',
    width: 120,
  },
];

// 进行中的生产订单数据
const prodOrderData = ref<any[]>([]);
const prodOrderTotal = ref(0);

// 列定义：进行中的生产订单
const PROD_ORDER_COLUMNS: TdBaseTableProps['columns'] = [
  {
    align: 'left',
    ellipsis: true,
    colKey: 'order_code',
    title: '订单编号',
    minWidth: 150,
  },
  {
    align: 'left',
    ellipsis: true,
    colKey: 'customer_name',
    title: '客户名称',
    minWidth: 120,
  },
  {
    align: 'center',
    colKey: 'status',
    title: '状态',
    width: 100,
  },
  {
    align: 'center',
    colKey: 'plan_end_date',
    title: '到期日期',
    width: 120,
  },
];

const loadMoldOrders = async () => {
  try {
    const res = await getInProgressMoldOrders();
    if (res) {
      moldOrderData.value = res.items.map(item => ({
        order_id: item.order_id,
        order_code: item.order_code,
        mold_codes: item.mold_codes,
        customer_name: item.customer_name,
        status: item.status,
        status_label: item.status_label,
        plan_end_date: item.plan_end_date,
      }));
      moldOrderTotal.value = res.total;
    }
  } catch (error) {
    console.error('加载进行中的模具订单失败:', error);
    moldOrderData.value = [];
  }
};

const loadProdOrders = async () => {
  try {
    const res = await getInProgressProductionOrders();
    if (res) {
      prodOrderData.value = res.items.map(item => ({
        order_id: item.order_id,
        order_code: item.order_code,
        customer_name: item.customer_name,
        status: item.status,
        status_label: item.status_label,
        plan_end_date: item.plan_end_date,
      }));
      prodOrderTotal.value = res.total;
    }
  } catch (error) {
    console.error('加载进行中的生产订单失败:', error);
    prodOrderData.value = [];
  }
};

const router = useRouter();

// 跳转到模具订单中心
const goToMoldOrders = () => {
  router.push('/mold/order');
};

// 跳转到生产订单列表
const goToProdOrders = () => {
  router.push('/production/order-list');
};

onMounted(() => {
  loadMoldOrders();
  loadProdOrders();
});

// ✨ KeepAlive 激活时刷新数据
onActivated(() => {
  loadMoldOrders();
  loadProdOrders();
});

</script>
<style lang="less" scoped>
.dashboard-rank-card {
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__header) {
    padding: 0;
  }

  :deep(.t-card__title) {
    font: var(--td-font-title-large);
    font-weight: 400;
  }

  :deep(.t-card__body) {
    padding: 0;
    margin-top: var(--td-comp-margin-xxl);
  }
}

.dashboard-rank__cell {
  display: inline-flex;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  color: white;
  font-size: 14px;
  background-color: var(--td-gray-color-5);
  align-items: center;
  justify-content: center;
  font-weight: 700;

  &--top {
    background: var(--td-brand-color);
  }
}

/* 列设置样式 */
.t-checkbox-group {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
</style>

<template>
  <div class="order-management-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="searchForm"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <t-tabs v-model="activeTab" @change="handleTabChange">
          <t-tab-panel value="master" label="综合订单" />
          <t-tab-panel value="mold" label="全部模具订单" />
          <t-tab-panel value="production" label="全部生产订单" />
        </t-tabs>
        <div class="toolbar-actions">
          <t-button theme="default" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
          <t-button theme="default" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button theme="primary" @click="handleCreate">
            <template #icon><add-icon /></template>
            新建订单
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <div v-if="activeTab === 'master'">
        <MasterOrderList
          ref="masterListRef"
          :order-code="searchForm.orderCode"
          :customer-name="searchForm.customerName"
          :create-time-start="dateRangeStart(searchForm.dateRange)"
          :create-time-end="dateRangeEnd(searchForm.dateRange)"
          :due-soon="searchForm.dueSoon"
        />
      </div>
      <div v-else-if="activeTab === 'mold'">
        <MoldOrderList
          ref="moldListRef"
          :create-time-start="dateRangeStart(searchForm.dateRange)"
          :create-time-end="dateRangeEnd(searchForm.dateRange)"
          :due-soon="searchForm.dueSoon"
        />
      </div>
      <div v-else>
        <ProductionOrderList
          ref="prodListRef"
          :search-keyword="searchForm.orderCode"
          :create-time-start="dateRangeStart(searchForm.dateRange)"
          :create-time-end="dateRangeEnd(searchForm.dateRange)"
          :due-soon="searchForm.dueSoon"
        />
      </div>
    </t-card>

    <!-- 列设置对话框 - 统一放在父组件 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      :title="currentColumnSettingTitle"
      :all-columns="currentAllColumns"
      :default-columns="currentDefaultColumns"
      :storage-key="currentStorageKey"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed } from 'vue';
import type { TabValue } from 'tdesign-vue-next';
import { AddIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import { useRouter } from 'vue-router';
import MasterOrderList from './components/MasterOrderList.vue';
import MoldOrderList from './components/MoldOrderList.vue';
import ProductionOrderList from './components/ProductionOrderList.vue';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';

const router = useRouter();
const activeTab = ref('master');
const searchForm = reactive({ orderCode: '', customerName: '', dueSoon: undefined as string | undefined, dateRange: [] as string[] });

// 到期预警选项
const dueSoonOptions = [
  { label: '1 天内到期', value: '1' },
  { label: '7 天内到期', value: '7' },
  { label: '30 天内到期', value: '30' },
];

// 筛选栏配置 - 根据当前Tab动态显示对应的日期框
const dateRangeLabels: Record<string, string> = {
  master: '创建日期',
  mold: '创建日期',
  production: '创建日期',
};
const filterConfig = computed<FilterItem[]>(() => [
  { type: 'input', field: 'orderCode', placeholder: '请输入单号' },
  { type: 'input', field: 'customerName', placeholder: '请输入客户' },
  { type: 'select', field: 'dueSoon', placeholder: '到期预警', options: dueSoonOptions, showMore: true },
  { type: 'dateRangePicker', field: 'dateRange', placeholder: dateRangeLabels[activeTab.value] || '日期范围', span: 8, showMore: true },
]);

const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);

// 综合订单列配置
const masterAllColumns: ColumnOption[] = [
  { colKey: 'order_code', title: '综合单号', pinned: true },
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
const masterDefaultColumns = ['order_code', 'customer.cus_name', 'order_type', 'mold_order_count', 'production_order_count', 'plan_start_date', 'plan_end_date', 'total_amount', 'status', 'create_time'];

// 模具订单列配置
const moldAllColumns: ColumnOption[] = [
  { colKey: 'order_code', title: '订单编号', pinned: true },
  { colKey: 'order_type', title: '订单类型' },
  { colKey: 'master_order_code', title: '来源订单' },
  { colKey: 'customer.cus_name', title: '客户名称' },
  { colKey: 'mold_codes', title: '模具编号' },
  { colKey: 'plan_start_date', title: '计划开始' },
  { colKey: 'plan_end_date', title: '订单交期' },
  { colKey: 'total_amount', title: '总金额' },
  { colKey: 'status', title: '状态' },
  { colKey: 'deliver_time', title: '交付时间' },
  { colKey: 'create_time', title: '创建时间' },
];
const moldDefaultColumns = ['order_code', 'order_type', 'master_order_code', 'customer.cus_name', 'mold_codes', 'plan_start_date', 'plan_end_date', 'total_amount', 'status', 'deliver_time', 'create_time'];

// 生产订单列配置
const productionAllColumns: ColumnOption[] = [
  { colKey: 'order_code', title: '订单号', pinned: true },
  { colKey: 'master_order_code', title: '来源订单' },
  { colKey: 'priority', title: '优先级' },
  { colKey: 'customer.cus_name', title: '客户名称' },
  { colKey: 'product_name', title: '产品名称' },
  { colKey: 'product_image', title: '产品图片' },
  { colKey: 'mold_code', title: '模具编号' },
  { colKey: 'production_type', title: '生产类型' },
  { colKey: 'progress', title: '进度' },
  { colKey: 'plan_qty', title: '计划数量' },
  { colKey: 'completed_qty', title: '完成数量' },
  { colKey: 'ok_qty', title: '合格数量' },
  { colKey: 'bad_qty', title: '不良数量' },
  { colKey: 'status', title: '状态' },
  { colKey: 'plan_start_date', title: '计划开始日期' },
  { colKey: 'plan_end_date', title: '订单交期' },
  { colKey: 'create_time', title: '创建时间' },
];
const productionDefaultColumns = [
  'order_code', 'master_order_code', 'priority', 'customer.cus_name', 'product_name', 'product_image', 'mold_code',
  'production_type', 'plan_qty', 'completed_qty', 'ok_qty', 'bad_qty', 'progress', 'status',
  'plan_start_date', 'plan_end_date', 'create_time',
];

// 根据当前Tab动态获取列配置
const currentColumnSettingTitle = computed(() => {
  const titles: Record<string, string> = {
    master: '综合订单 - 列设置',
    mold: '模具订单 - 列设置',
    production: '生产订单 - 列设置',
  };
  return titles[activeTab.value] || '列设置';
});

const currentAllColumns = computed(() => {
  const columnsMap: Record<string, ColumnOption[]> = {
    master: masterAllColumns,
    mold: moldAllColumns,
    production: productionAllColumns,
  };
  return columnsMap[activeTab.value] || [];
});

const currentDefaultColumns = computed(() => {
  const defaultColumnsMap: Record<string, string[]> = {
    master: masterDefaultColumns,
    mold: moldDefaultColumns,
    production: productionDefaultColumns,
  };
  return defaultColumnsMap[activeTab.value] || [];
});

const currentStorageKey = computed(() => {
  const keys: Record<string, string> = {
    master: 'master_order_table_columns',
    mold: 'mold_order_list_table_columns',
    production: 'production_order_table_columns',
  };
  return keys[activeTab.value] || 'order_table_columns';
});

// 打开列设置对话框
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 列配置变化时的回调
const handleColumnsChange = (cols: string[]) => {
  // 根据当前Tab通知对应的子组件更新列配置
  if (activeTab.value === 'master') {
    masterListRef.value?.updateColumns(cols);
  } else if (activeTab.value === 'mold') {
    moldListRef.value?.updateColumns(cols);
  } else if (activeTab.value === 'production') {
    prodListRef.value?.updateColumns(cols);
  }
};

const handleTabChange = (value: TabValue) => {
  activeTab.value = value as string;
};

const handleCreate = () => {
  // 根据当前tab自动选择订单类型
  const typeMap: Record<string, string> = {
    master: 'COMBINED',      // 综合订单 -> 综合下单
    mold: 'MOLD_ONLY',       // 模具订单 -> 仅制造模具
    production: 'PROD_ONLY', // 生产订单 -> 仅生产产品
  };
  const orderType = typeMap[activeTab.value] || 'COMBINED';
  router.push({ path: '/order-management/form', query: { type: orderType } });
};

const handleSearch = () => {
  // 触发当前激活的子组件搜索
  handleChildSearch();
};

const handleReset = () => {
  Object.assign(searchForm, { orderCode: '', customerName: '', dueSoon: undefined, dateRange: [] });
  handleSearch();
};

const handleRefresh = () => {
  // 触发当前激活的子组件刷新
  if (activeTab.value === 'master') {
    masterListRef.value?.fetchData();
  } else if (activeTab.value === 'mold') {
    moldListRef.value?.fetchData();
  } else {
    prodListRef.value?.fetchData();
  }
};

const handleChildSearch = () => {
  // 子组件通过 props 接收筛选参数后，通过 ref 触发主动刷新
  if (activeTab.value === 'master') {
    masterListRef.value?.fetchData();
  } else if (activeTab.value === 'mold') {
    moldListRef.value?.fetchData();
  } else {
    prodListRef.value?.fetchData();
  }
};

// 辅助函数：从日期范围数组中取开始/结束日期
const dateRangeStart = (range: any[]) => (range && range.length > 0 ? range[0] : '');
const dateRangeEnd = (range: any[]) => (range && range.length > 1 ? range[1] : '');

const masterListRef = ref<InstanceType<typeof MasterOrderList> | null>(null);
const moldListRef = ref<InstanceType<typeof MoldOrderList> | null>(null);
const prodListRef = ref<InstanceType<typeof ProductionOrderList> | null>(null);

// 暴露 fetchData 方法给父组件
defineExpose({
  fetchData: handleRefresh,
});
</script>

<style scoped>
.order-management-container {
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

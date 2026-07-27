<template>
  <div class="production-order-list-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="filterForm"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 工具栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">生产订单管理</span>
        <div class="toolbar-actions">
          <t-button theme="default" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
          <t-button theme="primary" @click="handleCreateOrder">
            <template #icon><add-icon /></template>
            新建订单
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <ProductionOrderList
        ref="productionOrderListRef"
        :show-search-bar="false"
        :show-create-button="false"
        :show-edit-button="true"
        :show-advanced-actions="true"
        :show-title="false"
        :search-keyword="searchKeyword"
        :filter-status="filterForm.status"
        :filter-production-type="filterForm.production_type"
        :filter-priority="filterForm.priority"
      />
    </t-card>
    
    <!-- 列设置对话框 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="production_order_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onActivated, onMounted, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { SettingIcon, AddIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import ProductionOrderList from '@/pages/order-management/components/ProductionOrderList.vue';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import { useDict, useDictRaw } from '@/hooks/useDict';

const router = useRouter();
const route = useRoute();

// 字典数据（useDict / useDictRaw 自动加载 + 全局缓存）
const { options: statusDictList } = useDict('production-order-status');
const { items: productionTypeDict } = useDictRaw('production-type');

// 搜索和筛选
const filterForm = reactive({
  searchKeyword: '',
  status: undefined as number | undefined,
  production_type: undefined as 1 | 2 | undefined,
  priority: undefined as 1 | 2 | 3 | 4 | undefined,
});

// 保持与子组件 searchKeyword props 兼容
const searchKeyword = computed(() => filterForm.searchKeyword);

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);

// 所有可选列配置（不包括操作列）
const allColumns: ColumnOption[] = [
  { colKey: 'order_code', title: '订单号', pinned: true },  // 固定列：不可隐藏/移动
  { colKey: 'master_order_code', title: '来源订单' },
  { colKey: 'priority', title: '优先级' },
  { colKey: 'customer', title: '客户名称' },
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
  { colKey: 'plan_end_date', title: '计划结束日期' },
];

// 默认显示的列
const DEFAULT_COLUMNS = [
  'order_code', 'master_order_code', 'priority', 'customer', 'product_name', 'product_image', 'mold_code',
  'production_type', 'plan_qty', 'completed_qty', 'ok_qty', 'bad_qty', 'progress', 'status',
  'plan_start_date', 'plan_end_date',
];

// 搜索
const productionOrderListRef = ref<InstanceType<typeof ProductionOrderList> | null>(null);

const handleSearch = () => {
  // 通过 props 传递搜索条件给子组件，子组件会自动监听并重新获取数据
  productionOrderListRef.value?.fetchData();
};

const handleReset = async () => {
  Object.assign(filterForm, {
    searchKeyword: '',
    status: undefined,
    production_type: undefined,
    priority: undefined,
  });
  // 等待 props 更新到子组件，避免子组件读到 SearchFilter 重置时写入的 '' 值
  await nextTick();
  handleSearch();
};

// 列设置
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 列配置变化时的回调
const handleColumnsChange = (columns: string[]) => {
  productionOrderListRef.value?.updateColumns(columns);
};

// 优先级选项
const priorityOptions = [
  { label: '普通', value: 1 },
  { label: '重要', value: 2 },
  { label: '紧急', value: 3 },
  { label: '非常紧急', value: 4 },
];

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'searchKeyword', placeholder: '搜索订单号、客户名称、产品名称' },
  { type: 'select', field: 'status', placeholder: '请选择状态', options: statusDictList },
  { type: 'select', field: 'production_type', placeholder: '请选择生产类型', options: productionTypeDict, labelKey: 'dict_label', valueKey: 'dict_value' },
  { type: 'select', field: 'priority', placeholder: '请选择优先级', options: priorityOptions },
];

const handleCreateOrder = () => {
  router.push('/production/order/create');
};

onMounted(async () => {
  // 支持从查询参数预填搜索关键词（如从其他页面跳转携带订单号）
  const orderCode = route.query.order_code as string | undefined;
  if (orderCode) {
    filterForm.searchKeyword = orderCode;
    // 等待子组件挂载后触发搜索
    setTimeout(() => {
      productionOrderListRef.value?.fetchData();
    }, 100);
  }
});

// ✨ KeepAlive 缓存下返回页面时自动刷新列表
onActivated(() => {
  productionOrderListRef.value?.fetchData();
});
</script>

<style scoped>
.production-order-list-container {
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
</style>

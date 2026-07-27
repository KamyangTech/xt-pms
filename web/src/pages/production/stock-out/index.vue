<template>
  <div class="stock-out-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="filterForm"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">发货记录</span>
        <div class="toolbar-actions">
          <t-button theme="primary" @click="handleCreate">
            <template #icon><add-icon /></template>
            新建送货单
          </t-button>
          <t-button 
            theme="default" 
            :disabled="selectedRowKeys.length === 0"
            @click="handleMergePrint"
          >
            <template #icon><print-icon /></template>
            合并送货单 ({{ selectedRowKeys.length }})
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
        row-key="id"
        :pagination="pagination"
        :loading="dataLoading"
        :selected-row-keys="selectedRowKeys"
        @page-change="handlePageChange"
        @select-change="handleSelectChange"
      >
        <template #mold_code="{ row }">
          <span>{{ row.mold?.mold_code || row.external_mold_code || '-' }}</span>
        </template>
        <template #order_code="{ row }">
          <span>{{ row.order?.order_code || '-' }}</span>
        </template>
        <template #customer_name="{ row }">
          <span>{{ row.customer_name || row.order?.customer?.cus_name || '-' }}</span>
        </template>
        <template #delivery_method="{ row }">
          <t-tag v-if="row.delivery_method" variant="light-outline">
            {{ getDeliveryMethodName(row.delivery_method) }}
          </t-tag>
          <span v-else>-</span>
        </template>
        <template #express_info="{ row }">
          <div v-if="row.express_company && row.express_no">
            <span>{{ getExpressCompanyName(row.express_company) }}</span>
            <span style="margin-left: 4px; color: var(--td-text-color-placeholder)">({{ row.express_no }})</span>
          </div>
          <span v-else>-</span>
        </template>
        <template #photos="{ row }">
          <t-image-viewer v-if="row.photos && getPhotos(row.photos).length > 0" :images="getPhotos(row.photos)">
            <template #trigger="{ open }">
              <t-tag variant="light-outline" @click="open">
                {{ getPhotos(row.photos).length }} 张照片
              </t-tag>
            </template>
          </t-image-viewer>
          <span v-else>-</span>
        </template>
        <template #delivery_date="{ row }">
          {{ formatDate(row.delivery_date) }}
        </template>
        <template #operation="{ row }">
          <t-space>
            <t-link theme="primary" @click="handleViewDetail(row)">详情</t-link>
            <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
            <t-link theme="primary" @click="handlePrint(row)">打印</t-link>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 编辑对话框 -->
    <edit-dialog ref="editDialogRef" @success="fetchStockOutList" />

    <!-- 列设置对话框 - 使用通用组件 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="stock_out_table_columns"
      @change="handleColumnsChange"
    />

    <!-- 打印组件 - 单个送货单 -->
    <PrintDeliveryNote
      v-if="printRecords.length === 1"
      :visible="showPrint"
      :records="printRecords"
      :config="printConfig"
      :customer-name="printCustomerName"
      :delivery-method-dict="deliveryMethodDict"
      :express-company-dict="expressCompanyDict"
    />
    
    <!-- 打印组件 - 合并送货单（使用统一模板） -->
    <PrintDeliveryNote
      v-if="printRecords.length > 1"
      :visible="showPrint"
      :records="printRecords"
      :config="printConfig"
      :customer-name="printCustomerName"
      :delivery-method-dict="deliveryMethodDict"
      :express-company-dict="expressCompanyDict"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onActivated, reactive, ref, computed, nextTick } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { useRouter } from 'vue-router';
import { AddIcon, PrintIcon, SettingIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';

import { getStockOutList, getStockOutListForPrint } from '@/api/production/stock';
import { getPrintNoticeConfig } from '@/api/reference';
import type { DeliveryNoteConfig } from '@/api/print-notice-config';
import { useDictRaw } from '@/hooks/useDict';
import EditDialog from './edit-dialog.vue';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import PrintDeliveryNote from '@/components/PrintDeliveryNote.vue';
import { formatDate } from '@/utils/date';

defineOptions({
  name: 'StockOutList',
});

const router = useRouter();

// 数据状态
const tableData = ref<any[]>([]);
const dataLoading = ref(false);
const selectedRowKeys = ref<number[]>([]);
const editDialogRef = ref<InstanceType<typeof EditDialog>>();

// 分页
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
});

// 搜索和筛选
const filterForm = reactive({
  customerName: '',
  deliveryMethod: '',
  dateRange: [] as string[],
});

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: deliveryMethodDict } = useDictRaw('delivery_method');
const { items: expressCompanyDict } = useDictRaw('express_company');

const showPrint = ref(false);
const printRecords = ref<any[]>([]);
const printCustomerName = ref('');
const printConfig = ref<DeliveryNoteConfig | null>(null);



// 表格列定义
const columns: PrimaryTableCol[] = [
  { title: '选择', colKey: 'row-select', type: 'multiple', width: 50, fixed: 'left' },
  { title: '出库单号', colKey: 'stock_code', width: 180, fixed: 'left' },
  { title: '关联订单', colKey: 'order_code', width: 150 },
  { title: '模具编号', colKey: 'mold_code', width: 150 },
  { title: '产品名称', colKey: 'product_name', width: 200 },
  { title: '接收客户', colKey: 'customer_name', width: 150 },
  { title: '出库数量', colKey: 'quantity', width: 100 },
  { title: '发货方式', colKey: 'delivery_method', width: 100 },
  { title: '快递信息', colKey: 'express_info', width: 180 },
  { title: '发货照片', colKey: 'photos', width: 120 },
  { title: '发货日期', colKey: 'delivery_date', width: 120 },
  { title: '操作', colKey: 'operation', width: 150, fixed: 'right' },
];

// 所有可选列配置（不包括操作列和出库单号）
const allColumns: ColumnOption[] = [
  { colKey: 'stock_code', title: '出库单号', pinned: true },  // ✨ 固定列：不可隐藏/移动
  { colKey: 'order_code', title: '关联订单' },
  { colKey: 'mold_code', title: '模具编号' },
  { colKey: 'product_name', title: '产品名称' },
  { colKey: 'customer_name', title: '接收客户' },
  { colKey: 'quantity', title: '出库数量' },
  { colKey: 'delivery_method', title: '发货方式' },
  { colKey: 'express_info', title: '快递信息' },
  { colKey: 'photos', title: '发货照片' },
  { colKey: 'delivery_date', title: '发货日期' },
];

// 默认显示的列（不包括出库单号，因为它固定显示）
const DEFAULT_COLUMNS = ['stock_code', 'order_code', 'mold_code', 'product_name', 'customer_name', 'quantity', 'delivery_method', 'express_info', 'photos', 'delivery_date'];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([...DEFAULT_COLUMNS]);

// 动态列配置（根据用户选择）
const displayColumns = computed(() => {
  const visibleColumns = columns.filter(col => 
    col.colKey === 'row-select' || 
    col.colKey === 'operation' || 
    col.colKey === 'stock_code' || 
    selectedColumns.value.includes(col.colKey)
  );
  
  // 按照 selectedColumns 的顺序排序
  return visibleColumns.sort((a, b) => {
    if (a.colKey === 'row-select') return -1;
    if (b.colKey === 'row-select') return 1;
    if (a.colKey === 'operation') return 1;
    if (b.colKey === 'operation') return -1;
    // 出库单号固定在第二位
    if (a.colKey === 'stock_code') return -1;
    if (b.colKey === 'stock_code') return 1;
    
    const indexA = selectedColumns.value.indexOf(a.colKey);
    const indexB = selectedColumns.value.indexOf(b.colKey);
    return indexA - indexB;
  });
});

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('stock_out_table_columns');
    if (saved) {
      const savedData = JSON.parse(saved);
      if (Array.isArray(savedData)) {
        selectedColumns.value = savedData;
      } else if (savedData.visible) {
        selectedColumns.value = savedData.visible;
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

const getDeliveryMethodName = (value: string) => {
  const item = deliveryMethodDict.value.find((d) => d.dict_value === value);
  return item ? item.dict_label : value;
};

const getExpressCompanyName = (value: string) => {
  const item = expressCompanyDict.value.find((d) => d.dict_value === value);
  return item ? item.dict_label : value;
};

const getPhotos = (photosStr: string) => {
  try {
    return photosStr ? JSON.parse(photosStr) : [];
  } catch {
    return [];
  }
};

const fetchStockOutList = async () => {
  dataLoading.value = true;
  try {
    const params: any = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
    };

    if (filterForm.customerName) {
      params.customer_name = filterForm.customerName;
    }
    if (filterForm.deliveryMethod) {
      params.delivery_method = filterForm.deliveryMethod;
    }
    if (filterForm.dateRange && filterForm.dateRange.length === 2) {
      params.start_date = filterForm.dateRange[0];
      params.end_date = filterForm.dateRange[1];
    }

    const res = await getStockOutList(params);
    tableData.value = res.list || [];
    pagination.value.total = res.total || 0;
  } catch (error) {
    console.error('获取送货单列表失败:', error);
    MessagePlugin.error('获取送货单列表失败');
  } finally {
    dataLoading.value = false;
  }
};

// 加载字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

// 搜索
const handleSearch = () => {
  pagination.value.current = 1;
  fetchStockOutList();
};

const handleReset = () => {
  Object.assign(filterForm, {
    customerName: '',
    deliveryMethod: '',
    dateRange: [],
  });
  handleSearch();
};

// 分页变化
const handlePageChange = (pageInfo: any) => {
  pagination.value.current = pageInfo.current;
  pagination.value.pageSize = pageInfo.pageSize;
  fetchStockOutList();
};

// 选择变化
const handleSelectChange = (keys: (string | number)[]) => {
  selectedRowKeys.value = keys.filter((key): key is number => typeof key === 'number');
};

const handleCreate = () => {
  router.push('/production/stock-out/create');
};

// 编辑
const handleEdit = (row: any) => {
  router.push(`/production/stock-out/edit/${row.id}`);
};

// 查看详情
const handleViewDetail = (row: any) => {
  router.push(`/production/stock-out/${row.id}`);
};

const handlePrint = async (row: any) => {
  if (!row) {
    MessagePlugin.warning('数据未加载');
    return;
  }
  
  try {
    const res = await getStockOutListForPrint([row.id]);
    
    printRecords.value = res.records;
    printCustomerName.value = res.customer_name || '';
    showPrint.value = true;
    
    // 延迟一下确保DOM渲染完成
    setTimeout(() => {
      // 监听打印完成或取消
      const afterPrint = () => {
        setTimeout(() => {
          showPrint.value = false;
        }, 500);
        // 移除监听器
        window.removeEventListener('afterprint', afterPrint);
      };
      
      // 兼容性处理：某些浏览器不支持 afterprint 事件
      if (window.matchMedia) {
        const mediaQueryList = window.matchMedia('print');
        const mediaQueryListener = (mql: MediaQueryList | MediaQueryListEvent) => {
          if (!mql.matches) {
            afterPrint();
            mediaQueryList.removeEventListener('change', mediaQueryListener);
          }
        };
        mediaQueryList.addEventListener('change', mediaQueryListener);
      } else {
        window.addEventListener('afterprint', afterPrint);
      }
      
      window.print();
    }, 100);
  } catch (error) {
    console.error('获取打印数据失败:', error);
    MessagePlugin.error('获取打印数据失败');
  }
};

const handleMergePrint = async () => {
  if (selectedRowKeys.value.length === 0) {
    MessagePlugin.warning('请选择要合并的送货单');
    return;
  }

  try {
    const res = await getStockOutListForPrint(selectedRowKeys.value);
    
    printRecords.value = res.records;
    printCustomerName.value = res.customer_name || '';
    showPrint.value = true;
    
    // 延迟打印，确保DOM渲染完成
    await nextTick();
    setTimeout(() => {
      window.print();
      showPrint.value = false;
    }, 100);
  } catch (error) {
    console.error('合并发货失败:', error);
    MessagePlugin.error('合并发货失败');
  }
};



// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'customerName', placeholder: '请输入客户名称' },
  { type: 'select', field: 'deliveryMethod', placeholder: '请选择发货方式', options: deliveryMethodDict, labelKey: 'dict_label', valueKey: 'dict_value' },
  { type: 'dateRangePicker', field: 'dateRange' },
];

const loadPrintConfig = async () => {
  try {
    const res = await getPrintNoticeConfig({ config_type: 'delivery_note' });
    if (res && res.basic_info && res.delivery_info && res.detail_columns) {
      printConfig.value = res;
    } else if (res && res.data) {
      printConfig.value = res.data;
    }
  } catch (error) {
    console.error('[StockOutList] 加载送货单打印配置失败:', error);
  }
};

onMounted(() => {
  // 字典已通过 useDictRaw 自动加载
  fetchStockOutList();
  loadColumnConfig();
  loadPrintConfig();
});

// keep-alive 激活时自动刷新列表（如从创建/编辑/详情页返回时）
onActivated(() => {
  fetchStockOutList();
});
</script>

<style lang="less" scoped>
.stock-out-container {
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

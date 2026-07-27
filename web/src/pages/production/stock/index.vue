<template>
  <div class="stock-container">
    <search-filter
      :filters="filterConfig"
      :model-value="filterForm"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">库存管理</span>
        <div class="toolbar-actions">
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
        row-key="row_key"
        :pagination="pagination"
        :loading="dataLoading"
        :header-affixed-top="headerAffixedTop"
        @page-change="handlePageChange"
      >
        <template #mold_code="{ row }">
          <span class="mold-code-text">{{ row.mold_code }}</span>
        </template>
        <template #mold_name="{ row }">
          <span class="mold-name-text">{{ row.mold_name }}</span>
        </template>
        <template #product_image_url="{ row }">
          <t-image-viewer v-if="row.product_image_url" :images="[row.product_image_url]">
            <template #trigger="{ open }">
              <img
                :src="getThumbnailUrl(row.product_image_url)"
                class="product-image-thumb"
                @click="open"
              />
            </template>
          </t-image-viewer>
          <div v-else class="product-image-placeholder">无图片</div>
        </template>
        <template #product_name="{ row }">
          <div class="product-names-wrapper">
            <span v-if="row.product_name" class="product-name-text" :title="row.product_name">
              {{ row.product_name }}
            </span>
            <span v-else class="product-name-empty">-</span>
          </div>
        </template>
        <template #current_stock="{ row }">
          <span class="stock-number" :class="{ 'stock-warning': row.current_stock === 0 }">
            {{ row.current_stock }}
          </span>
        </template>
        <template #last_stock_in_date="{ row }">
          {{ formatDate(row.last_stock_in_date) }}
        </template>
        <template #operation="{ row }">
          <t-space :size="8" break-line>
            <t-link theme="primary" @click="handleViewDetail(row)">详情</t-link>
            <t-link theme="success" @click="handleStockOut(row)">发货</t-link>
            <t-link theme="default" @click="handleCheckStock(row)">盘点</t-link>
            <t-link theme="danger" @click="handleDelete(row)">删除</t-link>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 列设置对话框 - 使用通用组件 -->
    <column-setting-dialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="production_stock_table_columns"
      @change="handleColumnsChange"
    />

    <!-- 盘点对话框 -->
    <t-dialog
      v-model:visible="checkStockDialogVisible"
      header="盘点库存"
      width="480px"
      @confirm="handleConfirmCheckStock"
      @cancel="handleCancelCheckStock"
     :close-on-overlay-click="false">
      <t-alert
        theme="info"
        message="输入实际盘点数量，系统将自动矫正库存并记录差异"
        style="margin-bottom: 16px"
      />

      <t-form
        ref="checkStockFormRef"
        :model="checkStockForm"
        :label-width="80"
        style="margin-top: 16px"
      >
        <t-form-item label="模具编号">
          <t-input
            :value="checkStockRow?.mold_code || checkStockRow?.external_mold_code"
            disabled
          />
        </t-form-item>

        <t-form-item label="产品名称">
          <t-input :value="checkStockRow?.product_name" disabled />
        </t-form-item>

        <t-form-item label="系统库存">
          <t-input :value="checkStockRow?.current_stock?.toString() || '0'" disabled />
        </t-form-item>

        <t-form-item label="盘点数量" required>
          <t-input-number
            v-model="checkStockForm.actualQuantity"
            :min="0"
            placeholder="请输入实际盘点数量"
            style="width: 100%"
          />
        </t-form-item>

        <t-form-item v-if="isCheckLoss" label="盘亏原因" required>
          <t-select v-model="checkStockForm.outReason" placeholder="请选择盘亏原因">
            <t-option
              v-for="item in outReasonDict"
              :key="item.dict_value"
              :label="item.dict_label"
              :value="item.dict_value"
            />
          </t-select>
        </t-form-item>

        <t-form-item label="备注">
          <t-textarea v-model="checkStockForm.remark" placeholder="请输入备注（可选）" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import { SettingIcon } from 'tdesign-icons-vue-next';
import type { FormInstanceFunctions, PrimaryTableCol } from 'tdesign-vue-next';
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next';
import { computed, h, onActivated, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { getThumbnailUrl } from '@/api/common/upload';
import { checkStock, deleteStock, getProductStockList } from '@/api/production/stock';
import type { ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import ColumnSettingDialog from '@/components/ColumnSettingDialog/index.vue';
import type { FilterItem } from '@/components/SearchFilter/index.vue';
import SearchFilter from '@/components/SearchFilter/index.vue';
import { prefix } from '@/config/global';
import { useDictRaw } from '@/hooks/useDict';
import { useSettingStore } from '@/store';

defineOptions({
  name: 'ProductionStock',
});

const router = useRouter();
const store = useSettingStore();

// 数据状态
const tableData = ref<any[]>([]);
const dataLoading = ref(false);

// 分页
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
});

// 搜索和筛选
const filterForm = reactive({
  moldCode: '',
  productName: '',
});

// 表格列定义
const COLUMNS: PrimaryTableCol[] = [
  { title: '模具编号', colKey: 'mold_code', width: 140, fixed: 'left' },
  { title: '产品图片', colKey: 'product_image_url', width: 100 },
  { title: '产品名称', colKey: 'product_name', width: 150 },
  { title: '订单客户', colKey: 'customer_name', width: 120 },
  { title: '总入库数量', colKey: 'total_stock_in', width: 100 },
  { title: '总出库数量', colKey: 'total_stock_out', width: 100 },
  { title: '当前库存', colKey: 'current_stock', width: 100 },
  { title: '最后入库日期', colKey: 'last_stock_in_date', width: 130 },
  { title: '操作', colKey: 'operation', width: 200, fixed: 'right' },
];

// 动态列配置（根据用户选择）
const displayColumns = computed(() => {
  // 根据 selectedColumns 的顺序来返回列配置
  const orderedColumns = selectedColumns.value
    .map((colKey) => COLUMNS.find((col) => col.colKey === colKey))
    .filter((col) => col !== undefined);

  // 将操作列添加到最后
  const opColumn = COLUMNS.find((col) => col.colKey === 'operation');
  if (opColumn) {
    orderedColumns.push(opColumn);
  }

  return orderedColumns;
});

// 所有可选列配置（不包括操作列）
const allColumns: ColumnOption[] = [
  { colKey: 'mold_code', title: '模具编号', pinned: true }, // ✨ 固定列：不可隐藏/移动
  { colKey: 'product_image_url', title: '产品图片' },
  { colKey: 'product_name', title: '产品名称' },
  { colKey: 'customer_name', title: '订单客户' },
  { colKey: 'total_stock_in', title: '总入库数量' },
  { colKey: 'total_stock_out', title: '总出库数量' },
  { colKey: 'current_stock', title: '当前库存' },
  { colKey: 'last_stock_in_date', title: '最后入库日期' },
];

// 默认显示的列
const DEFAULT_COLUMNS = [
  'mold_code',
  'product_image_url',
  'product_name',
  'customer_name',
  'total_stock_in',
  'total_stock_out',
  'current_stock',
  'last_stock_in_date',
];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

import { formatDate } from '@/utils/date';

const fetchProductStockList = async () => {
  dataLoading.value = true;
  try {
    const params: any = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
    };

    if (filterForm.moldCode) {
      params.mold_code = filterForm.moldCode;
    }
    if (filterForm.productName) {
      params.product_name = filterForm.productName;
    }

    const res = await getProductStockList(params);

    // 为每条数据生成唯一的 row_key
    tableData.value = (res.list || []).map((item: any) => ({
      ...item,
      row_key: item.mold_id || `ext-${item.external_mold_code}`,
    }));

    pagination.value.total = res.total || 0;
  } catch (error) {
    console.error('获取产品库存列表失败:', error);
    MessagePlugin.error('获取产品库存列表失败');
  } finally {
    dataLoading.value = false;
  }
};

// 搜索
const handleSearch = () => {
  pagination.value.current = 1;
  fetchProductStockList();
};

const handleReset = () => {
  Object.assign(filterForm, {
    moldCode: '',
    productName: '',
  });
  handleSearch();
};

// 分页变化
const handlePageChange = (pageInfo: any) => {
  pagination.value.current = pageInfo.current;
  pagination.value.pageSize = pageInfo.pageSize;
  fetchProductStockList();
};

// 查看产品出入库记录详情
const handleViewDetail = (row: any) => {
  if (row.mold_id) {
    // 自有模具：跳转到模具详情，并携带 product_id
    const params = new URLSearchParams();
    if (row.product_id) {
      params.set('product_id', String(row.product_id));
    }
    router.push(`/production/stock/detail/${row.mold_id}?${params.toString()}`);
  } else if (row.external_mold_code) {
    // 外来模具：跳转到外来模具详情（复用同一页面组件）
    const params = new URLSearchParams();
    if (row.product_id) {
      params.set('product_id', String(row.product_id));
    }
    router.push(
      `/production/stock/detail/ext/${encodeURIComponent(row.external_mold_code)}?${params.toString()}`,
    );
  }
};

// 出库操作 - 直接跳转到完整出库页面
const handleStockOut = async (row: any) => {
  if (!row.current_stock || row.current_stock <= 0) {
    MessagePlugin.warning('当前库存为0，无法出库');
    return;
  }

  // 直接跳转到完整出库页面，并携带当前产品信息
  const params = new URLSearchParams();
  if (row) {
    if (row.mold_id) {
      params.append('mold_id', row.mold_id);
    }
    if (row.external_mold_code) {
      params.append('external_mold_code', row.external_mold_code);
    }

    // 如果有 product_id 就传递
    if (row.product_id !== null && row.product_id !== undefined) {
      params.append('product_id', row.product_id);
    }

    // 传递模具编号和模具名称，用于自动填充显示
    if (row.mold_code) {
      params.append('mold_code', row.mold_code);
    }
    if (row.mold_name) {
      params.append('mold_name', row.mold_name);
    }
    if (row.product_name) {
      params.append('product_name', row.product_name);
    }

    // 传递客户名称，用于自动填充
    if (row.customer_name) {
      params.append('customer_name', row.customer_name);
    }

    // 传递当前库存，用于发货页面显示库存提示
    if (row.current_stock !== undefined && row.current_stock !== null) {
      params.append('current_stock', String(row.current_stock));
    }
  }
  params.append('from', 'stock');
  router.push(`/production/stock-out/create?${params.toString()}`);
};

const handleDelete = async (row: any) => {
  const deleteParams: any = {};
  if (row.mold_id) {
    deleteParams.mold_id = row.mold_id;
  } else if (row.external_mold_code) {
    deleteParams.external_mold_code = row.external_mold_code;
  }
  if (row.product_id) {
    deleteParams.product_id = row.product_id;
  }

  // 根据库存数量显示不同的确认对话框
  if (row.current_stock > 0) {
    // 库存不为0，显示警告并询问是否强制删除
    const confirmDialog = DialogPlugin.confirm({
      header: '警告：库存不为0',
      body: () => [
        h(
          'div',
          { style: 'margin-bottom: 8px' },
          `该记录当前库存为 ${row.current_stock}，删除后将直接清空库存数据。`,
        ),
        h('div', null, `模具编号：${row.mold_code || row.external_mold_code}`),
        h('div', null, `产品名称：${row.product_name || '-'}`),
        h('div', { style: 'margin-top: 8px; color: var(--td-error-color)' }, '确认要强制删除吗？'),
      ],
      confirmBtn: {
        content: '强制删除',
        theme: 'danger',
      },
      onConfirm: async () => {
        try {
          await deleteStock(deleteParams);
          MessagePlugin.success('删除成功');
          confirmDialog.hide();
          // 刷新列表
          fetchProductStockList();
        } catch (error: any) {
          console.error('删除失败:', error);
          const errorMsg = error.response?.data?.message || error.message || '删除失败';
          MessagePlugin.error(errorMsg);
          confirmDialog.hide();
        }
      },
      onCancel: () => {
        confirmDialog.hide();
      },
    });
  } else {
    // 库存为0，正常删除
    const confirmDialog = DialogPlugin.confirm({
      header: '确认删除',
      body: () => [
        h('div', null, `确认删除该库存记录吗？`),
        h('div', null, `模具编号：${row.mold_code || row.external_mold_code}`),
        h('div', null, `产品名称：${row.product_name || '-'}`),
        h('div', null, `当前库存：${row.current_stock}`),
      ],
      onConfirm: async () => {
        try {
          await deleteStock(deleteParams);
          MessagePlugin.success('删除成功');
          confirmDialog.hide();
          // 刷新列表
          fetchProductStockList();
        } catch (error: any) {
          console.error('删除失败:', error);
          const errorMsg = error.response?.data?.message || error.message || '删除失败';
          MessagePlugin.error(errorMsg);
          confirmDialog.hide();
        }
      },
      onCancel: () => {
        confirmDialog.hide();
      },
    });
  }
};

// 更多操作
const handleMoreAction = (value: any, row: any) => {
  switch (value) {
    case 'lost':
      MessagePlugin.info('遗失功能待实现');
      // TODO: 实现遗失逻辑
      break;
    case 'discarded':
      MessagePlugin.info('丢弃功能待实现');
      // TODO: 实现丢弃逻辑
      break;
  }
};

// ==================== 盘点相关 ====================
const checkStockDialogVisible = ref(false);
const checkStockFormRef = ref<FormInstanceFunctions>();
const checkStockRow = ref<any>(null);

const checkStockForm = ref({
  actualQuantity: undefined as number | undefined,
  outReason: '',
  remark: '',
});

// 出库原因字典
const { items: outReasonDict } = useDictRaw('out_reason');

// 是否盘亏（实际数量 < 系统库存）
const isCheckLoss = computed(() => {
  const actual = checkStockForm.value.actualQuantity;
  const current = checkStockRow.value?.current_stock;
  if (actual === undefined || actual === null || current === undefined || current === null)
    return false;
  return Number(actual) < Number(current);
});

// 打开盘点对话框
const handleCheckStock = (row: any) => {
  checkStockRow.value = row;
  checkStockForm.value = {
    actualQuantity: undefined,
    outReason: '',
    remark: '',
  };
  checkStockDialogVisible.value = true;
};

// 取消盘点
const handleCancelCheckStock = () => {
  checkStockDialogVisible.value = false;
  checkStockRow.value = null;
  checkStockFormRef.value?.reset();
};

// 确认盘点
const handleConfirmCheckStock = async () => {
  const errors: string[] = [];

  const actualQuantity = checkStockForm.value.actualQuantity;
  if (actualQuantity === undefined || actualQuantity === null) {
    errors.push('请输入盘点数量');
  } else if (Number(actualQuantity) < 0) {
    errors.push('盘点数量不能为负数');
  }

  // 盘亏时必须选择出库原因
  if (isCheckLoss.value && !checkStockForm.value.outReason) {
    errors.push('请选择盘亏原因');
  }

  if (errors.length > 0) {
    MessagePlugin.warning(errors.join('；'));
    return;
  }

  const row = checkStockRow.value;
  const diff = Number(actualQuantity) - (Number(row.current_stock) || 0);

  // 二次确认
  const bodyHtml = `<div style="line-height: 2;">
    <div>模具编号：${row.mold_code || row.external_mold_code}</div>
    <div>产品名称：${row.product_name || '-'}</div>
    <div>系统库存：${row.current_stock}</div>
    <div>盘点数量：${actualQuantity}</div>
    ${diff > 0 ? `<div style="color: var(--td-success-color); font-weight: 500;">盘盈：+${diff}</div>` : ''}
    ${diff < 0 ? `<div style="color: var(--td-error-color); font-weight: 500;">盘亏：${diff}</div><div>盘亏原因：${outReasonDict.value.find((item: any) => item.dict_value === checkStockForm.value.outReason)?.dict_label || '-'}</div>` : ''}
    ${diff === 0 ? '<div>库存数量无变化</div>' : ''}
    <div style="margin-top: 8px; font-weight: 500;">确认提交盘点吗？</div>
  </div>`;

  const confirmDialog = DialogPlugin.confirm({
    header: '确认盘点',
    body: () => h('div', { innerHTML: bodyHtml }),
    onConfirm: async () => {
      try {
        const params: any = {
          warehouse_id: row.warehouse_id || 1,
          actual_quantity: Number(actualQuantity),
          remark: checkStockForm.value.remark,
        };

        if (row.mold_id) {
          params.mold_id = row.mold_id;
        } else if (row.external_mold_code) {
          params.external_mold_code = row.external_mold_code;
        }

        if (row.product_id) {
          params.product_id = row.product_id;
        }

        // 盘亏时传递出库原因
        if (isCheckLoss.value) {
          params.out_reason = checkStockForm.value.outReason;
        }

        await checkStock(params);

        MessagePlugin.success('盘点成功');
        checkStockDialogVisible.value = false;
        confirmDialog.hide();

        fetchProductStockList();
      } catch (error: any) {
        console.error('盘点失败:', error);
        const errorMsg = error.response?.data?.message || error.message || '盘点失败';
        MessagePlugin.error(errorMsg);
      }
    },
    onCancel: () => {
      confirmDialog.hide();
    },
  });
};

const headerAffixedTop = computed(
  () =>
    ({
      offsetTop: store.isUseTabsRouter ? 48 : 0,
      container: `.${prefix}-layout`,
    }) as any,
);

// 打开列设置对话框
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 列配置变化时的回调
const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
};

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'moldCode', placeholder: '请输入模具编号' },
  { type: 'input', field: 'productName', placeholder: '请输入产品名称' },
];

onMounted(() => {
  fetchProductStockList();
  // 出库原因字典已通过 useDictRaw 自动加载
  try {
    const saved = localStorage.getItem('production_stock_table_columns');
    if (saved) {
      const savedColumns = JSON.parse(saved);
      if (Array.isArray(savedColumns)) {
        const validColumns = savedColumns.filter((colKey: string) =>
          allColumns.some((col) => col.colKey === colKey),
        );
        selectedColumns.value = validColumns.length > 0 ? validColumns : [...DEFAULT_COLUMNS];
      } else {
        // 如果不是数组，使用默认配置
        selectedColumns.value = [...DEFAULT_COLUMNS];
      }
    } else {
      selectedColumns.value = [...DEFAULT_COLUMNS];
    }
  } catch (error) {
    console.error('加载列配置失败:', error);
    selectedColumns.value = [...DEFAULT_COLUMNS];
  }
});

// KeepAlive 缓存恢复时自动刷新列表，确保从报工等页面返回后数据最新
onActivated(() => {
  fetchProductStockList();
});
</script>
<style lang="less" scoped>
.stock-container {
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

/* 列设置样式 */
.column-setting-content {
  padding: 16px 0;
}

.draggable-column-list {
  max-height: 400px;
  overflow-y: auto;
}

.column-item {
  padding: 8px 0;
  border-bottom: 1px solid var(--td-border-level-1-color, #e7e7e7);
  display: flex;
  align-items: center;
  justify-content: space-between;

  &:last-child {
    border-bottom: none;
  }
}

.draggable-item {
  cursor: move;
  transition: background-color 0.2s;

  &:hover {
    background-color: var(--td-bg-color-container-hover, #f0f0f0);
  }
}

.drag-handle {
  color: var(--td-text-color-placeholder, #999);
  cursor: grab;
  flex-shrink: 0;
  font-size: 20px;

  &:active {
    cursor: grabbing;
  }
}

.column-setting-footer {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.mold-code-text {
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.mold-name-text {
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.product-names-wrapper {
  max-width: 300px;
}

.product-name-text {
  font-size: 13px;
  color: var(--td-text-color-primary);
  line-height: 1.5;
  word-break: break-all;
  cursor: pointer;

  &:hover {
    color: var(--td-brand-color);
  }
}

.product-name-empty {
  color: var(--td-text-color-placeholder);
}

.stock-number {
  font-weight: 600;
  color: var(--td-brand-color);
  font-size: 16px;

  &.stock-warning {
    color: var(--td-error-color);
  }
}

/* 产品图片样式 - 统一使用70x70圆角 */
.product-image-thumb {
  width: 70px;
  height: 70px;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
  transition: opacity 0.2s;
}

.product-image-thumb:hover {
  opacity: 0.8;
}

.product-image-placeholder {
  width: 70px;
  height: 70px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  border-radius: 4px;
  color: #999;
  font-size: 12px;
}
</style>

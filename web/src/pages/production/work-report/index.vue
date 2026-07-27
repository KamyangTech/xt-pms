<template>
  <div class="work-report-container">
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
          <t-tab-panel value="all" label="全部" />
          <t-tab-panel value="IN_HOUSE" label="本厂" />
          <t-tab-panel value="OUTSOURCED" label="外发" />
        </t-tabs>
        <div class="toolbar-actions">
          <t-button theme="default" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
        </div>
      </div>
    </t-card>

    <template v-if="orderList.length > 0 && !dataLoading">
      <div class="work-report-items">
        <t-row :gutter="[16, 16]">
          <t-col
            v-for="order in orderList"
            :key="order.id"
            :lg="4"
            :xs="6"
            :xl="3"
          >
            <work-report-card
              class="work-report-item"
              :order="order"
              :production-type-dict="productionTypeDict"
              @report-work="handleReportWork"
              @view-detail="handleViewDetail"
              @start-production="handleStartProduction"
              @stock-in="handleStockIn"
              @complete-production="handleCompleteProduction"
            />
          </t-col>
        </t-row>
      </div>
      <div class="work-report-pagination">
        <t-pagination
          v-model="pagination.current"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-size-options="[12, 24, 36]"
          @page-size-change="onPageSizeChange"
          @current-change="onCurrentChange"
        />
      </div>
    </template>

    <div v-else-if="dataLoading" class="work-report-loading">
      <PageLoading />
    </div>

    <t-empty v-else-if="!dataLoading" description="暂无生产报工数据" />

    <!-- 报工对话框 -->
    <t-dialog
      v-model:visible="reportDialogVisible"
      :header="`报工 - ${currentOrder?.orderNo}`"
      width="600px"
      @confirm="handleConfirmReport"
      @cancel="reportDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form :model="reportForm" :label-width="100" style="margin-top: 16px">
        <t-form-item label="报工日期">
          <t-date-picker v-model="reportForm.reportDate" format="YYYY-MM-DD" />
        </t-form-item>
        <t-form-item label="班次">
          <t-select v-model="reportForm.shift" placeholder="请选择班次">
            <t-option
              v-for="item in shiftDict"
              :key="item.dict_value"
              :value="item.dict_value"
              :label="item.dict_label"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="良品数量">
          <t-input-number v-model="reportForm.okQty" placeholder="请输入良品数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="不良品数量">
          <t-input-number v-model="reportForm.badQty" placeholder="请输入不良品数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="reportForm.remark" placeholder="请输入备注" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 入库对话框 -->
    <t-dialog
      v-model:visible="stockInDialogVisible"
      :header="`入库 - ${currentOrder?.orderNo}`"
      width="600px"
      @confirm="handleConfirmStockIn"
      @cancel="stockInDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form :model="stockInForm" :label-width="100" style="margin-top: 16px">
        <t-form-item label="仓库">
          <t-select v-model="stockInForm.warehouseId" placeholder="请选择仓库">
            <t-option
              v-for="warehouse in warehouseList"
              :key="warehouse.id"
              :label="warehouse.wh_name"
              :value="warehouse.id"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="产品">
          <!-- ✨ 直接显示当前产品信息，不需要选择 -->
          <t-input 
            :value="currentOrder?.productName" 
            disabled 
            placeholder="产品"
          />
        </t-form-item>
        <t-form-item label="入库数量">
          <t-input-number v-model="stockInForm.quantity" :min="1" placeholder="请输入入库数量" />
        </t-form-item>
        <t-form-item label="批次号">
          <t-input v-model="stockInForm.batchNo" placeholder="自动生成" disabled />
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="stockInForm.remark" placeholder="请输入备注（可选）" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 确认对话框 -->
    <t-dialog
      v-model:visible="confirmDialogVisible"
      :header="confirmDialogTitle"
      width="500px"
      @confirm="handleConfirmAction"
      @cancel="confirmDialogVisible = false"
     :close-on-overlay-click="false">
      <p style="margin-top: 16px">{{ confirmDialogContent }}</p>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { RefreshIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onActivated, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { TabValue } from 'tdesign-vue-next';

import WorkReportCard from '@/components/work-report-card/index.vue';
import type { WorkReportOrderType } from '@/components/work-report-card/index.vue';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import { getProductionOrders, submitReport, startProduction, completeProduction, completeOrderItem } from '@/api/production/order';
import { getWarehouseList } from '@/api/reference';
import { createStockDirectly, stockInOrder } from '@/api/production/stock';
import { useDictRaw } from '@/hooks/useDict';
import type { ProductionOrder } from '@/types/production';
import { getDictColorRenderConfig } from '@/utils/dict-color';
import PageLoading from '@/components/PageLoading/index.vue';

defineOptions({
  name: 'WorkReport',
});

// 路由实例
const router = useRouter();

// 搜索表单
const searchForm = reactive({ orderCode: '', productName: '', status: '' });

// 生产类型筛选
const activeTab = ref('all');

// 分页
const dataLoading = ref(true);
const pagination = ref({ current: 1, pageSize: 12, total: 0 });

// 状态字典数据（useDictRaw 自动加载 + 全局缓存，含主题映射）
const { items: statusDictList } = useDictRaw('production-order-status', (items) => {
  const dictList = (items as any[]).filter((item: any) => item.status === 1);
  return dictList.map((item: any) => {
    const color = item.color || 'default';
    const renderConfig = getDictColorRenderConfig(color);
    return {
      value: item.dict_value,
      label: item.dict_label,
      color: color,
      theme: renderConfig.mode === 'theme' ? renderConfig.theme : 'default',
      themeStyle: renderConfig.mode === 'style' ? {
        background: renderConfig.style!.background + ' !important',
        color: renderConfig.style!.color + ' !important',
        borderColor: renderConfig.style!.borderColor + ' !important',
      } : undefined,
      semanticTag: item.semantic_tag || '',
    };
  });
});

// 生产类型字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: productionTypeDict } = useDictRaw('production-type', (items) =>
  (items as any[]).filter((item: any) => item.status === 1),
);

// 班次字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: shiftDict } = useDictRaw('shift-type', (items) =>
  (items as any[]).filter((item: any) => item.status === 1),
);

// 状态筛选选项（报工页仅展示可报工/可开始生产的状态）
const statusFilterOptions = [
  { label: '待排产', value: 'PENDING_SCHEDULE' },
  { label: '生产中', value: 'PRODUCING' },
];

// 筛选栏配置
const filterConfig = computed<FilterItem[]>(() => [
  { type: 'input', field: 'orderCode', placeholder: '请输入单号' },
  { type: 'input', field: 'productName', placeholder: '请输入产品名称' },
  { type: 'select', field: 'status', placeholder: '请选择状态', options: statusFilterOptions },
]);

// 订单列表
const orderList = ref<WorkReportOrderType[]>([]);

// 对话框状态
const reportDialogVisible = ref(false);
const confirmDialogVisible = ref(false);
const currentOrder = ref<WorkReportOrderType | null>(null);

// 报工表单
const reportForm = ref({
  reportDate: new Date().toISOString().split('T')[0],
  shift: undefined as string | undefined, // 班次字典值（语义化字符串）
  okQty: undefined as number | undefined,
  badQty: undefined as number | undefined,
  remark: '',
});

// 确认对话框
const confirmDialogTitle = ref('');
const confirmDialogContent = ref('');
let confirmActionCallback: (() => void) | null = null;

// 入库对话框
const stockInDialogVisible = ref(false);
const stockInForm = ref({
  warehouseId: undefined as number | undefined,
  moldId: undefined as number | undefined,
  externalMoldCode: undefined as string | undefined,
  productId: undefined as number | undefined,
  quantity: 0,
  batchNo: '',
  remark: '',
});

// 仓库列表
const warehouseList = ref<any[]>([]);

// 当前订单的模具列表（明细项）
const orderMolds = ref<any[]>([]);

// 将字典颜色映射到 TDesign 主题
const getTDesignTheme = (color: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' => {
  const renderConfig = getDictColorRenderConfig(color);
  if (renderConfig.mode === 'theme') {
    return renderConfig.theme!;
  }
  return 'default';
};

const fetchData = async () => {
  dataLoading.value = true;
  try {
    const params: any = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
      order_code: searchForm.orderCode || undefined,
      product_name: searchForm.productName || undefined,
      item_status: searchForm.status || undefined,
    };
    // 非"全部"时传递生产类型筛选
    if (activeTab.value !== 'all') {
      params.production_type = activeTab.value;
    }

    const res = await getProductionOrders(params);

    const expandedOrders: WorkReportOrderType[] = [];
    
    (res.list || []).forEach((order: ProductionOrder) => {
      // 如果有明细项，为每个明细项创建一个卡片
      if (order.items && order.items.length > 0) {
        order.items.forEach((item: any, index: number) => {
          const moldCode = item.mold?.mold_code || item.external_mold_code || '-';
          
          const productImageUrl = item.product_image_url || '';
          
          // ✨ 获取该明细项的依赖状态
          const dependencyStatus = item.dependency_status || 'NONE';
          
          expandedOrders.push({
            id: order.id,
            itemId: item.id, // ✨ 新增：明细项ID
            orderNo: order.order_code,
            productName: item.product_name || '-',
            productImageUrl: productImageUrl || undefined,
            customerName: order.customer_name || order.customer?.cus_name,
            moldCode: moldCode,
            planQuantity: item.plan_qty || order.plan_qty,
            completedQuantity: item.completed_qty || 0,
            // ✨ 显示状态以订单主状态为准：避免 item.status 默认值（PRODUCING）
            //   与 order.status（PENDING_SCHEDULE）不一致时导致卡片和详情页状态对不上
            //   只有订单开始生产后，明细项的独立状态（STOCKED_IN 等）才有意义
            status: order.status === 'PENDING_SCHEDULE' ? 'PENDING_SCHEDULE' : (item.status || order.status),
            statusInfo: getStatusInfo(order.status === 'PENDING_SCHEDULE' ? 'PENDING_SCHEDULE' : (item.status || order.status)),
            deadline: order.plan_end_date ? new Date(order.plan_end_date).toISOString().split('T')[0] : '-',
            productionType: order.production_type,
            dependencyStatus: dependencyStatus,
            // ✨ 新增：保存产品ID和模具ID，用于入库
            productId: item.product_id,
            moldId: item.mold_id,
            externalMoldCode: item.external_mold_code,
          });
        });
      } else {
        // 没有明细项，使用订单主表数据（兼容旧数据）
        const moldCode = order.mold?.mold_code || order.external_mold_info?.mold_code || '-';
        let productImageUrl = '';
        
        expandedOrders.push({
          id: order.id,
          itemId: undefined,
          orderNo: order.order_code,
          productName: order.product_name || '-',
          productImageUrl: productImageUrl || undefined,
          customerName: order.customer_name || order.customer?.cus_name,
          moldCode: moldCode,
          planQuantity: order.plan_qty,
          completedQuantity: order.completed_qty,
          status: order.status,
          statusInfo: getStatusInfo(order.status),
          deadline: order.plan_end_date ? new Date(order.plan_end_date).toISOString().split('T')[0] : '-',
          productionType: order.production_type,
          dependencyStatus: order.dependency_status || 'NONE',
          // ✨ 新增：保存产品ID和模具ID，用于入库
          productId: order.product_id,
          moldId: order.mold_id,
          externalMoldCode: order.external_mold_code,
        });
      }
    });
    
    orderList.value = expandedOrders;

    pagination.value.total = res.total || 0;
  } catch (e: any) {
    console.error('获取数据失败:', e);
    MessagePlugin.error(e.message || '获取数据失败');
    orderList.value = [];
    pagination.value.total = 0;
  } finally {
    dataLoading.value = false;
  }
};

// 根据后端状态值获取字典信息
const getStatusInfo = (status: any) => {
  // 尝试多种匹配方式：数字、字符串、语义化字符串（大小写不敏感）
  const dictItem = statusDictList.value.find((item: any) => {
    // 1. 精确匹配 value 字段
    if (item.value === status) return true;
    // 2. 大小写不敏感匹配
    if (typeof status === 'string' && item.value && String(item.value).toUpperCase() === status.toUpperCase()) return true;
    // 3. 如果 status 是数字，尝试与 dict_value 的数字形式匹配
    if (typeof status === 'number' && item.dict_value && parseInt(item.dict_value) === status) return true;
    // 4. 如果 status 是字符串，尝试与 dict_value 的字符串形式匹配
    if (typeof status === 'string' && item.dict_value && String(item.dict_value) === status) return true;
    // 5. 大小写不敏感匹配 dict_value
    if (typeof status === 'string' && item.dict_value && String(item.dict_value).toUpperCase() === status.toUpperCase()) return true;
    return false;
  });
  
  if (!dictItem) {
  }
  
  return dictItem || {
    value: status,
    label: '未知',
    theme: 'default' as const,
    themeStyle: undefined as { background: string; color: string; borderColor: string } | undefined,
    semanticTag: ''
  };
};

// 刷新数据
const handleRefresh = () => {
  pagination.value.current = 1;
  fetchData();
};

// 分页数据（后端已分页，直接返回当前页数据）
const paginatedOrders = computed(() => orderList.value);

// 分页事件
const onPageSizeChange = (size: number) => {
  pagination.value.pageSize = size;
  pagination.value.current = 1;
  fetchData();
};

const onCurrentChange = (current: number) => {
  pagination.value.current = current;
  fetchData();
};

// 报工操作
const handleReportWork = (order: WorkReportOrderType) => {
  currentOrder.value = order;
  reportForm.value = {
    reportDate: new Date().toISOString().split('T')[0],
    shift: undefined,
    okQty: undefined,
    badQty: undefined,
    remark: '',
  };
  reportDialogVisible.value = true;
};

// 确认报工
const handleConfirmReport = async () => {
  if (!currentOrder.value) return;
  
  try {
    await submitReport({
      order_id: currentOrder.value.id,
      item_id: currentOrder.value.itemId, // ✨ 新增：明细项ID
      report_date: reportForm.value.reportDate,
      shift: reportForm.value.shift,
      ok_qty: reportForm.value.okQty,
      bad_qty: reportForm.value.badQty,
      scrap_qty: 0,
      remark: reportForm.value.remark,
    });
    MessagePlugin.success('报工成功');
    reportDialogVisible.value = false;
    fetchData();
  } catch (error: any) {
    MessagePlugin.error(error.message || '报工失败');
  }
};

// 开始生产
const handleStartProduction = (order: WorkReportOrderType) => {
  currentOrder.value = order;
  confirmDialogTitle.value = '确认开始生产';
  confirmDialogContent.value = `确定要开始生产订单 ${order.orderNo} 吗？`;
  confirmActionCallback = async () => {
    try {
      await startProduction(order.id);
      MessagePlugin.success('已开始生产');
      fetchData();
    } catch (error: any) {
      MessagePlugin.error(error.message || '开始生产失败');
    }
  };
  confirmDialogVisible.value = true;
};

// 入库操作
const handleStockIn = async (order: WorkReportOrderType) => {
  currentOrder.value = order;
  
  // 直接使用卡片中保存的产品ID和模具ID
  const productId = order.productId;
  const moldId = order.moldId;
  const externalMoldCode = order.externalMoldCode;
  
  stockInForm.value = {
    warehouseId: undefined,
    moldId: moldId,
    externalMoldCode: externalMoldCode,
    productId: productId,
    quantity: 0,
    batchNo: '',
    remark: '',
  };
  
  
  await fetchWarehouseList();
  
  stockInDialogVisible.value = true;
};

const fetchWarehouseList = async () => {
  try {
    const res = await getWarehouseList({ page: 1, page_size: 100 });
    warehouseList.value = res.list || [];
  } catch (error) {
    console.error('获取仓库列表失败:', error);
    warehouseList.value = [];
  }
};

// 产品选择变化时同步 productId
const handleProductChange = (value: any) => {
  const selectedMold = orderMolds.value.find((m: any) => m.id === value);
  if (selectedMold) {
    stockInForm.value.productId = selectedMold.productId;
    stockInForm.value.externalMoldCode = selectedMold.externalMoldCode;
  }
};

// 确认入库
const handleConfirmStockIn = async () => {
  if (!currentOrder.value) return;
  
  if (!stockInForm.value.warehouseId) {
    MessagePlugin.warning('请选择仓库');
    return;
  }
  
  if (stockInForm.value.moldId == null && !stockInForm.value.externalMoldCode) {
    MessagePlugin.warning('请选择模具或输入外来模具编号');
    return;
  }
  
  if (!stockInForm.value.quantity || stockInForm.value.quantity <= 0) {
    MessagePlugin.warning('请输入有效的入库数量');
    return;
  }
  
  try {
    await stockInOrder({
      order_id: currentOrder.value.id,
      item_id: currentOrder.value.itemId, // ✨ 新增：传递明细项ID
      warehouse_id: stockInForm.value.warehouseId,
      mold_id: stockInForm.value.moldId,
      external_mold_code: stockInForm.value.externalMoldCode,
      product_id: stockInForm.value.productId,
      product_name: currentOrder.value.productName, // 传递产品名称
      quantity: stockInForm.value.quantity,
      batch_no: '', // 不传批次号，后端自动生成
      remark: stockInForm.value.remark,
    });
    MessagePlugin.success('入库成功');
    stockInDialogVisible.value = false;
    fetchData();
  } catch (error: any) {
    console.error('入库失败:', error);
    MessagePlugin.error(error.message || '入库失败');
  }
};

// 确认对话框确认
const handleConfirmAction = () => {
  if (confirmActionCallback) {
    confirmActionCallback();
    confirmDialogVisible.value = false;
  }
};

// 查看详情
const handleViewDetail = (order: WorkReportOrderType) => {
  // 跳转到订单详情页
  router.push({ name: 'ProductionOrderDetail', params: { id: String(order.id) } });
};



// 结束生产
const handleCompleteProduction = (order: WorkReportOrderType) => {
  currentOrder.value = order;
  
  // 生产中状态：调用完成生产API
  confirmDialogTitle.value = '确认结束生产';
  confirmDialogContent.value = `确定要结束生产产品 ${order.productName} 吗？结束生产后该产品将转为已入库，可在订单页进行发货。当前完成数量: ${order.completedQuantity || 0}`;
  confirmActionCallback = async () => {
    try {
      // ✨ 优先使用明细项级别 API（生产报工页是按明细项展开的）；
      //    没有明细项时回退到订单级 API（兼容旧数据）
      if (order.itemId) {
        await completeOrderItem(order.itemId);
      } else {
        await completeProduction(order.id);
      }
      MessagePlugin.success('生产已结束');
      fetchData();
    } catch (error: any) {
      MessagePlugin.error(error.message || '结束生产失败');
    }
  };
  
  confirmDialogVisible.value = true;
};

// 加载生产类型字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载班次字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载状态字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 页面加载时获取数据
onMounted(async () => {
  // 字典已通过 useDictRaw 自动加载
  await Promise.all([
  ]);
  await fetchData();
});

// ✨ KeepAlive 缓存下返回页面时自动刷新数据
onActivated(() => {
  fetchData();
});

// 搜索
const handleSearch = (formData?: Record<string, any>) => {
  if (formData) {
    Object.assign(searchForm, {
      orderCode: formData.orderCode ?? '',
      productName: formData.productName ?? '',
      status: formData.status ?? '',
    });
  }
  pagination.value.current = 1;
  fetchData();
};

const handleReset = (formData?: Record<string, any>) => {
  if (formData) {
    Object.assign(searchForm, {
      orderCode: formData.orderCode ?? '',
      productName: formData.productName ?? '',
      status: formData.status ?? '',
    });
  }
  pagination.value.current = 1;
  fetchData();
};

// Tab切换
const handleTabChange = (value: TabValue) => {
  activeTab.value = value as string;
  pagination.value.current = 1;
  fetchData();
};
</script>

<style lang="less" scoped>
.work-report-container {
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

.work-report-items {
  margin-bottom: 16px;
}

.work-report-item {
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingTB-xl);

  :deep(.t-card__header) {
    padding: 0;
  }

  :deep(.t-card__body) {
    padding: 0;
    margin-top: var(--td-comp-margin-xxl);
    margin-bottom: var(--td-comp-margin-xxl);
  }

  :deep(.t-card__footer) {
    padding: 0;
  }
}

.work-report-pagination {
  padding: var(--td-comp-paddingTB-xl) 0;
}

.work-report-loading {
  position: relative;
  height: calc(100vh - 120px);
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>

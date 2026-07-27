<template>
  <div class="production-order-list">
    <!-- 搜索栏（可选显示） -->
    <t-card v-if="showSearchBar" class="search-card" :bordered="false">
      <t-form layout="inline" :data="filterForm" :label-width="0">
        <t-form-item label="订单状态">
          <t-select v-model="filterForm.status" placeholder="请选择" clearable style="width: 150px">
            <t-option 
              v-for="status in statusDictList" 
              :key="status.value" 
              :value="status.value" 
              :label="status.label" 
            />
          </t-select>
        </t-form-item>
        <t-form-item label="生产类型">
          <t-select v-model="filterForm.production_type" placeholder="请选择" clearable style="width: 150px">
            <t-option 
              v-for="item in productionTypeDict" 
              :key="item.dict_value" 
              :value="item.dict_value" 
              :label="item.dict_label" 
            />
          </t-select>
        </t-form-item>
        <t-form-item label="优先级">
          <t-select v-model="filterForm.priority" placeholder="请选择" clearable style="width: 150px">
            <t-option :value="1" label="普通" />
            <t-option :value="2" label="重要" />
            <t-option :value="3" label="紧急" />
            <t-option :value="4" label="非常紧急" />
          </t-select>
        </t-form-item>
        <t-form-item label="关键词">
          <t-input v-model="searchKeyword" placeholder="搜索订单号、客户名称、产品名称" clearable style="width: 300px">
            <template #suffix-icon>
              <search-icon size="16px" />
            </template>
          </t-input>
        </t-form-item>
        <t-form-item>
          <t-button theme="primary" @click="handleSearch">查询</t-button>
          <t-button @click="handleReset" style="margin-left: 8px">重置</t-button>
        </t-form-item>
      </t-form>
    </t-card>

    <!-- 工具栏（在订单中心页面中隐藏） -->
    
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
      <!-- 订单号 -->
      <template #order_code="{ row }">
        <t-link @click="handleView(row)" class="order-code-link">
          {{ row.order_code }}
        </t-link>
      </template>
      
      <!-- 来源订单 -->
      <template #master_order_code="{ row }">
        <span v-if="row.master_order?.order_code" style="color: #0052d9; cursor: pointer" @click="handleViewMasterOrder(row)">
          {{ row.master_order.order_code }}
        </span>
        <span v-else style="color: #999">-</span>
      </template>
      
      <!-- 优先级 -->
      <template #priority="{ row }">
        <template v-for="item in priorityDict" :key="item.value">
          <t-tag
            v-if="String(row.priority) === String(item.value)"
            shape="round"
            :theme="item.theme || 'default'"
            :style="item.themeStyle"
            variant="outline"
          >
            {{ item.label }}
          </t-tag>
        </template>
        <span v-if="!priorityDict.find(item => String(item.value) === String(row.priority))" class="debug-info">
          未匹配: {{ row.priority }}
        </span>
      </template>
      
      <!-- 产品图片 -->
      <template #product_image="{ row }">
        <template v-if="row.items && row.items.length > 0">
          <div
            v-for="(item, index) in row.items"
            :key="index"
            class="product-image-item"
          >
            <img
              v-if="item.product_image_url"
              :src="getThumbnailUrl(item.product_image_url)"
              class="product-image-thumb"
              @click="handlePreviewImage(row, Number(index))"
            />
            <div v-else class="product-image-placeholder">无图片</div>
          </div>
        </template>
        <div v-else class="product-image-item">
          <img
            v-if="row.product_image_url"
            :src="getThumbnailUrl(row.product_image_url)"
            class="product-image-thumb"
            @click="handlePreviewSingleImage(row)"
          />
          <div v-else class="product-image-placeholder">无图片</div>
        </div>
      </template>
      
      <!-- 产品名称 -->
      <template #product_name="{ row }">
        <template v-if="row.items && row.items.length > 0">
          <div
            v-for="(item, index) in row.items"
            :key="index"
            class="product-name-item"
          >
            <div class="product-name-content">
              {{ item.product?.product_name || item.product_name || '-' }}
            </div>
          </div>
        </template>
        <span v-else>{{ row.product_name || '-' }}</span>
      </template>
      
      <!-- 模具编号 -->
      <template #mold_code="{ row }">
        <template v-if="row.items && row.items.length > 0">
          <div
            v-for="(item, index) in row.items"
            :key="index"
            class="mold-code-item"
          >
            <div class="mold-code-content">
              <!-- 多模具情况 -->
              <template v-if="item.molds && item.molds.length > 0">
                <div v-for="(mold, moldIndex) in item.molds" :key="moldIndex">
                  {{ mold.mold_code }}
                </div>
              </template>
              <!-- 单模具或外来模具 -->
              <template v-else>
                {{ item.mold?.mold_code || item.external_mold_code || '-' }}
              </template>
            </div>
          </div>
        </template>
        <span v-else>{{ row.mold_code || '-' }}</span>
      </template>
      
      <!-- 生产类型 -->
      <template #production_type="{ row }">
        <template v-for="item in productionTypeDict" :key="item.dict_value">
          <t-tag
            v-if="row.production_type === item.dict_value"
            shape="round"
            variant="outline"
            :color="getDictColorHex(item.color)"
          >
            {{ item.dict_label }}
          </t-tag>
        </template>
      </template>
      
      <!-- 完成数量 -->
      <template #completed_qty="{ row }">
        {{ row.completed_qty || 0 }}
      </template>
      
      <!-- 合格数量 -->
      <template #ok_qty="{ row }">
        {{ row.ok_qty || 0 }}
      </template>
      
      <!-- 不良数量 -->
      <template #bad_qty="{ row }">
        {{ row.bad_qty || 0 }}
      </template>
      
      <!-- 进度 -->
      <template #progress="{ row }">
        <t-progress
          v-if="row.plan_qty > 0"
          :percentage="Math.round((row.completed_qty / row.plan_qty) * 100)"
          :label="true"
        />
        <span v-else>-</span>
      </template>
      
      <!-- 计划开始日期 -->
      <template #plan_start_date="{ row }">
        {{ formatDate(row.plan_start_date) }}
      </template>
      
      <!-- 计划结束日期 -->
      <template #plan_end_date="{ row }">
        {{ formatDate(row.plan_end_date) }}
      </template>
      
      <!-- 创建时间 -->
      <template #create_time="{ row }">
        {{ formatDate(row.create_time) }}
      </template>
      
      <!-- 状态 -->
      <template #status="{ row }">
        <template v-for="status in statusDictList" :key="status.value">
          <t-tag 
            v-if="row.status === status.value"
            shape="round"
            :theme="status.theme || 'default'"
            :style="status.themeStyle"
          >
            {{ getDisplayStatus(row, status) }}
          </t-tag>
        </template>
      </template>
      
      <!-- 操作 -->
      <template #op="{ row }">
        <t-space :size="8" break-line>
          <t-link theme="primary" @click="handleView(row)">详情</t-link>
          <t-link v-if="showEditButton" theme="primary" @click="handleEdit(row)">编辑</t-link>
          <t-link theme="primary" @click="handlePrint(row)">
            <template #icon><print-icon /></template>
            打印
          </t-link>
          
          <!-- ✨ 等待模具提示（可点击跳转） -->
          <t-link
            v-if="row.dependency_status === 'WAITING_MOLD' && row.items && row.items.length > 0"
            theme="warning"
            @click="handleViewMoldWorkOrder(row)"
            style="margin: 4px 0; cursor: pointer"
          >
            <template #icon>
              <t-icon name="time" />
            </template>
            等待模具试模
          </t-link>
          <!-- 兼容旧数据格式 -->
          <t-tag
            v-else-if="row.dependency_status === 'WAITING_MOLD'"
            theme="warning"
            variant="light"
            style="margin: 4px 0"
          >
            <template #icon>
              <t-icon name="time" />
            </template>
            等待模具试模
          </t-tag>
          
          <!-- 动态操作按钮 - 根据订单状态显示可用操作 -->
          <t-link
            v-if="showAdvancedActions && hasAvailableAction(row, 'start_production')"
            theme="primary"
            @click="handleStartProduction(row)"
          >
            开始生产
          </t-link>
          <t-link
            v-if="showAdvancedActions && hasAvailableAction(row, 'complete_production')"
            theme="primary"
            @click="handleCompleteProduction(row)"
          >
            完成生产
          </t-link>
          <t-link
            v-if="showAdvancedActions && hasAvailableAction(row, 'stock_in')"
            theme="primary"
            @click="handleStockIn(row)"
          >
            入库
          </t-link>
          <t-link
            v-if="showAdvancedActions && hasAvailableAction(row, 'shipment')"
            theme="primary"
            @click="handleShipment(row)"
          >
            {{ hasShipmentRecord(row) ? '继续出货' : '出货' }}
          </t-link>
          <t-popconfirm
            v-if="showAdvancedActions && hasAvailableAction(row, 'finish_order')"
            :content="'已出货 '+ (row.shipped_qty || 0) +'/'+ row.plan_qty +'，确认完成该订单？'"
            @confirm="handleFinishOrder(row)"
          >
            <t-link theme="success">完成订单</t-link>
          </t-popconfirm>
          <t-link
            v-if="showAdvancedActions && row.status === 'COMPLETED'"
            theme="primary"
            @click="handleReplenish(row)"
          >
            补发
          </t-link>
          <t-popconfirm
            v-if="showAdvancedActions && hasAvailableAction(row, 'cancel_order')"
            content="确认取消该订单？"
            @confirm="handleCancel(row)"
          >
            <t-link theme="danger">取消</t-link>
          </t-popconfirm>
        </t-space>
      </template>
    </t-table>
    
    <!-- 取消订单对话框 -->
    <t-dialog
      v-model:visible="cancelDialogVisible"
      header="取消订单"
      placement="center"
      :body="cancelDialogBody"
      @confirm="onConfirmCancel"
     :close-on-overlay-click="false">
      <t-textarea
        v-model="cancelReason"
        placeholder="请输入取消原因（可选）"
        :autosize="{ minRows: 3, maxRows: 5 }"
      />
    </t-dialog>

    <!-- 补发对话框 -->
    <t-dialog
      v-model:visible="replenishDialogVisible"
      header="补发出货"
      placement="center"
      @confirm="onConfirmReplenish"
     :close-on-overlay-click="false">
      <t-form :data="replenishForm" label-width="100px">
        <t-form-item label="订单编号">
          <span>{{ replenishForm.orderCode }}</span>
        </t-form-item>
        <t-form-item label="产品名称">
          <span>{{ replenishForm.productName }}</span>
        </t-form-item>
        <t-form-item label="可补发数量">
          <t-tag theme="warning" variant="light">{{ replenishForm.remainingQty }}</t-tag>
        </t-form-item>
        <t-form-item label="补发备注">
          <t-textarea
            v-model="replenishForm.remark"
            placeholder="输入补发说明（可选）"
            :rows="2"
          />
        </t-form-item>
      </t-form>
    </t-dialog>
    
    <!-- 图片预览 -->
    <t-image-viewer
      v-model:visible="imagePreviewVisible"
      :images="previewImages"
      :default-index="previewIndex"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import type { PrimaryTableCol, PageInfo } from 'tdesign-vue-next';
import { SearchIcon, AddIcon, PrintIcon } from 'tdesign-icons-vue-next';
import { 
  getProductionOrders, 
  cancelProductionOrder,
  getOrderStatistics,
  startProduction,
  completeProduction,
  shipmentOrder,
  completeOrder,
  replenishShipment,
} from '@/api/production/order';
import { useDictRaw } from '@/hooks/useDict';
import { getDictColorHex, getDictColorRenderConfig } from '@/utils/dict-color';
import { MessagePlugin } from 'tdesign-vue-next';
import { OPERATION_COLUMN_WIDTH } from '@/utils/operation-column-width';
import { formatDate } from '@/utils/date';
import { getThumbnailUrl } from '@/api/common/upload';
import type { ProductionOrder } from '@/types/production';

const router = useRouter();
const loading = ref(false);
const tableData = ref<any[]>([]);
const pagination = ref({ current: 1, pageSize: 10, total: 0 });

// Props - 控制显示哪些功能
interface Props {
  showSearchBar?: boolean; // 是否显示搜索栏
  showCreateButton?: boolean; // 是否显示新建按钮
  showEditButton?: boolean; // 是否显示编辑按钮
  showAdvancedActions?: boolean; // 是否显示高级操作（开始生产、完成等）
  showTitle?: boolean; // 是否显示标题
  toolbarTitle?: string; // 工具栏标题
  searchKeyword?: string; // 从父组件传入的搜索关键词
  filterStatus?: number; // 从父组件传入的状态筛选
  filterProductionType?: 1 | 2; // 从父组件传入的生产类型筛选
  filterPriority?: 1 | 2 | 3 | 4; // 从父组件传入的优先级筛选
  createTimeStart?: string; // 创建日期范围-开始
  createTimeEnd?: string; // 创建日期范围-结束
  dueSoon?: string; // 到期预警天数
}

const props = withDefaults(defineProps<Props>(), {
  showSearchBar: false,
  showCreateButton: false,
  showEditButton: true,
  showAdvancedActions: false,
  showTitle: false,
  toolbarTitle: '生产订单管理',
  searchKeyword: '',
  filterStatus: undefined,
  filterProductionType: undefined,
  filterPriority: undefined,
  createTimeStart: '',
  createTimeEnd: '',
  dueSoon: '',
});

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: statusDictList } = useDictRaw('production-order-status', (items) => {
  const dictList = (items as any[]).filter(
    (item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '',
  );
  return dictList.map((item: any) => {
    const color = item.color || 'default';
    const renderConfig = getDictColorRenderConfig(color);
    return {
      value: item.dict_value,
      label: item.dict_label,
      color: color,
      theme: renderConfig.mode === 'theme' ? renderConfig.theme : 'default',
      themeStyle: renderConfig.mode === 'style'
        ? { background: renderConfig.style!.background, color: renderConfig.style!.color, borderColor: renderConfig.style!.borderColor }
        : undefined,
    };
  });
});
const { items: productionTypeDict } = useDictRaw('production-type', (items) =>
  (items as any[]).filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== ''),
);
const { items: priorityDict } = useDictRaw('order-priority', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '')
    .map((item: any) => {
      const color = item.color || 'default';
      const renderConfig = getDictColorRenderConfig(color);
      return {
        value: item.dict_value,
        label: item.dict_label,
        color,
        theme: renderConfig.mode === 'theme' ? renderConfig.theme : 'default',
        themeStyle: renderConfig.mode === 'style'
          ? { background: renderConfig.style!.background, color: renderConfig.style!.color, borderColor: renderConfig.style!.borderColor }
          : undefined,
      };
    }),
);

// 搜索和筛选
const searchKeyword = ref('');
const filterForm = ref({
  status: undefined as number | undefined,
  production_type: undefined as 1 | 2 | undefined,
  priority: undefined as 1 | 2 | 3 | 4 | undefined,
});

// 取消订单对话框
const cancelDialogVisible = ref(false);
const cancelReason = ref('');
const cancelOrderId = ref<number | null>(null);

// 补发对话框
const replenishDialogVisible = ref(false);
const replenishForm = ref({
  orderId: 0,
  orderCode: '',
  productName: '',
  remainingQty: 0,
  remark: '',
});
const replenishOrderId = ref<number | null>(null);

const cancelDialogBody = computed(() => {
  const order = tableData.value.find((item) => item.id === cancelOrderId.value);
  if (order) {
    return `确认要取消订单 "${order.order_code}" 吗？`;
  }
  return '';
});

const hasAvailableAction = (order: ProductionOrder, actionCode: string): boolean => {
  const actions = order.available_action_codes || [];
  return actions.includes(actionCode);
};

// 固定列配置（用于表格渲染）
const columns: PrimaryTableCol[] = [
  { colKey: 'order_code', title: '订单号', width: 180, fixed: 'left' },
  { colKey: 'master_order_code', title: '来源订单', width: 180 },
  { colKey: 'priority', title: '优先级', width: 100 },
  { colKey: 'customer.cus_name', title: '客户名称', width: 150 },
  { colKey: 'product_name', title: '产品名称', width: 200 },
  { colKey: 'product_image', title: '产品图片', width: 120 },
  { colKey: 'mold_code', title: '模具编号', width: 150 },
  { colKey: 'production_type', title: '生产类型', width: 120 },
  { colKey: 'progress', title: '进度', width: 150 },
  { colKey: 'plan_qty', title: '计划数量', width: 100 },
  { colKey: 'completed_qty', title: '完成数量', width: 100 },
  { colKey: 'ok_qty', title: '合格数量', width: 100 },
  { colKey: 'bad_qty', title: '不良数量', width: 100 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'plan_start_date', title: '计划开始日期', width: 120 },
  { colKey: 'plan_end_date', title: '订单交期', width: 120 },
  { colKey: 'create_time', title: '创建时间', width: 160 },
  { colKey: 'op', title: '操作', width: OPERATION_COLUMN_WIDTH.MEDIUM, fixed: 'right' },
];

// 动态列配置（根据用户选择显示）
const selectedColumns = ref<string[]>([]);
const columnSettingRef = ref();

const displayColumns = computed(() => {
  const orderedColumns = selectedColumns.value
    .map(colKey => columns.find(col => col.colKey === colKey))
    .filter(col => col !== undefined);
  
  // 确保操作列始终在最后且固定右侧
  const operationCol = columns.find(col => col.colKey === 'op');
  if (operationCol) {
    orderedColumns.push(operationCol);
  }
  
  return orderedColumns;
});

// 所有可选列配置（用于列管理）
const allColumns = [
  { colKey: 'order_code', title: '订单号', pinned: true },
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
  { colKey: 'plan_end_date', title: '订单交期' },
  { colKey: 'create_time', title: '创建时间' },
];

const DEFAULT_COLUMNS = [
  'order_code', 'master_order_code', 'priority', 'customer', 'product_name', 'product_image', 'mold_code',
  'production_type', 'plan_qty', 'completed_qty', 'ok_qty', 'bad_qty', 'progress', 'status',
  'plan_start_date', 'plan_end_date', 'create_time',
];

// 列设置相关方法
const updateColumns = (cols: string[]) => {
  selectedColumns.value = cols;
};

const handleRefresh = () => {
  fetchData();
};

// 搜索
const handleSearch = () => {
  pagination.value.current = 1;
  fetchData();
};

const handleReset = () => {
  searchKeyword.value = '';
  filterForm.value = {
    status: undefined,
    production_type: undefined,
    priority: undefined,
  };
  handleSearch();
};

const handleView = (row: any) => {
  router.push(`/production/order/detail/${row.id}`);
};

const handleEdit = (row: any) => {
  router.push(`/production/order/edit/${row.id}`);
};

// 查看来源订单（综合订单）
const handleViewMasterOrder = (row: any) => {
  if (row.master_order?.id) {
    router.push(`/order-management/detail/${row.master_order.id}`);
  }
};

// 查看模具工单（等待模具试模时跳转）
const handleViewMoldWorkOrder = (row: any) => {
  // 从订单明细中获取模具编号
  let moldCode = '';
  if (row.items && row.items.length > 0) {
    const firstItem = row.items[0];
    // 优先使用模具对象的编号
    if (firstItem.mold?.mold_code) {
      moldCode = firstItem.mold.mold_code;
    } else if (firstItem.mold_code) {
      moldCode = firstItem.mold_code;
    }
  }
  
  if (!moldCode) {
    MessagePlugin.warning('未找到模具编号');
    return;
  }
  
  // 跳转到模具工单页面并携带模具编号参数
  router.push({
    path: '/mold/work-order',
    query: {
      moldCode: moldCode,
    },
  });
};

const handleCreateOrder = () => {
  router.push('/production/order/create');
};

const handlePrint = (row: ProductionOrder) => {
  // 跳转到订单详情页并触发打印
  router.push({
    path: `/production/order/detail/${row.id}`,
    query: { print: '1' },
  });
};

// 取消订单
const handleCancel = (row: ProductionOrder) => {
  cancelOrderId.value = row.id;
  cancelReason.value = '';
  cancelDialogVisible.value = true;
};

// 确认取消
const onConfirmCancel = async () => {
  if (!cancelOrderId.value) return;

  try {
    await cancelProductionOrder(cancelOrderId.value, cancelReason.value);
    MessagePlugin.success('已取消订单');
    cancelDialogVisible.value = false;
    handleRefresh();
  } catch (error) {
    console.error('取消订单失败:', error);
    MessagePlugin.error('取消订单失败');
  }
};

// 开始生产
const handleStartProduction = async (row: ProductionOrder) => {
  try {
    await startProduction(row.id);
    MessagePlugin.success('已开始生产');
    handleRefresh();
  } catch (error) {
    console.error('开始生产失败:', error);
    MessagePlugin.error('开始生产失败');
  }
};

// 完成生产
const handleCompleteProduction = async (row: ProductionOrder) => {
  try {
    await completeProduction(row.id);
    MessagePlugin.success('已完成生产');
    handleRefresh();
  } catch (error) {
    console.error('完成生产失败:', error);
    MessagePlugin.error('完成生产失败');
  }
};

// 入库
const handleStockIn = (row: ProductionOrder) => {
  MessagePlugin.info('入库功能开发中');
};

// 出货 - 跳转到出库单填写页面
const handleShipment = (row: ProductionOrder) => {
  // 从订单明细中获取外部模具编号（如果有）
  let externalMoldCode = '';
  if (row.items && row.items.length > 0) {
    const firstItem = row.items[0];
    if (firstItem.external_mold_code) {
      externalMoldCode = firstItem.external_mold_code;
    }
  }
  
  // 跳转到出库单填写页面，传递订单信息
  router.push({
    path: '/production/stock-out/create',
    query: {
      order_id: row.id,
      order_code: row.order_code,
      mold_id: row.mold_id?.toString() || '',
      external_mold_code: externalMoldCode,
      product_name: row.product_name || '',
    },
  });
};

const hasShipmentRecord = (row: ProductionOrder): boolean => {
  // 从后端数据中获取，如果订单有 shipment_count 字段
  // 暂时根据订单状态判断：如果状态是 SHIPPED 则说明已有出货记录
  return row.status === 'SHIPPED';
};

const getDisplayStatus = (row: ProductionOrder, status: any): string => {
  // 如果有出货记录，强制显示"已出货"
  if (hasShipmentRecord(row)) {
    // 查找"已出货"状态的label
    const shippedStatus = statusDictList.value.find(s => s.value === 'SHIPPED');
    return shippedStatus ? shippedStatus.label : status.label;
  }
  return status.label;
};

// 完成订单
const handleFinishOrder = async (row: ProductionOrder) => {
  try {
    await completeOrder(row.id);
    MessagePlugin.success('订单已完成');
    handleRefresh();
  } catch (error) {
    console.error('完成订单失败:', error);
    MessagePlugin.error('完成订单失败');
  }
};

// 补发
const handleReplenish = (row: ProductionOrder) => {
  const remaining = (row.ok_qty || 0) - (row.shipped_qty || 0);
  replenishForm.value = {
    orderId: row.id,
    orderCode: row.order_code,
    productName: row.product_name || '',
    remainingQty: remaining > 0 ? remaining : 0,
    remark: '',
  };
  replenishOrderId.value = row.id;
  replenishDialogVisible.value = true;
};

// 确认补发
const onConfirmReplenish = async () => {
  if (!replenishOrderId.value) return;
  try {
    await replenishShipment(replenishOrderId.value, { remark: replenishForm.value.remark });
    MessagePlugin.success(`补发 ${replenishForm.value.remainingQty} 个成功`);
    replenishDialogVisible.value = false;
    handleRefresh();
  } catch (error) {
    console.error('补发失败:', error);
    MessagePlugin.error('补发失败');
  }
};

// 图片预览相关
const imagePreviewVisible = ref(false);
const previewImages = ref<string[]>([]);
const previewIndex = ref(0);

const handlePreviewImage = (row: any, index: number) => {
  if (row.items && row.items.length > 0) {
    const images = row.items
      .map((item: any) => item.product_image_url)
      .filter((url: string) => url);
    if (images.length > 0) {
      previewImages.value = images;
      previewIndex.value = index;
      imagePreviewVisible.value = true;
    }
  }
};

const handlePreviewSingleImage = (row: any) => {
  if (row.product_image_url) {
    previewImages.value = [row.product_image_url];
    previewIndex.value = 0;
    imagePreviewVisible.value = true;
  }
};

const handlePageChange = (pageInfo: PageInfo) => {
  pagination.value.current = pageInfo.current;
  pagination.value.pageSize = pageInfo.pageSize;
  fetchData();
};

const fetchData = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
    };

    // 优先使用 props 传入的搜索条件（订单中心模式），否则使用内部状态（生产管理模式）
    const keyword = props.searchKeyword || searchKeyword.value;
    const status = props.filterStatus !== undefined ? props.filterStatus : filterForm.value.status;
    const productionType = props.filterProductionType !== undefined ? props.filterProductionType : filterForm.value.production_type;
    const priority = props.filterPriority !== undefined ? props.filterPriority : filterForm.value.priority;

    if (keyword) {
      params.keyword = keyword;
    }
    if (status !== undefined) {
      params.status = status;
    }
    if (productionType !== undefined) {
      params.production_type = productionType;
    }
    if (priority !== undefined) {
      params.priority = priority;
    }
    if (props.createTimeStart) params.create_time_start = props.createTimeStart;
    if (props.createTimeEnd) params.create_time_end = props.createTimeEnd;
    if (props.dueSoon) params.due_soon = props.dueSoon;

    const res = await getProductionOrders(params);
    
    tableData.value = res.list || [];
    pagination.value.total = res.total || 0;
    // 可用操作编码已由后端直接返回在 order.available_action_codes 中
  } catch (error) {
    console.error('获取生产订单列表失败:', error);
    MessagePlugin.error('获取生产订单列表失败');
  } finally {
    loading.value = false;
  }
};

onMounted(async () => {
  try {
    const saved = localStorage.getItem('production_order_table_columns');
    if (saved) {
      const savedData = JSON.parse(saved);
      if (Array.isArray(savedData)) {
        selectedColumns.value = savedData;
      } else if (savedData.visible) {
        selectedColumns.value = savedData.visible;
      } else if (savedData.columns) {
        selectedColumns.value = savedData.columns;
      } else if (savedData.order) {
        selectedColumns.value = savedData.visible || savedData.order;
      } else {
        selectedColumns.value = [...DEFAULT_COLUMNS];
      }
    } else {
      selectedColumns.value = [...DEFAULT_COLUMNS];
    }
    
    // 确保列配置有效
    if (!selectedColumns.value.length) {
      selectedColumns.value = [...DEFAULT_COLUMNS];
    }
  } catch (error) {
    console.error('加载列配置失败:', error);
    selectedColumns.value = [...DEFAULT_COLUMNS];
  }
  
  await fetchData();
});

// 暴露方法给父组件
defineExpose({
  fetchData,
  updateColumns,
});
</script>

<style scoped>
.production-order-list {
  /* 移除了 padding，由父容器控制 */
}

.search-card {
  margin-bottom: 16px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.order-code-link {
  font-size: 15px;
}

.product-name-item,
.mold-code-item,
.product-image-item {
  min-height: 70px;
  padding: 2px 0;
  font-size: 13px;
  line-height: 1.5;
  display: flex;
  align-items: center;
}

.product-name-content,
.mold-code-content {
  width: 100%;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.product-name-item:not(:last-child),
.mold-code-item:not(:last-child),
.product-image-item:not(:last-child) {
  border-bottom: 1px dashed #e7e7e7;
  margin-bottom: 4px;
  padding-bottom: 4px;
}

/* 表格单元格垂直居中 */
:deep(.t-table td[data-colkey="product_name"]),
:deep(.t-table td[data-colkey="product_image"]),
:deep(.t-table td[data-colkey="mold_code"]) {
  vertical-align: middle !important;
}

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
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  border-radius: 4px;
  color: #999;
  font-size: 12px;
}
</style>

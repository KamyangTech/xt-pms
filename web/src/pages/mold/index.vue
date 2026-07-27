<template>
  <div class="mold-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 统计信息卡片 -->
    <t-row :gutter="16" class="statistics-row">
      <t-col :span="3">
        <t-card :bordered="false" class="stat-card stat-card-all" @click="handleStatFilter('all')">
          <div class="stat-content">
            <div class="stat-icon stat-icon-primary">
              <layers-icon />
            </div>
            <div class="stat-info">
              <div class="stat-label">模具总数</div>
              <div class="stat-value">{{ statistics.total }}</div>
            </div>
          </div>
        </t-card>
      </t-col>
      <t-col :span="3">
        <t-card :bordered="false" class="stat-card stat-card-processing" @click="handleStatFilter('PROCESSING')">
          <div class="stat-content">
            <div class="stat-icon stat-icon-warning">
              <time-icon />
            </div>
            <div class="stat-info">
              <div class="stat-label">加工中</div>
              <div class="stat-value">{{ statistics.processing }}</div>
            </div>
          </div>
        </t-card>
      </t-col>
      <t-col :span="3">
        <t-card :bordered="false" class="stat-card stat-card-completed" @click="handleStatFilter('completed')">
          <div class="stat-content">
            <div class="stat-icon stat-icon-success">
              <check-circle-icon />
            </div>
            <div class="stat-info">
              <div class="stat-label">已完成</div>
              <div class="stat-value">{{ statistics.completed }}</div>
            </div>
          </div>
        </t-card>
      </t-col>
      <t-col :span="3">
        <t-card :bordered="false" class="stat-card stat-card-repairing" @click="handleStatFilter('REPAIRING')">
          <div class="stat-content">
            <div class="stat-icon stat-icon-danger">
              <info-circle-icon />
            </div>
            <div class="stat-info">
              <div class="stat-label">维修中</div>
              <div class="stat-value">{{ statistics.repairing }}</div>
            </div>
          </div>
        </t-card>
      </t-col>
    </t-row>

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">模具列表</span>
        <div class="toolbar-actions">
          <t-button theme="default" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button theme="default" @click="handleExport">
            <template #icon><download-icon /></template>
            导出
          </t-button>
          <t-button theme="default" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
          <t-button theme="primary" @click="handleAdd">
            <template #icon><add-icon /></template>
            新增模具
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <t-table
        :data="tableData"
        :columns="displayColumns"
        :loading="loading"
        row-key="id"
        :pagination="pagination"
        :stripe="false"
        @page-change="rehandlePageChange"
      >
        <template #moldCode="{ row }">
          <t-link @click="handleView(row)" class="mold-code-link">
            {{ row.moldCode }}
          </t-link>
        </template>
        <template #primaryImage="{ row }">
          <t-image-viewer
            v-if="row.primaryImage"
            :images="[row.primaryImage]"
            :z-index="10000"
          >
            <template #trigger="{ open }">
              <img
                :src="row.primaryImage"
                class="product-image-thumb"
                @click="open"
              />
            </template>
          </t-image-viewer>
          <div v-else class="product-image-placeholder">无图片</div>
        </template>
        <template #lifeCycleStatus="{ row }">
          <t-tooltip :content="getLifeCycleStatusTooltip(row)">
            <t-tag
              v-if="getLifeCycleStatusRenderConfig(row.lifeCycleStatus).mode === 'theme'"
              :theme="getLifeCycleStatusRenderConfig(row.lifeCycleStatus).theme"
              shape="round"
            >
              {{ getLifeCycleStatusLabel(row.lifeCycleStatus) }}
            </t-tag>
            <t-tag
              v-else
              :style="getLifeCycleStatusRenderConfig(row.lifeCycleStatus).style"
              shape="round"
            >
              {{ getLifeCycleStatusLabel(row.lifeCycleStatus) }}
            </t-tag>
          </t-tooltip>
        </template>
        <template #operation="{ row }">
          <div style="width: 220px;">
            <t-space :size="8" break-line>
              <t-link theme="primary" @click="handleView(row)">详情</t-link>
              <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
              <t-dropdown
                trigger="click"
                :options="getMoreDropdownOptions(row)"
                @click="(data: any) => handleMoreAction(data, row)"
              >
                <t-link theme="default">更多</t-link>
              </t-dropdown>
              <t-tag
                v-if="row.lifeCycleStatus === 'SCRAPPED'"
                theme="default"
                variant="outline"
                size="small"
              >已报废</t-tag>
            </t-space>
          </div>
        </template>
      </t-table>
    </t-card>

    <!-- 查看模具详情对话框 -->
    <t-dialog
      v-model:visible="detailVisible"
      title="模具详情"
      width="800px"
      :footer="null"
     :close-on-overlay-click="false">
      <t-descriptions bordered :column="2">
        <t-descriptions-item label="模具编号">{{ detailData.moldCode }}</t-descriptions-item>
        <t-descriptions-item label="产品名称">{{ detailData.moldName || '-' }}</t-descriptions-item>
        <t-descriptions-item label="客户">{{ detailData.customerName || '-' }}</t-descriptions-item>
        <t-descriptions-item label="开始生产日期">{{ detailData.productionDate?.split('T')[0] || detailData.productionDate || '-' }}</t-descriptions-item>
        <t-descriptions-item label="状态">{{ getLifeCycleStatusText(detailData.lifeCycleStatus) }}</t-descriptions-item>
        <t-descriptions-item label="模胚情况">{{ getMoldTypeText(detailData.moldType) }}</t-descriptions-item>
        <t-descriptions-item label="模具工艺">{{ getMoldProcessText(detailData.moldProcess) }}</t-descriptions-item>
        <t-descriptions-item label="材质">{{ getMaterialTypeText(detailData.materialType) }}</t-descriptions-item>
        <t-descriptions-item label="创建时间">{{ detailData.createTime?.split('T')[0] || detailData.createTime || '-' }}</t-descriptions-item>
        <t-descriptions-item label="备注" :span="2">
          <div style="white-space: pre-wrap;">{{ detailData.remark || '-' }}</div>
        </t-descriptions-item>
        <t-descriptions-item label="模具图片" :span="2">
          <MoldImageGallery :images="detailImages" />
        </t-descriptions-item>
      </t-descriptions>
    </t-dialog>

    <!-- 图片预览对话框 -->
    <t-image-viewer
      v-model:visible="imagePreviewVisible"
      :images="previewImages"
      :default-index="previewIndex"
    />

    <!-- 列设置对话框 - 使用通用组件 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="mold_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onActivated, reactive, ref, computed } from 'vue';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import type { TableProps } from 'tdesign-vue-next';
import { 
  AddIcon, 
  SettingIcon, 
  RefreshIcon, 
  DownloadIcon,
  LayersIcon,
  TimeIcon,
  CheckCircleIcon,
  InfoCircleIcon,
} from 'tdesign-icons-vue-next';
import { getMoldList as getMoldListApi, getMoldStatistics, updateMold } from '@/api/mold';
import { getFileFullUrl } from '@/api/common/upload';
import { getCustomerList as getCustomerListApi, getDictTypes } from '@/api/reference';
import { getDictColorRenderConfig } from '@/utils/dict-color';
import { OPERATION_COLUMN_WIDTH } from '@/utils/operation-column-width';

import MoldImageGallery from '@/components/mold/MoldImageGallery.vue';
import ImagePlaceholder from '@/components/ImagePlaceholder/index.vue';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import { useRouter } from 'vue-router';

const router = useRouter();

defineOptions({
  name: 'MoldIndex',
});

// 表格相关
const loading = ref(false);
const tableData = ref<any[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

// 统计信息
const statistics = reactive({
  total: 0,
  processing: 0,
  completed: 0,
  repairing: 0,
});

// 搜索表单
const formData = reactive({
  moldCode: '',
  moldName: '',
  customerId: undefined,
  lifeCycleStatus: undefined,
  moldType: undefined,
  moldProcess: undefined,
  materialType: undefined,
  productionDateRange: [],
});

// 客户列表
const customerList = ref<any[]>([]);

const getCustomerList = async () => {
  try {
    const res = await getCustomerListApi();
    customerList.value = res.list || [];
  } catch (error) {
    console.error('获取客户列表失败:', error);
    customerList.value = [];
  }
};

// 表格列配置
const columns: TableProps['columns'] = [
  { 
    colKey: 'moldCode', 
    title: '模具编号', 
    width: 150,
    fixed: 'left',
  },
  { 
    colKey: 'moldName',
    title: '产品名称',
    width: 150,
    cell: (h, { row }) => row.moldName || '-',
  },
  { 
    colKey: 'primaryImage', 
    title: '模具图片', 
    width: 120,
  },
  { 
    colKey: 'customerName', 
    title: '客户名称', 
    width: 150,
    cell: (h, { row }) => row.customerName || row.customer?.cus_name || '-',
  },
  { 
    colKey: 'productionDate', 
    title: '开始生产日期', 
    width: 120,
    cell: (h, { row }) => row.productionDate?.split('T')[0] || row.productionDate || '-',
  },
  { colKey: 'lifeCycleStatus', title: '状态', width: 100 },
  { 
    colKey: 'moldType', 
    title: '模胚', 
    width: 80,
    cell: (h, { row }) => getMoldTypeText(row.moldType),
  },
  { 
    colKey: 'moldProcess', 
    title: '工艺', 
    width: 100,
    cell: (h, { row }) => getMoldProcessText(row.moldProcess),
  },
  { 
    colKey: 'materialType', 
    title: '材质', 
    width: 100,
    cell: (h, { row }) => getMaterialTypeText(row.materialType),
  },
];

// 将操作列添加到 columns 数组
const columnsWithOperation = [
  ...columns,
  { colKey: 'operation', title: '操作', width: OPERATION_COLUMN_WIDTH.LARGE, fixed: 'right' as const },
];

// 动态列配置（根据用户选择）
const displayColumns = computed(() => {
  // 根据 selectedColumns 的顺序来返回列配置
  const orderedColumns = selectedColumns.value
    .map(colKey => columnsWithOperation.find(col => col.colKey === colKey))
    .filter(col => col !== undefined);
  
  // 将操作列添加到最后
  const operationCol = columnsWithOperation.find(col => col.colKey === 'operation');
  if (operationCol) {
    orderedColumns.push(operationCol);
  }
  
  return orderedColumns;
});

// 所有可选列配置（不包括操作列）
const allColumns: ColumnOption[] = [
  { colKey: 'moldCode', title: '模具编号', pinned: true },  // ✨ 固定列：不可隐藏/移动
  { colKey: 'moldName', title: '产品名称' },
  { colKey: 'primaryImage', title: '模具图片' },
  { colKey: 'customerName', title: '客户名称' },
  { colKey: 'productionDate', title: '开始生产日期' },
  { colKey: 'lifeCycleStatus', title: '状态' },
  { colKey: 'moldType', title: '模胚' },
  { colKey: 'moldProcess', title: '工艺' },
  { colKey: 'materialType', title: '材质' },
];

// 默认显示的列
const DEFAULT_COLUMNS = ['moldCode', 'moldName', 'primaryImage', 'customerName', 'productionDate', 'lifeCycleStatus', 'moldType', 'moldProcess', 'materialType'];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([]);

// 对话框相关（已废弃，保留以便兼容）
const dialogVisible = ref(false);
const isEdit = ref(false);
const dialogTitle = ref('新增模具');
const formRef = ref();

// 模具编号解锁状态（已废弃，保留以便兼容）
const moldCodeUnlocked = ref(false);

// 模具图片列表（已废弃，保留以便兼容）
const imageList = ref<any[]>([]);

// 查看详情相关
const detailVisible = ref(false);
const detailData = ref<any>({});
const detailImages = ref<any[]>([]);

// 图片预览相关
const imagePreviewVisible = ref(false);
const previewImages = ref<string[]>([]);
const previewIndex = ref(0);

// 字典数据相关
const processStatusList = ref<any[]>([]); // 工序状态
const moldTypeList = ref<any[]>([]); // 模胚情况
const moldProcessList = ref<any[]>([]); // 模具工艺
const materialTypeList = ref<any[]>([]); // 材质类型


const getMoldList = async () => {
  loading.value = true;
  try {
    // 将前端驼峰命名转换为后端下划线命名
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize,
    };
    
    if (formData.moldCode) {
      params.mold_code = formData.moldCode;
    }
    if (formData.moldName) {
      params.mold_name = formData.moldName;
    }
    if (formData.customerId) {
      params.customer_id = formData.customerId;
    }
    if (formData.lifeCycleStatus !== undefined && formData.lifeCycleStatus !== null) {
      params.life_cycle_status = formData.lifeCycleStatus;
    }
    if (formData.moldType !== undefined && formData.moldType !== null) {
      params.mold_type = formData.moldType;
    }
    if (formData.materialType !== undefined && formData.materialType !== null) {
      params.material_type = formData.materialType;
    }
    if (formData.moldProcess !== undefined && formData.moldProcess !== null) {
      params.mold_process = formData.moldProcess;
    }
    if (formData.productionDateRange && formData.productionDateRange.length === 2) {
      if (formData.productionDateRange[0] && formData.productionDateRange[1]) {
        params.start_date = formData.productionDateRange[0];
        params.end_date = formData.productionDateRange[1];
      }
    }
    
    const res = await getMoldListApi(params);
    
    // 响应拦截器已经解包了 data，直接访问 res.list
    // 将后端下划线命名转换为前端驼峰命名
    const list = res.list || [];
    
    // 直接使用后端返回的 primary_image_url，无需额外请求
    const listWithImages = list.map((item: any) => {
      let primaryImage = '';
      if (item.primary_image_url) {
        // 图片URL直接使用，因为图片代理是通过 /uploads 路径处理的
        primaryImage = item.primary_image_url;
      }
      
      return {
        id: item.id,
        moldCode: item.mold_code,
        moldName: item.mold_name,
        primaryImage,
        images: [] as any[], // 存储所有图片，用于预览
        customerName: item.customer?.cus_name, // 从关联对象中获取 cus_name
        customer: { id: item.customer_id, cus_name: item.customer?.cus_name },
        productionDate: item.production_date,
        lifeCycleStatus: item.life_cycle_status,
        moldType: item.mold_type,
        moldProcess: item.mold_process,
        materialType: item.material_type,
        remark: item.remark,
        createTime: item.create_time,
        // 添加当前工序信息
        current_processes: item.current_processes || [],
        process_summary: item.process_summary || null,
      };
    });
    
    tableData.value = listWithImages;
    pagination.total = res.total || 0;
    
    await getStatistics();
  } catch (error) {
    MessagePlugin.error('获取模具列表失败');
  } finally {
    loading.value = false;
  }
};

const getStatistics = async () => {
  try {
    const res = await getMoldStatistics();
    // 响应拦截器已经解包了 data，直接使用 res
    statistics.total = res.total || 0;
    statistics.processing = res.processing || 0;
    statistics.completed = res.completed || 0;
    statistics.repairing = res.repairing || 0;
  } catch (error) {
    console.error('获取统计信息失败:', error);
  }
};

// 刷新
const handleRefresh = () => {
  getMoldList();
  MessagePlugin.success('刷新成功');
};

// 导出
const handleExport = () => {
  const exportData: Record<string, any>[] = tableData.value.map(item => ({
    '模具编号': item.moldCode,
    '产品名称': item.moldName || '-',
    '客户': item.customerName || '-',
    '开始生产日期': item.productionDate?.split('T')[0] || '-',
    '状态': getLifeCycleStatusLabel(item.lifeCycleStatus),
    '模胚': getMoldTypeText(item.moldType),
    '工艺': getMoldProcessText(item.moldProcess),
    '材质': getMaterialTypeText(item.materialType),
  }));
  
  const headers: string[] = Object.keys(exportData[0] || {});
  const csvContent = [
    headers.join(','),
    ...exportData.map(row => 
      headers.map(header => `"${(row as any)[header] || ''}"`).join(',')
    )
  ].join('\n');
  
  const blob = new Blob(['\ufeff' + csvContent], { type: 'text/csv;charset=utf-8;' });
  const link = document.createElement('a');
  const url = URL.createObjectURL(blob);
  link.setAttribute('href', url);
  link.setAttribute('download', `模具列表_${new Date().getTime()}.csv`);
  link.style.visibility = 'hidden';
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  
  MessagePlugin.success('导出成功');
};

/** 「更多」菜单项：未报废时含「报废结档」 */
const getMoreDropdownOptions = (row: any) => {
  const base: { content: string; value: string }[] = [
    { content: '使用历史', value: 'usage' },
    { content: '工单记录', value: 'workorder' },
    { content: '外发记录', value: 'outsourcing' },
  ];
  if (row.lifeCycleStatus !== 'SCRAPPED') {
    base.push({ content: '报废结档', value: 'scrap' });
  }
  return base;
};

// 更多操作（Dropdown 可能传 value 字符串或 { value }）
const handleMoreAction = (data: any, row: any) => {
  const value = typeof data === 'string' ? data : data?.value;
  switch (value) {
    case 'usage':
      router.push({ name: 'MoldUsageHistoryDetail', params: { moldId: String(row.id) } });
      break;
    case 'workorder':
      router.push({ name: 'MoldWorkOrder', query: { moldId: row.id } });
      break;
    case 'outsourcing':
      router.push({ name: 'OutsourcingManagement', query: { moldId: row.id } });
      break;
    case 'scrap':
      handleScrap(row);
      break;
  }
};

const getLifeCycleStatusRenderConfig = (status: string) => {
  const item = processStatusList.value.find((item: any) => String(item.dict_value) === String(status));
  const color = item?.color || 'default';
  
  return getDictColorRenderConfig(color);
};

const getLifeCycleStatusLabel = (status: string) => {
  const item = processStatusList.value.find((item: any) => String(item.dict_value) === String(status));
  return item?.dict_label || '未知';
};

const getLifeCycleStatusTooltip = (row: any): string => {
  const status = row.lifeCycleStatus;
  
  // 根据不同的状态类型返回不同的提示
  switch (status) {
    case 'DESIGNING': // 设计中
    case 'PROCESSING': // 加工中
      // 显示当前正在进行的工序
      if (row.current_processes && row.current_processes.length > 0) {
        const currentProcess = row.current_processes.find((p: any) => !p.is_next);
        if (currentProcess) {
          return `当前在 ${currentProcess.process_name} 中`;
        }
        // 如果没有进行中的，显示下一个待开始的工序
        const nextProcess = row.current_processes.find((p: any) => p.is_next);
        if (nextProcess) {
          return `下一步: ${nextProcess.process_name}`;
        }
      }
      return getLifeCycleStatusLabel(status);
      
    case 'TESTING': // 待试模
      return '工序已完成，等待试模';
      
    case 'TESTED': // 已试模
      return '试模合格，等待入库';
      
    case 'COMPLETED': // 已完成
      return '工序已完成，准备试模或入库';
      
    case 'REPAIRING': // 维修
      return '模具维修中';
      
    case 'IN_STOCK': // 仓库库存
      return '当前在仓库中';
      
    case 'DELIVERED': // 已交付客户
      return `已交付给客户${row.customerName || ''}`;
      
    case 'SCRAPPED': // 报废
      return '模具已报废';
      
    default:
      return getLifeCycleStatusLabel(status);
  }
};

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  getMoldList();
};

const handleReset = () => {
  formData.moldCode = '';
  formData.moldName = '';
  formData.customerId = undefined;
  formData.lifeCycleStatus = undefined;
  formData.moldType = undefined;
  formData.moldProcess = undefined;
  formData.materialType = undefined;
  formData.productionDateRange = [];
  getMoldList();
};

// 统计卡片筛选
const handleStatFilter = (status: string) => {
  formData.moldCode = '';
  formData.moldName = '';
  formData.customerId = undefined;
  formData.moldType = undefined;
  formData.moldProcess = undefined;
  formData.materialType = undefined;
  formData.productionDateRange = [];
  if (status === 'all') {
    formData.lifeCycleStatus = undefined;
  } else if (status === 'completed') {
    formData.lifeCycleStatus = 'COMPLETED,IN_STOCK,DELIVERED';
  } else {
    formData.lifeCycleStatus = status;
  }
  handleSearch();
};

// 新增 - 跳转到独立表单页面
const handleAdd = () => {
  router.push({ name: 'MoldForm' });
};

// 编辑 - 跳转到独立表单页面
const handleEdit = (row: any) => {
  router.push({ name: 'MoldForm', query: { id: row.id } });
};

// 查看模具详情 - 跳转到独立页面
const handleView = (row: any) => {
  if (!row.id) {
    MessagePlugin.error('模具ID不存在');
    return;
  }
  router.push({ name: 'MoldDetail', params: { id: String(row.id) } });
};

// 预览列表中的图片
const handlePreviewImage = async (row: any) => {
  if (!row.primaryImage) {
    MessagePlugin.warning('暂无图片');
    return;
  }
  
  try {
    const { getMoldImages } = await import('@/api/mold/image');
    const res = await getMoldImages(row.id);
    if (res && res.length > 0) {
      previewImages.value = res.map((img: any) => getFileFullUrl(img.url));
      
      // 找到当前点击图片的索引
      const currentIndex = previewImages.value.findIndex(url => url === row.primaryImage);
      previewIndex.value = currentIndex >= 0 ? currentIndex : 0;
      
      imagePreviewVisible.value = true;
    }
  } catch (error) {
    console.error('加载图片失败:', error);
    MessagePlugin.error('加载图片失败');
  }
};

// 获取状态文本（用于详情展示）
const getLifeCycleStatusText = (status: string) => {
  const item = processStatusList.value.find((item: any) => String(item.dict_value) === String(status));
  return item?.dict_label || '未知';
};

const getMoldTypeText = (type: string) => {
  const item = moldTypeList.value.find((d: any) => d.dict_value === type);
  return item ? item.dict_label : (type === 'NEW' ? '全新' : (type === 'USED' ? '二手' : '-'));
};

const getMoldProcessText = (process: string) => {
  const item = moldProcessList.value.find((d: any) => d.dict_value === process);
  return item ? item.dict_label : (process === 'NORMAL' ? '普通' : (process === 'HOT_RUNNER' ? '热流道' : '-'));
};

const getMaterialTypeText = (material: string) => {
  const item = materialTypeList.value.find((d: any) => d.dict_value === material);
  return item ? item.dict_label : (material === 'STEEL' ? '钢材' : (material === 'HARDENED_STEEL' ? '加硬钢材' : '-'));
};

/** 报废结档：生命周期置为 SCRAPPED，不做物理删除 */
const handleScrap = (row: any) => {
  if (!row?.id) {
    MessagePlugin.error('模具信息不完整');
    return;
  }
  if (row.lifeCycleStatus === 'SCRAPPED') {
    MessagePlugin.info('该模具已报废');
    return;
  }
  const dialog = DialogPlugin.confirm({
    header: '报废结档',
    body: `确认将模具「${row.moldCode || row.mold_code}」标记为报废？\n结档后档案仍保留，便于追溯；如需修改名称等信息可在编辑页调整。`,
    theme: 'warning',
    confirmBtn: '确认报废',
    cancelBtn: '取消',
    onConfirm: async () => {
      try {
        await updateMold({
          id: row.id,
          life_cycle_status: 'SCRAPPED',
        });
        MessagePlugin.success('已标记为报废');
        dialog.destroy();
        getMoldList();
      } catch (error: any) {
        MessagePlugin.error(error?.message || error?.msg || '操作失败');
        dialog.destroy();
      }
    },
  });
};

// 分页变化
const rehandlePageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  getMoldList();
};

const loadDictData = async () => {
  try {
    const res = await getDictTypes();
    const dictList = res.list || [];
    
    // 根据字典编码分类存储
    dictList.forEach((dict: any) => {
      switch (dict.dict_code) {
        case 'mold-status':  // 模具状态（工序状态）
          processStatusList.value = dict.data_list || [];
          break;
        case 'mold-type':  // 模胚情况
          moldTypeList.value = dict.data_list || [];
          break;
        case 'mold-process':  // 模具工艺
          moldProcessList.value = dict.data_list || [];
          break;
        case 'material-type':  // 材质类型
          materialTypeList.value = dict.data_list || [];
          break;
      }
    });
  } catch (error) {
    console.error('加载字典数据失败:', error);
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

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'moldCode', placeholder: '请输入模具编号' },
  { type: 'input', field: 'moldName', placeholder: '请输入产品名称' },
  { type: 'select', field: 'customerId', placeholder: '请选择客户', options: customerList, labelKey: 'cus_name', valueKey: 'id' },
  { type: 'select', field: 'lifeCycleStatus', placeholder: '请选择状态', options: processStatusList, labelKey: 'dict_label', valueKey: 'dict_value' },
  { type: 'select', field: 'moldType', placeholder: '请选择模胚类型', options: moldTypeList, labelKey: 'dict_label', valueKey: 'dict_value' },
  { type: 'select', field: 'materialType', placeholder: '请选择材质', options: materialTypeList, labelKey: 'dict_label', valueKey: 'dict_value' },
  { type: 'select', field: 'moldProcess', placeholder: '请选择模具工艺', options: moldProcessList, labelKey: 'dict_label', valueKey: 'dict_value' },
  { type: 'dateRangePicker', field: 'productionDateRange', span: 8 },
];

onMounted(async () => {
  const params = new URLSearchParams(window.location.search);
  const moldCode = params.get('moldCode');
  
  if (moldCode) {
    // 将模具编号填充到搜索框
    formData.moldCode = moldCode;
  }
  
  await getCustomerList();
  await loadDictData();
  
  try {
    const saved = localStorage.getItem('mold_table_columns');
    if (saved) {
      const savedColumns = JSON.parse(saved);
      if (Array.isArray(savedColumns)) {
        const validColumns = savedColumns.filter((colKey: string) => 
          allColumns.some(col => col.colKey === colKey)
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
  
  // 如果有URL参数，自动触发搜索
  getMoldList();
});

// keep-alive 激活时自动刷新列表和统计数据（如从表单页返回）
onActivated(() => {
  getMoldList();
  getStatistics();
});
</script>

<style scoped>
.mold-container {
  padding: 16px;
}

.search-card,
.toolbar-card {
  margin-bottom: 16px;
}

/* 统计信息卡片样式 */
.statistics-row {
  margin-bottom: 16px;
}

.stat-card {
  transition: all 0.3s ease;
  cursor: pointer;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #fff;
}

.stat-icon-primary {
  background: linear-gradient(135deg, #0052d9 0%, #266fe8 100%);
}

.stat-icon-warning {
  background: linear-gradient(135deg, #e37318 0%, #ed7b2f 100%);
}

.stat-icon-success {
  background: linear-gradient(135deg, #008858 0%, #059465 100%);
}

.stat-icon-danger {
  background: linear-gradient(135deg, #ad352f 0%, #c63a34 100%);
}

.stat-info {
  flex: 1;
}

.stat-label {
  font-size: 14px;
  color: var(--td-text-color-secondary, rgba(0, 0, 0, 0.6));
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: var(--td-text-color-primary, rgba(0, 0, 0, 0.9));
  line-height: 1;
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

/* 锁定状态下的输入框样式 */
.locked-input :deep(.t-input__inner) {
  background-color: var(--td-bg-color-container, #f5f5f5);
  cursor: not-allowed;
  color: var(--td-text-color-disabled, rgba(0, 0, 0, 0.36));
}

/* 模具编号链接样式 */
.mold-code-link {
  font-weight: 600;
  font-size: 15px;
}

/* 图片容器样式 - 已废弃，保留兼容 */
.image-wrapper {
  width: 70px;
  height: 70px;
  cursor: pointer;
  display: flex;
  justify-content: center;
  align-items: center;
  border-radius: 4px;
  overflow: hidden;
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

/* 暂无图片占位符样式 - 已废弃，保留兼容 */
.image-no-image {
  width: 70px;
  height: 70px;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>

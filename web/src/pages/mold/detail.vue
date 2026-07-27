<template>
  <div class="mold-detail-page">
    <!-- 加载中状态 -->
    <div v-if="loading" class="loading-container">
      <PageLoading />
    </div>
    
    <!-- 数据加载完成后显示 -->
    <template v-else>
    <!-- 返回按钮 -->
    <div class="back-button-container">
      <t-button theme="default" @click="goBack" variant="text">
        <template #icon><t-icon name="chevron-left" /></template>
        返回
      </t-button>
    </div>

    <!-- 基本信息 -->
    <t-card :bordered="false" title="模具信息" class="mold-info-card">
        <t-descriptions :column="3" bordered class="mold-descriptions">
          <t-descriptions-item label="模具编号">
            <span class="mold-code-text">{{ moldDetail.mold_code || '-' }}</span>
          </t-descriptions-item>
          <t-descriptions-item label="产品名称">{{ moldDetail.mold_name || '-' }}</t-descriptions-item>
          <t-descriptions-item label="客户">{{ moldDetail.customer?.cus_name || '-' }}</t-descriptions-item>
          <t-descriptions-item label="开始生产日期">{{ formatDate(moldDetail.production_date) }}</t-descriptions-item>
          <t-descriptions-item label="状态">
            <t-tag
              v-if="getLifeCycleStatusRenderConfig(moldDetail.life_cycle_status).mode === 'theme'"
              :theme="getLifeCycleStatusRenderConfig(moldDetail.life_cycle_status).theme"
            >
              {{ getLifeCycleStatusLabel(moldDetail.life_cycle_status) }}
            </t-tag>
            <t-tag
              v-else
              :style="getLifeCycleStatusRenderConfig(moldDetail.life_cycle_status).style"
            >
              {{ getLifeCycleStatusLabel(moldDetail.life_cycle_status) }}
            </t-tag>
          </t-descriptions-item>
          <t-descriptions-item label="模胚情况">{{ getMoldTypeText(moldDetail.mold_type) }}</t-descriptions-item>
          <t-descriptions-item label="模具工艺">{{ getMoldProcessText(moldDetail.mold_process) }}</t-descriptions-item>
          <t-descriptions-item label="材质">{{ getMaterialTypeText(moldDetail.material_type) }}</t-descriptions-item>
          <t-descriptions-item label="模腔数">{{ moldDetail.cavity_count ? `${moldDetail.cavity_count} 腔` : '-' }}</t-descriptions-item>
          <t-descriptions-item label="创建时间">{{ formatDateTime(moldDetail.create_time) }}</t-descriptions-item>
          <t-descriptions-item label="备注" :span="3">
            <div style="white-space: pre-wrap;">{{ moldDetail.remark || '-' }}</div>
          </t-descriptions-item>
        </t-descriptions>
    </t-card>

    <!-- 制造进度 -->
    <t-card title="制造进度" class="container-base-margin-top" :bordered="false">
      <!-- 历史模具提示 -->
      <t-alert
        v-if="isHistoricalMold"
        theme="info"
        message="此模具为历史模具，无关联工单和订单记录，无法显示完整的制造进度时间线"
        :close="false"
        style="margin-bottom: 16px"
      />
      
      <t-steps :current="progressCurrent" layout="horizontal" class="progress-steps">
        <t-step-item
          v-for="(step, index) in manufacturingSteps"
          :key="index"
          :title="step.title"
          :content="step.content"
          :class="{ 
            'step-clickable': !isHistoricalMold && index === 2 && ['TESTED', 'COMPLETED', 'REPAIRING', 'IN_STOCK', 'DELIVERED', 'SCRAPPED'].includes(String(moldDetail.life_cycle_status)),
            'step-disabled': isHistoricalMold && !step.content || step.content === '-' 
          }"
        />
      </t-steps>
      
    </t-card>

    <!-- 试模记录弹窗 -->
    <t-dialog
      v-model:visible="tryTestDialogVisible"
      header="模具试模记录"
      width="800px"
      :confirm-btn="null"
      :cancel-btn="null"
     :close-on-overlay-click="false">
      <t-table
        :data="tryTestRecords"
        :columns="tryTestColumns"
        row-key="id"
        :pagination="{ pageSize: 10 }"
        :stripe="false"
      >
        <template #index="{ rowIndex }">
          {{ rowIndex + 1 }}
        </template>
        <template #result="{ row }">
          <t-tag :theme="row.result === '合格' ? 'success' : 'danger'">
            {{ row.result || '-' }}
          </t-tag>
        </template>
        <template #test_time="{ row }">
          {{ formatDate(row.test_time) }}
        </template>
      </t-table>
      <template v-if="tryTestRecords.length === 0">
        <t-empty description="暂无试模记录" />
      </template>
    </t-dialog>

    <!-- 模具图片 -->
    <t-card title="模具图片" class="container-base-margin-top" :bordered="false">
        <div v-if="moldImages.length > 0" class="image-gallery">
          <t-image-viewer 
            v-for="(img, index) in previewImages" 
            :key="img" 
            :default-index="index" 
            :images="previewImages" 
            :z-index="10000"
          >
            <template #trigger="{ open }">
              <div class="image-item-wrapper">
                <t-image
                  :src="img"
                  fit="cover"
                  :style="{ width: '100%', height: '100%' }"
                  @click="open(index)"
                />
                <div class="image-overlay" @click="open(index)">
                  <span><BrowseIcon size="1.4em" /> 预览</span>
                </div>
                <div v-if="index === 0" class="primary-badge">主图</div>
              </div>
            </template>
          </t-image-viewer>
        </div>
        <div v-else class="image-gallery-placeholder">
          <t-empty title="暂无模具图片" />
        </div>
    </t-card>

    <!-- 模具生产明细 -->
    <t-card title="模具生产明细" class="container-base-margin-top" :bordered="false">
        <!-- 历史模具提示 -->
        <t-alert
          v-if="isHistoricalMold"
          theme="warning"
          message="此模具为历史模具，无关联工单记录，无法显示工序和生产明细数据"
          :close="false"
          style="margin-bottom: 16px"
        />
        
        <t-tabs v-if="!isHistoricalMold" defaultValue="manufacturing">
          <t-tab-panel value="manufacturing" label="模具制造明细">
            <div style="min-height: 400px; padding-top: 16px;">
              <t-table
              :data="manufacturingData"
              :columns="manufacturingColumns"
              row-key="id"
              :pagination="{ 
                pageSize: 10,
                total: manufacturingData.length
              }"
              :stripe="false"
            >
              <template #status_text="{ row }">
                <t-tag :theme="row.status === 1 ? 'success' : row.status === 2 ? 'default' : 'warning'">
                  {{ row.status_text }}
                </t-tag>
              </template>
            </t-table>
            </div>
          </t-tab-panel>
          
          <t-tab-panel value="production" label="产品生产明细">
            <div style="min-height: 400px; padding-top: 16px;">
            <t-table
              :data="productionData"
              :columns="productionColumns"
              row-key="id"
              :pagination="{ 
                pageSize: 10,
                total: productionData.length
              }"
              :stripe="false"
            >
              <template #status="{ row }">
                <t-tag :theme="row.status === 'COMPLETED' || row.status === 'STOCKED_IN' ? 'success' : row.status === 'PRODUCING' ? 'primary' : 'default'">
                  {{ getStatusText(row.status) }}
                </t-tag>
              </template>
            </t-table>
            <t-empty v-if="productionData.length === 0" description="暂无产品生产记录" />
            </div>
          </t-tab-panel>
        </t-tabs>
        
        <!-- 历史模具的空数据展示 -->
        <t-empty v-if="isHistoricalMold" description="无工序和生产记录" />
    </t-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import { getMoldById, getMoldTryTests, getWorkOrderProcesses, getMoldProductionItems } from '@/api/mold';
import { getMoldImages, type MoldFile } from '@/api/mold/image';
import { getFileFullUrl } from '@/api/common/upload';
import { getDictTypes } from '@/api/reference';
import { getDictColorRenderConfig } from '@/utils/dict-color';
import { BrowseIcon } from 'tdesign-icons-vue-next';
import PageLoading from '@/components/PageLoading/index.vue';

defineOptions({
  name: 'MoldDetail',
});

const route = useRoute();
const router = useRouter();

// 模具图片列表
const moldImages = ref<MoldFile[]>([]);
const previewImages = ref<string[]>([]);

// 模具详情数据
const moldDetail = ref<any>({});
const loading = ref(true);

// 字典数据
const processStatusList = ref<any[]>([]);
const moldTypeList = ref<any[]>([]);
const moldProcessList = ref<any[]>([]);
const materialTypeList = ref<any[]>([]);

// 试模记录相关
const tryTestDialogVisible = ref(false);
const tryTestRecords = ref<any[]>([]);

const tryTestColumns = [
  { colKey: 'index', title: '序号', width: 80 },
  { colKey: 'test_time', title: '试模时间', width: 180 },
  { colKey: 'result', title: '结果', width: 100 },
  { colKey: 'test_mode', title: '试模方式', width: 120 },
  { colKey: 'remark', title: '备注', ellipsis: true },
];

// 制造进度步骤
const manufacturingSteps = ref([
  { title: '订单创建', content: '-' },
  { title: '制造完成', content: '-' },
  { title: '试模', content: '-' },
  { title: '入库', content: '-' },
]);

const progressCurrent = computed(() => {
  const status = String(moldDetail.value.life_cycle_status);
  // 对于历史模具，只显示基本进度
  if (isHistoricalMold.value) {
    // 根据模具状态显示大致进度
    if (['IN_STOCK', 'DELIVERED', 'SCRAPPED'].includes(status)) return 3; // 已完成所有阶段
    if (['TESTED', 'COMPLETED', 'REPAIRING'].includes(status)) return 2; // 已试模及以上
    if (status === 'TESTING') return 1; // 待试模及以上
    return 0; // 其他状态
  }
  
  // TDesign Steps 的 current 从0开始，表示当前进行到第几步（4个步骤）
  if (['IN_STOCK', 'DELIVERED', 'SCRAPPED'].includes(status)) return 4; // 仓库库存/已交付/报废 - 所有4个步骤都已完成
  if (['COMPLETED', 'REPAIRING'].includes(status)) return 3; // 已完成/维修 - 前3步完成，到第4步（入库）
  if (status === 'TESTED') return 2; // 已试模 - 前2步完成，到第3步（试模）
  if (status === 'TESTING') return 1; // 待试模 - 第1步完成，到第2步（制造完成）
  return 0; // 未开始/加工中 - 第1步（订单创建）进行中
});

// 判断是否为历史模具（无工单/订单关联）
const isHistoricalMold = computed(() => {
  // 两者都有或只有工单 → 活跃模具
  // 只有订单 → 活跃模具
  // 两者都没有 → 历史模具
  const hasOrder = moldDetail.value.latest_order_id;
  const hasWorkOrder = moldDetail.value.latest_work_order_id;
  
  // 只要有订单或工单中的任何一个，就不是历史模具
  if (hasOrder || hasWorkOrder) {
    return false;
  }
  
  return true;
});

// 模具制造明细（从API获取真实数据）
const manufacturingColumns = [
  { colKey: 'index', title: '序号', width: 60 },
  { colKey: 'process_name', title: '工序名称', width: 100 },
  { colKey: 'group_name', title: '工序分组', width: 100 },
  { colKey: 'status_text', title: '状态', width: 80 },
  { colKey: 'completed_at', title: '完成时间', width: 170 },
];

const manufacturingData = ref<any[]>([]);

// 产品生产明细（从API获取真实数据）
const productionColumns = [
  { colKey: 'index', title: '序号', width: 80 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'product_name', title: '产品名称', width: 150 },
  { colKey: 'product_code', title: '产品编号', width: 150 },
  { colKey: 'order_quantity', title: '订单数量', width: 100 },
  { colKey: 'completed_quantity', title: '完成数量', width: 100 },
  { colKey: 'plan_start', title: '计划开始', width: 160 },
  { colKey: 'actual_start', title: '实际开始', width: 160 },
  { colKey: 'plan_end', title: '计划结束', width: 160 },
  { colKey: 'actual_end', title: '实际结束', width: 160 },
  { colKey: 'remark', title: '备注', ellipsis: true },
];

const productionData = ref<any[]>([]);

// 格式化日期（只显示 YYYY-MM-DD）
import { formatDate, formatDateTime } from '@/utils/date';

const getLifeCycleStatusRenderConfig = (status: string) => {
  const item = processStatusList.value.find((d) => String(d.dict_value) === String(status));
  const color = item?.color || 'default';
  return getDictColorRenderConfig(color);
};

const getLifeCycleStatusLabel = (status: string) => {
  const item = processStatusList.value.find((d) => String(d.dict_value) === String(status));
  return item ? item.dict_label : '-';
};

const getMoldTypeText = (type: string) => {
  const item = moldTypeList.value.find((d) => d.dict_value === type);
  return item ? item.dict_label : '-';
};

const getMoldProcessText = (process: string) => {
  const item = moldProcessList.value.find((d) => d.dict_value === process);
  return item ? item.dict_label : '-';
};

const getMaterialTypeText = (type: string) => {
  const item = materialTypeList.value.find((d) => d.dict_value === type);
  return item ? item.dict_label : '-';
};

const goBack = () => {
  router.back();
};

// 步骤点击事件
const handleStepClick = (index: number) => {
  // 历史模具的工序不可点击
  if (isHistoricalMold.value) {
    return;
  }
  // 只有试模步骤(index=2)且状态为已试模及以上可以点击
  if (index === 2 && ['TESTED', 'COMPLETED', 'REPAIRING', 'IN_STOCK', 'DELIVERED', 'SCRAPPED'].includes(String(moldDetail.value.life_cycle_status))) {
    loadTryTestRecords();
  }
};

const loadTryTestRecords = async () => {
  try {
    tryTestDialogVisible.value = true;
    const res = await getMoldTryTests({ mold_id: moldDetail.value.id });
    tryTestRecords.value = res?.list || [];
  } catch (err: any) {
    console.error('加载试模记录失败:', err);
    MessagePlugin.error('加载试模记录失败');
  }
};

const loadMoldImages = async () => {
  const id = route.params.id as string;
  if (!id) {
    return;
  }

  try {
    const images = await getMoldImages(Number(id));
    // 确保主图排在第一位（双重保险）
    const sortedImages = images ? [...images] : [];
    const primaryIndex = sortedImages.findIndex(img => img.is_primary === 1);
    if (primaryIndex > 0) {
      const primaryImage = sortedImages.splice(primaryIndex, 1)[0];
      sortedImages.unshift(primaryImage);
    }
    
    moldImages.value = sortedImages;
    previewImages.value = moldImages.value.map((img) => getFileFullUrl(img.url));
  } catch (error) {
    console.error('加载模具图片失败:', error);
  }
};

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    'PENDING_SCHEDULE': '待排产',
    'PRODUCING': '生产中',
    'PAUSED': '已暂停',
    'STOCKED_IN': '已入库',
    'COMPLETED': '已完成',
    'CANCELLED': '已取消'
  };
  return statusMap[status] || status || '-';
};

const loadProductionData = async () => {
  // 如果是历史模具，不加载生产数据
  if (isHistoricalMold.value || !moldDetail.value.id) {
    productionData.value = [];
    return;
  }

  try {
    const moldId = moldDetail.value.id;
    const res = await getMoldProductionItems(moldId);
    
    // 后端直接返回数组
    const items = Array.isArray(res) ? res : [];
    
    // 格式化数据，添加序号
    productionData.value = items.map((item: any, index: number) => {
      // 优先使用明细项自己的字段，其次使用订单字段
      return {
        id: item.id,
        index: index + 1,
        product_name: item.product_name || '-',
        // 使用产品的专属编号，而不是订单编号
        product_code: item.product?.pro_code || '-',
        order_quantity: item.plan_qty || 0,
        completed_quantity: item.completed_qty || 0,
        // 计划日期：优先使用明细项，其次使用订单
        plan_start: item.order?.plan_start_date ? formatDate(item.order.plan_start_date) : '-',
        // 实际日期：使用明细项的实际完成时间，或者订单的实际开始/结束时间
        actual_start: item.order?.actual_start_date ? formatDate(item.order.actual_start_date) : '-',
        plan_end: item.order?.plan_end_date ? formatDate(item.order.plan_end_date) : '-',
        // 实际结束时间：优先使用明细项的实际完成时间，其次使用订单的实际结束时间
        actual_end: item.actual_completion_time ? formatDate(item.actual_completion_time) : (item.order?.actual_end_date ? formatDate(item.order.actual_end_date) : '-'),
        // 状态：优先使用明细项自己的状态
        status: item.status || item.order?.status || 'PENDING_SCHEDULE',
        remark: item.remark || item.order?.remark || '',
      };
    });
  } catch (error) {
    console.error('加载产品生产明细失败:', error);
    // 不显示错误提示，允许为空
  }
};

const loadManufacturingData = async () => {
  // 如果是历史模具，不加载工序数据
  if (isHistoricalMold.value || !moldDetail.value.latest_work_order_id) {
    manufacturingData.value = [];
    return;
  }

  try {
    const workOrderId = moldDetail.value.latest_work_order_id;
    const res = await getWorkOrderProcesses(workOrderId);
    
    // 后端直接返回数组，不是 {list: [...]} 格式
    const items = Array.isArray(res) ? res : (res?.list || []);
    
    // 格式化数据，添加序号和关联信息
    manufacturingData.value = items.map((item: any, index: number) => {
      // 状态映射
      let statusText = '未完成';
      if (item.is_completed === 1) {
        statusText = '已完成';
      } else if (item.is_completed === 2) {
        statusText = '已跳过';
      }
      
      return {
        id: item.id,
        index: index + 1,
        process_name: item.process_template?.process_name || '-',
        group_name: item.process_template?.group_name || '-',
        status: item.is_completed,
        status_text: statusText,
        completed_at: item.completed_at ? formatDateTime(item.completed_at) : '-',
        remark: item.remark || '',
      };
    });
  } catch (error) {
    console.error('加载制造明细失败:', error);
    // 不显示错误提示，允许为空
  }
};

const loadMoldDetail = async () => {
  loading.value = true;
  const id = route.params.id as string;
  if (!id) {
    MessagePlugin.error('模具ID不能为空');
    goBack();
    return;
  }

  try {
    const res = await getMoldById(Number(id));
    moldDetail.value = res;
    
    // 使用后端返回的时间线数据更新制造进度步骤
    const timeline = res.manufacturing_timeline || {};
    
    // 步骤1: 订单创建
    if (timeline.order_create) {
      manufacturingSteps.value[0].content = `创建时间：${formatDate(timeline.order_create)}`;
    }
    
    // 步骤2: 制造完成
    if (timeline.manufacturing_complete) {
      manufacturingSteps.value[1].content = `完成时间：${formatDate(timeline.manufacturing_complete)}`;
    }
    
    // 步骤3: 试模
    if (timeline.try_test) {
      manufacturingSteps.value[2].content = `试模时间：${formatDate(timeline.try_test)}`;
    }
    
    // 步骤4: 入库
    if (timeline.storage) {
      manufacturingSteps.value[3].content = `入库时间：${formatDate(timeline.storage)}`;
    }
    
    await loadMoldImages();
    
    await loadManufacturingData();
    
    await loadProductionData();
  } catch (error) {
    console.error('加载模具详情失败:', error);
    MessagePlugin.error('获取模具详情失败');
    goBack();
  } finally {
    loading.value = false;
  }
};

const loadDictData = async () => {
  try {
    const res = await getDictTypes();
    const dictList = res.list || [];
    
    dictList.forEach((dict: any) => {
      switch (dict.dict_code) {
        case 'mold-status':
          processStatusList.value = dict.data_list || [];
          break;
        case 'mold-type':
          moldTypeList.value = dict.data_list || [];
          break;
        case 'mold-process':
          moldProcessList.value = dict.data_list || [];
          break;
        case 'material-type':
          materialTypeList.value = dict.data_list || [];
          break;
      }
    });
  } catch (error) {
    console.error('加载字典数据失败:', error);
  }
};

onMounted(() => {
  loadMoldDetail();
  loadDictData();
});

watch(
  () => route.params.id,
  (newId) => {
    if (newId && route.name === 'MoldDetail') {
      loadMoldDetail();
      loadDictData();
    }
  },
);
</script>

<style lang="less" scoped>
.mold-detail-page {
  padding: 16px;
  background-color: var(--td-bg-color-page, #f5f5f5);
  min-height: 100vh;
}

.loading-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 60vh;
}

.back-button-container {
  margin-bottom: 16px;
}

.container-base-margin-top {
  margin-top: 16px;
}

.image-gallery {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.image-item-wrapper {
  position: relative;
  width: 200px;
  height: 200px;
  border: 4px solid var(--td-bg-color-secondarycontainer);
  border-radius: var(--td-radius-medium);
  overflow: hidden;
  background-color: var(--td-bg-color-secondarycontainer);
  cursor: pointer;
  transition: all 0.3s ease;
}

.image-item-wrapper:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  transform: translateY(-2px);
}

.image-overlay {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  opacity: 0;
  background-color: rgba(0, 0, 0, 0.6);
  color: var(--td-text-color-anti);
  font-size: 14px;
  transition: opacity 0.3s ease;
  pointer-events: none;
}

.image-item-wrapper:hover .image-overlay {
  opacity: 1;
}

.primary-badge {
  position: absolute;
  top: 8px;
  left: 8px;
  padding: 2px 8px;
  background-color: var(--td-brand-color);
  color: white;
  font-size: var(--td-font-size-body-small);
  border-radius: var(--td-radius-round);
  z-index: 2;
  pointer-events: none;
}

.t-descriptions {
  :deep(.t-descriptions__body) {
    background-color: transparent;
  }
}

/* 试模步骤可点击样式 */
.progress-steps {
  padding-bottom: 16px;
}

.step-clickable {
  cursor: pointer;
  :deep(.t-steps-item__title) {
    color: var(--td-brand-color, #0052d9) !important;
    text-decoration: underline;
  }
  :deep(.t-steps-item__content) {
    color: var(--td-text-color-primary, rgba(0, 0, 0, 0.9)) !important;
  }
}
.step-clickable:hover {
  :deep(.t-steps-item__title) {
    opacity: 0.8;
  }
}

/* 步骤禁用样式(历史模具) */
.step-disabled {
  :deep(.t-steps-item__title) {
    color: var(--td-text-color-disabled, #999) !important;
  }
  :deep(.t-steps-item__content) {
    color: var(--td-text-color-disabled, #999) !important;
  }
}

/* 模具编号样式 */
.mold-code-text {
  font-size: 16px;
  font-weight: 600;
  color: var(--td-text-color-primary, rgba(0, 0, 0, 0.9));
}

</style>

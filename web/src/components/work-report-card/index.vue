<template>
  <t-card theme="poster2" :bordered="false" :class="{ 'external-production': props.order.productionType === 'OUTSOURCED' }">
    <template #avatar>
      <div class="product-image-wrapper">
        <ImagePlaceholder
          :src="props.order.productImageUrl || ''"
          :width="56"
          :height="56"
          fit="cover"
          shape="round"
        />
      </div>
    </template>
    <template #status>
      <t-tag 
        shape="round"
        :theme="props.order.statusInfo?.theme || 'default'" 
        :style="props.order.statusInfo?.themeStyle"
        :disabled="props.order.status === 'COMPLETED'"
      >
        {{ props.order.statusInfo?.label || '未知' }}
      </t-tag>
    </template>
    <template #content>
      <p class="list-card-item_detail--name">
        {{ props.order.orderNo }}
        <!-- 外发标签 -->
        <t-tag
          v-if="props.order.productionType === 'OUTSOURCED'"
          theme="primary"
          variant="light"
          size="small"
          style="margin-left: 8px; vertical-align: middle;"
        >
          外发
        </t-tag>
        <!-- ✨ 等待模具提示 - 小标签 -->
        <t-tag
          v-if="props.order.dependencyStatus === 'WAITING_MOLD'"
          theme="warning"
          variant="light"
          size="small"
          style="margin-left: 8px; vertical-align: middle;"
        >
          <template #icon>
            <t-icon name="time" size="12px" />
          </template>
          等待模具试模
        </t-tag>
      </p>
      
      <!-- 产品名称 - 带tooltip -->
      <t-tooltip v-if="props.order.productNameList && props.order.productNameList.length > 2" :content="props.order.productNameList.join('\n')" placement="top">
        <p class="list-card-item_detail--desc list-card-item_detail--desc-clamped">
          {{ props.order.productNameList.slice(0, 2).join('、') }}等{{ props.order.productNameList.length }}个产品
        </p>
      </t-tooltip>
      <p v-else class="list-card-item_detail--desc">{{ props.order.productName || '-' }}</p>

      <!-- 模具编号 - 带tooltip -->
      <t-tooltip v-if="props.order.moldCodeList && props.order.moldCodeList.length > 2" :content="props.order.moldCodeList.join('\n')" placement="top">
        <p class="list-card-item_detail--info list-card-item_detail--info-highlight">
          模具编号：{{ props.order.moldCodeList.slice(0, 2).join('、') }}等{{ props.order.moldCodeList.length }}个模具
        </p>
      </t-tooltip>
      <p v-else class="list-card-item_detail--info list-card-item_detail--info-highlight">模具编号：{{ props.order.moldCode || '-' }}</p>
      
      <div class="progress-section">
        <p class="progress-text">已完成 {{ props.order.completedQuantity || 0 }} / {{ props.order.planQuantity || 0 }}</p>
        <t-progress
          :percentage="progressPercent"
          :color="progressColor"
          :track-color="'#f3f3f3'"
          :label="true"
          size="small"
        />
      </div>
    </template>
    <template #footer>
      <div class="list-card-item_footer-wrapper">
        <div class="list-card-item_footer-info">
          <div class="deadline-label">截止日期：</div>
          <div class="deadline-value">{{ props.order.deadline || '-' }}</div>
        </div>
        <div class="list-card-item_footer-actions">
          <!-- 根据状态显示主要操作按钮 -->
          <t-button 
            v-if="showPrimaryButton" 
            :theme="getPrimaryButtonTheme()" 
            variant="outline"
            size="medium" 
            @click="handlePrimaryAction"
          >
            {{ getPrimaryButtonText() }}
          </t-button>
          <t-dropdown
            trigger="click"
            :options="getActionOptions(props.order)"
          >
            <t-button theme="default" shape="square" variant="text">
              <more-icon />
            </t-button>
          </t-dropdown>
        </div>
      </div>
    </template>
    <template #actions>
    </template>
  </t-card>
</template>
<script setup lang="ts">
import { MoreIcon, QueueIcon } from 'tdesign-icons-vue-next';
import { computed, type PropType } from 'vue';
import { getDictColorHex } from '@/utils/dict-color';
import ImagePlaceholder from '@/components/ImagePlaceholder/index.vue';

export interface WorkReportOrderType {
  id: number;
  itemId?: number;  //  新增：订单明细项ID（用于按明细报工和入库）
  orderNo: string;
  productName: string;
  productNameList?: string[];  // 产品名称列表，用于tooltip展示
  productImageUrl?: string;  // 产品图片URL
  customerName?: string;
  moldCode?: string;
  moldCodeList?: string[];  // 模具编号列表，用于tooltip展示
  planQuantity: number;
  completedQuantity: number;
  status: string;  // 后端状态值（语义化字符串）
  statusInfo: { 
    value: string; // 状态值（语义化字符串）
    label: string; 
    dictValue?: string; // 字典值（语义化字符串）
    theme: 'default' | 'primary' | 'success' | 'warning' | 'danger';
    themeStyle?: {
      background: string;
      color: string;
      borderColor: string;
    };
  };  // 状态字典信息
  deadline: string;
  productionType?: string;  // 生产类型（语义化字符串：IN_HOUSE, OUTSOURCED）
  dependencyStatus?: string;  // ✨ 依赖状态：NONE=无依赖, WAITING_MOLD=等待模具, READY=就绪
  // ✨ 新增：产品ID和模具ID（用于入库）
  productId?: number;
  moldId?: number;
  externalMoldCode?: string;
}

const props = defineProps({
  order: {
    type: Object as PropType<WorkReportOrderType>,
    default: undefined,
  },
  productionTypeDict: {
    type: Array as PropType<Array<{ dict_value: string; dict_label: string; color: string }>>,
    default: () => [] as Array<{ dict_value: string; dict_label: string; color: string }>,
  },
});

const emit = defineEmits(['report-work', 'view-detail', 'start-production', 'stock-in', 'complete-production']);

const getActionOptions = (order: WorkReportOrderType) => {
  const options: any[] = [
    {
      content: '查看详情',
      value: 'view',
      onClick: () => handleClickView(order),
    },
  ];

  // ✨ 检查依赖状态：如果是 WAITING_MOLD，禁用所有操作
  if (order.dependencyStatus === 'WAITING_MOLD') {
    return options; // 只返回查看详情
  }

  // 根据字典 dict_value 判断是否可以报工(转小写比较)
  const statusValue = (order.statusInfo?.dictValue || order.status || '').toLowerCase();
  const canReportWork = ['producing', 'production'].some(s => statusValue.includes(s));
  const canStart = ['pending_schedule', 'pending'].some(s => statusValue.includes(s));

  if (canReportWork) {
    // 已完成数量 >= 计划数量时，显示"继续报工"
    const isOverPlan = order.completedQuantity >= order.planQuantity;
    options.push({
      content: isOverPlan ? '继续报工' : '报工',
      value: 'report',
      onClick: () => handleClickReport(order),
    });
    // 生产中状态显示"结束生产"按钮
    options.push({
      content: '结束生产',
      value: 'complete',
      onClick: () => handleClickComplete(order),
    });
  } else if (canStart) {
    options.push({
      content: '开始生产',
      value: 'start',
      onClick: () => handleClickStart(order),
    });
  }

  return options;
};

const handleClickView = (order: WorkReportOrderType) => {
  emit('view-detail', order);
};

const handleClickReport = (order: WorkReportOrderType) => {
  emit('report-work', order);
};

// 根据生产类型获取头像背景色（低饱和度版本）
const getAvatarBackground = () => {
  // 从字典数据中查找对应的颜色
  const dictItem = props.productionTypeDict?.find(
    (item) => item.dict_value === props.order.productionType
  );
  
  if (dictItem?.color) {
    // 使用字典工具获取颜色的 Hex 值，然后降低饱和度
    const colorHex = getDictColorHex(dictItem.color);
    // 将纯色转换为低饱和度版本（添加透明度）
    return colorHex + '80'; // 60 = 约60%透明度，中等饱和度
  }
  
  // 默认返回品牌蓝的中等饱和度版本
  return '#0052d940';
};

// 进度百分比
const progressPercent = computed(() => {
  if (!props.order.planQuantity || props.order.planQuantity <= 0) return 0;
  const pct = Math.round((props.order.completedQuantity / props.order.planQuantity) * 100);
  return Math.min(pct, 100);
});

// 进度条颜色：超过 100% 用橙色，已完全达标用绿色，否则用蓝色
const progressColor = computed(() => {
  if (progressPercent.value >= 100) return '#2ba471';
  if (progressPercent.value > 80) return '#0052d9';
  return '#0052d9';
});

const canReportWork = computed(() => {
  const statusValue = (props.order.statusInfo?.dictValue || props.order.status || '').toLowerCase();
  return ['producing', 'production'].some(s => statusValue.includes(s));
});

const showPrimaryButton = computed(() => {
  // ✨ 如果依赖状态为 WAITING_MOLD，不显示主要按钮
  if (props.order.dependencyStatus === 'WAITING_MOLD') {
    return false;
  }
  
  const statusValue = (props.order.statusInfo?.dictValue || props.order.status || '').toLowerCase();
  return ['pending_schedule', 'producing'].some(s => statusValue.includes(s));
});

const getPrimaryButtonText = (): string => {
  // ✨ 如果依赖状态为 WAITING_MOLD，返回提示文本
  if (props.order.dependencyStatus === 'WAITING_MOLD') {
    return '等待模具';
  }
  
  const statusValue = (props.order.statusInfo?.dictValue || props.order.status || '').toLowerCase();
  if (['pending_schedule'].some(s => statusValue.includes(s))) {
    return '开始生产';
  } else if (['producing', 'production'].some(s => statusValue.includes(s))) {
    // 已完成数量 >= 计划数量时，显示"继续报工"
    return props.order.completedQuantity >= props.order.planQuantity ? '继续报工' : '报工';
  }
  return '';
};

const getPrimaryButtonTheme = (): 'primary' | 'success' | 'warning' | 'danger' | 'default' => {
  // ✨ 如果依赖状态为 WAITING_MOLD，返回 default（灰色）
  if (props.order.dependencyStatus === 'WAITING_MOLD') {
    return 'default';
  }
  
  const statusValue = (props.order.statusInfo?.dictValue || props.order.status || '').toLowerCase();
  if (['pending_schedule'].some(s => statusValue.includes(s))) {
    return 'primary';
  } else if (['producing', 'production'].some(s => statusValue.includes(s))) {
    return 'warning';
  } else if (['stocked_in'].some(s => statusValue.includes(s))) {
    return 'success';
  }
  return 'default';
};

const handlePrimaryAction = () => {
  // ✨ 如果依赖状态为 WAITING_MOLD，不执行任何操作
  if (props.order.dependencyStatus === 'WAITING_MOLD') {
    return;
  }
  
  const statusValue = (props.order.statusInfo?.dictValue || props.order.status || '').toLowerCase();
  if (['pending_schedule'].some(s => statusValue.includes(s))) {
    // 开始生产
    handleClickStart(props.order);
  } else if (['producing', 'production'].some(s => statusValue.includes(s))) {
    // 报工
    handleClickReport(props.order);
  }
};

// 已移除：图片URL处理逻辑已迁移到 ImagePlaceholder 组件

const handleClickStart = (order: WorkReportOrderType) => {
  emit('start-production', order);
};

const handleClickStockIn = (order: WorkReportOrderType) => {
  emit('stock-in', order);
};

const handleClickComplete = (order: WorkReportOrderType) => {
  emit('complete-production', order);
};
</script>
<style lang="less" scoped>
.list-card-item {
  display: flex;
  flex-direction: column;
  cursor: pointer;
  min-height: 200px;

  &_detail {
    min-height: 140px;

    &--name {
      margin-bottom: var(--td-comp-margin-s);
      font: var(--td-font-body-medium);
      color: var(--td-text-color-primary);
    }

    &--desc {
      color: var(--td-text-color-primary);
      font-size: var(--td-font-size-title-large, 20px);
      font-weight: 600;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      margin-bottom: var(--td-comp-margin-l);
    }

    &--desc-clamped {
      cursor: pointer;
    }

    &--info {
      color: var(--td-text-color-secondary);
      font: var(--td-font-body-small);
      margin-bottom: var(--td-comp-margin-xs);
    }

    &--info-highlight {
      color: var(--td-text-color-primary);
      font-size: var(--td-font-size-body-medium, 14px);
      font-weight: normal;
      line-height: 1.6;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  &_footer-wrapper {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  &_footer-info {
    color: var(--td-text-color-placeholder);
    font: var(--td-font-body-small);
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .deadline-label {
    color: var(--td-text-color-placeholder);
  }

  .deadline-value {
    color: var(--td-text-color-secondary);
  }

.progress-section {
  margin-top: var(--td-comp-margin-m);
}

.progress-text {
  font-size: var(--td-font-size-body-medium, 14px);
  color: var(--td-text-color-secondary);
  margin-bottom: 4px;
}

  &_footer-actions {
    display: flex;
    align-items: center;
    gap: var(--td-comp-margin-xs);
  }
}

.work-report-icon {
  width: 100%;
  height: 100%;
  transform: scale(0.5);
}

// 外发生产卡片样式 - 浅紫色描边
.external-production {
  border: 2px solid #D6C6FF !important;
  box-shadow: 0 2px 8px rgba(136, 68, 255, 0.08);
}
</style>

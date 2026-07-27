<template>
  <div class="mold-items-form">
    <!-- 修模/改造订单的模具选择器 -->
    <div v-if="showMoldSelector" class="mold-select-wrapper">
      <t-form-item label="模具" :name="selectorFieldName">
        <t-select
          v-model="selectedMoldId"
          placeholder="请输入模具编号筛选"
          clearable
          filterable
          :filter="filterMold"
          :disabled="disabled"
          @change="handleMoldSelect"
        >
          <t-option
            v-for="mold in moldList"
            :key="mold.id"
            :label="`${mold.mold_code} - ${mold.mold_name}`"
            :value="mold.id"
          >
            <span class="mold-option">
              <span class="mold-code">{{ mold.mold_code }}</span>
              <span class="mold-separator"> - </span>
              <span class="mold-name">{{ mold.mold_name }}</span>
            </span>
          </t-option>
        </t-select>
      </t-form-item>
    </div>

    <!-- 模具明细内容区域 -->
    <div v-if="canShowItems" class="mold-items-wrapper">
      <!-- 空状态提示 -->
      <div v-if="items.length === 0" class="empty-state">
        <t-empty description="暂无模具明细数据" />
      </div>

      <draggable
        v-else
        v-model="items"
        item-key="index"
        handle=".drag-handle"
        animation="200"
        class="mold-sortable-list"
      >
        <template #item="{ element, index }">
          <div class="mold-item-row">
            <div class="mold-item-header">
              <div v-if="allowDrag" class="drag-handle">
                <move-icon size="20" />
              </div>
              <div class="mold-item-index">#{{ index + 1 }}</div>
              <div class="mold-item-title">{{ element.mold_code || '模具信息' }}</div>
              <t-button
                v-if="items.length > 1 && allowDelete"
                theme="danger"
                variant="text"
                shape="square"
                size="small"
                class="mold-delete-btn"
                @click="handleRemoveItem(index)"
              >
                <template #icon><delete-icon /></template>
              </t-button>
            </div>

            <!-- 基础信息板块 -->
            <div class="mold-section">
              <div class="form-basic-container-title">基础信息</div>
              <div class="mold-item-content">
                <t-input
                  v-model="element.mold_code"
                  :placeholder="moldCodePlaceholder"
                  :disabled="!editable"
                  :status="getMoldCodeError(element, index) ? 'error' : 'default'"
                  class="mold-input"
                  @blur="() => validateMoldCodeDuplicate(index)"
                >
                  <template #suffix>
                    <span
                      v-if="getMoldCodeError(element, index)"
                      style="color: #e34d59; font-size: 12px; white-space: nowrap"
                    >
                      {{ getMoldCodeError(element, index) }}
                    </span>
                  </template>
                </t-input>
                <!-- 模具名称：用户自定义，用于标识模具本身 -->
                <t-input
                  v-model="element.mold_name"
                  :placeholder="moldNamePlaceholder"
                  :disabled="!editable"
                  :status="getMoldNameError(element, index) ? 'error' : 'default'"
                  class="mold-input mold-name-input"
                  @blur="() => validateMoldNameRequired(index)"
                >
                  <template #suffix>
                    <span
                      v-if="getMoldNameError(element, index)"
                      style="color: #e34d59; font-size: 12px; white-space: nowrap"
                    >
                      {{ getMoldNameError(element, index) }}
                    </span>
                    <t-tooltip
                      content="模具名称用于标识模具本身，如：塑料多产品模具、手机壳专用模具等"
                      placement="top"
                    >
                      <t-icon name="info-circle" size="16" style="color: #999; cursor: help" />
                    </t-tooltip>
                  </template>
                </t-input>
                <t-input-number
                  v-model="element.amount"
                  :min="0"
                  :precision="2"
                  placeholder="请输入金额"
                  class="mold-number"
                  theme="normal"
                >
                  <template #suffix>元</template>
                </t-input-number>
              </div>
              <div class="mold-item-content mold-item-row-2">
                <t-select
                  v-model="element.moldProcess"
                  :placeholder="moldProcessPlaceholder"
                  :disabled="!editable"
                  class="mold-select"
                  :options="moldProcessOptions"
                />
                <t-select
                  v-model="element.moldType"
                  :placeholder="moldTypePlaceholder"
                  :disabled="!editable"
                  class="mold-select"
                  :options="moldTypeOptions"
                />
                <t-select
                  v-model="element.materialType"
                  :placeholder="materialTypePlaceholder"
                  :disabled="!editable"
                  class="mold-select"
                  :options="materialTypeOptions"
                />
                <t-select
                  v-model="element.cavityCount"
                  placeholder="模腔数"
                  class="mold-select"
                  :options="cavityCountOptions"
                />
                <t-input
                  v-model="element.remark"
                  placeholder="备注"
                  :disabled="!editable"
                  class="mold-input"
                />
              </div>
            </div>

            <!-- 加工信息板块（根据生产方式动态显示字段） -->
            <div class="mold-section">
              <div class="form-basic-container-title form-title-gap">加工信息</div>
              <div class="mold-item-content">
                <t-select
                  v-model="element.workType"
                  placeholder="生产方式"
                  :disabled="!editable"
                  class="mold-select work-type-select"
                  :options="workTypeOptions"
                />
                <!-- 外发加工时显示加工商 -->
                <template v-if="element.workType === 'OUTSOURCED'">
                  <t-select
                    v-model="element.processorId"
                    placeholder="请选择加工商"
                    :disabled="!editable"
                    class="mold-select"
                    clearable
                    filterable
                  >
                    <t-option
                      v-for="processor in processorList"
                      :key="processor.id"
                      :label="processor.displayName"
                      :value="processor.id"
                    />
                  </t-select>
                </template>
              </div>
            </div>
          </div>
        </template>
      </draggable>

      <t-button v-if="showAddButton" variant="outline" class="add-mold-btn" @click="handleAddItem">
        <template #icon><add-icon /></template>
        添加模具
      </t-button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { AddIcon, DeleteIcon, MoveIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref, watch } from 'vue';
import Draggable from 'vuedraggable';

import { getMoldList, getProcessorList } from '@/api/reference';
import { useDictRaw } from '@/hooks/useDict';

const props = defineProps({
  // 订单类型：NEW_MOLD=新模, REPAIR=修模, MODIFICATION=改造
  orderType: {
    type: String,
    default: 'NEW_MOLD',
  },
  // 模具明细数据
  modelValue: {
    type: Array as () => Array<{
      mold_id?: number;
      mold_code: string;
      mold_name: string; // 产品名称（支持多产品，用 / 或 、 分隔）
      product_count?: number; // 可生产产品数量（默认1）
      amount?: number;
      workType?: string; // 生产方式：IN_HOUSE=本厂生产，OUTSOURCED=外发加工
      processorId?: number; // 加工商ID
      moldProcess?: number | string;
      moldType?: number | string;
      materialType?: number | string;
      cavityCount?: number;
      remark: string;
    }>,
    default: () =>
      [] as Array<{
        mold_id?: number;
        mold_code: string;
        mold_name: string; // 模具名称（用户自定义）
        amount?: number;
        moldProcess?: number | string;
        moldType?: number | string;
        materialType?: number | string;
        cavityCount?: number;
        remark: string;
      }>,
  },
  // 是否显示模具选择器（修模/改造订单）
  showMoldSelector: {
    type: Boolean,
    default: false,
  },
  // 模具选择器字段名
  selectorFieldName: {
    type: String,
    default: 'moldId',
  },
  // 已选择的模具ID（修模/改造订单）
  selectedMoldId: {
    type: Number,
    default: undefined,
  },
  // 是否可编辑
  editable: {
    type: Boolean,
    default: true,
  },
  // 是否允许拖拽排序
  allowDrag: {
    type: Boolean,
    default: true,
  },
  // 是否允许删除
  allowDelete: {
    type: Boolean,
    default: true,
  },
  // 是否禁用
  disabled: {
    type: Boolean,
    default: false,
  },
  // 客户ID（用于修模/改模时按客户过滤模具列表）
  customerId: {
    type: Number,
    default: undefined,
  },
});

const emit = defineEmits<{
  (e: 'update:modelValue', value: any[]): void;
  (e: 'update:selectedMoldId', value: number | undefined): void;
  (e: 'mold-select', value: number | undefined): void;
}>();

// 模具明细列表
const items = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
});

// 选中的模具ID
const selectedMoldId = computed({
  get: () => props.selectedMoldId,
  set: (value) => emit('update:selectedMoldId', value),
});

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: moldProcessOptions } = useDictRaw('mold-process', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1)
    .map((item: any) => ({
      label: item.dict_label,
      value: item.dict_value,
    })),
);
const { items: moldTypeOptions } = useDictRaw('mold-type', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1)
    .map((item: any) => ({
      label: item.dict_label,
      value: item.dict_value,
    })),
);
const { items: materialTypeOptions } = useDictRaw('material-type', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1)
    .map((item: any) => ({
      label: item.dict_label,
      value: item.dict_value,
    })),
);
const workTypeOptions = ref<any[]>([
  { label: '本厂生产', value: 'IN_HOUSE' },
  { label: '外发加工', value: 'OUTSOURCED' },
]);

// 模腔数选项（1-30）
const cavityCountOptions = ref<any[]>(
  Array.from({ length: 30 }, (_, i) => ({ label: `${i + 1} 腔`, value: i + 1 })),
);
const moldList = ref<any[]>([]);
const processorList = ref<any[]>([]);

// 是否显示明细区域
const canShowItems = computed(() => {
  // 新模订单、综合订单或有数据时显示
  return props.orderType === 'NEW_MOLD' || props.orderType === 'COMBINED' || items.value.length > 0;
});

// ✨ 模具编号重复校验
const moldCodeErrors = ref<Record<number, string>>({});

const validateMoldCodeDuplicate = (index: number) => {
  const currentCode = items.value[index]?.mold_code?.trim();

  if (!currentCode) {
    delete moldCodeErrors.value[index];
    return;
  }

  const duplicateIndex = items.value.findIndex(
    (item, idx) => idx !== index && item.mold_code?.trim() === currentCode,
  );

  if (duplicateIndex !== -1) {
    moldCodeErrors.value[index] = '模具编号重复';
  } else {
    delete moldCodeErrors.value[index];
  }
};

const getMoldCodeError = (element: any, index: number) => {
  return moldCodeErrors.value[index] || '';
};

// ✨ 模具名称必填校验
const moldNameErrors = ref<Record<number, string>>({});

const validateMoldNameRequired = (index: number) => {
  const currentName = items.value[index]?.mold_name?.trim();

  if (currentName) {
    delete moldNameErrors.value[index];
    return;
  }

  // 模具名称不能为空
  moldNameErrors.value[index] = '模具名称不能为空';
};

const getMoldNameError = (element: any, index: number) => {
  return moldNameErrors.value[index] || '';
};

// 是否显示添加按钮
const showAddButton = computed(() => {
  // 新模订单或综合订单，且可编辑时显示添加按钮
  return (props.orderType === 'NEW_MOLD' || props.orderType === 'COMBINED') && props.editable;
});

// Placeholder 文本
const moldCodePlaceholder = computed(() =>
  props.orderType === 'NEW_MOLD' ? '模具编号' : '模具编号',
);
const moldNamePlaceholder = computed(() =>
  props.orderType === 'NEW_MOLD' ? '模具名称' : '模具名称',
);
const moldProcessPlaceholder = computed(() =>
  props.orderType === 'NEW_MOLD' ? '请选择工艺' : '请选择工艺',
);
const moldTypePlaceholder = computed(() =>
  props.orderType === 'NEW_MOLD' ? '请选择模胚' : '请选择模胚',
);
const materialTypePlaceholder = computed(() =>
  props.orderType === 'NEW_MOLD' ? '请选择材质' : '请选择材质',
);

// 加载字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载模具列表（用于修模/改造订单，可按客户过滤）
const loadMoldList = async () => {
  try {
    const params: any = { page: 1, page_size: 100 };
    if (props.customerId) {
      params.customer_id = props.customerId;
    }
    const res = await getMoldList(params);
    moldList.value = res.list || [];
  } catch (error) {
    console.error('获取模具列表失败:', error);
  }
};

const loadProcessorList = async () => {
  try {
    const res = await getProcessorList({ page: 1, page_size: 100 });
    processorList.value = (res.list || []).map((item: any) => ({
      id: item.id,
      name: item.processor_name || item.name,
      processType: item.process_type,
      displayName: item.process_type
        ? `${item.processor_name || item.name} - ${item.process_type}`
        : item.processor_name || item.name,
    }));
  } catch (error) {
    console.error('获取加工商列表失败:', error);
  }
};

// 选择模具后填充信息
const handleMoldSelect = (value: any) => {
  const moldId = value as number | undefined;
  emit('update:selectedMoldId', moldId);
  emit('mold-select', moldId);

  if (!moldId) {
    emit('update:modelValue', []);
    return;
  }

  const selectedMold = moldList.value.find((m) => m.id === moldId);
  if (selectedMold) {
    emit('update:modelValue', [
      {
        mold_id: moldId,
        mold_code: selectedMold.mold_code || '',
        mold_name: selectedMold.mold_name || '',
        amount: undefined,
        moldProcess: selectedMold.mold_process,
        moldType: selectedMold.mold_type,
        materialType: selectedMold.material_type,
        cavityCount: selectedMold.cavity_count,
        remark: selectedMold.remark || '',
      },
    ]);
  }
};

// 模具筛选函数
const filterMold = (keyword: string, option: any) => {
  if (!keyword) return true;
  const mold = moldList.value.find((m) => m.id === option.value);
  if (!mold) return false;
  const searchText = keyword.toLowerCase();
  return (
    mold.mold_code?.toLowerCase().includes(searchText) ||
    mold.mold_name?.toLowerCase().includes(searchText)
  );
};

const handleAddItem = () => {
  const newItem: {
    mold_id?: number;
    mold_code: string;
    mold_name: string; // 模具名称（用户自定义）
    amount?: number;
    workType?: string; // 生产方式：IN_HOUSE=本厂生产，OUTSOURCED=外发加工
    processorId?: number; // 加工商ID
    moldProcess?: number | string;
    moldType?: number | string;
    materialType?: number | string;
    cavityCount?: number;
    remark: string;
  } = {
    mold_id: 0, // ✨ 设置为 0，表示新模具
    mold_code: '',
    mold_name: '',
    amount: undefined,
    workType: 'IN_HOUSE', // 默认本厂生产
    processorId: undefined,
    moldProcess: undefined,
    moldType: undefined,
    materialType: undefined,
    cavityCount: undefined,
    remark: '',
  };
  emit('update:modelValue', [...items.value, newItem]);
};

const handleRemoveItem = (index: number) => {
  if (items.value.length <= 1) {
    MessagePlugin.warning('至少保留一项模具明细');
    return;
  }
  const newItems = [...items.value];
  newItems.splice(index, 1);
  emit('update:modelValue', newItems);
};

// 监听 showMoldSelector 变化，当切换到修模/改模时加载模具列表
watch(
  () => props.showMoldSelector,
  async (newVal) => {
    if (newVal && moldList.value.length === 0) {
      await loadMoldList();
    }
  },
  { immediate: true },
);

// 监听客户变化，重新加载模具列表（不清空已选择，由父组件决定是否需要清空）
watch(
  () => props.customerId,
  async () => {
    if (props.showMoldSelector && props.customerId) {
      await loadMoldList();
    }
  },
);

onMounted(async () => {
  // 字典已通过 useDictRaw 自动加载
  await loadProcessorList(); // 加载加工商列表
});

defineExpose({
  loadMoldList,
  validateItems: () => {
    const errors: string[] = [];
    const codes = new Set<string>();
    const data = items.value;

    for (let i = 0; i < data.length; i++) {
      const item = data[i];
      const code = item.mold_code?.trim();

      if (!code) {
        errors.push(`第 ${i + 1} 条模具明细的模具编号不能为空`);
      } else if (codes.has(code)) {
        errors.push(`模具编号"${code}"重复`);
      } else {
        codes.add(code);
      }

      if (!item.mold_name?.trim()) {
        errors.push(`第 ${i + 1} 条模具明细的产品名称不能为空`);
      }
      if (!item.amount || item.amount <= 0) {
        errors.push(`第 ${i + 1} 条模具明细的金额必须大于 0`);
      }
      if (!item.moldProcess) {
        errors.push(`第 ${i + 1} 条模具明细的工艺不能为空`);
      }
      if (!item.moldType) {
        errors.push(`第 ${i + 1} 条模具明细的模胚不能为空`);
      }
      if (!item.materialType) {
        errors.push(`第 ${i + 1} 条模具明细的材质不能为空`);
      }
      if (!item.cavityCount) {
        errors.push(`第 ${i + 1} 条模具明细的模腔数不能为空`);
      }
      if (!item.workType) {
        errors.push(`第 ${i + 1} 条模具明细的加工方式不能为空`);
      }
      if (item.workType === 'OUTSOURCED' && !item.processorId) {
        errors.push(`第 ${i + 1} 条模具明细为外发加工，请选择加工商`);
      }
    }

    return { valid: errors.length === 0, errors, data };
  },
});
</script>
<style lang="less" scoped>
.mold-items-form {
  width: 100%;
}

.mold-select-wrapper {
  margin-bottom: var(--td-comp-margin-m);
  padding: var(--td-comp-paddingTB-l) var(--td-comp-paddingLR-l);
  background-color: var(--td-bg-color-secondarycontainer);
  border-radius: var(--td-radius-medium);
}

.mold-items-wrapper {
  background-color: var(--td-bg-color-secondarycontainer);
  padding: var(--td-comp-paddingTB-m) var(--td-comp-paddingLR-m);
  border-radius: var(--td-radius-medium);

  .mold-sortable-list {
    position: relative;
  }

  .empty-state {
    padding: var(--td-comp-paddingTB-xl);
    text-align: center;
  }

  .mold-item-header {
    display: flex;
    align-items: center;
    gap: var(--td-comp-margin-s);
    margin-bottom: var(--td-comp-margin-m);

    .drag-handle {
      flex-shrink: 0;
      cursor: move;
      color: var(--td-text-color-placeholder);
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 4px;
      border-radius: var(--td-radius-small);
      transition: all 0.2s;

      &:hover {
        color: var(--td-brand-color);
        background-color: var(--td-bg-color-secondarycontainer);
      }
    }

    .mold-item-index {
      flex-shrink: 0;
      width: 32px;
      height: 32px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #0052d9 !important;
      color: #fff !important;
      border-radius: var(--td-radius-circle);
      font-size: var(--td-font-size-body-medium);
      font-weight: 600;
      box-shadow: 0 2px 6px rgb(0 82 217 / 25%);
    }

    .mold-item-title {
      flex: 1;
      font-size: var(--td-font-size-body-large);
      font-weight: 500;
      color: var(--td-text-color-primary);
      line-height: 32px;
    }

    .mold-delete-btn {
      flex-shrink: 0;
    }
  }

  .mold-item-row {
    margin-bottom: var(--td-comp-margin-l);
    padding: var(--td-comp-paddingTB-l) var(--td-comp-paddingLR-l);
    border-radius: var(--td-radius-medium);
    background-color: #fff;
    box-shadow: 0 1px 4px rgb(0 0 0 / 8%);

    &:last-child {
      margin-bottom: var(--td-comp-margin-m);
    }
  }

  // 板块样式
  .mold-section {
    margin-top: var(--td-comp-margin-m);

    &:first-of-type {
      margin-top: 0;
    }

    // 小标题加粗
    .form-basic-container-title {
      font-weight: 600;
    }
  }

  .mold-item-content {
    display: flex;
    align-items: center;
    gap: var(--td-comp-margin-s);
    flex-wrap: wrap;
    margin-top: var(--td-comp-margin-s);

    .mold-input {
      flex: 3;
      min-width: 120px;
    }

    .mold-number {
      flex: 1.5;
      min-width: 120px;
    }

    .mold-select {
      flex: 2;
      min-width: 90px;
    }

    .work-type-select {
      flex: 1;
      max-width: 200px;
    }

    .mold-delete-btn {
      flex-shrink: 0;
    }
  }

  .add-mold-btn {
    width: 100%;
    margin-top: var(--td-comp-margin-m);
  }
}

.mold-option {
  display: flex;
  align-items: center;
}

.mold-code {
  font-weight: 600;
  font-size: 1.1em;
}

.mold-separator {
  margin: 0 6px;
  opacity: 0.6;
}

.mold-name {
  color: var(--td-text-color-placeholder);
}
</style>

<template>
  <t-card class="search-card" :bordered="false">
    <t-form :data="formData" label-width="0">
      <div class="search-area-inner">
        <!-- 左侧筛选条件 -->
        <div class="search-area-form">
          <t-row
            v-for="(row, rowIdx) in basicFilterRows"
            :key="'basic-' + rowIdx"
            :gutter="[16, 24]"
          >
            <t-col
              v-for="item in row"
              :key="item.field"
              :span="item.span ?? 4"
            >
              <t-form-item>
                <t-input
                  v-if="item.type === 'input'"
                  v-model="formData[item.field]"
                  :placeholder="item.placeholder"
                  clearable
                  @enter="handleSearch"
                />
                <t-select
                  v-else-if="item.type === 'select'"
                  v-model="formData[item.field]"
                  :placeholder="item.placeholder"
                  :multiple="item.multiple"
                  clearable
                  filterable
                >
                  <t-option
                    v-for="opt in resolveOptions(item)"
                    :key="item.valueKey ? opt[item.valueKey] : opt.value"
                    :label="item.labelKey ? opt[item.labelKey] : opt.label"
                    :value="item.valueKey ? opt[item.valueKey] : opt.value"
                  />
                </t-select>
                <t-date-range-picker
                  v-else-if="item.type === 'dateRangePicker'"
                  v-model="formData[item.field]"
                  mode="date"
                  clearable
                  style="width: 100%"
                />
              </t-form-item>
            </t-col>
          </t-row>

          <!-- 更多筛选 (展开收起) -->
          <Transition name="fade-slide">
            <div v-if="showMore && moreFilterRows.length" class="expand-wrap">
              <t-row
                v-for="(row, rowIdx) in moreFilterRows"
                :key="rowIdx"
                :gutter="[16, 24]"
              >
                <t-col
                  v-for="item in row"
                  :key="item.field"
                  :span="item.span ?? 4"
                >
                  <t-form-item>
                    <t-input
                      v-if="item.type === 'input'"
                      v-model="formData[item.field]"
                      :placeholder="item.placeholder"
                      clearable
                      @enter="handleSearch"
                    />
                    <t-select
                      v-else-if="item.type === 'select'"
                      v-model="formData[item.field]"
                      :placeholder="item.placeholder"
                      :multiple="item.multiple"
                      clearable
                      filterable
                    >
                      <t-option
                        v-for="opt in resolveOptions(item)"
                        :key="item.valueKey ? opt[item.valueKey] : opt.value"
                        :label="item.labelKey ? opt[item.labelKey] : opt.label"
                        :value="item.valueKey ? opt[item.valueKey] : opt.value"
                      />
                    </t-select>
                    <t-date-range-picker
                      v-else-if="item.type === 'dateRangePicker'"
                      v-model="formData[item.field]"
                      mode="date"
                      clearable
                      style="width: 100%"
                    />
                  </t-form-item>
                </t-col>
              </t-row>
            </div>
          </Transition>
        </div>

        <!-- 右侧固定按钮 -->
        <div class="search-area-actions">
          <t-space>
            <t-button
              v-if="moreFilterRows.length"
              variant="text"
              theme="primary"
              @click="showMore = !showMore"
            >
              <template #icon>
                <chevron-down-icon v-if="!showMore" />
                <chevron-up-icon v-else />
              </template>
              {{ showMore ? '收起筛选' : '更多筛选' }}
            </t-button>
            <t-button theme="primary" @click="handleSearch">
              <template #icon><search-icon /></template>
              查询
            </t-button>
            <t-button theme="default" @click="handleReset">
              <template #icon><refresh-icon /></template>
              重置
            </t-button>
          </t-space>
        </div>
      </div>
    </t-form>
  </t-card>
</template>

<script setup lang="ts">
import { ref, computed, toRef, toValue, nextTick, type MaybeRefOrGetter } from 'vue';
import {
  SearchIcon,
  RefreshIcon,
  ChevronDownIcon,
  ChevronUpIcon,
} from 'tdesign-icons-vue-next';

export type FilterType = 'input' | 'select' | 'dateRangePicker';

export interface FilterItem {
  /** 筛选类型 */
  type: FilterType
  /** 表单字段名 (v-model 绑定的 key) */
  field: string
  /** placeholder 提示文字 */
  placeholder?: string
  /** select 的选项列表 (支持响应式) */
  options?: MaybeRefOrGetter<any[]>
  /** 选项文本字段名，默认 'label' */
  labelKey?: string
  /** 选项值字段名，默认 'value' */
  valueKey?: string
  /** 是否属于"更多筛选"(收起区域)，默认 false */
  showMore?: boolean
  /** 占据列数，默认 4（一行 3 列），dateRangePicker 建议传 8 */
  span?: number
  /** select 是否多选，仅 type='select' 时生效，默认 false */
  multiple?: boolean
}

const props = withDefaults(defineProps<{
  /** 筛选条件配置 */
  filters: FilterItem[]
  /** 表单数据对象 (v-model) */
  modelValue: Record<string, any>
}>(), {});

const emit = defineEmits<{
  (e: 'search', formData?: Record<string, any>): void
  (e: 'reset', formData?: Record<string, any>): void
  (e: 'update:modelValue', val: Record<string, any>): void
}>();

/** 表单数据引用
 * 使用 toRef 让内部 formData 始终指向 props.modelValue，
 * 无论父组件传入的是 reactive 代理还是 ref 包裹的普通对象，
 * v-model 的修改都能直接同步到父组件。
 */
const formData = toRef(props, 'modelValue');

const showMore = ref(false);

/** 将筛选项按每行总 span ≤ 12 分组 */
function groupIntoRows(items: FilterItem[]): FilterItem[][] {
  const rows: FilterItem[][] = [];
  let currentRow: FilterItem[] = [];
  let currentSpan = 0;
  for (const item of items) {
    const span = item.span ?? 4;
    if (currentSpan + span > 12 && currentRow.length > 0) {
      rows.push(currentRow);
      currentRow = [];
      currentSpan = 0;
    }
    currentRow.push(item);
    currentSpan += span;
  }
  if (currentRow.length > 0) {
    rows.push(currentRow);
  }
  return rows;
}

/** 基础筛选行（默认展示第一行，其余自动折叠） */
/** 基础筛选行：第一行 */
const basicFilterRows = computed(() => {
  const rows = allFilterRows.value;
  return rows.length > 0 ? [rows[0]] : [];
});

/** 更多筛选行：从第二行开始 */
const moreFilterRows = computed(() => {
  const rows = allFilterRows.value;
  return rows.slice(1);
});

/** 所有筛选项按 span 分行的结果 */
const allFilterRows = computed(() =>
  groupIntoRows(props.filters),
);

/** 解析 options，支持 reactive ref / getter */
function resolveOptions(item: FilterItem): any[] {
  return item.options ? toValue(item.options) : [];
}

function handleSearch() {
  // 延迟到 nextTick 触发，确保父组件已完成重新渲染，
  // 子组件通过 props 接收到的筛选条件已更新
  nextTick(() => {
    emit('search', { ...formData.value });
  });
}

function handleReset() {
  // 直接修改响应式对象的属性，保持与父组件的数据同步
  for (const key of Object.keys(formData.value)) {
    formData.value[key] = '';
  }
  nextTick(() => {
    emit('reset', { ...formData.value });
  });
}

defineExpose({
  showMore,
});
</script>

<style scoped>
.search-card {
  margin-bottom: 16px;
}

.search-area-inner {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 32px;
}

.search-area-form {
  flex: 1;
  min-width: 0;
  margin-bottom: -16px;
}

.search-area-form :deep(.t-form__item) {
  margin-bottom: 16px !important;
}

.search-area-actions {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  white-space: nowrap;
}

/* 展开收起动画 */
.expand-wrap {
  overflow: hidden;
}

.fade-slide-enter-active {
  animation: fadeSlideIn 0.25s ease-out;
}

.fade-slide-leave-active {
  animation: fadeSlideIn 0.2s ease-in reverse;
}

@keyframes fadeSlideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
    max-height: 0;
  }
  to {
    opacity: 1;
    transform: translateY(0);
    max-height: 200px;
  }
}
</style>

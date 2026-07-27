<template>
  <t-dialog
    v-model:visible="visible"
    :title="title"
    :width="width"
    :footer="null"
    @close="handleClose"
   :close-on-overlay-click="false">
    <div class="column-setting-content">
      <div ref="columnListRef" class="draggable-column-list">
        <!-- 按照 columnOrder 的完整顺序渲染所有列 -->
        <div 
          v-for="colKey in columnOrder" 
          :key="colKey" 
          class="column-item"
          :class="{ 
            'draggable-item': draggable && selectedColumns.includes(colKey),
            'pinned-item': isColumnPinned(colKey)
          }"
          :data-col-key="colKey"
        >
          <t-checkbox 
            :checked="selectedColumns.includes(colKey)"
            :disabled="isColumnPinned(colKey) || isColLockHide(colKey)"
            @change="(checked: boolean) => handleColumnCheckChange(colKey, checked)"
          >
            {{ getColumnTitle(colKey) }}
          </t-checkbox>
          <!-- 固定列显示锁图标，锁定隐藏列和普通列显示拖拽图标 -->
          <LockOnIcon v-if="isColumnPinned(colKey)" class="lock-icon" />
          <MoveIcon v-else-if="draggable && selectedColumns.includes(colKey)" class="drag-handle" />
        </div>
      </div>
      <div class="column-setting-footer">
        <t-button theme="primary" @click="handleReset">重置默认</t-button>
        <t-button @click="handleClose">关闭</t-button>
      </div>
    </div>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { MoveIcon, LockOnIcon } from 'tdesign-icons-vue-next';
import { useDraggable } from 'vue-draggable-plus';

export interface ColumnOption {
  colKey: string;
  title: string;
  /** 是否固定列（不可隐藏、不可移动，始终保持在原始位置） */
  pinned?: boolean;
  /** 是否锁定隐藏（不可隐藏，但可以拖拽移动位置） */
  lockHide?: boolean;
}

export interface ColumnSettingProps {
  /** 对话框标题 */
  title?: string;
  /** 对话框宽度 */
  width?: string;
  /** 所有可选列配置 */
  allColumns: ColumnOption[];
  /** 默认显示的列 */
  defaultColumns: string[];
  /** localStorage 存储键名 */
  storageKey: string;
  /** 是否启用拖拽排序 */
  draggable?: boolean;
}

const props = withDefaults(defineProps<ColumnSettingProps>(), {
  title: '列设置',
  width: '500px',
  draggable: true,
});

const emit = defineEmits<{
  /** 列配置变化时触发 */
  change: [columns: string[]];
  /** 对话框关闭时触发 */
  close: [];
}>();

// ---- 响应式状态 ----
const visible = ref(false);
/** 完整的列顺序（包含所有列，无论是否显示） */
const columnOrder = ref<string[]>([]);
/** 当前可见的列 */
const selectedColumns = ref<string[]>([]);
const columnListRef = ref<HTMLElement | null>(null);
/** 标记拖拽是否已初始化，避免重复调用 useDraggable */
const dragInitialized = ref(false);

// ---- 工具 ----
/** 用 Map 实现 O(1) 列查找，替代模板中每行的 O(n) find */
const columnMap = computed(() => new Map(props.allColumns.map(c => [c.colKey, c])));

const getColumnTitle = (colKey: string) => columnMap.value.get(colKey)?.title || colKey;
const isColumnPinned = (colKey: string) => columnMap.value.get(colKey)?.pinned === true;
const isColLockHide = (colKey: string) => columnMap.value.get(colKey)?.lockHide === true;

const getForcedVisibleKeys = () =>
  props.allColumns.filter(col => col.pinned || col.lockHide).map(col => col.colKey);

// ---- 固定列与锁定隐藏列强化 ----
const ensurePinnedColumns = () => {
  const forcedVisibleKeys = getForcedVisibleKeys();
  const pinnedKeys = props.allColumns.filter(col => col.pinned).map(col => col.colKey);
  if (forcedVisibleKeys.length === 0) return;

  // 确保强制可见列始终在 selectedColumns 中
  selectedColumns.value = [...new Set([...selectedColumns.value, ...forcedVisibleKeys])];

  // 仅将固定列（pinned）移动到最前面，锁定隐藏列（lockHide）不移动
  pinnedKeys.forEach(colKey => {
    const idx = columnOrder.value.indexOf(colKey);
    if (idx > 0) {
      columnOrder.value.splice(idx, 1);
      columnOrder.value.unshift(colKey);
    }
  });

  // 同步 selectedColumns 顺序与 columnOrder 一致
  const set = new Set(selectedColumns.value);
  selectedColumns.value = columnOrder.value.filter(k => set.has(k));
};

// ---- 存储 ----
const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem(props.storageKey);
    const allColKeys = props.allColumns.map(col => col.colKey);
    const validKeys = new Set(allColKeys);

    if (saved) {
      const savedData = JSON.parse(saved);

      if (Array.isArray(savedData)) {
        columnOrder.value = [...allColKeys];
        selectedColumns.value = savedData.filter((k: string) => validKeys.has(k));
      } else if (savedData.columns) {
        const { columns = [], visible: visibleCols = [] } = savedData;
        columnOrder.value = columns.filter((k: string) => validKeys.has(k));
        selectedColumns.value = visibleCols.filter((k: string) => validKeys.has(k));
      } else if (savedData.order) {
        const { order = [], visible: visibleCols = [] } = savedData;
        columnOrder.value = order.filter((k: string) => validKeys.has(k));
        selectedColumns.value = visibleCols.filter((k: string) => validKeys.has(k));
      } else {
        columnOrder.value = [...allColKeys];
        selectedColumns.value = [...props.defaultColumns];
      }

      ensurePinnedColumns();

      // 补充可能新增的列到末尾
      props.allColumns.forEach(col => {
        if (!columnOrder.value.includes(col.colKey)) {
          columnOrder.value.push(col.colKey);
        }
      });
    } else {
      columnOrder.value = [...allColKeys];
      selectedColumns.value = [...new Set([...props.defaultColumns, ...getForcedVisibleKeys()])];
    }
  } catch (error) {
    console.error('加载列配置失败:', error);
    columnOrder.value = props.allColumns.map(col => col.colKey);
    selectedColumns.value = [...props.defaultColumns];
  }
};

const saveColumnConfig = () => {
  try {
    localStorage.setItem(props.storageKey, JSON.stringify({
      order: columnOrder.value,
      visible: selectedColumns.value,
    }));
  } catch (error) {
    console.error('保存列配置失败:', error);
  }
};

// ---- 操作 ----
const handleColumnCheckChange = (colKey: string, checked: boolean) => {
  if ((isColumnPinned(colKey) || isColLockHide(colKey)) && !checked) return;

  if (checked) {
    if (!selectedColumns.value.includes(colKey)) {
      const orderIndex = columnOrder.value.indexOf(colKey);
      let insertIndex = selectedColumns.value.length;
      for (let i = 0; i < selectedColumns.value.length; i++) {
        if (columnOrder.value.indexOf(selectedColumns.value[i]) > orderIndex) {
          insertIndex = i;
          break;
        }
      }
      selectedColumns.value.splice(insertIndex, 0, colKey);
    }
  } else {
    const index = selectedColumns.value.indexOf(colKey);
    if (index > -1) {
      selectedColumns.value.splice(index, 1);
    }
  }
  saveColumnConfig();
  emit('change', [...selectedColumns.value]);
};

const handleReset = () => {
  columnOrder.value = props.allColumns.map(col => col.colKey);
  selectedColumns.value = [...new Set([...props.defaultColumns, ...getForcedVisibleKeys()])];
  saveColumnConfig();
  emit('change', [...selectedColumns.value]);
  MessagePlugin.success('已重置为默认配置');
};

// ---- 初始化拖拽（推迟到首次打开时，避免组件挂载即注册） ----
const initDraggable = () => {
  if (!props.draggable || dragInitialized.value) return;
  dragInitialized.value = true;

  useDraggable(columnListRef, columnOrder, {
    filter: '.column-item:not(.draggable-item), .pinned-item',
    animation: 150,
    ghostClass: 'ghost-item',
    chosenClass: 'chosen-item',
    onEnd: () => {
      ensurePinnedColumns();
      saveColumnConfig();
      emit('change', [...selectedColumns.value]);
    },
  });
};

// ---- 公开 API ----
const open = () => {
  loadColumnConfig();
  initDraggable();
  visible.value = true;
};

const handleClose = () => {
  visible.value = false;
  emit('close');
};

/** 只读获取当前可见列，替代直接暴露 ref（防止父组件绕过 save/emit） */
const getSelectedColumns = () => [...selectedColumns.value];

defineExpose({
  open,
  getSelectedColumns,
});
</script>

<style lang="less" scoped>
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

/* 拖拽时的占位元素样式 */
.ghost-item {
  opacity: 0.4;
  background-color: var(--td-brand-color-light, #e0e7ff);
  border: 1px dashed var(--td-brand-color, #0052d9);
}

/* 正在拖拽的元素样式 */
.chosen-item {
  background-color: var(--td-bg-color-container-active, #f0f0f0);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
}

/* 固定列样式 */
.pinned-item {
  .t-checkbox {
    cursor: not-allowed;
  }
}

.lock-icon {
  color: var(--td-text-color-placeholder, #999);
  flex-shrink: 0;
  font-size: 16px;
}

.column-setting-footer {
  margin-top: 24px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>

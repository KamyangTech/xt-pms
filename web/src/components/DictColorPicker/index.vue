<template>
  <t-select
    :key="selectKey"
    :model-value="modelValue"
    @update:model-value="handleChange"
    :placeholder="placeholder"
    :clearable="clearable"
    :popup-props="{ 
      overlayStyle: { maxHeight: '400px', zIndex: 5000 },
      attach: 'body',
      placement: 'bottom-left',
      destroyOnClose: true
    }"
  >
    <!-- 自定义选中值显示 -->
    <template #valueDisplay="{ value }">
      <div v-if="value" class="color-selected-display">
        <span
          class="color-preview"
          :style="{ background: getColorHex(value) }"
        ></span>
        <span>{{ getColorName(value) }}</span>
      </div>
      <span v-else>{{ placeholder }}</span>
    </template>

    <!-- 颜色选项列表（按分类分组） -->
    <t-option-group
      v-for="group in colorGroups"
      :key="group.group"
      :label="group.group"
    >
      <t-option
        v-for="option in group.options"
        :key="option.value"
        :value="option.value"
      >
        <div class="color-option-item">
          <span
            class="color-preview"
            :style="{ background: option.hex }"
          ></span>
          <span>{{ option.label }}</span>
        </div>
      </t-option>
    </t-option-group>
  </t-select>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, watch } from 'vue';
import {
  getAllColorOptions,
  getColorHex,
  getColorName,
} from '@/config/dict-color-config';

interface Props {
  modelValue?: string;
  placeholder?: string;
  clearable?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '请选择颜色',
  clearable: true,
});

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

// 颜色分组选项
const colorGroups = computed(() => getAllColorOptions());

// 使用 key 确保组件正确重新渲染
const selectKey = ref(0);

// 监听 modelValue 变化，刷新组件
watch(() => props.modelValue, () => {
  selectKey.value++;
});

// 组件挂载时初始化
onMounted(() => {
  selectKey.value = Date.now();
});

const handleChange = (value: string) => {
  emit('update:modelValue', value);
};
</script>

<style scoped lang="less">
.color-selected-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.color-option-item {
  display: flex;
  align-items: center;
  gap: 8px;
}

.color-preview {
  display: inline-block;
  width: 20px;
  height: 20px;
  border: 1px solid rgba(0, 0, 0, 0.1);
  border-radius: 2px;
  flex-shrink: 0;
}
</style>

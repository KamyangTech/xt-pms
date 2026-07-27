<template>
  <t-tag
    shape="round"
    :variant="computedVariant"
    :theme="theme"
    :style="computedStyle"
    v-bind="$attrs"
  >
    <slot />
  </t-tag>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { TagProps } from 'tdesign-vue-next';

/**
 * 订单标签类型
 * - type: 类型标签（线描边），如订单类型、生产类型、优先级等
 * - status: 状态标签（实心），如订单状态、依赖状态等
 */
export type TagKind = 'type' | 'status';

interface Props {
  /** 标签类型：type-类型标签（线描边），status-状态标签（实心） */
  kind?: TagKind;
  /** 主题色 */
  theme?: TagProps['theme'];
  /** 自定义样式 */
  style?: Record<string, any>;
}

const props = withDefaults(defineProps<Props>(), {
  kind: 'status',
  theme: 'default',
});

// 计算 variant：类型标签用 outline，状态标签用默认（实心）
const computedVariant = computed<TagProps['variant']>(() => {
  return props.kind === 'type' ? 'outline' : undefined;
});

const computedStyle = computed(() => props.style || {});
</script>

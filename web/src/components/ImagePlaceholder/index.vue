<template>
  <t-image
    v-if="displaySrc"
    :src="getImageUrl(displaySrc)"
    :fit="fit"
    :shape="shape"
    :style="imageStyle"
    :error="renderErrorIcon"
    :loading="renderLoadingIcon"
    lazy
    @error="handleImageError"
    v-bind="$attrs"
  />
  <div v-else :class="['no-image-placeholder', shapeClass]" :style="imageStyle">
    <ImageIcon size="24" />
  </div>
</template>

<script setup lang="ts">
import { ImageErrorIcon, ImageIcon, LoadingIcon } from 'tdesign-icons-vue-next';
import type { ImageProps } from 'tdesign-vue-next';
import { computed, h, ref } from 'vue';

interface Props {
  /** 图片 URL（优先加载，如缩略图） */
  src?: string;
  /** fallback 图片 URL（src 加载失败时回退，如原图） */
  fallbackSrc?: string;
  /** 图片适应方式 */
  fit?: 'contain' | 'cover' | 'fill' | 'none' | 'scale-down';
  /** 图片形状 */
  shape?: 'circle' | 'round' | 'square';
  /** 图片宽度 */
  width?: string | number;
  /** 图片高度 */
  height?: string | number;
}

const props = withDefaults(defineProps<Props>(), {
  src: '',
  fallbackSrc: '',
  fit: 'cover',
  shape: 'square',
  width: 80,
  height: 80,
});

// 当前显示的图片源：优先 src，加载失败后回退到 fallbackSrc
const useFallback = ref(false);
const displaySrc = computed(() => {
  if (!useFallback.value) return props.src;
  return props.fallbackSrc;
});

const hasImage = computed(() => !!displaySrc.value && displaySrc.value.trim() !== '');

const getImageUrl = (url: string) => {
  if (!url) return '';
  // 如果已经是完整 URL，直接返回
  if (url.startsWith('http') || url.startsWith('//')) {
    return url;
  }
  // 开发环境使用代理时，返回相对路径让 Vite 代理处理（避免跨域问题）
  const isProxy = import.meta.env.VITE_IS_REQUEST_PROXY === 'true';
  if (isProxy) {
    return url;
  }
  // 生产环境或其他情况拼接完整 URL
  const baseUrl = import.meta.env.VITE_API_URL || '';
  return `${baseUrl}${url}`;
};

const imageUrl = computed(() => getImageUrl(displaySrc.value));

// 图片加载失败时回退到 fallbackSrc
const handleImageError = () => {
  if (props.fallbackSrc && !useFallback.value) {
    useFallback.value = true;
  }
};

// 图片样式
const imageStyle = computed(() => ({
  width: typeof props.width === 'number' ? `${props.width}px` : props.width,
  height: typeof props.height === 'number' ? `${props.height}px` : props.height,
}));

// 形状类名
const shapeClass = computed(() => {
  const shapeMap = {
    circle: 'shape-circle',
    round: 'shape-round',
    square: 'shape-square',
  };
  return shapeMap[props.shape] || 'shape-square';
});

// 渲染错误图标
const renderErrorIcon: ImageProps['error'] = () => h(ImageErrorIcon, { size: '24' });

// 渲染加载中图标
const renderLoadingIcon: ImageProps['loading'] = () => h(LoadingIcon, { size: '24', class: 'loading-icon' });
</script>

<style lang="less" scoped>
/* 加载中图标旋转动画 */
:deep(.loading-icon) {
  animation: rotate 1s linear infinite;
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

/* 确保图片组件内部元素居中 */
:deep(.t-image) {
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 无图片占位符样式 */
.no-image-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--td-bg-color-component);
  color: var(--td-text-color-placeholder);
  
  &.shape-circle {
    border-radius: 50%;
  }
  
  &.shape-round {
    border-radius: 6px;
  }
  
  &.shape-square {
    border-radius: 0;
  }
}
</style>

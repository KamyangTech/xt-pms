<template>
  <div class="mold-image-gallery">
    <div v-if="images.length === 0" class="empty-state">
      <t-icon name="image" size="48px" />
      <p>暂无图片</p>
    </div>

    <div v-else class="image-grid">
      <div
        v-for="(image, index) in images"
        :key="image.id"
        class="image-card"
        @click="handlePreview(index)"
      >
        <div class="image-wrapper">
          <img :src="getFileUrl(image.url)" :alt="image.file_name" />
          <div v-if="image.is_primary === 1" class="primary-badge">主图</div>
        </div>
        <div class="image-info">
          <span class="file-name" :title="image.file_name">{{ image.file_name }}</span>
          <span class="file-size">{{ formatFileSize(image.file_size) }}</span>
        </div>
      </div>
    </div>

    <!-- 图片预览 -->
    <t-image-viewer
      v-model:visible="previewVisible"
      :images="previewImages"
      :default-index="previewIndex"
      @close="handlePreviewClose"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import type { MoldFile } from '@/types/mold';
import { getFileFullUrl } from '@/api/common/upload';

interface Props {
  images?: MoldFile[];
}

const props = withDefaults(defineProps<Props>(), {
  images: () => [],
});

const previewVisible = ref(false);
const previewIndex = ref(0);

const previewImages = computed(() => {
  return props.images.map((img) => getFileUrl(img.url));
});

function getFileUrl(url: string) {
  return getFileFullUrl(url);
}

// 格式化文件大小
function formatFileSize(size: number) {
  if (size < 1024) {
    return `${size} B`;
  } else if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(2)} KB`;
  } else {
    return `${(size / (1024 * 1024)).toFixed(2)} MB`;
  }
}

// 预览图片
function handlePreview(index: number) {
  previewIndex.value = index;
  previewVisible.value = true;
}

// 关闭预览
function handlePreviewClose() {
  previewVisible.value = false;
}
</script>

<style scoped>
.mold-image-gallery {
  width: 100%;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 48px;
  color: var(--td-text-color-secondary);
}

.empty-state p {
  margin-top: 16px;
  font-size: 14px;
}

.image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 16px;
}

.image-card {
  cursor: pointer;
  transition: transform 0.3s;
}

.image-card:hover {
  transform: translateY(-4px);
}

.image-wrapper {
  position: relative;
  width: 100%;
  padding-bottom: 100%;
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid var(--td-border-level-1-color);
  background: #f5f5f5;
}

.image-wrapper img {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s;
}

.image-card:hover .image-wrapper img {
  transform: scale(1.05);
}

.primary-badge {
  position: absolute;
  top: 8px;
  left: 8px;
  padding: 2px 8px;
  background: var(--td-brand-color);
  color: #fff;
  font-size: 12px;
  border-radius: 2px;
  z-index: 1;
}

.image-info {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.file-name {
  font-size: 12px;
  color: var(--td-text-color-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-size {
  font-size: 12px;
  color: var(--td-text-color-secondary);
}
</style>

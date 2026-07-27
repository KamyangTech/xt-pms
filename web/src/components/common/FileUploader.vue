<template>
  <div class="file-uploader" :class="{ 'file-uploader--inline': inline }">
    <!-- ====== INLINE 紧凑模式（适配表格单元格） ====== -->
    <template v-if="inline">
      <!-- 有文件时显示预览 -->
      <div v-if="displayFiles.length > 0" class="inline-file-item">
        <div
          v-for="(item, index) in displayFiles"
          :key="item.key"
          class="inline-file-preview"
          :class="{ 'inline-file-preview--pending': item.isPending }"
        >
          <t-image-viewer v-if="item.isImage && previewableImages.length > 0" :images="previewableImages" :default-index="getImageIndex(index)" :z-index="10000">
            <template #trigger="{ open }">
              <div class="inline-image-wrapper" @click="open(getImageIndex(index))">
                <t-image :src="item.previewUrl" fit="cover" shape="square" :style="{ width: '100%', height: '100%' }" />
                <div v-if="item.isPending" class="file-badge file-badge--pending">待上传</div>
                <!-- 操作覆盖层 -->
                <div v-if="showActions && !disabled" class="file-overlay" @click.stop>
                  <div class="overlay-actions">
                    <div class="action-btn" @click.stop="handleDelete(item, index)">
                      <delete-icon class="action-icon" />
                    </div>
                  </div>
                </div>
              </div>
            </template>
          </t-image-viewer>
          <div v-else class="inline-image-wrapper">
            <t-image :src="item.previewUrl" fit="cover" shape="square" :style="{ width: '100%', height: '100%' }" />
            <div v-if="showActions && !disabled" class="file-overlay" @click.stop>
              <div class="overlay-actions">
                <div class="action-btn" @click.stop="handleDelete(item, index)">
                  <delete-icon class="action-icon" />
                </div>
              </div>
            </div>
          </div>
        </div>
        <!-- 多图模式下额外上传入口 -->
        <t-upload
          v-if="multiple && displayFiles.length < maxCount"
          :auto-upload="false"
          :multiple="multiple"
          :accept="acceptAttr"
          :size-limit="sizeLimit"
          theme="custom"
          :disabled="disabled"
          @change="handleLocalFileSelect"
        >
          <div class="inline-upload-trigger">
            <add-icon class="inline-upload-icon" />
          </div>
        </t-upload>
      </div>

      <!-- 无文件时的上传占位：t-upload 包裹，点击即触发文件选择 -->
      <t-upload
        v-else
        :auto-upload="false"
        :multiple="multiple"
        :accept="acceptAttr"
        :size-limit="sizeLimit"
        theme="custom"
        :disabled="disabled"
        @change="handleLocalFileSelect"
      >
        <div class="inline-upload-placeholder" :class="{ 'is-disabled': disabled }">
          <div class="inline-upload-box">
            <upload-icon class="inline-upload-icon" />
            <div class="inline-upload-text">上传图片</div>
          </div>
        </div>
      </t-upload>
    </template>
    <!-- ====== 标准模式 ====== -->
    <template v-else>
    <!-- 上传按钮/区域 -->
    <div class="upload-area">
      <!-- 自动保存模式：直接上传到后端 -->
      <t-upload
        v-if="autoSave"
        ref="uploadRef"
        v-model="fileList"
        :action="actionUrl"
        :name="fileFieldName"
        :multiple="multiple"
        :max="maxCount - uploadedFiles.length"
        :size-limit="sizeLimit"
        :accept="acceptAttr"
        :disabled="disabled || !actionUrl"
        :headers="uploadHeaders"
        :drag="drag"
        theme="custom"
        @success="handleUploadSuccess"
        @fail="handleUploadFail"
        @remove="handleRemove"
      >
        <template #trigger>
          <slot name="trigger">
            <t-button theme="default" variant="outline" :disabled="disabled || (multiple ? uploadedFiles.length >= maxCount : uploadedFiles.length > 0)">
              <template #icon><upload-icon /></template>
              {{ uploadBtnText }}
            </t-button>
          </slot>
        </template>
      </t-upload>

      <!-- 延迟保存模式：本地选择 -->
      <t-upload
        v-else
        ref="uploadRef"
        :multiple="multiple"
        :max="maxCount - displayFiles.length"
        :size-limit="sizeLimit"
        :accept="acceptAttr"
        :auto-upload="false"
        :disabled="disabled"
        :drag="drag"
        theme="custom"
        @change="handleLocalFileSelect"
      >
        <template #trigger>
          <slot name="trigger">
            <div v-if="drag" class="upload-drag-area">
              <upload-icon class="drag-icon" />
              <p class="drag-text">{{ uploadBtnText }}</p>
              <p class="drag-hint">{{ dragHint }}</p>
            </div>
            <t-button v-else theme="default" variant="outline" :disabled="disabled">
              <template #icon><upload-icon /></template>
              {{ uploadBtnText }}
            </t-button>
          </slot>
        </template>
      </t-upload>

      <!-- 提示信息 -->
      <span v-if="!autoSave && hasPendingChanges" class="upload-tip">文件已暂存，请点击保存按钮提交</span>
      <span v-if="!autoSave && !hasPendingChanges && showHint" class="upload-hint">{{ hint }}</span>
    </div>

    <!-- 文件列表 -->
    <div v-if="displayFiles.length > 0" class="file-grid" :class="{ 'is-image-grid': isImageOnly }">
      <div
        v-for="(item, index) in displayFiles"
        :key="item.key"
        class="file-item"
        :class="{
          'file-item--pending': item.isPending,
          'file-item--image': item.isImage,
          'file-item--doc': !item.isImage && !item.isPdf,
          'file-item--pdf': item.isPdf,
        }"
      >
        <!-- 图片预览 -->
        <template v-if="item.isImage">
          <t-image-viewer v-if="previewableImages.length > 0" :images="previewableImages" :default-index="getImageIndex(index)" :z-index="10000">
            <template #trigger="{ open }">
              <div class="file-preview" @click="open(getImageIndex(index))">
                <t-image
                  :src="item.previewUrl"
                  fit="cover"
                  :style="{ width: '100%', height: '100%' }"
                  shape="square"
                />
                <!-- 主图标识 -->
                <div v-if="item.isPrimary" class="file-badge file-badge--primary">主图</div>
                <!-- 待上传标识 -->
                <div v-if="item.isPending" class="file-badge file-badge--pending">待上传</div>
              </div>
              <!-- 操作覆盖层（图片：使用 t-image-viewer 预览） -->
              <div v-if="showActions && !disabled" class="file-overlay" @click.stop>
                <div class="overlay-actions">
                  <div class="action-btn" @click.stop="open(getImageIndex(index))">
                    <t-icon name="browse" class="action-icon" />
                  </div>
                  <div
                    v-if="supportPrimary && !item.isPrimary && !item.isPending"
                    class="action-btn"
                    @click.stop="handleSetPrimary(item)"
                  >
                    <star-icon class="action-icon" />
                  </div>
                  <div class="action-btn action-btn--danger" @click.stop="handleDelete(item, index)">
                    <delete-icon class="action-icon" />
                  </div>
                </div>
              </div>
            </template>
          </t-image-viewer>
          <div v-else class="file-preview">
            <t-image
              :src="item.previewUrl"
              fit="cover"
              :style="{ width: '100%', height: '100%' }"
              shape="square"
            />
          </div>
        </template>

        <!-- PDF 预览 -->
        <template v-else-if="item.isPdf">
          <div class="file-preview file-preview--pdf" @click="openPdf(item)">
            <FilePdfIcon class="file-type-icon" />
            <div class="file-name-text">{{ item.fileName }}</div>
            <div class="file-size-text">{{ formatSize(item.fileSize) }}</div>
          </div>
          <!-- 操作覆盖层（PDF/文档：在新标签页打开） -->
          <div v-if="showActions && !disabled" class="file-overlay">
            <div class="overlay-actions">
              <div v-if="!item.isPending" class="action-btn" @click.stop="openFileUrl(item)">
                <t-icon name="browse" class="action-icon" />
              </div>
              <div class="action-btn action-btn--danger" @click.stop="handleDelete(item, index)">
                <delete-icon class="action-icon" />
              </div>
            </div>
          </div>
        </template>

        <!-- 文档/压缩包等 -->
        <template v-else>
          <div class="file-preview file-preview--doc" @click="openFileUrl(item)">
            <component :is="getFileIconComponent(item.fileType)" class="file-type-icon" />
            <div class="file-name-text">{{ item.fileName }}</div>
            <div class="file-size-text">{{ formatSize(item.fileSize) }}</div>
          </div>
          <!-- 操作覆盖层（PDF/文档：在新标签页打开） -->
          <div v-if="showActions && !disabled" class="file-overlay">
            <div class="overlay-actions">
              <div v-if="!item.isPending" class="action-btn" @click.stop="openFileUrl(item)">
                <t-icon name="browse" class="action-icon" />
              </div>
              <div class="action-btn action-btn--danger" @click.stop="handleDelete(item, index)">
                <delete-icon class="action-icon" />
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- 空状态提示 -->
    <div v-else-if="showEmptyHint" class="file-empty">
      <t-icon name="file" class="empty-icon" />
      <p class="empty-text">{{ emptyText }}</p>
    </div>
    </template>

    <!-- PDF 预览弹窗 -->
    <t-dialog
      v-model:visible="pdfPreviewVisible"
      :header="pdfPreviewTitle"
      placement="center"
      width="85vw"
      :footer="false"
      destroy-on-close
     :close-on-overlay-click="false">
      <div class="pdf-preview-wrapper">
        <iframe :src="pdfPreviewUrl" class="pdf-preview-iframe" frameborder="0" />
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed, onMounted, onUnmounted } from 'vue';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import {
  UploadIcon,
  DeleteIcon,
  StarIcon,
  AddIcon,
  FilePdfIcon,
  FileWordIcon,
  FileExcelIcon,
  FileIcon as FileDefaultIcon,
} from 'tdesign-icons-vue-next';
import { uploadFiles, isImageType, formatFileSize, getFileFullUrl, revokeFilePreviewUrl, createFilePreviewUrl, isPreviewable, downloadFile, getPdfPreviewUrl, type UploadFileResult } from '@/api/common/upload';

// ---------- Props ----------
interface Props {
  /** v-model：已上传的文件列表 */
  modelValue?: UploadedFileItem[];
  /** 自动保存模式 */
  autoSave?: boolean;
  /** 是否多文件上传 */
  multiple?: boolean;
  /** 最大文件数量 */
  maxCount?: number;
  /** 是否显示拖拽上传区域 */
  drag?: boolean;
  /** 是否禁用 */
  disabled?: boolean;
  /** 单个文件大小限制 (MB) */
  maxSizeMB?: number;
  /** 接受的文件类型，例: "image/*,.pdf,.zip,.rar" */
  accept?: string;
  /** 上传 URL（自动保存模式） */
  uploadUrl?: string;
  /** 上传字段名 */
  fileFieldName?: string;
  /** 业务分类（如 mold-images, product-images, expense-attachments） */
  category?: string;
  /** 上传按钮文字 */
  uploadBtnText?: string;
  /** 是否支持主图设置 */
  supportPrimary?: boolean;
  /** 是否显示操作按钮覆盖层 */
  showActions?: boolean;
  /** 是否显示空状态提示 */
  showEmptyHint?: boolean;
  /** 空状态文字 */
  emptyText?: string;
  /** 是否显示大小/类型提示文字 */
  showHint?: boolean;
  /** 提示文字 */
  hint?: string;
  /** 拖拽区域的提示文字 */
  dragHint?: string;
  /** 是否只接受图片 */
  imageOnly?: boolean;
  /** inline 紧凑模式：120x120 方块样式，适合表格单元格 */
  inline?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: () => [],
  autoSave: true,
  multiple: false,
  maxCount: 9,
  drag: false,
  disabled: false,
  maxSizeMB: 5,
  accept: '',
  uploadUrl: '',
  fileFieldName: 'files[]',
  category: 'general',
  uploadBtnText: '上传文件',
  supportPrimary: false,
  showActions: true,
  showEmptyHint: false,
  emptyText: '暂无文件',
  showHint: true,
  hint: '',
  dragHint: '支持拖拽上传',
  imageOnly: false,
  inline: false,
});

const emit = defineEmits<{
  (e: 'update:modelValue', value: UploadedFileItem[]): void;
  (e: 'change', value: { uploaded: UploadedFileItem[]; pending: PendingFileItem[]; hasChanges: boolean }): void;
  (e: 'files-uploaded', files: UploadedFileItem[]): void;
}>();

// ---------- 内部类型 ----------
export interface UploadedFileItem {
  id?: string;
  key: string;
  url: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  isPrimary?: boolean;
}

export interface PendingFileItem {
  key: string;
  file: File;
  previewUrl: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  isImage: boolean;
  isPdf: boolean;
}

export interface DisplayFileItem {
  key: string;
  previewUrl: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  isPrimary: boolean;
  isPending: boolean;
  isImage: boolean;
  isPdf: boolean;
  source: 'uploaded' | 'pending';
  uploadedItem?: UploadedFileItem;
  pendingItem?: PendingFileItem;
}

// ---------- 状态 ----------
const uploadRef = ref();
const fileList = ref<any[]>([]);
const pendingFiles = ref<PendingFileItem[]>([]);
const pendingDeleteIndices = ref<number[]>([]);
const pendingPrimaryKey = ref<string | null>(null);
// PDF 预览弹窗
const pdfPreviewVisible = ref(false);
const pdfPreviewUrl = ref('');
const pdfPreviewTitle = ref('');
const pdfPreviewIsBlob = ref(false);

// ---------- 计算属性 ----------
/** 已上传文件列表 */
const uploadedFiles = computed<UploadedFileItem[]>(() => props.modelValue);

/** accept 属性 */
const acceptAttr = computed(() => {
  if (props.accept) return props.accept;
  if (props.imageOnly) return 'image/*';
  return 'image/*,.pdf,.doc,.docx,.xls,.xlsx,.zip,.rar,.7z,.tar,.gz';
});

/** 大小限制 */
const sizeLimit = computed(() => ({
  size: props.maxSizeMB,
  unit: 'MB' as const,
}));

/** 上传 URL */
const actionUrl = computed(() => {
  if (props.uploadUrl) return props.uploadUrl;
  if (props.autoSave) return '/api/upload/files';
  return '';
});

/** 上传请求头 */
const uploadHeaders = computed(() => {
  const token = localStorage.getItem('token');
  return token ? { 'Authorization': `Bearer ${token}` } : {};
});

/** 是否有待处理的变更 */
const hasPendingChanges = computed(() => {
  return pendingFiles.value.length > 0 || pendingDeleteIndices.value.length > 0 || pendingPrimaryKey.value !== null;
});

/** 是否纯图片模式 */
const isImageOnly = computed(() => props.imageOnly);

/** 可预览的图片链接列表 */
const previewableImages = computed(() => {
  return displayFiles.value
    .filter((item) => item.isImage && !item.isPending)
    .map((item) => item.previewUrl);
});

/** 合并显示的文件列表 */
const displayFiles = computed<DisplayFileItem[]>(() => {
  const result: DisplayFileItem[] = [];
  const deletedKeys = new Set(pendingDeleteIndices.value.map((i) => `uploaded-${i}`));

  // 已上传文件
  uploadedFiles.value.forEach((item, index) => {
    if (deletedKeys.has(`uploaded-${index}`)) return;
    const key = `uploaded-${index}`;
    const fullUrl = getFileFullUrl(item.url || '');
    const ext = getFileExt(item.fileName);
    // pendingPrimaryKey 优先级高于 item.isPrimary，立即反映主图变更
    const isPrimary = pendingPrimaryKey.value !== null
      ? key === pendingPrimaryKey.value
      : item.isPrimary || false;
    result.push({
      key,
      previewUrl: fullUrl,
      fileName: item.fileName || '',
      fileSize: item.fileSize || 0,
      fileType: ext,
      isPrimary,
      isPending: false,
      isImage: isImageType(ext),
      isPdf: ext === '.pdf',
      source: 'uploaded',
      uploadedItem: item,
    });
  });

  // 待上传文件
  pendingFiles.value.forEach((item) => {
    result.push({
      key: item.key,
      previewUrl: item.previewUrl,
      fileName: item.fileName,
      fileSize: item.fileSize,
      fileType: item.fileType,
      isPrimary: false,
      isPending: true,
      isImage: item.isImage,
      isPdf: item.isPdf,
      source: 'pending',
      pendingItem: item,
    });
  });

  // 主图排在第一个
  const primaryIdx = result.findIndex((f) => f.isPrimary);
  if (primaryIdx > 0) {
    const primary = result.splice(primaryIdx, 1)[0];
    result.unshift(primary);
  }

  return result;
});

// ---------- 方法 ----------
/** 获取文件扩展名（防御性检查：fileName 可能为 undefined/null） */
function getFileExt(fileName: string | undefined | null): string {
  if (!fileName || typeof fileName !== 'string') return '';
  const idx = fileName.lastIndexOf('.');
  return idx >= 0 ? fileName.substring(idx).toLowerCase() : '';
}

/** 格式化文件大小 */
function formatSize(bytes: number): string {
  return formatFileSize(bytes);
}

/** 获取图标组件 */
function getFileIconComponent(fileType: string) {
  const ext = fileType.toLowerCase();
  if (['.pdf'].includes(ext)) return FilePdfIcon;
  if (['.doc', '.docx'].includes(ext)) return FileWordIcon;
  if (['.xls', '.xlsx'].includes(ext)) return FileExcelIcon;
  // 压缩包等使用默认文件图标
  return FileDefaultIcon;
}

/** 获取图片在预览数组中的索引 */
function getImageIndex(currentIndex: number): number {
  let imgIdx = 0;
  for (let i = 0; i < currentIndex; i++) {
    if (displayFiles.value[i]?.isImage && !displayFiles.value[i]?.isPending) {
      imgIdx++;
    }
  }
  return imgIdx;
}

/** 打开文件 URL */
async function openFileUrl(item: DisplayFileItem) {
  if (item.isPending) return;
  if (item.isPdf) {
    openPdf(item);
    return;
  }
  // 非图片文件：弹窗确认后下载
  if (!item.isImage && item.uploadedItem?.id) {
    const confirmDialog = DialogPlugin.confirm({
      header: '确认下载',
      body: `确定要下载文件「${item.fileName}」吗？`,
      confirmBtn: '下载',
      cancelBtn: '取消',
      onConfirm: async () => {
        confirmDialog.hide();
        try {
          await downloadFile(item.uploadedItem!.id, item.fileName);
        } catch (error: any) {
          MessagePlugin.error(error.message || '文件下载失败');
        }
      },
      onCancel: () => confirmDialog.hide(),
      onClose: () => confirmDialog.destroy(),
    });
    return;
  }
  window.open(item.previewUrl, '_blank');
}

/** 打开 PDF 预览弹窗 */
async function openPdf(item: DisplayFileItem) {
  if (pdfPreviewIsBlob.value && pdfPreviewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(pdfPreviewUrl.value);
  }
  pdfPreviewIsBlob.value = false;
  if (!item.uploadedItem?.id) {
    pdfPreviewUrl.value = item.previewUrl;
    pdfPreviewTitle.value = item.fileName;
    pdfPreviewVisible.value = true;
    return;
  }
  try {
    const blobUrl = await getPdfPreviewUrl(item.uploadedItem.id);
    pdfPreviewUrl.value = blobUrl;
    pdfPreviewIsBlob.value = true;
    pdfPreviewTitle.value = item.fileName;
    pdfPreviewVisible.value = true;
  } catch (error: any) {
    MessagePlugin.error(error.message || 'PDF 预览失败');
  }
}

/** 延迟保存模式：本地文件选择 */
function handleLocalFileSelect(context: any) {
  let files: any[] = [];
  if (Array.isArray(context)) {
    files = context;
  } else if (typeof context === 'object' && context !== null) {
    files = context.currentFiles || context.files || [];
    if (files.length === 0 && context.currentFile) {
      files = [context.currentFile];
    }
  }

  if (!files || files.length === 0) return;

  // 先提取有效的 File 实例，防止非文件事件错误触发单图清理逻辑
  const validFiles: File[] = [];
  files.forEach((file: any) => {
    const actualFile = file.raw || file;
    if (actualFile && actualFile instanceof File) {
      validFiles.push(actualFile);
    }
  });

  if (validFiles.length === 0) return;

  // 单图模式：选择新文件时自动清空旧文件，只保留新选的文件
  if (!props.multiple && props.maxCount <= 1) {
    pendingFiles.value.forEach((pf) => revokeFilePreviewUrl(pf.previewUrl));
    pendingFiles.value = [];
    // 已有已上传图片也标记为待删除
    if (uploadedFiles.value.length > 0) {
      uploadedFiles.value.forEach((_, i) => {
        if (!pendingDeleteIndices.value.includes(i)) {
          pendingDeleteIndices.value.push(i);
        }
      });
    }
  }

  validFiles.forEach((actualFile: File) => {
    const ext = getFileExt(actualFile.name);
    const previewUrl = createFilePreviewUrl(actualFile);

    pendingFiles.value.push({
      key: `pending-${Date.now()}-${Math.random()}`,
      file: actualFile,
      previewUrl,
      fileName: actualFile.name,
      fileSize: actualFile.size,
      fileType: ext,
      isImage: isImageType(ext),
      isPdf: ext === '.pdf',
    });
  });

  emitChange();
}

/** 自动保存模式：上传成功回调 */
async function handleUploadSuccess(context: any) {
  const { response, file } = context;
  const responseData = response?.data || response;

  if (responseData) {
    // 响应可能是数组（多文件）或对象（单文件）
    const results = Array.isArray(responseData) ? responseData : [responseData];
    const newItems = results.map((r: any) => ({
      id: r.id || '',
      key: `uploaded-${Date.now()}-${Math.random()}`,
      url: r.url,
      fileName: r.file_name || file.name,
      fileSize: r.file_size || file.size,
      fileType: r.file_type || getFileExt(r.file_name || file.name),
    }));

    fileList.value = [];

    emit('update:modelValue', [...uploadedFiles.value, ...newItems]);
    emit('files-uploaded', newItems);
    emitChange();
    MessagePlugin.success(`${newItems.length} 个文件上传成功`);
  } else {
    MessagePlugin.error(`${file.name} 上传失败`);
  }
}

/** 自动保存模式：上传失败 */
function handleUploadFail(context: any) {
  const { file } = context;
  MessagePlugin.error(`${file.name} 上传失败`);
}

/** 自动保存模式：移除文件 */
function handleRemove(context: any) {
  // 由 handleDelete 统一处理
}

/** 设置主图 */
async function handleSetPrimary(item: DisplayFileItem) {
  if (props.disabled) return;
  if (item.isPending) {
    MessagePlugin.warning('请先上传图片后再设置主图');
    return;
  }

  if (props.autoSave) {
    // 自动保存模式：直接更新主图标记
    const newList = uploadedFiles.value.map((f) => ({
      ...f,
      isPrimary: f.key === item.key,
    }));
    emit('update:modelValue', newList);
    emitChange();
    MessagePlugin.success('已设置为主图');
  } else {
    // 延迟保存模式
    pendingPrimaryKey.value = item.key;
    emitChange();
    MessagePlugin.success('已标记为主图，请点击保存按钮提交');
  }
}

/** 删除文件 */
function handleDelete(item: DisplayFileItem, index: number) {
  if (props.disabled) return;

  const displayName = item.fileName || '此文件';
  const isPending = item.isPending;

  const confirmDialog = DialogPlugin.confirm({
    header: '确认删除',
    body: `确定要删除 "${displayName}" 吗？${isPending ? '此操作将取消上传' : '此操作不可恢复'}`,
    confirmBtn: isPending ? '取消上传' : '确定删除',
    cancelBtn: '取消',
    onConfirm: () => {
      if (isPending) {
        // 清除待上传文件
        const idx = pendingFiles.value.findIndex((pf) => pf.key === item.key);
        if (idx !== -1) {
          revokeFilePreviewUrl(pendingFiles.value[idx].previewUrl);
          pendingFiles.value.splice(idx, 1);
        }
      } else if (props.autoSave) {
        // 自动保存模式：直接更新 modelValue
        const newList = uploadedFiles.value.filter((_, i) => `uploaded-${i}` !== item.key);
        emit('update:modelValue', newList);
      } else {
        // 延迟保存模式：标记为待删除
        const idx = uploadedFiles.value.findIndex((_, i) => `uploaded-${i}` === item.key);
        if (idx !== -1 && !pendingDeleteIndices.value.includes(idx)) {
          pendingDeleteIndices.value.push(idx);
        }
      }
      emitChange();
      MessagePlugin.success(isPending ? '已取消上传' : '删除成功');
      confirmDialog.hide();
    },
    onCancel: () => confirmDialog.hide(),
    onClose: () => confirmDialog.destroy(),
  });
}

/** 触发变更事件 */
function emitChange() {
  emit('change', {
    uploaded: uploadedFiles.value,
    pending: pendingFiles.value,
    hasChanges: hasPendingChanges.value,
  });
}

// ---------- 重置与提交 ----------
/** 重置所有待处理的变更 */
function resetPendingChanges() {
  pendingFiles.value.forEach((pf) => revokeFilePreviewUrl(pf.previewUrl));
  pendingFiles.value = [];
  pendingDeleteIndices.value = [];
  pendingPrimaryKey.value = null;
}

/** 获取待处理的变更数据 */
function getPendingChanges() {
  return {
    uploads: pendingFiles.value.map((pf) => pf.file),
    deleteIndices: [...pendingDeleteIndices.value],
    primaryKey: pendingPrimaryKey.value,
    hasChanges: hasPendingChanges.value,
  };
}

/** 提交待处理的上传文件（批量上传） */
async function submitPendingChanges(): Promise<UploadFileResult[]> {
  if (pendingFiles.value.length === 0) {
    MessagePlugin.info('没有需要上传的文件');
    return [];
  }

  try {
    const files = pendingFiles.value.map((pf) => pf.file);
    const result = await uploadFiles({
      files,
      category: props.category,
    });

    const newUploaded = result.map((r, i) => ({
      id: r.id,
      key: `uploaded-${Date.now()}-${i}`,
      url: r.url,
      fileName: r.file_name,
      fileSize: r.file_size,
      fileType: r.file_type,
    }));

    let finalList = [...uploadedFiles.value];
    if (pendingDeleteIndices.value.length > 0) {
      finalList = finalList.filter((_, i) => !pendingDeleteIndices.value.includes(i));
    }
    finalList = [...finalList, ...newUploaded];

    // 主图处理
    if (pendingPrimaryKey.value) {
      finalList = finalList.map((f) => ({
        ...f,
        isPrimary: f.key === pendingPrimaryKey.value,
      }));
    }

    emit('update:modelValue', finalList);
    emitChange();
    resetPendingChanges();
    MessagePlugin.success(`成功上传 ${newUploaded.length} 个文件`);
    return result;
  } catch (error: any) {
    MessagePlugin.error(error.message || '文件上传失败');
    throw error;
  }
}

// ---------- 生命周期 ----------
onMounted(() => {
  if (!props.autoSave) {
    resetPendingChanges();
  }
});

onUnmounted(() => {
  pendingFiles.value.forEach((pf) => revokeFilePreviewUrl(pf.previewUrl));
});

// ---------- 监听外部 modelValue ----------
watch(
  () => props.modelValue,
  () => {
    // 外部变更时不重置 pending 状态，保留 pending 操作
  },
);

// ---------- 监听 PDF 预览弹窗 ----------
watch(pdfPreviewVisible, (visible) => {
  if (!visible && pdfPreviewIsBlob.value && pdfPreviewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(pdfPreviewUrl.value);
    pdfPreviewUrl.value = '';
    pdfPreviewIsBlob.value = false;
  }
});

// ---------- 暴露方法 ----------
defineExpose({
  resetPendingChanges,
  getPendingChanges,
  submitPendingChanges,
});
</script>

<style scoped>
.file-uploader {
  width: 100%;
}

/* ---------- 上传区域 ---------- */
.upload-area {
  display: flex;
  align-items: center;
  margin-bottom: var(--td-comp-margin-m);
}

.upload-area :deep(.t-upload__trigger) {
  display: block;
}

.upload-tip {
  display: inline-block;
  margin-left: var(--td-comp-margin-s);
  font-size: var(--td-font-size-body-medium);
  color: var(--td-warning-color);
  font-weight: bold;
}

.upload-hint {
  display: inline-block;
  margin-left: var(--td-comp-margin-s);
  font-size: var(--td-font-size-body-small);
  color: var(--td-text-color-placeholder);
}

/* 拖拽区域 */
.upload-drag-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  border: 2px dashed var(--td-border-level-2-color);
  border-radius: var(--td-radius-medium);
  background-color: var(--td-bg-color-secondarycontainer);
  cursor: pointer;
  transition: all 0.2s;
}

.upload-drag-area:hover {
  border-color: var(--td-brand-color);
  background-color: var(--td-bg-color-container-hover);
}

.drag-icon {
  font-size: 48px;
  color: var(--td-text-color-placeholder);
  margin-bottom: 16px;
}

.drag-text {
  font-size: var(--td-font-size-body-large);
  color: var(--td-text-color-primary);
  margin-bottom: 8px;
}

.drag-hint {
  font-size: var(--td-font-size-body-small);
  color: var(--td-text-color-placeholder);
}

/* 隐藏 TDesign Upload 组件的文件列表 */
.upload-area :deep(.t-upload__list-container) {
  display: none !important;
}

/* ---------- 文件网格 ---------- */
.file-grid {
  display: flex;
  flex-wrap: wrap;
  gap: var(--td-comp-margin-m);
}

.file-grid.is-image-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
}

/* ---------- 文件项 ---------- */
.file-item {
  position: relative;
  border-radius: var(--td-radius-medium);
  overflow: hidden;
  border: 1px solid var(--td-border-level-1-color);
  background-color: var(--td-bg-color-secondarycontainer);
  width: 140px;
  height: 140px;
}

.file-item--pending {
  border: 2px dashed var(--td-brand-color);
  opacity: 0.85;
}

.file-item--doc,
.file-item--pdf {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

/* ---------- 文件预览 ---------- */
.file-preview {
  position: relative;
  width: 100%;
  height: 100%;
  cursor: pointer;
}

.file-preview--pdf,
.file-preview--doc {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 12px;
  text-align: center;
}

.file-type-icon {
  font-size: 40px;
  color: var(--td-text-color-placeholder);
  margin-bottom: 8px;
}

.file-name-text {
  font-size: var(--td-font-size-body-small);
  color: var(--td-text-color-primary);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
}

.file-size-text {
  font-size: var(--td-font-size-body-small);
  color: var(--td-text-color-placeholder);
}

/* ---------- 徽标 ---------- */
.file-badge {
  position: absolute;
  top: 6px;
  left: 6px;
  padding: 2px 6px;
  font-size: var(--td-font-size-body-small);
  border-radius: var(--td-radius-round);
  z-index: 2;
  line-height: 1.4;
}

.file-badge--primary {
  background-color: var(--td-brand-color);
  color: white;
}

.file-badge--pending {
  background-color: var(--td-warning-color);
  color: white;
}

/* ---------- 操作覆盖层 ---------- */
.file-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  opacity: 0;
  transition: opacity 0.2s ease;
  z-index: 1;
}

.file-item:hover .file-overlay,
.inline-file-preview:hover .file-overlay,
.inline-image-wrapper:hover .file-overlay {
  opacity: 1;
}

.overlay-actions {
  position: absolute;
  bottom: 10px;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  gap: 2px;
  align-items: center;
  z-index: 2;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 20px;
  backdrop-filter: blur(10px);
  overflow: hidden;
}

.action-btn {
  width: 34px;
  height: 34px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background-color 0.2s;
}

.action-btn:hover {
  background-color: rgba(255, 255, 255, 0.9);
}

.action-btn--danger:hover {
  background-color: rgba(255, 80, 80, 0.9);
}

.action-icon {
  font-size: 16px;
  color: rgba(255, 255, 255, 0.9);
  transition: color 0.2s;
}

.action-btn:hover .action-icon {
  color: var(--td-text-color-primary);
}

.action-btn--danger:hover .action-icon {
  color: white;
}

/* ---------- 空状态 ---------- */
.file-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  border: 1px dashed var(--td-border-level-1-color);
  border-radius: var(--td-radius-medium);
}

.empty-icon {
  font-size: 48px;
  color: var(--td-text-color-disabled);
  margin-bottom: 8px;
}

.empty-text {
  font-size: var(--td-font-size-body-medium);
  color: var(--td-text-color-disabled);
}

/* ===== inline 紧凑模式 ===== */
.file-uploader--inline {
  display: inline-flex;
  align-items: center;
}

.inline-upload-placeholder {
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
}

.inline-upload-placeholder.is-disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.inline-upload-box {
  width: 120px;
  height: 120px;
  border: 1px dashed var(--td-border-level-1-color);
  border-radius: var(--td-radius-medium);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  background-color: var(--td-bg-color-secondarycontainer);
  transition: all 0.2s;
}

.inline-upload-placeholder:not(.is-disabled) .inline-upload-box:hover {
  border-color: var(--td-brand-color);
  background-color: var(--td-bg-color-container-hover);
}

.inline-upload-icon {
  font-size: 32px;
  color: var(--td-text-color-placeholder);
  margin-bottom: 8px;
}

.inline-upload-text {
  font-size: var(--td-font-size-body-small);
  color: var(--td-text-color-placeholder);
}

.inline-file-item {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.inline-file-preview {
  width: 120px;
  height: 120px;
  position: relative;
  border-radius: var(--td-radius-medium);
  overflow: hidden;
  border: 1px solid var(--td-border-level-1-color);
  background-color: var(--td-bg-color-secondarycontainer);
}

.inline-file-preview--pending {
  border: 2px dashed var(--td-brand-color);
  opacity: 0.85;
}

.inline-image-wrapper {
  position: relative;
  width: 100%;
  height: 100%;
  cursor: pointer;
}

.inline-upload-trigger {
  width: 120px;
  height: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed var(--td-border-level-1-color);
  border-radius: var(--td-radius-medium);
  cursor: pointer;
  background-color: var(--td-bg-color-secondarycontainer);
  transition: all 0.2s;
}

.inline-upload-trigger:hover {
  border-color: var(--td-brand-color);
  background-color: var(--td-bg-color-container-hover);
}

/* 隐藏 TDesign Upload 组件显示 */
.upload-hidden {
  display: none;
}

/* PDF 预览弹窗 */
.pdf-preview-wrapper {
  width: 100%;
  height: 78vh;
}
.pdf-preview-iframe {
  width: 100%;
  height: 100%;
  border-radius: var(--td-radius-medium);
}
</style>

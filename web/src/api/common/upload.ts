import { request } from '@/utils/request';

/** 上传文件响应 */
export interface UploadFileResult {
  id: string;
  url: string;
  file_name: string;
  file_size: number;
  file_type: string;
}

/** 上传文件参数 */
export interface UploadFilesParams {
  /** 文件列表 */
  files: File[];
  /** 业务分类，用于决定存储子目录（如 mold-images, product-images, expense-attachments） */
  category?: string;
  /** 上传进度回调 */
  onProgress?: (percent: number) => void;
}

/** 待上传的本地文件（用于延迟保存模式） */
export interface PendingFile {
  file: File;
  previewUrl: string;
  fileName: string;
  fileSize: number;
  fileType: string;
  key: string;
}

/** 已上传的文件信息 */
export interface UploadedFileInfo {
  id: string;
  url: string;
  file_name: string;
  file_size: number;
  file_type: string;
}

/**
 * 通用多文件上传
 * 支持图片(png/jpg/gif/webp)、文档(pdf/doc/docx/xls/xlsx)、压缩包(zip/rar/7z/tar/gz)
 */
export function uploadFiles(params: UploadFilesParams): Promise<UploadFileResult[]> {
  const formData = new FormData();
  params.files.forEach((file) => {
    formData.append('files[]', file);
  });
  if (params.category) {
    formData.append('category', params.category);
  }
  return request.post({
    url: '/upload/files',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    // 上传请求单独设置超时（2分钟），避免大文件/多文件上传超时
    timeout: 120 * 1000,
    onUploadProgress: params.onProgress
      ? (progressEvent: any) => {
          const percent = Math.round((progressEvent.loaded / progressEvent.total) * 100);
          params.onProgress!(percent);
        }
      : undefined,
    // 上传请求不重试（大文件重试浪费带宽且大概率仍超时）
    requestOptions: {
      retry: { count: 0, delay: 0 },
    },
  } as any);
}

/**
 * 通用单图片上传（向后兼容）
 */
export function uploadImage(file: File): Promise<UploadFileResult> {
  const formData = new FormData();
  formData.append('image', file);
  return request.post({
    url: '/upload/image',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data',
    },
    timeout: 120 * 1000,
    requestOptions: {
      retry: { count: 0, delay: 0 },
    },
  } as any);
}

/**
 * 创建本地文件预览 Object URL
 */
export function createFilePreviewUrl(file: File): string {
  return URL.createObjectURL(file);
}

/**
 * 释放本地文件预览 Object URL
 */
export function revokeFilePreviewUrl(url: string) {
  URL.revokeObjectURL(url);
}

/**
 * 格式化文件大小
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const k = 1024;
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + units[i];
}

/**
 * 获取文件类型图标名称
 */
export function getFileTypeIcon(fileType: string): string {
  const ext = fileType.toLowerCase();
  if (['.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp', '.svg'].includes(ext)) {
    return 'image';
  }
  if (['.pdf'].includes(ext)) {
    return 'pdf';
  }
  if (['.doc', '.docx'].includes(ext)) {
    return 'word';
  }
  if (['.xls', '.xlsx'].includes(ext)) {
    return 'excel';
  }
  if (['.zip', '.rar', '.7z', '.tar', '.gz'].includes(ext)) {
    return 'archive';
  }
  return 'file';
}

/**
 * 判断是否为图片类型
 */
export function isImageType(fileType: string): boolean {
  return ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp', '.svg'].includes(fileType.toLowerCase());
}

/**
 * 获取图片完整访问 URL
 * 使用当前页面 origin 作为基础 URL，确保图片等静态资源通过同源地址访问，
 * 避免开发环境（Vite 代理）下跨域被 CORP 策略拦截。
 */
export function getFileFullUrl(url: string): string {
  if (!url) return '';
  if (url.startsWith('http')) return url;
  const baseUrl = typeof window !== 'undefined' ? window.location.origin : '';
  return `${baseUrl}${url}`;
}

/**
 * 判断是否为可预览的文件类型（图片和 PDF 支持预览）
 */
export function isPreviewable(fileType: string): boolean {
  const previewable = ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp', '.svg', '.pdf'];
  return previewable.includes(fileType.toLowerCase());
}

/**
 * 获取文件下载 URL（通过后端下载接口，会使用 Content-Disposition 返回原始文件名）
 * 适用于非图片文件（doc/xls/zip 等）
 */
export function getDownloadUrl(fileId: string): string {
  if (!fileId) return '';
  return `/api/files/${fileId}/download`;
}

/**
 * 带认证的下载文件
 * window.open 不会携带 Authorization header，所以下载接口会 401，
 * 这里用 fetch 带 token 拉取文件流，再用临时 <a> 触发浏览器下载
 */
export async function downloadFile(fileId: string, fileName: string): Promise<void> {
  if (!fileId) throw new Error('缺少文件ID');

  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : '';
  const response = await fetch(getDownloadUrl(fileId), {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });

  if (!response.ok) {
    const err = await response.json().catch(() => ({ message: '下载失败' }));
    throw new Error(err.message || '下载失败');
  }

  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = fileName || 'download';
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

/**
 * 获取 PDF 预览用 blob URL（带认证 fetch，避免开发环境静态文件代理问题）
 */
export async function getPdfPreviewUrl(fileId: string): Promise<string> {
  if (!fileId) throw new Error('缺少文件ID');

  const token = typeof window !== 'undefined' ? localStorage.getItem('token') : '';
  const response = await fetch(getDownloadUrl(fileId), {
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });

  if (!response.ok) {
    const err = await response.json().catch(() => ({ message: '预览失败' }));
    throw new Error(err.message || '预览失败');
  }

  const blob = await response.blob();
  return URL.createObjectURL(blob);
}

/**
 * 判断是否为图片类型（可通过 StaticFS 直接显示）
 */
export function isDisplayableImage(fileType: string): boolean {
  return isImageType(fileType);
}

/**
 * 根据原图 URL 推导缩略图 URL
 * 后端上传时自动生成 \`_thumb\` 后缀的缩略图文件，前端通过命名规则推导
 *
 * 规则：
 *   /uploads/ab/cd/uuid_name.jpg  →  /uploads/ab/cd/uuid_name_thumb.jpg
 *   /uploads/ab/cd/uuid_name.png  →  /uploads/ab/cd/uuid_name_thumb.png
 *
 * @param originalUrl 原图 URL
 * @returns 缩略图 URL，若 URL 为空或非图片则返回原 URL
 */
export function getThumbnailUrl(originalUrl: string): string {
  if (!originalUrl) return '';

  // 只对支持的图片格式生成缩略图 URL
  const thumbnailExtensions = ['.jpg', '.jpeg', '.png'];
  const lowerUrl = originalUrl.toLowerCase();
  const hasThumbnail = thumbnailExtensions.some((ext) => lowerUrl.endsWith(ext));

  if (!hasThumbnail) return originalUrl;

  // 在扩展名前插入 _thumb
  const lastDot = originalUrl.lastIndexOf('.');
  const base = originalUrl.substring(0, lastDot);
  const ext = originalUrl.substring(lastDot);
  return `${base}_thumb${ext}`;
}

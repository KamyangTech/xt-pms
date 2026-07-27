<template>
  <div v-if="visible" class="print-template-container">
    <!-- 打印标题 -->
    <div class="print-header">
      <p v-if="companyName" class="print-company-name">{{ companyName }}</p>
      <h1 class="print-title">{{ title }}</h1>
      <p v-if="subtitle" class="print-subtitle">{{ subtitle }}</p>
      <!-- 打印时间和制单操作员 - 在横线上面 -->
      <div class="print-header-info">
        <p class="print-time-header">打印时间: {{ formatDateTime(getCurrentDateTime()) }}</p>
        <p v-if="operator" class="print-operator-header">制单: {{ operator }}</p>
      </div>
    </div>

    <!-- 动态区块渲染 -->
    <div v-for="(section, sectionIndex) in sections" :key="sectionIndex" class="print-section">
      <!-- 区块标题 -->
      <h3 v-if="section.title" class="section-title">{{ section.title }}</h3>

      <!-- 信息表格类型 -->
      <table v-if="section.type === 'info-table' && section.fields" class="print-info-table">
        <colgroup>
          <col style="width: 12%" />
          <col style="width: 21%" />
          <col style="width: 12%" />
          <col style="width: 21%" />
          <col style="width: 12%" />
          <col style="width: 22%" />
        </colgroup>
        <tbody>
          <!-- 第一行：最多3个字段 -->
          <tr>
            <template v-for="(field, index) in section.fields" :key="field.key">
              <template v-if="index < 3">
                <td class="label-cell">{{ field.label }}</td>
                <td class="value-cell">{{ field.value }}</td>
              </template>
            </template>
            <td
              v-if="getFirstRowColspan(section.fields)"
              :colspan="getFirstRowColspan(section.fields) || undefined"
              class="empty-cell"
            ></td>
          </tr>
          <!-- 第二行：剩余字段（最多3组） -->
          <tr v-if="section.fields.length > 3">
            <template v-for="(field, index) in section.fields.slice(3)" :key="field.key">
              <template v-if="index < 3">
                <td class="label-cell">{{ field.label }}</td>
                <td class="value-cell">{{ field.value }}</td>
              </template>
            </template>
            <td
              v-if="getSecondRowColspan(section.fields)"
              :colspan="getSecondRowColspan(section.fields) || undefined"
              class="empty-cell"
            ></td>
          </tr>
        </tbody>
      </table>

      <!-- 数据表格类型 -->
      <table
        v-if="section.type === 'data-table' && section.columns && section.data"
        class="print-table"
      >
        <thead>
          <tr>
            <th
              v-for="col in section.columns"
              :key="col.key"
              :style="`width: ${col.width || 'auto'};`"
            >
              {{ col.label }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, rowIndex) in section.data" :key="rowIndex">
            <td v-for="col in section.columns" :key="col.key">
              <!-- 图片列渲染 -->
              <template v-if="col.imageUrl">
                <div class="print-image-cell">
                  <img
                    v-if="getImageUrl(row[col.imageUrl])"
                    :src="getImageUrl(row[col.imageUrl])"
                    class="print-product-image"
                    alt="产品图片"
                  />
                  <div v-else class="print-image-placeholder">无图片</div>
                </div>
              </template>
              <!-- 自定义渲染插槽 -->
              <slot v-else-if="col.slot" :name="col.slot" :row="row" :column="col">
                {{ row[col.key] || '-' }}
              </slot>
              <!-- 默认渲染 -->
              <template v-else>
                {{ row[col.key] || '-' }}
              </template>
            </td>
          </tr>
          <!-- 汇总行 -->
          <tr v-if="section.summary" class="summary-row">
            <td v-for="(col, colIndex) in section.columns" :key="colIndex">
              <template v-if="colIndex < section.summary.colspan">
                {{ colIndex === section.summary.colspan - 1 ? section.summary.label : '' }}
              </template>
              <template v-else-if="colIndex === section.summary.colspan">
                {{ section.summary.value }}
              </template>
            </td>
          </tr>
        </tbody>
      </table>

      <!-- 自定义内容类型 -->
      <div v-if="section.type === 'custom'" class="custom-content">
        <slot :name="`section-${sectionIndex}`"></slot>
      </div>
    </div>

    <!-- 签收栏（可选） -->
    <div v-if="showSignature" class="signature-section">
      <div class="signature-row">
        <div class="signature-item">
          <span class="stamp-label">{{ signatureLabels?.sender || '发货方签章' }}</span>
        </div>
        <div class="signature-item">
          <span class="stamp-label">{{ signatureLabels?.receiver || '收货方签章' }}</span>
        </div>
      </div>
      <div class="signature-row">
        <div class="signature-item date-item">
          <span class="signature-label">日 期</span>:<span class="date-underline"></span>
          <span class="date-text">年</span>
          <span class="date-underline"></span>
          <span class="date-text">月</span>
          <span class="date-underline"></span>
          <span class="date-text">日</span>
        </div>
        <div class="signature-item date-item">
          <span class="signature-label">日 期</span>:<span class="date-underline"></span>
          <span class="date-text">年</span>
          <span class="date-underline"></span>
          <span class="date-text">月</span>
          <span class="date-underline"></span>
          <span class="date-text">日</span>
        </div>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { formatDateTime, getCurrentDateTime } from '@/utils/date';

defineOptions({
  name: 'PrintTemplate',
});

withDefaults(defineProps<Props>(), {
  subtitle: '',
  showSignature: false,
  companyName: '',
  operator: '',
  signatureLabels: () => ({
    sender: '发货方签章',
    receiver: '收货方签章',
  }),
});

export interface PrintSection {
  type: 'info-table' | 'data-table' | 'custom';
  title?: string;
  fields?: Array<{ key: string; label: string; value: string }>;
  columns?: Array<{ key: string; label: string; width?: string; slot?: string; imageUrl?: string }>;
  data?: any[];
  summary?: { label: string; value: string | number; colspan: number };
}

interface Props {
  visible: boolean;
  title: string;
  subtitle?: string;
  sections: PrintSection[];
  showSignature?: boolean;
  companyName?: string;
  operator?: string; // 制单操作员
  signatureLabels?: {
    sender?: string;
    receiver?: string;
  };
}

const getImageUrl = (url: string | undefined | null) => {
  if (!url) return '';
  // 如果已经是完整URL（http开头），直接返回
  if (url.startsWith('http')) {
    return url;
  }
  // 开发环境使用代理时，返回相对路径让 Vite 代理处理（避免跨域和鉴权问题）
  const isProxy = import.meta.env.VITE_IS_REQUEST_PROXY === 'true';
  if (isProxy) {
    return url;
  }
  // 生产环境或其他情况拼接 API 基础地址
  const baseUrl = import.meta.env.VITE_API_URL || '';
  return `${baseUrl}${url.startsWith('/') ? url : `/${url}`}`;
};

const getFirstRowColspan = (fields: Array<{ key: string; label: string; value: string }>) => {
  const count = fields.length;
  const firstRowCount = Math.min(count, 3);
  const colspan = 6 - firstRowCount * 2;
  return colspan > 0 ? colspan : undefined;
};

const getSecondRowColspan = (fields: Array<{ key: string; label: string; value: string }>) => {
  if (fields.length <= 3) return undefined;
  const secondRowCount = Math.min(fields.length - 3, 3);
  const colspan = 6 - secondRowCount * 2;
  return colspan > 0 ? colspan : undefined;
};
</script>
<style lang="less" scoped>
/* 打印容器（屏幕下隐藏，但保留在 DOM 中以便图片提前加载） */
.print-template-container {
  position: fixed;
  left: -9999px;
  top: -9999px;
  width: 0;
  height: 0;
  overflow: hidden;
  opacity: 0;
  pointer-events: none;
}
</style>
<style lang="less">
/* 打印样式必须是全局的 */
@media print {
  @page {
    size: a4;
    margin: 0;
  }

  /* 启用背景图形打印（包括图片） */
  * {
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }

  /* 隐藏所有非打印元素 */
  body * {
    visibility: hidden;
  }

  /* 打印容器 */
  .print-template-container {
    display: block !important;
    visibility: visible !important;
    position: absolute !important;
    left: 0 !important;
    top: 0 !important;
    width: 100% !important;
    height: auto !important;
    overflow: visible !important;
    opacity: 1 !important;
    pointer-events: auto !important;
    padding: 40px 40px 0;
    margin: 0;
    font-family: 'Microsoft YaHei', SimSun, sans-serif;
    color: #000;
  }

  .print-template-container * {
    visibility: visible !important;
  }

  /* 顶部打印时间 */
  .print-template-container .print-time-top {
    display: none;
    font-size: 14px;
    color: #666;
    margin: 0 0 15px;
    text-align: left;
  }

  /* 打印头部 */
  .print-template-container .print-header {
    text-align: center;
    margin-bottom: 30px;
    border-bottom: 2px solid #000;
    padding-bottom: 20px;
    position: relative;
  }

  .print-template-container .print-header .print-company-name {
    margin: 0 0 8px;
    font-size: 18px;
    font-weight: bold;
    color: #000;
    line-height: 1.4;
  }

  .print-template-container .print-header .print-title {
    margin: 0 0 12px;
    font-size: 32px;
    font-weight: bold;
    color: #000;
    line-height: 1.2;
  }

  .print-template-container .print-header .print-subtitle {
    margin: 0;
    font-size: 16px;
    color: #333;
  }

  /* 打印头部信息区域（时间和操作员） */
  .print-template-container .print-header .print-header-info {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 5px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  /* 打印时间 - 在横线上面，靠左对齐 */
  .print-template-container .print-header .print-time-header {
    font-size: 14px;
    color: #666;
    margin: 0;
    text-align: left;
  }

  /* 制单操作员 - 在横线上面，靠右对齐 */
  .print-template-container .print-header .print-operator-header {
    font-size: 14px;
    color: #666;
    margin: 0;
    text-align: right;
  }

  /* 打印区块 */
  .print-template-container .print-section {
    margin-bottom: 20px;
    page-break-inside: avoid;
  }

  .print-template-container .print-section .section-title {
    font-size: 18px;
    font-weight: bold;
    color: #000;
    margin: 0 0 15px;
    padding: 8px 0;
  }

  /* 信息表格 */
  .print-template-container .print-info-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
    margin-top: 10px;
  }

  .print-template-container .print-info-table td {
    padding: 10px 8px;
    border: 1px solid #333;
    color: #000;
    text-align: center;
    vertical-align: middle;
  }

  .print-template-container .print-info-table td.label-cell {
    font-weight: bold;
    color: #000;
    width: 14%;
  }

  .print-template-container .print-info-table td.value-cell {
    color: #000;
    width: 19%;
  }

  .print-template-container .print-info-table td.empty-cell {
    background: #fff;
    min-height: 40px;
  }

  /* 数据表格 */
  .print-template-container .print-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
    margin-top: 10px;
  }

  .print-template-container .print-table thead tr {
    background-color: #fff;
  }

  .print-template-container .print-table thead th {
    padding: 14px 8px;
    border: 1px solid #333;
    font-weight: bold;
    color: #000;
    text-align: center;
    font-size: 14px;
    background-color: #fff;
  }

  .print-template-container .print-table tbody tr td {
    padding: 10px 8px;
    border: 1px solid #333;
    color: #000;
    text-align: center;
    vertical-align: middle;
    background-color: #fff;
  }

  .print-template-container .print-table tbody tr.summary-row {
    background-color: #fff;
    font-weight: bold;
  }

  /* 打印产品图片容器 */
  .print-template-container .print-image-cell {
    width: 100%;
    height: 150px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0;
    margin: 0;
  }

  /* 打印产品图片 */
  .print-template-container .print-product-image {
    width: 95% !important;
    height: 95% !important;
    max-width: 95% !important;
    max-height: 95% !important;
    object-fit: contain !important;
    display: block !important;
    visibility: visible !important;
    -webkit-print-color-adjust: exact !important;
    print-color-adjust: exact !important;
  }

  /* 无图片占位 */
  .print-template-container .print-image-placeholder {
    width: 100%;
    height: 150px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #999;
    font-size: 14px;
  }

  /* 自定义内容 */
  .print-template-container .custom-content {
    margin-top: 10px;
  }

  /* 签收栏 */
  .print-template-container .signature-section {
    margin-top: 40px;
    padding-top: 40px;
    border-top: 1px dashed #999;
  }

  .print-template-container .signature-row {
    display: flex;
    justify-content: space-between;
    margin-bottom: 40px;
  }

  .print-template-container .signature-item {
    display: flex;
    align-items: center;
    width: 48%;
  }

  .print-template-container .signature-label {
    display: inline-block;
    width: 5em;
    font-weight: bold;
    white-space: nowrap;
    text-align: justify;
    text-align-last: justify;
    margin-right: 10px;
    color: #000;
  }

  .print-template-container .signature-line {
    flex: 1;
    border-bottom: 1px solid #000;
    height: 20px;
  }

  .print-template-container .stamp-label {
    display: inline-block;
    width: 5em;
    font-weight: bold;
    font-size: 14px;
    color: #000;
    text-align: left;
  }

  .print-template-container .date-item {
    width: 48%;
  }

  .print-template-container .date-underline {
    width: 40px;
    border-bottom: 1px solid #000;
    height: 20px;
    margin: 0 2px;
  }

  .print-template-container .date-text {
    font-size: 14px;
    color: #000;
    margin: 0 2px;
  }

  /* 隐藏原始页面 */
  .t-card,
  .page-header,
  .info-card,
  .t-dialog,
  .t-image-viewer {
    display: none !important;
  }
}
</style>

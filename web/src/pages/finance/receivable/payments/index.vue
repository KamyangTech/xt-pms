<template>
  <div class="payment-records-page">
    <!-- 返回按钮 -->
    <div class="back-button-container">
      <t-button theme="default" variant="text" @click="goBack">
        <template #icon><chevron-left-icon /></template>
        返回应收账款列表
      </t-button>
    </div>

    <!-- 应收信息 -->
    <t-card :bordered="false" title="应收信息">
      <t-descriptions :column="4" bordered>
        <t-descriptions-item label="客户名称">
          {{ receivableDetail.customer?.cus_name || '-' }}
        </t-descriptions-item>
        <t-descriptions-item label="订单编号">
          {{ receivableDetail.order_code || '-' }}
        </t-descriptions-item>
        <t-descriptions-item label="款项类型">
          <t-tag
            v-if="receivableDetail.payment_type"
            size="small"
            :theme="getPaymentTypeTheme(receivableDetail.payment_type)"
            :style="getPaymentTypeStyle(receivableDetail.payment_type)"
          >
            {{ getPaymentTypeText(receivableDetail.payment_type) }}
          </t-tag>
          <span v-else>-</span>
        </t-descriptions-item>
        <t-descriptions-item label="状态">
          <t-tag v-if="receivableDetail.status" :theme="getStatusTheme(receivableDetail.status)" :style="getStatusStyle(receivableDetail.status)">
            {{ getStatusText(receivableDetail.status) }}
          </t-tag>
          <span v-else>-</span>
        </t-descriptions-item>
        <t-descriptions-item label="到期日期">
          {{ receivableDetail.due_date ? formatDate(receivableDetail.due_date) : '-' }}
        </t-descriptions-item>
      </t-descriptions>
    </t-card>

    <!-- 资金汇总 -->
    <t-card :bordered="false">
      <div class="summary-stats">
        <div class="stat-item">
          <div class="stat-label">应收总额</div>
          <div class="stat-value">¥{{ formatAmount(receivableDetail.amount) }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">已收金额</div>
          <div class="stat-value success">
            ¥{{ formatAmount(receivableDetail.received_amount) }}
          </div>
        </div>
        <div v-if="(receivableDetail.discount_amount || 0) > 0" class="stat-item">
          <div class="stat-label">已优惠金额</div>
          <div class="stat-value" style="color: var(--td-warning-color)">
            -¥{{ formatAmount(receivableDetail.discount_amount) }}
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-label">余额</div>
          <div class="stat-value warning">¥{{ formatAmount(receivableDetail.balance) }}</div>
        </div>
      </div>
    </t-card>

    <!-- 收款记录表格 -->
    <t-card :bordered="false" title="收款明细">
      <template #actions>
        <t-button theme="primary" @click="handleReceive">
          <template #icon><add-icon /></template>
          收款
        </t-button>
        <t-button theme="default" @click="handleColumnSetting">
          <template #icon><setting-icon /></template>
          列设置
        </t-button>
      </template>
      <t-table
        :data="paymentRecords"
        :columns="displayColumns"
        :loading="loading"
        row-key="id"
        hover
      >
        <template #payment_amount="{ row }">
          <span class="amount-text green">¥{{ formatAmount(row.payment_amount) }}</span>
        </template>
        <template #discount_amount="{ row }">
          <span
            v-if="(row.discount_amount || 0) > 0"
            class="amount-text"
            style="color: var(--td-warning-color)"
            >-¥{{ formatAmount(row.discount_amount) }}</span
          >
          <span v-else>-</span>
        </template>
        <template #payment_date="{ row }">
          {{ formatDate(row.payment_date) }}
        </template>
        <template #payment_method="{ row }">
          <t-tag v-if="row.payment_method" size="small" theme="default" variant="light">
            {{ getPaymentMethodText(row.payment_method) }}
          </t-tag>
          <span v-else>-</span>
        </template>
        <template #receiving_account="{ row }">
          {{ getReceivingAccountText(row.receiving_account) }}
        </template>
        <template #receipt_file="{ row }">
          <div v-if="row.receipt_files?.length" class="file-list">
            <div v-for="(file, idx) in row.receipt_files" :key="idx" class="file-item">
              <t-tag
                size="small"
                :theme="getFileTagTheme(file.file_type)"
                variant="light"
                style="flex-shrink: 0; margin-right: 4px"
              >
                {{ getFileCategory(file.file_name) }}
              </t-tag>
              <t-tooltip
                :content="file.file_name"
                placement="top"
                :show-arrow="false"
                style="flex: 1; min-width: 0"
              >
                <t-link theme="primary" class="file-name-link" @click="handleFilePreview(file)">
                  {{ file.file_name }}
                </t-link>
              </t-tooltip>
            </div>
          </div>
          <span v-else>-</span>
        </template>
        <template #invoice_file="{ row }">
          <div v-if="row.invoice_files?.length" class="file-list">
            <div v-for="(file, idx) in row.invoice_files" :key="idx" class="file-item">
              <t-tag
                size="small"
                :theme="getFileTagTheme(file.file_type)"
                variant="light"
                style="flex-shrink: 0; margin-right: 4px"
              >
                {{ getFileCategory(file.file_name) }}
              </t-tag>
              <t-tooltip
                :content="file.file_name"
                placement="top"
                :show-arrow="false"
                style="flex: 1; min-width: 0"
              >
                <t-link theme="primary" class="file-name-link" @click="handleFilePreview(file)">
                  {{ file.file_name }}
                </t-link>
              </t-tooltip>
            </div>
          </div>
          <span v-else>-</span>
        </template>
        <template #create_time="{ row }">
          {{ formatDateTime(row.create_time) }}
        </template>
        <template #op="{ row }">
          <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
        </template>
      </t-table>
    </t-card>

    <!-- 编辑收款记录对话框 -->
    <t-dialog
      v-model:visible="editDialogVisible"
      header="编辑收款记录"
      width="680px"
      placement="center"
      @confirm="handleEditSubmit"
      @cancel="editDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form :data="editForm" label-width="120px">
        <t-form-item label="收款金额">
          <span class="amount-text">¥{{ formatAmount(editForm.payment_amount) }}</span>
          <span style="margin-left: 8px; color: var(--td-text-color-placeholder); font-size: 12px"
            >（金额不可修改）</span
          >
        </t-form-item>
        <t-form-item v-if="(editForm.discount_amount || 0) > 0" label="优惠金额">
          <span class="amount-text" style="color: var(--td-warning-color)"
            >-¥{{ formatAmount(editForm.discount_amount) }}</span
          >
          <span style="margin-left: 8px; color: var(--td-text-color-placeholder); font-size: 12px"
            >（优惠金额不可修改）</span
          >
        </t-form-item>
        <t-form-item label="收款日期">
          <t-date-picker
            v-model="editForm.payment_date"
            format="YYYY-MM-DD"
            placeholder="选择收款日期"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item label="收款方式">
          <t-select v-model="editForm.payment_method" placeholder="请选择收款方式" clearable>
            <t-option
              v-for="item in paymentMethodOptions"
              :key="item.dict_value"
              :value="item.dict_value"
              :label="item.dict_label"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="流水号">
          <t-input v-model="editForm.transaction_no" placeholder="请输入银行回单号或交易流水号" />
        </t-form-item>
        <t-form-item label="收款账户">
          <t-select
            v-model="editForm.receiving_account"
            placeholder="请选择收款账户"
            clearable
            filterable
          >
            <t-option
              v-for="item in receivingAccountOptions"
              :key="item.id"
              :value="String(item.id)"
              :label="item.account_name + (item.account_no ? `（${item.account_no}）` : '')"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea
            v-model="editForm.remark"
            placeholder="请输入备注信息"
            :autosize="{ minRows: 2, maxRows: 4 }"
          />
        </t-form-item>
        <t-form-item label="回单文件">
          <t-upload
            :key="`edit-receipt-${editDialogVisible}`"
            :auto-upload="false"
            multiple
            :size-limit="{ size: 5, unit: 'MB' }"
            accept="image/*,.pdf,.ofd"
            theme="file"
            :default-files="editReceiptFiles"
            @change="(ctx: any) => handleEditFileChange('receipt', ctx)"
          >
            <template #trigger>
              <t-button variant="outline" theme="default">
                <template #icon><upload-icon /></template>
                选择文件
              </t-button>
            </template>
            <template #fileListDisplay="{ files, onRemove }">
              <ul v-if="files.length > 0" class="upload-file-list">
                <li v-for="file in files" :key="file.name" class="upload-file-item">
                  <t-tag
                    size="small"
                    :theme="getFileTagTheme(`.${getFileExt(file.name)}`)"
                    variant="light"
                    style="flex-shrink: 0"
                  >
                    {{ getFileCategory(file.name) }}
                  </t-tag>
                  <t-tooltip
                    :content="file.name"
                    placement="top"
                    :show-arrow="false"
                    style="flex: 1; min-width: 0"
                  >
                    <t-link
                      theme="primary"
                      hover="color"
                      class="file-name-link"
                      @click="previewEditFile(file)"
                    >
                      {{ file.name }}
                    </t-link>
                  </t-tooltip>
                  <delete-icon class="file-delete-btn" @click.stop="onRemove(file)" />
                </li>
              </ul>
            </template>
          </t-upload>
        </t-form-item>
        <t-form-item label="发票文件">
          <t-upload
            :key="`edit-invoice-${editDialogVisible}`"
            :auto-upload="false"
            multiple
            :size-limit="{ size: 5, unit: 'MB' }"
            accept="image/*,.pdf,.ofd"
            theme="file"
            :default-files="editInvoiceFiles"
            @change="(ctx: any) => handleEditFileChange('invoice', ctx)"
          >
            <template #trigger>
              <t-button variant="outline" theme="default">
                <template #icon><upload-icon /></template>
                选择文件
              </t-button>
            </template>
            <template #fileListDisplay="{ files, onRemove }">
              <ul v-if="files.length > 0" class="upload-file-list">
                <li v-for="file in files" :key="file.name" class="upload-file-item">
                  <t-tag
                    size="small"
                    :theme="getFileTagTheme(`.${getFileExt(file.name)}`)"
                    variant="light"
                    style="flex-shrink: 0"
                  >
                    {{ getFileCategory(file.name) }}
                  </t-tag>
                  <t-tooltip
                    :content="file.name"
                    placement="top"
                    :show-arrow="false"
                    style="flex: 1; min-width: 0"
                  >
                    <t-link
                      theme="primary"
                      hover="color"
                      class="file-name-link"
                      @click="previewEditFile(file)"
                    >
                      {{ file.name }}
                    </t-link>
                  </t-tooltip>
                  <delete-icon class="file-delete-btn" @click.stop="onRemove(file)" />
                </li>
              </ul>
            </template>
          </t-upload>
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 收款对话框 -->
    <t-dialog
      v-model:visible="receiveDialogVisible"
      header="收款"
      width="900px"
      placement="center"
      class="receive-dialog"
      @confirm="handleReceiveSubmit"
      @cancel="receiveDialogVisible = false"
     :close-on-overlay-click="false">
      <div class="receive-dialog-content">
        <!-- 左侧：表单信息 -->
        <div class="receive-dialog-left">
          <div class="stat-card">
            <div class="stat-grid">
              <div class="stat-item">
                <div class="stat-label">应收总额</div>
                <div class="stat-number">¥{{ formatAmount(receivableDetail.amount) }}</div>
              </div>
              <div class="stat-item">
                <div class="stat-label">已收金额</div>
                <div class="stat-number stat-received">
                  ¥{{ formatAmount(receivableDetail.received_amount) }}
                </div>
              </div>
              <div v-if="(receivableDetail.discount_amount || 0) > 0" class="stat-item">
                <div class="stat-label">已优惠金额</div>
                <div class="stat-number stat-discount">
                  -¥{{ formatAmount(receivableDetail.discount_amount) }}
                </div>
              </div>
              <div class="stat-item">
                <div class="stat-label stat-label-emphasis">待收余额</div>
                <div class="stat-number stat-balance">
                  ¥{{ formatAmount(receivableDetail.balance) }}
                </div>
              </div>
            </div>
          </div>

          <t-form :data="receiveForm" label-width="90px" class="receive-form">
            <t-form-item label="收款金额" required>
              <t-input
                v-model="receiveAmountStr"
                type="number"
                placeholder="请输入收款金额"
                style="width: 100%"
                @change="onReceiveAmountChange"
              >
                <template #suffix>元</template>
              </t-input>
            </t-form-item>
            <t-form-item label="优惠减免">
              <t-space direction="vertical" style="width: 100%">
                <t-input
                  v-model="receiveDiscountStr"
                  type="number"
                  placeholder="老板同意减免的金额"
                  style="width: 100%"
                >
                  <template #suffix>元</template>
                </t-input>
                <t-alert
                  v-if="(receiveForm.discount_amount || 0) > 0"
                  theme="warning"
                  :close="false"
                  size="small"
                >
                  <template #message>
                    优惠 ¥{{ formatAmount(receiveForm.discount_amount || 0) }}， 实收 ¥{{
                      formatAmount(receiveForm.amount || 0)
                    }}， 合计 ¥{{ formatAmount((receiveForm.amount || 0) + (receiveForm.discount_amount || 0)) }}
                  </template>
                </t-alert>
              </t-space>
            </t-form-item>
            <t-form-item label="收款日期">
              <t-date-picker
                v-model="receiveForm.payment_date"
                format="YYYY-MM-DD"
                placeholder="选择收款日期"
                style="width: 100%"
              />
            </t-form-item>
            <t-form-item label="收款方式">
              <t-select v-model="receiveForm.payment_method" placeholder="请选择收款方式" clearable>
                <t-option
                  v-for="item in paymentMethodOptions"
                  :key="item.dict_value"
                  :value="item.dict_value"
                  :label="item.dict_label"
                />
              </t-select>
            </t-form-item>
            <t-form-item label="流水号">
              <t-input v-model="receiveForm.transaction_no" placeholder="银行回单号或交易流水号" />
            </t-form-item>
            <t-form-item label="收款账户">
              <t-select
                v-model="receiveForm.receiving_account"
                placeholder="请选择收款账户"
                clearable
                filterable
              >
                <t-option
                  v-for="item in receivingAccountOptions"
                  :key="item.id"
                  :value="String(item.id)"
                  :label="item.account_name + (item.account_no ? `（${item.account_no}）` : '')"
                />
              </t-select>
            </t-form-item>
            <t-form-item label="备注">
              <t-textarea
                v-model="receiveForm.remark"
                placeholder="请输入备注信息"
                :autosize="{ minRows: 2, maxRows: 3 }"
              />
            </t-form-item>
          </t-form>
        </div>

        <!-- 右侧：文件上传 -->
        <div class="receive-dialog-right">
          <div class="file-upload-section">
            <div class="file-upload-header">
              <div class="file-upload-title">
                <file-icon />
                <span>凭证附件</span>
              </div>
              <div class="file-upload-subtitle">支持图片、PDF、OFD格式，单个文件不超过5MB</div>
            </div>
            <div class="file-upload-item">
              <div class="file-upload-label">
                <span class="file-label-text">回单文件</span>
              </div>
              <t-upload
                :key="`receive-receipt-${receiveDialogVisible}`"
                :auto-upload="false"
                multiple
                :size-limit="{ size: 5, unit: 'MB' }"
                accept="image/*,.pdf,.ofd"
                theme="file"
                @change="(ctx: any) => handleReceiveUploadChange('receipt', ctx)"
              >
                <template #fileListDisplay="{ files, onRemove }">
                  <ul v-if="files.length > 0" class="upload-file-list">
                    <li v-for="file in files" :key="file.name" class="upload-file-item">
                      <t-tag size="small" theme="danger" variant="light">PDF</t-tag>
                      <t-link
                        theme="primary"
                        hover="color"
                        class="file-name-link"
                        @click="previewReceiveFile(file.raw)"
                        >{{ file.name }}</t-link
                      >
                      <delete-icon class="file-delete-btn" @click.stop="onRemove(file)" />
                    </li>
                  </ul>
                </template>
              </t-upload>
            </div>
            <div class="file-upload-item">
              <div class="file-upload-label">
                <span class="file-label-text">发票文件</span>
              </div>
              <t-upload
                :key="`receive-invoice-${receiveDialogVisible}`"
                :auto-upload="false"
                multiple
                :size-limit="{ size: 5, unit: 'MB' }"
                accept="image/*,.pdf,.ofd"
                theme="file"
                @change="(ctx: any) => handleReceiveUploadChange('invoice', ctx)"
              >
                <template #fileListDisplay="{ files, onRemove }">
                  <ul v-if="files.length > 0" class="upload-file-list">
                    <li v-for="file in files" :key="file.name" class="upload-file-item">
                      <t-tag size="small" theme="danger" variant="light">PDF</t-tag>
                      <t-link
                        theme="primary"
                        hover="color"
                        class="file-name-link"
                        @click="previewReceiveFile(file.raw)"
                        >{{ file.name }}</t-link
                      >
                      <delete-icon class="file-delete-btn" @click.stop="onRemove(file)" />
                    </li>
                  </ul>
                </template>
              </t-upload>
            </div>
          </div>
        </div>
      </div>
    </t-dialog>

    <!-- 图片预览 -->
    <t-image-viewer
      v-model:visible="editImagePreviewVisible"
      :images="editImagePreviewList"
      :default-index="editImagePreviewIndex"
    />

    <!-- 列设置 -->
    <column-setting-dialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="finance_receivable_payment_columns"
      @change="handleColumnsChange"
    />

    <!-- 表格文件图片预览 -->
    <t-image-viewer
      v-model:visible="fileImagePreviewVisible"
      :images="fileImagePreviewList"
      :default-index="fileImagePreviewIndex"
    />

    <!-- PDF 预览弹窗 -->
    <t-dialog v-model:visible="pdfPreviewVisible" :header="pdfPreviewTitle" placement="center" width="85vw" :footer="false" destroy-on-close :close-on-overlay-click="false">
      <div class="pdf-preview-wrapper">
        <iframe :src="pdfPreviewUrl" class="pdf-preview-iframe" frameborder="0" />
      </div>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import dayjs from 'dayjs';
import {
  AddIcon,
  ChevronLeftIcon,
  DeleteIcon,
  FileIcon,
  SettingIcon,
  UploadIcon,
} from 'tdesign-icons-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { getFileFullUrl, downloadFile, getPdfPreviewUrl, uploadFiles } from '@/api/common/upload';
import {
  getReceivableDetail,
  getReceivablePayments,
  receivePayment,
} from '@/api/finance/receivable';
import { updateReceivablePayment } from '@/api/finance/receivable-payment';
import { getPaymentById } from '@/api/mold/payment';
import { getReceivingAccountList } from '@/api/reference';
import type { ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import ColumnSettingDialog from '@/components/ColumnSettingDialog/index.vue';
import { useDictRaw } from '@/hooks/useDict';
import { getDictColorRenderConfig } from '@/utils/dict-color';

const route = useRoute();
const router = useRouter();

const receivableId = computed(() => Number(route.params.id));
const loading = ref(false);
const receivableDetail = ref<any>({});
const paymentRecords = ref<any[]>([]);
// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: paymentMethodOptions } = useDictRaw('payment-method');
// 收款账户选项（来自 getReceivingAccountList API，不是字典）
const receivingAccountOptions = ref<any[]>([]);
const { items: paymentTypeOptions } = useDictRaw('payment-type');
const { items: receivableStatusOptions } = useDictRaw('receivable-status');

// 收款对话框
const receiveDialogVisible = ref(false);
const receiveForm = reactive({
  amount: null as number | null,
  discount_amount: null as number | null,
  payment_date: '',
  payment_method: '',
  transaction_no: '',
  receiving_account: '',
  remark: '',
});
const receiveReceiptFiles = ref<File[]>([]);
const receiveInvoiceFiles = ref<File[]>([]);

// 字符串-数字桥梁，解决 t-input-number 空值显示为 0 的问题
const receiveAmountStr = computed({
  get: () => (receiveForm.amount != null ? String(receiveForm.amount) : ''),
  set: (val: string) => {
    receiveForm.amount = val !== '' ? Number(val) : null;
  },
});
const receiveDiscountStr = computed({
  get: () => (receiveForm.discount_amount != null ? String(receiveForm.discount_amount) : ''),
  set: (val: string) => {
    receiveForm.discount_amount = val !== '' ? Number(val) : null;
  },
});

// 最大可优惠金额 = 余额 - 收款金额
const maxDiscountAmount = computed(() => {
  const balance = receivableDetail.value?.balance || 0;
  const amount = receiveForm.amount || 0;
  const canDiscount = balance - amount;
  return canDiscount > 0 ? canDiscount : 0;
});

// 当收款金额变化时，自动调整优惠金额不超过上限
const onReceiveAmountChange = () => {
  const discount = receiveForm.discount_amount || 0;
  if (discount > maxDiscountAmount.value) {
    receiveForm.discount_amount = maxDiscountAmount.value;
  }
};

const handleReceive = () => {
  receiveForm.amount = null;
  receiveForm.discount_amount = null;
  receiveForm.payment_date = '';
  receiveForm.payment_method = '';
  receiveForm.transaction_no = '';
  receiveForm.receiving_account = '';
  receiveForm.remark = '';
  receiveReceiptFiles.value = [];
  receiveInvoiceFiles.value = [];
  receiveDialogVisible.value = true;
};

const handleReceiveUploadChange = (type: 'receipt' | 'invoice', ctx: any) => {
  const files = Array.isArray(ctx) ? ctx : ctx?.files || [];
  const rawFiles: File[] = [];
  files.forEach((f: any) => {
    if (f.raw instanceof File) rawFiles.push(f.raw);
  });
  if (type === 'receipt') receiveReceiptFiles.value = rawFiles;
  else receiveInvoiceFiles.value = rawFiles;
};

const previewReceiveFile = (raw: File | undefined) => {
  if (!raw) return;
  if (raw.type.startsWith('image/')) {
    const reader = new FileReader();
    reader.onload = () => {
      const blob = new Blob([reader.result as ArrayBuffer], { type: raw.type });
      const url = URL.createObjectURL(blob);
      window.open(url, '_blank');
      setTimeout(() => URL.revokeObjectURL(url), 60000);
    };
    reader.readAsArrayBuffer(raw);
  } else {
    const reader = new FileReader();
    reader.onload = () => {
      const blob = new Blob([reader.result as ArrayBuffer], { type: raw.type });
      const url = URL.createObjectURL(blob);
      window.open(url, '_blank');
      setTimeout(() => URL.revokeObjectURL(url), 60000);
    };
    reader.onerror = () => MessagePlugin.warning('文件读取失败');
    reader.readAsArrayBuffer(raw);
  }
};

const receiveSubmitting = ref(false);
const handleReceiveSubmit = async () => {
  if (receiveForm.amount <= 0 && receiveForm.discount_amount <= 0) {
    MessagePlugin.warning('请输入收款金额或优惠金额');
    return;
  }
  if (receiveSubmitting.value) return;
  receiveSubmitting.value = true;
  try {
    let receiptFileId = '';
    let invoiceFileId = '';
    if (receiveReceiptFiles.value.length > 0) {
      const results = await uploadFiles({
        files: receiveReceiptFiles.value,
        category: 'receivable-payment',
      });
      receiptFileId = results.map((r: any) => r.id).join(',');
    }
    if (receiveInvoiceFiles.value.length > 0) {
      const results = await uploadFiles({
        files: receiveInvoiceFiles.value,
        category: 'receivable-payment',
      });
      invoiceFileId = results.map((r: any) => r.id).join(',');
    }
    await receivePayment(receivableId.value, {
      amount: receiveForm.amount,
      discount_amount: receiveForm.discount_amount > 0 ? receiveForm.discount_amount : undefined,
      payment_date: receiveForm.payment_date || dayjs().format('YYYY-MM-DD'),
      payment_method: receiveForm.payment_method,
      transaction_no: receiveForm.transaction_no,
      receiving_account: receiveForm.receiving_account,
      remark: receiveForm.remark,
      receipt_file_id: receiptFileId || undefined,
      invoice_file_id: invoiceFileId || undefined,
    });
    MessagePlugin.success('收款成功');
    receiveDialogVisible.value = false;
    loadData();
  } catch (error: any) {
    MessagePlugin.error(error.message || '收款失败');
  } finally {
    receiveSubmitting.value = false;
  }
};

// 列管理
const selectedColumns = ref<string[]>([]);
const columnSettingRef = ref();

const allColumns: ColumnOption[] = [
  { colKey: 'payment_date', title: '收款日期', pinned: true },
  { colKey: 'payment_amount', title: '收款金额', pinned: true },
  { colKey: 'discount_amount', title: '优惠金额' },
  { colKey: 'payment_method', title: '收款方式' },
  { colKey: 'transaction_no', title: '流水号' },
  { colKey: 'receiving_account', title: '收款账户' },
  { colKey: 'receipt_file', title: '回单' },
  { colKey: 'invoice_file', title: '发票' },
  { colKey: 'create_time', title: '录入时间' },
  { colKey: 'remark', title: '备注' },
  { colKey: 'op', title: '操作', pinned: true },
];

const DEFAULT_COLUMNS = [
  'payment_date',
  'payment_amount',
  'discount_amount',
  'payment_method',
  'transaction_no',
  'receiving_account',
  'receipt_file',
  'invoice_file',
  'op',
];

const baseColumns: PrimaryTableCol[] = [
  { colKey: 'payment_date', title: '收款日期', width: 120, fixed: 'left' },
  { colKey: 'payment_amount', title: '收款金额', width: 120, fixed: 'left' },
  { colKey: 'discount_amount', title: '优惠金额', width: 100 },
  { colKey: 'payment_method', title: '收款方式', width: 110 },
  { colKey: 'transaction_no', title: '流水号', width: 200 },
  { colKey: 'receiving_account', title: '收款账户', width: 220 },
  { colKey: 'receipt_file', title: '回单', width: 260, minWidth: 180 },
  { colKey: 'invoice_file', title: '发票', width: 260, minWidth: 180 },
  { colKey: 'create_time', title: '录入时间', width: 160 },
  { colKey: 'remark', title: '备注', ellipsis: true },
  { colKey: 'op', title: '操作', width: 80, fixed: 'right' },
];

const displayColumns = computed(() => {
  const pinnedKeys = ['payment_date', 'payment_amount'];
  const otherKeys = selectedColumns.value.filter((k) => !pinnedKeys.includes(k) && k !== 'op');

  const ordered: PrimaryTableCol[] = [
    ...pinnedKeys.map((k) => baseColumns.find((col) => col.colKey === k)!),
    ...(otherKeys
      .map((k) => baseColumns.find((col) => col.colKey === k))
      .filter(Boolean) as PrimaryTableCol[]),
    baseColumns.find((col) => col.colKey === 'op')!,
  ];

  return ordered;
});

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('finance_receivable_payment_columns');
    if (saved) {
      selectedColumns.value = JSON.parse(saved);
    } else {
      selectedColumns.value = [...DEFAULT_COLUMNS, 'create_time', 'remark'];
    }
  } catch {
    selectedColumns.value = [...DEFAULT_COLUMNS, 'create_time', 'remark'];
  }
  // 确保固定列始终存在
  const pinnedKeys = ['payment_date', 'payment_amount', 'op'];
  pinnedKeys.forEach((key) => {
    if (!selectedColumns.value.includes(key)) {
      selectedColumns.value.unshift(key);
    }
  });
};

const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
  localStorage.setItem('finance_receivable_payment_columns', JSON.stringify(columns));
};

const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 编辑对话框
const editDialogVisible = ref(false);
const editingPayment = ref<any>(null);
const editForm = ref({
  payment_amount: 0,
  discount_amount: 0,
  payment_date: '',
  payment_method: '',
  transaction_no: '',
  receiving_account: '',
  remark: '',
});
const editNewReceiptFiles = ref<File[]>([]);
const editNewInvoiceFiles = ref<File[]>([]);
const editReceiptFiles = ref<any[]>([]);
const editInvoiceFiles = ref<any[]>([]);

const editImageBlobUrls = ref<string[]>([]);
const editImagePreviewVisible = ref(false);
const editImagePreviewList = ref<string[]>([]);
const editImagePreviewIndex = ref(0);

// 表格文件图片预览
const fileImagePreviewVisible = ref(false);
const fileImagePreviewList = ref<string[]>([]);
const fileImagePreviewIndex = ref(0);
// PDF 预览弹窗
const pdfPreviewVisible = ref(false);
const pdfPreviewUrl = ref('');
const pdfPreviewTitle = ref('');
const pdfPreviewIsBlob = ref(false);

const handleEditFileChange = (type: 'receipt' | 'invoice', ctx: any) => {
  const files = Array.isArray(ctx) ? ctx : ctx?.files || [];
  const rawFiles: File[] = [];
  files.forEach((f: any) => {
    if (f.raw instanceof File) rawFiles.push(f.raw);
  });
  if (type === 'receipt') editNewReceiptFiles.value = rawFiles;
  else editNewInvoiceFiles.value = rawFiles;
};

const previewEditFile = async (file: any) => {
  // 已有文件（有 url 属性）
  if (file.url) {
    const imgExts = ['JPG', 'JPEG', 'PNG', 'GIF', 'WEBP', 'BMP', 'SVG', 'PDF'];
    if (file.id && file.fileType && !imgExts.includes(file.fileType?.toUpperCase())) {
      // 非图片/非 PDF 文件弹窗确认后下载
      const confirmDialog = DialogPlugin.confirm({
        header: '确认下载',
        body: `确定要下载文件「${file.name}」吗？`,
        confirmBtn: '下载',
        cancelBtn: '取消',
        onConfirm: async () => {
          confirmDialog.hide();
          try {
            await downloadFile(file.id!, file.name);
          } catch (error: any) {
            MessagePlugin.error(error.message || '文件下载失败');
          }
        },
        onCancel: () => confirmDialog.hide(),
        onClose: () => confirmDialog.destroy(),
      });
      return;
    }
    window.open(file.url, '_blank');
    return;
  }
  // 新选择的文件（有 raw 属性）
  if (file.raw) {
    const reader = new FileReader();
    reader.onload = () => {
      const blob = new Blob([reader.result as ArrayBuffer], { type: file.raw.type });
      const url = URL.createObjectURL(blob);
      editImageBlobUrls.value.push(url);
      if (file.raw.type.startsWith('image/')) {
        editImagePreviewList.value = [url];
        editImagePreviewIndex.value = 0;
        editImagePreviewVisible.value = true;
      } else {
        window.open(url, '_blank');
        setTimeout(() => URL.revokeObjectURL(url), 60000);
      }
    };
    reader.onerror = () => MessagePlugin.warning('文件读取失败');
    reader.readAsArrayBuffer(file.raw);
  }
};

const handleEdit = (row: any) => {
  editingPayment.value = row;
  editForm.value = {
    payment_amount: row.payment_amount,
    discount_amount: row.discount_amount || 0,
    payment_date: row.payment_date ? dayjs(row.payment_date).format('YYYY-MM-DD') : '',
    payment_method: row.payment_method || '',
    transaction_no: row.transaction_no || '',
    receiving_account: row.receiving_account || '',
    remark: row.remark || '',
  };
  editNewReceiptFiles.value = [];
  editNewInvoiceFiles.value = [];
  editImageBlobUrls.value.forEach((url) => URL.revokeObjectURL(url));
  editImageBlobUrls.value = [];
  // 已有的文件显示为默认文件
  editReceiptFiles.value = (row.receipt_files || []).map((f: any) => ({
    name: f.file_name,
    url: getFileFullUrl(f.url),
    status: 'success',
    id: f.id,
    fileType: f.file_type,
  }));
  editInvoiceFiles.value = (row.invoice_files || []).map((f: any) => ({
    name: f.file_name,
    url: getFileFullUrl(f.url),
    status: 'success',
    id: f.id,
    fileType: f.file_type,
  }));
  editDialogVisible.value = true;
};

const handleEditSubmit = async () => {
  if (!editingPayment.value) return;
  try {
    // 上传新文件
    let receiptFileId = editingPayment.value.receipt_file_id || '';
    let invoiceFileId = editingPayment.value.invoice_file_id || '';

    if (editNewReceiptFiles.value.length > 0) {
      const results = await uploadFiles({
        files: editNewReceiptFiles.value,
        category: 'receivable-payment',
      });
      const newIds = results.map((r: any) => r.id).join(',');
      receiptFileId = receiptFileId ? `${receiptFileId},${newIds}` : newIds;
    }
    if (editNewInvoiceFiles.value.length > 0) {
      const results = await uploadFiles({
        files: editNewInvoiceFiles.value,
        category: 'receivable-payment',
      });
      const newIds = results.map((r: any) => r.id).join(',');
      invoiceFileId = invoiceFileId ? `${invoiceFileId},${newIds}` : newIds;
    }

    await updateReceivablePayment(editingPayment.value.id, {
      payment_date: editForm.value.payment_date || undefined,
      payment_method: editForm.value.payment_method || undefined,
      transaction_no: editForm.value.transaction_no || undefined,
      receiving_account: editForm.value.receiving_account || undefined,
      remark: editForm.value.remark || undefined,
      receipt_file_id: receiptFileId || undefined,
      invoice_file_id: invoiceFileId || undefined,
    });
    MessagePlugin.success('更新成功');
    editDialogVisible.value = false;
    loadData();
  } catch (error: any) {
    MessagePlugin.error(error.message || '更新失败');
  }
};

const goBack = () => {
  router.push({ name: 'FinanceReceivable' });
};

const loadData = async () => {
  loading.value = true;
  try {
    const [receivableRes, paymentsRes] = await Promise.all([
      getReceivableDetail(receivableId.value).catch((err) => {
        // 如果应收记录不存在，尝试从 mold_order_payment 获取
        if (err?.response?.status === 404) {
          return null;
        }
        throw err;
      }),
      getReceivablePayments(receivableId.value).catch(() => []),
    ]);

    if (receivableRes === null) {
      // 尝试从 mold_order_payment 获取款项信息
      try {
        const payment = await getPaymentById(receivableId.value);
        if (payment) {
          // 从 mold_order_payment 构造 receivableDetail
          receivableDetail.value = {
            customer: { cus_name: payment.customer?.cus_name || '-' },
            order_code: payment.order_code || `模具订单 #${payment.order_id}`,
            payment_type: payment.payment_type || '',
            status: payment.status === 3 ? 'PAID' : payment.status === 2 ? 'PARTIAL' : 'PENDING',
            amount: payment.contract_amount || 0,
            received_amount: payment.actual_amount || 0,
            discount_amount: 0,
            balance: (payment.contract_amount || 0) - (payment.actual_amount || 0),
            due_date: payment.due_date || payment.payment_date || '',
            remark: '（数据来源于模具订单款项表，非应收账款系统）',
          };
          paymentRecords.value = [];
          return;
        }
      } catch {
        // mold_order_payment also not found
      }
      MessagePlugin.warning('找不到对应的应收记录');
      router.replace({ name: 'FinanceReceivable' });
      return;
    }

    receivableDetail.value = receivableRes || {};
    paymentRecords.value = paymentsRes || [];
  } catch (error: any) {
    MessagePlugin.error(error.message || '加载数据失败');
  } finally {
    loading.value = false;
  }
};

// 加载字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）
const loadReceivingAccountOptions = async () => {
  try {
    const res = await getReceivingAccountList();
    receivingAccountOptions.value = res.list || [];
  } catch (error) {
    console.error('加载收款账户列表失败:', error);
  }
};

// 文件预览
const imageExts = ['JPG', 'JPEG', 'PNG', 'GIF', 'WEBP', 'BMP', 'SVG'];
const handleFilePreview = async (file: { id?: string; url: string; file_name: string; file_type: string }) => {
  if (!file?.url) return;
  const ext = getFileExt(file.file_name);
  if (imageExts.includes(ext)) {
    // 图片使用 TDesign 图片预览器
    fileImagePreviewList.value = [getFileFullUrl(file.url)];
    fileImagePreviewIndex.value = 0;
    fileImagePreviewVisible.value = true;
  } else if (ext === 'PDF') {
    // PDF 弹窗预览（带认证 fetch 成 blob URL）
    if (pdfPreviewIsBlob.value && pdfPreviewUrl.value.startsWith('blob:')) {
      URL.revokeObjectURL(pdfPreviewUrl.value);
    }
    pdfPreviewIsBlob.value = false;
    if (!file.id) {
      pdfPreviewTitle.value = file.file_name;
      pdfPreviewUrl.value = getFileFullUrl(file.url);
      pdfPreviewVisible.value = true;
      return;
    }
    try {
      const blobUrl = await getPdfPreviewUrl(file.id);
      pdfPreviewTitle.value = file.file_name;
      pdfPreviewUrl.value = blobUrl;
      pdfPreviewIsBlob.value = true;
      pdfPreviewVisible.value = true;
    } catch (error: any) {
      MessagePlugin.error(error.message || 'PDF 预览失败');
    }
    return;
  } else if (file.id) {
    // 其他文件（doc/xls/zip 等）弹窗确认后下载
    const confirmDialog = DialogPlugin.confirm({
      header: '确认下载',
      body: `确定要下载文件「${file.file_name}」吗？`,
      confirmBtn: '下载',
      cancelBtn: '取消',
      onConfirm: async () => {
        confirmDialog.hide();
        try {
          await downloadFile(file.id!, file.file_name);
        } catch (error: any) {
          MessagePlugin.error(error.message || '文件下载失败');
        }
      },
      onCancel: () => confirmDialog.hide(),
      onClose: () => confirmDialog.destroy(),
    });
  } else {
    window.open(getFileFullUrl(file.url), '_blank');
  }
};

// 文件类型辅助
const getFileExt = (fileName: string) => {
  const ext = (fileName || '').split('.').pop()?.toUpperCase() || '?';
  return ext;
};

const getFileCategory = (fileName: string) => {
  const ext = (fileName || '').split('.').pop()?.toUpperCase() || '';
  const imageExts = ['JPG', 'JPEG', 'PNG', 'GIF', 'WEBP', 'BMP', 'SVG'];
  const docExts = ['DOC', 'DOCX', 'TXT', 'RTF'];
  const sheetExts = ['XLS', 'XLSX', 'CSV'];
  const archiveExts = ['ZIP', 'RAR', '7Z', 'TAR', 'GZ'];

  if (ext === 'PDF') return 'PDF';
  if (ext === 'OFD') return 'OFD';
  if (imageExts.includes(ext)) return '图片';
  if (sheetExts.includes(ext)) return '表格';
  if (docExts.includes(ext)) return '文档';
  if (archiveExts.includes(ext)) return '压缩包';
  return ext || '文件';
};

const fileTagThemeMap: Record<string, string> = {
  PDF: 'danger',
  OFD: 'danger',
  JPG: 'primary',
  JPEG: 'primary',
  PNG: 'primary',
  GIF: 'primary',
  WEBP: 'primary',
  BMP: 'primary',
  XLS: 'success',
  XLSX: 'success',
  DOC: 'warning',
  DOCX: 'warning',
};

const getFileTagTheme = (fileType?: string) => {
  const ext = (fileType || '' || '').replace('.', '').toUpperCase();
  return (fileTagThemeMap[ext] || 'default') as
    | 'primary'
    | 'success'
    | 'warning'
    | 'danger'
    | 'default';
};

// 格式化
const formatAmount = (amount: number) => {
  if (!amount && amount !== 0) return '-';
  return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

const formatDate = (date: string) => {
  return date ? dayjs(date).format('YYYY-MM-DD') : '-';
};

const formatDateTime = (date: string) => {
  return date ? dayjs(date).format('YYYY-MM-DD HH:mm') : '-';
};

// 字典映射
const getPaymentMethodText = (value: string) => {
  if (!value) return '-';
  const option = paymentMethodOptions.value.find((item: any) => item.dict_value === value);
  return option ? option.dict_label : value;
};

const getReceivingAccountText = (value: string) => {
  if (!value) return '-';
  const numericId = Number(value);
  const account = receivingAccountOptions.value.find((item: any) => item.id === numericId);
  if (account) {
    return account.account_name + (account.account_no ? `（${account.account_no}）` : '');
  }
  return value;
};

const getPaymentTypeText = (type: string) => {
  const option = paymentTypeOptions.value.find((item: any) => item.dict_value === type);
  return option ? option.dict_label : type;
};

const getPaymentTypeStyle = (type: string) => {
  const option = paymentTypeOptions.value.find((item: any) => item.dict_value === type);
  if (option && option.color) {
    const renderConfig = getDictColorRenderConfig(option.color);
    if (renderConfig.mode === 'style' && renderConfig.style) {
      return renderConfig.style;
    }
    return {};
  }
  return {};
};
const getPaymentTypeTheme = (type: string) => {
  const option = paymentTypeOptions.value.find((item: any) => item.dict_value === type);
  if (option && option.color) {
    const renderConfig = getDictColorRenderConfig(option.color);
    if (renderConfig.mode === 'theme') {
      return renderConfig.theme || 'default';
    }
  }
  return 'default';
};

const getStatusText = (status: string) => {
  if (status === 'locked') return '已锁定';
  const option = receivableStatusOptions.value.find((item: any) => item.dict_value === status);
  if (option && option.dict_label) return option.dict_label;
  const textMap: Record<string, string> = {
    pending: '待收款',
    partial: '部分收款',
    paid: '已收款',
    overdue: '逾期',
  };
  return textMap[status?.toLowerCase()] || status;
};

const getStatusStyle = (status: string) => {
  if (status === 'locked') return {};
  const option = receivableStatusOptions.value.find((item: any) => item.dict_value === status) as any;
  if (option?.color) {
    const config = getDictColorRenderConfig(option.color);
    return config.mode === 'style' ? (config.style || {}) : {};
  }
  return {};
};

const getStatusTheme = (status: string) => {
  if (status === 'locked') return undefined;
  const option = receivableStatusOptions.value.find((item: any) => item.dict_value === status) as any;
  if (option?.color) {
    return getDictColorRenderConfig(option.color).theme;
  }
  const themeMap: Record<string, string> = { pending: 'default', partial: 'warning', paid: 'success', overdue: 'danger' };
  return themeMap[status?.toLowerCase()] as any;
};

onMounted(() => {
  loadColumnConfig();
  loadData();
  loadReceivingAccountOptions();
});

// 监听路由参数变化，切换应收账款时重新加载数据
watch(
  () => route.params.id,
  (newId) => {
    if (newId && route.name === 'FinanceReceivablePayments') {
      loadData();
      loadReceivingAccountOptions();
    }
  },
);

// 关闭 PDF 弹窗时释放 blob URL
watch(pdfPreviewVisible, (visible) => {
  if (!visible && pdfPreviewIsBlob.value && pdfPreviewUrl.value.startsWith('blob:')) {
    URL.revokeObjectURL(pdfPreviewUrl.value);
    pdfPreviewUrl.value = '';
    pdfPreviewIsBlob.value = false;
  }
});
</script>
<style scoped>
.payment-records-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
}

.back-button-container {
  display: flex;
  align-items: center;
}

.summary-stats {
  display: flex;
  gap: 24px;
  justify-content: space-around;
  padding: 16px 0;
}

.stat-item {
  text-align: center;
  flex: 1;
}

.stat-label {
  font-size: 14px;
  color: var(--td-text-color-secondary);
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.stat-value.success {
  color: var(--td-success-color);
}

.stat-value.warning {
  color: var(--td-warning-color);
}

.amount-text {
  font-weight: 500;
}

.amount-text.green {
  color: var(--td-success-color);
}

.file-list {
  display: flex;
  flex-direction: column;
}

.file-item {
  display: flex;
  align-items: center;
  gap: 4px;
  line-height: 1.5;
  padding: 4px 0;
  border-bottom: 1px solid var(--td-border-level-1-color);
}

.file-item:last-child {
  border-bottom: none;
}

.file-name-link {
  display: inline-block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  vertical-align: bottom;
  max-width: 100%;
}

/* 上传文件列表（与收款录入一致） */
.upload-file-list {
  list-style: none;
  padding: 0;
  margin: 8px 0 0;
  width: 100%;
}

.upload-file-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 8px;
  border: 1px solid var(--td-border-level-1-color);
  border-radius: var(--td-radius-medium);
  background: var(--td-bg-color-secondarycontainer);
}

.upload-file-item:last-child {
  margin-bottom: 0;
}

/* 收款对话框 - 响应式布局 */
.receive-dialog-content {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 20px;
  min-height: 380px;
}

.receive-dialog-left {
  min-width: 0;
  padding-right: 20px;
  border-right: 1px solid var(--td-border-level-1-color);
}

.receive-dialog-right {
  min-width: 280px;
  display: flex;
  flex-direction: column;
}

/* 窄屏时弹窗本身缩小适配 */
@media (width <= 960px) {
  .receive-dialog :deep(.t-dialog__wrap) {
    min-width: unset;
  }
}

/* 响应式：窄屏幕下单栏布局 */
@media (width <= 768px) {
  .receive-dialog-content {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .receive-dialog-left {
    padding-right: 0;
    border-right: none;
    border-bottom: 1px solid var(--td-border-level-1-color);
    padding-bottom: 16px;
  }

  .receive-dialog-right {
    padding-left: 0;
  }
}

/* 左侧 - 统计卡片 */
.stat-card {
  background: linear-gradient(
    135deg,
    var(--td-brand-color-light, #e8f4ff) 0%,
    var(--td-bg-color-secondarycontainer) 60%
  );
  border-radius: var(--td-radius-medium);
  padding: 14px 16px;
  margin-bottom: 16px;
  border: 1px solid var(--td-border-level-1-color);
}

.stat-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.stat-grid .stat-item {
  text-align: center;
  padding: 6px 0;
}

.stat-grid .stat-item:last-child {
  background: var(--td-warning-color-light, #fff3e0);
  border-radius: var(--td-radius-small);
  padding: 10px 6px;
  margin: -4px 0;
}

.stat-grid .stat-label {
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  margin-bottom: 4px;
  line-height: 1.4;
}

.stat-label-emphasis {
  font-weight: 500;
  color: var(--td-warning-color);
}

.stat-number {
  font-size: 16px;
  font-weight: 700;
  color: var(--td-text-color-primary);
  font-variant-numeric: tabular-nums;
}

.stat-received {
  color: var(--td-success-color);
}

.stat-discount {
  color: var(--td-warning-color);
}

.stat-balance {
  font-size: 20px;
  color: var(--td-warning-color);
}

/* 左侧 - 表单 */
.receive-form {
  margin-top: 0;
}

.receive-form :deep(.t-form__item) {
  margin-bottom: 14px;
}

/* 右侧 - 文件上传区 */
.file-upload-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
  height: 100%;
}

.file-upload-header {
  padding-bottom: 8px;
  margin-bottom: 12px;
  border-bottom: 1px solid var(--td-border-level-1-color);
}

.file-upload-title {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.file-upload-subtitle {
  font-size: 11px;
  color: var(--td-text-color-placeholder);
  margin-top: 4px;
}

.file-upload-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.file-upload-label {
  display: flex;
  align-items: center;
}

.file-label-text {
  font-size: 12px;
  font-weight: 500;
  color: var(--td-text-color-secondary);
}

.file-preview-item .file-name-link {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-preview-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.file-preview-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  border: 1px solid var(--td-border-level-1-color);
  border-radius: var(--td-radius-small);
  background: var(--td-bg-color-secondarycontainer);
  transition: background 0.2s;
}

.file-preview-item:hover {
  background: var(--td-bg-color-container-hover);
}

.file-preview-info {
  flex: 1;
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
  color: var(--td-brand-color);
  font-size: 12px;
}

.file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.upload-trigger-btn {
  width: 100%;
}

.file-delete-btn {
  margin-left: auto;
  cursor: pointer;
  color: var(--td-text-color-placeholder);
  flex-shrink: 0;
  transition: color 0.2s;
}

.file-delete-btn:hover {
  color: var(--td-error-color);
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

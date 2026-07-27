<template>
  <div class="payment-records-page">
    <!-- 返回按钮 -->
    <div class="back-button-container">
      <t-button theme="default" @click="goBack" variant="text">
        <template #icon><chevron-left-icon /></template>
        返回应付账款列表
      </t-button>
    </div>

    <!-- 应付信息 -->
    <t-card :bordered="false" title="应付信息">
      <t-descriptions :column="4" bordered>
        <t-descriptions-item label="供应商/加工商">
          {{ payableDetail.supplier?.sup_name || payableDetail.processor?.processor_name || '-' }}
        </t-descriptions-item>
        <t-descriptions-item label="订单编号">
          {{ payableDetail.order_code || '-' }}
        </t-descriptions-item>
        <t-descriptions-item label="款项类型">
          <span v-if="!payableDetail.source_type && !payableDetail.payment_type">-</span>
          <t-space v-else>
            <t-tag v-if="payableDetail.source_type" size="small" :style="getSourceTypeStyle(payableDetail.source_type)">
              {{ getSourceTypeText(payableDetail.source_type) }}
            </t-tag>
            <t-tag v-if="payableDetail.payment_type" size="small" :style="getPaymentTypeStyle(payableDetail.payment_type)">
              {{ getPaymentTypeText(payableDetail.payment_type) }}
            </t-tag>
          </t-space>
        </t-descriptions-item>
        <t-descriptions-item label="状态">
          <t-tag v-if="payableDetail.status" :theme="getStatusTheme(payableDetail.status)" :style="getStatusStyle(payableDetail.status)">
            {{ getStatusText(payableDetail.status) }}
          </t-tag>
          <span v-else>-</span>
        </t-descriptions-item>
        <t-descriptions-item label="到期日期">
          {{ payableDetail.due_date ? formatDate(payableDetail.due_date) : '-' }}
        </t-descriptions-item>
      </t-descriptions>
    </t-card>

    <!-- 资金汇总 -->
    <t-card :bordered="false">
      <div class="summary-stats">
        <div class="stat-item">
          <div class="stat-label">应付总额</div>
          <div class="stat-value">¥{{ formatAmount(payableDetail.amount) }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">已付金额</div>
          <div class="stat-value success">¥{{ formatAmount(payableDetail.paid_amount) }}</div>
        </div>
        <div class="stat-item" v-if="(payableDetail.discount_amount || 0) > 0">
          <div class="stat-label">已优惠金额</div>
          <div class="stat-value" style="color: var(--td-warning-color);">-¥{{ formatAmount(payableDetail.discount_amount) }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">余额</div>
          <div class="stat-value warning">¥{{ formatAmount(payableDetail.balance) }}</div>
        </div>
      </div>
    </t-card>

    <!-- 付款记录表格 -->
    <t-card :bordered="false" title="付款明细">
      <template #actions>
        <t-button theme="primary" @click="handlePay">
          <template #icon><add-icon /></template>
          付款
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
          <span v-if="(row.discount_amount || 0) > 0" class="amount-text" style="color: var(--td-warning-color);">-¥{{ formatAmount(row.discount_amount) }}</span>
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
        <template #payment_account="{ row }">
          {{ getPaymentAccountText(row.payment_account) }}
        </template>
        <template #receipt_file="{ row }">
          <div v-if="row.receipt_files?.length" class="file-list">
            <div v-for="(file, idx) in row.receipt_files" :key="idx" class="file-item">
              <t-tag size="small" :theme="getFileTagTheme(file.file_type)" variant="light" style="flex-shrink: 0; margin-right: 4px;">
                {{ getFileCategory(file.file_name) }}
              </t-tag>
              <t-tooltip :content="file.file_name" placement="top" :show-arrow="false" style="flex: 1; min-width: 0;">
                <t-link theme="primary" @click="handleFilePreview(file)" class="file-name-link">
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
              <t-tag size="small" :theme="getFileTagTheme(file.file_type)" variant="light" style="flex-shrink: 0; margin-right: 4px;">
                {{ getFileCategory(file.file_name) }}
              </t-tag>
              <t-tooltip :content="file.file_name" placement="top" :show-arrow="false" style="flex: 1; min-width: 0;">
                <t-link theme="primary" @click="handleFilePreview(file)" class="file-name-link">
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

    <!-- 付款对话框 -->
    <t-dialog
      v-model:visible="payDialogVisible"
      header="付款"
      width="900px"
      placement="center"
      class="pay-dialog"
      @confirm="handlePaySubmit"
      @cancel="payDialogVisible = false"
     :close-on-overlay-click="false">
      <div class="pay-dialog-content">
        <!-- 左侧：表单信息 -->
        <div class="pay-dialog-left">
          <div class="stat-card">
            <div class="stat-grid">
              <div class="stat-item">
                <div class="stat-label">应付总额</div>
                <div class="stat-number">¥{{ formatAmount(payableDetail.amount) }}</div>
              </div>
              <div class="stat-item">
                <div class="stat-label">已付金额</div>
                <div class="stat-number stat-paid">¥{{ formatAmount(payableDetail.paid_amount) }}</div>
              </div>
              <div class="stat-item">
                <div class="stat-label stat-label-emphasis">待付余额</div>
                <div class="stat-number stat-balance">¥{{ formatAmount(payableDetail.balance) }}</div>
              </div>
            </div>
          </div>

          <t-form :data="payForm" label-width="90px" class="pay-form">
            <t-form-item label="付款金额" required>
              <t-input-number
                v-model="payForm.amount"
                :min="0.01"
                :max="payableDetail.balance || 0"
                :precision="2"
                placeholder="请输入付款金额"
                theme="normal"
                style="width: 100%"
                @change="onPayAmountChange"
              >
                <template #suffix>元</template>
              </t-input-number>
            </t-form-item>
            <t-form-item label="优惠减免">
              <t-space direction="vertical" style="width: 100%">
                <t-input-number
                  v-model="payForm.discount_amount"
                  :min="0"
                  :max="maxDiscountAmount"
                  :precision="2"
                  placeholder="供应商同意减免的金额"
                  theme="normal"
                  style="width: 100%"
                >
                  <template #suffix>元</template>
                </t-input-number>
                <t-alert v-if="payForm.discount_amount > 0" theme="warning" :close="false" size="small">
                  <template #message>
                    优惠 ¥{{ formatAmount(payForm.discount_amount) }}，
                    实付 ¥{{ formatAmount(payForm.amount) }}，
                    合计 ¥{{ formatAmount(payForm.amount + payForm.discount_amount) }}
                  </template>
                </t-alert>
              </t-space>
            </t-form-item>
            <t-form-item label="付款日期">
              <t-date-picker
                v-model="payForm.payment_date"
                format="YYYY-MM-DD"
                placeholder="选择付款日期"
                style="width: 100%"
              />
            </t-form-item>
            <t-form-item label="付款方式">
              <t-select v-model="payForm.payment_method" placeholder="请选择付款方式" clearable>
                <t-option
                  v-for="item in paymentMethodOptions"
                  :key="item.dict_value"
                  :value="item.dict_value"
                  :label="item.dict_label"
                />
              </t-select>
            </t-form-item>
            <t-form-item label="流水号">
              <t-input v-model="payForm.transaction_no" placeholder="银行回单号或交易流水号" />
            </t-form-item>
            <t-form-item label="付款账户">
              <t-select v-model="payForm.payment_account" placeholder="请选择付款账户" clearable filterable :disabled="filteredAccountOptions.length === 0">
                <t-option
                  v-for="item in filteredAccountOptions"
                  :key="item.dict_value"
                  :value="item.dict_value"
                  :label="item.dict_label"
                />
              </t-select>
            </t-form-item>
            <t-form-item label="收款方账户">
              <t-input v-model="payForm.receiver_account" placeholder="请输入收款方账户" />
            </t-form-item>
            <t-form-item label="备注">
              <t-textarea v-model="payForm.remark" placeholder="请输入备注信息" :autosize="{ minRows: 2, maxRows: 3 }" />
            </t-form-item>
          </t-form>
        </div>

        <!-- 右侧：文件上传 -->
        <div class="pay-dialog-right">
          <div class="file-upload-section">
            <div class="file-upload-header">
              <div class="file-upload-title">
                <file-icon />
                <span>凭证附件</span>
              </div>
              <div class="file-upload-subtitle">支持图片、PDF格式，单个文件不超过5MB</div>
            </div>
            <div class="file-upload-item">
              <div class="file-upload-label">
                <span class="file-label-text">回单文件</span>
              </div>
              <t-upload
                :key="`pay-receipt-${payDialogVisible}`"
                :auto-upload="false"
                multiple
                :size-limit="{ size: 5, unit: 'MB' }"
                accept="image/*,.pdf"
                theme="file"
                @change="(ctx: any) => handlePayUploadChange('receipt', ctx)"
              >
                <template #fileListDisplay="{ files, onRemove }">
                  <ul v-if="files.length > 0" class="upload-file-list">
                    <li v-for="file in files" :key="file.name" class="upload-file-item">
                      <t-tag size="small" theme="danger" variant="light">PDF</t-tag>
                      <t-link theme="primary" hover="color" @click="previewPayFile(file.raw)" class="file-name-link">{{ file.name }}</t-link>
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
                :key="`pay-invoice-${payDialogVisible}`"
                :auto-upload="false"
                multiple
                :size-limit="{ size: 5, unit: 'MB' }"
                accept="image/*,.pdf"
                theme="file"
                @change="(ctx: any) => handlePayUploadChange('invoice', ctx)"
              >
                <template #fileListDisplay="{ files, onRemove }">
                  <ul v-if="files.length > 0" class="upload-file-list">
                    <li v-for="file in files" :key="file.name" class="upload-file-item">
                      <t-tag size="small" theme="danger" variant="light">PDF</t-tag>
                      <t-link theme="primary" hover="color" @click="previewPayFile(file.raw)" class="file-name-link">{{ file.name }}</t-link>
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

    <!-- 列设置 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="finance_payable_payment_columns"
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
import { ref, reactive, computed, onMounted, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ChevronLeftIcon, AddIcon, SettingIcon, FileIcon, DeleteIcon } from 'tdesign-icons-vue-next';
import { getPayableDetail } from '@/api/finance/payable';
import { getPayablePayments } from '@/api/finance/payable-payment';
import { payPayable } from '@/api/finance/payable';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { getReceivingAccountList } from '@/api/reference';
import { useDictRaw } from '@/hooks/useDict';
import { getFileFullUrl, downloadFile, getPdfPreviewUrl, uploadFiles } from '@/api/common/upload';
import { getDictColorRenderConfig } from '@/utils/dict-color';
import dayjs from 'dayjs';
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';

const route = useRoute();
const router = useRouter();

const payableId = computed(() => Number(route.params.id));
const loading = ref(false);
const payableDetail = ref<any>({});
const paymentRecords = ref<any[]>([]);
// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: paymentMethodOptions } = useDictRaw('payment-method');
// 收款账户选项（来自 getReceivingAccountList API，不是字典）
const receivingAccountOptions = ref<any[]>([]);
const { items: payableStatusOptions } = useDictRaw('payable-status', (items) =>
  (items as any[]).filter((i: any) => i.status === 1),
);

// 付款对话框
const payDialogVisible = ref(false);
const payForm = reactive({
  amount: 0,
  discount_amount: 0,
  payment_date: '',
  payment_method: '',
  transaction_no: '',
  payment_account: '',
  receiver_account: '',
  remark: '',
});
const payReceiptFiles = ref<File[]>([]);
const payInvoiceFiles = ref<File[]>([]);

// 最大可优惠金额 = 余额 - 付款金额
const maxDiscountAmount = computed(() => {
  const balance = payableDetail.value?.balance || 0;
  const canDiscount = balance - payForm.amount;
  return canDiscount > 0 ? canDiscount : 0;
});

// 当付款金额变化时，自动调整优惠金额不超过上限
const onPayAmountChange = () => {
  if (payForm.discount_amount > maxDiscountAmount.value) {
    payForm.discount_amount = maxDiscountAmount.value;
  }
};

// 根据付款方式获取对应的账户类型列表（从字典 ext 配置读取）
const getAccountTypesByMethod = (method: string): string[] => {
  if (!method) return [];
  const item = paymentMethodOptions.value.find(
    (item: any) => item.dict_value === method
  );
  if (!item || !item.ext) return [];
  try {
    const ext = typeof item.ext === 'string' ? JSON.parse(item.ext) : item.ext;
    return ext.account_types || [];
  } catch {
    return [];
  }
};

// 根据选择的付款方式过滤付款账户
const filteredAccountOptions = computed(() => {
  const method = payForm.payment_method;
  if (!method) return receivingAccountOptions.value;
  const allowedTypes = getAccountTypesByMethod(method);
  if (allowedTypes.length === 0) return [];
  return receivingAccountOptions.value.filter(
    (item: any) => allowedTypes.includes(item.dict_value)
  );
});

// 监听付款方式变化，联动清空不匹配的付款账户
watch(() => payForm.payment_method, (newMethod) => {
  if (!payForm.payment_account) return;
  if (!newMethod) return;
  const allowedTypes = getAccountTypesByMethod(newMethod);
  if (allowedTypes.length === 0) {
    payForm.payment_account = '';
  } else {
    const current = receivingAccountOptions.value.find(
      (item: any) => item.dict_value === payForm.payment_account
    );
    if (current && !allowedTypes.includes(current.dict_value)) {
      payForm.payment_account = '';
    }
  }
});

// 列管理
const selectedColumns = ref<string[]>([]);
const columnSettingRef = ref();

const allColumns: ColumnOption[] = [
  { colKey: 'payment_date', title: '付款日期', pinned: true },
  { colKey: 'payment_amount', title: '付款金额', pinned: true },
  { colKey: 'discount_amount', title: '优惠金额' },
  { colKey: 'payment_method', title: '付款方式' },
  { colKey: 'transaction_no', title: '流水号' },
  { colKey: 'payment_account', title: '付款账户' },
  { colKey: 'receiver_account', title: '收款方账户' },
  { colKey: 'receipt_file', title: '回单' },
  { colKey: 'invoice_file', title: '发票' },
  { colKey: 'remark', title: '备注' },
  { colKey: 'create_time', title: '录入时间' },
  { colKey: 'op', title: '操作', pinned: true },
];

const DEFAULT_COLUMNS = ['payment_date', 'payment_amount', 'discount_amount', 'payment_method', 'transaction_no', 'payment_account', 'receiver_account', 'receipt_file', 'invoice_file', 'op'];

const baseColumns: PrimaryTableCol[] = [
  { colKey: 'payment_date', title: '付款日期', width: 120, fixed: 'left' },
  { colKey: 'payment_amount', title: '付款金额', width: 120, fixed: 'left' },
  { colKey: 'discount_amount', title: '优惠金额', width: 100 },
  { colKey: 'payment_method', title: '付款方式', width: 110 },
  { colKey: 'transaction_no', title: '流水号', width: 200 },
  { colKey: 'payment_account', title: '付款账户', width: 220 },
  { colKey: 'receiver_account', title: '收款方账户', width: 220 },
  { colKey: 'receipt_file', title: '回单', width: 260, minWidth: 180 },
  { colKey: 'invoice_file', title: '发票', width: 260, minWidth: 180 },
  { colKey: 'remark', title: '备注', ellipsis: true },
  { colKey: 'create_time', title: '录入时间', width: 160 },
  { colKey: 'op', title: '操作', width: 80, fixed: 'right' },
];

const displayColumns = computed(() => {
  const pinnedKeys = ['payment_date', 'payment_amount'];
  const otherKeys = selectedColumns.value.filter(k => !pinnedKeys.includes(k) && k !== 'op');

  const ordered: PrimaryTableCol[] = [
    ...pinnedKeys.map(k => baseColumns.find(col => col.colKey === k)!),
    ...otherKeys.map(k => baseColumns.find(col => col.colKey === k)).filter(Boolean) as PrimaryTableCol[],
    baseColumns.find(col => col.colKey === 'op')!,
  ];

  return ordered;
});

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('finance_payable_payment_columns');
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
  pinnedKeys.forEach(key => {
    if (!selectedColumns.value.includes(key)) {
      selectedColumns.value.unshift(key);
    }
  });
};

const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
  localStorage.setItem('finance_payable_payment_columns', JSON.stringify(columns));
};

const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 付款
const handlePay = () => {
  payForm.amount = 0;
  payForm.discount_amount = 0;
  payForm.payment_date = '';
  payForm.payment_method = '';
  payForm.transaction_no = '';
  payForm.payment_account = '';
  payForm.receiver_account = '';
  payForm.remark = '';
  payReceiptFiles.value = [];
  payInvoiceFiles.value = [];
  payDialogVisible.value = true;
};

// 付款上传文件变化
const handlePayUploadChange = (type: 'receipt' | 'invoice', ctx: any) => {
  const files = Array.isArray(ctx) ? ctx : (ctx?.files || []);
  const rawFiles: File[] = [];
  files.forEach((f: any) => {
    if (f.raw instanceof File) rawFiles.push(f.raw);
  });
  if (type === 'receipt') payReceiptFiles.value = rawFiles;
  else payInvoiceFiles.value = rawFiles;
};

// 预览付款上传的文件
const previewPayFile = (raw: File | undefined) => {
  if (!raw) return;
  const reader = new FileReader();
  reader.onload = () => {
    const blob = new Blob([reader.result as ArrayBuffer], { type: raw.type });
    const url = URL.createObjectURL(blob);
    window.open(url, '_blank');
    setTimeout(() => URL.revokeObjectURL(url), 60000);
  };
  reader.onerror = () => MessagePlugin.warning('文件读取失败');
  reader.readAsArrayBuffer(raw);
};

const paySubmitting = ref(false);

// 提交付款
const handlePaySubmit = async () => {
  if (!payableDetail.value || (payForm.amount <= 0 && payForm.discount_amount <= 0)) {
    MessagePlugin.warning('请输入有效的付款金额或优惠金额');
    return;
  }
  if (paySubmitting.value) return;
  paySubmitting.value = true;

  try {
    let receiptFileId = '';
    let invoiceFileId = '';
    if (payReceiptFiles.value.length > 0) {
      const results = await uploadFiles({
        files: payReceiptFiles.value,
        category: 'payable-payment',
      });
      receiptFileId = results.map((r: any) => r.id).join(',');
    }
    if (payInvoiceFiles.value.length > 0) {
      const results = await uploadFiles({
        files: payInvoiceFiles.value,
        category: 'payable-payment',
      });
      invoiceFileId = results.map((r: any) => r.id).join(',');
    }
    await payPayable(payableId.value, {
      amount: payForm.amount,
      discount_amount: payForm.discount_amount > 0 ? payForm.discount_amount : undefined,
      payment_date: payForm.payment_date || dayjs().format('YYYY-MM-DD'),
      payment_method: payForm.payment_method,
      transaction_no: payForm.transaction_no,
      payment_account: payForm.payment_account,
      receiver_account: payForm.receiver_account,
      remark: payForm.remark,
      receipt_file_id: receiptFileId || undefined,
      invoice_file_id: invoiceFileId || undefined,
    });
    MessagePlugin.success('付款成功');
    payDialogVisible.value = false;
    payForm.amount = 0;
    payForm.discount_amount = 0;
    payForm.payment_date = '';
    payForm.payment_method = '';
    payForm.transaction_no = '';
    payForm.payment_account = '';
    payForm.receiver_account = '';
    payForm.remark = '';
    payReceiptFiles.value = [];
    payInvoiceFiles.value = [];
    loadData();
  } catch (error: any) {
    MessagePlugin.error(error.message || '付款失败');
  } finally {
    paySubmitting.value = false;
  }
};

// 编辑付款记录
const handleEdit = (row: any) => {
  MessagePlugin.info('编辑功能暂未实现');
};

const goBack = () => {
  router.push({ name: 'FinancePayable' });
};

const loadData = async () => {
  loading.value = true;
  try {
    const [detailRes, paymentsRes] = await Promise.all([
      getPayableDetail(payableId.value),
      getPayablePayments(payableId.value),
    ]);
    payableDetail.value = detailRes || {};
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
const fileImagePreviewVisible = ref(false);
const fileImagePreviewList = ref<string[]>([]);
const fileImagePreviewIndex = ref(0);
// PDF 预览弹窗
const pdfPreviewVisible = ref(false);
const pdfPreviewUrl = ref('');
const pdfPreviewTitle = ref('');
const pdfPreviewIsBlob = ref(false);
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
  if (imageExts.includes(ext)) return '图片';
  if (sheetExts.includes(ext)) return '表格';
  if (docExts.includes(ext)) return '文档';
  if (archiveExts.includes(ext)) return '压缩包';
  return ext || '文件';
};

const fileTagThemeMap: Record<string, string> = {
  PDF: 'danger',
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
  const ext = ((fileType || '') || '').replace('.', '').toUpperCase();
  return (fileTagThemeMap[ext] || 'default') as 'primary' | 'success' | 'warning' | 'danger' | 'default';
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

const getPaymentAccountText = (value: string) => {
  if (!value) return '-';
  const numericId = Number(value);
  const account = receivingAccountOptions.value.find((item: any) => item.id === numericId);
  if (account) {
    return account.account_name + (account.account_no ? '（' + account.account_no + '）' : '');
  }
  // 也尝试从字典中匹配
  const dictItem = receivingAccountOptions.value.find((item: any) => item.dict_value === value);
  return dictItem ? dictItem.dict_label : value;
};

const getSourceTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    mold: '模具',
    production: '生产',
    purchase: '采购',
    outsource: '外发',
  };
  return textMap[type] || type;
};

const getSourceTypeStyle = (type: string) => {
  const styleMap: Record<string, any> = {
    mold: { background: '#e8f4ff', color: '#0052d9', border: '1px solid #b8dcff' },
    production: { background: '#e8f8e8', color: '#00a870', border: '1px solid #b8e8b8' },
    purchase: getDictColorRenderConfig('default').style || {},
    outsource: getDictColorRenderConfig('warning').style || {},
  };
  return styleMap[type] || {};
};

const getPaymentTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    total: '全款',
    first: '首期款',
    final: '尾款',
    deposit: '首付款',
    'FIRST_PAYMENT': '首付款',
    'FINAL_PAYMENT': '尾款',
    installment: '分期款',
    production: '生产款',
  };
  return textMap[type] || type;
};

const getPaymentTypeStyle = (type: string) => {
  const styleMap: Record<string, any> = {
    total: getDictColorRenderConfig('primary').style || {},
    first: getDictColorRenderConfig('success').style || {},
    'FIRST_PAYMENT': getDictColorRenderConfig('success').style || {},
    deposit: getDictColorRenderConfig('success').style || {},
    final: getDictColorRenderConfig('warning').style || {},
    'FINAL_PAYMENT': getDictColorRenderConfig('warning').style || {},
    installment: getDictColorRenderConfig('default').style || {},
    production: getDictColorRenderConfig('brand').style || {},
  };
  return styleMap[type] || {};
};

const getStatusText = (status: string) => {
  const option = payableStatusOptions.value.find((item: any) => item.dict_value === status);
  if (option && option.dict_label) return option.dict_label;
  const textMap: Record<string, string> = { pending: '待付款', partial: '部分付款', paid: '已付款', overdue: '逾期' };
  return textMap[status?.toLowerCase()] || status;
};

const getStatusStyle = (status: string) => {
  const option = payableStatusOptions.value.find((item: any) => item.dict_value === status);
  if (option?.color) {
    const config = getDictColorRenderConfig(option.color);
    return config.mode === 'style' ? (config.style || {}) : {};
  }
  return {};
};

const getStatusTheme = (status: string) => {
  const option = payableStatusOptions.value.find((item: any) => item.dict_value === status);
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

// 监听路由参数变化，切换应付账款时重新加载数据
watch(
  () => route.params.id,
  (newId) => {
    if (newId && route.name === 'FinancePayablePayments') {
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

.summary-stats .stat-item {
  text-align: center;
  flex: 1;
}

.summary-stats .stat-label {
  font-size: 14px;
  color: var(--td-text-color-secondary);
  margin-bottom: 8px;
}

.summary-stats .stat-value {
  font-size: 24px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.summary-stats .stat-value.success {
  color: var(--td-success-color);
}

.summary-stats .stat-value.warning {
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

/* 付款对话框 */
.pay-dialog :deep(.t-dialog__wrap) {
  min-width: unset;
}

.pay-dialog-content {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 20px;
  min-height: 380px;
}

.pay-dialog-left {
  min-width: 0;
  padding-right: 20px;
  border-right: 1px solid var(--td-border-level-1-color);
}

.pay-dialog-right {
  min-width: 280px;
  display: flex;
  flex-direction: column;
}

@media (max-width: 960px) {
  .pay-dialog :deep(.t-dialog__wrap) {
    min-width: unset;
  }
}

@media (max-width: 768px) {
  .pay-dialog-content {
    grid-template-columns: 1fr;
    gap: 16px;
  }

  .pay-dialog-left {
    padding-right: 0;
    border-right: none;
    border-bottom: 1px solid var(--td-border-level-1-color);
    padding-bottom: 16px;
  }

  .pay-dialog-right {
    padding-left: 0;
  }
}

/* 统计卡片 */
.stat-card {
  background: linear-gradient(135deg, var(--td-brand-color-light, #e8f4ff) 0%, var(--td-bg-color-secondarycontainer) 60%);
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

.stat-label {
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

.stat-paid {
  color: var(--td-success-color);
}

.stat-balance {
  font-size: 20px;
  color: var(--td-warning-color);
}

.pay-form :deep(.t-form__item) {
  margin-bottom: 14px;
}

/* 文件上传区 */
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

.file-delete-btn {
  cursor: pointer;
  color: var(--td-text-color-placeholder);
  flex-shrink: 0;
  transition: color 0.2s;
  margin-left: auto;
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

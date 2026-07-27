<template>
  <div class="finance-receivable-container">
    <search-filter
      :filters="filterConfig"
      :model-value="searchParams"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <div class="toolbar-actions">
          <t-tooltip content="应收账款由订单确认时自动生成，请在订单管理中确认合同" placement="top">
            <t-button theme="default" disabled>
              <template #icon><add-icon /></template>
              新增应收账款
            </t-button>
          </t-tooltip>
          <t-button theme="default" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button theme="default" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 汇总统计 -->
    <t-card :bordered="false">
      <div class="summary-stats">
        <div class="stat-item">
          <div class="stat-label">总应收金额</div>
          <div class="stat-value">¥{{ formatAmount(summaryStats.totalAmount) }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">已收金额</div>
          <div class="stat-value success">¥{{ formatAmount(summaryStats.totalReceived) }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">待收金额</div>
          <div class="stat-value warning">¥{{ formatAmount(summaryStats.totalBalance) }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">已优惠金额</div>
          <div class="stat-value" style="color: var(--td-warning-color)">
            ¥{{ formatAmount(summaryStats.totalDiscount) }}
          </div>
        </div>
        <div class="stat-item">
          <div class="stat-label">已锁定金额</div>
          <div class="stat-value locked">¥{{ formatAmount(summaryStats.totalLocked) }}</div>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <t-table
        :data="tableData"
        :columns="displayColumns"
        :loading="loading"
        :pagination="pagination"
        row-key="id"
        hover
        @page-change="onPageChange"
      >
        <template #payment_type="{ row }">
          <span
            v-if="row.payment_type || row.source_type"
            class="payment-type-cell"
            style="white-space: nowrap; display: inline-flex; align-items: center"
          >
            <t-tag
              v-if="row.source_type"
              size="small"
              :style="getSourceTypeStyle(row.source_type)"
              style="margin-right: 4px; flex-shrink: 0"
            >
              {{ getSourceTypeText(row.source_type) }}
            </t-tag>
            <t-tag
              v-if="row.payment_type"
              size="small"
              :style="getPaymentTypeStyle(row.payment_type)"
              style="flex-shrink: 0"
            >
              {{ getPaymentTypeText(row.payment_type) }}
            </t-tag>
          </span>
          <span v-else>-</span>
        </template>

        <template #mold_code="{ row }">
          {{ row.mold_code || '-' }}
        </template>

        <template #product_name="{ row }">
          {{ row.product_name || '-' }}
        </template>

        <template #amount="{ row }">
          <span class="amount-text">¥{{ formatAmount(row.amount) }}</span>
        </template>

        <template #received_amount="{ row }">
          <span class="amount-text success">¥{{ formatAmount(row.received_amount) }}</span>
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

        <template #balance="{ row }">
          <span class="amount-text" :class="{ warning: row.balance > 0 }"
            >¥{{ formatAmount(row.balance) }}</span
          >
        </template>

        <template #status="{ row }">
          <div class="status-cell">
            <t-tag :theme="getStatusTheme(row.status)" :style="getStatusStyle(row.status)">
              {{ getStatusText(row.status) }}
            </t-tag>
            <!-- 只有当主状态不是locked，但is_locked为true时才显示锁定标识 -->
            <t-tooltip
              v-if="row.is_locked && row.status !== 'locked'"
              content="该应收账款已锁定，无法进行收款操作"
              placement="top"
            >
              <t-tag theme="warning" size="small" variant="light">锁定</t-tag>
            </t-tooltip>
          </div>
        </template>

        <template #aging_category="{ row }">
          <t-tooltip
            v-if="row.aging_category === 'OVERDUE'"
            content="该款项已逾期，请尽快处理"
            placement="top"
          >
            <t-tag :style="getAgingStyle(row.aging_category)" variant="light">
              {{ getAgingText(row.aging_category) }}
            </t-tag>
          </t-tooltip>
          <t-tag v-else-if="row.aging_category" :style="getAgingStyle(row.aging_category)">
            {{ getAgingText(row.aging_category) }}
          </t-tag>
          <span v-else>-</span>
        </template>

        <template #due_date="{ row }">
          {{ row.due_date ? formatDate(row.due_date) : '-' }}
        </template>

        <template #op="{ row }">
          <div class="operation-cell">
            <t-space :size="8">
              <t-link
                v-if="row.status === 'PENDING' && !row.is_locked"
                theme="default"
                @click="handleEdit(row)"
              >
                编辑
              </t-link>
              <t-link
                v-if="row.balance > 0 && !row.is_locked"
                theme="warning"
                @click="handleReceive(row)"
              >
                收款
              </t-link>
              <t-link v-else-if="row.is_locked" theme="default" disabled> 已锁定 </t-link>
              <t-link theme="primary" @click="handleViewPayments(row)"> 收款记录 </t-link>
            </t-space>
          </div>
        </template>
      </t-table>
    </t-card>

    <!-- 编辑金额对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      header="调增/调减金额"
      width="500px"
      :close-on-overlay-click="false"
      @confirm="handleSubmit"
      @cancel="handleCancel"
    >
      <t-alert theme="info" :close="false" style="margin-bottom: 16px">
        <template #message>
          客户追加需求或协商调价时，可调整应收金额。
          <strong>仅未开始收款的记录允许调整。</strong>
        </template>
      </t-alert>
      <t-form ref="formRef" :data="formData" :rules="rules" label-width="100px">
        <t-form-item v-if="isEdit" label="原金额">
          <span class="amount-text" style="line-height: 32px"
            >¥{{ formatAmount(originalAmount) }}</span
          >
        </t-form-item>

        <t-form-item label="新金额" name="amount">
          <t-input-number
            v-model="formData.amount"
            :min="0"
            :precision="2"
            placeholder="请输入调整后的金额"
            style="width: 100%"
          >
            <template #suffix>元</template>
          </t-input-number>
        </t-form-item>

        <t-form-item label="调整原因" name="remark">
          <t-textarea
            v-model="formData.remark"
            placeholder="请说明调整原因，例如：客户增加需求追加200元"
            :maxlength="500"
            show-limit-number
          />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 列设置对话框 - 使用通用组件 -->
    <column-setting-dialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="finance_receivable_table_columns"
      @change="handleColumnsChange"
    />

    <!-- 收款对话框 -->
    <t-dialog
      v-model:visible="receiveDialogVisible"
      header="收款"
      width="900px"
      placement="center"
      :close-on-overlay-click="false"
      class="receive-dialog"
      @confirm="handleReceiveSubmit"
      @cancel="receiveDialogVisible = false"
    >
      <div class="receive-dialog-content">
        <!-- 左侧：表单信息 -->
        <div class="receive-dialog-left">
          <!-- 金额信息卡片 -->
          <div class="stat-card">
            <div class="stat-grid">
              <div class="stat-item">
                <div class="stat-label">应收总额</div>
                <div class="stat-number">¥{{ formatAmount(currentReceivable?.amount || 0) }}</div>
              </div>
              <div class="stat-item">
                <div class="stat-label">已收金额</div>
                <div class="stat-number stat-received">
                  ¥{{ formatAmount(currentReceivable?.received_amount || 0) }}
                </div>
              </div>
              <div v-if="(currentReceivable?.discount_amount || 0) > 0" class="stat-item">
                <div class="stat-label">已优惠金额</div>
                <div class="stat-number stat-discount">
                  -¥{{ formatAmount(currentReceivable?.discount_amount || 0) }}
                </div>
              </div>
              <div class="stat-item">
                <div class="stat-label stat-label-emphasis">待收余额</div>
                <div class="stat-number stat-balance">
                  ¥{{ formatAmount(currentReceivable?.balance || 0) }}
                </div>
              </div>
            </div>
          </div>

          <!-- 表单字段 -->
          <t-form :data="receiveForm" label-width="90px" class="receive-form">
            <t-form-item label="收款金额" required>
              <t-input-number
                v-model="receiveForm.amount"
                :min="0.01"
                :max="currentReceivable?.balance || 0"
                :precision="2"
                placeholder="请输入收款金额"
                :tips="amountUpperText || '\u00A0'"
                theme="normal"
                style="width: 100%"
                @change="onReceiveAmountChange"
              >
                <template #suffix>元</template>
              </t-input-number>
            </t-form-item>
            <t-form-item label="优惠减免">
              <t-space direction="vertical" style="width: 100%">
                <t-input-number
                  v-model="receiveForm.discount_amount"
                  :min="0"
                  :max="maxDiscountAmount"
                  :precision="2"
                  placeholder="老板同意减免的金额"
                  theme="normal"
                  style="width: 100%"
                >
                  <template #suffix>元</template>
                </t-input-number>
                <t-alert
                  v-if="receiveForm.discount_amount > 0"
                  theme="warning"
                  :close="false"
                  size="small"
                >
                  <template #message>
                    优惠 ¥{{ formatAmount(receiveForm.discount_amount) }}， 实收 ¥{{
                      formatAmount(receiveForm.amount)
                    }}， 合计 ¥{{ formatAmount(receiveForm.amount + receiveForm.discount_amount) }}
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
            <t-form-item label="收款账户">
              <t-select
                v-model="receiveForm.receiving_account"
                placeholder="请选择收款账户"
                clearable
                filterable
                :disabled="filteredAccountOptions.length === 0"
              >
                <t-option
                  v-for="item in filteredAccountOptions"
                  :key="item.id"
                  :value="String(item.id)"
                  :label="item.account_name + (item.account_no ? `（${item.account_no}）` : '')"
                />
              </t-select>
            </t-form-item>
            <t-form-item label="流水号">
              <t-input v-model="receiveForm.transaction_no" placeholder="银行回单号或交易流水号" />
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
              <div class="file-upload-subtitle">支持图片、PDF格式，单个文件不超过5MB</div>
            </div>

            <!-- 回单上传 -->
            <!-- 回单上传 -->
            <div class="file-upload-item">
              <div class="file-upload-label">
                <span class="file-label-text">回单文件</span>
              </div>
              <t-upload
                :key="`receipt-${receiveDialogVisible}`"
                :auto-upload="false"
                multiple
                :size-limit="{ size: 5, unit: 'MB' }"
                accept="image/*,.pdf"
                theme="file"
                @change="(ctx: any) => handleUploadChange('receipt', ctx)"
              >
                <template #fileListDisplay="{ files, onRemove }">
                  <ul v-if="files.length > 0" class="upload-file-list">
                    <li v-for="file in files" :key="file.name" class="upload-file-item">
                      <t-tag size="small" theme="danger" variant="light">PDF</t-tag>
                      <t-link
                        theme="primary"
                        hover="color"
                        class="file-name-link"
                        @click="previewFile(file.raw)"
                        >{{ file.name }}</t-link
                      >
                      <delete-icon class="file-delete-btn" @click.stop="onRemove(file)" />
                    </li>
                  </ul>
                </template>
              </t-upload>
            </div>

            <!-- 发票上传 -->
            <div class="file-upload-item">
              <div class="file-upload-label">
                <span class="file-label-text">发票文件</span>
              </div>
              <t-upload
                :key="`invoice-${receiveDialogVisible}`"
                :auto-upload="false"
                multiple
                :size-limit="{ size: 5, unit: 'MB' }"
                accept="image/*,.pdf"
                theme="file"
                @change="(ctx: any) => handleUploadChange('invoice', ctx)"
              >
                <template #fileListDisplay="{ files, onRemove }">
                  <ul v-if="files.length > 0" class="upload-file-list">
                    <li v-for="file in files" :key="file.name" class="upload-file-item">
                      <t-tag size="small" theme="danger" variant="light">PDF</t-tag>
                      <t-link
                        theme="primary"
                        hover="color"
                        class="file-name-link"
                        @click="previewFile(file.raw)"
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
      v-model:visible="imagePreviewVisible"
      :images="imagePreviewList"
      :default-index="imagePreviewIndex"
    />

    <!-- 文件上传进度 -->
    <t-dialog
      v-model:visible="uploadProgressVisible"
      header="文件上传中"
      width="500px"
      :footer="false"
      :close-btn="false"
      :close-on-overlay-click="false"
      :close-on-esc="false"
      attach="body"
    >
      <div class="upload-progress-list">
        <div v-for="(item, index) in uploadProgressList" :key="index" class="upload-progress-item">
          <div class="upload-progress-info">
            <file-icon />
            <span class="upload-progress-name">{{ item.name }}</span>
            <t-tag v-if="item.status === 'pending'" theme="default" variant="light" size="small"
              >等待中</t-tag
            >
            <t-tag
              v-else-if="item.status === 'uploading'"
              theme="primary"
              variant="light"
              size="small"
              >{{ item.percent }}%</t-tag
            >
            <t-tag
              v-else-if="item.status === 'success'"
              theme="success"
              variant="light"
              size="small"
              >完成</t-tag
            >
            <t-tag v-else-if="item.status === 'fail'" theme="danger" variant="light" size="small"
              >失败</t-tag
            >
          </div>
          <t-progress
            :percentage="item.percent"
            :status="item.status === 'fail' ? 'error' : undefined"
            :stroke-width="6"
          />
        </div>
      </div>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import dayjs from 'dayjs';
import { AddIcon, DeleteIcon, FileIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, nextTick, onMounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { uploadFiles } from '@/api/common/upload';
import type { ReceivableRecord } from '@/api/finance/receivable';
import {
  createReceivable,
  getReceivableList,
  receivePayment,
  updateReceivable,
} from '@/api/finance/receivable';
import { getCustomerList, getReceivingAccountList } from '@/api/reference';
import type { ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import ColumnSettingDialog from '@/components/ColumnSettingDialog/index.vue';
import type { FilterItem } from '@/components/SearchFilter/index.vue';
import SearchFilter from '@/components/SearchFilter/index.vue';
import { useDictRaw } from '@/hooks/useDict';
import { toChineseUpperAmount } from '@/utils/amount-upper';
import { getDictColorRenderConfig } from '@/utils/dict-color';

const router = useRouter();

// 搜索参数
const searchParams = reactive({
  customer_id: undefined as number | undefined,
  status: '',
  source_types: [] as string[],
  payment_types: [] as string[],
});

// 款项来源筛选项（多选：mold=模具, production=生产）
const sourceTypeOptions = [
  { value: 'mold', label: '模具' },
  { value: 'production', label: '生产' },
];

// 表格数据
const loading = ref(false);
const tableData = ref<ReceivableRecord[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
});

// 汇总统计
const summaryStats = reactive({
  totalAmount: 0,
  totalReceived: 0,
  totalBalance: 0,
  totalDiscount: 0,
  totalLocked: 0,
});

// 对话框
const dialogVisible = ref(false);
const isEdit = ref(false);
const originalAmount = ref(0);
const formRef = ref();
const formData = ref<Partial<ReceivableRecord>>({
  customer_id: undefined,
  amount: 0,
  due_date: '',
  remark: '',
});

// 收款对话框
const receiveDialogVisible = ref(false);
const currentReceivable = ref<ReceivableRecord | null>(null);
const receiveForm = reactive<{
  amount: number | undefined;
  discount_amount: number | undefined;
  payment_date: string;
  payment_method: string;
  transaction_no: string;
  receiving_account: string;
  remark: string;
  receipt_file_id: string;
  invoice_file_id: string;
}>({
  amount: undefined,
  discount_amount: undefined,
  payment_date: '',
  payment_method: '',
  transaction_no: '',
  receiving_account: '',
  remark: '',
  receipt_file_id: '',
  invoice_file_id: '',
});

// 最大可优惠金额 = 余额 - 收款金额
const maxDiscountAmount = computed(() => {
  const balance = currentReceivable.value?.balance || 0;
  const canDiscount = balance - (receiveForm.amount || 0);
  return canDiscount > 0 ? canDiscount : 0;
});

// 收款金额大写（实时显示）
const amountUpperText = computed(() => toChineseUpperAmount(receiveForm.amount));

// 当收款金额变化时，自动调整优惠金额不超过上限
const onReceiveAmountChange = () => {
  if (receiveForm.discount_amount > maxDiscountAmount.value) {
    receiveForm.discount_amount = maxDiscountAmount.value;
  }
};

// 上传文件管理
const receiptFiles = ref<File[]>([]);
const invoiceFiles = ref<File[]>([]);

// 图片预览（TDesign ImageViewer）
const imagePreviewVisible = ref(false);
const imagePreviewList = ref<string[]>([]);
const imagePreviewIndex = ref(0);

// 收集当前所有图片的 blob URL，便于在合适时机清理
const imageBlobUrls = ref<string[]>([]);

// 文件上传进度
interface UploadProgressItem {
  name: string;
  percent: number;
  status: 'pending' | 'uploading' | 'success' | 'fail';
}
const uploadProgressVisible = ref(false);
const uploadProgressList = ref<UploadProgressItem[]>([]);

const handleUploadChange = (type: 'receipt' | 'invoice', ctx: any) => {
  // ctx 是 TDesign Upload 的 change 事件的第一个参数 fileList: UploadFile[]
  // 即当前已选择的文件列表数组
  const files = Array.isArray(ctx) ? ctx : ctx?.files || [];
  // 清理旧的图片 blob URL
  imageBlobUrls.value.forEach((url) => URL.revokeObjectURL(url));
  imageBlobUrls.value = [];

  const rawFiles: File[] = [];
  files.forEach((f: any) => {
    if (f.raw instanceof File) rawFiles.push(f.raw);
  });
  if (type === 'receipt') receiptFiles.value = rawFiles;
  else invoiceFiles.value = rawFiles;
};

// 图片用 ImageViewer，PDF 等用浏览器新标签页
const previewFile = (raw: File | undefined) => {
  if (!raw) return;

  if (raw.type.startsWith('image/')) {
    const reader = new FileReader();
    reader.onload = () => {
      const blob = new Blob([reader.result as ArrayBuffer], { type: raw.type });
      const url = URL.createObjectURL(blob);
      imageBlobUrls.value.push(url);
      imagePreviewList.value = [url];
      imagePreviewIndex.value = 0;
      imagePreviewVisible.value = true;
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
    reader.onerror = () => {
      MessagePlugin.warning('文件读取失败，请重新选择文件');
    };
    reader.readAsArrayBuffer(raw);
  }
};

// 客户列表
const customerList = ref<any[]>([]);

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: paymentMethodOptions } = useDictRaw('payment-method');
const { items: receivableStatusOptions } = useDictRaw('receivable-status');
const { items: agingCategoryOptions } = useDictRaw('aging-category');
const { items: paymentTypeOptions } = useDictRaw('payment-type');

// 收款账户选项（来自 getReceivingAccountList API，不是字典）
const receivingAccountOptions = ref<any[]>([]);

// 根据收款方式获取对应的账户类型列表（从字典 ext 配置读取）
const getAccountTypesByMethod = (method: string): string[] => {
  if (!method) return [];
  const item = paymentMethodOptions.value.find((item) => item.dict_value === method) as any;
  if (!item || !item.ext) return [];
  try {
    const ext = typeof item.ext === 'string' ? JSON.parse(item.ext) : item.ext;
    return ext.account_types || [];
  } catch {
    return [];
  }
};

// 根据选择的收款方式过滤收款账户
const filteredAccountOptions = computed(() => {
  const method = receiveForm.payment_method;
  if (!method) return receivingAccountOptions.value; // 未选择方式时全部展示
  const allowedTypes = getAccountTypesByMethod(method);
  if (allowedTypes.length === 0) return []; // 现金等无账户方式
  return receivingAccountOptions.value.filter((item: any) =>
    allowedTypes.includes(item.account_type),
  );
});

// 监听收款方式变化，联动清空不匹配的收款账户
watch(
  () => receiveForm.payment_method,
  (newMethod) => {
    if (!receiveForm.receiving_account) return;
    if (!newMethod) return;
    const allowedTypes = getAccountTypesByMethod(newMethod);
    if (allowedTypes.length === 0) {
      receiveForm.receiving_account = '';
    } else {
      const current = receivingAccountOptions.value.find(
        (item: any) => item.id === receiveForm.receiving_account,
      );
      if (current && !allowedTypes.includes(current.account_type)) {
        receiveForm.receiving_account = '';
      }
    }
  },
);

// 表单验证规则
const rules = {
  amount: [{ required: true, message: '请输入应收金额', type: 'error' as const }],
};

// 所有可选列配置（用于列管理）
const allColumns: ColumnOption[] = [
  { colKey: 'order_code', title: '订单编号', pinned: true },
  { colKey: 'customer.cus_name', title: '客户名称' },
  { colKey: 'payment_type', title: '款项类型' },
  { colKey: 'amount', title: '应收金额' },
  { colKey: 'received_amount', title: '已收金额' },
  { colKey: 'discount_amount', title: '优惠金额' },
  { colKey: 'balance', title: '余额' },
  { colKey: 'status', title: '状态' },
  { colKey: 'due_date', title: '到期日期' },
  { colKey: 'aging_category', title: '账龄分类' },
  { colKey: 'mold_code', title: '模具编号' },
  { colKey: 'product_name', title: '产品名称' },
  { colKey: 'create_time', title: '创建时间' },
];

// 默认显示的列（按业务优先级排序）
const DEFAULT_COLUMNS = [
  'order_code',
  'customer.cus_name',
  'payment_type',
  'amount',
  'received_amount',
  'balance',
  'status',
];

// 动态列配置（根据用户选择显示）
const selectedColumns = ref<string[]>([]);
const columnSettingRef = ref();

// 原始列配置（不包含ID列）
const baseColumns: PrimaryTableCol[] = [
  { colKey: 'order_code', title: '订单编号', width: 160, fixed: 'left' },
  { colKey: 'customer.cus_name', title: '客户名称', width: 120 },
  { colKey: 'payment_type', title: '款项类型', width: 160, minWidth: 140 },
  { colKey: 'amount', title: '应收金额', width: 120 },
  { colKey: 'received_amount', title: '已收金额', width: 120 },
  { colKey: 'discount_amount', title: '优惠金额', width: 120 },
  { colKey: 'balance', title: '余额', width: 120 },
  { colKey: 'status', title: '状态', width: 150 },
  { colKey: 'due_date', title: '到期日期', width: 120 },
  { colKey: 'aging_category', title: '账龄分类', width: 100 },
  { colKey: 'mold_code', title: '模具编号', width: 140 },
  { colKey: 'product_name', title: '产品名称', width: 160 },
  { colKey: 'create_time', title: '创建时间', width: 160 },
  { colKey: 'op', title: '操作', width: 160, fixed: 'right' },
];

const displayColumns = computed(() => {
  const orderedColumns = selectedColumns.value
    .map((colKey) => baseColumns.find((col) => col.colKey === colKey))
    .filter((col) => col !== undefined);

  // 确保操作列始终在最后且固定右侧
  const operationCol = baseColumns.find((col) => col.colKey === 'op');
  if (operationCol) {
    orderedColumns.push(operationCol);
  }

  return orderedColumns;
});

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('finance_receivable_table_columns');
    if (saved) {
      selectedColumns.value = JSON.parse(saved);
    } else {
      selectedColumns.value = [...DEFAULT_COLUMNS];
    }
  } catch (error) {
    console.error('加载列配置失败:', error);
    selectedColumns.value = [...DEFAULT_COLUMNS];
  }
};

const saveColumnConfig = () => {
  try {
    localStorage.setItem('finance_receivable_table_columns', JSON.stringify(selectedColumns.value));
  } catch (error) {
    console.error('保存列配置失败:', error);
  }
};

const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
  saveColumnConfig();
};

// 打开列设置对话框
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getReceivableList({
      page: pagination.current,
      page_size: pagination.pageSize,
      ...searchParams,
    });

    // request 封装在 code===200 时会直接返回 data.data
    const list = Array.isArray((res as any)?.list) ? (res as any).list : [];
    tableData.value = list;
    pagination.total = typeof (res as any)?.total === 'number' ? (res as any).total : 0;

    calculateSummaryStats(list);
  } catch (error) {
    console.error('加载数据失败:', error);
    MessagePlugin.error('加载数据失败');
  } finally {
    loading.value = false;
  }
};

const calculateSummaryStats = (data: ReceivableRecord[]) => {
  summaryStats.totalAmount = data.reduce((sum, item) => sum + (item.amount || 0), 0);
  summaryStats.totalReceived = data.reduce((sum, item) => sum + (item.received_amount || 0), 0);
  summaryStats.totalDiscount = data.reduce((sum, item) => sum + (item.discount_amount || 0), 0);
  summaryStats.totalBalance = data.reduce((sum, item) => sum + (item.balance || 0), 0);
  summaryStats.totalLocked = data
    .filter((item) => item.is_locked)
    .reduce((sum, item) => sum + (item.balance || 0), 0);
};

const loadCustomerList = async () => {
  try {
    const res = await getCustomerList({ page: 1, page_size: 1000 });
    if (res.code === 200 && res.data) {
      customerList.value = res.data.list || [];
    }
  } catch (error) {
    console.error('加载客户列表失败:', error);
  }
};

const loadReceivingAccountOptions = async () => {
  try {
    const res = await getReceivingAccountList();
    receivingAccountOptions.value = res.list || [];
  } catch (error) {
    console.error('加载收款账户选项失败:', error);
  }
};

// 加载收款方式字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载应收状态字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载账龄分类字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载款项类型字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  loadData();
};

const handleReset = () => {
  searchParams.customer_id = undefined;
  searchParams.status = '';
  searchParams.source_types = [];
  searchParams.payment_types = [];
  handleSearch();
};

// 刷新
const handleRefresh = () => {
  loadData();
};

// 分页变化
const onPageChange = (pageInfo: any) => {
  pagination.current = pageInfo.current;
  pagination.pageSize = pageInfo.pageSize;
  loadData();
};

// 编辑
const handleEdit = (row: ReceivableRecord) => {
  isEdit.value = true;
  originalAmount.value = row.amount;
  formData.value = {
    id: row.id,
    customer_id: row.customer_id,
    amount: row.amount,
    due_date: row.due_date,
    remark: row.remark || '',
  };
  dialogVisible.value = true;
};

// 提交表单
const handleSubmit = async () => {
  const validateResult = await formRef.value?.validate();
  if (!validateResult) return;

  try {
    if (isEdit.value && formData.value.id) {
      await updateReceivable(formData.value.id, formData.value);
      MessagePlugin.success('更新成功');
    } else {
      await createReceivable(formData.value);
      MessagePlugin.success('创建成功');
    }
    dialogVisible.value = false;
    loadData();
  } catch (error: any) {
    MessagePlugin.error(error.message || '操作失败');
  }
};

// 取消
const handleCancel = () => {
  dialogVisible.value = false;
};

// 收款
const handleReceive = (row: ReceivableRecord) => {
  currentReceivable.value = row;
  receiveForm.amount = undefined;
  receiveForm.discount_amount = undefined;
  receiveForm.receipt_file_id = '';
  receiveForm.invoice_file_id = '';
  receiptFiles.value = [];
  invoiceFiles.value = [];
  imageBlobUrls.value.forEach((url) => URL.revokeObjectURL(url));
  imageBlobUrls.value = [];
  receiveDialogVisible.value = true;
};

// 查看收款记录
const handleViewPayments = (row: ReceivableRecord) => {
  router.push({ name: 'FinanceReceivablePayments', params: { id: row.id } });
};

// 提交收款（支持多文件逐一上传 + 进度展示）
const handleReceiveSubmit = async () => {
  if (!currentReceivable.value || !receiveForm.amount || receiveForm.amount <= 0) {
    MessagePlugin.warning('请输入有效的收款金额');
    return;
  }

  const allFiles: { file: File; type: string }[] = [];
  receiptFiles.value.forEach((f) => allFiles.push({ file: f, type: 'receipt' }));
  invoiceFiles.value.forEach((f) => allFiles.push({ file: f, type: 'invoice' }));

  try {
    let receiptFileId = '';
    let invoiceFileId = '';

    if (allFiles.length > 0) {
      // 先初始化进度列表，让对话框渲染
      uploadProgressList.value = allFiles.map((f) => ({
        name: f.file.name,
        percent: 0,
        status: 'pending' as const,
      }));
      uploadProgressVisible.value = true;
      // 确保对话框 DOM 渲染完成后再开始上传
      await nextTick();

      // 逐文件上传，每个文件独立跟踪进度
      const uploadedIds: { id: string; type: string }[] = [];
      for (let i = 0; i < allFiles.length; i++) {
        const { file, type } = allFiles[i];
        uploadProgressList.value[i].status = 'uploading';
        // 让 UI 先刷新状态
        await nextTick();

        const results = await uploadFiles({
          files: [file],
          category: 'receivable-payment',
          onProgress: (percent: number) => {
            uploadProgressList.value[i].percent = percent;
          },
        });
        uploadedIds.push({ id: results[0].id, type });
        uploadProgressList.value[i].percent = 100;
        uploadProgressList.value[i].status = 'success';
      }

      receiptFileId = uploadedIds
        .filter((x) => x.type === 'receipt')
        .map((x) => x.id)
        .join(',');
      invoiceFileId = uploadedIds
        .filter((x) => x.type === 'invoice')
        .map((x) => x.id)
        .join(',');
    }

    // 提交收款（进度弹窗保持打开，让用户看到全部完成的状态）
    await receivePayment(currentReceivable.value.id, {
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

    // 全部完成后，一起关闭进度弹窗和收款弹窗
    uploadProgressVisible.value = false;
    receiveDialogVisible.value = false;
    MessagePlugin.success('收款成功');
    receiveForm.amount = undefined;
    receiveForm.discount_amount = undefined;
    receiveForm.payment_date = '';
    receiveForm.payment_method = '';
    receiveForm.transaction_no = '';
    receiveForm.receiving_account = '';
    receiveForm.remark = '';
    receiveForm.receipt_file_id = '';
    receiveForm.invoice_file_id = '';
    receiptFiles.value = [];
    invoiceFiles.value = [];
    imageBlobUrls.value.forEach((url) => URL.revokeObjectURL(url));
    imageBlobUrls.value = [];
    loadData();
  } catch (error: any) {
    MessagePlugin.error(error.message || '收款失败');
  }
};

// 格式化金额
const formatAmount = (amount: number) => {
  return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

// 格式化日期
const formatDate = (date: string) => {
  return date ? dayjs(date).format('YYYY-MM-DD') : '-';
};

const getStatusStyle = (status: string) => {
  if (status === 'locked') return {};
  const option = receivableStatusOptions.value.find(
    (item: any) => item.dict_value === status,
  ) as any;
  if (option?.color) {
    const config = getDictColorRenderConfig(option.color);
    return config.mode === 'style' ? config.style || {} : {};
  }
  return {};
};

const getStatusTheme = (status: string) => {
  if (status === 'locked') return undefined;
  const option = receivableStatusOptions.value.find(
    (item: any) => item.dict_value === status,
  ) as any;
  if (option?.color) {
    return getDictColorRenderConfig(option.color).theme;
  }
  const themeMap: Record<string, string> = {
    pending: 'default',
    partial: 'warning',
    paid: 'success',
    overdue: 'danger',
  };
  return themeMap[status?.toLowerCase()] as any;
};

const getStatusText = (status: string) => {
  // 特殊处理 locked 状态
  if (status === 'locked') return '已锁定';

  // 从应收状态字典中查找对应的标签
  const option = receivableStatusOptions.value.find((item: any) => item.dict_value === status);
  if (option && option.dict_label) {
    return option.dict_label;
  }

  // 降级到默认文本映射（键与字典 dict_value 统一为小写）
  const textMap: Record<string, string> = {
    pending: '待收款',
    partial: '部分收款',
    paid: '已收款',
    overdue: '逾期',
  };
  return textMap[status?.toLowerCase()] || status;
};

const getAgingStyle = (category: string) => {
  // 从账龄分类字典中查找对应的颜色配置
  const option = agingCategoryOptions.value.find((item) => item.dict_value === category) as any;
  if (option && option.color) {
    return getDictColorRenderConfig(option.color).style || {};
  }

  // 降级到默认主题映射
  const styleMap: Record<string, any> = {
    CURRENT: getDictColorRenderConfig('success').style || {},
    WITHIN_30: {},
    WITHIN_60: getDictColorRenderConfig('warning').style || {},
    WITHIN_90: getDictColorRenderConfig('warning').style || {},
    OVERDUE: getDictColorRenderConfig('danger').style || {},
  };
  return styleMap[category] || {};
};

const getAgingText = (category: string) => {
  // 从账龄分类字典中查找对应的标签
  const option = agingCategoryOptions.value.find((item: any) => item.dict_value === category);
  if (option && option.dict_label) {
    return option.dict_label;
  }

  // 降级到默认文本映射
  const textMap: Record<string, string> = {
    CURRENT: '当期',
    WITHIN_30: '30天内',
    WITHIN_60: '60天内',
    WITHIN_90: '90天内',
    OVERDUE: '逾期',
  };
  return textMap[category] || category;
};

const getPaymentTypeStyle = (type: string) => {
  const option = paymentTypeOptions.value.find((item) => item.dict_value === type) as any;
  if (option && option.color) {
    return getDictColorRenderConfig(option.color).style || {};
  }

  // 降级到默认主题映射
  const styleMap: Record<string, any> = {
    FIRST_PAYMENT: getDictColorRenderConfig('primary').style || {},
    MIDDLE_PAYMENT: getDictColorRenderConfig('warning').style || {},
    FINAL_PAYMENT: getDictColorRenderConfig('success').style || {},
    total: {},
  };
  return styleMap[type] || {};
};

const getPaymentTypeText = (type: string) => {
  const option = paymentTypeOptions.value.find((item) => item.dict_value === type);
  return option ? option.dict_label : type;
};

// ✨ 获取款项来源样式
const getSourceTypeStyle = (type: string) => {
  const styleMap: Record<string, any> = {
    mold: { background: '#e8f4ff', color: '#0052d9', border: '1px solid #b8dcff' },
    production: { background: '#e8f8e8', color: '#00a870', border: '1px solid #b8e8b8' },
  };
  return styleMap[type] || {};
};

// ✨ 获取款项来源文本
const getSourceTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    mold: '模具',
    production: '生产',
  };
  return textMap[type] || type;
};

// 筛选栏配置
const filterConfig: FilterItem[] = [
  {
    type: 'select',
    field: 'customer_id',
    placeholder: '请选择客户',
    options: customerList,
    labelKey: 'cus_name',
    valueKey: 'id',
  },
  {
    type: 'select',
    field: 'status',
    placeholder: '请选择状态',
    options: receivableStatusOptions,
    labelKey: 'dict_label',
    valueKey: 'dict_value',
  },
  {
    type: 'select',
    field: 'source_types',
    placeholder: '请选择款项来源（可多选）',
    options: sourceTypeOptions,
    multiple: true,
  },
  {
    type: 'select',
    field: 'payment_types',
    placeholder: '请选择款项类型（可多选）',
    options: paymentTypeOptions,
    labelKey: 'dict_label',
    valueKey: 'dict_value',
    multiple: true,
  },
];

onMounted(() => {
  loadData();
  loadCustomerList();
  loadReceivingAccountOptions();
  loadColumnConfig();
});
</script>
<style scoped>
.finance-receivable-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
}

.finance-receivable-container :deep(.search-card) {
  margin-bottom: 0;
}

.toolbar-card {
  margin-bottom: 0;
}

.search-form {
  margin-bottom: 0;
}

.toolbar-content {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

/* 汇总统计样式 */
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

.stat-value.locked {
  color: var(--td-gray-color-7);
}

.amount-text {
  font-weight: 500;
}

.amount-text.success {
  color: var(--td-success-color);
}

.amount-text.warning {
  color: var(--td-warning-color);
}

/* 状态单元格 */
.status-cell {
  display: flex;
  gap: 4px;
  align-items: center;
}

/* 操作列单元格 */
.operation-cell {
  min-width: 140px;
  display: flex;
  align-items: center;
}

/* 收款对话框 - 响应式布局 */
.receive-dialog-content {
  display: grid;
  grid-template-columns: 1fr 320px;
  gap: 20px;
  min-height: 420px;
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
@media (max-width: 960px) {
  .receive-dialog :deep(.t-dialog__wrap) {
    min-width: unset;
  }
}

/* 响应式：窄屏幕下单栏布局 */
@media (max-width: 768px) {
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

.stat-item {
  text-align: center;
  padding: 6px 0;
}

.stat-item:last-child {
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

.file-name-link {
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

/* 上传文件列表（旧样式保留兼容） */
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
  margin-left: auto;
  cursor: pointer;
  color: var(--td-text-color-placeholder);
  flex-shrink: 0;
  transition: color 0.2s;
}

.file-delete-btn:hover {
  color: var(--td-error-color);
}

/* 上传进度弹窗 */
.upload-progress-list {
  max-height: 400px;
  overflow-y: auto;
}

.upload-progress-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 12px 0;
  border-bottom: 1px solid var(--td-border-level-1-color);
}

.upload-progress-item:last-child {
  border-bottom: none;
}

.upload-progress-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.upload-progress-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>

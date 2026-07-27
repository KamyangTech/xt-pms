<template>
  <div class="finance-payable-container">
    <search-filter
      :filters="filterConfig"
      :model-value="searchParams"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">应付账款管理</span>
        <div class="toolbar-actions">
          <t-button theme="primary" @click="handleAdd">
            <template #icon><add-icon /></template>
            新增应付账款
          </t-button>
          <t-button variant="outline" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button variant="outline" @click="handleColumnSetting">
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
          <div class="stat-label">总应付金额</div>
          <div class="stat-value">¥{{ formatAmount(summaryStats.totalAmount) }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">已付金额</div>
          <div class="stat-value success">¥{{ formatAmount(summaryStats.totalPaid) }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">待付金额</div>
          <div class="stat-value warning">¥{{ formatAmount(summaryStats.totalBalance) }}</div>
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
        stripe
        @page-change="onPageChange"
      >
        <template #supplier_name="{ row }">
          {{ row.supplier?.sup_name || row.processor?.processor_name || '-' }}
        </template>

        <template #order_code="{ row }">
          {{ row.order_code || '-' }}
        </template>

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

        <template #amount="{ row }">
          <span class="amount-text">¥{{ formatAmount(row.amount) }}</span>
        </template>

        <template #paid_amount="{ row }">
          <span class="amount-text success">¥{{ formatAmount(row.paid_amount) }}</span>
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
          <t-tag :theme="getStatusTheme(row.status)" :style="getStatusStyle(row.status)">
            {{ getStatusText(row.status) }}
          </t-tag>
        </template>

        <template #due_date="{ row }">
          {{ row.due_date ? formatDate(row.due_date) : '-' }}
        </template>

        <template #op="{ row }">
          <t-space>
            <t-link theme="primary" @click="handleView(row)">详情</t-link>
            <t-link v-if="row.balance > 0" theme="primary" @click="handlePay(row)"> 付款 </t-link>
            <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
            <t-popconfirm content="确认删除该记录？" @confirm="handleDelete(row)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 新增/编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :header="isEdit ? '编辑应付账款' : '新增应付账款'"
      width="600px"
      :close-on-overlay-click="false"
      @confirm="handleSubmit"
      @cancel="handleCancel"
    >
      <t-form ref="formRef" :data="formData" :rules="rules" label-width="120px">
        <t-form-item label="供应商" name="supplier_id">
          <!-- 编辑模式：从订单带出，禁止修改 -->
          <t-input
            v-if="isEdit"
            :value="
              formData.supplier?.sup_name ||
              formData.processor?.processor_name ||
              (formData.supplier_id ? `供应商#${formData.supplier_id}` : '')
            "
            placeholder="关联订单自动带出"
            readonly
            disabled
          />
          <!-- 新增模式：允许选择 -->
          <t-select v-else v-model="formData.supplier_id" placeholder="请选择供应商" clearable>
            <t-option
              v-for="supplier in supplierList"
              :key="supplier.id"
              :value="supplier.id"
              :label="supplier.sup_name"
            />
          </t-select>
        </t-form-item>

        <t-form-item label="应付金额" name="amount">
          <t-input-number
            v-model="formData.amount"
            :min="0"
            :precision="2"
            placeholder="请输入应付金额"
            style="width: 100%"
          >
            <template #suffix>元</template>
          </t-input-number>
        </t-form-item>

        <t-form-item label="到期日期" name="due_date">
          <t-date-picker
            v-model="formData.due_date"
            placeholder="请选择到期日期"
            style="width: 100%"
          />
        </t-form-item>

        <t-form-item label="备注" name="remark">
          <t-textarea
            v-model="formData.remark"
            placeholder="请输入备注信息"
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
      storage-key="finance_payable_table_columns"
      @change="handleColumnsChange"
    />

    <!-- 付款对话框 -->
    <t-dialog
      v-model:visible="payDialogVisible"
      header="付款"
      width="900px"
      placement="center"
      :close-on-overlay-click="false"
      class="pay-dialog"
      @confirm="handlePaySubmit"
      @cancel="payDialogVisible = false"
    >
      <div class="pay-dialog-content">
        <!-- 左侧：表单信息 -->
        <div class="pay-dialog-left">
          <div class="stat-card">
            <div class="stat-grid">
              <div class="stat-item">
                <div class="stat-label">应付总额</div>
                <div class="stat-number">¥{{ formatAmount(currentPayable?.amount || 0) }}</div>
              </div>
              <div class="stat-item">
                <div class="stat-label">已付金额</div>
                <div class="stat-number stat-paid">
                  ¥{{ formatAmount(currentPayable?.paid_amount || 0) }}
                </div>
              </div>
              <div class="stat-item">
                <div class="stat-label stat-label-emphasis">待付余额</div>
                <div class="stat-number stat-balance">
                  ¥{{ formatAmount(currentPayable?.balance || 0) }}
                </div>
              </div>
            </div>
          </div>

          <t-form :data="payForm" label-width="90px" class="pay-form">
            <t-form-item label="付款金额" required>
              <t-input-number
                v-model="payForm.amount"
                :min="0.01"
                :max="currentPayable?.balance || 0"
                :precision="2"
                placeholder="请输入付款金额"
                :tips="amountUpperText || '\u00A0'"
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
                <t-alert
                  v-if="payForm.discount_amount > 0"
                  theme="warning"
                  :close="false"
                  size="small"
                >
                  <template #message>
                    优惠 ¥{{ formatAmount(payForm.discount_amount) }}， 实付 ¥{{
                      formatAmount(payForm.amount)
                    }}， 合计 ¥{{ formatAmount(payForm.amount + payForm.discount_amount) }}
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
              <t-select
                v-model="payForm.payment_account"
                placeholder="请选择付款账户"
                clearable
                filterable
                :disabled="filteredAccountOptions.length === 0"
              >
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
              <t-textarea
                v-model="payForm.remark"
                placeholder="请输入备注信息"
                :autosize="{ minRows: 2, maxRows: 3 }"
              />
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
                      <t-link
                        theme="primary"
                        hover="color"
                        class="file-name-link"
                        @click="previewPayFile(file.raw)"
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
                      <t-link
                        theme="primary"
                        hover="color"
                        class="file-name-link"
                        @click="previewPayFile(file.raw)"
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
  </div>
</template>
<script setup lang="ts">
import dayjs from 'dayjs';
import { AddIcon, DeleteIcon, FileIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { uploadFiles } from '@/api/common/upload';
import type { PayableRecord } from '@/api/finance/payable';
import {
  createPayable,
  deletePayable,
  getPayableList,
  payPayable,
  updatePayable,
} from '@/api/finance/payable';
import { getSupplierList } from '@/api/reference';
import type { ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import ColumnSettingDialog from '@/components/ColumnSettingDialog/index.vue';
import type { FilterItem } from '@/components/SearchFilter/index.vue';
import SearchFilter from '@/components/SearchFilter/index.vue';
import { useDictRaw } from '@/hooks/useDict';
import { toChineseUpperAmount } from '@/utils/amount-upper';
import { getDictColorRenderConfig } from '@/utils/dict-color';

// 搜索参数
const searchParams = reactive({
  supplier_id: undefined as number | undefined,
  status: '',
});

// 表格数据
const loading = ref(false);
const tableData = ref<PayableRecord[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
});

// 汇总统计
const summaryStats = reactive({
  totalAmount: 0,
  totalPaid: 0,
  totalDiscount: 0,
  totalBalance: 0,
  totalLocked: 0,
});

// 对话框
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref();
const formData = ref<Partial<PayableRecord>>({
  supplier_id: undefined,
  amount: undefined,
  due_date: '',
  remark: '',
});

// 付款对话框
const payDialogVisible = ref(false);
const currentPayable = ref<PayableRecord | null>(null);
const payForm = reactive<{
  amount: number | undefined;
  discount_amount: number | undefined;
  payment_date: string;
  payment_method: string;
  transaction_no: string;
  payment_account: string;
  receiver_account: string;
  remark: string;
}>({
  amount: undefined,
  discount_amount: undefined,
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
  const balance = currentPayable.value?.balance || 0;
  const canDiscount = balance - (payForm.amount || 0);
  return canDiscount > 0 ? canDiscount : 0;
});

// 付款金额大写（实时显示）
const amountUpperText = computed(() => toChineseUpperAmount(payForm.amount));

// 当付款金额变化时，自动调整优惠金额不超过上限
const onPayAmountChange = () => {
  if (payForm.discount_amount > maxDiscountAmount.value) {
    payForm.discount_amount = maxDiscountAmount.value;
  }
};

// 供应商列表
const supplierList = ref<any[]>([]);

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: paymentMethodOptions } = useDictRaw('payment-method');
const { items: receivingAccountOptions } = useDictRaw('receiving-account');
const { items: payableStatusOptions } = useDictRaw('payable-status');

// 根据付款方式获取对应的账户类型列表（从字典 ext 配置读取）
const getAccountTypesByMethod = (method: string): string[] => {
  if (!method) return [];
  const item = paymentMethodOptions.value.find((item: any) => item.dict_value === method);
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
  if (!method) return receivingAccountOptions.value; // 未选择方式时全部展示
  const allowedTypes = getAccountTypesByMethod(method);
  if (allowedTypes.length === 0) return []; // 现金等无账户方式
  return receivingAccountOptions.value.filter((item: any) =>
    allowedTypes.includes(item.dict_value),
  );
});

// 监听付款方式变化，联动清空不匹配的付款账户
watch(
  () => payForm.payment_method,
  (newMethod) => {
    if (!payForm.payment_account) return;
    if (!newMethod) return;
    const allowedTypes = getAccountTypesByMethod(newMethod);
    if (allowedTypes.length === 0) {
      payForm.payment_account = '';
    } else {
      const current = receivingAccountOptions.value.find(
        (item: any) => item.dict_value === payForm.payment_account,
      );
      if (current && !allowedTypes.includes(current.dict_value)) {
        payForm.payment_account = '';
      }
    }
  },
);

// 应付状态字典（合并到上方 useDictRaw 声明中）

// 表单验证规则
const rules = {
  supplier_id: [{ required: true, message: '请选择供应商', type: 'error' as const }],
  amount: [{ required: true, message: '请输入应付金额', type: 'error' as const }],
};

// 所有可选列配置（用于列管理）
const allColumns: ColumnOption[] = [
  { colKey: 'id', title: 'ID' },
  { colKey: 'supplier_name', title: '供应商' },
  { colKey: 'order_code', title: '订单编号' },
  { colKey: 'payment_type', title: '款项类型' },
  { colKey: 'amount', title: '应付金额' },
  { colKey: 'paid_amount', title: '已付金额' },
  { colKey: 'discount_amount', title: '优惠金额' },
  { colKey: 'balance', title: '余额' },
  { colKey: 'due_date', title: '到期日期' },
  { colKey: 'status', title: '付款状态' },
  { colKey: 'remark', title: '备注' },
  { colKey: 'create_time', title: '创建时间' },
];

// 默认显示的列
const DEFAULT_COLUMNS = [
  'order_code',
  'payment_type',
  'supplier_name',
  'amount',
  'paid_amount',
  'balance',
  'due_date',
  'status',
];

// 动态列配置（根据用户选择显示）
const selectedColumns = ref<string[]>([]);
const columnSettingRef = ref();

// 原始列配置
const baseColumns: PrimaryTableCol[] = [
  { colKey: 'id', title: 'ID', width: 80 },
  { colKey: 'supplier_name', title: '供应商', width: 150 },
  { colKey: 'order_code', title: '订单编号', width: 160, fixed: 'left' },
  { colKey: 'payment_type', title: '款项类型', width: 160 },
  { colKey: 'amount', title: '应付金额', width: 120 },
  { colKey: 'paid_amount', title: '已付金额', width: 120 },
  { colKey: 'discount_amount', title: '优惠金额', width: 120 },
  { colKey: 'balance', title: '余额', width: 120 },
  { colKey: 'due_date', title: '到期日期', width: 120 },
  { colKey: 'status', title: '付款状态', width: 100 },
  { colKey: 'remark', title: '备注', width: 200 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
  { colKey: 'op', title: '操作', width: 240, fixed: 'right' },
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
    const saved = localStorage.getItem('finance_payable_table_columns');
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
    localStorage.setItem('finance_payable_table_columns', JSON.stringify(selectedColumns.value));
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
    const res = await getPayableList({
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

const calculateSummaryStats = (data: PayableRecord[]) => {
  summaryStats.totalAmount = data.reduce((sum, item) => sum + (item.amount || 0), 0);
  summaryStats.totalPaid = data.reduce((sum, item) => sum + (item.paid_amount || 0), 0);
  summaryStats.totalBalance = data.reduce((sum, item) => sum + (item.balance || 0), 0);
  summaryStats.totalDiscount = data.reduce((sum, item) => sum + (item.discount_amount || 0), 0);
  summaryStats.totalLocked = 0; // 应付账款暂不支持锁定
};

const loadSupplierList = async () => {
  try {
    const res = await getSupplierList({ page: 1, page_size: 1000 });

    // request 封装在 code===200 时会直接返回 data.data
    const list = Array.isArray((res as any)?.list) ? (res as any).list : [];
    supplierList.value = list;
  } catch (error) {
    console.error('加载供应商列表失败:', error);
  }
};

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  loadData();
};

const handleReset = () => {
  searchParams.supplier_id = undefined;
  searchParams.status = '';
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

// 新增
const handleAdd = () => {
  isEdit.value = false;
  formData.value = {
    supplier_id: undefined,
    amount: undefined,
    due_date: '',
    remark: '',
  };
  dialogVisible.value = true;
};

// 编辑
const handleEdit = (row: PayableRecord) => {
  isEdit.value = true;
  formData.value = { ...row };
  dialogVisible.value = true;
};

// 提交表单
const handleSubmit = async () => {
  const validateResult = await formRef.value?.validate();
  if (!validateResult) return;

  try {
    if (isEdit.value && formData.value.id) {
      await updatePayable(formData.value.id, formData.value);
      MessagePlugin.success('更新成功');
    } else {
      await createPayable(formData.value);
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

// 路由
const router = useRouter();

// 查看详情
const handleView = (row: PayableRecord) => {
  router.push({ name: 'FinancePayablePayments', params: { id: row.id } });
};

const handleDelete = async (row: PayableRecord) => {
  try {
    await deletePayable(row.id);
    MessagePlugin.success('删除成功');
    loadData();
  } catch (error: any) {
    MessagePlugin.error(error.message || '删除失败');
  }
};

// 付款
const handlePay = (row: PayableRecord) => {
  currentPayable.value = row;
  payForm.amount = undefined;
  payForm.discount_amount = undefined;
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
  const files = Array.isArray(ctx) ? ctx : ctx?.files || [];
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
  if (
    !currentPayable.value ||
    ((!payForm.amount || payForm.amount <= 0) &&
      (!payForm.discount_amount || payForm.discount_amount <= 0))
  ) {
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
    await payPayable(currentPayable.value.id, {
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
    payForm.amount = undefined;
    payForm.discount_amount = undefined;
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

// 格式化金额
const formatAmount = (amount: number) => {
  return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

// 格式化日期
const formatDate = (date: string) => {
  return date ? dayjs(date).format('YYYY-MM-DD') : '-';
};

const getStatusStyle = (status: string) => {
  const option = payableStatusOptions.value.find((item: any) => item.dict_value === status);
  if (option?.color) {
    const config = getDictColorRenderConfig(option.color);
    return config.mode === 'style' ? config.style || {} : {};
  }
  return {};
};

const getStatusTheme = (status: string) => {
  const option = payableStatusOptions.value.find((item: any) => item.dict_value === status);
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
  // 从应付状态字典中查找对应的标签
  const option = payableStatusOptions.value.find((item: any) => item.dict_value === status);
  if (option && option.dict_label) {
    return option.dict_label;
  }

  // 降级到默认文本映射（键与字典 dict_value 统一为小写）
  const textMap: Record<string, string> = {
    pending: '待付款',
    partial: '部分付款',
    paid: '已付款',
    overdue: '逾期',
  };
  return textMap[status?.toLowerCase()] || status;
};

// ✨ 获取款项来源文本（与应收账款页面一致）
const getSourceTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    mold: '模具',
    production: '生产',
    purchase: '采购',
    outsource: '外发',
  };
  return textMap[type] || textMap[type?.toLowerCase()] || type;
};

// ✨ 获取款项来源样式（与应收账款页面一致）
const getSourceTypeStyle = (type: string) => {
  const styleMap: Record<string, any> = {
    mold: { background: '#e8f4ff', color: '#0052d9', border: '1px solid #b8dcff' },
    production: { background: '#e8f8e8', color: '#00a870', border: '1px solid #b8e8b8' },
    purchase: getDictColorRenderConfig('default').style || {},
    outsource: getDictColorRenderConfig('warning').style || {},
  };
  return styleMap[type] || styleMap[type?.toLowerCase()] || {};
};

// ✨ 获取款项类型文本
const getPaymentTypeText = (type: string) => {
  const textMap: Record<string, string> = {
    total: '全款',
    first: '首期款',
    final: '尾款',
    deposit: '首付款',
    FIRST_PAYMENT: '首付款',
    FINAL_PAYMENT: '尾款',
    installment: '分期款',
    production: '生产款',
  };
  // 先精确匹配，再尝试小写匹配，兜底显示原始值
  return textMap[type] || textMap[type?.toLowerCase()] || type;
};

// ✨ 获取款项类型样式
const getPaymentTypeStyle = (type: string) => {
  const styleMap: Record<string, any> = {
    total: getDictColorRenderConfig('primary').style || {},
    first: getDictColorRenderConfig('success').style || {},
    FIRST_PAYMENT: getDictColorRenderConfig('success').style || {},
    deposit: getDictColorRenderConfig('success').style || {},
    final: getDictColorRenderConfig('warning').style || {},
    FINAL_PAYMENT: getDictColorRenderConfig('warning').style || {},
    installment: getDictColorRenderConfig('default').style || {},
    production: getDictColorRenderConfig('brand').style || {},
  };
  return styleMap[type] || styleMap[type?.toLowerCase()] || {};
};

// 筛选栏配置
const filterConfig: FilterItem[] = [
  {
    type: 'select',
    field: 'supplier_id',
    placeholder: '请选择供应商',
    options: supplierList,
    labelKey: 'sup_name',
    valueKey: 'id',
  },
  {
    type: 'select',
    field: 'status',
    placeholder: '请选择状态',
    options: payableStatusOptions,
    labelKey: 'dict_label',
    valueKey: 'dict_value',
  },
];

onMounted(() => {
  loadData();
  loadSupplierList();
  loadColumnConfig();
});
</script>
<style scoped>
.finance-payable-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
}

.finance-payable-container :deep(.search-card) {
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
  justify-content: space-between;
  align-items: center;
}

.toolbar-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.toolbar-actions {
  display: flex;
  gap: 8px;
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

.file-name-link {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
</style>

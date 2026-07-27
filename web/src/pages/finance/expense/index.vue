<template>
  <div class="finance-expense-container">
    <SearchFilter
      :filters="filterConfig"
      :model-value="searchParams"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">费用管理</span>
        <div class="toolbar-actions">
          <t-button theme="primary" @click="handleAdd">
            <template #icon><add-icon /></template>
            新增费用
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
          <div class="stat-label">总费用金额</div>
          <div class="stat-value">¥{{ formatAmount(summaryStats.totalAmount) }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">已付金额</div>
          <div class="stat-value success">¥{{ formatAmount(summaryStats.totalPaid) }}</div>
        </div>
        <div class="stat-item">
          <div class="stat-label">未付金额</div>
          <div class="stat-value warning">¥{{ formatAmount(summaryStats.totalBalance) }}</div>
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
        <template #expense_type="{ row }">
          <t-tag :style="getExpenseTypeStyle(row.expense_type)">
            {{ getExpenseTypeText(row.expense_type) }}
          </t-tag>
        </template>

        <template #amount="{ row }">
          <span class="amount-text">¥{{ formatAmount(row.amount) }}</span>
        </template>

        <template #paid_amount="{ row }">
          <span class="amount-text success">¥{{ formatAmount(row.paid_amount) }}</span>
        </template>

        <template #balance="{ row }">
          <span class="amount-text" :class="{ warning: row.balance > 0 }">¥{{ formatAmount(row.balance) }}</span>
        </template>

        <template #status="{ row }">
          <t-tag :style="getStatusStyle(row.status)" :theme="getStatusTheme(row.status)">
            {{ getStatusText(row.status) }}
          </t-tag>
        </template>

        <template #payment_date="{ row }">
          {{ row.payment_date ? formatDate(row.payment_date) : '-' }}
        </template>

        <template #create_time="{ row }">
          {{ row.create_time ? formatDate(row.create_time) : '-' }}
        </template>

        <template #op="{ row }">
          <t-space>
            <t-link v-if="row.balance > 0" theme="primary" @click="handlePay(row)">支付</t-link>
            <t-link v-if="row.paid_amount > 0" theme="primary" @click="handleViewPayments(row)">支付记录</t-link>
            <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
            <t-popconfirm content="确认删除该费用记录？" @confirm="handleDelete(row)">
              <t-link theme="danger">删除</t-link>
            </t-popconfirm>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 新增/编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :header="isEdit ? '编辑费用' : '新增费用'"
      width="600px"
      @confirm="handleSubmit"
      @cancel="handleCancel"
     :close-on-overlay-click="false">
      <t-form ref="formRef" :data="formData" :rules="rules" label-width="100px">
        <t-form-item label="费用类型" name="expense_type">
          <t-select v-model="formData.expense_type" placeholder="请选择费用类型" clearable>
            <t-option
              v-for="item in expenseTypeOptions"
              :key="item.dict_value"
              :value="item.dict_value"
              :label="item.dict_label"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="费用金额" name="amount">
          <t-input-number
            v-model="formData.amount"
            :min="0"
            :precision="2"
            placeholder="请输入费用金额"
            style="width: 100%"
          >
            <template #suffix>元</template>
          </t-input-number>
        </t-form-item>
        <t-form-item label="支付方">
          <t-input v-model="formData.payer" placeholder="请输入支付方" />
        </t-form-item>
        <t-form-item label="收款方">
          <t-input v-model="formData.receiver" placeholder="请输入收款方" />
        </t-form-item>
        <t-form-item label="支付日期">
          <t-date-picker v-model="formData.payment_date" placeholder="选择支付日期" style="width: 100%" />
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="formData.remark" placeholder="请输入备注" :maxlength="500" show-limit-number />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 支付对话框 -->
    <t-dialog
      v-model:visible="payDialogVisible"
      header="支付费用"
      width="600px"
      @confirm="handlePaySubmit"
      @cancel="payDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form :data="payForm" label-width="120px">
        <t-form-item label="费用金额">
          <span class="amount-text">¥{{ formatAmount(currentExpense?.amount || 0) }}</span>
        </t-form-item>
        <t-form-item label="已付金额">
          <span class="amount-text success">¥{{ formatAmount(currentExpense?.paid_amount || 0) }}</span>
        </t-form-item>
        <t-form-item label="余额">
          <span class="amount-text warning">¥{{ formatAmount(balanceAmount) }}</span>
        </t-form-item>
        <t-form-item label="支付金额" required>
          <t-input-number
            v-model="payForm.amount"
            :min="0.01"
            :max="balanceAmount"
            :precision="2"
            placeholder="请输入支付金额"
            style="width: 100%"
          >
            <template #suffix>元</template>
          </t-input-number>
        </t-form-item>
        <t-form-item label="支付日期">
          <t-date-picker
            v-model="payForm.payment_date"
            format="YYYY-MM-DD"
            placeholder="选择支付日期"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item label="支付方式">
          <t-select v-model="payForm.payment_method" placeholder="请选择支付方式" clearable>
            <t-option
              v-for="item in paymentMethodOptions"
              :key="item.dict_value"
              :value="item.dict_value"
              :label="item.dict_label"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="回单号/流水号">
          <t-input v-model="payForm.transaction_no" placeholder="请输入银行回单号或交易流水号" />
        </t-form-item>
        <t-form-item label="付款账户">
          <t-select v-model="payForm.payment_account" placeholder="请选择付款账户" clearable filterable>
            <t-option
              v-for="item in receivingAccountOptions"
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
          <t-textarea v-model="payForm.remark" placeholder="请输入备注信息" :autosize="{ minRows: 2, maxRows: 4 }" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 支付记录对话框 -->
    <t-dialog
      v-model:visible="paymentDialogVisible"
      header="支付记录"
      width="700px"
      :footer="false"
      @cancel="paymentDialogVisible = false"
     :close-on-overlay-click="false">
      <t-table
        :data="paymentRecords"
        :columns="paymentColumns"
        row-key="id"
        hover
        stripe
        :loading="paymentLoading"
      >
        <template #payment_amount="{ row }">
          <span class="amount-text">¥{{ formatAmount(row.payment_amount) }}</span>
        </template>
        <template #payment_date="{ row }">
          {{ row.payment_date ? formatDate(row.payment_date) : '-' }}
        </template>
        <template #payment_method="{ row }">
          <span v-if="row.payment_method">{{ getPaymentMethodText(row.payment_method) }}</span>
          <span v-else>-</span>
        </template>
      </t-table>
    </t-dialog>

    <!-- 列设置对话框 -->
    <ColumnSettingDialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="finance_expense_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { RefreshIcon, AddIcon, SettingIcon } from 'tdesign-icons-vue-next';
import SearchFilter, { type FilterItem } from '@/components/SearchFilter/index.vue';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import {
  getExpenseList,
  getExpenseStats,
  createExpense,
  updateExpense,
  deleteExpense,
  payExpense,
  getExpensePayments,
  type ExpenseRecord,
} from '@/api/finance/expense';
import { useDictRaw } from '@/hooks/useDict';
import { getDictColorRenderConfig } from '@/utils/dict-color';
import dayjs from 'dayjs';
import ColumnSettingDialog, { type ColumnOption } from '@/components/ColumnSettingDialog/index.vue';

defineOptions({
  name: 'FinanceExpenseIndex',
});

// 搜索参数
const searchParams = reactive({
  expense_type: '',
  date_range: [] as string[],
});

// 表格
const loading = ref(false);
const tableData = ref<ExpenseRecord[]>([]);
const pagination = reactive({ current: 1, pageSize: 20, total: 0 });

// 汇总统计
const summaryStats = reactive({
  totalAmount: 0,
  totalPaid: 0,
  totalBalance: 0,
});

// 对话框
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref();
const formData = ref<Partial<ExpenseRecord>>({
  expense_type: '',
  amount: 0,
  payer: '',
  receiver: '',
  payment_date: '',
  remark: '',
});

// 支付对话框
const payDialogVisible = ref(false);
const currentExpense = ref<ExpenseRecord | null>(null);
const payForm = reactive({
  amount: 0,
  payment_date: '',
  payment_method: '',
  transaction_no: '',
  payment_account: '',
  receiver_account: '',
  remark: '',
});

// 余额
const balanceAmount = computed(() => {
  if (!currentExpense.value) return 0;
  return currentExpense.value.amount - currentExpense.value.paid_amount;
});

// 支付记录对话框
const paymentDialogVisible = ref(false);
const paymentRecords = ref<any[]>([]);
const paymentLoading = ref(false);

// 费用类型字典（useDictRaw 自动加载 + 全局缓存）
const { items: expenseTypeOptions } = useDictRaw('expense-type');
const expenseTypeMap = computed<Record<string, string>>(() => {
  const map: Record<string, string> = {};
  expenseTypeOptions.value.forEach((item) => {
    map[item.dict_value] = (item as any).color || '#0052d9';
  });
  return map;
});

// 支付方式字典（useDictRaw 自动加载 + 全局缓存）
const { items: paymentMethodOptions } = useDictRaw('payment-method');

// 付款账户字典（useDictRaw 自动加载 + 全局缓存）
const { items: receivingAccountOptions } = useDictRaw('receiving-account');

// 列设置
const columnSettingRef = ref();
const selectedColumns = ref<string[]>([]);

// 所有可选列配置
const allColumns: ColumnOption[] = [
  { colKey: 'id', title: 'ID' },
  { colKey: 'expense_type', title: '费用类型' },
  { colKey: 'amount', title: '费用金额' },
  { colKey: 'paid_amount', title: '已付金额' },
  { colKey: 'balance', title: '余额' },
  { colKey: 'status', title: '状态' },
  { colKey: 'payer', title: '支付方' },
  { colKey: 'receiver', title: '收款方' },
  { colKey: 'payment_date', title: '支付日期' },
  { colKey: 'remark', title: '备注' },
  { colKey: 'create_time', title: '创建时间' },
];

// 默认显示的列
const DEFAULT_COLUMNS = ['expense_type', 'amount', 'paid_amount', 'balance', 'status', 'payer', 'payment_date'];

// 基础列配置
const baseColumns: PrimaryTableCol[] = [
  { colKey: 'id', title: 'ID', width: 80 },
  { colKey: 'expense_type', title: '费用类型', width: 120 },
  { colKey: 'amount', title: '费用金额', width: 120 },
  { colKey: 'paid_amount', title: '已付金额', width: 120 },
  { colKey: 'balance', title: '余额', width: 120 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'payer', title: '支付方', width: 120 },
  { colKey: 'receiver', title: '收款方', width: 120 },
  { colKey: 'payment_date', title: '支付日期', width: 120 },
  { colKey: 'remark', title: '备注', minWidth: 200 },
  { colKey: 'create_time', title: '创建时间', width: 180 },
  { colKey: 'op', title: '操作', width: 220, fixed: 'right' },
];

// 支付记录表格列
const paymentColumns: PrimaryTableCol[] = [
  { colKey: 'id', title: 'ID', width: 60 },
  { colKey: 'payment_amount', title: '支付金额', width: 120 },
  { colKey: 'payment_date', title: '支付日期', width: 120 },
  { colKey: 'payment_method', title: '支付方式', width: 100 },
  { colKey: 'transaction_no', title: '流水号/回单号', width: 160 },
  { colKey: 'remark', title: '备注', minWidth: 160 },
];

const displayColumns = computed(() => {
  const orderedColumns = selectedColumns.value
    .map(colKey => baseColumns.find(col => col.colKey === colKey))
    .filter(col => col !== undefined);

  // 确保操作列始终在最后且固定右侧
  const operationCol = baseColumns.find(col => col.colKey === 'op');
  if (operationCol) {
    orderedColumns.push(operationCol);
  }

  return orderedColumns;
});

const loadColumnConfig = () => {
  try {
    const saved = localStorage.getItem('finance_expense_table_columns');
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
    localStorage.setItem('finance_expense_table_columns', JSON.stringify(selectedColumns.value));
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

// 表单验证
const rules = {
  expense_type: [{ required: true, message: '请选择费用类型', type: 'error' as const }],
  amount: [{ required: true, message: '请输入费用金额', type: 'error' as const }],
};

// 格式化金额
function formatAmount(amount: number) {
  return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

// 格式化日期
function formatDate(date: string) {
  return date ? dayjs(date).format('YYYY-MM-DD') : '-';
}

// 费用类型
function getExpenseTypeStyle(type: string) {
  const color = expenseTypeMap.value[type];
  if (color) {
    // 先尝试用字典颜色解析
    const config = getDictColorRenderConfig(color);
    if (config.style) return config.style;
    // 如果颜色解析无结果，直接用颜色值作为降级
    return { color: color, borderColor: color, background: color + '20' };
  }
  // 默认颜色映射
  const colorMap: Record<string, any> = {
    freight: getDictColorRenderConfig('warning').style || {},
    packaging: getDictColorRenderConfig('success').style || {},
    insurance: getDictColorRenderConfig('primary').style || {},
    other: getDictColorRenderConfig('default').style || {},
  };
  return colorMap[type] || {};
}

function getExpenseTypeText(type: string) {
  const opt = expenseTypeOptions.value.find((o: any) => o.dict_value === type);
  return opt?.dict_label || type;
}

// 状态标签
function getStatusStyle(status: string) {
  const styleMap: Record<string, any> = {
    pending: getDictColorRenderConfig('default').style || {},
    partial: getDictColorRenderConfig('warning').style || {},
    paid: getDictColorRenderConfig('success').style || {},
  };
  return styleMap[status] || {};
}

function getStatusTheme(status: string): 'default' | 'primary' | 'success' | 'warning' | 'danger' {
  const themeMap: Record<string, 'default' | 'primary' | 'success' | 'warning' | 'danger'> = {
    pending: 'default',
    partial: 'warning',
    paid: 'success',
  };
  return themeMap[status] || 'default';
}

function getStatusText(status: string) {
  const textMap: Record<string, string> = {
    pending: '待支付',
    partial: '部分支付',
    paid: '已支付',
  };
  return textMap[status] || status;
}

// 支付方式文本
function getPaymentMethodText(method: string) {
  const opt = paymentMethodOptions.value.find((o: any) => o.dict_value === method);
  return opt?.dict_label || method;
}

// 加载字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 搜索

const loadStats = async () => {
  try {
    const res = await getExpenseStats();
    if (res) {
      summaryStats.totalAmount = (res as any)?.total_amount || 0;
      summaryStats.totalPaid = (res as any)?.total_paid || 0;
      summaryStats.totalBalance = (res as any)?.total_balance || 0;
    }
  } catch (error) {
    console.error('加载汇总统计失败:', error);
  }
};

const loadData = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.current,
      page_size: pagination.pageSize,
    };
    if (searchParams.expense_type) {
      params.expense_type = searchParams.expense_type;
    }
    if (searchParams.date_range?.length === 2) {
      params.start_date = searchParams.date_range[0];
      params.end_date = searchParams.date_range[1];
    }
    const res = await getExpenseList(params);
    const list = Array.isArray((res as any)?.list) ? (res as any).list : [];
    tableData.value = list.map((item: any) => ({
      ...item,
      payment_date: item.payment_date || '',
    }));
    pagination.total = typeof (res as any)?.total === 'number' ? (res as any).total : 0;
  } catch (error) {
    console.error('加载数据失败:', error);
    MessagePlugin.error('加载数据失败');
  } finally {
    loading.value = false;
  }
};

// 搜索
const handleSearch = () => { pagination.current = 1; loadData(); };
const handleReset = () => { searchParams.expense_type = ''; searchParams.date_range = []; handleSearch(); };
const handleRefresh = () => { loadData(); loadStats(); };
const onPageChange = (pageInfo: any) => { pagination.current = pageInfo.current; pagination.pageSize = pageInfo.pageSize; loadData(); };

// 新增
const handleAdd = () => {
  isEdit.value = false;
  formData.value = { expense_type: '', amount: 0, payer: '', receiver: '', payment_date: '', remark: '' };
  dialogVisible.value = true;
};

// 编辑
const handleEdit = (row: ExpenseRecord) => {
  isEdit.value = true;
  formData.value = { ...row };
  dialogVisible.value = true;
};

// 提交
const handleSubmit = async () => {
  const validateResult = await formRef.value?.validate();
  if (!validateResult) return;
  try {
    if (isEdit.value && formData.value.id) {
      await updateExpense(formData.value.id, formData.value);
      MessagePlugin.success('更新成功');
    } else {
      await createExpense(formData.value);
      MessagePlugin.success('创建成功');
    }
    dialogVisible.value = false;
    loadData();
    loadStats();
  } catch (error) {
    MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
  }
};

const handleCancel = () => { dialogVisible.value = false; };

const handleDelete = async (row: ExpenseRecord) => {
  try {
    await deleteExpense(row.id);
    MessagePlugin.success('删除成功');
    loadData();
    loadStats();
  } catch (error) {
    MessagePlugin.error('删除失败');
  }
};

// 支付
const handlePay = (row: ExpenseRecord) => {
  currentExpense.value = row;
  payForm.amount = 0;
  payForm.payment_date = '';
  payForm.payment_method = '';
  payForm.transaction_no = '';
  payForm.payment_account = '';
  payForm.receiver_account = '';
  payForm.remark = '';
  payDialogVisible.value = true;
};

const handlePaySubmit = async () => {
  if (!currentExpense.value || payForm.amount <= 0) {
    MessagePlugin.warning('请输入有效的支付金额');
    return;
  }
  if (payForm.amount > balanceAmount.value) {
    MessagePlugin.warning(`支付金额不能超过余额 ¥${formatAmount(balanceAmount.value)}`);
    return;
  }
  try {
    await payExpense(currentExpense.value.id, {
      amount: payForm.amount,
      payment_date: payForm.payment_date || dayjs().format('YYYY-MM-DD'),
      payment_method: payForm.payment_method,
      transaction_no: payForm.transaction_no,
      payment_account: payForm.payment_account,
      receiver_account: payForm.receiver_account,
      remark: payForm.remark,
    });
    MessagePlugin.success('支付成功');
    payDialogVisible.value = false;
    loadData();
    loadStats();
  } catch (error) {
    MessagePlugin.error('支付失败');
  }
};

// 查看支付记录
const handleViewPayments = async (row: ExpenseRecord) => {
  paymentDialogVisible.value = true;
  paymentLoading.value = true;
  try {
    const res = await getExpensePayments(row.id);
    paymentRecords.value = Array.isArray(res) ? res : (Array.isArray((res as any)?.list) ? (res as any).list : []);
  } catch (error) {
    console.error('加载支付记录失败:', error);
    MessagePlugin.error('加载支付记录失败');
    paymentRecords.value = [];
  } finally {
    paymentLoading.value = false;
  }
};

// 筛选栏配置
const filterConfig = computed<FilterItem[]>(() => [
  {
    type: 'select',
    field: 'expense_type',
    placeholder: '请选择费用类型',
    options: expenseTypeOptions.value,
    labelKey: 'dict_label',
    valueKey: 'dict_value',
  },
  { type: 'dateRangePicker', field: 'date_range', placeholder: '日期范围', span: 8 },
]);

onMounted(() => {
  loadData();
  loadStats();
  loadColumnConfig();
});
</script>

<style scoped>
.finance-expense-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 16px;
}

.finance-expense-container :deep(.search-card) {
  margin-bottom: 0;
}

.toolbar-card {
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
}

.toolbar-actions {
  display: flex;
  gap: 8px;
}

.amount-text {
  font-weight: 500;
}

.amount-text.success {
  color: #00a870;
}

.amount-text.warning {
  color: #ed7b2f;
}

/* 汇总统计 */
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
  color: rgba(0, 0, 0, 0.6);
  margin-bottom: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: rgba(0, 0, 0, 0.9);
}

.stat-value.success {
  color: #00a870;
}

.stat-value.warning {
  color: #ed7b2f;
}
</style>

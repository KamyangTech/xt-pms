<template>
  <div class="payment-list">
    <div class="toolbar">
      <span class="title">款项明细</span>
      <t-button theme="primary" @click="handleAdd">
        <template #icon><add-icon /></template>
        新增款项
      </t-button>
    </div>

    <t-table
      :data="payments"
      :columns="columns"
      :loading="loading"
      row-key="id"
      hover
      stripe
    >
      <template #payment_type="{ row }">
        {{ getPaymentTypeLabel(row.payment_type) }}
      </template>

      <template #source_type="{ row }">
        <t-tag v-if="row.source_type" :theme="row.source_type === 'mold' ? 'primary' : 'warning'" size="small">
          {{ row.source_type === 'mold' ? '模具款' : row.source_type === 'outsource' ? '外发应付款' : row.source_type }}
        </t-tag>
        <span v-else>-</span>
      </template>

      <template #contract_amount="{ row }">
        ¥{{ formatAmount(row.contract_amount) }}
      </template>

      <template #actual_amount="{ row }">
        <span :class="{ 'text-success': row.actual_amount > 0 }">
          ¥{{ formatAmount(row.actual_amount) }}
        </span>
      </template>

      <template #balance="{ row }">
        <span :class="{ 'text-warning': getBalance(row) > 0, 'text-success': getBalance(row) <= 0 }">
          ¥{{ formatAmount(getBalance(row)) }}
        </span>
      </template>

      <template #payment_date="{ row }">
        {{ row.payment_date ? formatDate(row.payment_date) : '-' }}
      </template>

      <template #due_date="{ row }">
        {{ row.due_date ? formatDate(row.due_date) : '-' }}
      </template>

      <template #status="{ row }">
        <t-tag :theme="getStatusTheme(row.status)">
          {{ getStatusText(row.status) }}
        </t-tag>
      </template>

      <template #op="{ row }">
        <t-space>
          <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
          <t-popconfirm content="确认删除该款项记录？" @confirm="handleDelete(row)">
            <t-link theme="danger">删除</t-link>
          </t-popconfirm>
        </t-space>
      </template>
    </t-table>

    <!-- 款项表单对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :header="isEdit ? '编辑款项' : '新增款项'"
      width="640px"
      @confirm="handleSubmit"
      @cancel="handleCancel"
     :close-on-overlay-click="false">
      <t-form ref="formRef" :data="formData" :rules="rules" label-width="120px">
        <t-form-item label="款项归属" name="source_type">
          <t-select v-model="formData.source_type" placeholder="请选择款项归属" clearable>
            <t-option value="mold" label="模具款" />
            <t-option value="outsource" label="外发应付款" />
          </t-select>
        </t-form-item>

        <t-form-item label="款项类型" name="payment_type">
          <t-select v-model="formData.payment_type" placeholder="请选择款项类型" clearable>
            <t-option
              v-for="item in paymentTypes"
              :key="item.dict_value || item.item_code"
              :value="item.dict_value || item.item_code"
              :label="item.dict_label || item.item_name"
            />
          </t-select>
        </t-form-item>

        <t-form-item label="合同约定金额" name="contract_amount">
          <t-input-number
            v-model="formData.contract_amount"
            :min="0"
            :precision="2"
            placeholder="请输入合同约定金额"
            style="width: 100%"
          >
            <template #suffix>元</template>
          </t-input-number>
        </t-form-item>

        <t-form-item label="实际到账金额" name="actual_amount">
          <t-input-number
            v-model="formData.actual_amount"
            :min="0"
            :precision="2"
            placeholder="请输入实际到账金额"
            style="width: 100%"
            @change="handleActualAmountChange"
          >
            <template #suffix>元</template>
          </t-input-number>
        </t-form-item>

        <!-- 自动计算的应收余额 -->
        <t-form-item label="应收余额">
          <div class="auto-calc-field">
            <span class="calc-value" :class="{ 'text-success': calcBalance <= 0, 'text-warning': calcBalance > 0 }">
              ¥{{ formatAmount(calcBalance) }}
            </span>
            <span class="calc-hint">自动计算（合同金额 - 实际到账）</span>
          </div>
        </t-form-item>

        <t-form-item label="到期日期" name="due_date">
          <t-date-picker
            v-model="formData.due_date"
            placeholder="请选择到期日期"
            format="YYYY-MM-DD"
            clearable
            style="width: 100%"
          />
        </t-form-item>

        <t-form-item label="收款日期" name="payment_date">
          <t-date-picker
            v-model="formData.payment_date"
            placeholder="请选择收款日期"
            clearable
            style="width: 100%"
          />
        </t-form-item>

        <t-form-item label="状态" name="status">
          <t-select v-model="formData.status" placeholder="请选择状态">
            <t-option :value="1" label="未收" />
            <t-option :value="2" label="部分收款" />
            <t-option :value="3" label="已收款" />
          </t-select>
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
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { AddIcon } from 'tdesign-icons-vue-next';
import type { PaymentRecord } from '@/api/mold/payment';
import { getOrderPayments, getPaymentsByWorkOrderId, createPayment, updatePayment, deletePayment } from '@/api/mold/payment';
import { useDictRaw } from '@/hooks/useDict';
import { formatDate as formatDateUtil } from '@/utils/date';

interface Props {
  orderId?: number;
  workOrderId?: number;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: 'refresh'): void;
}>();

const loading = ref(false);
const payments = ref<PaymentRecord[]>([]);
const dialogVisible = ref(false);
const isEdit = ref(false);
const formRef = ref();
const formData = ref<Partial<PaymentRecord>>({
  payment_type: '',
  source_type: 'mold',
  contract_amount: 0,
  actual_amount: 0,
  should_amount: 0,
  payment_date: '',
  due_date: '',
  status: 1,
  remark: '',
});

// 款项类型字典（useDictRaw 自动加载 + 全局缓存）
const { items: paymentTypes } = useDictRaw('payment-type');

// 自动计算的应收余额
const calcBalance = computed(() => {
  const contract = formData.value.contract_amount || 0;
  const actual = formData.value.actual_amount || 0;
  return Math.max(contract - actual, 0);
});

// 表单验证规则
const rules = {
  payment_type: [{ required: true, message: '请选择款项类型', type: 'error' as const }],
  contract_amount: [{ required: true, message: '请输入合同约定金额', type: 'error' as const }],
};

// 表格列配置（使用 computed 启用响应式）
const columns = computed(() => [
  { colKey: 'source_type', title: '款项归属', width: 110 },
  { colKey: 'payment_type', title: '款项类型', width: 110 },
  { colKey: 'contract_amount', title: '合同约定金额', width: 140 },
  { colKey: 'actual_amount', title: '实际到账金额', width: 140 },
  { colKey: 'balance', title: '应收余额', width: 120 },
  { colKey: 'payment_date', title: '收款日期', width: 110 },
  { colKey: 'due_date', title: '到期日期', width: 110 },
  { colKey: 'status', title: '状态', width: 90 },
  { colKey: 'remark', title: '备注', ellipsis: true, minWidth: 120 },
  { colKey: 'op', title: '操作', width: 130, fixed: 'right' as const },
]);

function getBalance(row: PaymentRecord): number {
  return Math.max((row.contract_amount || 0) - (row.actual_amount || 0), 0);
}

async function loadPayments() {
  if (!props.orderId && !props.workOrderId) return;

  loading.value = true;
  try {
    let res;
    // 优先使用工单ID查询，否则使用订单ID
    if (props.workOrderId) {
      res = await getPaymentsByWorkOrderId(props.workOrderId);
    } else if (props.orderId) {
      res = await getOrderPayments(props.orderId);
    }
    
    if (res.code === 200 && res.data) {
      payments.value = res.data;
    }
  } catch (error) {
    MessagePlugin.error('加载款项列表失败');
  } finally {
    loading.value = false;
  }
}

// 格式化金额
function formatAmount(amount: number) {
  return amount.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

// 格式化日期
function formatDate(date: string) {
  if (!date) return '-';
  return formatDateUtil(date);
}

function getPaymentTypeLabel(type: string) {
  if (!type) return '-';
  const matched = paymentTypes.value.find(
    (item: any) => item.dict_value === type || item.item_code === type
  );
  return matched ? (matched.dict_label || matched.item_name) : type;
}

function getStatusText(status: number) {
  const statusMap: Record<number, string> = {
    1: '未收',
    2: '部分收款',
    3: '已收款',
  };
  return statusMap[status] || '未知';
}

function getStatusTheme(status: number) {
  const themeMap: Record<number, any> = {
    1: 'default',
    2: 'warning',
    3: 'success',
  };
  return themeMap[status] || 'default';
}

// 新增款项
function handleAdd() {
  isEdit.value = false;
  formData.value = {
    payment_type: '',
    source_type: 'mold',
    contract_amount: 0,
    actual_amount: 0,
    should_amount: 0,
    payment_date: '',
    due_date: '',
    status: 1,
    remark: '',
  };
  dialogVisible.value = true;
}

// 编辑款项
function handleEdit(record: PaymentRecord) {
  isEdit.value = true;
  formData.value = { ...record };
  dialogVisible.value = true;
}

async function handleDelete(record: PaymentRecord) {
  try {
    await deletePayment(record.id);
    MessagePlugin.success('删除成功');
    await loadPayments();
    emit('refresh');
  } catch (error) {
    MessagePlugin.error('删除失败');
  }
}

// 实际到账金额变化时自动计算 should_amount
function handleActualAmountChange(val: number) {
  const contract = formData.value.contract_amount || 0;
  formData.value.should_amount = Math.max(contract - (val || 0), 0);
}

// 提交表单
async function handleSubmit() {
  const valid = await formRef.value?.validate();
  if (valid !== true) return;

  try {
    // 自动计算 should_amount
    const contract = formData.value.contract_amount || 0;
    const actual = formData.value.actual_amount || 0;
    formData.value.should_amount = Math.max(contract - actual, 0);

    if (isEdit.value && formData.value.id) {
      await updatePayment(formData.value.id, formData.value);
      MessagePlugin.success('更新成功');
    } else {
      if (!props.orderId) {
        MessagePlugin.warning('订单ID不存在');
        return;
      }
      await createPayment({ ...formData.value, order_id: props.orderId });
      MessagePlugin.success('创建成功');
    }
    dialogVisible.value = false;
    await loadPayments();
    emit('refresh');
  } catch (error) {
    MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
  }
}

// 取消
function handleCancel() {
  dialogVisible.value = false;
  formRef.value?.reset();
}

// 加载款项类型字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 暴露方法供父组件调用
defineExpose({
  loadPayments,
});

onMounted(() => {
  loadPayments();
  // 款项类型字典已通过 useDictRaw 自动加载
});
</script>

<style scoped>
.payment-list {
  width: 100%;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.text-success {
  color: var(--td-success-color, #00a870);
  font-weight: 500;
}

.text-warning {
  color: var(--td-warning-color, #ed7b2f);
  font-weight: 500;
}

.auto-calc-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px 0;
}

.calc-value {
  font-size: 18px;
  font-weight: 600;
}

.calc-hint {
  font-size: 12px;
  color: var(--td-text-color-secondary, rgba(0, 0, 0, 0.6));
}
</style>

<template>
  <div class="payment-plan-config">
    <div class="config-header">
      <span class="title">款项配置</span>
      <t-switch v-model="enabled" :value="enabled" @change="(val: any) => handleToggle(val as boolean)">
        <template #label="slotProps">
          {{ slotProps.value ? '自定义' : '默认' }}
        </template>
      </t-switch>
    </div>

    <t-alert
      v-if="!enabled"
      theme="info"
      :message="alertMessage"
      style="margin-top: 12px"
    />

    <div v-show="enabled" class="payment-items-container">
      <t-table
        :data="paymentItems"
        :columns="columns"
        row-key="index"
        bordered
        hover
        size="small"
      >
        <template #payment_type="{ row, rowIndex }">
          <t-select
            v-model="row.payment_type"
            placeholder="选择款项类型"
            @change="handlePaymentTypeChange(rowIndex)"
          >
            <t-option
              v-for="item in getAvailablePaymentTypes(rowIndex)"
              :key="item.dict_value"
              :value="item.dict_value"
              :label="item.dict_label"
            />
          </t-select>
        </template>

        <template #source_type="{ row, rowIndex }">
          <t-select
            v-model="row.source_type"
            placeholder="款项归属"
            style="width: 130px"
          >
            <t-option value="mold" label="模具款" />
            <t-option value="outsource" label="外发应付款" />
          </t-select>
        </template>

        <template #amount_type="{ row, rowIndex }">
          <t-radio-group v-model="row.amount_type" size="small">
            <t-radio-button value="amount">金额</t-radio-button>
            <t-radio-button value="ratio">比例</t-radio-button>
          </t-radio-group>
        </template>

        <template #amount_value="{ row, rowIndex }">
          <t-input-number
            v-model="row.amount_value"
            :min="0"
            :max="row.amount_type === 'ratio' ? 100 : totalAmount"
            :precision="2"
            :placeholder="row.amount_type === 'ratio' ? '请输入比例' : '请输入金额'"
            theme="normal"
            style="width: 100%"
            @change="calculateAmounts"
          >
            <template #suffix>{{ row.amount_type === 'ratio' ? '%' : '元' }}</template>
          </t-input-number>
        </template>

        <template #calculated_amount="{ row }">
          <span class="amount-text">¥{{ formatAmount(row.calculated_amount || 0) }}</span>
        </template>

        <template #due_date="{ row, rowIndex }">
          <t-date-picker
            v-model="row.due_date"
            placeholder="选择到期日"
            format="YYYY-MM-DD"
            style="width: 100%"
          />
        </template>

        <template #op="{ rowIndex }">
          <t-link theme="danger" @click="removeItem(rowIndex)">删除</t-link>
        </template>
      </t-table>

      <div class="add-button-container">
        <t-button theme="default" variant="dashed" block @click="addItem">
          <template #icon><add-icon /></template>
          添加款项
        </t-button>
      </div>

      <div class="summary-info">
        <t-space>
          <span>订单总额：<strong>¥{{ formatAmount(totalAmount) }}</strong></span>
          <span>已配置：<strong :class="{ 'text-success': isAmountMatched, 'text-warning': !isAmountMatched }">
            ¥{{ formatAmount(configuredTotal) }}
          </strong></span>
          <span v-if="!isAmountMatched" class="text-warning">
            （差额：¥{{ formatAmount(Math.abs(totalAmount - configuredTotal)) }}）
          </span>
        </t-space>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { AddIcon } from 'tdesign-icons-vue-next';
import { useDictRaw } from '@/hooks/useDict';

interface PaymentItem {
  index: number;
  payment_type: string;
  source_type: string;    // 款项归属：mold=模具款, outsource=外发应付款
  amount_type: 'amount' | 'ratio';
  amount_value: number;
  calculated_amount: number;
  due_date: string;
}

const props = defineProps<{
  totalAmount: number;
  defaultDepositRatio?: number; // 默认首付款比例（仅综合订单使用）
  defaultItems?: PaymentItem[]; // 默认款项列表
}>();

const emit = defineEmits<{
  (e: 'change', data: { enabled: boolean; items: PaymentItem[] }): void;
}>();

const enabled = ref(false);
const paymentItems = ref<PaymentItem[]>([]);
// 款项类型字典（useDictRaw 自动加载 + 全局缓存）
const { items: paymentTypeOptions } = useDictRaw('payment-type');

// 表格列配置
const columns = [
  { colKey: 'source_type', title: '款项归属', width: 150 },
  { colKey: 'payment_type', title: '款项类型', width: 150 },
  { colKey: 'amount_type', title: '金额类型', width: 140 },
  { colKey: 'amount_value', title: '金额/比例', width: 160 },
  { colKey: 'calculated_amount', title: '计算金额', width: 120 },
  { colKey: 'due_date', title: '到期日期', width: 150 },
  { colKey: 'op', title: '操作', width: 80, fixed: 'right' as const },
];

// 提示信息
const alertMessage = computed(() => {
  if (props.defaultDepositRatio) {
    return `系统将按首付款 ${props.defaultDepositRatio}% 自动拆分为首付款和尾款`;
  }
  return '系统将自动生成单笔应收账款';
});

// 已配置总额
const configuredTotal = computed(() => {
  return paymentItems.value.reduce((sum, item) => sum + (item.calculated_amount || 0), 0);
});

// 金额是否匹配
const isAmountMatched = computed(() => {
  return Math.abs(configuredTotal.value - props.totalAmount) < 0.01;
});

// 加载款项类型字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 切换开关
const handleToggle = (value: boolean) => {
  if (value) {
    // 启用时，初始化默认款项
    initDefaultItems();
  } else {
    // 禁用时，清空款项
    paymentItems.value = [];
  }
  emitChange();
};

// 初始化默认款项
const initDefaultItems = () => {
  // 如果有传入的默认款项，使用它
  if (props.defaultItems && props.defaultItems.length > 0) {
    paymentItems.value = props.defaultItems.map((item, index) => ({
      ...item,
      index: index + 1,
      source_type: item.source_type || 'mold',
    }));
    return;
  }

  const now = new Date();
  const dueDate30 = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
  const dueDate60 = new Date(now.getTime() + 60 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

  if (props.defaultDepositRatio) {
    // 综合订单：首款 + 尾款
    const depositAmount = props.totalAmount * props.defaultDepositRatio / 100;
    const balanceAmount = props.totalAmount - depositAmount;
    
    paymentItems.value = [
      {
        index: 1,
        payment_type: 'FIRST_PAYMENT',
        source_type: 'mold',
        amount_type: 'amount',
        amount_value: depositAmount,
        calculated_amount: depositAmount,
        due_date: dueDate30,
      },
      {
        index: 2,
        payment_type: 'FINAL_PAYMENT',
        source_type: 'mold',
        amount_type: 'amount',
        amount_value: balanceAmount,
        calculated_amount: balanceAmount,
        due_date: dueDate60,
      },
    ];
  } else {
    // 模具订单：单笔（模具款）
    paymentItems.value = [
      {
        index: 1,
        payment_type: 'TOTAL',
        source_type: 'mold',
        amount_type: 'amount',
        amount_value: props.totalAmount,
        calculated_amount: props.totalAmount,
        due_date: dueDate30,
      },
    ];
  }
};

const addItem = () => {
  // 如果已有全款，不允许添加
  if (paymentItems.value.some(p => p.payment_type?.toUpperCase() === 'TOTAL')) {
    MessagePlugin.warning('全款已配置，无法添加更多款项');
    return;
  }
  // 如果已有尾款，不允许添加（尾款必须最后）
  if (paymentItems.value.some(p => p.payment_type?.toUpperCase() === 'FINAL_PAYMENT')) {
    MessagePlugin.warning('尾款必须是最后一笔款项，无法再添加');
    return;
  }
  // 至少需要有首款才能添加
  if (!paymentItems.value.some(p => p.payment_type?.toUpperCase() === 'FIRST_PAYMENT')) {
    MessagePlugin.warning('请先配置首款');
    return;
  }

  const newIndex = paymentItems.value.length + 1;
  const now = new Date();
  const dueDate = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
  
  paymentItems.value.push({
    index: newIndex,
    payment_type: '',
    source_type: 'mold',
    amount_type: 'amount',
    amount_value: undefined as any,
    calculated_amount: 0,
    due_date: dueDate,
  });
};

const removeItem = (index: number) => {
  const payments = paymentItems.value;
  // 如果只有一项且有类型，删除后重置为空行
  if (payments.length <= 1 && payments[index]?.payment_type) {
    const now = new Date();
    const dueDate = new Date(now.getTime() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
    payments[index] = {
      index: 1,
      payment_type: '',
      source_type: 'mold',
      amount_type: 'amount',
      amount_value: undefined as any,
      calculated_amount: 0,
      due_date: dueDate,
    };
    emitChange();
    return;
  }
  if (paymentItems.value.length <= 1) {
    MessagePlugin.warning('至少保留一项款项');
    return;
  }
  paymentItems.value.splice(index, 1);
  recalculateIndexes();
  calculateAmounts();
  emitChange();
};

// 重新计算索引
const recalculateIndexes = () => {
  paymentItems.value.forEach((item, idx) => {
    item.index = idx + 1;
  });
};

const getAvailablePaymentTypes = (rowIndex: number) => {
  const hasTotal = paymentItems.value.some((p: any) => p.payment_type?.toUpperCase() === 'TOTAL');
  const hasFirst = paymentItems.value.some((p: any) => p.payment_type?.toUpperCase() === 'FIRST_PAYMENT');
  const hasFinal = paymentItems.value.some((p: any) => p.payment_type?.toUpperCase() === 'FINAL_PAYMENT');
  const currentType = paymentItems.value[rowIndex]?.payment_type?.toUpperCase();
  const isFirstRow = rowIndex === 0;

  return paymentTypeOptions.value.filter((item: any) => {
    const val = item.dict_value?.toUpperCase();

    // 如果已选了全款，允许「全款」或「首款」（可以从全款改回分期）
    if (hasTotal) {
      return val === 'TOTAL' || val === 'FIRST_PAYMENT';
    }

    // 第一行：只能选全款或首款
    if (isFirstRow) {
      return val === 'TOTAL' || val === 'FIRST_PAYMENT';
    }

    // 非第一行：必须已有首款才能添加其他类型
    if (!hasFirst) {
      return false;
    }

    // 已经有尾款了：不再允许选择
    if (hasFinal) {
      return val === 'FINAL_PAYMENT' && currentType === 'FINAL_PAYMENT';
    }

    // 有首款、无尾款的情况
    if (val === 'FIRST_PAYMENT') return false; // 首款只能有一个（第一行）
    if (val === 'TOTAL') return false; // 有了首款就不能选全款了
    return val === 'MIDDLE_PAYMENT' || val === 'FINAL_PAYMENT';
  });
};

// 款项类型变化
const handlePaymentTypeChange = (index: number) => {
  const row = paymentItems.value[index];
  if (!row) return;

  // 选择全款时：自动填充总金额，删除其他款项
  if (row.payment_type?.toUpperCase() === 'TOTAL') {
    row.amount_value = props.totalAmount;
    row.calculated_amount = props.totalAmount;
    row.amount_type = 'amount';
    paymentItems.value = [row];
    recalculateIndexes();
    emitChange();
    return;
  }

  // 选择首款时：如果之前是全款，清空金额让用户重新填写
  if (row.payment_type?.toUpperCase() === 'FIRST_PAYMENT') {
    // 从全款切换回首款，清空金额
    row.amount_value = undefined as any;
    row.calculated_amount = 0;
    emitChange();
    return;
  }
};

const calculateAmounts = () => {
  paymentItems.value.forEach(item => {
    if (item.amount_type === 'ratio') {
      item.calculated_amount = props.totalAmount * item.amount_value / 100;
    } else {
      item.calculated_amount = item.amount_value;
    }
  });

  const firstPayment = paymentItems.value.find(p => p.payment_type?.toUpperCase() === 'FIRST_PAYMENT');
  if (firstPayment && firstPayment.calculated_amount === props.totalAmount && props.totalAmount > 0) {
    firstPayment.payment_type = 'TOTAL';
    firstPayment.amount_value = props.totalAmount;
    firstPayment.amount_type = 'amount';
    paymentItems.value = [firstPayment];
    recalculateIndexes();
  }

  emitChange();
};

// 格式化金额
const formatAmount = (amount: number) => {
  return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

// 触发变更事件
const emitChange = () => {
  if (!enabled.value) {
    emit('change', { enabled: false, items: [] });
  } else {
    if (!isAmountMatched.value) {
      MessagePlugin.warning(`款项总额与订单总额不匹配，差额：¥${formatAmount(Math.abs(props.totalAmount - configuredTotal.value))}`);
    }
    emit('change', { 
      enabled: true, 
      items: paymentItems.value.map(item => ({
        index: item.index,
        payment_type: item.payment_type,
        source_type: item.source_type,
        amount_type: item.amount_type,
        amount_value: item.amount_value,
        calculated_amount: item.calculated_amount,
        due_date: item.due_date,
      }))
    });
  }
};

// 监听总额变化，重新计算
watch(() => props.totalAmount, () => {
  if (enabled.value) {
    calculateAmounts();
  }
});

onMounted(() => {
  // 款项类型字典已通过 useDictRaw 自动加载
});

// 暴露方法供父组件调用
defineExpose({
  getData: () => ({
    enabled: enabled.value,
    items: paymentItems.value,
  }),
  validate: () => {
    if (!enabled.value) {
      return true;
    }
    if (!isAmountMatched.value) {
      MessagePlugin.error('款项总额与订单总额不匹配');
      return false;
    }
    const hasEmptyType = paymentItems.value.some(item => !item.payment_type);
    if (hasEmptyType) {
      MessagePlugin.error('请为所有款项选择款项类型');
      return false;
    }

    // 有首款必须有尾款
    const hasFirst = paymentItems.value.some(p => p.payment_type?.toUpperCase() === 'FIRST_PAYMENT');
    const hasFinal = paymentItems.value.some(p => p.payment_type?.toUpperCase() === 'FINAL_PAYMENT');
    const hasTotal = paymentItems.value.some(p => p.payment_type?.toUpperCase() === 'TOTAL');

    if (!hasTotal && hasFirst && !hasFinal) {
      MessagePlugin.error('已配置首款，必须配置尾款');
      return false;
    }

    // 尾款必须是最后一项
    if (hasFinal) {
      const lastType = paymentItems.value[paymentItems.value.length - 1]?.payment_type?.toUpperCase();
      if (lastType !== 'FINAL_PAYMENT') {
        MessagePlugin.error('尾款必须是最后一笔款项，请调整款项顺序');
        return false;
      }
    }

    return true;
  },
});
</script>

<style scoped lang="less">
.payment-plan-config {
  margin-top: 24px;
  padding: 16px;
  background: var(--td-bg-color-container);
  border-radius: 4px;
  border: 1px solid var(--td-component-border);

  .config-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;

    .title {
      font-size: 14px;
      font-weight: 500;
      color: var(--td-text-color-primary);
    }
  }

  .payment-items-container {
    margin-top: 12px;

    .add-button-container {
      margin-top: 12px;
    }

    .summary-info {
      margin-top: 12px;
      padding: 12px;
      background: var(--td-bg-color-secondarycontainer);
      border-radius: 4px;
      font-size: 14px;

      .amount-text {
        font-weight: 500;
      }

      .text-success {
        color: var(--td-success-color);
      }

      .text-warning {
        color: var(--td-warning-color);
      }
    }
  }
}
</style>

<template>
  <div class="payment-plan-form">
    <div v-if="!props.hideStepUI" class="form-header">
      <div class="step-title">步骤 4：配置收款与应付计划</div>
      <t-alert
        theme="info"
        message="为模具、产品配置分期收款计划，外发加工的产品或模具可配置分期应付款项"
        style="margin-bottom: 16px"
      />
    </div>

    <!-- 订单汇总概览 -->
    <t-card :bordered="true" class="summary-card">
      <div class="summary-header">
        <span class="summary-title">订单汇总</span>
        <t-tag :theme="allValid ? 'success' : 'warning'" variant="light" size="small">
          {{ configuredCount }}/{{ totalItemsCount }} 项已配置
        </t-tag>
      </div>
      <div class="summary-body">
        <div class="summary-item">
          <span class="summary-label">模具</span>
          <span class="summary-count brand">{{ moldItems.length }} 项</span>
          <span class="summary-amount">¥{{ formatAmount(totalMoldAmount) }}</span>
        </div>
        <div class="summary-divider" />
        <div class="summary-item">
          <span class="summary-label">产品（应收）</span>
          <span class="summary-count success">{{ productionItems.length }} 项</span>
          <span class="summary-amount">¥{{ formatAmount(totalProductAmount) }}</span>
        </div>
        <div class="summary-divider" />
        <div class="summary-item">
          <span class="summary-label">订单总额</span>
          <span class="summary-amount large">¥{{ formatAmount(totalAmount) }}</span>
          <t-progress
            v-if="totalItemsCount > 0"
            :percentage="configuredPercent"
            :label="`${configuredPercent}% 已配置`"
            status="active"
            class="summary-progress"
          />
        </div>
        <div v-if="isOutsourced" class="summary-divider" />
        <div v-if="isOutsourced" class="summary-item primary">
          <span class="summary-label">生产外发应付</span>
          <span class="summary-amount outsource">¥{{ formatAmount(totalOutsourceAmount) }}</span>
        </div>
        <div v-if="hasMoldOutsource" class="summary-divider" />
        <div v-if="hasMoldOutsource" class="summary-item primary">
          <span class="summary-label">模具外发应付</span>
          <span class="summary-amount outsource"
            >¥{{ formatAmount(totalMoldOutsourceAmount) }}</span
          >
        </div>
      </div>
    </t-card>

    <!-- 模具款项配置 -->
    <div v-if="orderType !== 'PROD_ONLY' && moldItems.length > 0" class="section">
      <div class="section-header">
        <div class="header-left">
          <span class="section-icon">⚙️</span>
          <span class="section-title">模具应收款</span>
        </div>
        <t-tag theme="primary" variant="light">{{ moldItems.length }} 项</t-tag>
      </div>

      <t-collapse v-model="moldActivePanels" :borderless="true" expand-on-row-click>
        <t-collapse-panel
          v-for="(mold, index) in moldItems"
          :key="`mold-${index}`"
          :value="`mold-${index}`"
        >
          <template #header>
            <div class="panel-header">
              <t-tag theme="primary" size="small" variant="light-outline" class="item-index"
                >M{{ index + 1 }}</t-tag
              >
              <div class="panel-title">
                <span class="item-code">{{ mold.mold_code }}</span>
                <span v-if="mold.mold_name" class="item-subtitle">{{ mold.mold_name }}</span>
              </div>
              <div class="panel-amount">
                <span class="amount-total">¥{{ formatAmount(mold.amount || 0) }}</span>
              </div>
            </div>
          </template>

          <div class="panel-body">
            <div class="payment-toolbar">
              <t-button
                theme="default"
                variant="dashed"
                size="small"
                @click="addMoldPayment(index)"
              >
                <template #icon><add-icon /></template>
                添加期次
              </t-button>
              <t-tag v-if="isMoldAmountValid(index)" theme="success" variant="light" size="small">
                已配平
              </t-tag>
              <t-tag v-else theme="warning" variant="light" size="small">
                差额 ¥{{ formatAmount(getMoldDiff(index)) }}
              </t-tag>
            </div>

            <t-table
              :data="getMoldPayments(index)"
              :columns="paymentColumns"
              row-key="paymentIndex"
              :bordered="false"
              size="small"
              hover
              table-layout="fixed"
              @change="() => {}"
            >
              <template #seq="{ rowIndex }">{{ rowIndex + 1 }}</template>
              <template #payment_type="{ row, rowIndex }">
                <t-select
                  v-model="row.payment_type"
                  placeholder="选择款项类型"
                  style="width: 100%"
                  @change="(val: any) => handleMoldPaymentTypeChange(index, rowIndex, val)"
                >
                  <t-option
                    v-for="item in getAvailablePaymentTypes(index, rowIndex, 'mold')"
                    :key="item.dict_value"
                    :value="item.dict_value"
                    :label="item.dict_label"
                  />
                </t-select>
              </template>
              <template #amount="{ row, rowIndex }">
                <t-input-number
                  v-model="row.amount"
                  :min="0"
                  :max="mold.amount || 0"
                  :precision="2"
                  placeholder="金额"
                  theme="normal"
                  style="width: 100%"
                  @blur="handleMoldPaymentAmountBlur(index, rowIndex)"
                >
                  <template #suffix>元</template>
                </t-input-number>
              </template>
              <template #due_date="{ row }">
                <t-date-picker
                  v-model="row.due_date"
                  placeholder="到期日"
                  format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </template>
              <template #op="{ rowIndex }">
                <t-link
                  theme="danger"
                  :underline="false"
                  @click="removeMoldPayment(index, rowIndex)"
                  >删除</t-link
                >
              </template>
            </t-table>
          </div>
        </t-collapse-panel>
      </t-collapse>
    </div>

    <!-- ⚙️ 模具外发应付款配置（仅模具中有外发加工时显示） -->
    <div v-if="hasMoldOutsource" class="section">
      <div class="section-header outsource-header">
        <div class="header-left">
          <span class="section-icon">🔧</span>
          <span class="section-title">模具外发应付款</span>
        </div>
        <t-tag theme="warning" variant="light">{{ moldOutsourceItems.length }} 项</t-tag>
      </div>

      <t-collapse v-model="moldOutsourceActivePanels" :borderless="true" expand-on-row-click>
        <t-collapse-panel
          v-for="(mold, index) in moldOutsourceItems"
          :key="`mold-outsource-${index}`"
          :value="`mold-outsource-${index}`"
        >
          <template #header>
            <div class="panel-header">
              <t-tag theme="warning" size="small" variant="light-outline" class="item-index"
                >MO{{ index + 1 }}</t-tag
              >
              <div class="panel-title">
                <span class="item-code">{{ mold.mold_code }}</span>
                <span class="item-subtitle processor-name">{{ mold.mold_name }}</span>
              </div>
              <div class="panel-amount">
                <span class="amount-total outsource-amount"
                  >¥{{ formatAmount(getMoldOutsourceConfiguredAmount(index)) }}</span
                >
              </div>
            </div>
          </template>

          <div class="panel-body">
            <div class="payment-toolbar">
              <t-button
                theme="default"
                variant="dashed"
                size="small"
                @click="addMoldOutsourcePayment(index)"
              >
                <template #icon><add-icon /></template>
                添加期次
              </t-button>
              <t-tag
                v-if="moldOutsourcePayments[index]?.length"
                theme="warning"
                variant="light"
                size="small"
              >
                共 {{ moldOutsourcePayments[index].length }} 期
              </t-tag>
              <t-tag v-else theme="default" variant="light" size="small">未配置</t-tag>
            </div>

            <t-table
              :data="getMoldOutsourcePayments(index)"
              :columns="paymentColumns"
              row-key="paymentIndex"
              :bordered="false"
              size="small"
              hover
              table-layout="fixed"
            >
              <template #seq="{ rowIndex }">{{ rowIndex + 1 }}</template>
              <template #payment_type="{ row, rowIndex }">
                <t-select
                  v-model="row.payment_type"
                  placeholder="选择应付类型"
                  style="width: 100%"
                  @change="(val: any) => handleMoldOutsourcePaymentTypeChange(index, rowIndex, val)"
                >
                  <t-option
                    v-for="item in getAvailablePaymentTypes(index, rowIndex, 'mold_outsource')"
                    :key="item.dict_value"
                    :value="item.dict_value"
                    :label="item.dict_label"
                  />
                </t-select>
              </template>
              <template #amount="{ row }">
                <t-input-number
                  v-model="row.amount"
                  :min="0"
                  :precision="2"
                  placeholder="金额"
                  theme="normal"
                  style="width: 100%"
                >
                  <template #suffix>元</template>
                </t-input-number>
              </template>
              <template #due_date="{ row }">
                <t-date-picker
                  v-model="row.due_date"
                  placeholder="到期日"
                  format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </template>
              <template #op="{ rowIndex }">
                <t-link
                  theme="danger"
                  :underline="false"
                  @click="removeMoldOutsourcePayment(index, rowIndex)"
                  >删除</t-link
                >
              </template>
            </t-table>
          </div>
        </t-collapse-panel>
      </t-collapse>
    </div>

    <!-- 产品款项配置 -->
    <div v-if="orderType !== 'MOLD_ONLY' && productionItems.length > 0" class="section">
      <div class="section-header product-header">
        <div class="header-left">
          <span class="section-icon">📦</span>
          <span class="section-title">产品应收款</span>
        </div>
        <t-tag theme="success" variant="light">{{ productionItems.length }} 项</t-tag>
      </div>

      <t-collapse v-model="productActivePanels" :borderless="true" expand-on-row-click>
        <t-collapse-panel
          v-for="(product, index) in productionItems"
          :key="`product-${index}`"
          :value="`product-${index}`"
        >
          <template #header>
            <div class="panel-header">
              <t-tag theme="success" size="small" variant="light-outline" class="item-index"
                >P{{ index + 1 }}</t-tag
              >
              <div class="panel-title">
                <span class="item-code">{{ product.product_name }}</span>
                <span v-if="getProductMoldCode(product)" class="item-subtitle">{{
                  getProductMoldCode(product)
                }}</span>
              </div>
              <div class="panel-amount">
                <span class="amount-total"
                  >¥{{ formatAmount(calculateProductAmount(product)) }}</span
                >
              </div>
            </div>
          </template>

          <div class="panel-body">
            <!-- 金额计算行 -->
            <div class="calc-row">
              <div class="calc-row-header">
                <span class="calc-row-title">金额计算</span>
              </div>
              <div class="calc-grid">
                <div class="calc-field">
                  <label class="calc-label">单价</label>
                  <t-input-number
                    v-model="product.unit_price"
                    :min="0"
                    :precision="2"
                    theme="normal"
                    style="width: 100%"
                    @change="() => {}"
                  >
                    <template #suffix>元</template>
                  </t-input-number>
                </div>
                <div class="calc-field">
                  <label class="calc-label">数量</label>
                  <span class="calc-qty">{{ product.plan_qty || 1 }}</span>
                </div>
                <div v-if="isOutsourced" class="calc-field">
                  <label class="calc-label">外发单价</label>
                  <t-input-number
                    v-model="product.outsource_price"
                    :min="0"
                    :precision="2"
                    theme="normal"
                    style="width: 100%"
                    @change="() => {}"
                  >
                    <template #suffix>元</template>
                  </t-input-number>
                </div>
                <div class="calc-field">
                  <label class="calc-label">包装费</label>
                  <t-input-number
                    v-model="product.packaging_fee"
                    :min="0"
                    :precision="2"
                    theme="normal"
                    style="width: 100%"
                    @change="() => {}"
                  >
                    <template #suffix>元</template>
                  </t-input-number>
                </div>
                <div class="calc-field">
                  <label class="calc-label">运费</label>
                  <t-input-number
                    v-model="product.freight_fee"
                    :min="0"
                    :precision="2"
                    theme="normal"
                    style="width: 100%"
                    @change="() => {}"
                  >
                    <template #suffix>元</template>
                  </t-input-number>
                </div>
                <div class="calc-field calc-result">
                  <label class="calc-label">计算结果</label>
                  <div class="calc-result-value">
                    <span class="calc-amount"
                      >¥{{ formatAmount(calculateProductAmount(product)) }}</span
                    >
                  </div>
                </div>
              </div>
            </div>

            <div class="payment-toolbar">
              <t-button
                theme="default"
                variant="dashed"
                size="small"
                @click="addProductPayment(index)"
              >
                <template #icon><add-icon /></template>
                添加期次
              </t-button>
              <t-tag
                v-if="isProductAmountValid(index)"
                theme="success"
                variant="light"
                size="small"
              >
                已配平
              </t-tag>
              <t-tag v-else theme="warning" variant="light" size="small">
                差额 ¥{{ formatAmount(getProductDiff(index)) }}
              </t-tag>
            </div>

            <t-table
              :data="getProductPayments(index)"
              :columns="paymentColumns"
              row-key="paymentIndex"
              :bordered="false"
              size="small"
              hover
              table-layout="fixed"
            >
              <template #seq="{ rowIndex }">{{ rowIndex + 1 }}</template>
              <template #payment_type="{ row, rowIndex }">
                <t-select
                  v-model="row.payment_type"
                  placeholder="选择款项类型"
                  style="width: 100%"
                  @change="(val: any) => handleProductPaymentTypeChange(index, rowIndex, val)"
                >
                  <t-option
                    v-for="item in getAvailablePaymentTypes(index, rowIndex, 'product')"
                    :key="item.dict_value"
                    :value="item.dict_value"
                    :label="item.dict_label"
                  />
                </t-select>
              </template>
              <template #amount="{ row, rowIndex }">
                <t-input-number
                  v-model="row.amount"
                  :min="0"
                  :max="getProductPaymentMax(index)"
                  :precision="2"
                  placeholder="金额"
                  theme="normal"
                  style="width: 100%"
                  @blur="handleProductPaymentAmountBlur(index, rowIndex)"
                >
                  <template #suffix>元</template>
                </t-input-number>
              </template>
              <template #due_date="{ row }">
                <t-date-picker
                  v-model="row.due_date"
                  placeholder="到期日"
                  format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </template>
              <template #op="{ rowIndex }">
                <t-link
                  theme="danger"
                  :underline="false"
                  @click="removeProductPayment(index, rowIndex)"
                  >删除</t-link
                >
              </template>
            </t-table>
          </div>
        </t-collapse-panel>
      </t-collapse>
    </div>

    <!-- ✨ 外发应付款配置（仅外发生产时显示） -->
    <div v-if="isOutsourced && productionItems.length > 0" class="section">
      <div class="section-header outsource-header">
        <div class="header-left">
          <span class="section-icon">💰</span>
          <span class="section-title">外发应付款</span>
        </div>
        <t-tag theme="warning" variant="light">{{ productionItems.length }} 项</t-tag>
      </div>

      <t-collapse v-model="outsourceActivePanels" :borderless="true" expand-on-row-click>
        <t-collapse-panel
          v-for="(product, index) in productionItems"
          :key="`outsource-${index}`"
          :value="`outsource-${index}`"
        >
          <template #header>
            <div class="panel-header">
              <t-tag theme="warning" size="small" variant="light-outline" class="item-index"
                >O{{ index + 1 }}</t-tag
              >
              <div class="panel-title">
                <span class="item-code">{{ product.product_name }}</span>
                <span v-if="processorName" class="item-subtitle processor-name">{{
                  processorName
                }}</span>
              </div>
              <div class="panel-amount">
                <span class="amount-total outsource-amount"
                  >¥{{ formatAmount(getOutsourceConfiguredAmount(index)) }}</span
                >
              </div>
            </div>
          </template>

          <div class="panel-body">
            <div class="payment-toolbar">
              <t-button
                theme="default"
                variant="dashed"
                size="small"
                @click="addOutsourcePayment(index)"
              >
                <template #icon><add-icon /></template>
                添加期次
              </t-button>
              <t-tag
                v-if="outsourcePayments[index]?.length"
                theme="warning"
                variant="light"
                size="small"
              >
                共 {{ outsourcePayments[index].length }} 期
              </t-tag>
              <t-tag v-else theme="default" variant="light" size="small">未配置</t-tag>
            </div>

            <t-table
              :data="getOutsourcePayments(index)"
              :columns="paymentColumns"
              row-key="paymentIndex"
              :bordered="false"
              size="small"
              hover
              table-layout="fixed"
            >
              <template #seq="{ rowIndex }">{{ rowIndex + 1 }}</template>
              <template #payment_type="{ row, rowIndex }">
                <t-select
                  v-model="row.payment_type"
                  placeholder="选择应付类型"
                  style="width: 100%"
                  @change="(val: any) => handleOutsourcePaymentTypeChange(index, rowIndex, val)"
                >
                  <t-option
                    v-for="item in getAvailablePaymentTypes(index, rowIndex, 'outsource')"
                    :key="item.dict_value"
                    :value="item.dict_value"
                    :label="item.dict_label"
                  />
                </t-select>
              </template>
              <template #amount="{ row }">
                <t-input-number
                  v-model="row.amount"
                  :min="0"
                  :precision="2"
                  placeholder="金额"
                  theme="normal"
                  style="width: 100%"
                >
                  <template #suffix>元</template>
                </t-input-number>
              </template>
              <template #due_date="{ row }">
                <t-date-picker
                  v-model="row.due_date"
                  placeholder="到期日"
                  format="YYYY-MM-DD"
                  style="width: 100%"
                />
              </template>
              <template #op="{ rowIndex }">
                <t-link
                  theme="danger"
                  :underline="false"
                  @click="removeOutsourcePayment(index, rowIndex)"
                  >删除</t-link
                >
              </template>
            </t-table>
          </div>
        </t-collapse-panel>
      </t-collapse>
    </div>

    <!-- 操作按钮 -->
    <div v-if="!props.hideStepUI" class="action-buttons">
      <t-button theme="default" variant="outline" @click="$emit('prev')">上一步</t-button>
      <t-button theme="primary" @click="handleNext">
        <template #icon><check-icon /></template>
        完成配置
      </t-button>
    </div>
  </div>
</template>
<script setup lang="ts">
import { AddIcon, CheckIcon } from 'tdesign-icons-vue-next';
import type { PrimaryTableCol } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, ref, watch } from 'vue';

import { useDictRaw } from '@/hooks/useDict';

const props = defineProps<{
  orderType: string;
  moldItems: any[];
  productionItems: any[];
  productionType?: string; // ✨ 生产类型：IN_HOUSE / OUTSOURCED
  outsourcePrice?: number; // ✨ 外发单价（订单级别，同步到每个明细项用于外发应付款计算）
  processorName?: string; // ✨ 加工商名称（外发时显示）
  hideStepUI?: boolean; // ✨ 隐藏步骤标题和按钮（用于单页表单）
  additionalAmount?: number; // ✨ 额外附加金额（如运费等）
  initialMoldPayments?: Record<
    number,
    Array<{
      payment_type: string;
      amount: number;
      due_date: string;
    }>
  >; // ✨ 编辑模式：预填充的模具款项数据（key 为模具索引）
  initialProductPayments?: Record<
    number,
    Array<{
      payment_type: string;
      amount: number;
      due_date: string;
    }>
  >; // ✨ 编辑模式：预填充的产品款项数据（key 为产品索引）
}>();

const emit = defineEmits<{
  (e: 'next', data: any): void;
  (e: 'prev'): void;
}>();

// 款项类型字典（useDictRaw 自动加载 + 全局缓存）
const { items: paymentTypeOptions } = useDictRaw('payment-type');
const moldActivePanels = ref<string[]>([]);
const productActivePanels = ref<string[]>([]);
const outsourceActivePanels = ref<string[]>([]); // ✨ 外发应付款折叠面板
const moldOutsourceActivePanels = ref<string[]>([]); // ✨ 模具外发应付款折叠面板

// 表格列配置（紧凑版）
const paymentColumns: PrimaryTableCol[] = [
  { colKey: 'seq', title: '#', width: 50 },
  { colKey: 'payment_type', title: '款项类型', width: 160 },
  { colKey: 'amount', title: '金额（元）', width: 160 },
  { colKey: 'due_date', title: '到期日期', width: 150 },
  { colKey: 'op', title: '操作', width: 60, fixed: 'right' },
];

// 款项配置数据
const moldPayments = ref<Record<number, any[]>>({});
const productPayments = ref<Record<number, any[]>>({});
const outsourcePayments = ref<Record<number, any[]>>({}); // ✨ 外发应付款
const moldOutsourcePayments = ref<Record<number, any[]>>({}); // ✨ 模具外发应付款

// ✨ 是否为外发生产模式
const isOutsourced = computed(() => props.productionType === 'OUTSOURCED');

// ✨ 模具中是否有外发加工项
const hasMoldOutsource = computed(() => {
  return props.moldItems.some(
    (item: any) => item.workType === 'OUTSOURCED' || item.work_type === 'OUTSOURCED',
  );
});

// ✨ 外发加工的模具项
const moldOutsourceItems = computed(() => {
  return props.moldItems.filter(
    (item: any) => item.workType === 'OUTSOURCED' || item.work_type === 'OUTSOURCED',
  );
});

// ✨ 计算外发应付金额（外发单价 × 数量）
// 优先使用 props.outsourcePrice（订单级），其次回退到 item.outsource_price
const calculateOutsourceAmount = (product: any) => {
  const price = props.outsourcePrice ?? (product.outsource_price || 0);
  return price * (product.plan_qty || 1);
};

// ✨ 监听外发单价/数量变化，自动联动"全款"行的金额
watch(
  () => [props.outsourcePrice, ...props.productionItems.map((p: any) => p.plan_qty)],
  () => {
    props.productionItems.forEach((product: any, i: number) => {
      const payments = outsourcePayments.value[i];
      if (!payments || payments.length === 0) return;
      payments.forEach((row: any) => {
        if (row.payment_type?.toUpperCase() === 'TOTAL') {
          row.amount = calculateOutsourceAmount(product);
        }
      });
    });
  },
  { deep: true },
);

// ✨ 外发应付总额（实际已填写的付款总额）
const totalOutsourceAmount = computed(() => {
  let total = 0;
  for (let i = 0; i < productionItemsInternal.value.length; i++) {
    total += getOutsourceConfiguredAmount(i);
  }
  return total;
});

const calculateProductAmount = (product: any) => {
  return (
    (product.unit_price || 0) * (product.plan_qty || 1) +
    (product.packaging_fee || 0) +
    (product.freight_fee || 0)
  );
};

// ✨ 获取产品款项金额上限（编辑模式下优先使用预填充数据的总和）
const getProductPaymentMax = (index: number) => {
  // 如果有预填充数据，使用预填充金额的最大值作为上限（避免红框）
  const initialData = props.initialProductPayments;
  if (initialData && initialData[index] && initialData[index].length > 0) {
    const prefillTotal = initialData[index].reduce((sum, p) => sum + (p.amount || 0), 0);
    if (prefillTotal > 0) {
      // 取预填充总额的1.5倍或产品计算金额中的较大值
      const calcAmount = calculateProductAmount(props.productionItems[index] || {});
      return Math.max(prefillTotal * 1.5, calcAmount);
    }
  }
  return calculateProductAmount(props.productionItems[index] || {});
};

const getProductMoldCode = (product: any): string => {
  // 优先使用 ProductItemsForm 选择模具时保存的显示字段（不会被 watch 清除）
  if (product.mold_code_display) return product.mold_code_display;
  // 使用产品自带的 mold_code（未被清除时）
  if (product.mold_code) return product.mold_code;
  // 外来模具编号
  if (product.external_mold_code) return product.external_mold_code;
  // 通过 mold_id 在 moldItems 中查找
  if (product.mold_id) {
    const mold = props.moldItems.find((m: any) => String(m.mold_id) === String(product.mold_id));
    if (mold?.mold_code) return mold.mold_code;
  }
  // 通过多模具 IDs 查找
  if (product.mold_ids?.length) {
    const codes = product.mold_ids
      .map((id: number) => {
        const mold = props.moldItems.find((m: any) => String(m.mold_id) === String(id));
        return mold?.mold_code;
      })
      .filter(Boolean);
    if (codes.length) return codes.join(', ');
  }
  return '';
};

// 总额计算
// ✨ 计算模具总金额（含附加金额）
const totalMoldAmount = computed(() => {
  const moldAmount = props.moldItems.reduce((s, i) => s + (i.amount || 0), 0);
  return moldAmount + (props.additionalAmount || 0);
});
// ✨ 使用 ref + watch 维护 productionItems 的副本，确保响应式追踪
const productionItemsInternal = ref<any[]>([]);

watch(
  () => props.productionItems,
  (val) => {
    productionItemsInternal.value = JSON.parse(JSON.stringify(val || []));
  },
  { immediate: true, deep: true },
);

// 同步暴露给模板使用
const productionItemsView = computed(() => productionItemsInternal.value);

const totalProductAmount = computed(() => {
  return productionItemsInternal.value.reduce(
    (s: number, i: any) => s + calculateProductAmount(i),
    0,
  );
});
const totalAmount = computed(
  () => totalMoldAmount.value + totalProductAmount.value + (props.additionalAmount || 0),
);
const totalItemsCount = computed(() => props.moldItems.length + props.productionItems.length);

// 已配置项数量
const configuredCount = computed(() => {
  let count = 0;
  for (let i = 0; i < props.moldItems.length; i++) {
    if (isMoldAmountValid(i)) count++;
  }
  for (let i = 0; i < props.productionItems.length; i++) {
    if (isProductAmountValid(i)) count++;
  }
  return count;
});

const allValid = computed(() => configuredCount.value === totalItemsCount.value);
const configuredPercent = computed(() =>
  totalItemsCount.value > 0 ? Math.round((configuredCount.value / totalItemsCount.value) * 100) : 0,
);

// 格式化金额
const formatAmount = (amount: number) =>
  amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });

// 款项 CRUD（模具）
const getMoldPayments = (index: number) => {
  if (!moldPayments.value[index]) moldPayments.value[index] = [];
  return moldPayments.value[index];
};

const addMoldPayment = (moldIndex: number) => {
  if (
    (moldPayments.value[moldIndex] || []).some((p) => p.payment_type?.toUpperCase() === 'TOTAL')
  ) {
    MessagePlugin.warning('全款已配置，无法添加更多期次');
    return;
  }
  if (paymentListHasFinal(moldIndex, 'mold')) {
    MessagePlugin.warning('尾款必须是最后一笔款项，无法再添加');
    return;
  }
  // 至少需要有首款才能添加（但如果用户还没选任何类型，允许直接添加）
  const hasAnyConfigured = (moldPayments.value[moldIndex] || []).some((p) => !!p.payment_type);
  if (hasAnyConfigured && !paymentListHasFirst(moldIndex, 'mold')) {
    MessagePlugin.warning('请先配置首款');
    return;
  }
  const dueDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
  if (!moldPayments.value[moldIndex]) moldPayments.value[moldIndex] = [];
  moldPayments.value[moldIndex].push({
    paymentIndex: Date.now(),
    payment_type: '',
    amount: undefined as any,
    due_date: dueDate,
  });
};

const removeMoldPayment = (moldIndex: number, paymentIndex: number) => {
  const payments = moldPayments.value[moldIndex];
  if (!payments) return;
  // 如果只有一项且有类型，删除后重置为空行
  if (payments.length <= 1 && payments[paymentIndex]?.payment_type) {
    payments[paymentIndex] = {
      paymentIndex: Date.now(),
      payment_type: '',
      amount: undefined as any,
      due_date: getDefaultDueDate(),
    };
    return;
  }
  if (payments.length <= 1) {
    MessagePlugin.warning('至少保留一项');
    return;
  }
  payments.splice(paymentIndex, 1);
};

// 款项 CRUD（产品）
const getProductPayments = (index: number) => {
  if (!productPayments.value[index]) productPayments.value[index] = [];
  return productPayments.value[index];
};

const addProductPayment = (productIndex: number) => {
  if (
    (productPayments.value[productIndex] || []).some(
      (p) => p.payment_type?.toUpperCase() === 'TOTAL',
    )
  ) {
    MessagePlugin.warning('全款已配置，无法添加更多期次');
    return;
  }
  if (paymentListHasFinal(productIndex, 'product')) {
    MessagePlugin.warning('尾款必须是最后一笔款项，无法再添加');
    return;
  }
  // 至少需要有首款才能添加（但如果用户还没选任何类型，允许直接添加）
  const hasAnyConfigured = (productPayments.value[productIndex] || []).some((p) => !!p.payment_type);
  if (hasAnyConfigured && !paymentListHasFirst(productIndex, 'product')) {
    MessagePlugin.warning('请先配置首款');
    return;
  }
  const dueDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
  if (!productPayments.value[productIndex]) productPayments.value[productIndex] = [];
  productPayments.value[productIndex].push({
    paymentIndex: Date.now(),
    payment_type: '',
    amount: undefined as any,
    due_date: dueDate,
  });
};

const removeProductPayment = (productIndex: number, paymentIndex: number) => {
  const payments = productPayments.value[productIndex];
  if (!payments) return;
  if (payments.length <= 1 && payments[paymentIndex]?.payment_type) {
    payments[paymentIndex] = {
      paymentIndex: Date.now(),
      payment_type: '',
      amount: undefined as any,
      due_date: getDefaultDueDate(),
    };
    return;
  }
  if (payments.length <= 1) {
    MessagePlugin.warning('至少保留一项');
    return;
  }
  payments.splice(paymentIndex, 1);
};

// ✨ 外发应付款 CRUD
const getOutsourcePayments = (index: number) => {
  if (!outsourcePayments.value[index]) outsourcePayments.value[index] = [];
  return outsourcePayments.value[index];
};

const addOutsourcePayment = (productIndex: number) => {
  if (
    (outsourcePayments.value[productIndex] || []).some(
      (p) => p.payment_type?.toUpperCase() === 'TOTAL',
    )
  ) {
    MessagePlugin.warning('全款已配置，无法添加更多期次');
    return;
  }
  if (paymentListHasFinal(productIndex, 'outsource')) {
    MessagePlugin.warning('尾款必须是最后一笔款项，无法再添加');
    return;
  }
  // 至少需要有首款才能添加（但如果用户还没选任何类型，允许直接添加）
  const hasAnyConfigured = (outsourcePayments.value[productIndex] || []).some((p) => !!p.payment_type);
  if (hasAnyConfigured && !paymentListHasFirst(productIndex, 'outsource')) {
    MessagePlugin.warning('请先配置首款');
    return;
  }
  const dueDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
  if (!outsourcePayments.value[productIndex]) outsourcePayments.value[productIndex] = [];
  outsourcePayments.value[productIndex].push({
    paymentIndex: Date.now(),
    payment_type: '',
    amount: undefined as any,
    due_date: dueDate,
  });
};

const removeOutsourcePayment = (productIndex: number, paymentIndex: number) => {
  const payments = outsourcePayments.value[productIndex];
  if (!payments) return;
  if (payments.length <= 1 && payments[paymentIndex]?.payment_type) {
    payments[paymentIndex] = {
      paymentIndex: Date.now(),
      payment_type: '',
      amount: undefined as any,
      due_date: getDefaultDueDate(),
    };
    return;
  }
  if (!payments || payments.length <= 1) {
    MessagePlugin.warning('至少保留一项');
    return;
  }
  payments.splice(paymentIndex, 1);
};

// 金额验证：金额配平 + 如果选了首款则必须有尾款才算配平
const getMoldConfiguredAmount = (index: number) =>
  (moldPayments.value[index] || []).reduce((s, i) => s + (i.amount || 0), 0);
const getMoldDiff = (index: number) =>
  Math.abs((props.moldItems[index]?.amount || 0) - getMoldConfiguredAmount(index));
const isMoldAmountValid = (index: number) => {
  const payments = moldPayments.value[index] || [];
  const amountOk = (props.moldItems[index]?.amount || 0) > 0 && getMoldDiff(index) < 0.01;
  if (!amountOk) return false;
  // 有首款必须有尾款才配平，全款模式直接通过
  const hasFirst = payments.some((p: any) => p.payment_type?.toUpperCase() === 'FIRST_PAYMENT');
  const hasFinal = payments.some((p: any) => p.payment_type?.toUpperCase() === 'FINAL_PAYMENT');
  const hasTotal = payments.some((p: any) => p.payment_type?.toUpperCase() === 'TOTAL');
  if (hasTotal) return true;
  if (hasFirst) return hasFinal;
  return true;
};

const getProductConfiguredAmount = (index: number) =>
  (productPayments.value[index] || []).reduce((s, i) => s + (i.amount || 0), 0);
const getProductDiff = (index: number) =>
  Math.abs(
    calculateProductAmount(props.productionItems[index]) - getProductConfiguredAmount(index),
  );
const isProductAmountValid = (index: number) => {
  const payments = productPayments.value[index] || [];
  const amountOk =
    calculateProductAmount(props.productionItems[index]) > 0 && getProductDiff(index) < 0.01;
  if (!amountOk) return false;
  const hasFirst = payments.some((p: any) => p.payment_type?.toUpperCase() === 'FIRST_PAYMENT');
  const hasFinal = payments.some((p: any) => p.payment_type?.toUpperCase() === 'FINAL_PAYMENT');
  const hasTotal = payments.some((p: any) => p.payment_type?.toUpperCase() === 'TOTAL');
  if (hasTotal) return true;
  if (hasFirst) return hasFinal;
  return true;
};

const getOutsourceConfiguredAmount = (index: number) => {
  const payments = outsourcePayments.value[index];
  return payments ? payments.reduce((s: number, i: any) => s + (i.amount || 0), 0) : 0;
};

// ✨ 模具外发应付款 CRUD
const getMoldOutsourcePayments = (index: number) => {
  if (!moldOutsourcePayments.value[index]) moldOutsourcePayments.value[index] = [];
  return moldOutsourcePayments.value[index];
};

const addMoldOutsourcePayment = (moldIndex: number) => {
  if (
    (moldOutsourcePayments.value[moldIndex] || []).some(
      (p) => p.payment_type?.toUpperCase() === 'TOTAL',
    )
  ) {
    MessagePlugin.warning('全款已配置，无法添加更多期次');
    return;
  }
  if (paymentListHasFinal(moldIndex, 'mold_outsource')) {
    MessagePlugin.warning('尾款必须是最后一笔款项，无法再添加');
    return;
  }
  // 至少需要有首款才能添加（但如果用户还没选任何类型，允许直接添加）
  const hasAnyConfigured = (moldOutsourcePayments.value[moldIndex] || []).some((p) => !!p.payment_type);
  if (hasAnyConfigured && !paymentListHasFirst(moldIndex, 'mold_outsource')) {
    MessagePlugin.warning('请先配置首款');
    return;
  }
  const dueDate = new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
  if (!moldOutsourcePayments.value[moldIndex]) moldOutsourcePayments.value[moldIndex] = [];
  moldOutsourcePayments.value[moldIndex].push({
    paymentIndex: Date.now(),
    payment_type: '',
    amount: undefined as any,
    due_date: dueDate,
  });
};

const removeMoldOutsourcePayment = (moldIndex: number, paymentIndex: number) => {
  const payments = moldOutsourcePayments.value[moldIndex];
  if (!payments) return;
  if (payments.length <= 1 && payments[paymentIndex]?.payment_type) {
    payments[paymentIndex] = {
      paymentIndex: Date.now(),
      payment_type: '',
      amount: undefined as any,
      due_date: getDefaultDueDate(),
    };
    return;
  }
  if (!payments || payments.length <= 1) {
    MessagePlugin.warning('至少保留一项');
    return;
  }
  payments.splice(paymentIndex, 1);
};

const getMoldOutsourceConfiguredAmount = (index: number) => {
  const payments = moldOutsourcePayments.value[index];
  return payments ? payments.reduce((s: number, i: any) => s + (i.amount || 0), 0) : 0;
};

// ✨ 模具外发应付总额
const totalMoldOutsourceAmount = computed(() => {
  let total = 0;
  for (let i = 0; i < moldOutsourceItems.value.length; i++) {
    total += getMoldOutsourceConfiguredAmount(i);
  }
  return total;
});

const getAvailablePaymentTypes = (
  listIndex: number,
  rowIndex: number,
  source: 'mold' | 'product' | 'outsource' | 'mold_outsource',
) => {
  const payments = getPaymentListBySource(listIndex, source);
  const isFirstRow = rowIndex === 0;
  const hasTotal = payments.some((p: any) => p.payment_type?.toUpperCase() === 'TOTAL');
  const hasFirst = payments.some((p: any) => p.payment_type?.toUpperCase() === 'FIRST_PAYMENT');
  const hasFinal = payments.some((p: any) => p.payment_type?.toUpperCase() === 'FINAL_PAYMENT');
  const currentType = payments[rowIndex]?.payment_type?.toUpperCase();

  return paymentTypeOptions.value.filter((item: any) => {
    const val = item.dict_value?.toUpperCase();

    // 如果已选了全款，只允许「全款」或「首款」（可以从全款改回分期）
    if (hasTotal) {
      return val === 'TOTAL' || val === 'FIRST_PAYMENT';
    }

    // 第一行：只能选全款或首款
    if (isFirstRow) {
      return val === 'TOTAL' || val === 'FIRST_PAYMENT';
    }

    // 非第一行：必须已有首款才能添加其他类型
    // 但如果用户还没配置任何款项，允许非第一行也选择首款或全款
    if (!hasFirst) {
      const hasAnyConfigured = payments.some((p: any) => !!p.payment_type);
      if (hasAnyConfigured) {
        return false;
      }
      return val === 'TOTAL' || val === 'FIRST_PAYMENT';
    }

    // 已经有尾款了：不再允许选择（尾款必须最后）
    if (hasFinal) {
      // 如果当前行就是尾款，允许保留
      return val === 'FINAL_PAYMENT' && currentType === 'FINAL_PAYMENT';
    }

    // 有首款、无尾款的情况
    if (val === 'FIRST_PAYMENT') return false; // 首款只能有一个（第一行）
    if (val === 'TOTAL') return false; // 有了首款就不能选全款了
    return val === 'MIDDLE_PAYMENT' || val === 'FINAL_PAYMENT';
  });
};

// 根据 source 类型获取对应的 payment 列表
const getPaymentListBySource = (listIndex: number, source: string) => {
  switch (source) {
    case 'mold':
      return getMoldPayments(listIndex);
    case 'product':
      return getProductPayments(listIndex);
    case 'outsource':
      return getOutsourcePayments(listIndex);
    case 'mold_outsource':
      return getMoldOutsourcePayments(listIndex);
    default:
      return [];
  }
};

const paymentListHas = (listIndex: number, source: string, type: string) => {
  const payments = getPaymentListBySource(listIndex, source);
  return payments.some((p: any) => p.payment_type?.toUpperCase() === type.toUpperCase());
};

const paymentListHasFinal = (listIndex: number, source: string) =>
  paymentListHas(listIndex, source, 'FINAL_PAYMENT');
const paymentListHasFirst = (listIndex: number, source: string) =>
  paymentListHas(listIndex, source, 'FIRST_PAYMENT');

// 模具款项类型变更处理
const handleMoldPaymentTypeChange = (moldIndex: number, rowIndex: number, val: string) => {
  const payments = moldPayments.value[moldIndex];
  if (!payments) return;

  const totalAmount = props.moldItems[moldIndex]?.amount || 0;

  if (val?.toUpperCase() === 'TOTAL') {
    const row = payments[rowIndex];
    if (!row) return;
    row.amount = totalAmount;
    moldPayments.value[moldIndex] = [row];
    return;
  }

  // 从全款切换为首款：清空金额，保留当前行
  if (val?.toUpperCase() === 'FIRST_PAYMENT') {
    const row = payments[rowIndex];
    if (!row) return;
    // 如果之前是全款，切回首款时清空金额让用户重新填写
    if (payments.length === 1 && payments.some((p) => p.payment_type?.toUpperCase() === 'TOTAL')) {
      // 已经是干净的单行（上面的TOTAL分支已处理），直接到清空逻辑
    }
    row.amount = undefined as any;
  }
};

// 模具款项金额失焦处理（首款金额达到总额时自动切换为全款）
const handleMoldPaymentAmountBlur = (moldIndex: number, rowIndex: number) => {
  const payments = moldPayments.value[moldIndex];
  if (!payments || rowIndex !== 0) return;

  const totalAmount = Number(props.moldItems[moldIndex]?.amount || 0);
  const row = payments[rowIndex];
  if (!row || !row.payment_type) return;

  const currentAmount = Number(row.amount || 0);
  if (
    row.payment_type?.toUpperCase() === 'FIRST_PAYMENT' &&
    currentAmount === totalAmount &&
    totalAmount > 0
  ) {
    row.payment_type = 'TOTAL';
    moldPayments.value[moldIndex] = [row];
  }
};

// 产品款项金额变更处理（首款金额达到总额时自动切换为全款）
const handleProductPaymentAmountBlur = (productIndex: number, rowIndex: number) => {
  const payments = productPayments.value[productIndex];
  if (!payments || rowIndex !== 0) return;

  const totalAmount = Number(calculateProductAmount(props.productionItems[productIndex]));
  const row = payments[rowIndex];
  if (!row || !row.payment_type) return;

  const currentAmount = Number(row.amount || 0);
  if (
    row.payment_type?.toUpperCase() === 'FIRST_PAYMENT' &&
    currentAmount === totalAmount &&
    totalAmount > 0
  ) {
    row.payment_type = 'TOTAL';
    productPayments.value[productIndex] = [row];
  }
};

// 产品款项类型变更处理
const handleProductPaymentTypeChange = (productIndex: number, rowIndex: number, val: string) => {
  const payments = productPayments.value[productIndex];
  if (!payments) return;

  const totalAmount = calculateProductAmount(props.productionItems[productIndex]);

  if (val?.toUpperCase() === 'TOTAL') {
    const row = payments[rowIndex];
    if (!row) return;
    row.amount = totalAmount;
    productPayments.value[productIndex] = [row];
    return;
  }

  // 从全款切换为首款：清空金额，保留当前行
  if (val?.toUpperCase() === 'FIRST_PAYMENT') {
    const row = payments[rowIndex];
    if (!row) return;
    row.amount = undefined as any;
  }
};

// ✨ 外发应付款类型变更处理
const handleOutsourcePaymentTypeChange = (productIndex: number, rowIndex: number, val: string) => {
  if (val?.toUpperCase() === 'TOTAL') {
    const payments = outsourcePayments.value[productIndex];
    const row = payments?.[rowIndex];
    if (!row) return;
    const product = props.productionItems?.[productIndex];
    if (!product) return;
    row.amount = calculateOutsourceAmount(product);
    outsourcePayments.value[productIndex] = [row];
  }
};

// ✨ 模具外发应付款类型变更处理
const handleMoldOutsourcePaymentTypeChange = (moldIndex: number, rowIndex: number, val: string) => {
  if (val?.toUpperCase() === 'TOTAL') {
    const payments = moldOutsourcePayments.value[moldIndex];
    const row = payments?.[rowIndex];
    if (!row) return;
    moldOutsourcePayments.value[moldIndex] = [row];
  }
};

// 下一步
const validatePaymentList = (
  listIndex: number,
  source: string,
  itemName: string,
): string | null => {
  const payments = getPaymentListBySource(listIndex, source);
  if (!payments || payments.length === 0) return null;

  const hasFirst = payments.some((p: any) => p.payment_type?.toUpperCase() === 'FIRST_PAYMENT');
  const hasFinal = payments.some((p: any) => p.payment_type?.toUpperCase() === 'FINAL_PAYMENT');
  const hasTotal = payments.some((p: any) => p.payment_type?.toUpperCase() === 'TOTAL');

  // 全款模式：无需更多检查
  if (hasTotal) return null;

  // 有首款则必须有尾款
  if (hasFirst && !hasFinal) {
    return `"${itemName}" 已配置首款，必须配置尾款`;
  }

  // 尾款必须是最后一项
  if (hasFinal) {
    const lastType = payments[payments.length - 1]?.payment_type?.toUpperCase();
    if (lastType !== 'FINAL_PAYMENT') {
      return `"${itemName}" 的尾款必须是最后一笔款项，请调整款项顺序`;
    }
    // 最后一个必须是尾款，且不能有空的款项类型后续
    for (let i = 0; i < payments.length - 1; i++) {
      if (payments[i]?.payment_type?.toUpperCase() === 'FINAL_PAYMENT') {
        return `"${itemName}" 的尾款必须是最后一笔款项，请删除尾款后面的款项`;
      }
    }
  }

  return null;
};

// 下一步
const handleNext = () => {
  for (let i = 0; i < props.moldItems.length; i++) {
    if (!isMoldAmountValid(i)) {
      MessagePlugin.error(`模具"${props.moldItems[i].mold_code}" 金额未配平`);
      return;
    }
    if ((moldPayments.value[i] || []).some((item: any) => !item.payment_type)) {
      MessagePlugin.error(`模具"${props.moldItems[i].mold_code}" 有未选择类型的款项`);
      return;
    }
    const moldErr = validatePaymentList(i, 'mold', `模具 ${props.moldItems[i].mold_code}`);
    if (moldErr) {
      MessagePlugin.error(moldErr);
      return;
    }
  }
  for (let i = 0; i < props.productionItems.length; i++) {
    if (!isProductAmountValid(i)) {
      MessagePlugin.error(`产品"${props.productionItems[i].product_name}" 金额未配平`);
      return;
    }
    if ((productPayments.value[i] || []).some((item: any) => !item.payment_type)) {
      MessagePlugin.error(`产品"${props.productionItems[i].product_name}" 有未选择类型的款项`);
      return;
    }
    const productErr = validatePaymentList(
      i,
      'product',
      `产品 ${props.productionItems[i].product_name}`,
    );
    if (productErr) {
      MessagePlugin.error(productErr);
      return;
    }
  }
  // ✨ 外发应付款不做强制校验，由用户自由填写
  emit('next', {
    mold_payments: moldPayments.value,
    product_payments: productPayments.value,
    outsource_payments: outsourcePayments.value,
    mold_outsource_payments: moldOutsourcePayments.value,
  });
};

defineExpose({
  getData: () => ({
    mold_payments: moldPayments.value,
    product_payments: productPayments.value,
    outsource_payments: outsourcePayments.value,
    mold_outsource_payments: moldOutsourcePayments.value,
  }),
  validate: handleNext,
  // 纯校验方法：返回 { valid, error }，不 emit，不弹 MessagePlugin
  validateOnly: () => {
    for (let i = 0; i < props.moldItems.length; i++) {
      if (!isMoldAmountValid(i)) {
        return { valid: false, error: `模具"${props.moldItems[i].mold_code}" 金额未配平` };
      }
      if ((moldPayments.value[i] || []).some((item: any) => !item.payment_type)) {
        return { valid: false, error: `模具"${props.moldItems[i].mold_code}" 有未选择类型的款项` };
      }
      const moldErr = validatePaymentList(i, 'mold', `模具 ${props.moldItems[i].mold_code}`);
      if (moldErr) {
        return { valid: false, error: moldErr };
      }
    }
    for (let i = 0; i < props.productionItems.length; i++) {
      if (!isProductAmountValid(i)) {
        return { valid: false, error: `产品"${props.productionItems[i].product_name}" 金额未配平` };
      }
      if ((productPayments.value[i] || []).some((item: any) => !item.payment_type)) {
        return {
          valid: false,
          error: `产品"${props.productionItems[i].product_name}" 有未选择类型的款项`,
        };
      }
      const productErr = validatePaymentList(
        i,
        'product',
        `产品 ${props.productionItems[i].product_name}`,
      );
      if (productErr) {
        return { valid: false, error: productErr };
      }
    }
    return { valid: true };
  },
});

const getDefaultDueDate = () =>
  new Date(Date.now() + 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

// ✨ 初始化模具款项（支持预填充）
const initializeMoldPayments = (forceOverwrite = false) => {
  const initialData = props.initialMoldPayments;
  const hasInitialData = initialData && Object.keys(initialData).length > 0;

  props.moldItems.forEach((mold, i) => {
    // 优先使用预填充数据（来自后台的已配置款项）
    if (hasInitialData && initialData![i] && initialData![i].length > 0) {
      moldPayments.value[i] = initialData![i].map((p) => ({
        paymentIndex: Date.now() + i + Math.random(),
        payment_type: p.payment_type,
        amount: p.amount,
        due_date: p.due_date || getDefaultDueDate(),
      }));
      return;
    }
    // 没有预填充数据时，如果已有用户编辑过的数据且非强制覆盖，则跳过
    if (!forceOverwrite && moldPayments.value[i] && moldPayments.value[i].length > 0) {
      return;
    }
    // 创建默认空行
    moldPayments.value[i] = [
      {
        paymentIndex: Date.now() + i,
        payment_type: '',
        amount: undefined as any,
        due_date: getDefaultDueDate(),
      },
    ];
  });
};

// ✨ 初始化产品款项（支持预填充）
const initializeProductPayments = (forceOverwrite = false) => {
  const initialData = props.initialProductPayments;
  const hasInitialData = initialData && Object.keys(initialData).length > 0;

  props.productionItems.forEach((product, i) => {
    // 优先使用预填充数据（来自后台的已配置款项）
    if (hasInitialData && initialData![i] && initialData![i].length > 0) {
      productPayments.value[i] = initialData![i].map((p) => ({
        paymentIndex: Date.now() + i + 1000 + Math.random(),
        payment_type: p.payment_type,
        amount: p.amount,
        due_date: p.due_date || getDefaultDueDate(),
      }));
      return;
    }
    // 没有预填充数据时，如果已有用户编辑过的数据且非强制覆盖，则跳过
    if (!forceOverwrite && productPayments.value[i] && productPayments.value[i].length > 0) {
      return;
    }
    // 创建默认空行
    productPayments.value[i] = [
      {
        paymentIndex: Date.now() + i + 1000,
        payment_type: '',
        amount: undefined as any,
        due_date: getDefaultDueDate(),
      },
    ];
  });
};

// ✨ 初始化面板展开状态（默认为空数组，所有卡片折叠）
const initializePanels = () => {
  // 保持所有面板折叠：不向 active 数组中添加任何值
  moldActivePanels.value = [];
  productActivePanels.value = [];
  outsourceActivePanels.value = [];
  moldOutsourceActivePanels.value = [];
};

// ✨ 监听模具明细变化（只在数组引用变化时重新初始化，避免修改 item 属性导致面板折叠）
watch(
  () => props.moldItems,
  (newItems, oldItems) => {
    if (newItems !== oldItems && newItems.length > 0) {
      initializePanels();
      initializeMoldPayments();
    }
  },
  { immediate: false },
);

// ✨ 监听产品明细变化（只在数组引用变化时重新初始化，避免修改 unit_price/packaging_fee/freight_fee 导致面板折叠）
watch(
  () => props.productionItems,
  (newItems, oldItems) => {
    if (newItems !== oldItems && newItems.length > 0) {
      initializePanels();
      initializeProductPayments();
    }
  },
  { immediate: false },
);

// ✨ 监听预填充数据变化（编辑模式下加载完款项后传入，强制覆盖默认空行）
watch(
  () => props.initialMoldPayments,
  (newVal) => {
    if (newVal && Object.keys(newVal).length > 0) {
      initializeMoldPayments(true);
    }
  },
  { deep: true },
);

// ✨ 监听产品预填充数据变化（编辑模式下加载完款项后传入，强制覆盖默认空行）
watch(
  () => props.initialProductPayments,
  (newVal) => {
    if (newVal && Object.keys(newVal).length > 0) {
      initializeProductPayments(true);
    }
  },
  { deep: true },
);

onMounted(() => {
  initializePanels();

  // 模具款项初始化
  initializeMoldPayments();

  // 产品款项初始化（支持预填充）
  initializeProductPayments();

  // ✨ 外发应付款初始默认行
  if (isOutsourced.value) {
    props.productionItems.forEach((product, i) => {
      if (!outsourcePayments.value[i] || outsourcePayments.value[i].length === 0) {
        outsourcePayments.value[i] = [
          {
            paymentIndex: Date.now() + i + 2000,
            payment_type: '',
            amount: undefined as any,
            due_date: getDefaultDueDate(),
          },
        ];
      }
    });
  }

  // ✨ 模具外发应付款初始默认行
  if (hasMoldOutsource.value) {
    moldOutsourceItems.value.forEach((mold, i) => {
      if (!moldOutsourcePayments.value[i] || moldOutsourcePayments.value[i].length === 0) {
        moldOutsourcePayments.value[i] = [
          {
            paymentIndex: Date.now() + i + 3000,
            payment_type: '',
            amount: undefined as any,
            due_date: getDefaultDueDate(),
          },
        ];
      }
    });
  }
});
</script>
<style scoped lang="less">
.payment-plan-form {
  .form-header {
    .step-title {
      font-size: 16px;
      font-weight: 500;
      color: var(--td-text-color-primary);
      margin-bottom: 24px;
    }
  }

  // 订单汇总概览
  .summary-card {
    margin-bottom: 24px;

    :deep(.t-card__body) {
      padding: 16px 24px;
    }

    .summary-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 14px;

      .summary-title {
        font-size: 16px;
        font-weight: 600;
        color: var(--td-text-color-primary);
      }
    }

    .summary-body {
      display: flex;
      align-items: center;
      gap: 24px;

      .summary-item {
        display: flex;
        align-items: center;
        gap: 10px;

        .summary-label {
          font-size: 13px;
          color: var(--td-text-color-secondary);
        }

        .summary-count {
          font-size: 13px;
          font-weight: 500;

          &.brand {
            color: var(--td-brand-color);
          }

          &.success {
            color: var(--td-success-color);
          }
        }

        .summary-amount {
          font-size: 16px;
          font-weight: 600;
          color: var(--td-text-color-primary);

          &.outsource {
            color: var(--td-warning-color);
          }
        }

        &.primary {
          flex: 1;

          .summary-amount {
            font-size: 20px;
            color: var(--td-brand-color);
          }

          .summary-progress {
            flex: 1;
            max-width: 200px;
            margin-left: auto;

            :deep(.t-progress__inner) {
              height: 8px;
            }
          }
        }
      }

      .summary-divider {
        width: 1px;
        height: 28px;
        background: var(--td-component-border);
      }
    }
  }

  // 分区标题 - TDesign 标准样式
  .section {
    margin-bottom: 24px;

    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 16px;
      background: var(--td-brand-color-1);
      border-radius: 6px 6px 0 0;
      margin-bottom: 0;

      &.product-header {
        background: var(--td-success-color-1);
      }

      &.outsource-header {
        background: var(--td-warning-color-1);
      }

      .header-left {
        display: flex;
        align-items: center;
        gap: 8px;

        .section-icon {
          font-size: 18px;
        }

        .section-title {
          font-size: 16px;
          font-weight: 600;
          color: var(--td-text-color-primary);
        }
      }
    }
  }

  // 折叠面板整体
  :deep(.t-collapse) {
    border: 1px solid var(--td-component-border);
    border-top: none;
    border-radius: 0 0 6px 6px;
  }

  // 面板之间添加分隔线
  :deep(.t-collapse-panel) {
    border-bottom: 1px solid var(--td-component-border);

    &:last-child {
      border-bottom: none;
    }
  }

  // 面板头部悬停效果优化
  :deep(.t-collapse-panel__header) {
    transition: background-color 0.2s;

    &:hover {
      background-color: var(--td-bg-color-secondarycontainer);
    }
  }

  // 面板头部
  .panel-header {
    display: flex;
    align-items: center;
    gap: 12px;
    width: 100%;

    .item-index {
      min-width: 32px;
      text-align: center;
      font-weight: 600;
    }

    .panel-title {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 2px;

      .item-code {
        font-size: 16px;
        font-weight: 600;
        color: var(--td-text-color-primary);
      }

      .item-subtitle {
        font-size: 12px;
        color: var(--td-text-color-secondary);
      }

      .processor-name {
        color: var(--td-warning-color);
        font-weight: 500;
      }
    }

    .panel-amount {
      .amount-total {
        font-size: 15px;
        font-weight: 600;
      }

      .outsource-amount {
        color: var(--td-warning-color);
      }
    }
  }

  // 面板内容
  .panel-body {
    padding: 16px;
    background: var(--td-bg-color-container);

    // 产品金额计算行 - 网格布局
    .calc-row {
      margin-bottom: 16px;
      padding: 0;
      background: var(--td-bg-color-secondarycontainer);
      border-radius: 6px;
      border: 1px solid var(--td-component-border);
      overflow: hidden;

      :deep(.t-input-number) {
        background: var(--td-bg-color-container);
      }

      :deep(.calc-field) {
        .t-input-number {
          width: 100%;
        }

        .t-input-number .t-input {
          width: 100%;
        }

        .t-input__wrap {
          width: 100%;
        }
      }

      .calc-row-header {
        padding: 8px 16px;
        border-bottom: 1px solid var(--td-component-border);
        background: var(--td-bg-color-container);

        .calc-row-title {
          font-size: 13px;
          font-weight: 500;
          color: var(--td-text-color-primary);
        }
      }

      .calc-grid {
        display: grid;
        grid-template-columns: 1.5fr 1fr 1.5fr 1.5fr 1fr 1.5fr;
        gap: 0;

        .calc-field {
          padding: 12px;
          border-right: 1px solid var(--td-component-border);
          min-width: 0;

          &:last-child {
            border-right: none;
          }

          .calc-label {
            display: block;
            font-size: 12px;
            color: var(--td-text-color-secondary);
            margin-bottom: 6px;
          }

          .calc-qty {
            display: flex;
            align-items: center;
            height: 32px;
            padding: 0 12px;
            background: var(--td-bg-color-container);
            border-radius: 4px;
            font-weight: 600;
            font-size: 14px;
          }

          &.calc-result {
            .calc-result-value {
              display: flex;
              align-items: center;
              height: 32px;
              padding: 0 12px;
              background: var(--td-bg-color-container);
              border-radius: 4px;

              .calc-amount {
                font-size: 15px;
                font-weight: 700;
                color: var(--td-warning-color);
              }
            }
          }
        }
      }
    }

    // 表格 - TDesign 官方白色简洁风格
    :deep(.t-table) {
      border-radius: 4px;
      border: 1px solid var(--td-component-border);
      background: #fff;

      th {
        background: #fff;
        font-weight: 500;
        padding: 8px 12px !important;
        border-bottom: 1px solid var(--td-component-border);
      }

      td {
        padding: 8px 12px !important;
        border-bottom: 1px solid var(--td-component-border);
        background: #fff;
      }

      tr:hover td {
        background: var(--td-bg-color-secondarycontainer);
      }

      .t-table__body .t-table__row:last-child td {
        border-bottom: none;
      }

      .t-input,
      .t-select .t-input,
      .t-date-picker .t-input {
        height: 32px;
        font-size: 13px;
      }

      .t-input-number {
        .t-input {
          height: 32px;
        }
      }
    }

    // 工具栏（表格上方）
    .payment-toolbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
    }
  }

  // 操作按钮
  .action-buttons {
    display: flex;
    justify-content: center;
    gap: 16px;
    padding-top: 8px;
  }
}
</style>

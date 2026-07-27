<template>
  <div class="mold-order-form-page">
    <div class="back-button-container">
      <t-button theme="default" variant="text" @click="goBack">
        <template #icon><t-icon name="chevron-left" /></template>
        返回
      </t-button>
    </div>
    <t-loading :loading="loading" size="small" text="加载中...">
      <t-form
        ref="formRef"
        class="mold-order-form"
        :data="formModel"
        :rules="FORM_RULES"
        label-align="top"
        :label-width="100"
        @reset="onReset"
        @submit="onSubmit"
      >
        <div class="form-basic-container">
          <div class="form-basic-item">
            <div class="form-basic-container-title">订单信息</div>

            <!-- 订单基础信息 -->
            <t-row class="row-gap" :gutter="[32, 24]">
              <t-col :span="6">
                <t-form-item label="订单编号">
                  <t-input
                    v-model="formModel.orderCode"
                    placeholder="订单号将在创建后自动生成"
                    disabled
                  />
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="订单类型" name="orderType">
                  <t-select
                    v-model="formModel.orderType"
                    placeholder="请选择订单类型"
                    :disabled="isEdit"
                  >
                    <t-option
                      v-for="type in orderTypeList"
                      :key="type.value"
                      :label="`${type.label}订单`"
                      :value="type.value"
                    />
                  </t-select>
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="优先级">
                  <t-select v-model="formModel.priority" placeholder="请选择优先级">
                    <t-option
                      v-for="item in priorityList"
                      :key="item.value"
                      :label="item.label"
                      :value="item.value"
                    />
                  </t-select>
                </t-form-item>
              </t-col>
              <t-col :span="12">
                <t-form-item label="客户" name="customerId">
                  <t-select v-model="formModel.customerId" placeholder="请选择客户" clearable>
                    <t-option
                      v-for="customer in customerList"
                      :key="customer.id"
                      :label="customer.cus_name"
                      :value="customer.id"
                    />
                  </t-select>
                </t-form-item>
              </t-col>
              <!-- 计划开始日期 -->
              <t-col :span="6">
                <t-form-item label="计划开始日期" name="planStartDate">
                  <t-date-picker
                    v-model="formModel.planStartDate"
                    theme="primary"
                    mode="date"
                    style="width: 100%"
                  />
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="订单交期" name="planEndDate">
                  <t-date-picker
                    v-model="formModel.planEndDate"
                    theme="primary"
                    mode="date"
                    style="width: 100%"
                  />
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="订单总金额">
                  <t-statistic
                    :value="formModel.totalAmount"
                    prefix="¥"
                    unit="元"
                    trend="increase"
                  />
                  <template #help>模具费用总和</template>
                </t-form-item>
              </t-col>
            </t-row>

            <!-- 模具明细 -->
            <div class="form-basic-container-title form-title-gap">模具明细</div>

            <!-- 使用公共模具明细组件 -->
            <mold-items-form
              ref="moldItemsRef"
              v-model="formModel.items"
              v-model:selected-mold-id="formModel.moldId"
              :order-type="formModel.orderType"
              :show-mold-selector="
                formModel.orderType === 'REPAIR' || formModel.orderType === 'MODIFICATION'
              "
              :customer-id="formModel.customerId"
              selector-field-name="moldId"
              :editable="true"
              :allow-drag="formModel.orderType === 'NEW_MOLD'"
              :allow-delete="formModel.orderType === 'NEW_MOLD'"
              :disabled="isEdit"
              @mold-select="handleMoldSelect"
            />

            <!-- 备注 -->
            <div class="form-basic-container-title form-title-gap">其他信息</div>

            <t-form-item label="备注">
              <t-textarea v-model="formModel.remark" :height="124" placeholder="请输入备注" />
            </t-form-item>

            <!-- 款项配置 -->
            <div class="form-basic-container-title form-title-gap">款项配置</div>
            <payment-plan-form
              ref="paymentPlanRef"
              order-type="MOLD_ONLY"
              :mold-items="formModel.items"
              :production-items="[]"
              :initial-mold-payments="initialMoldPayments"
              :hide-step-u-i="true"
              @next="handlePaymentPlanNext"
              @prev="handlePaymentPlanPrev"
            />
          </div>
        </div>

        <div class="form-submit-container">
          <div class="form-submit-sub">
            <div class="form-submit-left">
              <t-button theme="primary" class="form-submit-confirm" type="submit">
                {{ isEdit ? '更新' : '确认' }}
              </t-button>
              <t-button type="reset" class="form-submit-cancel" theme="default" variant="base">
                取消
              </t-button>
            </div>
          </div>
        </div>
      </t-form>
    </t-loading>
  </div>
</template>
<script setup lang="ts">
import type { FormRule, SubmitContext } from 'tdesign-vue-next';
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next';
import { nextTick, onMounted, reactive, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { getReceivablesByOrder } from '@/api/finance/receivable';
import {
  createMoldOrder,
  getMoldList,
  getMoldOrderById,
  getMoldWorkOrderList,
  updateMoldOrder,
} from '@/api/mold';
import { getOrderPayments } from '@/api/mold/payment';
import { getCustomerList } from '@/api/reference';
import MoldItemsForm from '@/components/MoldItemsForm.vue';
import { useDictRaw } from '@/hooks/useDict';
import PaymentPlanForm from '@/pages/order-management/components/PaymentPlanForm.vue';

defineOptions({
  name: 'MoldOrderForm',
});

const route = useRoute();
const router = useRouter();
const formRef = ref();
const paymentPlanRef = ref();
const moldItemsRef = ref();
const isEdit = ref(false);
const loading = ref(false); // 编辑模式下的详情加载中状态
const originalItems = ref<any[]>([]); // ✨ 保存原始数据用于比对
const initialMoldPayments = ref<
  Record<
    number,
    Array<{
      payment_type: string;
      amount: number;
      due_date: string;
    }>
  >
>({}); // ✨ 编辑模式：预填充的模具款项数据

// 客户列表
const customerList = ref<any[]>([]);

// 模具列表（用于修模/改造订单的校验）
const moldList = ref<any[]>([]);

// 防止选模具自动填客户时循环触发
const isAutoFillingCustomer = ref(false);

// 字典数据列表（useDictRaw 自动加载 + 全局缓存）
const { items: orderTypeList } = useDictRaw('mold_order_type', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '')
    .map((item: any) => ({
      value: item.dict_value,
      label: item.dict_label,
      color: item.color || 'default',
    })),
);
const { items: priorityList } = useDictRaw('order-priority', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '')
    .map((item: any) => ({
      value: item.dict_value,
      label: item.dict_label,
      color: item.color || 'default',
    })),
);

// 直接使用语义化字符串，不需要转换函数

// 表单数据
const formModel = reactive({
  id: undefined as number | undefined,
  orderCode: '',
  orderType: 'NEW_MOLD', // 默认新模订单（字符串类型）
  customerId: undefined as number | undefined,
  moldId: undefined as number | undefined, // 修模/改造订单选择的模具ID
  priority: 'NORMAL', // ✨ 优先级，默认为普通
  planStartDate: '',
  planEndDate: '',
  totalAmount: 0,
  status: 'NOT_STARTED', // 语义化字符串状态
  remark: '',
  items: [] as Array<{
    id?: number;
    mold_id?: number;
    mold_code: string;
    mold_name: string;
    amount: number | undefined;
    moldType: number | undefined;
    moldProcess: number | undefined;
    materialType: number | undefined;
    remark: string;
    cavityCount?: number;
    // ✨ 工单相关信息（生产方式）
    workType?: string;
    processorId?: number;
    outsourcePrice?: number;
  }>,
});

const getDefaultFormModel = () => ({
  id: undefined as number | undefined,
  orderCode: '',
  orderType: 'NEW_MOLD',
  customerId: undefined as number | undefined,
  moldId: undefined as number | undefined,
  priority: 'NORMAL',
  planStartDate: '',
  planEndDate: '',
  totalAmount: 0,
  status: 'NOT_STARTED',
  remark: '',
  items: [] as any[],
});

const resetFormState = () => {
  Object.assign(formModel, getDefaultFormModel());
  isEdit.value = false;
  originalItems.value = [];
  initialMoldPayments.value = {};
  formRef.value?.clearValidate?.();
};

// 表单验证规则
const FORM_RULES: Record<string, FormRule[]> = {
  orderType: [{ required: true, message: '请选择订单类型', type: 'error' }],
  customerId: [{ required: true, message: '请选择客户', type: 'error' }],
  planStartDate: [{ required: true, message: '请选择计划开始日期', type: 'error' }],
  planEndDate: [{ required: true, message: '请选择订单交期', type: 'error' }],
};

// 监听订单类型变化
watch(
  () => formModel.orderType,
  (newType) => {
    if (newType === 'NEW_MOLD') {
      // 新模订单：清空模具选择
      formModel.moldId = undefined;
    } else {
      // 修模/改造订单：清空手动输入的明细，加载模具列表
      formModel.items = [];
      formModel.moldId = undefined;
      loadMoldList();
    }
  },
);

// 监听客户变化（用户手动切换客户时，清空已有模具选择）
watch(
  () => formModel.customerId,
  (newVal, oldVal) => {
    // 跳过选模具自动填客户的场景
    if (isAutoFillingCustomer.value) {
      isAutoFillingCustomer.value = false;
      return;
    }
    // 用户手动切换客户时，清空模具选择（组件 watch customerId 会重新加载列表）
    if (newVal !== oldVal && formModel.moldId) {
      formModel.moldId = undefined;
      formModel.items = [];
    }
  },
);

// 监听模具明细金额变化，自动计算订单总金额
watch(
  () => formModel.items,
  (items) => {
    formModel.totalAmount = items.reduce((sum, item) => sum + (item.amount || 0), 0);
  },
  { deep: true },
);

// 选择模具后由公共组件处理，这里保留用于双向绑定客户
const handleMoldSelect = (value: any) => {
  const moldId = value as number | undefined;
  if (!moldId) {
    return;
  }

  const selectedMold = moldList.value.find((m: any) => m.id === moldId);
  if (selectedMold) {
    // 双向绑定：选模具自动带出客户
    if (!formModel.customerId && selectedMold.customer_id) {
      isAutoFillingCustomer.value = true;
      formModel.customerId = selectedMold.customer_id;
    }
  }
};

// ✨ 获取款项配置数据，转换为后端 PaymentPlanDTO 格式
const getPaymentPlanData = () => {
  if (paymentPlanRef.value) {
    const raw = paymentPlanRef.value.getData(); // { mold_payments, product_payments, outsource_payments, mold_outsource_payments }
    if (!raw) return null;

    const { mold_payments, mold_outsource_payments } = raw;
    const items: Array<{
      index: number;
      payment_type: string;
      amount_type: string;
      amount_value: number;
      calculated_amount: number;
      due_date: string;
      source_type: string;
    }> = [];
    let index = 0;

    for (const key of Object.keys(mold_payments || {})) {
      for (const payment of mold_payments[key] || []) {
        if (payment.payment_type && (payment.amount || 0) > 0) {
          items.push({
            index: index++,
            payment_type: payment.payment_type,
            amount_type: 'amount',
            amount_value: payment.amount || 0,
            calculated_amount: payment.amount || 0,
            due_date: payment.due_date || '',
            source_type: 'mold',
          });
        }
      }
    }

    // ✨ 转换模具外发应付款（应付）
    for (const key of Object.keys(mold_outsource_payments || {})) {
      for (const payment of mold_outsource_payments[key] || []) {
        if (payment.payment_type && (payment.amount || 0) > 0) {
          items.push({
            index: index++,
            payment_type: payment.payment_type,
            amount_type: 'amount',
            amount_value: payment.amount || 0,
            calculated_amount: payment.amount || 0,
            due_date: payment.due_date || '',
            source_type: 'outsource', // 模具外发应付
          });
        }
      }
    }

    return {
      enabled: items.length > 0,
      items,
    };
  }
  return null;
};

// ✨ 款项配置下一步
const handlePaymentPlanNext = async () => {
  // 在模具订单中，款项配置后直接提交订单
  // 这里不需要做任何事情，因为提交按钮会调用 getPaymentPlanData()
};

// ✨ 款项配置上一步
const handlePaymentPlanPrev = () => {
  // 在单页表单中，上一步不需要处理
};

const goBack = () => {
  router.back();
};

// 提交表单
const onSubmit = async (ctx: SubmitContext) => {
  if (ctx.validateResult === true) {
    try {
      if (!formModel.planStartDate) {
        MessagePlugin.warning('请选择计划开始日期');
        return;
      }
      if (!formModel.planEndDate) {
        MessagePlugin.warning('请选择订单交期');
        return;
      }

      // 模具明细校验（复用 MoldItemsForm 组件统一校验）
      if (formModel.orderType === 'NEW_MOLD') {
        const result = moldItemsRef.value?.validateItems();
        if (!result?.valid) {
          MessagePlugin.warning(result?.errors?.[0] || '请完善模具明细');
          return;
        }
      }

      if (formModel.orderType === 'REPAIR' || formModel.orderType === 'MODIFICATION') {
        if (!formModel.moldId) {
          MessagePlugin.warning('请选择模具');
          return;
        }
        const result = moldItemsRef.value?.validateItems();
        if (!result?.valid) {
          MessagePlugin.warning(result?.errors?.[0] || '请完善模具明细');
          return;
        }
      }

      // ✨ 检查是否有生产方式变更，需要用户确认
      if (isEdit.value && originalItems.value.length > 0) {
        const changes: string[] = [];
        let hasOutsourceToInHouse = false; // 是否有外发转本厂
        let hasInHouseToOutsource = false; // 是否有本厂转外发

        formModel.items.forEach((item, index) => {
          const originalItem = originalItems.value[index];
          if (originalItem) {
            if (item.workType !== originalItem.workType) {
              const oldType = originalItem.workType === 'OUTSOURCED' ? '外发' : '本厂';
              const newType = item.workType === 'OUTSOURCED' ? '外发' : '本厂';
              changes.push(`明细 ${index + 1} (${item.mold_code}): ${oldType} → ${newType}`);

              // 记录变更方向
              if (originalItem.workType === 'OUTSOURCED' && item.workType === 'IN_HOUSE') {
                hasOutsourceToInHouse = true;
              } else if (originalItem.workType === 'IN_HOUSE' && item.workType === 'OUTSOURCED') {
                hasInHouseToOutsource = true;
              }
            }
          }
        });

        if (changes.length > 0) {
          // 根据不同的变更方向给出不同的提示
          let message = `检测到以下生产方式变更：\n\n${changes.join('\n')}\n\n`;

          if (hasOutsourceToInHouse && !hasInHouseToOutsource) {
            message += '系统将冻结旧的外发记录。是否继续？';
          } else if (hasInHouseToOutsource && !hasOutsourceToInHouse) {
            message += '系统将创建新的外发记录。是否继续？';
          } else {
            message += '系统将自动处理外发记录的变更。是否继续？';
          }

          const confirmResult = await DialogPlugin.confirm({
            header: '生产方式变更确认',
            body: message,
            theme: 'warning',
            cancelBtn: '取消',
            confirmBtn: '确认',
          });

          // DialogPlugin.confirm 返回 Promise，resolve 表示确认，reject 表示取消
          if (!confirmResult) {
            return; // 用户取消操作
          }
        }
      }

      const submitData: any = {
        order_type: formModel.orderType,
        customer_id: formModel.customerId,
        priority: formModel.priority, // ✨ 添加优先级
        total_amount: formModel.totalAmount,
        status: formModel.status,
        remark: formModel.remark,
      };

      if (formModel.planStartDate) {
        submitData.plan_start_date = formModel.planStartDate;
      }
      if (formModel.planEndDate) {
        submitData.plan_end_date = formModel.planEndDate;
      }

      // 新增模式下添加模具明细或模具ID
      if (!isEdit.value) {
        // 修模/改造订单：添加模具ID
        if (formModel.orderType === 'REPAIR' || formModel.orderType === 'MODIFICATION') {
          if (formModel.moldId) {
            try {
              const selectedMold = moldList.value.find((m: any) => m.id === formModel.moldId);
              if (!selectedMold) {
                MessagePlugin.error('未找到模具信息');
                return;
              }
              const workOrders = await getMoldWorkOrderList({
                moldCode: selectedMold.mold_code,
                page: 1,
                page_size: 100,
              });
              const unfinishedNewMoldWorkOrder = (workOrders.list || []).find((wo: any) => {
                const status = Number(wo.status);
                const orderType = wo.order_type;
                return orderType === 'NEW_MOLD' && status < 3;
              });
              if (unfinishedNewMoldWorkOrder) {
                MessagePlugin.error(
                  '该模具的新模制造工单尚未完成，无法创建修模订单。请等待模具制造完成后再创建修模订单',
                );
                return;
              }
            } catch (error) {
              console.error('检查模具工单失败:', error);
            }
            submitData.mold_id = formModel.moldId;
          }
        }
      }

      // ✨ 所有订单类型（包括编辑模式）都添加模具明细，用于更新工单信息
      submitData.items = formModel.items.map((item) => ({
        id: item.id || undefined,
        mold_id: item.mold_id || undefined,
        mold_code: item.mold_code,
        mold_name: item.mold_name,
        amount: item.amount,
        mold_type: item.moldType,
        material_type: item.materialType,
        mold_process: item.moldProcess,
        cavity_count: item.cavityCount,
        remark: item.remark,
        // ✨ 工单相关信息（生产方式）
        work_type: item.workType || 'IN_HOUSE',
        processor_id: item.workType === 'OUTSOURCED' ? item.processorId : undefined,
        outsource_price: item.workType === 'OUTSOURCED' ? item.outsourcePrice : undefined,
      }));

      // ✨ 校验并添加款项配置（复用 PaymentPlanForm 组件统一校验）
      if (!isEdit.value && paymentPlanRef.value) {
        const paymentPlan = getPaymentPlanData();
        const validResult = paymentPlanRef.value?.validateOnly();
        if (validResult && !validResult.valid) {
          MessagePlugin.warning(validResult.error || '请完善款项配置');
          return;
        }
        submitData.payment_plan = paymentPlan;
      }

      if (isEdit.value) {
        submitData.id = formModel.id;
        await updateMoldOrder(submitData);
        MessagePlugin.success('更新成功');
      } else {
        await createMoldOrder(submitData);
        MessagePlugin.success('创建成功');

        // 新增成功后立即重置表单，防止 KeepAlive 缓存导致下次进入时显示旧数据
        Object.assign(formModel, {
          id: undefined,
          orderCode: '',
          orderType: 'NEW_MOLD',
          customerId: undefined,
          moldId: undefined,
          priority: 'NORMAL',
          planStartDate: '',
          planEndDate: '',
          totalAmount: 0,
          status: 'NOT_STARTED',
          remark: '',
          items: [],
        });
      }

      router.push('/mold/order');
    } catch (error: any) {
      console.error('提交订单失败:', error);
      if (error?.message) {
        MessagePlugin.error(error.message);
      } else {
        MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
      }
    }
  }
};

const onReset = () => {
  MessagePlugin.warning('取消操作');
  router.push('/mold/order');
};

const loadOrderDetail = async (id: number) => {
  loading.value = true;
  try {
    const res = await getMoldOrderById(id);
    if (res) {
      formModel.id = res.id;
      formModel.orderCode = res.order_code;
      formModel.orderType = res.order_type || 'NEW_MOLD';
      formModel.customerId = res.customer_id;
      formModel.priority = res.priority || 'NORMAL'; // ✨ 加载优先级
      formModel.planStartDate = res.plan_start_date;
      formModel.planEndDate = res.plan_end_date;
      formModel.totalAmount = res.total_amount || 0; // 确保订单总金额正确加载
      formModel.status = res.status;
      formModel.remark = res.remark;


      if (res.items && Array.isArray(res.items) && res.items.length > 0) {
        formModel.items = res.items.map((item: any) => {
          // 如果 Mold 关联数据存在，从 Mold 中获取信息
          const mold = item.mold || {};
          return {
            id: item.id,
            mold_id: item.mold_id,
            mold_code: item.mold_code || mold.mold_code || '',
            mold_name: item.mold_name || mold.mold_name || '',
            amount: item.amount,
            moldType: item.mold_type || mold.mold_type,
            materialType: item.material_type || mold.material_type,
            moldProcess: item.mold_process || mold.mold_process,
            cavityCount: item.cavity_count || mold.cavity_count || undefined,
            remark: item.remark || '',
            // ✨ 加载工单信息（生产方式）- 从 item 的直接字段读取
            workType: item.work_type || 'IN_HOUSE',
            processorId: item.processor_id || undefined,
            outsourcePrice: item.outsource_price || undefined,
          };
        });

        // ✨ 保存原始数据用于后续比对
        originalItems.value = JSON.parse(JSON.stringify(formModel.items));
      } else {
        console.warn('[loadOrderDetail] 订单没有明细数据，items:', res.items);
        formModel.items = [];
        return;
      }

      // ✨ 加载应收款项数据，用于预填充款项配置
      // 策略：多数据源依次回退
      //   1️⃣ accounts_receivable 表（PaymentPlan 创建时写入）
      //   2️⃣ mold_order_payment 表（工单生成时自动创建）
      //   3️⃣ master_order 的 accounts_receivable（综合订单模式）
      try {
        let receivables: any[] = [];

        // 1. 优先从应收账款表查询（创建订单时通过 PaymentPlan 写入的数据）
        const arList = await getReceivablesByOrder(id, 'mold_order');

        if (arList && Array.isArray(arList) && arList.length > 0) {
          receivables = arList;
        } else {
          // 2. 回退：从模具订单款项表查询（工单生成时自动创建的默认款项）
          const paymentList = await getOrderPayments(id);
          if (paymentList && Array.isArray(paymentList) && paymentList.length > 0) {
            receivables = paymentList.map((p: any) => ({
              payment_type: p.payment_type || '',
              amount: p.should_amount || p.contract_amount || 0,
              due_date: p.payment_date
                ? typeof p.payment_date === 'string'
                  ? p.payment_date.split('T')[0]
                  : p.payment_date
                : '',
            }));
          } else if (res.master_order_id) {
            // 3. 回退：从综合订单的应收账款查询
            try {
              const masterARList = await getReceivablesByOrder(res.master_order_id, 'master_order');
              if (masterARList && Array.isArray(masterARList) && masterARList.length > 0) {
                // 只取属于该模具订单的应收款
                receivables = masterARList
                  .filter((ar: any) => {
                    return ar.source_type === 'mold' || !ar.source_type;
                  })
                  .map((ar: any) => ({
                    payment_type: ar.payment_type || '',
                    amount: ar.amount || 0,
                    due_date: ar.due_date
                      ? typeof ar.due_date === 'string'
                        ? ar.due_date.split('T')[0]
                        : ar.due_date
                      : '',
                  }));
              }
            } catch (e) {
              console.warn('[loadOrderDetail] 查询综合订单应收款失败:', e);
            }
          }
        }

        // 只要有应收款数据就进行预填充
        if (receivables && receivables.length > 0) {
          const moldPaymentsMap: Record<
            number,
            Array<{
              payment_type: string;
              amount: number;
              due_date: string;
            }>
          > = {};

          // 将所有应收款作为第一个模具的期数计划（模具订单通常只有一个模具）
          moldPaymentsMap[0] = receivables.map((ar: any) => ({
            payment_type: ar.payment_type || '',
            amount: ar.amount || 0,
            due_date: ar.due_date || '',
          }));

          initialMoldPayments.value = moldPaymentsMap;
        } else {
          initialMoldPayments.value = {};
          console.warn(
            '[loadOrderDetail] ⚠️ 所有数据源均为空（accounts_receivable/mold_order_payment/master_order），无法预填充款项',
          );
        }
      } catch (payError) {
        console.error('[loadOrderDetail] 加载应收款项数据失败:', payError);
        initialMoldPayments.value = {};
      }

      // 如果是修模/改造订单，加载模具列表
      if (formModel.orderType === 'REPAIR' || formModel.orderType === 'MODIFICATION') {
        await loadMoldList();
      }
    }
  } catch (error) {
    console.error('[loadOrderDetail] 加载订单详情失败:', error);
    MessagePlugin.error('加载订单详情失败');
  } finally {
    loading.value = false;
  }
};

// 加载模具列表（用于修模/改造订单）
const loadMoldList = async () => {
  try {
    const res = await getMoldList({ page: 1, page_size: 100 });
    moldList.value = res.list || [];
  } catch (error) {
    console.error('获取模具列表失败:', error);
  }
};

const loadCustomerList = async () => {
  try {
    const res = await getCustomerList();
    customerList.value = res.list || [];
  } catch (error) {
    console.error('获取客户列表失败:', error);
  }
};

// 加载字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

const initFormMode = async (idParam: string | undefined) => {
  // ✨ 先开启 loading 避免数据填充时的闪烁
  loading.value = true;
  // ✨ 始终先重置表单状态，防止编辑数据残留到新增模式
  resetFormState();

  if (idParam) {
    isEdit.value = true;
    await loadOrderDetail(Number(idParam));
    // 编辑模式：动态更新页面标题
    route.meta.title = { zh_CN: '编辑模具订单', en_US: 'Edit Mold Order' };
    loading.value = false;
    await nextTick();
    formRef.value?.clearValidate?.();
  } else {
    // 新增模式：恢复页面标题
    route.meta.title = { zh_CN: '新增模具订单', en_US: 'New Mold Order' };
    loading.value = false;
  }
};

onMounted(async () => {
  await loadCustomerList();
  // 字典已通过 useDictRaw 自动加载

  const idParam = route.query.id;
  await initFormMode(typeof idParam === 'string' ? idParam : undefined);
});

// ✨ 监听路由 query.id 变化，处理组件复用时从编辑切换到新增的场景
watch(
  () => route.query.id,
  async (newId, oldId) => {
    if (newId !== oldId) {
      await initFormMode(typeof newId === 'string' ? newId : undefined);
    }
  },
);
</script>
<style lang="less" scoped>
@import '@/pages/form/base/index.less';

.back-button-container {
  margin-bottom: 16px;
}
</style>

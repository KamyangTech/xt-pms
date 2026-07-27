<template>
  <div class="step-form">
    <div class="step-title">填写订单基础信息</div>
    <t-form
      ref="formRef"
      :data="formData"
      :rules="FORM_RULES"
      label-align="top"
      @reset="onReset"
      @submit="onSubmit"
    >
      <t-row :gutter="[32, 24]">
        <!-- 客户单独一行 -->
        <t-col :span="6">
          <t-form-item label="客户" name="customer_id">
            <t-select
              v-model="formData.customer_id"
              placeholder="请选择客户"
              clearable
              filterable
            >
              <t-option
                v-for="customer in customerList"
                :key="customer.id"
                :value="customer.id"
                :label="customer.cus_name"
              />
            </t-select>
          </t-form-item>
        </t-col>

        <!-- 两个日期并排 -->
        <t-col :span="3">
          <t-form-item label="计划开始日期" name="plan_start_date">
            <t-date-picker
              v-model="formData.plan_start_date"
              placeholder="请选择计划开始日期"
              clearable
            />
          </t-form-item>
        </t-col>

        <t-col :span="3">
          <t-form-item label="计划结束日期" name="plan_end_date">
            <t-date-picker
              v-model="formData.plan_end_date"
              placeholder="请选择计划结束日期"
              clearable
            />
          </t-form-item>
        </t-col>

        <!-- 两个备注并排 -->
        <t-col :span="6">
          <t-form-item label="订单备注" name="remark">
            <t-textarea
              v-model="formData.remark"
              placeholder="请输入订单备注（主订单）"
              :autosize="{ minRows: 2, maxRows: 4 }"
            />
          </t-form-item>
        </t-col>

        <t-col :span="6">
          <t-form-item label="模具订单备注" name="mold_order_remark">
            <t-textarea
              v-model="formData.mold_order_remark"
              placeholder="请输入模具订单备注（可选）"
              :autosize="{ minRows: 2, maxRows: 4 }"
            />
          </t-form-item>
        </t-col>
      </t-row>

      <!-- 按钮区域 -->
      <t-row class="form-actions">
        <t-col :span="24">
          <t-form-item>
            <t-button type="reset" theme="default" variant="base">上一步</t-button>
            <t-button theme="primary" type="submit">下一步</t-button>
          </t-form-item>
        </t-col>
      </t-row>
    </t-form>
  </div>
</template>

<script setup lang="ts">
import type { FormRule, SubmitContext } from 'tdesign-vue-next';
import { onMounted, reactive, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { getCustomerList } from '@/api/reference';

const emit = defineEmits<{
  (e: 'next', data: any): void
  (e: 'prev'): void
}>();

const formRef = ref();

// 客户列表
const customerList = ref<any[]>([]);

// 表单数据
const formData = reactive({
  customer_id: undefined as number | undefined,
  plan_start_date: '',
  plan_end_date: '',
  remark: '',
  mold_order_remark: '', // 模具订单备注
});

// 今天的日期字符串（YYYY-MM-DD）
const todayStr = (() => {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
})();

// 表单验证规则
const FORM_RULES: Record<string, FormRule[]> = {
  customer_id: [{ required: true, message: '请选择客户', type: 'error' }],
  plan_start_date: [{ required: true, message: '请选择计划开始日期', type: 'error' }],
  plan_end_date: [
    { required: true, message: '请选择计划结束日期', type: 'error' },
    {
      validator: (val: string) => {
        if (!val || !formData.plan_start_date) return true;
        // 规则1：结束日期不能早于开始日期
        if (val < formData.plan_start_date) {
          return { result: false, message: '计划结束日期不能早于计划开始日期', type: 'error' };
        }
        // 规则2：如果开始日期为历史时间，结束日期最早只能是今天
        if (formData.plan_start_date < todayStr && val < todayStr) {
          return { result: false, message: '计划结束日期不能早于今天', type: 'error' };
        }
        return true;
      },
      trigger: 'blur',
    },
  ],
};

const loadCustomers = async () => {
  try {
    const res = await getCustomerList({ page_size: 1000 });
    customerList.value = res.list || [];
  } catch (error) {
    console.error('加载客户列表失败:', error);
    MessagePlugin.error('加载客户列表失败');
  }
};

// 提交表单
const onSubmit = async (ctx: SubmitContext) => {
  if (ctx.validateResult === true) {
    try {
      // 触发下一步事件
      emit('next', { ...formData });
    } catch (error: any) {
      MessagePlugin.error(error.message || '提交失败');
    }
  }
};

const onReset = () => {
  emit('prev');
};

onMounted(() => {
  loadCustomers();
});

defineExpose({
  validate: async () => {
    return await formRef.value?.validate();
  },
  getData: () => ({ ...formData }),
});
</script>

<style scoped lang="less">
.step-form {
  .step-title {
    font-size: 16px;
    font-weight: 500;
    color: var(--td-text-color-primary);
    margin-bottom: 24px;
  }

  .form-actions {
    margin-top: 32px;
  }
}
</style>

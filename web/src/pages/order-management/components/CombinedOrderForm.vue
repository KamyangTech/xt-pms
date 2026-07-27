<template>
  <div class="combined-order-form">
    <t-divider>基础信息</t-divider>
    <t-form :data="formData" label-width="100px">
      <t-form-item label="客户名称" name="customer_id">
        <t-select v-model="formData.customer_id" placeholder="请选择客户" />
      </t-form-item>
      <t-form-item label="订单备注" name="remark">
        <t-textarea v-model="formData.remark" />
      </t-form-item>
    </t-form>

    <t-divider>模具信息</t-divider>
    <!-- 这里可以复用或简化现有的模具订单表单逻辑 -->
    <t-alert theme="info" message="请填写新模具的相关信息，系统将自动创建模具订单。" />

    <t-divider>生产信息</t-divider>
    <!-- 这里可以复用或简化现有的生产订单表单逻辑 -->
    <t-alert theme="info" message="请填写产品信息及生产要求，该订单将等待模具完工后启动。" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const formData = ref({
  customer_id: undefined,
  remark: '',
});

defineExpose({
  validate: async () => {
    // TODO: 实现验证逻辑
    return true;
  },
  submit: async () => {
    // TODO: 实现提交逻辑
  },
});
</script>

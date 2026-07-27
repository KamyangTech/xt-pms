<template>
  <div class="step-content">
    <div class="step-title">请选择对应的订单类型</div>
    <t-radio-group :model-value="orderType" @update:model-value="$emit('update:orderType', $event)">
      <t-radio-button value="COMBINED">综合下单 (模具+产品)</t-radio-button>
      <t-radio-button value="MOLD_ONLY">仅制造模具</t-radio-button>
      <t-radio-button value="PROD_ONLY">仅生产产品</t-radio-button>
    </t-radio-group>
    
    <!-- 模式提示框 -->
    <t-alert v-if="orderType === 'COMBINED'" theme="info" message="同时创建模具订单和生产订单，适用于需要定制模具并生产产品的场景。" style="margin-top: 16px" />
    <t-alert v-else-if="orderType === 'MOLD_ONLY'" theme="info" message="仅创建模具订单，适用于只需要制造模具的场景。" style="margin-top: 16px" />
    <t-alert v-else theme="info" message="仅创建生产订单，适用于使用已有模具进行产品生产的场景。" style="margin-top: 16px" />
    
    <div class="action-bar">
      <t-button theme="primary" @click="$emit('next')">下一步</t-button>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  orderType: string;
}>();

defineEmits<{
  (e: 'update:orderType', value: string): void;
  (e: 'next'): void;
}>();
</script>

<style scoped>
.step-content {
  min-height: 400px;
}

.step-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  margin-bottom: 16px;
}

.action-bar {
  margin-top: 24px;
  text-align: left;
}

:deep(.t-radio-group) {
  display: inline-block;
}
</style>

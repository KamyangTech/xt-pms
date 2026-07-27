<template>
  <div class="step-content success-step">
    <t-result status="success" title="下单成功" tip="您可以前往订单中心查看进度">
      <template #operations>
        <t-button theme="primary" @click="$emit('viewOrders')">查看订单</t-button>
        <t-button variant="base" theme="default" @click="$emit('createAgain')">再次创建</t-button>
      </template>
    </t-result>
  </div>
</template>

<script setup lang="ts">
defineEmits<{
  (e: 'viewOrders'): void;
  (e: 'createAgain'): void;
}>();
</script>

<style scoped>
.step-content {
  min-height: 400px;
}

.success-step {
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>

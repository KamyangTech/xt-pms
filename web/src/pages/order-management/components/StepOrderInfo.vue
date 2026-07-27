<template>
  <div class="step-content">
    <OrderInfoForm
      ref="orderInfoFormRef"
      @next="$emit('next', $event)"
      @prev="$emit('prev')"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import OrderInfoForm from './OrderInfoForm.vue';

const orderInfoFormRef = ref<any>(null);

defineEmits<{
  (e: 'next', data: any): void;
  (e: 'prev'): void;
}>();

defineExpose({
  validate: async () => {
    return await orderInfoFormRef.value?.validate();
  },
  getData: () => orderInfoFormRef.value?.getData(),
});
</script>

<style scoped>
.step-content {
  min-height: 400px;
}
</style>

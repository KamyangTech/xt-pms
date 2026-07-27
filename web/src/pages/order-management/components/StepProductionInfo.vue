<template>
  <div class="step-content">
    <ProductionInfoForm
      ref="productionInfoFormRef"
      :order-type="orderType"
      :prefilled-customer-id="prefilledCustomerId"
      :prefilled-mold-items="prefilledMoldItems"
      @next="$emit('next', $event)"
      @prev="$emit('prev')"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import ProductionInfoForm from './ProductionInfoForm.vue';

const productionInfoFormRef = ref<any>(null);

defineProps<{
  orderType: string;
  prefilledCustomerId?: number;
  prefilledMoldItems?: any[];
}>();

defineEmits<{
  (e: 'next', data: any): void;
  (e: 'prev'): void;
}>();

defineExpose({
  validate: async () => {
    return await productionInfoFormRef.value?.validate();
  },
  getData: () => productionInfoFormRef.value?.getData(),
});
</script>

<style scoped>
.step-content {
  min-height: 400px;
}
</style>

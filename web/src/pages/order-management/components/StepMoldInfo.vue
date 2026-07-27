<template>
  <div class="step-content">
    <MoldInfoForm
      ref="moldInfoFormRef"
      :order-type="orderType"
      @next="$emit('next', $event)"
      @prev="$emit('prev')"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import MoldInfoForm from './MoldInfoForm.vue';

const moldInfoFormRef = ref<any>(null);

defineProps<{
  orderType: string;
}>();

defineEmits<{
  (e: 'next', data: any): void;
  (e: 'prev'): void;
}>();

defineExpose({
  validate: async () => {
    return await moldInfoFormRef.value?.validate();
  },
  getData: () => moldInfoFormRef.value?.getData(),
});
</script>

<style scoped>
.step-content {
  min-height: 400px;
}
</style>

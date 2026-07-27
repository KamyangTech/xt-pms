<template>
  <div :class="`${prefix}-footer`">© {{ new Date().getFullYear() }} {{ companyName || 'Kamyang' }}</div>
</template>
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { prefix } from '@/config/global';
import { getCompanyInfo } from '@/api/company-management';

const companyName = ref('');

onMounted(async () => {
  try {
    const res = await getCompanyInfo();
    if (res) {
      companyName.value = res.company_name || '';
    }
  } catch {
    // 获取失败时使用默认值（skip403Redirect 已在 API 层配置，不会触发 403 跳转循环）
  }
});
</script>
<style lang="less" scoped>
.@{starter-prefix}-footer {
  color: var(--td-text-color-placeholder);
  line-height: 20px;
  text-align: center;
}
</style>

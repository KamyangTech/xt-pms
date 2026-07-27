<template>
  <t-config-provider :global-config="getComponentsLocale">
    <router-view :class="[mode]" />
  </t-config-provider>
</template>
<script setup lang="ts">
import { computed, onMounted } from 'vue';

import { useLocale } from '@/locales/useLocale';
import { useSettingStore } from '@/store';
import { getSystemInfo } from '@/api/system/settings';

const store = useSettingStore();

const mode = computed(() => {
  return store.displayMode;
});

const { getComponentsLocale, locale } = useLocale();

onMounted(async () => {
  try {
    const res = await getSystemInfo();
    if (res.system_name) {
      document.title = res.system_name;
      localStorage.setItem('pms_system_name', res.system_name);
    }
    if (res.favicon_url) {
      const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
      if (link) link.href = `${res.favicon_url}?v=${Date.now()}`;
    }
  } catch {
    // 使用 index.html 默认
  }
});
</script>
<style lang="less" scoped>
#nprogress .bar {
  background: var(--td-brand-color) !important;
}
</style>

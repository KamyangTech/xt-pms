<template>
  <t-breadcrumb max-item-width="150" class="tdesign-breadcrumb">
    <t-breadcrumbItem v-for="item in crumbs" :key="item.to" :to="item.to">
      {{ item.title }}
    </t-breadcrumbItem>
  </t-breadcrumb>
</template>
<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';

import { useLocale } from '@/locales/useLocale';
import type { RouteMeta } from '@/types/interface';

const { locale } = useLocale();
const route = useRoute();

const crumbs = computed(() => {
  return route.matched
    .filter((routeItem) => !routeItem.meta.hiddenBreadcrumb)
    .map((routeItem) => {
      const meta = routeItem.meta as RouteMeta;
      let title = routeItem.path;
      if (meta.title) {
        if (typeof meta.title === 'string') {
          title = meta.title;
        } else {
          title = meta.title[locale.value] || routeItem.path;
        }
      }
      return {
        path: routeItem.path,
        to: routeItem.path,
        title,
      };
    });
});
</script>
<style scoped>
.tdesign-breadcrumb {
  margin-bottom: 24px;
}
</style>

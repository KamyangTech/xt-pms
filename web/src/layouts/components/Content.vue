<template>
  <div v-if="!isRefreshing">
    <router-view v-if="!isFramePage" v-slot="{ Component }">
      <NoPermissionContent v-if="!isAuthorized" />
      <KeepAlive v-else :include="aliveViews">
        <component :is="Component" :key="$route.name" />
      </KeepAlive>
    </router-view>
    <frame-page v-else />
  </div>

  <t-loading v-else />
</template>
<script setup lang="ts">
import isBoolean from 'lodash/isBoolean';
import isUndefined from 'lodash/isUndefined';
import type { ComputedRef } from 'vue';
import { computed } from 'vue';
import { useRoute } from 'vue-router';

import FramePage from '@/layouts/frame/index.vue';
import NoPermissionContent from '@/pages/result/no-permission/NoPermissionContent.vue';
import { useTabsRouterStore, useUserStore } from '@/store';

// 检查当前路由是否有权限
const route = useRoute();
const userStore = useUserStore();
const isAuthorized = computed(() => {
  const routePermission = route.meta?.permission as string | undefined;
  if (!routePermission) return true;
  return userStore.perms.includes(routePermission);
});

const aliveViews = computed(() => {
  const tabsRouterStore = useTabsRouterStore();
  const { tabRouters } = tabsRouterStore;
  return tabRouters
    .filter((route) => {
      const keepAliveConfig = route.meta?.keepAlive;
      const isRouteKeepAlive = isUndefined(keepAliveConfig) || (isBoolean(keepAliveConfig) && keepAliveConfig); // 默认开启 keepalive
      return route.isAlive && isRouteKeepAlive;
    })
    .map((route) => route.name);
}) as ComputedRef<string[]>;

const isRefreshing = computed(() => {
  const tabsRouterStore = useTabsRouterStore();
  const { refreshing } = tabsRouterStore;
  return refreshing;
});

const isFramePage = computed(() => {
  return !!route.meta?.frameSrc;
});
</script>
<style lang="less" scoped>
</style>

<template>
  <div class="server-error-wrapper">
    <div class="server-error-container">
      <div class="server-error-bg">
        <component :is="ServerErrorIcon" />
      </div>
      <div class="server-error-title">500 Internal Server Error</div>
      <div class="server-error-tip">{{ t('pages.result.500.subtitle') }}</div>
      <t-button @click="handleRetry">{{ t('pages.result.networkError.reload') }}</t-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router';
import Result500Icon from '@/assets/assets-result-500.svg?component';
import { t } from '@/locales';

const route = useRoute();
const ServerErrorIcon = Result500Icon;

defineOptions({
  name: 'ServerError',
});

function handleRetry() {
  // 跳回出错前的页面，路由守卫会重新尝试连接后端
  const redirect = route.query.redirect as string || '/';
  window.location.href = redirect;
}
</script>

<style lang="less" scoped>
.server-error {
  &-wrapper {
    height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--td-bg-color-page);
  }
  &-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 400px;
  }
  &-bg {
    width: 200px;
    color: var(--td-brand-color);
  }
  &-title {
    font: var(--td-font-title-large);
    font-style: normal;
    margin-top: var(--td-comp-margin-l);
    color: var(--td-text-color-primary);
  }
  &-tip {
    margin: var(--td-comp-margin-s) 0 var(--td-comp-margin-xxxl);
    font: var(--td-font-body-medium);
    color: var(--td-text-color-secondary);
    text-align: center;
    max-width: 400px;
    line-height: 1.6;
  }
}
</style>

<template>
  <div class="no-permission-wrapper">
    <div class="no-permission-container">
      <div class="no-permission-bg">
        <component :is="Result403Icon" />
      </div>
      <div class="no-permission-title">403 Forbidden</div>
      <div class="no-permission-tip">{{ t('pages.result.noPermission.tip') }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Result403Icon from '@/assets/assets-result-403.svg?component';
import { t } from '@/locales';

defineOptions({
  name: 'ResultNoPermission',
});
</script>

<style lang="less" scoped>
.no-permission {
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

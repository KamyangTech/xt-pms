<template>
  <div class="page-loading-container">
    <t-loading />
  </div>
</template>

<script setup lang="ts">
defineOptions({
  name: 'PageLoading',
});
</script>

<style lang="less" scoped>
.page-loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  width: 100%;
}
</style>

<template>
  <div class="login-wrapper">
    <login-header />

    <div class="login-container">
      <div class="title-container">
        <h1 class="title margin-no">{{ t('pages.login.loginTitle') }}</h1>
        <h1 class="title">生产管理系统</h1>
        <div class="sub-title">
          <!-- 临时弹窗提示 -->
          <p class="tip">{{ t('pages.login.noAccount') }}</p>
          <p class="tip" @click="showRegisterDialog = true">{{ t('pages.login.createAccount') }}</p>
        </div>
      </div>

      <login v-if="type === 'login'" />
      <tdesign-setting />
    </div>
    
    <!-- 注册提示弹窗 -->
    <t-dialog
      v-model:visible="showRegisterDialog"
      header="账号注册"
      :closable="false"
      :footer="false"
      width="400px"
     :close-on-overlay-click="false">
      <div class="register-dialog-content">
        <t-icon name="info-circle-filled" size="48" style="color: #0052d9; margin-bottom: 16px;" />
        <p style="margin-bottom: 8px; font-size: 14px; line-height: 1.5;">
          如需注册账号，请联系对应的管理员进行账号分配。
        </p>
      </div>
      <template #footer>
        <t-button theme="primary" @click="showRegisterDialog = false">我知道了</t-button>
      </template>
    </t-dialog>

    <footer class="copyright">© {{ currentYear }} Kamyang Tech｜v{{ appVersion }}</footer>
  </div>
</template>
<script setup lang="ts">
import { computed, ref } from 'vue';

import TdesignSetting from '@/layouts/setting.vue';
import { t } from '@/locales';

import LoginHeader from './components/Header.vue';
import Login from './components/Login.vue';

defineOptions({
  name: 'LoginIndex',
});
const type = ref('login');
const showRegisterDialog = ref(false);
const currentYear = computed(() => new Date().getFullYear());
const appVersion = __APP_VERSION__;
</script>
<style lang="less" scoped>
@import './index.less';

.register-dialog-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 24px 0;
}
</style>

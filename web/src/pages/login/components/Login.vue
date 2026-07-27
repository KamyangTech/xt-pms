<template>
  <t-form
    ref="form"
    class="item-container"
    :class="[`login-${type}`]"
    :data="formData"
    :rules="FORM_RULES"
    label-width="0"
    @submit="onSubmit"
  >
    <template v-if="type === 'password'">
      <t-form-item name="account">
        <t-input
          :value="formData.account"
          size="large"
          :placeholder="t('pages.login.input.account')"
          @update:value="formData.account = $event"
        >
          <template #prefix-icon>
            <t-icon name="user" />
          </template>
        </t-input>
      </t-form-item>

      <t-form-item name="password">
        <t-input
          :value="formData.password"
          size="large"
          :type="showPsw ? 'text' : 'password'"
          clearable
          :placeholder="t('pages.login.input.password')"
          @update:value="formData.password = $event"
        >
          <template #prefix-icon>
            <t-icon name="lock-on" />
          </template>
          <template #suffix-icon>
            <t-icon :name="showPsw ? 'browse' : 'browse-off'" @click="showPsw = !showPsw" />
          </template>
        </t-input>
      </t-form-item>

      <!-- 数学运算验证码（首次登录失败后才出现） -->
      <t-form-item v-if="captchaRequired" name="captchaCode" class="captcha-item">
        <div class="captcha-wrapper">
          <img
            v-if="captchaImg"
            :src="captchaImg"
            class="captcha-image"
            alt="验证码"
            @click="refreshCaptcha"
          />
          <div v-else class="captcha-placeholder" @click="refreshCaptcha">获取验证码</div>
          <t-tooltip content="刷新验证码">
            <t-icon name="refresh" class="captcha-refresh" @click="refreshCaptcha" />
          </t-tooltip>
          <t-input
            :value="formData.captchaCode"
            size="large"
            :placeholder="t('pages.login.input.captcha')"
            class="captcha-input"
            @update:value="formData.captchaCode = $event"
          />
        </div>
      </t-form-item>

      <div class="check-container remember-pwd">
        <t-checkbox v-model="keepLogin">{{ t('pages.login.keepLogin') }}</t-checkbox>
        <span class="tip" @click="showForgotDialog = true">{{ t('pages.login.forget') }}</span>
      </div>
    </template>

    <t-form-item v-if="type !== 'qrcode'" class="btn-container">
      <t-button block size="large" type="submit"> {{ t('pages.login.signIn') }} </t-button>
    </t-form-item>

    <div class="switch-container">
      <span v-if="type !== 'password'" class="tip" @click="switchType('password')">{{
        t('pages.login.accountLogin')
      }}</span>
    </div>
  </t-form>

  <!-- 忘记密码弹窗 -->
  <t-dialog
    v-model:visible="showForgotDialog"
    header="忘记密码"
    :closable="false"
    :footer="false"
    width="400px"
    :close-on-overlay-click="false"
  >
    <div class="forgot-password-content">
      <t-icon name="info-circle-filled" size="48" style="color: #0052d9; margin-bottom: 16px" />
      <p style="margin-bottom: 8px; font-size: 14px; line-height: 1.5">
        如忘记密码，请联系系统管理员重置密码。
      </p>
      <p style="margin: 0; font-size: 14px; color: #666">管理员联系方式：</p>
      <p style="margin: 8px 0 0; font-size: 16px; font-weight: 500; color: #0052d9">
        admin@company.com
      </p>
    </div>
    <template #footer>
      <t-button theme="primary" @click="showForgotDialog = false">我知道了</t-button>
    </template>
  </t-dialog>
</template>
<script setup lang="ts">
import type { FormInstanceFunctions, FormRule, SubmitContext } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, nextTick, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { getCaptcha } from '@/api/user';
import { t } from '@/locales';
import { useUserStore } from '@/store';

const CAPTCHA_STORAGE_KEY = 'login_captcha_required';

const userStore = useUserStore();

const INITIAL_DATA = {
  account: '',
  password: '',
  captchaCode: '',
  checked: false,
};

const captchaRequired = ref(localStorage.getItem(CAPTCHA_STORAGE_KEY) === 'true');

const FORM_RULES = computed<Record<string, FormRule[]>>(() => ({
  account: [{ required: true, message: t('pages.login.required.account'), type: 'error' }],
  password: [{ required: true, message: t('pages.login.required.password'), type: 'error' }],
  ...(captchaRequired.value
    ? {
        captchaCode: [
          { required: true, message: t('pages.login.required.captcha'), type: 'error' },
        ],
      }
    : {}),
}));

const type = ref('password');

const form = ref<FormInstanceFunctions>();
const formData = ref({ ...INITIAL_DATA });
const showPsw = ref(false);
const showForgotDialog = ref(false);
const keepLogin = ref(false);

const captchaId = ref('');
const captchaImg = ref('');

const switchType = (val: string) => {
  type.value = val;
};

const router = useRouter();
const route = useRoute();

/** 刷新验证码 */
const refreshCaptcha = async () => {
  formData.value.captchaCode = '';
  try {
    const res = await getCaptcha();
    captchaId.value = res.captcha_id;
    captchaImg.value = res.captcha_img;
    // 图片更新后再清除验证状态，避免 DOM 更新时重新触发校验
    await nextTick();
    form.value?.clearValidate(['captchaCode']);
  } catch {
    // 静默失败
  }
};

/** 启用验证码（首次登录失败后调用） */
const enableCaptcha = async () => {
  // 保存当前表单数据，防止 rules 变化导致 t-form 重置表单
  const savedAccount = formData.value.account;
  const savedPassword = formData.value.password;

  captchaRequired.value = true;
  localStorage.setItem(CAPTCHA_STORAGE_KEY, 'true');

  // 等待 DOM 更新后恢复表单数据
  await nextTick();
  formData.value.account = savedAccount;
  formData.value.password = savedPassword;

  await refreshCaptcha(); // 内部会清空验证码输入
};

const onSubmit = async (ctx: SubmitContext) => {
  if (ctx.validateResult === true) {
    try {
      await userStore.login({
        username: formData.value.account,
        password: formData.value.password,
        captcha_id: captchaId.value,
        captcha_val: formData.value.captchaCode,
        keep_login: keepLogin.value,
      });

      // 登录成功，清除验证码需求
      localStorage.removeItem(CAPTCHA_STORAGE_KEY);
      captchaRequired.value = false;

      MessagePlugin.success('登录成功');
      const redirect = route.query.redirect as string;
      const redirectUrl = redirect ? decodeURIComponent(redirect) : '/dashboard';
      router.push(redirectUrl);
    } catch (e: any) {
      const errMsg = e?.message || '';

      // 网络错误不处理（拦截器已做提示）
      if (
        errMsg.includes('Network Error') ||
        errMsg.includes('ECONNREFUSED') ||
        errMsg.includes('econnrefused') ||
        errMsg.includes('timeout')
      ) {
        return;
      }

      // 从后端 JSON 响应中提取真实错误消息（e.message 已被 Axios 改写）
      const backendMsg = e?.response?.data?.message || '';

      // 仅当是登录相关的业务错误时才启用验证码（排除服务器内部错误）
      const isLoginError =
        backendMsg.includes('用户名或密码错误') ||
        backendMsg.includes('账号已被禁用') ||
        backendMsg.includes('账号已被锁定') ||
        backendMsg.includes('验证码错误');

      if (isLoginError && !captchaRequired.value) {
        await enableCaptcha();
      } else if (captchaRequired.value) {
        await refreshCaptcha();
      }
    }
  } else {
    MessagePlugin.error('请填写完整的登录信息');
  }
};

// 页面加载时：如果之前已经需要验证码，直接获取
if (captchaRequired.value) {
  refreshCaptcha();
}
</script>
<style lang="less" scoped>
@import '../index.less';

.forgot-password-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 24px 0;
}

.captcha-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
}

.captcha-image {
  height: 40px;
  width: auto;
  border-radius: var(--td-radius-default);
  cursor: pointer;
  flex-shrink: 0;
  border: 1px solid var(--td-border-level-2-color);
}

.captcha-placeholder {
  height: 40px;
  width: 120px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed var(--td-border-level-2-color);
  border-radius: var(--td-radius-default);
  cursor: pointer;
  font-size: 12px;
  color: var(--td-text-color-placeholder);
  flex-shrink: 0;
}

.captcha-refresh {
  font-size: 18px;
  cursor: pointer;
  color: var(--td-text-color-secondary);
  flex-shrink: 0;
  transition: transform 0.3s;
}

.captcha-refresh:hover {
  color: var(--td-brand-color);
  transform: rotate(90deg);
}

.captcha-input {
  flex: 1;
}

.captcha-item :deep(.t-form__controls-content) {
  width: 100%;
}
</style>

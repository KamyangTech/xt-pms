<template>
  <div>
    <t-row :gutter="[24, 24]">
      <t-col :flex="3">
      <div class="user-left-greeting">
        <div>
          Hi，{{ userInfo.name || 'Image' }}
          <span class="regular"> {{ markDayText }}</span>
        </div>
        <img :src="companyLogo" class="logo" />
      </div>

      <t-card class="user-info-list" :title="t('pages.user.personalInfo.title')" :bordered="false">
        <template #actions>
          <t-dropdown :options="profileOptions" @click="handleProfileAction">
            <t-button theme="default" shape="square" variant="text">
              <t-icon name="ellipsis" />
            </t-button>
          </t-dropdown>
        </template>
        <t-descriptions :column="4" item-layout="vertical">
          <t-descriptions-item v-for="(item, index) in USER_INFO_LIST" :key="index" :label="t(item.title)">
            {{ item.content }}
          </t-descriptions-item>
        </t-descriptions>
      </t-card>

      <!-- 工作概览 -->
      <t-card class="work-overview-card" :title="t('pages.user.workOverview.title')" :bordered="false">
        <t-row :gutter="[16, 16]">
          <t-col v-for="item in WORK_STAT_LIST" :key="item.title" :xs="6">
            <t-card
              :bordered="false"
              class="work-stat-card"
              :class="item.theme"
            >
              <div class="work-stat-card__top">
                <span>{{ item.number }}</span>
              </div>
      <div class="work-stat-card__icon-wrap">
        <span>
          <t-icon :name="item.icon" />
        </span>
      </div>
              <template #footer>
                <div class="work-stat-card__bottom">
                  <span>{{ t(item.title) }}</span>
                </div>
              </template>
            </t-card>
          </t-col>
        </t-row>
      </t-card>
    </t-col>

    <t-col :flex="1">
      <t-card class="user-intro" :bordered="false">
        <t-avatar size="80px">{{ userInfo.name?.charAt(0).toUpperCase() || 'T' }}</t-avatar>
        <div class="name">{{ userInfo.name || 'My Account' }}</div>
        <div class="position">{{ userRoles }}</div>
      </t-card>


    </t-col>
  </t-row>

  <!-- 编辑个人资料对话框 -->
  <t-dialog
    v-model:visible="editProfileVisible"
    :title="t('pages.user.editProfile.title')"
    width="520px"
    placement="center"
    :confirm-btn="{ content: '保存', loading: editProfileLoading }"
    @confirm="handleEditProfileSubmit"
    @cancel="editProfileVisible = false"
   :close-on-overlay-click="false">
    <t-form ref="editProfileFormRef" :data="editProfileModel" :rules="editProfileRules" label-width="100px">
      <t-form-item :label="t('pages.user.editProfile.username')" name="username">
        <t-input v-model="editProfileModel.username" disabled />
      </t-form-item>
      <t-form-item label="姓名" name="real_name">
        <t-input v-model="editProfileModel.real_name" placeholder="请输入姓名" />
      </t-form-item>
      <t-form-item label="手机号" name="phone">
        <t-input v-model="editProfileModel.phone" placeholder="请输入手机号" />
      </t-form-item>
      <t-form-item label="邮箱" name="email">
        <t-input v-model="editProfileModel.email" placeholder="请输入邮箱" />
      </t-form-item>
      <t-form-item label="职位" name="position">
        <t-input v-model="editProfileModel.position" placeholder="请输入职位" />
      </t-form-item>
      <t-form-item label="入职日期" name="hire_date">
        <t-date-picker
          v-model="editProfileModel.hire_date"
          placeholder="请选择入职日期"
          format="YYYY-MM-DD"
          clearable
        />
      </t-form-item>
    </t-form>
  </t-dialog>

  <!-- 修改密码对话框 -->
  <t-dialog
    v-model:visible="changePwdVisible"
    :title="t('pages.user.changePasswordDialog.title')"
    width="440px"
    placement="center"
    :confirm-btn="{ content: '确定', loading: changePwdLoading }"
    @confirm="handleChangePwdSubmit"
    @cancel="handleChangePwdCancel"
   :close-on-overlay-click="false">
    <t-form ref="changePwdFormRef" :data="changePwdModel" :rules="changePwdRules" label-width="100px">
      <t-form-item :label="t('pages.user.changePasswordDialog.oldPassword')" name="old_password">
        <t-input
          v-model="changePwdModel.old_password"
          type="password"
          :placeholder="t('pages.user.changePasswordDialog.oldPasswordPlaceholder')"
        />
      </t-form-item>
      <t-form-item :label="t('pages.user.changePasswordDialog.newPassword')" name="new_password">
        <t-input
          v-model="changePwdModel.new_password"
          type="password"
          :placeholder="t('pages.user.changePasswordDialog.newPasswordPlaceholder')"
        />
      </t-form-item>
      <t-form-item :label="t('pages.user.changePasswordDialog.confirmPassword')" name="confirm_password">
        <t-input
          v-model="changePwdModel.confirm_password"
          type="password"
          :placeholder="t('pages.user.changePasswordDialog.confirmPasswordPlaceholder')"
        />
      </t-form-item>
    </t-form>
  </t-dialog>
</div>
</template>
<script setup lang="ts">
import { computed, h, onMounted, reactive, ref } from 'vue';
import dayjs from 'dayjs';
import timezone from 'dayjs/plugin/timezone';
import utc from 'dayjs/plugin/utc';

dayjs.extend(utc);
dayjs.extend(timezone);

import defaultLogo from '@/assets/assets-kamyang-logo.png';
import { getCompanyInfo } from '@/api/company-management';
import { getFileFullUrl } from '@/api/common/upload';
import { changePassword, type UserProfile, updateProfile, getUserStatistics } from '@/api/user';
import { t } from '@/locales';
import { useUserStore } from '@/store';
import { EditIcon, LockOnIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import type { DropdownProps } from 'tdesign-vue-next';

defineOptions({
  name: 'UserIndex',
});

const userStore = useUserStore();

// 从 store 获取用户信息
const userInfo = computed(() => userStore.userInfo);

// 用户详细数据
const userProfile = ref<UserProfile | null>(null);

// 公司信息
const companyInfo = ref<{ short_name: string; logo?: string } | null>(null);

// 公司 logo URL：优先使用上传的 logo，否则使用默认图片
const companyLogo = computed(() => {
  if (companyInfo.value?.logo) {
    return getFileFullUrl(companyInfo.value.logo);
  }
  return defaultLogo;
});

const joinDays = computed(() => {
  // 优先使用入职日期，如果没有则使用创建时间
  const dateStr = userProfile.value?.hire_date || userProfile.value?.create_time;
  if (!dateStr) return 0;
  
  const joinDate = dayjs(dateStr).tz('Asia/Shanghai');
  const today = dayjs().tz('Asia/Shanghai');
  return today.diff(joinDate, 'day');
});

const markDayText = computed(() => {
  const days = joinDays.value;
  const companyName = companyInfo.value?.short_name || '公司';
  
  // 获取当前时段（使用东八区）
  const hour = dayjs().tz('Asia/Shanghai').hour();
  let timeGreeting = '你好';
  if (hour < 12) {
    timeGreeting = '上午好';
  } else if (hour < 18) {
    timeGreeting = '下午好';
  } else {
    timeGreeting = '晚上好';
  }
  
  return `${timeGreeting}，今天是你加入${companyName}的第 ${days} 天`;
});

const userRoles = computed(() => {
  if (!userProfile.value?.roles || userProfile.value.roles.length === 0) {
    return '-';
  }
  return userProfile.value.roles.map(role => role.role_name).join(' ');
});

// 动态生成用户信息列表 - 从后端 API 获取真实数据
const USER_INFO_LIST = computed(() => {
  // 如果没有用户数据，返回空数组
  if (!userProfile.value) {
    // 尝试从 localStorage 获取备用数据
    const userInfoStr = localStorage.getItem('userInfo');
    if (userInfoStr) {
      try {
        const localData: any = JSON.parse(userInfoStr);
        return [
          {
            title: 'pages.user.personalInfo.desc.mobile',
            content: localData.phone || '-',
          },
          {
            title: 'pages.user.personalInfo.desc.email',
            content: localData.email || '-',
          },
          {
            title: 'pages.user.personalInfo.desc.position',
            content: localData.position || '-',
          },
          {
            title: 'pages.user.personalInfo.desc.joinDay',
            content: localData.create_time ? localData.create_time.split('T')[0] : '-',
          },
        ];
      } catch (e) {
        console.error('解析本地用户信息失败:', e);
      }
    }
    return [];
  }
  
  return [
    {
      title: 'pages.user.personalInfo.desc.mobile',
      content: userProfile.value.phone || '-',
    },
    {
      title: 'pages.user.personalInfo.desc.email',
      content: userProfile.value.email || '-',
    },
    {
      title: 'pages.user.personalInfo.desc.position',
      content: userProfile.value.position || '-',
    },
    {
      title: 'pages.user.personalInfo.desc.joinDay',
      content: userProfile.value.hire_date 
        ? userProfile.value.hire_date.split('T')[0] 
        : (userProfile.value.create_time ? userProfile.value.create_time.split('T')[0] : '-'),
    },
  ];
});

// 工作概览数据（当前用户本月新增）
const userStats = ref<{ mold_order_count: number; production_order_count: number }>({
  mold_order_count: 0,
  production_order_count: 0,
});

const WORK_STAT_LIST = computed(() => [
  {
    title: 'pages.user.workOverview.moldOrder',
    number: userStats.value.mold_order_count || 0,
    icon: 'layers',
    theme: 'work-stat-card--approval',
  },
  {
    title: 'pages.user.workOverview.productionOrder',
    number: userStats.value.production_order_count || 0,
    icon: 'view-list',
    theme: 'work-stat-card--progress',
  },
]);

const loadUserStats = async () => {
  try {
    const res = await getUserStatistics();
    if (res) {
      userStats.value = res;
    }
  } catch {
    // 静默失败
  }
};

onMounted(() => {
  userStore.fetchUserInfo().then((res) => {
    if (res) {
      userProfile.value = res;
    } else {
      // API 调用失败，使用 localStorage 中的数据
      const userInfoStr = localStorage.getItem('userInfo');
      if (userInfoStr) {
        try {
          userProfile.value = JSON.parse(userInfoStr);
        } catch (e) {
          console.error('解析本地用户信息失败:', e);
        }
      }
    }
  });
  
  getCompanyInfo().then((res) => {
    if (res) {
      companyInfo.value = res;
    }
  }).catch((err) => {
    console.error('获取公司信息失败:', err);
  });

  loadUserStats();
  });

// 个人信息卡片下拉菜单选项
const profileOptions: DropdownProps['options'] = [
  {
    content: t('pages.user.personalInfo.edit'),
    value: 'editProfile',
    prefixIcon: () => h(EditIcon),
  },
  {
    content: t('pages.user.personalInfo.changePassword'),
    value: 'changePassword',
    prefixIcon: () => h(LockOnIcon),
  },
];

// 个人信息卡片下拉菜单点击处理
const handleProfileAction: DropdownProps['onClick'] = (data) => {
  if (data.value === 'editProfile') {
    handleEditProfile();
  } else if (data.value === 'changePassword') {
    handleChangePassword();
  }
};

// ==================== 编辑个人资料 ====================

const editProfileVisible = ref(false);
const editProfileLoading = ref(false);
const editProfileFormRef = ref();
const editProfileModel = reactive({
  username: '',
  real_name: '',
  phone: '',
  email: '',
  position: '',
  hire_date: undefined as string | undefined,
});

// 手机号正则
const phoneReg = /^1[3-9]\d{9}$/;

const editProfileRules = {
  phone: [
    {
      pattern: phoneReg,
      message: '请输入正确的手机号',
      trigger: 'blur' as const,
    },
  ],
};

// 打开编辑资料对话框
const handleEditProfile = () => {
  if (userProfile.value) {
    editProfileModel.username = userProfile.value.username;
    editProfileModel.real_name = userProfile.value.real_name;
    editProfileModel.phone = userProfile.value.phone || '';
    editProfileModel.email = userProfile.value.email || '';
    editProfileModel.position = userProfile.value.position || '';
    editProfileModel.hire_date = userProfile.value.hire_date
      ? userProfile.value.hire_date.split('T')[0]
      : undefined;
  } else {
    // 从 localStorage 兜底
    const userInfoStr = localStorage.getItem('userInfo');
    if (userInfoStr) {
      try {
        const localData: any = JSON.parse(userInfoStr);
        editProfileModel.username = localData.username || '';
        editProfileModel.real_name = localData.real_name || '';
        editProfileModel.phone = localData.phone || '';
        editProfileModel.email = localData.email || '';
        editProfileModel.position = localData.position || '';
        editProfileModel.hire_date = localData.hire_date
          ? localData.hire_date.split('T')[0]
          : undefined;
      } catch (e) {
        console.error('解析本地用户信息失败:', e);
      }
    }
  }
  editProfileVisible.value = true;
};

// 提交编辑个人资料
const handleEditProfileSubmit = async () => {
  const valid = await editProfileFormRef.value?.validate();
  if (valid !== true) return;

  editProfileLoading.value = true;
  try {
    await updateProfile({
      real_name: editProfileModel.real_name,
      phone: editProfileModel.phone,
      email: editProfileModel.email,
      position: editProfileModel.position,
      hire_date: editProfileModel.hire_date,
    });
    MessagePlugin.success(t('pages.user.editProfile.success'));
    editProfileVisible.value = false;

    // 刷新用户信息
    userStore.fetchUserInfo().then((res) => {
      if (res) {
        userProfile.value = res;
      }
    });
  } catch {
    MessagePlugin.error(t('pages.user.editProfile.error'));
  } finally {
    editProfileLoading.value = false;
  }
};

// ==================== 修改密码 ====================

const changePwdVisible = ref(false);
const changePwdLoading = ref(false);
const changePwdFormRef = ref();
const changePwdModel = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
});

const changePwdRules = {
  old_password: [
    { required: true, message: t('pages.user.changePasswordDialog.oldPasswordRequired'), trigger: 'blur' as const },
  ],
  new_password: [
    { required: true, message: t('pages.user.changePasswordDialog.newPasswordRequired'), trigger: 'blur' as const },
    { min: 6, message: t('pages.user.changePasswordDialog.newPasswordMin'), trigger: 'blur' as const },
  ],
  confirm_password: [
    { required: true, message: t('pages.user.changePasswordDialog.confirmPasswordRequired'), trigger: 'blur' as const },
    {
      validator: (val: string) => val === changePwdModel.new_password,
      message: t('pages.user.changePasswordDialog.passwordMismatch'),
      trigger: 'blur' as const,
    },
  ],
};

const handleChangePassword = () => {
  changePwdModel.old_password = '';
  changePwdModel.new_password = '';
  changePwdModel.confirm_password = '';
  changePwdVisible.value = true;
};

const handleChangePwdCancel = () => {
  changePwdVisible.value = false;
};

const handleChangePwdSubmit = async () => {
  const valid = await changePwdFormRef.value?.validate();
  if (valid !== true) return;

  changePwdLoading.value = true;
  try {
    await changePassword({
      old_password: changePwdModel.old_password,
      new_password: changePwdModel.new_password,
    });
    MessagePlugin.success(t('pages.user.changePasswordDialog.success'));
    changePwdVisible.value = false;
  } catch {
    MessagePlugin.error(t('pages.user.changePasswordDialog.error'));
  } finally {
    changePwdLoading.value = false;
  }
};

// ==================== 工具函数 ====================
</script>
<style lang="less" scoped>
@import './index.less';

.work-overview-card {
  margin-top: 16px;
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__header) {
    padding: 0;
  }

  :deep(.t-card__body) {
    padding: 0;
  }

  :deep(.t-card__footer) {
    padding: 0;
  }
}

.work-stat-card {
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__header) {
    padding: 0;
  }

  :deep(.t-card__footer) {
    padding: 0;
    margin-top: var(--td-comp-margin-xxl);
  }

  :deep(.t-card__body) {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    flex: 1;
    position: relative;
    padding: 0;
    margin-top: var(--td-comp-margin-s);
    margin-bottom: var(--td-comp-margin-xxl);
  }

  &__top {
    > span {
      display: inline-block;
      color: var(--td-text-color-primary);
      font-size: var(--td-font-size-headline-medium);
      line-height: var(--td-line-height-headline-medium);
    }
  }

  &__bottom {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;

    > span {
      color: var(--td-text-color-placeholder);
      font-size: var(--td-font-size-body-small);
    }
  }

  &__icon-wrap {
    position: absolute;
    top: 0;
    right: 0;

    > span {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: 40px;
      height: 40px;
      border-radius: 50%;
    }

    .t-icon {
      font-size: 22px;
    }
  }

  // 不同状态的卡片主题色
  &--pending {
    .work-stat-card__icon-wrap > span {
      background: var(--td-brand-color-light);
    }
    .work-stat-card__icon-wrap .t-icon {
      color: var(--td-brand-color);
    }
  }

  &--approval {
    .work-stat-card__icon-wrap > span {
      background: var(--td-success-color-light);
    }
    .work-stat-card__icon-wrap .t-icon {
      color: var(--td-success-color);
    }
  }

  &--progress {
    .work-stat-card__icon-wrap > span {
      background: var(--td-warning-color-light);
    }
    .work-stat-card__icon-wrap .t-icon {
      color: var(--td-warning-color);
    }
  }

  &--completed {
    .work-stat-card__icon-wrap > span {
      background: var(--td-danger-color-light);
    }
    .work-stat-card__icon-wrap .t-icon {
      color: var(--td-danger-color);
    }
  }
}

.t-descriptions {
  margin-top: 24px;
}
</style>

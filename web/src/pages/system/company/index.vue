<template>
  <div class="company-container">
    <!-- 操作栏 -->
    <t-card class="header-card">
      <div class="header-content">
        <span class="page-title">公司信息</span>
        <div class="header-actions">
          <t-button variant="outline" @click="fetchCompanyInfo">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button theme="primary" @click="handleSaveCompany">
            <template #icon><check-icon /></template>
            保存
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 公司信息表单 -->
    <t-card :bordered="false">
      <t-form
        ref="formRef"
        :data="formData"
        :rules="formRules"
        label-width="100px"
        style="max-width: 600px;"
      >
        <t-form-item label="公司名称" name="company_name">
          <t-input
            v-model="formData.company_name"
            placeholder="请输入公司全称"
            :maxlength="100"
          />
        </t-form-item>
        <t-form-item label="公司简称" name="short_name">
          <t-input
            v-model="formData.short_name"
            placeholder="请输入公司简称"
            :maxlength="50"
          />
        </t-form-item>

        <t-form-item label="公司 Logo" name="logo">
          <t-upload
            v-model="logoFiles"
            theme="image"
            :max="1"
            accept=".png,.jpg,.jpeg,.gif,.svg"
            :request-method="handleLogoUpload"
            :abridge-name="[8, 6]"
            tips="建议尺寸 336*44px，支持 JPG/PNG/GIF/SVG 格式"
            @remove="handleLogoRemove"
          />
        </t-form-item>
      </t-form>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { CheckIcon, RefreshIcon } from 'tdesign-icons-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { onMounted, reactive, ref } from 'vue';
import type { UploadFile } from 'tdesign-vue-next';

import { getCompanyInfo, updateCompanyInfo } from '@/api/company-management';
import { request } from '@/utils/request';

defineOptions({
  name: 'SystemCompany',
});

interface CompanyForm {
  company_name: string;
  short_name: string;
}

const formRef = ref();
const loading = ref(false);

const formData = reactive<CompanyForm>({
  company_name: '',
  short_name: '',
});

// Logo 图片上传
const logoFiles = ref<UploadFile[]>([]);
const currentLogoUrl = ref<string>('');

const formRules = {
  company_name: [{ required: true, message: '请输入公司名称', trigger: 'blur' as const }],
};

const fetchCompanyInfo = async () => {
  loading.value = true;
  try {
    const res = await getCompanyInfo();
    if (res) {
      formData.company_name = res.company_name || '';
      formData.short_name = res.short_name || '';
      currentLogoUrl.value = res.logo || '';

      // 如果已有 logo，回显到上传组件
      if (res.logo) {
        logoFiles.value = [
          {
            name: 'company-logo',
            url: res.logo.startsWith('http') ? res.logo : `${import.meta.env.VITE_API_URL || ''}${res.logo}`,
            status: 'success',
          } as UploadFile,
        ];
      } else {
        logoFiles.value = [];
      }
    }
  } catch (error) {
    console.error('获取公司信息失败:', error);
    MessagePlugin.error('获取公司信息失败');
  } finally {
    loading.value = false;
  }
};

// 上传 logo
const handleLogoUpload = async (file: any) => {
  const formData = new FormData();
  formData.append('file', file.raw || file);
  try {
    const res = await request.post<{ url: string }>({ url: '/system/company/logo', data: formData, headers: { 'Content-Type': 'multipart/form-data' } });
    currentLogoUrl.value = res.url;
    return { status: 'success', url: res.url } as any;
  } catch {
    MessagePlugin.error('Logo 上传失败');
    return { status: 'fail' } as any;
  }
};

// 移除 logo
const handleLogoRemove = async () => {
  try {
    await request.delete({ url: '/system/company/logo' });
    currentLogoUrl.value = '';
    logoFiles.value = [];
    MessagePlugin.success('已恢复默认 Logo');
  } catch {
    MessagePlugin.error('操作失败');
  }
};

const handleSaveCompany = async () => {
  const valid = await formRef.value?.validate();
  if (valid !== true) return;

  try {
    await updateCompanyInfo({
      company_name: formData.company_name.trim(),
      short_name: formData.short_name.trim(),
      logo: currentLogoUrl.value,
    });
    MessagePlugin.success('保存成功');
    fetchCompanyInfo(); // 刷新
  } catch (error) {
    console.error('保存失败:', error);
    MessagePlugin.error('保存失败');
  }
};

onMounted(() => {
  fetchCompanyInfo();
});
</script>

<style lang="less" scoped>
.company-container {
  padding: 16px;

  .header-card {
    margin-bottom: 16px;

    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .page-title {
        font-size: 16px;
        font-weight: 500;
      }

      .header-actions {
        display: flex;
        gap: 8px;
      }
    }
  }
}
</style>

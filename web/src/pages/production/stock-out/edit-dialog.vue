<template>
  <t-dialog
    v-model:visible="visible"
    header="编辑送货单"
    :close-on-overlay-click="false"
    width="800px"
    :confirm-btn="{
      content: '保存',
      theme: 'primary',
      loading: submitting,
    }"
    :cancel-btn="'取消'"
    @confirm="handleSubmit"
    @close="handleClose"
  >
    <t-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      :label-width="120"
    >
      <!-- 不可编辑的基础信息 -->
      <t-divider align="left">基础信息（不可修改）</t-divider>
      
      <t-row :gutter="[16, 16]">
        <t-col :span="8">
          <t-form-item label="出库单号">
            <t-input :value="stockOutData?.stock_code" disabled />
          </t-form-item>
        </t-col>
        <t-col :span="8">
          <t-form-item label="模具编号">
            <t-input :value="getMoldCode()" disabled />
          </t-form-item>
        </t-col>
        <t-col :span="8">
          <t-form-item label="产品名称">
            <t-input :value="stockOutData?.product_name" disabled />
          </t-form-item>
        </t-col>
      </t-row>

      <t-row :gutter="[16, 16]">
        <t-col :span="8">
          <t-form-item label="出库数量">
            <t-input-number :value="stockOutData?.quantity" disabled style="width: 100%" />
          </t-form-item>
        </t-col>
        <t-col :span="8">
          <t-form-item label="出库日期">
            <t-input :value="formatDateTime(stockOutData?.stock_out_date)" disabled />
          </t-form-item>
        </t-col>
        <t-col :span="8">
          <t-form-item label="批次号">
            <t-input :value="stockOutData?.batch_no || '-'" disabled />
          </t-form-item>
        </t-col>
      </t-row>

      <!-- 可编辑的收货信息 -->
      <t-divider align="left">收货信息</t-divider>
      
      <t-row :gutter="[16, 16]">
        <t-col :span="12">
          <t-form-item label="接收客户" name="customerName">
            <t-input
              v-model="formData.customerName"
              placeholder="请输入客户名称"
            />
          </t-form-item>
        </t-col>
        
        <t-col :span="12">
          <t-form-item label="发货日期" name="deliveryDate">
            <t-date-picker
              v-model="formData.deliveryDate"
              placeholder="请选择发货日期"
              format="YYYY-MM-DD"
              style="width: 100%"
            />
          </t-form-item>
        </t-col>
      </t-row>

      <!-- 可编辑的发货方式 -->
      <t-divider align="left">发货方式</t-divider>
      
      <t-row :gutter="[16, 16]">
        <t-col :span="12">
          <t-form-item label="发货方式" name="deliveryMethod">
            <t-select
              v-model="formData.deliveryMethod"
              placeholder="请选择发货方式"
              clearable
            >
              <t-option
                v-for="item in deliveryMethodDict"
                :key="item.dict_value"
                :label="item.dict_label"
                :value="item.dict_value"
              />
            </t-select>
          </t-form-item>
        </t-col>
        
        <t-col :span="12" v-if="formData.deliveryMethod === 'EXPRESS'">
          <t-form-item label="快递公司">
            <t-select
              v-model="formData.expressCompany"
              placeholder="请选择快递公司"
              clearable
            >
              <t-option
                v-for="item in expressCompanyDict"
                :key="item.dict_value"
                :label="item.dict_label"
                :value="item.dict_value"
              />
            </t-select>
          </t-form-item>
        </t-col>
        
        <t-col :span="12" v-if="formData.deliveryMethod === 'EXPRESS'">
          <t-form-item label="快递单号" help="仅做保存不校验，推荐复制单号/扫描输入">
            <t-input
              v-model="formData.expressNo"
              placeholder="请输入或扫描快递单号"
              clearable
              @blur="handleRecognizeExpress"
            />
          </t-form-item>
        </t-col>
        
        <t-col :span="12" v-if="formData.deliveryMethod === 'LALAMOVE'">
          <t-form-item label="司机电话">
            <t-input
              v-model="formData.driverPhone"
              placeholder="请输入司机联系电话"
            />
          </t-form-item>
        </t-col>
        
        <t-col :span="12" v-if="formData.deliveryMethod === 'LALAMOVE'">
          <t-form-item label="车牌号">
            <t-input
              v-model="formData.driverPlate"
              placeholder="请输入车牌号"
            />
          </t-form-item>
        </t-col>
      </t-row>

      <!-- 照片上传 -->
      <t-divider align="left">发货照片</t-divider>
      
      <t-form-item label="上传照片">
        <FileUploader
          ref="photoUploaderRef"
          v-model="formData.photos"
          :auto-save="false"
          :image-only="true"
          :max-count="5"
          :multiple="true"
          category="stock-out"
          upload-btn-text="上传发货照片"
        />
      </t-form-item>

      <!-- 备注 -->
      <t-form-item label="备注">
        <t-textarea
          v-model="formData.remark"
          placeholder="请输入备注（可选）"
          :rows="3"
          maxlength="500"
        />
      </t-form-item>
    </t-form>
  </t-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import type { FormInstanceFunctions, FormRule } from 'tdesign-vue-next';

import { updateStockOut, recognizeExpressCompany } from '@/api/production/stock';
import { useDictRaw } from '@/hooks/useDict';
import { uploadFiles } from '@/api/common/upload';
import FileUploader from '@/components/common/FileUploader.vue';
import { formatDate, formatDateTime } from '@/utils/date';

const visible = ref(false);
const submitting = ref(false);
const formRef = ref<FormInstanceFunctions>();
const photoUploaderRef = ref();
const stockOutData = ref<any>(null);

// 表单数据
const formData = ref({
  customerName: '',
  deliveryDate: '',
  deliveryMethod: '',
  expressCompany: '',
  expressNo: '',
  photos: [] as any[],
  driverPhone: '',
  driverPlate: '',
  remark: '',
});

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: deliveryMethodDict } = useDictRaw('delivery_method');
const { items: expressCompanyDict } = useDictRaw('express_company');

// 表单验证规则
const rules: Record<string, FormRule[]> = {
  customerName: [{ required: true, message: '请输入客户名称', type: 'error' }],
  deliveryDate: [{ required: true, message: '请选择发货日期', type: 'error' }],
  deliveryMethod: [{ required: true, message: '请选择发货方式', type: 'error' }],
};

const getMoldCode = () => {
  if (!stockOutData.value) return '-';
  return stockOutData.value.mold_code || stockOutData.value.external_mold_code || '-';
};

const parsePhotos = (photosStr: string) => {
  try {
    return photosStr ? JSON.parse(photosStr) : [];
  } catch {
    return [];
  }
};

// 识别快递公司
const handleRecognizeExpress = async () => {
  if (!formData.value.expressNo) return;
  
  try {
    const res = await recognizeExpressCompany(formData.value.expressNo);
    if (res.express_company) {
      formData.value.expressCompany = res.express_company;
      MessagePlugin.success(`已自动识别快递公司: ${res.express_company_name}`);
    }
  } catch (error) {
    console.error('识别快递公司失败:', error);
  }
};

// 打开对话框
const open = async (row: any) => {
  stockOutData.value = row;
  
  // 字典已通过 useDictRaw 自动加载
  
  // 解析日期字符串，确保正确填充到日期选择器
  let parsedDeliveryDate = '';
  if (row.delivery_date) {
    // 如果后端返回的是 YYYY-MM-DD 格式，直接使用
    // 如果是其他格式，尝试解析
    const dateStr = row.delivery_date;
    if (dateStr.includes('T')) {
      // ISO 8601 格式，提取日期部分
      parsedDeliveryDate = dateStr.split('T')[0];
    } else if (dateStr.includes(' ')) {
      // 日期时间格式，提取日期部分
      parsedDeliveryDate = dateStr.split(' ')[0];
    } else {
      // 已经是 YYYY-MM-DD 格式
      parsedDeliveryDate = dateStr;
    }
  }
  
  // 获取客户名称：优先使用出库单的 customer_name，如果没有则从关联订单获取
  let customerName = row.customer_name || '';
  if (!customerName && row.order?.customer?.cus_name) {
    customerName = row.order.customer.cus_name;
  }
  
  // 填充表单数据
  formData.value = {
    customerName,
    deliveryDate: parsedDeliveryDate,
    deliveryMethod: row.delivery_method || '',
    expressCompany: row.express_company || '',
    expressNo: row.express_no || '',
    photos: parsePhotos(row.photos),
    driverPhone: row.driver_phone || '',
    driverPlate: row.driver_plate || '',
    remark: row.remark || '',
  };
  
  visible.value = true;
};

// 关闭对话框
const handleClose = () => {
  formRef.value?.reset();
  stockOutData.value = null;
};

// 提交表单
const handleSubmit = async () => {
  const valid = await formRef.value?.validate();
  if (valid !== true) return;

  submitting.value = true;
  try {
    // 1. 先上传待上传的照片（统一上传接口）
    // 从 FileUploader v-model 提取已有照片的 URL
    const existingPhotoUrls = (formData.value.photos || []).map((p: any) => p.url || p);
    const pending = photoUploaderRef.value?.getPendingChanges() || {};
    const pendingFiles = pending.uploads || [];
    
    let allUrls = [...existingPhotoUrls];
    if (pendingFiles.length > 0 && stockOutData.value?.id) {
      const results = await uploadFiles({ files: pendingFiles, category: 'stock-out' });
      if (results && results.length > 0) {
        allUrls = [...allUrls, ...results.map(r => r.url)];
      }
    }
    
    // 2. 格式化日期为 YYYY-MM-DD
    let formattedDeliveryDate = '';
    if (formData.value.deliveryDate) {
      const date = new Date(formData.value.deliveryDate);
      if (!isNaN(date.getTime())) {
        formattedDeliveryDate = formatDate(formData.value.deliveryDate);
      }
    }

    // 3. 更新出库单信息
    const params: any = {
      customer_name: formData.value.customerName || undefined,
      delivery_date: formattedDeliveryDate || undefined,
      delivery_method: formData.value.deliveryMethod || undefined,
      express_company: formData.value.expressCompany || undefined,
      express_no: formData.value.expressNo || undefined,
      photos: allUrls.length > 0 ? JSON.stringify(allUrls) : undefined,
      driver_phone: formData.value.driverPhone || undefined,
      driver_plate: formData.value.driverPlate || undefined,
      remark: formData.value.remark || undefined,
    };

    await updateStockOut(stockOutData.value.id, params);
    
    MessagePlugin.success('送货单更新成功');
    visible.value = false;
    
    // 触发刷新
    emit('success');
  } catch (error: any) {
    console.error('更新送货单失败:', error);
    const errorMsg = error.response?.data?.message || error.message || '更新送货单失败';
    MessagePlugin.error(errorMsg);
  } finally {
    submitting.value = false;
  }
};

const emit = defineEmits(['success']);

defineExpose({
  open,
});

// 组件挂载时字典已通过 useDictRaw 自动加载
</script>

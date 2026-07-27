<template>
  <div class="production-info-form-container">
    <div class="step-title">填写生产交付要求</div>

    <!-- 仅制造模具时跳过生产信息 -->
    <t-alert
      v-if="orderType === 'MOLD_ONLY'"
      theme="info"
      message="当前为仅制造模具模式，无需填写生产信息，请直接点击下一步"
      :close="false"
      style="margin-bottom: 24px"
    />

    <template v-else>
      <!-- 综合订单模式下，显示带入的模具信息 -->
      <t-alert
        v-if="orderType === 'COMBINED' && prefilledMoldItems && prefilledMoldItems.length > 0"
        theme="success"
        :message="`已带入 ${prefilledMoldItems.length} 个模具信息，模具来源已自动设置为'自制'`"
        :close="false"
        style="margin-bottom: 16px"
      />

      <!-- 生产配置与外发信息 -->
      <div class="form-section" style="margin-bottom: 24px">
        <t-form :label-width="100" label-align="top">
          <t-row :gutter="[32, 24]">
            <t-col :span="4">
              <t-form-item label="生产类型">
                <t-select v-model="formData.production_type" placeholder="请选择生产类型">
                  <t-option label="本厂生产" value="IN_HOUSE" />
                  <t-option label="外发生产" value="OUTSOURCED" />
                </t-select>
              </t-form-item>
            </t-col>
            <t-col :span="4">
              <t-form-item label="优先级">
                <t-select v-model="formData.priority" placeholder="请选择优先级">
                  <t-option v-for="item in priorityOptions" :key="item.value" :label="item.label" :value="item.value" />
                </t-select>
              </t-form-item>
            </t-col>
            <t-col v-if="formData.production_type === 'OUTSOURCED'" :span="4">
              <t-form-item label="外发加工商">
                <t-select v-model="formData.processor_id" placeholder="请选择加工商" clearable>
                  <t-option v-for="processor in processorList" :key="processor.id" :label="processor.displayName" :value="processor.id" />
                </t-select>
              </t-form-item>
            </t-col>

          </t-row>
        </t-form>
      </div>

      <!-- 产品配套明细 -->
      <ProductItemsForm
        ref="productItemsFormRef"
        v-model="formData.items"
        :show-customer-selector="false"
        :show-mold-source-selector="orderType !== 'COMBINED'"
        :show-mold-selector="true"
        :show-image-upload="true"
        :show-plastic-material="false"
        :editable="true"
        :allow-drag="true"
        :allow-delete="true"
        :prefilled-customer-id="prefilledCustomerId"
        :prefilled-mold-items="orderType === 'COMBINED' ? prefilledMoldItems : []"
        @mold-source-type-change="handleMoldSourceTypeChange"
      />
    </template>

    <div class="form-actions">
      <t-button type="reset" theme="default" variant="base" @click="onReset">上一步</t-button>
      <t-button theme="primary" @click="handleNext" :loading="submitting">
        下一步
      </t-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FormRule, SubmitContext } from 'tdesign-vue-next';
import { onMounted, reactive, ref, watch } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import ProductItemsForm from '@/components/ProductItemsForm.vue';
import { getProcessorList } from '@/api/reference';
import { dictUtils } from '@/utils/dict';

const props = defineProps({
  orderType: {
    type: String,
    default: 'COMBINED',
  },
  // 预选客户ID（从订单信息带入）
  prefilledCustomerId: {
    type: Number,
    default: undefined,
  },
  // 预选模具信息（从模具信息带入）
  prefilledMoldItems: {
    type: Array as () => Array<{
      mold_id?: number; // 模具ID
      mold_code: string;
      mold_name: string; // 模具名称（用户自定义）
      amount?: number;
      moldProcess?: number | string;
      moldType?: number | string;
      materialType?: number | string;
      remark: string;
    }>,
    default: () => [] as Array<{
      mold_id?: number;
      mold_code: string;
      mold_name: string;
      amount?: number;
      moldProcess?: number | string;
      moldType?: number | string;
      materialType?: number | string;
      remark: string;
    }>,
  },
});

const emit = defineEmits<{
  (e: 'next', data: any): void
  (e: 'prev'): void
}>();

const formRef = ref();
const productItemsFormRef = ref();
const submitting = ref(false);

// 加工商列表
const processorList = ref<any[]>([]);

// 优先级选项（从字典加载）
const priorityOptions = ref<any[]>([]);

// 表单数据
const formData = reactive({
  mold_source_type: 'SELF_MADE' as string, // ✨ 模具来源类型（订单级别）
  production_type: 'IN_HOUSE' as string,   // ✨ 生产类型
  priority: 'NORMAL' as string,            // ✨ 优先级
  processor_id: undefined as number | undefined, // ✨ 加工商ID（外发时填写）

  items: [] as Array<{
    id?: string;
    product_name: string;
    product_image_url?: string; // 产品图片URL
    mold_code?: string;
    mold_name?: string;
    mold_id?: number; // 单模具ID
    mold_ids?: number[]; // 多模具列表
    external_mold_code?: string; // ✨ 外来模具编号
    plan_qty: number;
    device_id?: number | string;
    process_params: string;
    remark: string;
  }>,
});

// 表单验证规则
const FORM_RULES: Record<string, FormRule[]> = {};

const handleAddItem = () => {
  // 综合订单模式下，自动带入模具信息
  let newItem: any = {
    product_name: '',
    product_image_url: '',
    plan_qty: undefined,
    device_id: undefined,
    process_params: '',
    remark: '',
  };
  
  // ✨ 综合订单模式下，自动绑定模具信息
  if (props.orderType === 'COMBINED' && props.prefilledMoldItems && props.prefilledMoldItems.length > 0) {
    // 如果只有 1 套模具，自动绑定
    if (props.prefilledMoldItems.length === 1) {
      const moldItem = props.prefilledMoldItems[0];
      
      // ✨ 只使用有效的数字模具ID
      const validMoldId = moldItem.mold_id && !isNaN(Number(moldItem.mold_id)) 
        ? Number(moldItem.mold_id) 
        : undefined;
      
      newItem.mold_code = moldItem.mold_code || '';
      newItem.mold_name = moldItem.mold_name || '';
      newItem.mold_id = validMoldId;
      newItem.mold_ids = validMoldId ? [validMoldId] : [];
    }
  }
  
  formData.items.push(newItem);
};

const handleRemoveItem = (index: number) => {
  if (formData.items.length <= 1) {
    MessagePlugin.warning('至少保留一项产品明细');
    return;
  }
  formData.items.splice(index, 1);
};

// 判断指定行是否有待上传的图片（用于验证时替代 product_image_url 检查）
const hasPendingImageForRow = (rowIndex: number): boolean => {
  const componentRef = productItemsFormRef.value?.productImageRefs?.get(rowIndex);
  if (!componentRef?.getPendingChanges) return false;
  const changes = componentRef.getPendingChanges();
  return changes.hasChanges && changes.uploads && changes.uploads.length > 0;
};

// 上传所有待处理的图片（仅在最终提交时调用，不提前上传）
const uploadPendingImages = async (): Promise<boolean> => {
  if (productItemsFormRef.value?.productImageRefs) {
    const refs = productItemsFormRef.value.productImageRefs;
    for (const [rowIndex, componentRef] of refs.entries()) {
      if (componentRef?.getPendingChanges) {
        const changes = componentRef.getPendingChanges();
        if (changes.hasChanges && changes.uploads && changes.uploads.length > 0) {
          try {
            const { uploadImage } = await import('@/api/common/upload');
            const response = await uploadImage(changes.uploads[0]);
            const imageUrl: string = response?.url || (typeof response === 'string' ? response : '');
            if (formData.items[rowIndex]) {
              formData.items[rowIndex].product_image_url = imageUrl;
            }
            componentRef.resetPendingChanges();
          } catch (error: any) {
            MessagePlugin.error(`产品图片上传失败: ${error.message || '未知错误'}`);
            return false;
          }
        }
      }
    }
  }
  return true;
};

const handleNext = async () => {
  // ✨ 生产信息校验
  if (props.orderType !== 'MOLD_ONLY') {
    // 1. 至少有一条产品明细
    if (!formData.items || formData.items.length === 0) {
      MessagePlugin.warning('请至少添加一条产品明细');
      return;
    }

    // 2. 逐项校验产品明细的必填字段
    for (let i = 0; i < formData.items.length; i++) {
      const item = formData.items[i];
      const itemLabel = `第${i + 1}行产品`;

      // 产品名称不能为空
      if (!item.product_name?.trim()) {
        MessagePlugin.warning(`${itemLabel}的产品名称不能为空`);
        return;
      }

      // 产品图片：检查 product_image_url 或是否有待上传的图片
      const hasImage = !!item.product_image_url?.trim() || hasPendingImageForRow(i);
      if (!hasImage) {
        MessagePlugin.warning(`${itemLabel}的产品图片不能为空，请上传产品图片`);
        return;
      }

      // 计划数量必须大于 0
      if (!item.plan_qty || item.plan_qty <= 0) {
        MessagePlugin.warning(`${itemLabel}的计划数量必须大于 0`);
        return;
      }
    }

    // 3. 外发生产时校验加工商
    if (formData.production_type === 'OUTSOURCED' && !formData.processor_id) {
      MessagePlugin.warning('外发生产时，请选择加工商');
      return;
    }
  }

  // 触发onSubmit的逻辑
  await onSubmit({ validateResult: true } as SubmitContext);
};

// 提交表单（仅传递数据，不直接提交订单）
const onSubmit = async (ctx: SubmitContext) => {
  if (ctx.validateResult === true) {
    try {
      submitting.value = true;
      
      // 仅制造模具模式时，直接传递空数据
      if (props.orderType === 'MOLD_ONLY') {
        emit('next', { items: [] });
        return;
      }

      // 只传递数据，不直接提交订单
      // 注意：图片尚未上传，仅传递待处理状态。最终提交时由 form.vue 调用 uploadPendingImages() 统一上传
      emit('next', {
        mold_source_type: formData.mold_source_type,
        production_type: formData.production_type,
        priority: formData.priority,
        processor_id: formData.production_type === 'OUTSOURCED' ? formData.processor_id : undefined,
        processor_name: formData.production_type === 'OUTSOURCED' && formData.processor_id
          ? (processorList.value.find((p: any) => p.id === formData.processor_id)?.name || '')
          : '',
        items: formData.items,
      });
    } catch (error: any) {
      MessagePlugin.error(error.message || '提交失败');
    } finally {
      submitting.value = false;
    }
  }
};

const onReset = () => {
  emit('prev');
};

// ✨ 处理模具来源变更
const handleMoldSourceTypeChange = (value: string | undefined) => {
  if (value) {
    formData.mold_source_type = value;
  }
};

onMounted(() => {
});

// 监听模具信息变化，自动初始化产品明细
watch(
  () => props.prefilledMoldItems,
  (newMoldItems) => {    
    // 如果是综合订单模式，且有模具数据，根据模具信息初始化产品明细
    if (props.orderType === 'COMBINED' && newMoldItems && newMoldItems.length > 0) {
      formData.items.splice(0, formData.items.length);
      
      // 为每个模具生成一行空白的产品明细，等待用户填写
      newMoldItems.forEach((moldItem, moldIndex) => {        
        // ✨ 只使用有效的数字模具ID，如果是临时ID则设为 undefined
        const validMoldId = moldItem.mold_id && !isNaN(Number(moldItem.mold_id)) 
          ? Number(moldItem.mold_id) 
          : undefined;
        
        formData.items.push({
          id: `mold_${moldIndex}_${Date.now()}`, // 唯一标识
          product_name: '', // ✨ 空白，由用户手动填写产品名称
          product_image_url: '', //  空白，由用户手动上传产品图片
          mold_code: moldItem.mold_code || '', //  带入模具编号（用于显示）
          mold_name: moldItem.mold_name || '', //  带入模具名称（用于显示）
          mold_id: validMoldId, // ✨ 单选模式：只绑定一个模具（仅有效数字ID）
          mold_ids: validMoldId ? [validMoldId] : [], // ✨ 多模具列表：只包含有效数字ID
          external_mold_code: undefined, // ✨ 外来模具编号（综合订单不使用）
          plan_qty: undefined,
          device_id: undefined as number | undefined,
          process_params: '',
          remark: '', // ✨ 产品备注初始化为空，不自动带入模具备注
        });
      });
    }
  },
  { immediate: true, deep: true }
);


defineExpose({
  validate: async () => {
    return await formRef.value?.validate();
  },
  getData: () => ({
    items: formData.items,
  }),
  uploadPendingImages, // 暴露给 form.vue 在最终提交时调用
});

const loadProcessorList = async () => {
  try {
    const res = await getProcessorList();
    processorList.value = (res.list || []).map((item: any) => ({
      id: item.id,
      name: item.processor_name || item.name,
      processType: item.process_type,
      displayName: item.process_type 
        ? `${item.processor_name || item.name} - ${item.process_type}`
        : item.processor_name || item.name,
    }));
  } catch (error) {
    console.error('获取加工商列表失败:', error);
  }
};

const loadPriorityDict = async () => {
  try {
    const options = await dictUtils.getDictOptions('order-priority');
    priorityOptions.value = options;
    // 设置默认值（如果字典中有默认值）
    if (options.length > 0 && !formData.priority) {
      formData.priority = options[0].value;
    }
  } catch (error) {
    console.error('加载优先级字典失败:', error);
    // 失败时使用硬编码默认值
    priorityOptions.value = [
      { label: '普通', value: 'NORMAL' },
      { label: '紧急', value: 'URGENT' },
      { label: '非常紧急', value: 'CRITICAL' }
    ];
  }
};

onMounted(() => {
  loadProcessorList();
  loadPriorityDict();
});
</script>

<style scoped lang="less">
.step-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  margin-bottom: 24px;
}

.form-actions {
  margin-top: var(--td-comp-margin-xxl);
  display: flex;
  gap: var(--td-comp-margin-s);
}

.items-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--td-comp-margin-l);
}

.empty-tip {
  padding: var(--td-comp-paddingTB-xl) 0;
  text-align: center;
}

.production-items {
  display: flex;
  flex-direction: column;
  gap: var(--td-comp-margin-l);
}

.production-item-card {
  padding: var(--td-comp-paddingTB-l) var(--td-comp-paddingLR-l);
  background-color: var(--td-bg-color-secondarycontainer);
  border-radius: var(--td-radius-medium);
}

.production-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--td-comp-margin-m);
}

.production-item-title {
  font-weight: 600;
  font-size: 14px;
  color: var(--td-text-color-primary);
}
</style>

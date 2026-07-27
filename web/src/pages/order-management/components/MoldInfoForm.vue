<template>
  <div class="mold-info-form-container">
    <div class="step-title">填写模具相关明细</div>
    
    <!-- 仅生产产品时跳过模具信息 -->
    <t-alert
      v-if="orderType === 'PROD_ONLY'"
      theme="info"
      message="当前为仅生产产品模式，无需填写模具信息，请直接点击下一步"
      :close="false"
      style="margin-bottom: 24px"
    />

    <template v-else>
      <!-- 金额统计区域（横向排列） -->
      <div class="amount-summary-section">
        <t-row :gutter="[32, 24]">
          <t-col :span="2">
            <t-form-item label="模具总金额">
              <t-statistic :value="moldTotalAmount" prefix="¥" unit="元" trend="increase" />
            </t-form-item>
          </t-col>

        </t-row>
      </div>
      
      <!-- 模具配置 -->
      <div class="form-section" style="margin-bottom: 24px">
        <t-form :label-width="100" label-align="top">
          <t-row :gutter="[32, 24]">
            <t-col :span="4">
              <t-form-item label="优先级">
                <t-select v-model="formData.priority" placeholder="请选择优先级">
                  <t-option v-for="item in priorityOptions" :key="item.value" :label="item.label" :value="item.value" />
                </t-select>
              </t-form-item>
            </t-col>
          </t-row>
        </t-form>
      </div>
      
      <!-- 模具明细 - 使用公共组件 -->
      <MoldItemsForm
        v-model="formData.items"
        :order-type="orderType"
        :editable="orderType === 'COMBINED' || orderType === 'MOLD_ONLY'"
        :allow-drag="orderType === 'COMBINED' || orderType === 'MOLD_ONLY'"
        :allow-delete="orderType === 'COMBINED' || orderType === 'MOLD_ONLY'"
      />
    </template>

    <div class="form-actions">
      <t-button type="reset" theme="default" variant="base" @click="handlePrev">上一步</t-button>
      <t-button theme="primary" @click="handleSubmit">
        下一步
      </t-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { FormRule } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import MoldItemsForm from '@/components/MoldItemsForm.vue';
import { dictUtils } from '@/utils/dict';

const props = defineProps({
  orderType: {
    type: String,
    default: 'COMBINED',
  },
});

const emit = defineEmits<{
  (e: 'next', data: any): void
  (e: 'prev'): void
}>();

const formRef = ref();

// 表单数据
const formData = reactive({
  priority: 'NORMAL' as string,            // ✨ 优先级
  items: [
    {
      mold_id: 0, //  设置为 0，表示新模具
      mold_code: '',
      mold_name: '',
      amount: undefined as number | undefined,
      moldProcess: undefined as number | string | undefined,
      moldType: undefined as number | string | undefined,
      materialType: undefined as number | string | undefined,
      cavityCount: undefined as number | undefined,
      workType: undefined as string | undefined,
      processorId: undefined as number | undefined,
      remark: '',
    },
  ],
});


const priorityOptions = ref<any[]>([]);

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

// 表单验证规则
const FORM_RULES: Record<string, FormRule[]> = {};

const moldTotalAmount = computed(() => {
  return formData.items.reduce((sum, item) => sum + (item.amount || 0), 0);
});

// 提交表单
const handleSubmit = async () => {
  try {
    // 仅生产产品模式时，直接传递空数据
    if (props.orderType === 'PROD_ONLY') {
      emit('next', { items: [] });
      return;
    }

    // ✨ 校验模具编号是否重复
    const moldCodes = formData.items.map(item => item.mold_code?.trim()).filter(Boolean);
    const uniqueCodes = new Set(moldCodes);
    
    if (moldCodes.length !== uniqueCodes.size) {
      MessagePlugin.warning('模具编号重复，请修改后再提交');
      return;
    }

    const emptyCodeIndex = formData.items.findIndex(item => !item.mold_code?.trim());
    if (emptyCodeIndex !== -1) {
      const moldIdentifier = formData.items[emptyCodeIndex].mold_name || `${emptyCodeIndex + 1}`;
      MessagePlugin.warning(`模具【${moldIdentifier}】的编号不能为空`);
      return;
    }

    // ✨ 校验模具名称是否为空
    const emptyNameIndex = formData.items.findIndex(item => !item.mold_name?.trim());
    if (emptyNameIndex !== -1) {
      const moldIdentifier = formData.items[emptyNameIndex].mold_code || `${emptyNameIndex + 1}`;
      MessagePlugin.warning(`模具【${moldIdentifier}】的名称不能为空`);
      return;
    }

    // ✨ 校验金额是否为正数
    const invalidAmountIndex = formData.items.findIndex(item => item.amount === undefined || item.amount <= 0);
    if (invalidAmountIndex !== -1) {
      const moldIdentifier = formData.items[invalidAmountIndex].mold_code || formData.items[invalidAmountIndex].mold_name || `${invalidAmountIndex + 1}`;
      MessagePlugin.warning(`模具【${moldIdentifier}】的金额必须大于 0`);
      return;
    }

    // ✨ 校验工艺是否已选择
    const emptyMoldProcessIndex = formData.items.findIndex(item => !item.moldProcess);
    if (emptyMoldProcessIndex !== -1) {
      const moldIdentifier = formData.items[emptyMoldProcessIndex].mold_code || formData.items[emptyMoldProcessIndex].mold_name || `${emptyMoldProcessIndex + 1}`;
      MessagePlugin.warning(`模具【${moldIdentifier}】的工艺不能为空`);
      return;
    }

    // ✨ 校验模胚是否已选择
    const emptyMoldTypeIndex = formData.items.findIndex(item => !item.moldType);
    if (emptyMoldTypeIndex !== -1) {
      const moldIdentifier = formData.items[emptyMoldTypeIndex].mold_code || formData.items[emptyMoldTypeIndex].mold_name || `${emptyMoldTypeIndex + 1}`;
      MessagePlugin.warning(`模具【${moldIdentifier}】的模胚不能为空`);
      return;
    }

    // ✨ 校验材质是否已选择
    const emptyMaterialTypeIndex = formData.items.findIndex(item => !item.materialType);
    if (emptyMaterialTypeIndex !== -1) {
      const moldIdentifier = formData.items[emptyMaterialTypeIndex].mold_code || formData.items[emptyMaterialTypeIndex].mold_name || `${emptyMaterialTypeIndex + 1}`;
      MessagePlugin.warning(`模具【${moldIdentifier}】的材质不能为空`);
      return;
    }

    // ✨ 校验模腔数是否已选择
    const emptyCavityCountIndex = formData.items.findIndex(item => !item.cavityCount);
    if (emptyCavityCountIndex !== -1) {
      const moldIdentifier = formData.items[emptyCavityCountIndex].mold_code || formData.items[emptyCavityCountIndex].mold_name || `${emptyCavityCountIndex + 1}`;
      MessagePlugin.warning(`模具【${moldIdentifier}】的模腔数不能为空`);
      return;
    }

    // ✨ 校验每个模具的加工方式（workType）是否已选择
    const emptyWorkTypeIndex = formData.items.findIndex(item => !item.workType);
    if (emptyWorkTypeIndex !== -1) {
      const moldIdentifier = formData.items[emptyWorkTypeIndex].mold_code || formData.items[emptyWorkTypeIndex].mold_name || `${emptyWorkTypeIndex + 1}`;
      MessagePlugin.warning(`模具【${moldIdentifier}】的加工方式不能为空`);
      return;
    }

    // ✨ 校验外发时加工商必填
    const emptyProcessorIndex = formData.items.findIndex(item => item.workType === 'OUTSOURCED' && !item.processorId);
    if (emptyProcessorIndex !== -1) {
      const moldIdentifier = formData.items[emptyProcessorIndex].mold_code || formData.items[emptyProcessorIndex].mold_name || `${emptyProcessorIndex + 1}`;
      MessagePlugin.warning(`模具【${moldIdentifier}】生产方式为外发加工时，请选择加工商`);
      return;
    }

    emit('next', {
      items: formData.items.map(item => ({
        ...item,
        work_type: item.workType,
        processor_id: item.processorId,
      })),
      totalAmount: moldTotalAmount.value,
      priority: formData.priority,
    });
  } catch (error: any) {
    MessagePlugin.error(error.message || '提交失败');
  }
};

// 上一步
const handlePrev = () => {
  emit('prev');
};

onMounted(() => {
  // 字典数据由 MoldItemsForm 组件内部加载
  loadPriorityDict();
});

defineExpose({
  validate: async () => {
    return await formRef.value?.validate();
  },
  getData: () => ({
    items: formData.items,
    totalAmount: moldTotalAmount.value,
  }),
});
</script>

<style scoped lang="less">
.step-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  margin-bottom: 24px;
}

.amount-summary-section {
  margin-bottom: var(--td-comp-margin-l);
}

.form-actions {
  margin-top: var(--td-comp-margin-xxl);
  display: flex;
  gap: var(--td-comp-margin-s);
}
</style>

<template>
  <div class="production-order-form-page">
    <div class="back-button-container">
      <t-button theme="default" @click="goBack" variant="text">
        <template #icon><t-icon name="chevron-left" /></template>
        返回
      </t-button>
    </div>
    <t-form
      ref="formRef"
      class="production-order-form"
    :data="formData"
    :rules="currentFormRules"
    label-align="top"
    :label-width="100"
    @reset="onReset"
    @submit="onSubmit"
  >
    <div class="form-basic-container">
      <div class="form-basic-item">
        <!-- 基本信息 -->
        <div class="form-basic-container-title">基本信息</div>
        <t-row class="row-gap" :gutter="[32, 24]">
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="订单号" name="order_code">
              <t-input v-model="formData.order_code" placeholder="系统自动生成" disabled />
            </t-form-item>
          </t-col>
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="客户" name="customer_id">
              <t-select
                v-model="formData.customer_id"
                placeholder="请选择客户"
                clearable
                filterable
                :disabled="isReadOnly"
                @change="handleCustomerChange"
              >
                <t-option
                  v-for="customer in customerList"
                  :key="customer.id"
                  :value="customer.id"
                  :label="customer.name"
                />
              </t-select>
            </t-form-item>
          </t-col>
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="优先级" name="priority">
              <t-select v-model="formData.priority" placeholder="请选择优先级">
                <t-option
                  v-for="item in priorityDictList"
                  :key="item.dict_value"
                  :value="item.dict_value"
                  :label="item.dict_label"
                />
              </t-select>
            </t-form-item>
          </t-col>
        </t-row>

        <!-- 生产信息 -->
        <div class="form-basic-container-title form-title-gap">生产信息</div>
        <t-row class="row-gap" :gutter="[32, 24]">
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="生产类型" name="production_type">
              <t-select 
                v-model="formData.production_type" 
                placeholder="请选择生产类型"
                clearable
                :disabled="isReadOnly"
              >
                <t-option
                  v-for="item in productionTypeOptions"
                  :key="item.value"
                  :value="item.value"
                  :label="item.label"
                />
              </t-select>
            </t-form-item>
          </t-col>
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="模具来源" name="mold_source_type">
              <t-select 
                v-model="formData.mold_source_type" 
                placeholder="请选择模具来源"
                clearable
                :disabled="isReadOnly"
              >
                <t-option
                  v-for="item in moldSourceDictList"
                  :key="item.dict_value"
                  :value="item.dict_value"
                  :label="item.dict_label"
                />
              </t-select>
            </t-form-item>
          </t-col>
        </t-row>

        <t-row class="row-gap" :gutter="[32, 24]">
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="计划开始日期" name="plan_start_date">
              <t-date-picker
                v-model="formData.plan_start_date"
                theme="primary"
                mode="date"
                style="width: 100%"
                :disabled="isReadOnly"
              />
            </t-form-item>
          </t-col>
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="计划结束日期" name="plan_end_date">
              <t-date-picker
                v-model="formData.plan_end_date"
                theme="primary"
                mode="date"
                style="width: 100%"
              />
            </t-form-item>
          </t-col>
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="负责人" name="worker_id">
              <t-select
                v-model="formData.worker_id"
                placeholder="请选择负责人"
                clearable
                filterable
                :disabled="isReadOnly"
              >
                <t-option
                  v-for="item in employeeDictList"
                  :key="item.dict_value"
                  :value="item.dict_value"
                  :label="item.dict_label"
                />
              </t-select>
            </t-form-item>
          </t-col>
        </t-row>

        <!-- 订单明细（配套生产） -->
        <div class="form-basic-container-title form-title-gap">订单明细</div>
        <t-row class="row-gap" :gutter="[32, 24]">
          <t-col :span="24">
            <div class="items-header">
              <span>产品配套明细</span>
              <span class="pending-upload-tip" v-if="hasAnyPendingImageChanges">
                提示:图片已暂存,请点击底部更新按钮保存
              </span>
              <t-button theme="primary" variant="outline" size="small" :disabled="isItemsDisabled" @click="handleAddItem">
                <template #icon><add-icon /></template>
                添加产品
              </t-button>
            </div>
            <t-table
              v-if="formData.items && formData.items.length > 0"
              :data="formData.items"
              :columns="itemColumns"
              row-key="index"
              bordered
              :pagination="{ pageSize: 5 }"
            >
              <template #product_name="{ row, rowIndex }">
                <!-- 产品名称始终可编辑 -->
                <t-input
                  v-model="row.product_name"
                  :placeholder="formData.mold_source_type === 'SELF_MADE' ? '选择模具后自动填充，也可手动修改' : (formData.mold_source_type === 'CUSTOMER_PROVIDED' ? '请输入产品名称' : '请先选择模具')"
                  :disabled="isItemsDisabled || isReadOnly"
                />
              </template>
              <template #mold_id="{ row, rowIndex }">
                <t-select
                  v-if="formData.mold_source_type === 'SELF_MADE'"
                  v-model="row.mold_id"
                  placeholder="请选择模具"
                  clearable
                  filterable
                  :disabled="!formData.customer_id || isItemsDisabled || isReadOnly"
                  @change="(value: any) => handleMoldChange(value, row, rowIndex)"
                >
                  <t-option
                    v-for="mold in availableMoldList"
                    :key="mold.id"
                    :value="mold.id"
                    :label="`${mold.mold_code} - ${mold.mold_name || ''}`"
                  />
                </t-select>
                <!-- 外来模具：输入编号时自动查询档案 -->
                <t-input
                  v-else
                  v-model="row.external_mold_code"
                  placeholder="请输入外来模具编号"
                  clearable
                  :disabled="isItemsDisabled || isReadOnly"
                  @blur="() => handleExternalMoldCodeBlur(row, rowIndex)"
                />
              </template>
              <template #device_id="{ row, rowIndex }">
                <t-select
                  v-model="row.device_id"
                  placeholder="请选择设备"
                  clearable
                  filterable
                  :disabled="isItemsDisabled || isReadOnly"
                >
                  <t-option
                    v-for="item in equipmentDictList"
                    :key="item.dict_value"
                    :value="Number(item.dict_value)"
                    :label="item.dict_label"
                  />
                </t-select>
              </template>
              <template #plan_qty="{ row, rowIndex }">
                <t-input-number v-model="row.plan_qty" :min="1" style="width: 100%" :disabled="isItemsDisabled || isReadOnly" />
              </template>
              <template #product_image="{ row, rowIndex }">
                <div class="image-upload-cell">
                  <FileUploader
                    :ref="(el) => setProductImageRef(el, rowIndex)"
                    :model-value="row.product_image_url ? [{ id: '', key: 'img-'+rowIndex, url: row.product_image_url, fileName: '', fileSize: 0, fileType: '' }] : []"
                    :auto-save="false"
                    :max-count="1"
                    :multiple="false"
                    image-only
                    :show-empty-hint="false"
                    :show-actions="!(isItemsDisabled || (isReadOnly && !canUploadImage))"
                    :disabled="isItemsDisabled || (isReadOnly && !canUploadImage)"
                    inline
                    category="product-images"
                    upload-btn-text=""
                    @change="handleProductImageChange(rowIndex, $event)"
                  />
                </div>
              </template>
              <template #plastic_material_id="{ row, rowIndex }">
                <t-select
                  v-model="row.plastic_material_id"
                  placeholder="请选择塑胶原料"
                  clearable
                  filterable
                  :disabled="isItemsDisabled || isReadOnly"
                >
                  <t-option
                    v-for="item in plasticMaterialDictList"
                    :key="item.id"
                    :value="Number(item.id)"
                    :label="item.dict_label"
                  />
                </t-select>
              </template>
              <template #op="{ rowIndex }">
                <t-link theme="danger" @click="handleDeleteItem(rowIndex)" :disabled="isItemsDisabled || isReadOnly">删除</t-link>
              </template>
            </t-table>
            <div v-else class="empty-items-tip">
              暂无明细，请点击"添加产品"按钮添加配套产品
            </div>
          </t-col>
        </t-row>

        <!-- 外发信息（仅在外发生产时显示） -->
        <div v-if="formData.production_type === 'OUTSOURCED'" class="form-basic-container-title form-title-gap">外发信息</div>
        <t-row v-if="formData.production_type === 'OUTSOURCED'" class="row-gap" :gutter="[32, 24]">
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="加工商" name="processor_id">
              <t-select
                v-model="formData.processor_id"
                placeholder="请选择加工商"
                clearable
                filterable
                :disabled="isReadOnly"
              >
                <t-option
                  v-for="processor in processorList"
                  :key="processor.id"
                  :value="processor.id"
                  :label="processor.name"
                />
              </t-select>
            </t-form-item>
          </t-col>
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="外发单价" name="outsource_price">
              <t-input-number
                v-model="formData.outsource_price"
                :min="0"
                :precision="2"
                placeholder="请输入外发单价"
                style="width: 100%"
                :disabled="isReadOnly"
              />
            </t-form-item>
          </t-col>
          <t-col :span="12" :xs="12" :sm="8" :md="8" :lg="6" :xl="4">
            <t-form-item label="总金额">
              <t-input
                :value="totalAmountDisplay"
                disabled
                placeholder="自动计算"
              />
            </t-form-item>
          </t-col>
        </t-row>

        <!-- 备注 -->
        <div class="form-basic-container-title form-title-gap">其他信息</div>
        <t-form-item label="备注" name="remark">
          <t-textarea
            v-model="formData.remark"
            :autosize="{ minRows: 3, maxRows: 5 }"
            placeholder="请输入备注信息"
          />
        </t-form-item>

        <!-- 款项配置 -->
        <div class="form-basic-container-title form-title-gap">款项配置</div>
        <PaymentPlanForm
          ref="paymentPlanRef"
          :order-type="'PROD_ONLY'"
          :mold-items="[]"
          :production-items="formData.items || []"
          :production-type="formData.production_type"
          :outsource-price="formData.outsource_price || 0"
          :hide-step-u-i="true"
          :initial-product-payments="initialProductPayments"
          @next="handlePaymentPlanNext"
          @prev="handlePaymentPlanPrev"
        />
      </div>
    </div>

    <div class="form-submit-container">
      <div class="form-submit-sub">
        <div class="form-submit-left">
          <t-button 
            theme="primary" 
            class="form-submit-confirm" 
            type="submit" 
            :loading="isSubmitting" 
            :disabled="isSubmitting"
          >
            {{ isEdit ? '更新' : '确认' }}
          </t-button>
          <t-button type="reset" class="form-submit-cancel" theme="default" variant="base" :disabled="isSubmitting">
            取消
          </t-button>
        </div>
      </div>
    </div>
  </t-form>
  </div>
</template>

<script setup lang="ts">
import { MessagePlugin, FormInstanceFunctions, DialogPlugin } from 'tdesign-vue-next';
import { AddIcon, DeleteIcon } from 'tdesign-icons-vue-next';
import { computed, onMounted, ref, h, nextTick } from 'vue';

/** HTML 转义，防止 XSS */
const escapeHtml = (str: string): string => {
  const map: Record<string, string> = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' };
  return str.replace(/[&<>"']/g, (c) => map[c]);
};
import { useRoute, useRouter } from 'vue-router';

import { getCustomerList, getStatusFlowConfigs, getProcessorList } from '@/api/reference';
import { useDictRaw } from '@/hooks/useDict';
import { getProductionOrderById, createProductionOrder, updateProductionOrder } from '@/api/production/order';
import { getMoldList, getMoldById } from '@/api/mold';
import { getExternalMoldArchiveByCode } from '@/api/production/external-mold-archive';
import { getReceivablesByOrder } from '@/api/finance/receivable';
import type { ProductionOrder } from '@/types/production';
import FileUploader from '@/components/common/FileUploader.vue';
import PaymentPlanForm from '@/pages/order-management/components/PaymentPlanForm.vue';

defineOptions({
  name: 'ProductionOrderForm',
});

const router = useRouter();
const route = useRoute();
const formRef = ref<FormInstanceFunctions>();
const paymentPlanRef = ref();

// ✨ 编辑模式：预填充的产品应收款数据（从 accounts_receivable 或 master_order 加载）
const initialProductPayments = ref<Record<number, Array<{
  payment_type: string;
  amount: number;
  due_date: string;
}>>>({});

// 产品图片上传组件引用数组
const productImageUploaderRefs = ref<Map<number, any>>(new Map());
// 记录用户明确交互过的行（用于防御非预期的 pendingDeleteIndices）
const rowsWithImageInteraction = ref<Set<number>>(new Set());

// 是否有任何待处理的图片变更
const hasAnyPendingImageChanges = computed(() => {
  // 无论新增还是编辑模式，都使用延迟保存模式，需要检查待处理变更
  
  for (const [rowIndex, component] of productImageUploaderRefs.value.entries()) {
    if (component && component.getPendingChanges) {
      const pendingChanges = component.getPendingChanges();
      if (pendingChanges.hasChanges) {
        return true;
      }
    }
  }
  
  return false;
});

const setProductImageRef = (el: any, rowIndex: number) => {
  if (el) {
    productImageUploaderRefs.value.set(rowIndex, el);
  }
};

const handleProductImageChange = (rowIndex: number, event: any) => {
  if (event.hasChanges) {
    rowsWithImageInteraction.value.add(rowIndex);
  }
};

const isEdit = computed(() => !!route.params.id);
const orderId = computed(() => Number(route.params.id));

const isReadOnly = computed(() => {
  // 新增模式下，所有字段可编辑
  if (!isEdit.value) return false;
  // 编辑模式下，如果订单状态 > 2（已完成/已入库等），则部分字段只读
  // 状态 1(待排产) 和 2(生产中) 仍然允许编辑
  const status = formData.value.status;
  const editableStatuses = ['PENDING_SCHEDULE', 'PRODUCING'];
  return status && !editableStatuses.includes(status);
});

// 1. 未选择客户或模具来源时
// 2. 从综合订单来的附属生产订单（master_order_id存在时），模具已预绑定，不允许修改
const isItemsDisabled = computed(() => {
  return !formData.value.customer_id || !formData.value.mold_source_type || !!formData.value.master_order_id;
});

const canUploadImage = computed(() => {
  // 新增模式：允许上传
  if (!isEdit.value) return true;
  // 编辑模式：状态 <= 3（已完成）时允许上传图片
  // 状态 > 3（如已入库等终态）则不允许
  const status = formData.value.status;
  const allowImageUploadStatuses = ['PENDING_SCHEDULE', 'PRODUCING', 'STOCKED_IN'];
  return status && allowImageUploadStatuses.includes(status);
});

// 表单数据
const formData = ref<Partial<ProductionOrder>>({
  order_code: '',
  customer_id: undefined,
  production_type: undefined,
  mold_source_type: undefined,
  mold_id: undefined,
  external_mold_info: undefined,
  plan_start_date: '',
  plan_end_date: '',
  items: [], // 订单明细
  device_id: undefined,
  worker_id: undefined,
  processor_id: undefined,
  outsource_price: 0,
  priority: undefined,
  status: 'PENDING_SCHEDULE', // 默认状态
  remark: '',
});

// ✨ 保存原始数据用于比对生产方式变更
const originalProductionType = ref<string>('');
const originalProcessorId = ref<number | undefined>(undefined);

// 计算客户模具编号（用于表单显示和验证）
// 注意：此字段已废弃，模具信息在产品明细中管理
const customerMoldCode = computed({
  get: () => '',
  set: (_value: string) => {},
});

// 生产类型字典列表（直接使用语义化字符串）
const productionTypeOptions = computed(() => 
  productionTypeDictList.value.map(item => ({
    label: item.dict_label,
    value: item.dict_value
  }))
);

// 今天的日期字符串（YYYY-MM-DD）
const todayStr = (() => {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
})();

// 表单验证规则
const FORM_RULES: any = {
  customer_id: [{ required: true, message: '请选择客户', type: 'error' }],
  production_type: [{ required: true, message: '请选择生产类型', type: 'error' }],
  mold_source_type: [{ required: true, message: '请选择模具来源', type: 'error' }],
  priority: [{ required: true, message: '请选择优先级', type: 'error' }],
  plan_start_date: [{ required: true, message: '请选择计划开始日期', type: 'error' }],
  plan_end_date: [
    { required: true, message: '请选择计划结束日期', type: 'error' },
    {
      validator: (val: string) => {
        if (!val || !formData.value.plan_start_date) return true;
        // 规则1：结束日期不能早于开始日期
        if (val < formData.value.plan_start_date) {
          return { result: false, message: '计划结束日期不能早于计划开始日期', type: 'error' };
        }
        // 规则2：如果开始日期为历史时间，结束日期最早只能是今天
        if (formData.value.plan_start_date < todayStr && val < todayStr) {
          return { result: false, message: '计划结束日期不能早于今天', type: 'error' };
        }
        return true;
      },
      trigger: 'blur',
    },
  ],
};

// 当订单已开始生产时，动态调整验证规则（禁用的字段不需要验证）
const currentFormRules = computed(() => {
  if (isReadOnly.value) {
    // 只验证可编辑的字段
    return {
      priority: [{ required: true, message: '请选择优先级', type: 'error' }],
    };
  }
  return FORM_RULES;
});

// 下拉列表数据
const customerList = ref<any[]>([]);
const moldList = ref<any[]>([]);
const processorList = ref<any[]>([]);
const deviceList = ref<any[]>([]);
const workerList = ref<any[]>([]);

// 字典数据列表（useDictRaw 自动加载 + 全局缓存）
const { items: priorityDictList } = useDictRaw('order-priority', (items) =>
  (items as any[]).filter((item: any) => item.status === 1),
);
const { items: productionTypeDictList } = useDictRaw('production-type');
const { items: moldSourceDictList } = useDictRaw('mold-source');
const { items: equipmentDictList } = useDictRaw('production-equipment', (items) =>
  (items as any[]).filter((item: any) => item.status === 1),
);
const { items: employeeDictList } = useDictRaw('responsible-employee');
const { items: plasticMaterialDictList } = useDictRaw('plastic-material', (items) =>
  (items as any[]).filter((item: any) => item.status === 1),
);

// 可用于生产的模具状态（动态从状态流转配置获取）
const availableMoldStatusList = ref<number[]>([]);

const totalAmountDisplay = computed(() => {
  const items = formData.value.items || [];
  const price = formData.value.outsource_price || 0;
  const totalQty = items.reduce((sum, item) => sum + (item.plan_qty || 0), 0);
  return (totalQty * price).toFixed(2);
});

// 明细列定义
const itemColumns = [
  { title: '产品名称', colKey: 'product_name', width: 180 },
  { title: '模具编号', colKey: 'mold_id', width: 200 },
  { title: '产品图片', colKey: 'product_image', width: 150 },
  { title: '计划数量', colKey: 'plan_qty', width: 200 },
  { title: '塑胶原料', colKey: 'plastic_material_id', width: 180 },
  { title: '建议设备', colKey: 'device_id', width: 150 },
  { title: '操作', colKey: 'op', width: 80, fixed: 'right' as const },
];

// 根据模具ID获取模具名称
const getMoldNameById = (moldId: number): string => {
  const mold = availableMoldList.value.find(m => m.id === moldId);
  return mold?.mold_name || '';
};

// 模具选择变更 - 自动填充产品名称
const handleMoldChange = (value: any, row: any, rowIndex: number) => {
  if (value && availableMoldList.value.length > 0) {
    const selectedMold = availableMoldList.value.find(mold => mold.id === value);
    if (selectedMold && selectedMold.mold_name) {
      // 自动填充产品名称
      row.product_name = selectedMold.mold_name;
    }
  }
};

// 外来模具编号失焦 - 自动查询档案并填充产品名称
const handleExternalMoldCodeBlur = async (row: any, rowIndex: number) => {
  if (!row.external_mold_code || formData.value.mold_source_type !== 'CUSTOMER_PROVIDED') {
    return;
  }

  // 去除前后空格和制表符
  const moldCode = row.external_mold_code.trim();
  if (!moldCode) return;

  if (!formData.value.customer_id) {
    MessagePlugin.warning('请先选择客户');
    row.external_mold_code = '';
    return;
  }

  try {
    // 如果是编辑模式，传入当前订单ID以排除该订单
    const excludeOrderId = isEdit.value ? orderId.value : undefined;
    const res = await getExternalMoldArchiveByCode(moldCode, excludeOrderId);
    if (res.exists && res.data) {
      // 找到历史档案，检查客户是否匹配
      if (res.data.customer_id && res.data.customer_id !== formData.value.customer_id) {
        // 客户不匹配，提示用户
        MessagePlugin.warning(`该外来模具属于客户「${res.data.customer_name}」，与当前选择的客户不匹配`);
        row.external_mold_code = '';
        return;
      }
      
      // 客户匹配，自动填充产品名称
      row.product_name = res.data.product_name;
      
      if (res.pending_orders && res.pending_orders.length > 0) {
        row._pendingOrders = res.pending_orders;
        row._hasPendingOrders = true;
      } else {
        row._hasPendingOrders = false;
      }
      
      MessagePlugin.success(`已自动填充产品名称：${res.data.product_name}`);
    } else {
      // 未找到档案，提示用户手动输入
      MessagePlugin.info('首次使用此外来模具，请手动输入产品名称');
    }
  } catch (error: any) {
    console.error('查询外来模具档案失败:', error);
  }
};

const handleAddItem = async () => {
  // 等待响应式更新完成
  await nextTick();
  
  
  if (isItemsDisabled.value) {
    MessagePlugin.warning('当前订单无法添加产品');
    return;
  }
  
  if (!formData.value.items) {
    formData.value.items = [];
  }
  formData.value.items.push({
    product_id: undefined,
    product_name: '',
    mold_id: undefined,
    external_mold_code: undefined,
    device_id: undefined,
    plastic_material_id: undefined,
    plan_qty: 0,
    product_image_url: '',
    sort: formData.value.items.length,
    // ✨ 款项配置字段（金额计算）
    unit_price: 0,
    packaging_fee: 0,
    freight_fee: 0,
    outsource_price: 0,
  });
};

const handleDeleteItem = (index: number) => {
  formData.value.items?.splice(index, 1);
};

const getImageUrl = (url: string) => {
  if (!url) return '';
  if (url.startsWith('http')) {
    return url;
  }
  const baseUrl = import.meta.env.VITE_API_URL || '';
  return `${baseUrl}${url}`;
};

const loadCustomers = async () => {
  try {
    const res = await getCustomerList({ page_size: 1000 });
    // 将客户数据映射为下拉框所需的格式
    customerList.value = (res.list || []).map((item: any) => ({
      id: item.id,
      name: item.cus_name,
    }));
  } catch (error) {
    console.error('加载客户列表失败:', error);
  }
};

const loadMolds = async () => {
  try {
    const res = await getMoldList({ page_size: 1000 });
    // 排除已报废的模具（SCRAPPED），其他状态都可选择
    moldList.value = (res.list || []).filter(
      (mold: any) => mold.life_cycle_status !== 'SCRAPPED'
    );
  } catch (error) {
    console.error('加载模具列表失败:', error);
  }
};

// 可用的模具列表（用于下拉选择）
// 根据客户过滤：只显示该客户的模具；但已绑定到订单明细中的模具始终可选
const availableMoldList = computed(() => {
  if (!formData.value.customer_id) {
    return [];
  }
  
  // 获取当前订单明细中已绑定的模具ID列表
  const assignedMoldIds = new Set(
    (formData.value.items || [])
      .filter(item => item.mold_id)
      .map(item => item.mold_id)
  );
  
  // 统一转为 number 比较，避免后端返回 string 导致类型不匹配
  const selectedCustomerId = Number(formData.value.customer_id);
  
  return moldList.value.filter(mold => {
    const moldCustomerId = Number(mold.customer_id);
    return moldCustomerId === selectedCustomerId || assignedMoldIds.has(mold.id);
  });
});

const loadProcessors = async () => {
  try {
    const res = await getProcessorList({ page_size: 1000 });
    processorList.value = (res.list || []).map((item: any) => ({
      id: item.id,
      name: item.processor_name || item.name,  // 兼容两种字段名
    }));
  } catch (error) {
    console.error('加载加工商列表失败:', error);
  }
};

// 加载设备列表（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载工人列表（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载优先级字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载生产类型字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载模具来源字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载模具状态字典（已替换为 useDictRaw('mold-status')，自动加载 + 全局缓存）

// 加载塑胶原料字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载模具位置字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载可用于生产的模具状态（从状态流转配置动态获取）
const loadAvailableMoldStatuses = async () => {
  try {
    const configs = await getStatusFlowConfigs();
    // 从配置中提取所有模具状态值
    const statuses = new Set<number>();
    configs.forEach((config: any) => {
      if (config.mold_status !== null && config.mold_status !== undefined) {
        statuses.add(config.mold_status);
      }
    });
    availableMoldStatusList.value = Array.from(statuses);
  } catch (error) {
    console.error('加载模具状态配置失败:', error);
  }
};

const loadOrderDetail = async () => {
  if (!isEdit.value) return;

  try {
    const order = await getProductionOrderById(orderId.value);
    formData.value = { ...order };
    // 重置图片交互记录（确保每次加载数据时从干净状态开始）
    rowsWithImageInteraction.value = new Set();
    
    // ✨ 保存原始生产方式和加工商信息用于后续比对
    originalProductionType.value = formData.value.production_type || '';
    originalProcessorId.value = formData.value.processor_id;
    
    // 编辑模式下，确保已选模具在可用列表中
    // 如果订单明细中有模具ID，但该模具不在moldList中，需要手动添加到moldList
    if (formData.value.items && formData.value.items.length > 0) {
      for (const item of formData.value.items) {
        if (item.mold_id) {
          const existsInMoldList = moldList.value.some(m => m.id === item.mold_id);
          if (!existsInMoldList) {
            // 如果模具不在列表中，使用 getMoldById 精确获取该模具信息
            try {
              const mold = await getMoldById(item.mold_id);
              if (mold) {
                moldList.value.push(mold);
              }
            } catch (error) {
              console.error('[订单编辑] 获取模具信息失败:', error);
            }
          }
        }
      }
    }

    // ✨ 加载产品应收款项数据（多数据源回退）
    // 策略：1️⃣ accounts_receivable(prod_order) → 2️⃣ master_order 应收款
    try {
      let receivables: any[] = [];
      const id = orderId.value;

      // 1. 优先从应收账款表查询（生产订单类型）
      const arList = await getReceivablesByOrder(id, 'production_order');

      if (arList && Array.isArray(arList) && arList.length > 0) {
        receivables = arList;
      } else if (order.master_order_id) {
        // 2. 回退：从综合订单的应收账款查询
        try {
          const masterARList = await getReceivablesByOrder(order.master_order_id, 'master_order');
          if (masterARList && Array.isArray(masterARList) && masterARList.length > 0) {
            // 只取属于生产订单的应收款（source_type 为 production 或空）
            receivables = masterARList.filter((ar: any) => {
              return ar.source_type === 'production' || !ar.source_type;
            }).map((ar: any) => ({
              payment_type: ar.payment_type || '',
              amount: ar.amount || 0,
              due_date: ar.due_date
                ? (typeof ar.due_date === 'string' ? ar.due_date.split('T')[0] : ar.due_date)
                : '',
            }));
          }
        } catch (e) {
          console.warn('[loadOrderDetail] 查询综合订单应收款失败:', e);
        }
      }

      // 只要有应收款数据就进行预填充
      if (receivables && receivables.length > 0) {
        const productPaymentsMap: Record<number, Array<{
          payment_type: string;
          amount: number;
          due_date: string;
        }>> = {};

        // 将所有应收款作为第一个产品的期数计划
        productPaymentsMap[0] = receivables.map((ar: any) => ({
          payment_type: ar.payment_type || '',
          amount: ar.amount || 0,
          due_date: ar.due_date || '',
        }));

        initialProductPayments.value = productPaymentsMap;
      } else {
        initialProductPayments.value = {};
        console.warn('[loadOrderDetail] ⚠️ 所有数据源均为空，无法预填充产品款项');
      }
    } catch (payError) {
      console.error('[loadOrderDetail] 加载应收款项数据失败:', payError);
      initialProductPayments.value = {};
    }
  } catch (error) {
    console.error('加载订单详情失败:', error);
    MessagePlugin.error('加载订单详情失败');
  }
};

// 客户变更
const handleCustomerChange = (value: any) => {
  const customerId = typeof value === 'number' || typeof value === 'string' ? value : undefined;
  if (!customerId) return;
  
  const customer = customerList.value.find((c) => c.id === customerId);
  if (customer) {
    formData.value.customer_name = customer.name;
  }
  
  // 客户切换时，清空所有明细中的模具选择
  if (formData.value.items && formData.value.items.length > 0) {
    formData.value.items.forEach(item => {
      item.mold_id = undefined;
      item.product_name = '';
    });
    MessagePlugin.warning('已切换客户，模具选择已清空，请重新选择');
  }
};

// 防重复提交标志
const isSubmitting = ref(false);

// ✨ 获取款项配置数据，转换为后端 PaymentPlanDTO 格式
const getPaymentPlanData = () => {
  if (paymentPlanRef.value) {
    const raw = paymentPlanRef.value.getData();
    if (!raw) return null;

    const { mold_payments, product_payments, outsource_payments, mold_outsource_payments } = raw;
    const items: Array<{
      index: number;
      payment_type: string;
      amount_type: string;
      amount_value: number;
      calculated_amount: number;
      due_date: string;
      source_type: string;
    }> = [];
    let index = 0;

    for (const key of Object.keys(product_payments || {})) {
      for (const payment of product_payments[key] || []) {
        if (payment.payment_type && (payment.amount || 0) > 0) {
          items.push({
            index: index++,
            payment_type: payment.payment_type,
            amount_type: 'amount',
            amount_value: payment.amount || 0,
            calculated_amount: payment.amount || 0,
            due_date: payment.due_date || '',
            source_type: 'production',
          });
        }
      }
    }

    for (const key of Object.keys(outsource_payments || {})) {
      for (const payment of outsource_payments[key] || []) {
        if (payment.payment_type && (payment.amount || 0) > 0) {
          items.push({
            index: index++,
            payment_type: payment.payment_type,
            amount_type: 'amount',
            amount_value: payment.amount || 0,
            calculated_amount: payment.amount || 0,
            due_date: payment.due_date || '',
            source_type: 'outsource',
          });
        }
      }
    }

    return {
      enabled: items.length > 0,
      items,
    };
  }
  return null;
};

// ✨ 款项配置下一步
const handlePaymentPlanNext = async () => {
  // 生产订单中，款项配置后直接提交订单
};

// ✨ 款项配置上一步
const handlePaymentPlanPrev = () => {
};

// 提交表单
const onSubmit = async ({ validateResult }: any) => {
  // 防止重复提交
  if (isSubmitting.value) {
    console.warn('[表单提交] 表单正在提交中，忽略重复提交');
    return;
  }

  if (validateResult !== true) {
    MessagePlugin.error('请检查表单填写');
    return;
  }

  // ✨ 检查是否有生产方式变更，需要用户确认
  if (isEdit.value && originalProductionType.value) {
    const changes: string[] = [];
    
    if (formData.value.production_type !== originalProductionType.value) {
      const oldType = originalProductionType.value === 'OUTSOURCED' ? '外发' : '本厂';
      const newType = formData.value.production_type === 'OUTSOURCED' ? '外发' : '本厂';
      changes.push(`生产类型: ${oldType} → ${newType}`);
    }
    
    // 如果从外发改为本厂，或加工商变更
    if (originalProductionType.value === 'OUTSOURCED' && formData.value.production_type !== 'OUTSOURCED') {
      changes.push('将撤回当前外发记录');
    } else if (formData.value.production_type === 'OUTSOURCED' && 
               formData.value.processor_id !== originalProcessorId.value) {
      changes.push('加工商信息已变更');
    }
    
    if (changes.length > 0) {
      try {
        await DialogPlugin.confirm({
          header: '生产方式变更确认',
          body: `检测到以下变更，系统将自动冻结旧的外发记录并创建新记录：\n\n${changes.join('\n')}\n\n是否继续？`,
          theme: 'warning',
          cancelBtn: '取消',
          confirmBtn: '确认',
        });
        // 用户点击确认，继续执行
      } catch (error) {
        // 用户点击取消或关闭对话框
        return;
      }
    }
  }

  // 锁定提交状态
  isSubmitting.value = true;
  
  // 锁定编辑状态，防止在提交过程中路由变化导致判断错误
  const currentIsEdit = isEdit.value;
  const currentOrderId = orderId.value;

  // 确保必填字段有值
  if (!formData.value.production_type) {
    MessagePlugin.error('请选择生产类型');
    isSubmitting.value = false; // 释放锁
    return;
  }

  // 计划日期必填校验（与模具订单保持一致）
  if (!formData.value.plan_start_date) {
    MessagePlugin.warning('请选择计划开始日期');
    isSubmitting.value = false;
    return;
  }
  if (!formData.value.plan_end_date) {
    MessagePlugin.warning('请选择计划结束日期');
    isSubmitting.value = false;
    return;
  }

  // ✨ 订单明细非空校验
  if (!formData.value.items || formData.value.items.length === 0) {
    MessagePlugin.warning('请添加订单明细');
    isSubmitting.value = false;
    return;
  }

  {
    try {

      // ===================== 安全快照：保存本次提交前的原始URL =====================
      const originalImageUrls = new Map<number, string>();
      if (formData.value.items) {
        formData.value.items.forEach((item: any, idx: number) => {
          originalImageUrls.set(idx, item.product_image_url || '');
        });
      }

      const imageUploaderComponents = Array.from(productImageUploaderRefs.value.entries());
      if (imageUploaderComponents && imageUploaderComponents.length > 0) {
        for (const [rowIndex, component] of imageUploaderComponents) {
          if (component && component.getPendingChanges) {
            const pendingChanges = component.getPendingChanges();
            // 记录实际需要处理的行
            const userInteracted = rowsWithImageInteraction.value.has(rowIndex);

            if (pendingChanges.hasChanges) {
              // === 防御1：只有用户实际交互过的行才处理变更，防止组件库异常触发 ===
              if (!userInteracted) {
                console.warn('[表单提交] 跳过未交互行的错误变更触发，rowIndex:', rowIndex);
                if (component.resetPendingChanges) {
                  component.resetPendingChanges();
                }
                // 从快照恢复URL（防止被意外篡改）
                if (formData.value.items && formData.value.items[rowIndex] && originalImageUrls.has(rowIndex)) {
                  formData.value.items[rowIndex].product_image_url = originalImageUrls.get(rowIndex) || '';
                }
                continue;
              }

              // 标记该行是否有上传操作（用于判断是否为替换操作）
              let hasUploadForThisRow = false;

              // 优先处理文件上传
              if (pendingChanges.uploads && pendingChanges.uploads.length > 0) {
                const uploadedUrl = await uploadProductImage(pendingChanges.uploads[0]);
                if (uploadedUrl && formData.value.items && formData.value.items[rowIndex]) {
                  formData.value.items[rowIndex].product_image_url = uploadedUrl;
                  hasUploadForThisRow = true;
                }
              }

              // 如果需要删除，清空对应行的图片URL
              // 注意：如果同一行既有上传又有删除（替换场景），上传已经设置了新URL，不应再清空
              if (!hasUploadForThisRow && pendingChanges.deleteIndices && pendingChanges.deleteIndices.length > 0) {
                if (formData.value.items && formData.value.items[rowIndex]) {
                  formData.value.items[rowIndex].product_image_url = '';
                }
              }
              // 清除组件的待处理状态，避免新旧图片同时显示
              if (component.resetPendingChanges) {
                component.resetPendingChanges();
              }
            }
          }
        }
      }

      // ===================== 防御2：最终校验 - 恢复被意外清空的URL =====================
      if (formData.value.items) {
        formData.value.items.forEach((item: any, idx: number) => {
          const original = originalImageUrls.get(idx);
          const current = item.product_image_url || '';
          if (original && !current) {
            // 只有当此行没有用户交互（既无上传也无明确的删除意图）时，才恢复
            if (!rowsWithImageInteraction.value.has(idx)) {
              console.warn('[表单提交] 防御恢复: 行', idx, 'URL被意外清空，从快照恢复:', original);
              item.product_image_url = original;
            }
          }
        });
      }

      // 提交完成后清空交互记录
      rowsWithImageInteraction.value.clear();

    } catch (error) {
      console.error('处理图片变更失败:', error);
      MessagePlugin.error('图片处理失败');
      isSubmitting.value = false; // 释放锁
      return;
    }
  }

  // 外来模具双重验证
  if (formData.value.mold_source_type === 'CUSTOMER_PROVIDED' && formData.value.items) {
    const externalMoldItems = formData.value.items.filter(
      (item: any) => item.external_mold_code && item.product_name
    );

    if (externalMoldItems.length > 0) {
      const warnings: Array<{ moldCode: string; oldName: string; newName: string; pendingOrders?: any[] }> = [];

      for (const item of externalMoldItems) {
        const moldCode = item.external_mold_code.trim();
        try {
          // 如果是编辑模式，传入当前订单ID以排除该订单
          const excludeOrderId = currentIsEdit ? currentOrderId : undefined;
          const res = await getExternalMoldArchiveByCode(moldCode, excludeOrderId);
          if (res.exists && res.data) {
            // 新增模式：检查产品名称是否一致
            // 编辑模式：不检查名称一致性，只检查是否有未完成订单
            const nameConflict = !currentIsEdit && res.data.product_name !== item.product_name;
            const hasPending = res.pending_orders && res.pending_orders.length > 0;
            
            if (nameConflict || hasPending) {
              warnings.push({
                moldCode,
                oldName: res.data.product_name,
                newName: item.product_name,
                pendingOrders: res.pending_orders,
              });
            }
          }
        } catch (error) {
        }
      }

      // 如果有警告，弹窗确认
      if (warnings.length > 0) {
        const confirmResult = await handleMoldWarningsConfirm(warnings, currentIsEdit);
        if (!confirmResult) {
          isSubmitting.value = false;
          return;
        }
      }
    }
  }

  // ✨ 自有模具试模合格校验
  if (formData.value.mold_source_type === 'OWN' && formData.value.items) {
    for (const item of formData.value.items) {
      if (item.mold_id) {
        try {
          const res = await getMoldList({ page_size: 1, mold_id: item.mold_id });
          if (res.list && res.list.length > 0) {
            const mold = res.list[0];
            // 如果模具没有工单（理论上不会发生，但为了安全）
            if (!mold.work_order_status) {
              MessagePlugin.error(`模具 [${mold.mold_code}] 尚未加工完成，无法创建生产订单`);
              isSubmitting.value = false;
              return;
            }
            const allowedStatuses = ['TESTED', 'STOCKED_IN', 'DELIVERED'];
            if (!allowedStatuses.includes(mold.work_order_status)) {
              MessagePlugin.error(`模具 [${mold.mold_code}] 尚未试模合格，无法创建生产订单`);
              isSubmitting.value = false;
              return;
            }
          }
        } catch (error) {
          console.error('校验模具试模状态失败:', error);
        }
      }
    }
  }

  // 提交数据 - 深拷贝避免响应式对象问题
  const submitData: any = JSON.parse(JSON.stringify(formData.value));

  // ✨ 类型安全转换：确保 items 中的数字字段为数字类型
  if (submitData.items && Array.isArray(submitData.items)) {
    submitData.items.forEach((item: any) => {
      if (item.device_id && typeof item.device_id === 'string') {
        item.device_id = Number(item.device_id);
      }
      if (item.plastic_material_id && typeof item.plastic_material_id === 'string') {
        item.plastic_material_id = Number(item.plastic_material_id);
      }
      if (item.mold_id && typeof item.mold_id === 'string') {
        item.mold_id = Number(item.mold_id);
      }
    });
  }

  // 使用锁定的编辑状态和订单ID
  if (currentIsEdit && !submitData.id) {
    submitData.id = currentOrderId;
  }

  // ✨ 添加款项配置
  submitData.payment_plan = getPaymentPlanData();

  try {
    if (currentIsEdit) {
      await updateProductionOrder(submitData as any);
      MessagePlugin.success('更新订单成功');
    } else {
      await createProductionOrder(submitData as any);
      MessagePlugin.success('创建订单成功');

      // 新增成功后立即重置表单，防止 KeepAlive 缓存导致下次进入时显示旧数据
      formData.value = {
        order_code: '',
        customer_id: undefined,
        production_type: undefined,
        mold_source_type: undefined,
        mold_id: undefined,
        external_mold_info: undefined,
        plan_start_date: '',
        plan_end_date: '',
        items: [],
        device_id: undefined,
        worker_id: undefined,
        processor_id: undefined,
        outsource_price: 0,
        priority: undefined,
        status: 'PENDING_SCHEDULE',
        remark: '',
      };
    }
    router.push('/production/order/list');
  } catch (error) {
    console.error('[表单提交] 保存订单失败:', error);
    MessagePlugin.error('保存订单失败');
  } finally {
    isSubmitting.value = false;
  }
};

// 上传产品图片（用于延迟保存模式）
const uploadProductImage = async (file: File) => {
  try {
    const { uploadFiles } = await import('@/api/common/upload');
    const result = await uploadFiles({
      files: [file],
      category: 'product-images',
    });
    if (result && result.length > 0 && result[0].url) {
      return result[0].url;
    }
    throw new Error('图片上传失败');
  } catch (error) {
    console.error('图片上传失败:', error);
    throw error;
  }
};

const handleMoldWarningsConfirm = (warnings: Array<{ moldCode: string; oldName: string; newName: string; pendingOrders?: any[] }>, isEdit: boolean) => {
  return new Promise<boolean>((resolve) => {
    const warningList = warnings.map((item) => {
      let html = `<div style="margin-bottom: 16px; padding: 12px; background: #f5f5f5; border-radius: 4px;">`;
      html += `<div style="margin-bottom: 8px;"><b>模具编号：${escapeHtml(item.moldCode)}</b></div>`;

      // 名称不一致提示
      if (item.oldName !== item.newName) {
        html += `<div style="margin-bottom: 4px;">原名称：<span style="color: #e34d59;">${escapeHtml(item.oldName)}</span></div>`;
        html += `<div style="margin-bottom: 4px;">新名称：<span style="color: #0052d9;">${escapeHtml(item.newName)}</span></div>`;
      }

      // 未完成订单提示
      if (item.pendingOrders && item.pendingOrders.length > 0) {
        html += `<div style="margin-top: 8px; color: #e37900; font-weight: bold;">⚠️ 该模具还有 ${item.pendingOrders.length} 个未完成订单：</div>`;
        html += `<ul style="margin: 4px 0; padding-left: 20px; font-size: 13px;">`;
        item.pendingOrders.forEach((order: any) => {
          const statusMap: any = {
            1: '待排产', 2: '生产中', 3: '已完成', 4: '已入库'
          };
          html += `<li>${escapeHtml(order.order_code)} - ${statusMap[order.status] || '未知'} - 计划数量：${escapeHtml(String(order.plan_qty))}</li>`;
        });
        html += `</ul>`;
      }

      html += `</div>`;
      return html;
    }).join('');

    const dialogInstance = DialogPlugin.confirm({
      header: '模具使用提示',
      body: () => h('div', [
        h('p', '以下外来模具存在以下情况：'),
        h('div', { innerHTML: warningList }),
        h('p', { style: 'color: #e34d59; font-weight: bold; margin-top: 12px;' }, isEdit ? '是否确认继续编辑订单？' : '是否确认继续创建订单？'),
      ]),
      confirmBtn: '确认继续',
      cancelBtn: '取消',
      onConfirm: () => {
        dialogInstance.hide();
        resolve(true);
      },
      onCancel: () => {
        dialogInstance.hide();
        resolve(false);
      },
    });
  });
};

const onReset = () => {
  MessagePlugin.warning('取消操作');
  router.push('/production/order/list');
};

const goBack = () => {
  router.back();
};

onMounted(async () => {
  await Promise.all([
    loadCustomers(),
    // 模具状态字典已通过 useDictRaw('mold-status') 自动加载
    loadMolds(),
    loadProcessors(),
  ]);

  if (isEdit.value) {
    await loadOrderDetail();
  }
});
</script>

<style lang="less" scoped>
@import '@/pages/form/base/index.less';

.production-order-form-page {
  padding: 16px;
  background-color: var(--td-bg-color-page, #f5f5f5);
  min-height: 100vh;
}

.items-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  font-size: var(--td-font-size-body-medium);
  font-weight: 500;
  gap: 16px;
}

.pending-upload-tip {
  font-size: var(--td-font-size-body-medium);
  color: #e37900;
  line-height: 1.5;
  flex: 1;
}

.empty-items-tip {
  text-align: center;
  padding: 32px;
  color: var(--td-text-color-secondary);
  background: var(--td-bg-color-container);
  border: 1px dashed var(--td-component-border);
  border-radius: 3px;
}

.image-upload-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 80px;
}

.uploaded-image-preview {
  position: relative;
  display: inline-block;
}

.image-overlay-actions {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s;
  border-radius: 4px;
}

.uploaded-image-preview:hover .image-overlay-actions {
  opacity: 1;
}

.back-button-container {
  margin-bottom: 16px;
}
</style>

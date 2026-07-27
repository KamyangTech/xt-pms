<template>
  <div class="mold-work-order-form-page">
    <div class="back-button-container">
      <t-button theme="default" variant="text" @click="goBack">
        <template #icon><t-icon name="chevron-left" /></template>
        返回
      </t-button>
    </div>
    <t-form
      ref="formRef"
      class="mold-work-order-form"
      :data="formModel"
      :rules="FORM_RULES"
      label-align="top"
      :label-width="100"
      @reset="onReset"
      @submit="onSubmit"
    >
      <div class="form-basic-container">
        <div class="form-basic-item">
          <div class="form-basic-container-title">工单信息</div>

          <!-- 工单基础信息 -->
          <t-row class="row-gap" :gutter="[32, 24]">
            <t-col :span="6">
              <t-form-item label="工单号">
                <t-input v-model="formModel.workCode" placeholder="系统自动生成" disabled />
              </t-form-item>
            </t-col>
            <t-col :span="6">
              <t-form-item label="订单类型" name="orderType">
                <t-select
                  v-model="formModel.orderType"
                  placeholder="请选择订单类型"
                  :disabled="isEdit"
                  @change="handleOrderTypeChange"
                >
                  <t-option
                    v-for="item in orderTypeList"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </t-select>
              </t-form-item>
            </t-col>
            <t-col :span="6">
              <t-form-item label="工单类型" name="workType">
                <t-select
                  v-model="formModel.workType"
                  placeholder="请选择工单类型"
                  @change="handleWorkTypeChange"
                >
                  <t-option label="内部生产" value="IN_HOUSE" />
                  <t-option label="外发加工" value="OUTSOURCED" />
                </t-select>
              </t-form-item>
            </t-col>
            <t-col :span="6">
              <t-form-item label="模具订单" name="moldOrderId">
                <t-select
                  v-model="formModel.moldOrderId"
                  :placeholder="orderSelectPlaceholder"
                  clearable
                  filterable
                >
                  <t-option
                    v-for="order in filteredOrderList"
                    :key="order.id"
                    :label="getOrderLabel(order)"
                    :value="order.id"
                  />
                </t-select>
              </t-form-item>
            </t-col>
            <t-col
              v-if="formModel.orderType === 'REPAIR' || formModel.orderType === 'MODIFICATION'"
              :span="6"
            >
              <t-form-item label="模具" name="moldId">
                <t-select v-model="formModel.moldId" placeholder="请选择模具" :disabled="isEdit" filterable @change="handleMoldChange">
                  <t-option
                    v-for="mold in moldList"
                    :key="mold.id"
                    :label="mold.mold_code"
                    :value="mold.id"
                  />
                </t-select>
              </t-form-item>
            </t-col>
          </t-row>

          <!-- 新模工单 - 新模具信息 -->
          <template v-if="formModel.orderType === 'NEW_MOLD'">
            <div class="form-basic-container-title form-title-gap">模具信息</div>

            <t-row class="row-gap" :gutter="[32, 24]">
              <t-col :span="6">
                <t-form-item label="模具编号" name="newMoldCode">
                  <t-input
                    v-if="!isEdit"
                    v-model="formModel.newMoldCode"
                    placeholder="请输入模具编号"
                  />
                  <t-input
                    v-else
                    v-model="existingMoldInfo.moldCode"
                    disabled
                    :class="{ 'locked-input': isEdit }"
                  />
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="产品名称" name="newMoldName">
                  <t-input
                    v-if="!isEdit"
                    v-model="formModel.newMoldName"
                    placeholder="请输入产品名称"
                  />
                  <t-input
                    v-else
                    v-model="existingMoldInfo.moldName"
                    disabled
                    :class="{ 'locked-input': isEdit }"
                  />
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="客户" name="customerId">
                  <t-select
                    v-if="!isEdit"
                    v-model="formModel.customerId"
                    placeholder="请选择客户"
                    clearable
                  >
                    <t-option
                      v-for="customer in customerList"
                      :key="customer.id"
                      :label="customer.cus_name"
                      :value="customer.id"
                    />
                  </t-select>
                  <t-input
                    v-else
                    v-model="existingMoldInfo.customerName"
                    disabled
                    :class="{ 'locked-input': isEdit }"
                  />
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="模具类型" name="newMoldType">
                  <t-select
                    v-if="!isEdit"
                    v-model="formModel.newMoldType"
                    placeholder="请选择模具类型"
                    clearable
                  >
                    <t-option
                      v-for="item in moldTypeList"
                      :key="item.value"
                      :label="item.label"
                      :value="item.value"
                    />
                  </t-select>
                  <t-input
                    v-else
                    v-model="existingMoldInfo.moldTypeText"
                    disabled
                    :class="{ 'locked-input': isEdit }"
                  />
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="材质类型" name="newMaterialType">
                  <t-select
                    v-if="!isEdit"
                    v-model="formModel.newMaterialType"
                    placeholder="请选择材质类型"
                    clearable
                  >
                    <t-option
                      v-for="item in materialTypeList"
                      :key="item.value"
                      :label="item.label"
                      :value="item.value"
                    />
                  </t-select>
                  <t-input
                    v-else
                    v-model="existingMoldInfo.materialTypeText"
                    disabled
                    :class="{ 'locked-input': isEdit }"
                  />
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="模具工艺" name="newMoldProcess">
                  <t-select
                    v-if="!isEdit"
                    v-model="formModel.newMoldProcess"
                    placeholder="请选择模具工艺"
                    clearable
                  >
                    <t-option
                      v-for="item in moldProcessList"
                      :key="item.value"
                      :label="item.label"
                      :value="item.value"
                    />
                  </t-select>
                  <t-input
                    v-else
                    v-model="existingMoldInfo.moldProcessText"
                    disabled
                    :class="{ 'locked-input': isEdit }"
                  />
                </t-form-item>
              </t-col>
            </t-row>
          </template>

          <!-- 外发信息 -->
          <template v-if="formModel.workType === 'OUTSOURCED'">
            <div class="form-basic-container-title form-title-gap">外发信息</div>

            <t-row class="row-gap" :gutter="[32, 24]">
              <t-col :span="6">
                <t-form-item label="加工商" name="processorId">
                  <t-select v-model="formModel.processorId" placeholder="请选择加工商" clearable>
                    <t-option
                      v-for="processor in processorList"
                      :key="processor.id"
                      :label="processor.name"
                      :value="processor.id"
                    />
                  </t-select>
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="外发单价" name="outsourcePrice">
                  <t-input-number
                    v-model="formModel.outsourcePrice"
                    :min="0"
                    :decimal-places="2"
                    placeholder="请输入外发单价"
                    style="width: 100%"
                  >
                    <template #prefix>¥</template>
                  </t-input-number>
                </t-form-item>
              </t-col>
              <t-col :span="6">
                <t-form-item label="外发总金额" name="outsourceTotalAmount">
                  <t-input-number
                    v-model="formModel.outsourceTotalAmount"
                    :min="0"
                    :decimal-places="2"
                    placeholder="自动计算或手动输入"
                    style="width: 100%"
                  >
                    <template #prefix>¥</template>
                  </t-input-number>
                </t-form-item>
              </t-col>
            </t-row>
          </template>

          <!-- 计划时间和状态 -->
          <div class="form-basic-container-title form-title-gap">计划信息</div>

          <t-row class="row-gap" :gutter="[32, 24]">
            <t-col :span="6">
              <t-form-item label="计划开始时间" name="planStart">
                <t-date-picker
                  v-model="formModel.planStart"
                  placeholder="请选择日期"
                  format="YYYY-MM-DD HH:mm:ss"
                  show-time
                  style="width: 100%"
                />
              </t-form-item>
            </t-col>
            <t-col :span="6">
              <t-form-item label="计划结束时间" name="planEnd">
                <t-date-picker
                  v-model="formModel.planEnd"
                  placeholder="请选择日期"
                  format="YYYY-MM-DD HH:mm:ss"
                  show-time
                  style="width: 100%"
                />
              </t-form-item>
            </t-col>
          </t-row>

          <!-- 备注 -->
          <div class="form-basic-container-title form-title-gap">其他信息</div>

          <t-form-item label="备注">
            <t-textarea v-model="formModel.remark" :height="124" placeholder="请输入备注" />
          </t-form-item>
        </div>
      </div>

      <div class="form-submit-container">
        <div class="form-submit-sub">
          <div class="form-submit-left">
            <t-button theme="primary" class="form-submit-confirm" type="submit">
              {{ isEdit ? '更新' : '确认' }}
            </t-button>
            <t-button type="reset" class="form-submit-cancel" theme="default" variant="base">
              取消
            </t-button>
          </div>
        </div>
      </div>
    </t-form>
  </div>
</template>
<script setup lang="ts">
import type { FormRule, SubmitContext } from 'tdesign-vue-next';
import { MessagePlugin } from 'tdesign-vue-next';
import { computed, onMounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import {
  createMoldWorkOrder,
  getMoldList,
  getMoldOrderList,
  updateMoldWorkOrder,
} from '@/api/mold';
import { getCustomerList, getProcessorList } from '@/api/reference';
import { useDictRaw } from '@/hooks/useDict';

defineOptions({
  name: 'MoldWorkOrderForm',
});

const route = useRoute();
const router = useRouter();
const formRef = ref();
const isEdit = ref(false);

// 客户列表
const customerList = ref<any[]>([]);

// 模具列表
const moldList = ref<any[]>([]);

// 订单列表
const orderList = ref<any[]>([]);

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: orderTypeList } = useDictRaw('mold_order_type', (items) =>
  (items as any[]).map((item: any) => ({
    value: item.dict_value,
    label: item.dict_label,
  })),
);
const { items: moldTypeList } = useDictRaw('mold-type', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1)
    .map((item: any) => ({
      label: item.dict_label,
      value: item.dict_value,
    })),
);
const { items: materialTypeList } = useDictRaw('material-type', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1)
    .map((item: any) => ({
      label: item.dict_label,
      value: item.dict_value,
    })),
);
const { items: moldProcessList } = useDictRaw('mold-process', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1)
    .map((item: any) => ({
      label: item.dict_label,
      value: item.dict_value,
    })),
);
const processorList = ref<any[]>([]); // 加工商列表

// 直接使用语义化字符串，不需要转换函数

// 已有模具信息（编辑模式显示）
const existingMoldInfo = reactive({
  moldCode: '',
  moldName: '',
  customerName: '',
  moldTypeText: '',
  materialTypeText: '',
  moldProcessText: '',
});

// 表单数据
const formModel = reactive({
  id: undefined as number | undefined,
  workCode: '',
  orderType: 'NEW_MOLD',
  workType: 'IN_HOUSE', // 工单类型：IN_HOUSE=内部生产，OUTSOURCED=外发加工
  moldOrderId: undefined as number | undefined,
  moldId: undefined as number | undefined,
  planStart: '',
  planEnd: '',
  remark: '',
  // 新模工单字段
  newMoldCode: '',
  newMoldName: '',
  customerId: undefined as number | undefined,
  newMoldType: undefined as string | undefined,
  newMaterialType: undefined as string | undefined,
  newMoldProcess: undefined as string | undefined,
  newMoldRemark: '',
  // 外发信息字段
  processorId: undefined as number | undefined,
  outsourcePrice: 0,
  outsourceTotalAmount: 0,
});

// 表单验证规则（根据订单类型动态计算）
const FORM_RULES = computed<Record<string, FormRule[]>>(() => {
  const rules: Record<string, FormRule[]> = {
    orderType: [{ required: true, message: '请选择订单类型', type: 'error' }],
  };

  // 新模工单：模具信息必填
  if (formModel.orderType === 'NEW_MOLD') {
    rules.newMoldCode = [
      { required: true, message: '请输入模具编号', type: 'error' },
      { max: 50, message: '模具编号长度不能超过50字符', type: 'warning' },
    ];
    rules.newMoldName = [
      { required: true, message: '请输入产品名称', type: 'error' },
      { max: 100, message: '产品名称长度不能超过100字符', type: 'warning' },
    ];
    rules.customerId = [{ required: true, message: '请选择客户', type: 'error' }];
  }

  // 修模/改模工单：模具必选
  if (formModel.orderType === 'REPAIR' || formModel.orderType === 'MODIFICATION') {
    rules.moldId = [{ required: true, message: '请选择模具', type: 'error' }];
  }

  return rules;
});

// 监听订单类型变化
const handleOrderTypeChange = (value: any) => {
  const newType = value as string;
  if (newType === 'NEW_MOLD') {
    // 新模订单：清空模具选择
    formModel.moldId = undefined;
  } else {
    // 修模/改造订单：清空新模具信息
    formModel.newMoldCode = '';
    formModel.newMoldName = '';
    formModel.customerId = undefined;
    formModel.newMoldType = undefined;
    formModel.newMaterialType = undefined;
    formModel.newMoldRemark = '';
  }
};

const goBack = () => {
  router.back();
};

// 监听工单类型变化
const handleWorkTypeChange = (value: any) => {
  const newType = value as string;
  if (newType === 'OUTSOURCED') {
    // 外发加工：设置默认值
    formModel.outsourcePrice = 0;
    formModel.outsourceTotalAmount = 0;
  } else {
    // 内部生产：清空外发信息
    formModel.processorId = undefined;
    formModel.outsourcePrice = 0;
    formModel.outsourceTotalAmount = 0;
  }
};

// 订单类型名称映射
const orderTypeLabelMap: Record<string, string> = {
  NEW_MOLD: '新模',
  REPAIR: '修模',
  MODIFICATION: '改模',
};

const getOrderLabel = (order: any) => {
  const parts = [order.order_code];
  if (order.customer?.cus_name) {
    parts.push(order.customer.cus_name);
  }
  const typeLabel = orderTypeLabelMap[order.order_type] || order.order_type;
  if (typeLabel) {
    parts.push(typeLabel);
  }
  return parts.join(' - ');
};

// 根据选中的模具过滤订单列表（修模/改模时只能选该模具所属的订单）
const filteredOrderList = computed(() => {
  if (!formModel.moldId) return orderList.value;

  const selectedMold = moldList.value.find((m: any) => m.id === formModel.moldId);
  if (!selectedMold) return orderList.value;

  // 只显示该模具关联的订单：order.id 等于模具的 latest_order_id，或者订单的 mold_codes 包含该模具编号
  return orderList.value.filter((o: any) => {
    // 通过 latest_order_id 匹配
    if (selectedMold.latest_order_id && o.id === selectedMold.latest_order_id) return true;
    // 通过订单的模具编号列表匹配
    if (o.mold_codes && Array.isArray(o.mold_codes) && selectedMold.mold_code) {
      return o.mold_codes.includes(selectedMold.mold_code);
    }
    return false;
  });
});

// 订单下拉提示文字
const orderSelectPlaceholder = computed(() => {
  if (!formModel.moldId) return '请选择订单（可选）';
  if (filteredOrderList.value.length === 0) return '该模具暂无关联订单';
  return '请选择订单（可选）';
});

// 选择模具时，筛选并自动关联该模具的来源订单
const handleMoldChange = (moldId: any) => {
  if (!moldId) {
    formModel.moldOrderId = undefined;
    return;
  }
  // 仅在修模/改模时自动关联
  if (formModel.orderType !== 'REPAIR' && formModel.orderType !== 'MODIFICATION') return;

  // 等 filteredOrderList 更新后（nextTick），如果只有一个关联订单则自动选中
  const selectedMold = moldList.value.find((m: any) => m.id === moldId);
  if (selectedMold?.latest_order_id) {
    formModel.moldOrderId = selectedMold.latest_order_id;
  } else {
    formModel.moldOrderId = undefined;
  }
};

// 提交表单
const onSubmit = async (ctx: SubmitContext) => {
  if (ctx.validateResult === true) {
    try {
      // 手动二次校验（根据订单类型验证必填字段，防止表单验证规则漏过）
      if (formModel.orderType === 'NEW_MOLD') {
        if (!formModel.newMoldCode?.trim()) {
          MessagePlugin.warning('请输入模具编号');
          return;
        }
        if (!formModel.newMoldName?.trim()) {
          MessagePlugin.warning('请输入产品名称');
          return;
        }
        if (!formModel.customerId) {
          MessagePlugin.warning('请选择客户');
          return;
        }
      }

      if (formModel.orderType === 'REPAIR' || formModel.orderType === 'MODIFICATION') {
        if (!formModel.moldId) {
          MessagePlugin.warning('请选择模具');
          return;
        }
      }

      const submitData: any = {
        order_type: formModel.orderType,
        work_type: formModel.workType, // 工单类型
        plan_start: formModel.planStart ? formModel.planStart.replace(' ', 'T') : '',
        plan_end: formModel.planEnd ? formModel.planEnd.replace(' ', 'T') : '',
        remark: formModel.remark,
      };

      if (formModel.moldOrderId) {
        submitData.mold_order_id = formModel.moldOrderId;
      }

      if (formModel.moldId) {
        submitData.mold_id = formModel.moldId;
      }

      // 外发信息
      if (formModel.workType === 'OUTSOURCED') {
        if (formModel.processorId) {
          submitData.processor_id = formModel.processorId;
        }
        submitData.outsource_price = formModel.outsourcePrice || 0;
        submitData.outsource_total_amount = formModel.outsourceTotalAmount || 0;
      }

      // 新模工单的特殊字段
      if (formModel.orderType === 'NEW_MOLD') {
        submitData.new_mold_code = formModel.newMoldCode;
        submitData.new_mold_name = formModel.newMoldName;
        submitData.customer_id = formModel.customerId;
        if (formModel.newMoldType) submitData.new_mold_type = formModel.newMoldType;
        if (formModel.newMaterialType) submitData.new_material_type = formModel.newMaterialType;
        if (formModel.newMoldProcess) submitData.new_mold_process = formModel.newMoldProcess;
        if (formModel.newMoldRemark) submitData.new_mold_remark = formModel.newMoldRemark;
      }

      if (isEdit.value) {
        submitData.id = formModel.id;
        await updateMoldWorkOrder(submitData);
        MessagePlugin.success('更新成功');
      } else {
        await createMoldWorkOrder(submitData);
        MessagePlugin.success('创建成功');

        // 新增成功后立即重置表单，防止 KeepAlive 缓存导致下次进入时显示旧数据
        Object.assign(formModel, {
          id: undefined,
          workCode: '',
          orderType: 'NEW_MOLD',
          workType: 'IN_HOUSE',
          moldOrderId: undefined,
          moldId: undefined,
          planStart: '',
          planEnd: '',
          remark: '',
          newMoldCode: '',
          newMoldName: '',
          customerId: undefined,
          newMoldType: undefined,
          newMaterialType: undefined,
          newMoldProcess: undefined,
          newMoldRemark: '',
          processorId: undefined,
          outsourcePrice: 0,
          outsourceTotalAmount: 0,
        });
      }

      router.push('/mold/work-order');
    } catch (error: any) {
      console.error('提交工单失败:', error);
      if (error?.message) {
        MessagePlugin.error(error.message);
      } else {
        MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
      }
    }
  }
};

const onReset = () => {
  MessagePlugin.warning('取消操作');
  router.push('/mold/work-order');
};

const loadWorkOrderDetail = async (_id: number) => {
  try {
    // TODO: 需要后端提供获取工单详情的API
    // const res = await getMoldWorkOrderById(id);
    MessagePlugin.warning('编辑功能暂未实现，需要先添加获取工单详情的API');
    router.push('/mold/work-order');
  } catch (error) {
    console.error('加载工单详情失败:', error);
    MessagePlugin.error('加载工单详情失败');
  }
};

const loadCustomerList = async () => {
  try {
    const res = await getCustomerList();
    customerList.value = res.list || [];
  } catch (error) {
    console.error('获取客户列表失败:', error);
  }
};

const loadMoldList = async () => {
  try {
    const res = await getMoldList({ page: 1, page_size: 100 });
    moldList.value = res.list || [];
  } catch (error) {
    console.error('获取模具列表失败:', error);
  }
};

const loadOrderList = async () => {
  try {
    const res = await getMoldOrderList({ page: 1, page_size: 100 });
    orderList.value = res.list || [];
  } catch (error) {
    console.error('获取订单列表失败:', error);
  }
};

const loadProcessorList = async () => {
  try {
    const res = await getProcessorList();
    processorList.value = res.list || [];
  } catch (error) {
    console.error('获取加工商列表失败:', error);
  }
};

// 加载字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

onMounted(async () => {
  // 字典已通过 useDictRaw 自动加载

  await Promise.all([loadCustomerList(), loadMoldList(), loadOrderList(), loadProcessorList()]);

  const idParam = route.query.id;
  if (idParam) {
    isEdit.value = true;
    await loadWorkOrderDetail(Number(idParam));
    // 编辑模式：动态更新页面标题
    route.meta.title = { zh_CN: '编辑模具工单', en_US: 'Edit Mold Work Order' };
  }
});
</script>
<style lang="less" scoped>
@import '@/pages/form/base/index.less';

.locked-input {
  :deep(.t-input__inner) {
    color: var(--td-text-color-placeholder) !important;
    background-color: var(--td-bg-color-secondarycontainer) !important;
  }
}

.back-button-container {
  margin-bottom: 16px;
}
</style>

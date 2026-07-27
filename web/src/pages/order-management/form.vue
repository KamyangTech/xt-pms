<template>
  <div class="order-form-container">
    <div class="back-button-container">
      <t-button theme="default" @click="goBack" variant="text">
        <template #icon><t-icon name="chevron-left" /></template>
        返回
      </t-button>
    </div>
    <!-- 步骤条（替代搜索栏） -->
    <t-card :bordered="false" class="step-card">
      <t-steps :current="actualStep" layout="horizontal">
        <t-step-item title="选择模式" content="选择下单类型" />
        <t-step-item title="订单信息" content="填写订单基础信息" />
        
        <!-- 只有 COMBINED 和 MOLD_ONLY 显示模具信息 -->
        <t-step-item 
          v-if="orderType !== 'PROD_ONLY'" 
          title="模具信息" 
          content="填写模具相关明细" 
        />
        
        <!-- 只有 COMBINED 和 PROD_ONLY 显示生产信息 -->
        <t-step-item 
          v-if="orderType !== 'MOLD_ONLY'" 
          title="生产信息" 
          content="填写生产交付要求" 
        />
        
        <!-- 款项配置步骤 -->
        <t-step-item 
          title="款项配置" 
          content="配置分期收款计划" 
        />
        
        <t-step-item title="完成" content="提交成功" />
      </t-steps>
    </t-card>

    <!-- 表单内容（替代数据表格） -->
    <t-card :bordered="false" class="form-card">
      <!-- 步骤1：选择模式 -->
      <div v-show="currentStep === 0" class="step-content">
        <div class="step-title">请选择对应的订单类型</div>
        <t-radio-group v-model="orderType">
          <t-radio-button value="COMBINED">综合下单 (模具+产品)</t-radio-button>
          <t-radio-button value="MOLD_ONLY">仅制造模具</t-radio-button>
          <t-radio-button value="PROD_ONLY">仅生产产品</t-radio-button>
        </t-radio-group>
        
        <!-- 模式提示框 -->
        <t-alert v-if="orderType === 'COMBINED'" theme="info" message="同时创建模具订单和生产订单，适用于需要定制模具并生产产品的场景。" style="margin-top: 16px" />
        <t-alert v-else-if="orderType === 'MOLD_ONLY'" theme="info" message="仅创建模具订单，适用于只需要制造模具的场景。" style="margin-top: 16px" />
        <t-alert v-else theme="info" message="仅创建生产订单，适用于使用已有模具进行产品生产的场景。" style="margin-top: 16px" />
        
        <div class="action-bar">
          <t-button theme="primary" @click="nextStep">下一步</t-button>
        </div>
      </div>

      <!-- 步骤2：订单信息 -->
      <div v-show="currentStep === 1" class="step-content">
        <OrderInfoForm
          ref="orderInfoFormRef"
          @next="handleOrderInfoNext"
          @prev="prevStep"
        />
      </div>

      <!-- 步骤3：模具信息 -->
      <div v-show="currentStep === 2" class="step-content">
        <MoldInfoForm
          ref="moldInfoFormRef"
          :order-type="orderType"
          @next="handleMoldInfoNext"
          @prev="prevStep"
        />
      </div>

      <!-- 步骤4：生产信息 -->
      <div v-show="currentStep === 3" class="step-content">
        <ProductionInfoForm
          ref="productionInfoFormRef"
          :order-type="orderType"
          :prefilled-customer-id="formData.orderInfo.customer_id"
          :prefilled-mold-items="formData.moldInfo.items || []"
          @next="handleProductionInfoNext"
          @prev="prevStep"
        />
      </div>

      <!-- 步骤5：款项配置 -->
      <div v-show="currentStep === 4" class="step-content">
        <PaymentPlanForm
          ref="paymentPlanFormRef"
          :order-type="orderType"
          :mold-items="formData.moldInfo.items || []"
          :production-items="formData.productionInfo.items || []"
          :production-type="formData.productionInfo.production_type"
          :processor-name="formData.productionInfo.processor_name || ''"
          @next="handlePaymentPlanNext"
          @prev="prevStep"
        />
      </div>

      <!-- 步骤6：完成 -->
      <div v-show="currentStep === 5" class="step-content success-page">
        <div class="success-container">
          <!-- 成功图标 -->
          <div class="success-icon">
            <t-icon name="check-circle" size="80" style="color: #00a870;" />
          </div>
          
          <!-- 成功标题 -->
          <div class="success-title">订单创建成功</div>
          
          <!-- 成功提示 -->
          <div class="success-tip">订单已成功提交，您可以查看订单进度或继续创建新订单</div>
          
          <!-- 订单信息（如果有订单号） -->
          <div v-if="createdOrderCode" class="order-info-card">
            <div class="order-info-label">订单编号</div>
            <div class="order-info-value">{{ createdOrderCode }}</div>
          </div>
          
          <!-- 操作按钮 -->
          <div class="success-actions">
            <t-button variant="outline" @click="resetForm">
              继续创建
            </t-button>
            <t-button theme="primary" @click="goToDashboard">
              回到首页
            </t-button>
          </div>
        </div>
      </div>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, computed, watch } from 'vue';
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router';
import { MessagePlugin, DialogPlugin } from 'tdesign-vue-next';
import { createCombinedOrder } from '@/api/order-management';
import { createMoldOrder } from '@/api/mold/index';
import { createProductionOrder } from '@/api/production/order';
import OrderInfoForm from './components/OrderInfoForm.vue';
import MoldInfoForm from './components/MoldInfoForm.vue';
import PaymentPlanForm from './components/PaymentPlanForm.vue';
import ProductionInfoForm from './components/ProductionInfoForm.vue';

const route = useRoute();
const router = useRouter();
const currentStep = ref(0);
const orderType = ref('COMBINED');
const submitting = ref(false);
const createdOrderCode = ref(''); // ✨ 保存创建的订单号

// 根据订单类型计算实际步骤索引
const actualStep = computed(() => {
  if (orderType.value === 'PROD_ONLY') {
    // PROD_ONLY: 0选择, 1订单, 2生产, 3款项, 4完成 (跳过模具)
    if (currentStep.value === 0) return 0;
    if (currentStep.value === 1) return 1;
    if (currentStep.value === 3) return 2; // 生产信息
    if (currentStep.value === 4) return 3; // 款项配置
    if (currentStep.value === 5) return 4; // 完成
    return currentStep.value;
  } else if (orderType.value === 'MOLD_ONLY') {
    // MOLD_ONLY: 0选择, 1订单, 2模具, 4款项, 5完成 (跳过生产)
    if (currentStep.value === 0) return 0;
    if (currentStep.value === 1) return 1;
    if (currentStep.value === 2) return 2; // 模具信息
    if (currentStep.value === 4) return 3; // 款项配置
    if (currentStep.value === 5) return 4; // 完成
    return currentStep.value;
  }
  // COMBINED: 0选择, 1订单, 2模具, 3生产, 4款项, 5完成
  return currentStep.value;
});

// 表单引用
const orderInfoFormRef = ref<any>(null);
const moldInfoFormRef = ref<any>(null);
const paymentPlanFormRef = ref<any>(null);
const productionInfoFormRef = ref<any>(null);

// 表单数据（用于跨步骤传递）
const formData = ref<{
  orderInfo: {
    customer_id?: number;  // ✨ 修复：使用下划线命名，与 OrderInfoForm 一致
    [key: string]: any;
  };
  moldInfo: {
    items?: Array<{
      mold_code: string;
      mold_name: string;
      amount?: number;
      moldProcess?: number | string;
      moldType?: number | string;
      materialType?: number | string;
      cavityCount?: number;
      remark: string;
    }>;
    [key: string]: any;
  };
  productionInfo: {
    items?: any[];
    [key: string]: any;
  };
  paymentPlan?: any; // 款项配置数据
}>({
  orderInfo: {},
  moldInfo: {
    items: [], // 初始化为空数组
  },
  productionInfo: {},
});

const hasUnsavedData = computed(() => {
  // 如果已经到最后一步（完成），说明已经提交成功，不需要提示
  if (currentStep.value === 5) return false;
  
  // 如果还在第一步（选择模式），不需要提示
  if (currentStep.value === 0) return false;
  
  // 其他情况（步骤1-4）说明有填写数据，需要提示
  return true;
});

// 使用浏览器原生的 beforeunload 事件（仅对刷新/关闭浏览器有效）
const handleBeforeUnload = (e: BeforeUnloadEvent) => {
  if (hasUnsavedData.value) {
    e.preventDefault();
    e.returnValue = '';
    return '';
  }
};

onMounted(() => {
  const type = route.query.type as string;
  if (type) orderType.value = type;
  
  window.addEventListener('beforeunload', handleBeforeUnload);
});

// 监听步骤变化
watch(currentStep, (newStep) => {
  if (newStep === 3) {
  }
});

onBeforeUnmount(() => {
  // 移除事件监听
  window.removeEventListener('beforeunload', handleBeforeUnload);
});

// SPA 内部路由跳转的确认
// 使用 onBeforeRouteLeave 返回 Promise
onBeforeRouteLeave((to, from) => {
  // 如果没有未保存数据，直接放行
  if (!hasUnsavedData.value) {
    return true;
  }
  
  // 有未保存数据，返回 Promise 等待用户选择
  return new Promise<boolean>((resolve) => {
    const dialog = DialogPlugin.confirm({
      header: '提示',
      body: '当前订单尚未提交，离开此页面将丢失已填写的数据。确定要离开吗？',
      confirmBtn: '确定离开',
      cancelBtn: '继续填写',
      onConfirm: () => {
        dialog.destroy();
        resolve(true);
      },
      onCancel: () => {
        dialog.destroy();
        resolve(false);
      },
      onClose: () => {
        dialog.destroy();
        resolve(false);
      },
    });
  });
});

const nextStep = () => {
  // 根据模式自动跳转到正确的步骤
  if (currentStep.value === 1 && orderType.value === 'PROD_ONLY') {
    // PROD_ONLY 模式：从订单信息直接跳到生产信息（跳过模具信息）
    currentStep.value = 3;
  } else if (currentStep.value === 2 && orderType.value === 'MOLD_ONLY') {
    // MOLD_ONLY 模式：从模具信息直接跳到款项配置（跳过生产信息）
    currentStep.value = 4;
  } else {
    currentStep.value++;
  }
};

const prevStep = () => {
  // 根据模式返回到正确的步骤
  if (orderType.value === 'PROD_ONLY') {
    // PROD_ONLY 模式：0选择, 1订单, 3生产, 4款项, 5完成 (跳过模具)
    if (currentStep.value === 3) {
      currentStep.value = 1; // 从生产信息返回到订单信息
    } else if (currentStep.value === 4) {
      currentStep.value = 3; // 从款项配置返回到生产信息
    } else if (currentStep.value > 0) {
      currentStep.value--;
    }
  } else if (orderType.value === 'MOLD_ONLY') {
    // MOLD_ONLY 模式：0选择, 1订单, 2模具, 4款项, 5完成 (跳过生产)
    if (currentStep.value === 5) {
      currentStep.value = 4; // 从完成返回到款项配置
    } else if (currentStep.value === 4) {
      currentStep.value = 2; // 从款项配置返回到模具信息
    } else if (currentStep.value > 0) {
      currentStep.value--;
    }
  } else {
    // COMBINED 模式：0选择, 1订单, 2模具, 3生产, 4款项, 5完成
    if (currentStep.value > 0) {
      currentStep.value--;
    }
  }
};

// 订单信息下一步
const handleOrderInfoNext = async (data: any) => {
  formData.value.orderInfo = data;
  nextStep();
};

// ✨ 获取款项配置数据，转换为后端 PaymentPlanDTO 格式
const getPaymentPlanData = () => {
  if (paymentPlanFormRef.value) {
    const raw = paymentPlanFormRef.value.getData(); // { mold_payments, product_payments, outsource_payments }
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

    for (const key of Object.keys(mold_payments || {})) {
      for (const payment of mold_payments[key] || []) {
        if (payment.payment_type && (payment.amount || 0) > 0) {
          items.push({
            index: index++,
            payment_type: payment.payment_type,
            amount_type: 'amount',
            amount_value: payment.amount || 0,
            calculated_amount: payment.amount || 0,
            due_date: payment.due_date || '',
            source_type: 'mold',
          });
        }
      }
    }

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

    // ✨ 转换外发应付款（应付）
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
            source_type: 'outsource', // 外发应付
          });
        }
      }
    }

    // ✨ 转换模具外发应付款（应付）
    for (const key of Object.keys(mold_outsource_payments || {})) {
      for (const payment of mold_outsource_payments[key] || []) {
        if (payment.payment_type && (payment.amount || 0) > 0) {
          items.push({
            index: index++,
            payment_type: payment.payment_type,
            amount_type: 'amount',
            amount_value: payment.amount || 0,
            calculated_amount: payment.amount || 0,
            due_date: payment.due_date || '',
            source_type: 'outsource', // 模具外发应付
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

// 模具信息下一步
const handleMoldInfoNext = async (data: any) => {
  formData.value.moldInfo = data;
  
  // MOLD_ONLY 模式：模具信息之后直接进入款项配置
  // COMBINED 模式：模具信息之后进入生产信息
  nextStep();
};

// 款项配置下一步
const handlePaymentPlanNext = async (data: any) => {
  formData.value.paymentPlan = data;

  // 提交订单
  await handleSubmitOrder();
};

// 生产信息下一步（进入款项配置）
const handleProductionInfoNext = async (data: any) => {
  formData.value.productionInfo = data;
  
  // 进入款项配置步骤
  nextStep();
};

const handleSubmitOrder = async () => {
  submitting.value = true;
  
  try {
    // ✨ 最终提交时，先上传所有待处理的图片
    if (productionInfoFormRef.value?.uploadPendingImages) {
      const ok = await productionInfoFormRef.value.uploadPendingImages();
      if (!ok) {
        submitting.value = false;
        return;
      }
    }

    // 根据订单类型构建请求数据
    let requestData: any = {};
    
    // ✨ 确保 customerId 是数字类型（使用 customer_id 字段）
    const customerId = Number(formData.value.orderInfo.customer_id);
    
    // ✨ 验证 customerId 是否有效
    if (!customerId || isNaN(customerId) || customerId <= 0) {
      MessagePlugin.error('客户信息无效，请返回订单信息步骤重新选择客户');
      submitting.value = false;
      return;
    }
    
    //  辅助函数：计算生产订单总金额（单价 × 数量 + 包装费 + 运费）
    const calculateProductionTotalAmount = () => {
      if (!formData.value.productionInfo.items) return 0;
      return formData.value.productionInfo.items.reduce((sum: number, item: any) => {
        const productAmount = (item.unit_price || 0) * (item.plan_qty || 1);
        const packagingFee = item.packaging_fee || 0;
        const freightFee = item.freight_fee || 0;
        return sum + productAmount + packagingFee + freightFee;
      }, 0);
    };

    // ✨ 辅助函数：计算生产订单总计划数量
    const calculateProductionTotalQty = () => {
      if (!formData.value.productionInfo.items) return 0;
      return formData.value.productionInfo.items.reduce((sum: number, item: any) => {
        return sum + (item.plan_qty || 0);
      }, 0);
    };

    // ✨ 辅助函数：从各产品外发单价取最大值作为订单级外发单价
    const calculateMaxOutsourcePrice = () => {
      if (!formData.value.productionInfo.items) return 0;
      return formData.value.productionInfo.items.reduce((max: number, item: any) => {
        return Math.max(max, item.outsource_price || 0);
      }, 0);
    };
    
    if (orderType.value === 'COMBINED') {
      // 综合订单：模具 + 产品
      requestData = {
        master: {
          customer_id: customerId,
          order_type: 'COMBINED',
          remark: formData.value.orderInfo.remark || '',
          total_amount: (formData.value.moldInfo.totalAmount || 0) + calculateProductionTotalAmount(), // 计算总金额
        },
        mold_order: {
          customer_id: customerId,
          plan_start_date: formData.value.orderInfo.plan_start_date || null,  // ✨ 添加计划开始日期
          plan_end_date: formData.value.orderInfo.plan_end_date || null,      // ✨ 添加计划结束日期
          remark: formData.value.orderInfo.mold_order_remark || '',           // ✨ 模具订单备注
          priority: formData.value.moldInfo.priority || 'NORMAL',             // ✨ 传递优先级
          items: (formData.value.moldInfo.items || []).map((item: any) => ({
            mold_code: item.mold_code,
            mold_name: item.mold_name,
            amount: item.amount || 0,
            mold_process: item.moldProcess,
            mold_type: item.moldType,
            material_type: item.materialType,
            cavity_count: item.cavityCount,
            remark: item.remark || '',
            mold_id: 0, //  确保 mold_id 为 0，触发后端创建新模具档案和工单
            work_type: item.work_type || 'IN_HOUSE',
            processor_id: item.processor_id || null,
          })),
          total_amount: formData.value.moldInfo.totalAmount || 0,
        },
        prod_order: {
          customer_id: customerId,  // ✨ 修复：使用已验证的 customerId 变量
          production_type: formData.value.productionInfo.production_type || 'IN_HOUSE',          // ✨ 生产类型
          mold_source_type: formData.value.productionInfo.mold_source_type || 'SELF_MADE',        // ✨ 模具来源
          plan_qty: calculateProductionTotalQty(), // ✨ 计算总计划数量
          priority: formData.value.productionInfo.priority || 'NORMAL',                           // ✨ 优先级
          // ✨ 外发相关字段（从 PaymentPlanForm 的 items 中获取）
          processor_id: formData.value.productionInfo.production_type === 'OUTSOURCED' ? (formData.value.productionInfo.processor_id || null) : null, // ✨ 加工商ID
          outsource_price: calculateMaxOutsourcePrice(), // ✨ 各产品外发单价的最大值
          items: (formData.value.productionInfo.items || []).map((item: any) => {
            // ✨ 过滤 mold_ids，只保留有效的数字 ID
            const validMoldIds = (item.mold_ids || [])
              .filter((id: any) => id !== undefined && id !== null && id !== '')
              .map((id: any) => {
                const numId = Number(id);
                return isNaN(numId) ? null : numId;
              })
              .filter((id: number | null) => id !== null) as number[];
            
            return {
              product_name: item.product_name,
              product_image_url: item.product_image_url || '', // ✨ 添加产品图片
              mold_id: item.mold_id ? Number(item.mold_id) : null, // ✨ 转换为数字类型
              mold_ids: validMoldIds, // ✨ 只发送有效的数字 ID
              external_mold_code: item.external_mold_code || null, // ✨ 添加外来模具编号
              plan_qty: item.plan_qty || 1,
              outsource_price: item.outsource_price || 0, // ✨ 传递每个产品的外发单价
              device_id: item.device_id ? Number(item.device_id) : null, // ✨ 转换为数字类型
              process_remark: item.process_params || '', // ✨ 修正字段映射：process_params -> process_remark
              remark: item.remark || '',
              // ✨ 款项配置字段（金额计算）
              unit_price: item.unit_price || 0,
              packaging_fee: item.packaging_fee || 0,
              freight_fee: item.freight_fee || 0,
            };
          }),
        },
        // ✨ 添加款项配置
        payment_plan: getPaymentPlanData(),
      };
    } else if (orderType.value === 'MOLD_ONLY') {
      // 仅模具订单 → 调用独立模具订单 API，不会生成综合订单号
      requestData = {
        customer_id: customerId,
        order_type: 'NEW_MOLD',
        priority: formData.value.moldInfo.priority || 'NORMAL',
        plan_start_date: formData.value.orderInfo.plan_start_date || null,
        plan_end_date: formData.value.orderInfo.plan_end_date || null,
        remark: formData.value.orderInfo.remark || '',
        total_amount: formData.value.moldInfo.totalAmount || 0,
        items: (formData.value.moldInfo.items || []).map((item: any) => ({
          mold_code: item.mold_code,
          mold_name: item.mold_name,
          amount: item.amount || 0,
          mold_process: item.moldProcess,
          mold_type: item.moldType,
          material_type: item.materialType,
          cavity_count: item.cavityCount,
          remark: item.remark || '',
          work_type: item.work_type || 'IN_HOUSE',
          processor_id: item.processor_id || null,
          outsource_price: item.work_type === 'OUTSOURCED' ? item.outsource_price : undefined,
        })),
        payment_plan: getPaymentPlanData(),
      };
      const response = await createMoldOrder(requestData);
      createdOrderCode.value = response?.order_code || '';
    } else if (orderType.value === 'PROD_ONLY') {
      // 仅生产订单 → 调用独立生产订单 API，不会生成综合订单号
      const items = (formData.value.productionInfo.items || []).map((item: any) => {
        const validMoldIds = (item.mold_ids || [])
          .filter((id: any) => id !== undefined && id !== null && id !== '')
          .map((id: any) => {
            const numId = Number(id);
            return isNaN(numId) ? null : numId;
          })
          .filter((id: number | null) => id !== null) as number[];
        return {
          product_name: item.product_name,
          product_image_url: item.product_image_url || '',
          mold_id: item.mold_id ? Number(item.mold_id) : null,
          mold_ids: validMoldIds,
          external_mold_code: item.external_mold_code || null,
          plan_qty: item.plan_qty || 1,
          outsource_price: item.outsource_price || 0,
          device_id: item.device_id ? Number(item.device_id) : null,
          process_remark: item.process_params || '',
          remark: item.remark || '',
          // ✨ 款项配置字段（金额计算）
          unit_price: item.unit_price || 0,
          packaging_fee: item.packaging_fee || 0,
          freight_fee: item.freight_fee || 0,
        };
      });
      requestData = {
        customer_id: customerId,
        production_type: formData.value.productionInfo.production_type || 'IN_HOUSE',
        mold_source_type: formData.value.productionInfo.mold_source_type || 'SELF_MADE',
        product_name: '',
        plan_start_date: formData.value.orderInfo.plan_start_date || null,
        plan_end_date: formData.value.orderInfo.plan_end_date || null,
        plan_qty: calculateProductionTotalQty(),
        priority: formData.value.productionInfo.priority || 'NORMAL',
        processor_id: formData.value.productionInfo.production_type === 'OUTSOURCED' ? (formData.value.productionInfo.processor_id || null) : null,
        outsource_price: calculateMaxOutsourcePrice(),
        remark: formData.value.orderInfo.remark || '',
        unit_price: calculateProductionTotalQty() > 0 ? calculateProductionTotalAmount() / calculateProductionTotalQty() : 0,
        items,
      };
      const prodResponse: any = await createProductionOrder(requestData);
      createdOrderCode.value = prodResponse?.order_code || '';
    }
    
    if (orderType.value === 'COMBINED') {
      const response = await createCombinedOrder(requestData);
      createdOrderCode.value = response?.order_code || '';
    }
    
    MessagePlugin.success('订单创建成功');
    nextStep();
  } catch (error: any) {
    console.error('订单创建失败:', error);
    MessagePlugin.error(error.message || error.response?.data?.message || '订单创建失败');
  } finally {
    submitting.value = false;
  }
};

const goToList = () => {
  router.push('/order-management/list');
};

const goBack = () => {
  router.back();
};

const goToDashboard = () => {
  router.push('/dashboard');
};

const resetForm = () => {
  currentStep.value = 0;
  orderType.value = 'COMBINED';
  createdOrderCode.value = '';
  formData.value = {
    orderInfo: {},
    moldInfo: {},
    productionInfo: {},
  };
};
</script>

<style scoped>
.order-form-container {
  padding: 16px;
  background-color: var(--td-bg-color-page, #f5f5f5);
  min-height: 100%;
}

.step-card {
  padding: 18px 0 0 0;
  margin-bottom: 16px;
}

.form-card {
  min-height: 500px;
  background-color: var(--td-bg-color-container, #fff);
}

.step-content {
  min-height: 400px;
  /* 确保每个步骤内容都是独立的块级元素 */
  display: block;
  padding: 16px 0;
}

.step-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  margin-bottom: 24px;
}

.action-bar {
  margin-top: 24px;
  text-align: left;
}

.success-step {
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ✨ 成功页面样式 - 参考 TDesign 成功页 */
.success-page {
  min-height: 500px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.success-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
}

.success-icon {
  margin-bottom: 24px;
  animation: scaleIn 0.3s ease-out;
}

@keyframes scaleIn {
  from {
    transform: scale(0);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

.success-title {
  font-size: 20px;
  font-weight: 500;
  color: var(--td-text-color-primary);
  margin-bottom: 8px;
  text-align: center;
}

.success-tip {
  font-size: 14px;
  color: var(--td-text-color-secondary);
  margin-bottom: 24px;
  text-align: center;
  line-height: 1.6;
}

.order-info-card {
  background-color: var(--td-bg-color-container);
  border: 1px solid var(--td-component-border);
  border-radius: 6px;
  padding: 16px 24px;
  margin-bottom: 32px;
  text-align: center;
  min-width: 300px;
}

.order-info-label {
  font-size: 12px;
  color: var(--td-text-color-secondary);
  margin-bottom: 8px;
}

.order-info-value {
  font-size: 18px;
  font-weight: 500;
  color: var(--td-brand-color);
  font-family: 'Courier New', monospace;
}

.success-actions {
  display: flex;
  gap: 12px;
}

.success-actions .t-button {
  min-width: 120px;
}

/* 步骤1选择模式的样式优化 */
:deep(.t-radio-group) {
  display: inline-block;
}

.back-button-container {
  margin-bottom: 16px;
}
</style>

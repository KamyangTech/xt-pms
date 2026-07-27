<template>
  <div class="mold-order-payment-plan">
    <t-card title="模具订单款项配置" :bordered="false">
      <!-- 订单信息 -->
      <div v-if="orderDetail" class="order-info">
        <t-descriptions :column="3" bordered>
          <t-descriptions-item label="订单编号">{{ orderDetail.order_code }}</t-descriptions-item>
          <t-descriptions-item label="客户名称">{{ orderDetail.customer?.cus_name || '-' }}</t-descriptions-item>
          <t-descriptions-item label="订单总额">¥{{ formatAmount(orderDetail.total_amount) }}</t-descriptions-item>
        </t-descriptions>
      </div>

      <!-- 模具/产品列表 -->
      <div class="items-section">
        <div class="section-header">
          <span class="title">模具/产品列表</span>
          <t-button theme="primary" @click="handleSaveAll" :loading="saving">
            <template #icon><save-icon /></template>
            保存所有配置
          </t-button>
        </div>

        <t-table
          :data="items"
          :columns="itemColumns"
          row-key="id"
          bordered
          hover
          :loading="loading"
        >
          <template #mold_info="{ row }">
            <div class="mold-info">
              <div class="mold-code">{{ row.mold?.mold_code || '-' }}</div>
              <div class="mold-name">{{ row.mold?.mold_name || '-' }}</div>
            </div>
          </template>

          <template #amount="{ row }">
            <span class="amount-text">¥{{ formatAmount(row.amount || 0) }}</span>
          </template>

          <template #payment_configured="{ row }">
            <t-tag v-if="row.payment_items && row.payment_items.length > 0" theme="success">
              已配置 ({{ row.payment_items.length }}项)
            </t-tag>
            <t-tag v-else theme="default">未配置</t-tag>
          </template>

          <template #op="{ row }">
            <t-space>
              <t-link theme="primary" @click="handleConfigPayment(row)">配置款项</t-link>
              <t-link theme="primary" @click="handleViewPayments(row)">查看款项</t-link>
            </t-space>
          </template>
        </t-table>
      </div>
    </t-card>

    <!-- 款项配置对话框 -->
    <t-dialog
      v-model:visible="configDialogVisible"
      :header="`配置款项 - ${currentItem?.mold?.mold_code || ''}`"
      width="1200px"
      :footer="null"
      @close="handleDialogClose"
     :close-on-overlay-click="false">
      <div v-if="currentItem" class="config-container">
        <div class="config-header">
          <t-alert
            theme="info"
            message="您可以为该模具单独配置多笔款项，款项总额应等于该模具的金额"
            style="margin-bottom: 16px"
          />
        </div>

        <PaymentPlanConfig
          ref="paymentPlanRef"
          :total-amount="currentItem.amount || 0"
          :default-items="currentItem.payment_items || []"
          @change="handlePaymentPlanChange"
        />

        <div class="dialog-footer">
          <t-space>
            <t-button @click="handleDialogClose">取消</t-button>
            <t-button theme="primary" @click="handleSaveItem" :loading="saving">
              <template #icon><save-icon /></template>
              保存配置
            </t-button>
          </t-space>
        </div>
      </div>
    </t-dialog>

    <!-- 查看款项对话框 -->
    <t-dialog
      v-model:visible="viewDialogVisible"
      :header="`款项明细 - ${currentItem?.mold?.mold_code || ''}`"
      width="1000px"
      :footer="null"
     :close-on-overlay-click="false">
      <div v-if="currentItem" class="view-container">
        <PaymentList 
          v-if="currentWorkOrderId"
          :work-order-id="currentWorkOrderId" 
          @refresh="loadItems"
        />
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { MessagePlugin } from 'tdesign-vue-next';
import { SaveIcon } from 'tdesign-icons-vue-next';
import { getMoldOrderById } from '@/api/mold';
import PaymentPlanConfig from '@/components/PaymentPlanConfig.vue';
import PaymentList from '@/components/order/PaymentList.vue';
import { getPaymentsByWorkOrderId, replacePaymentsByWorkOrder } from '@/api/mold/payment';
import { getMoldWorkOrdersByOrderId } from '@/api/mold/work-order';

defineOptions({
  name: 'MoldOrderPaymentPlan',
});

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const saving = ref(false);
const orderDetail = ref<any>(null);
const items = ref<any[]>([]);
const configDialogVisible = ref(false);
const viewDialogVisible = ref(false);
const currentItem = ref<any>(null);
const currentWorkOrderId = ref<number | null>(null);
const paymentPlanRef = ref();
const currentPaymentPlanData = ref<any>(null);

// 表格列配置
const itemColumns = [
  { colKey: 'index', title: '序号', width: 80 },
  { colKey: 'mold_info', title: '模具信息', width: 250 },
  { colKey: 'amount', title: '金额', width: 150 },
  { colKey: 'payment_configured', title: '款项配置', width: 150 },
  { colKey: 'remark', title: '备注', ellipsis: true },
  { colKey: 'op', title: '操作', width: 200, fixed: 'right' as const },
];

const loadOrderData = async () => {
  const orderId = route.query.orderId as string;
  if (!orderId) {
    MessagePlugin.error('缺少订单ID参数');
    router.back();
    return;
  }

  loading.value = true;
  try {
    const orderRes = await getMoldOrderById(Number(orderId));
    orderDetail.value = orderRes;

    const workOrdersRes = await getMoldWorkOrdersByOrderId(Number(orderId));
    const workOrders = workOrdersRes.list || [];

    items.value = workOrders.map((wo: any, index: number) => ({
      id: wo.id,
      index: index + 1,
      mold: wo.mold,
      amount: wo.outsource_price || orderDetail.value?.total_amount / workOrders.length || 0,
      remark: wo.remark || '',
      payment_items: [] as any[], // 将从款项API加载
      work_order_id: wo.id,
    }));

    // 为每个工单加载款项配置
    await loadAllPaymentItems();
  } catch (error: any) {
    console.error('加载订单数据失败:', error);
    MessagePlugin.error(error.message || '加载订单数据失败');
  } finally {
    loading.value = false;
  }
};

const loadAllPaymentItems = async () => {
  for (const item of items.value) {
    try {
      const payments = await getPaymentsByWorkOrderId(item.work_order_id);
      item.payment_items = payments || [];
    } catch (error) {
      console.error(`加载工单 ${item.work_order_id} 的款项失败:`, error);
      item.payment_items = [];
    }
  }
};

// 配置款项
const handleConfigPayment = (item: any) => {
  currentItem.value = item;
  currentWorkOrderId.value = item.work_order_id;
  configDialogVisible.value = true;
};

// 查看款项
const handleViewPayments = (item: any) => {
  currentItem.value = item;
  currentWorkOrderId.value = item.work_order_id;
  viewDialogVisible.value = true;
};

const handlePaymentPlanChange = (data: any) => {
  currentPaymentPlanData.value = data;
};

const handleSaveItem = async () => {
  if (!paymentPlanRef.value || !currentPaymentPlanData.value) {
    MessagePlugin.warning('请先配置款项');
    return;
  }

  if (!paymentPlanRef.value.validate()) {
    return;
  }

  saving.value = true;
  try {
    const { enabled, items: paymentItems } = currentPaymentPlanData.value;

    if (!enabled || !paymentItems || paymentItems.length === 0) {
      MessagePlugin.warning('未启用款项配置或没有款项数据');
      return;
    }

    const workOrderId = currentWorkOrderId.value;
    if (!workOrderId) {
      MessagePlugin.error('工单ID不存在');
      return;
    }

    // 批量替换该工单的所有款项记录（事务保护）
    await replacePaymentsByWorkOrder({
      order_id: orderDetail.value.id,
      work_order_id: workOrderId,
      items: paymentItems.map((item: any) => ({
        payment_type: item.payment_type,
        contract_amount: item.calculated_amount,
        should_amount: item.calculated_amount,
        actual_amount: 0,
        status: 1,
        remark: `到期日: ${item.due_date || '-'}`,
      })),
    });

    MessagePlugin.success('保存成功');
    configDialogVisible.value = false;
    
    // 重新加载数据
    await loadOrderData();
  } catch (error: any) {
    console.error('保存款项配置失败:', error);
    MessagePlugin.error(error.message || '保存失败');
  } finally {
    saving.value = false;
  }
};

const handleSaveAll = async () => {
  MessagePlugin.info('请逐个为每个模具配置款项后保存');
};

// 重新加载数据
const loadItems = async () => {
  await loadOrderData();
};

// 关闭对话框
const handleDialogClose = () => {
  configDialogVisible.value = false;
  currentItem.value = null;
  currentPaymentPlanData.value = null;
};

// 格式化金额
const formatAmount = (amount: number) => {
  return amount.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

onMounted(() => {
  loadOrderData();
});
</script>

<style scoped lang="less">
.mold-order-payment-plan {
  padding: 16px;

  .order-info {
    margin-bottom: 24px;
  }

  .items-section {
    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      .title {
        font-size: 16px;
        font-weight: 500;
        color: var(--td-text-color-primary);
      }
    }

    .mold-info {
      .mold-code {
        font-weight: 500;
        color: var(--td-text-color-primary);
      }

      .mold-name {
        font-size: 12px;
        color: var(--td-text-color-secondary);
        margin-top: 4px;
      }
    }

    .amount-text {
      font-weight: 500;
      color: var(--td-success-color);
    }
  }

  .config-container {
    .config-header {
      margin-bottom: 16px;
    }

    .dialog-footer {
      margin-top: 24px;
      text-align: right;
    }
  }

  .view-container {
    min-height: 300px;
  }
}
</style>

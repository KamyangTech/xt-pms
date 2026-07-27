<template>
  <div class="mold-order-container">
    <search-filter
      :filters="filterConfig"
      :model-value="formData"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 操作栏 -->
    <t-card class="toolbar-card" :bordered="false">
      <div class="toolbar-content">
        <span class="toolbar-title">模具订单管理</span>
        <div class="toolbar-actions">
          <t-button theme="default" @click="handleRefresh">
            <template #icon><refresh-icon /></template>
            刷新
          </t-button>
          <t-button theme="default" @click="handleColumnSetting">
            <template #icon><setting-icon /></template>
            列设置
          </t-button>
          <t-button theme="primary" @click="handleAdd">
            <template #icon><add-icon /></template>
            新增订单
          </t-button>
        </div>
      </div>
    </t-card>

    <!-- 数据表格 -->
    <t-card :bordered="false">
      <t-table
        :data="tableData"
        :columns="displayColumns"
        :loading="loading"
        row-key="id"
        :pagination="pagination"
        @page-change="rehandlePageChange"
      >
        <template #order_code="{ row }">
          <t-link theme="default" @click="handleViewDetail(row)">
            {{ row.order_code }}
          </t-link>
        </template>
        <template #order_type="{ row }">
          <t-tag shape="round" variant="outline" :theme="getOrderTypeTheme(row.order_type)">
            {{ getOrderTypeText(row.order_type) }}
          </t-tag>
        </template>
        <template #master_order_code="{ row }">
          <span
            v-if="row.master_order_code"
            style="color: #0052d9; cursor: pointer"
            @click="handleViewMasterOrder(row)"
          >
            {{ row.master_order_code }}
          </span>
          <span v-else style="color: #999">-</span>
        </template>
        <template #priority="{ row }">
          <template v-for="item in priorityList" :key="item.value">
            <t-tag
              v-if="row.priority === item.value"
              shape="round"
              :theme="item.theme || 'default'"
              :style="item.themeStyle"
              variant="outline"
            >
              {{ item.label }}
            </t-tag>
          </template>
          <span v-if="!priorityList.find((item) => item.value === row.priority)">-</span>
        </template>
        <template #mold_codes="{ row }">
          <div v-if="row.mold_codes && row.mold_codes.length > 0" class="mold-codes-list">
            <div v-for="(code, index) in row.mold_codes" :key="index" class="mold-code-item">
              <t-link @click="handleViewMoldWorkOrders(row, code)">
                {{ code }}
              </t-link>
            </div>
          </div>
          <span v-else>-</span>
        </template>
        <template #status="{ row }">
          <t-tag
            shape="round"
            :theme="getStatusTheme(row.status)"
            :style="getStatusStyle(row.status)"
          >
            {{ getStatusText(row.status) }}
          </t-tag>
        </template>
        <template #plan_start_date="{ row }">
          {{ formatDate(row.plan_start_date) }}
        </template>
        <template #plan_end_date="{ row }">
          {{ formatDate(row.plan_end_date) }}
        </template>
        <template #deliver_time="{ row }">
          {{ formatDate(row.deliver_time) }}
        </template>
        <template #totalAmount="{ row }"> ¥{{ row.total_amount?.toFixed(2) || '0.00' }} </template>
        <template #operation="{ row }">
          <div style="width: 220px">
            <t-space :size="8" break-line>
              <t-link theme="primary" @click="handleViewDetail(row)">详情</t-link>
              <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
              <t-link theme="success" @click="handlePaymentManage(row)">款项</t-link>
              <t-link v-if="canDeliver(row)" theme="warning" @click="handleDeliver(row)">
                交付
              </t-link>
              <t-popconfirm content="确认删除该订单？" @confirm="handleDelete(row)">
                <t-link theme="danger">删除</t-link>
              </t-popconfirm>
            </t-space>
          </div>
        </template>
      </t-table>
    </t-card>

    <!-- 新增/编辑对话框 -->
    <t-dialog
      v-model:visible="dialogVisible"
      :header="dialogTitle"
      width="800px"
      @confirm="handleSubmit"
      @cancel="handleCancel"
      @opened="handleDialogOpened"
     :close-on-overlay-click="false">
      <t-form
        ref="formRef"
        :model="formModel"
        :rules="formRules"
        :label-width="100"
        style="margin-top: 16px"
      >
        <t-row :gutter="24" style="margin-left: 0; margin-right: 0">
          <t-col :span="12">
            <t-form-item label="订单编号">
              <t-input placeholder="订单号将在创建后自动生成（格式：D-YYYYMMDD-NNN）" disabled />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="订单类型" name="orderType" style="margin-bottom: 16px">
              <t-select
                v-model="formModel.orderType"
                placeholder="请选择订单类型"
                :disabled="isEdit"
              >
                <t-option
                  v-for="type in orderTypeList"
                  :key="type.value"
                  :label="`${type.label}订单`"
                  :value="Number(type.value)"
                />
              </t-select>
            </t-form-item>
          </t-col>
          <!-- 一单多款：模具明细表 -->
          <template v-if="!isEdit">
            <t-col :span="24" style="margin-bottom: 16px">
              <div class="mold-items-header">
                <t-divider align="left" class="section-title">模具明细</t-divider>
                <t-button theme="primary" variant="outline" size="small" @click="handleAddItem">
                  <template #icon><add-icon /></template>
                  添加模具
                </t-button>
              </div>
            </t-col>
            <t-col :span="24">
              <div class="mold-items-container">
                <t-table
                  :data="formModel.items"
                  :columns="itemColumns"
                  row-key="index"
                  :pagination="{ pageSize: 5 }"
                  bordered
                >
                  <template #mold_code="{ row }">
                    <t-input v-model="row.mold_code" placeholder="请输入模具编号" />
                  </template>
                  <template #mold_name="{ row }">
                    <t-input v-model="row.mold_name" placeholder="请输入产品名称" />
                  </template>
                  <template #amount="{ row }">
                    <t-input-number
                      v-model="row.amount"
                      :min="0"
                      :precision="2"
                      placeholder="金额"
                    />
                  </template>
                  <template #operation="{ rowIndex }">
                    <t-link theme="danger" @click="handleRemoveItem(rowIndex)">删除</t-link>
                  </template>
                </t-table>
              </div>
            </t-col>
          </template>
          <t-col :span="12">
            <t-form-item label="计划开始日期" name="planStartDate" style="margin-bottom: 16px">
              <t-date-picker
                v-model="formModel.planStartDate"
                placeholder="请选择日期"
                format="YYYY-MM-DD"
              />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="订单交期" name="planEndDate" style="margin-bottom: 16px">
              <t-date-picker
                v-model="formModel.planEndDate"
                placeholder="请选择日期"
                format="YYYY-MM-DD"
              />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="订单总金额" name="totalAmount" style="margin-bottom: 16px">
              <t-input-number
                v-model="formModel.totalAmount"
                placeholder="0.00"
                :min="0"
                :precision="2"
                style="width: 100%"
              />
            </t-form-item>
          </t-col>
          <t-col :span="12">
            <t-form-item label="订单状态" name="status" style="margin-bottom: 16px">
              <t-select v-model="formModel.status" placeholder="请选择状态">
                <t-option label="未开始" value="NOT_STARTED" />
                <t-option label="加工中" value="PROCESSING" />
                <t-option label="待试模" value="PENDING_TEST" />
                <t-option label="已试模" value="TESTED" />
                <t-option label="已完成" value="COMPLETED" />
                <t-option label="已取消" value="CANCELLED" />
                <t-option label="已交付" value="DELIVERED" />
              </t-select>
            </t-form-item>
          </t-col>
          <t-col :span="24">
            <t-form-item label="备注" name="remark">
              <t-textarea v-model="formModel.remark" placeholder="请输入备注" :rows="3" />
            </t-form-item>
          </t-col>
        </t-row>
      </t-form>
    </t-dialog>

    <!-- 款项管理对话框 -->
    <t-dialog
      v-model:visible="paymentDialogVisible"
      title="款项管理"
      width="1200px"
      :footer="null"
      destroy-on-close
     :close-on-overlay-click="false">
      <div v-if="currentOrder" class="payment-dialog-content">
        <!-- 应收信息 -->
        <t-card :bordered="true" title="应收信息" class="payment-section">
          <t-descriptions :column="2" bordered>
            <t-descriptions-item label="客户名称">
              {{ currentOrder.customer?.cus_name || '-' }}
            </t-descriptions-item>
            <t-descriptions-item label="订单编号">
              <span style="white-space: nowrap; font-weight: 500">{{
                currentOrder.order_code || '-'
              }}</span>
            </t-descriptions-item>
            <t-descriptions-item label="订单金额">
              ¥{{ (currentOrder.total_amount || 0).toFixed(2) }}
            </t-descriptions-item>
            <t-descriptions-item label="订单状态">
              <t-tag
                shape="round"
                :theme="getStatusTheme(currentOrder.status)"
                :style="getStatusStyle(currentOrder.status)"
              >
                {{ getStatusText(currentOrder.status) }}
              </t-tag>
            </t-descriptions-item>
          </t-descriptions>
        </t-card>

        <!-- 资金汇总 -->
        <t-card :bordered="true" class="payment-section">
          <div class="summary-stats">
            <div class="stat-item">
              <div class="stat-label">应收总额</div>
              <div class="stat-value">¥{{ formatAmount(receivableTotal) }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">已收金额</div>
              <div class="stat-value success">¥{{ formatAmount(receivableReceived) }}</div>
            </div>
            <div class="stat-item">
              <div class="stat-label">余额</div>
              <div class="stat-value warning">¥{{ formatAmount(receivableBalance) }}</div>
            </div>
          </div>
        </t-card>

        <!-- 应收明细 -->
        <t-card :bordered="true" title="应收明细" class="payment-section">
          <t-table
            :data="receivableList"
            :columns="receivableColumns"
            row-key="id"
            hover
            stripe
            :loading="receivableLoading"
          >
            <template #payment_type="{ row }">
              <t-tag
                size="small"
                :theme="getPaymentTypeTheme(row.payment_type)"
                :style="getPaymentTypeStyle(row.payment_type)"
              >
                {{ getPaymentTypeLabel(row.payment_type) }}
              </t-tag>
            </template>
            <template #amount="{ row }"> ¥{{ formatAmount(row.amount) }} </template>
            <template #received_amount="{ row }">
              <span class="text-success">¥{{ formatAmount(row.received_amount) }}</span>
            </template>
            <template #discount_amount="{ row }">
              <span v-if="row.discount_amount > 0" style="color: var(--td-warning-color)"
                >-¥{{ formatAmount(row.discount_amount) }}</span
              >
              <span v-else>-</span>
            </template>
            <template #balance="{ row }">
              <span :class="row.balance > 0 ? 'text-warning' : 'text-success'"
                >¥{{ formatAmount(row.balance) }}</span
              >
            </template>
            <template #due_date="{ row }">
              {{ row.due_date ? formatDate(row.due_date) : '-' }}
            </template>
            <template #status="{ row }">
              <t-tag size="small" :theme="getReceivableStatusTheme(row.status)" :style="getReceivableStatusStyle(row.status)">
                {{ getReceivableStatusText(row.status) }}
              </t-tag>
            </template>
            <template #op="{ row }">
              <t-link theme="primary" @click="handleViewReceivablePayments(row)">收款明细</t-link>
            </template>
          </t-table>
          <t-empty
            v-if="!receivableLoading && receivableList.length === 0"
            description="暂无应收款项记录"
          />
        </t-card>
      </div>
    </t-dialog>

    <!-- 订单交付对话框 -->
    <t-dialog
      v-model:visible="deliverDialogVisible"
      header="模具订单交付"
      width="600px"
      @confirm="handleConfirmDeliver"
      @cancel="deliverDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form ref="deliverFormRef" :model="deliverForm" :label-width="100" style="margin-top: 16px">
        <t-form-item label="订单编号">
          <t-input :value="deliverForm.orderCode" disabled />
        </t-form-item>
        <t-form-item label="客户">
          <t-input :value="deliverForm.customerName" disabled />
        </t-form-item>
        <t-form-item label="交付时间" name="deliverTime">
          <t-date-picker
            v-model="deliverForm.deliverTime"
            placeholder="请选择交付时间"
            format="YYYY-MM-DD HH:mm:ss"
            show-time
          />
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="deliverForm.remark" placeholder="请输入备注（可选）" :rows="3" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 列设置对话框 - 使用通用组件 -->
    <column-setting-dialog
      ref="columnSettingRef"
      title="列设置"
      :all-columns="allColumns"
      :default-columns="DEFAULT_COLUMNS"
      storage-key="mold_order_table_columns"
      @change="handleColumnsChange"
    />
  </div>
</template>
<script setup lang="ts">
import { AddIcon, RefreshIcon, SettingIcon } from 'tdesign-icons-vue-next';
import type { TableProps } from 'tdesign-vue-next';
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next';
import { computed, nextTick, onActivated, onMounted, reactive, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

import { getReceivablesByOrder } from '@/api/finance/receivable';
import {
  createMoldOrder,
  deleteMoldOrder,
  getMoldList,
  getMoldOrderById,
  getMoldOrderList,
  updateMoldOrder,
} from '@/api/mold';
import { getOrderPayments } from '@/api/mold/payment';
import { getCustomerList } from '@/api/reference';
import type { ColumnOption } from '@/components/ColumnSettingDialog/index.vue';
import ColumnSettingDialog from '@/components/ColumnSettingDialog/index.vue';
import type { FilterItem } from '@/components/SearchFilter/index.vue';
import SearchFilter from '@/components/SearchFilter/index.vue';
import { useDictRaw } from '@/hooks/useDict';
import { useUserStore } from '@/store';
import { formatDate } from '@/utils/date';
import { getDictColorRenderConfig } from '@/utils/dict-color';
import { OPERATION_COLUMN_WIDTH } from '@/utils/operation-column-width';

defineOptions({
  name: 'MoldOrder',
});

const router = useRouter();

// 表格相关
const loading = ref(false);
const tableData = ref<any[]>([]);
const pagination = reactive({
  current: 1,
  pageSize: 10,
  total: 0,
});

// 搜索表单（参数名与后端 snake_case 保持一致）
const formData = reactive({
  mold_code: '',
  order_code: '',
  customer_id: undefined as number | undefined,
  status: undefined as number | undefined,
  createTime: [] as string[],
  due_soon: undefined as number | undefined,
});

// 订单状态选项
const orderStatusOptions = [
  { label: '未开始', value: 1 },
  { label: '加工中', value: 2 },
  { label: '待试模', value: 3 },
  { label: '已试模', value: 4 },
  { label: '已完成', value: 5 },
  { label: '已取消', value: 6 },
  { label: '已交付', value: 7 },
];

// 客户列表
const customerList = ref<any[]>([]);

// 模具列表
const moldList = ref<any[]>([]);

// 订单状态字典列表（useDictRaw 自动加载 + 全局缓存）
const { items: orderStatusList } = useDictRaw('mold-order-status', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '')
    .map((item: any) => ({
      value: item.dict_value,
      label: item.dict_label,
      color: item.color || 'default',
      semanticTag: item.semantic_tag || '',
    })),
);

// 订单/工单类型字典列表（useDictRaw 自动加载 + 全局缓存）
const { items: orderTypeList } = useDictRaw('mold_order_type', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.list !== '')
    .map((item: any) => ({
      value: item.dict_value,
      label: item.dict_label,
      color: item.color || 'default',
    })),
);

// ✨ 优先级字典列表（useDictRaw 自动加载 + 全局缓存）
const { items: priorityList } = useDictRaw('order-priority', (items) =>
  (items as any[])
    .filter((item: any) => item.status === 1 && item.dict_value != null && item.dict_value !== '')
    .map((item: any) => ({
      value: item.dict_value,
      label: item.dict_label,
      color: item.color || 'default',
      themeStyle: item.ext?.themeStyle || '',
      theme: item.ext?.theme || 'default',
    })),
);

// 表格列配置
const columns: TableProps['columns'] = [
  { colKey: 'order_code', title: '订单编号', width: 180, fixed: 'left' },
  { colKey: 'order_type', title: '订单类型', width: 100 },
  { colKey: 'master_order_code', title: '来源订单', width: 150 },
  { colKey: 'priority', title: '优先级', width: 100 }, // ✨ 添加优先级列
  { colKey: 'customer.cus_name', title: '客户名称', width: 120 },
  { colKey: 'mold_codes', title: '模具编号', width: 180 },
  { colKey: 'plan_start_date', title: '计划开始', width: 120 },
  { colKey: 'plan_end_date', title: '订单交期', width: 120 },
  { colKey: 'totalAmount', title: '总金额', width: 120 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'deliver_time', title: '交付时间', width: 160 },
  { colKey: 'operation', title: '操作', width: OPERATION_COLUMN_WIDTH.LARGE, fixed: 'right' },
];

// 子表列配置
const itemColumns = [
  { colKey: 'mold_code', title: '模具编号', width: 150 },
  { colKey: 'mold_name', title: '产品名称', width: 150 },
  { colKey: 'amount', title: '金额', width: 120 },
  { colKey: 'operation', title: '操作', width: 80 },
];

// 对话框相关
const dialogVisible = ref(false);
const isEdit = ref(false);
const dialogTitle = ref('新增订单');
const formRef = ref();

// 保存原始表单数据（用于对比是否修改）
const originalFormData = ref<any>(null);

// 对比两个值是否相等（处理各种边界情况）
const isValueEqual = (val1: any, val2: any): boolean => {
  // 都未定义或都为null，视为相等
  if (val1 === undefined && val2 === undefined) return true;
  if (val1 === null && val2 === null) return true;

  // 将undefined和null视为相等
  if ((val1 === undefined || val1 === null) && (val2 === undefined || val2 === null)) return true;

  // 都为空字符串，视为相等
  if (val1 === '' && val2 === '') return true;

  // 空字符串和undefined/null视为相等
  if (
    (val1 === '' && (val2 === undefined || val2 === null)) ||
    (val2 === '' && (val1 === undefined || val1 === null))
  ) {
    return true;
  }

  // 数值比较（0 和 undefined 视为不同）
  if (typeof val1 === 'number' || typeof val2 === 'number') {
    return val1 === val2;
  }

  return val1 === val2;
};

const hasFormDataChanged = (): boolean => {
  if (!originalFormData.value) {
    return true;
  }

  // 对比表单字段
  const fieldsToCompare = [
    'orderType',
    'customerId',
    'moldId',
    'planStartDate',
    'planEndDate',
    'totalAmount',
    'status',
    'remark',
    'moldCode',
    'moldName',
    'workPlanStart',
    'workPlanEnd',
  ];

  // 详细对比每个字段
  for (const field of fieldsToCompare) {
    const val1 = formModel[field as keyof typeof formModel];
    const val2 = originalFormData.value[field];

    if (!isValueEqual(val1, val2)) {
      return true;
    }
  }

  return false;
};

// 款项管理对话框相关（对接应收账款）
const paymentDialogVisible = ref(false);
const currentOrder = ref<any>(null);
const receivableList = ref<any[]>([]);
const receivableLoading = ref(false);
// 款项相关字典（useDictRaw 自动加载 + 全局缓存）
const { items: paymentTypeOptions } = useDictRaw('payment-type');
const { items: receivableStatusOptions } = useDictRaw('receivable-status', (items) =>
  (items as any[]).filter((i: any) => i.status === 1),
);
const { items: _paymentMethodOptions } = useDictRaw('payment-method');

// 应收汇总计算
const receivableTotal = computed(() => {
  return receivableList.value.reduce((sum, r) => sum + (r.amount || 0), 0);
});
const receivableReceived = computed(() => {
  return receivableList.value.reduce((sum, r) => sum + (r.received_amount || 0), 0);
});
const receivableBalance = computed(() => {
  return receivableList.value.reduce((sum, r) => sum + (r.balance || 0), 0);
});

// 应收明细表格列
const receivableColumns = [
  { colKey: 'payment_type', title: '款项类型', width: 110 },
  { colKey: 'amount', title: '应收金额', width: 130 },
  { colKey: 'received_amount', title: '已收金额', width: 130 },
  { colKey: 'discount_amount', title: '优惠金额', width: 100 },
  { colKey: 'balance', title: '余额', width: 120 },
  { colKey: 'due_date', title: '到期日期', width: 140 },
  { colKey: 'status', title: '状态', width: 100 },
  { colKey: 'remark', title: '备注', ellipsis: true, minWidth: 80 },
  { colKey: 'op', title: '操作', width: 110, fixed: 'right' as const },
];

// 交付对话框相关
const deliverDialogVisible = ref(false);
const deliverFormRef = ref();
const currentOrderForDeliver = ref<any>(null);
const deliverForm = reactive({
  orderCode: '',
  customerName: '',
  deliverTime: '',
  remark: '',
});

const formModel = reactive({
  id: undefined,
  orderCode: '',
  orderType: 1,
  customerId: undefined,
  planStartDate: '',
  planEndDate: '',
  totalAmount: 0,
  status: 1,
  remark: '',
  items: [] as Array<{ mold_code: string; mold_name: string; amount: number }>,
});

// 监听订单类型变化
watch(
  () => formModel.orderType,
  () => {
    // 可以在这里处理一些联动逻辑
  },
);

const formRules: Record<string, any> = {
  orderType: [{ required: true, message: '请选择订单类型', trigger: 'change' as const }],
  customerId: [{ required: true, message: '请选择客户', trigger: 'change' as const }],
  status: [{ required: true, message: '请选择状态', trigger: 'change' as const }],
};

// 动态列配置（根据用户选择）
const displayColumns = computed(() => {
  // 根据 selectedColumns 的顺序来返回列配置
  const orderedColumns = selectedColumns.value
    .map((colKey) => columns.find((col) => col.colKey === colKey))
    .filter((col) => col !== undefined);

  // 将操作列添加到最后
  const operationCol = columns.find((col) => col.colKey === 'operation');
  if (operationCol) {
    orderedColumns.push(operationCol);
  }

  return orderedColumns;
});

// 所有可选列配置（不包括操作列）
const allColumns: ColumnOption[] = [
  { colKey: 'order_code', title: '订单编号', pinned: true }, // ✨ 固定列：不可隐藏/移动
  { colKey: 'order_type', title: '订单类型' },
  { colKey: 'master_order_code', title: '来源订单' },
  { colKey: 'priority', title: '优先级' }, // ✨ 添加优先级列配置
  { colKey: 'customer.cus_name', title: '客户名称' },
  { colKey: 'mold_codes', title: '模具编号' },
  { colKey: 'plan_start_date', title: '计划开始' },
  { colKey: 'plan_end_date', title: '订单交期' },
  { colKey: 'totalAmount', title: '总金额' },
  { colKey: 'status', title: '状态' },
  { colKey: 'deliver_time', title: '交付时间' },
];

// 默认显示的列
const DEFAULT_COLUMNS = [
  'order_code',
  'order_type',
  'master_order_code',
  'priority',
  'customer.cus_name',
  'mold_codes',
  'plan_start_date',
  'plan_end_date',
  'totalAmount',
  'status',
  'deliver_time',
];

// 列设置相关
const columnSettingRef = ref<InstanceType<typeof ColumnSettingDialog> | null>(null);
const selectedColumns = ref<string[]>([...DEFAULT_COLUMNS]);

const getOrderList = async () => {
  loading.value = true;
  try {
    const params: Record<string, any> = {
      page: pagination.current,
      page_size: pagination.pageSize,
      order_code: formData.order_code || undefined,
      mold_code: formData.mold_code || undefined,
      customer_id: formData.customer_id || undefined,
      status: formData.status || undefined,
      due_soon: formData.due_soon || undefined,
    };
    if (formData.createTime && formData.createTime.length === 2) {
      params.create_time_start = formData.createTime[0];
      params.create_time_end = formData.createTime[1];
    }
    const res = await getMoldOrderList(params);

    // 将后端下划线命名转换为前端驼峰命名
    tableData.value = (res.list || []).map((item: any) => ({
      id: item.id,
      order_code: item.order_code,
      order_type: item.order_type,
      master_order_code: item.master_order_code,
      priority: item.priority,
      customer_id: item.customer_id,
      mold_id: item.mold_id,
      customer: item.customer || { name: item.customer_name },
      mold: item.mold || { mold_code: item.mold_code },
      mold_codes: item.mold_codes || [],
      plan_start_date: item.plan_start_date,
      plan_end_date: item.plan_end_date,
      total_amount: item.total_amount,
      status: item.status,
      deliver_time: item.deliver_time,
      remark: item.remark,
      master_order_id: item.master_order_id,
    }));
    // 数据已加载完成
    pagination.total = res.total || 0;
  } catch {
    MessagePlugin.error('获取订单列表失败');
  } finally {
    loading.value = false;
  }
};

// 状态标签颜色样式（使用统一的颜色配置系统）
const getStatusStyle = (status: string) => {
  if (!status) return {};
  const statusItem = orderStatusList.value.find((s) => s.value === status);
  if (!statusItem) return {};

  const color = statusItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);

  // 如果是 theme 模式，直接返回空样式，由 t-tag 的 theme 属性处理
  if (renderConfig.mode === 'theme') {
    return {};
  }

  // style 模式，使用自定义样式
  if (renderConfig.style) {
    return {
      background: `${renderConfig.style.background} !important`,
      color: `${renderConfig.style.color} !important`,
      borderColor: `${renderConfig.style.borderColor} !important`,
    };
  }

  return {};
};

// 获取状态主题（用于 t-tag 的 theme 属性）
const getStatusTheme = (status: string) => {
  if (!status) return 'default';
  const statusItem = orderStatusList.value.find((s) => s.value === status);
  if (!statusItem) return 'default';

  const color = statusItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);

  return renderConfig.mode === 'theme' ? renderConfig.theme : 'default';
};

// 状态文本转换（使用字典数据）
const getStatusText = (status: string) => {
  if (!status) return '未知';
  const statusItem = orderStatusList.value.find((s) => s.value === status);
  return statusItem ? statusItem.label : '未知';
};

// 订单类型标签颜色（使用统一的颜色配置系统）
const getOrderTypeTheme = (orderType: number) => {
  const typeItem = orderTypeList.value.find((t) => t.value === orderType);
  if (typeItem?.color && typeItem.color !== 'default') {
    return typeItem.color;
  }
  // 如果没有字典配置，返回默认值
  return 'default';
};

// 订单类型文本转换（从字典数据获取）
const getOrderTypeText = (orderType: number) => {
  const typeItem = orderTypeList.value.find((t) => t.value === orderType);
  return typeItem ? typeItem.label : '未知';
};

// ✨ 优先级标签样式（使用统一的颜色配置系统）
const _getPriorityStyle = (priority: string) => {
  if (!priority) return getDictColorRenderConfig('default');
  const priorityItem = priorityList.value.find((p) => p.value === priority);
  if (!priorityItem) return getDictColorRenderConfig('default');

  const color = priorityItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);

  // 如果是 theme 模式，直接返回空样式，由 t-tag 的 theme 属性处理
  if (renderConfig.mode === 'theme') {
    return {};
  }

  // style 模式，使用自定义样式
  if (renderConfig.style) {
    return {
      background: `${renderConfig.style.background} !important`,
      color: `${renderConfig.style.color} !important`,
      borderColor: `${renderConfig.style.borderColor} !important`,
    };
  }

  return getDictColorRenderConfig('default');
};

// ✨ 获取优先级主题（用于 t-tag 的 theme 属性）
const _getPriorityTheme = (priority: string) => {
  if (!priority) return 'default';
  const priorityItem = priorityList.value.find((p) => p.value === priority);
  if (!priorityItem) return 'default';

  const color = priorityItem.color || 'default';
  const renderConfig = getDictColorRenderConfig(color);

  return renderConfig.mode === 'theme' ? renderConfig.theme : 'default';
};

// ✨ 优先级文本转换
const _getPriorityText = (priority: string) => {
  if (!priority) return '未知';
  const priorityItem = priorityList.value.find((p) => p.value === priority);
  return priorityItem ? priorityItem.label : '未知';
};

// 搜索
const handleSearch = () => {
  pagination.current = 1;
  getOrderList();
};

// 刷新
const handleRefresh = () => {
  getOrderList();
};

const handleReset = () => {
  formData.mold_code = '';
  formData.order_code = '';
  formData.customer_id = undefined;
  formData.status = undefined;
  formData.createTime = [];
  formData.due_soon = undefined;
  getOrderList();
};

// 新增 - 跳转到表单页面
const handleAdd = () => {
  router.push('/mold/order/form');
};

const handleAddItem = () => {
  formModel.items.push({ mold_code: '', mold_name: '', amount: 0 });
};

const handleRemoveItem = (index: number) => {
  if (formModel.items.length <= 1) {
    MessagePlugin.warning('至少保留一项模具明细');
    return;
  }
  formModel.items.splice(index, 1);
};

// 编辑 - 跳转到表单页面
const handleEdit = (row: any) => {
  router.push(`/mold/order/form?id=${row.id}`);
};

// 查看详情
const handleViewDetail = (row: any) => {
  router.push(`/mold/order/detail/${row.id}`);
};

// 查看关联的综合订单
const handleViewMasterOrder = (row: any) => {
  if (row.master_order_id) {
    router.push(`/order-management/detail/${row.master_order_id}`);
  }
};

// 款项管理 - 按模具订单查询款项（多源回退，与编辑页面一致）
const handlePaymentManage = async (row: any) => {
  currentOrder.value = row;
  receivableList.value = [];
  paymentDialogVisible.value = true;
  receivableLoading.value = true;

  // 额外：用 getMoldOrderById 获取完整数据，检查是否有内嵌款项
  try {
    const fullOrder = await getMoldOrderById(row.id);
    if (fullOrder) {
      // 预检查款项相关字段是否存在
      const paymentFields = [
        'payments',
        'payment_plan',
        'payment_plans',
        'receivables',
        'mold_order_payments',
      ];
      for (const f of paymentFields) {
        void (fullOrder as any)[f];
      }
    }
  } catch {
    // 静默处理，款项数据可选
  }

  try {
    let receivables: any[] = [];

    // 1. 优先：模具订单自身的应收账款表（创建订单时 PaymentPlan 写入）
    const arList = await getReceivablesByOrder(row.id, 'mold_order');
    if (arList && Array.isArray(arList) && arList.length > 0) {
      receivables = arList;
    }

    // 2. 回退：模具订单款项表(mold_order_payment，工单生成时自动创建)
    if (receivables.length === 0) {
      const paymentList = await getOrderPayments(row.id);
      if (paymentList && Array.isArray(paymentList) && paymentList.length > 0) {
        receivables = paymentList.map((p: any) => ({
          id: p.id,
          order_id: row.id,
          payment_type: p.payment_type || '',
          source_type: p.source_type || 'mold',
          amount: p.contract_amount || p.should_amount || 0,
          received_amount: p.actual_amount || 0,
          should_amount: p.should_amount || 0,
          discount_amount: p.discount_amount || 0,
          balance: (p.should_amount || p.contract_amount || 0) - (p.actual_amount || 0),
          due_date: p.due_date || p.payment_date || '',
          status: p.status === 3 ? 'PAID' : p.status === 2 ? 'PARTIAL' : 'PENDING',
          remark: p.remark || '',
        }));
      }
    }

    // 3. 回退：综合订单的应收账款
    if (receivables.length === 0 && row.master_order_id) {
      try {
        const masterARList = await getReceivablesByOrder(row.master_order_id, 'master_order');
        if (masterARList && Array.isArray(masterARList) && masterARList.length > 0) {
          receivables = masterARList.filter((ar: any) => {
            return ar.source_type === 'mold' || !ar.source_type;
          });
        }
      } catch (e) {
        console.warn('[款项] 查询综合订单应收款失败:', e);
      }
    }

    receivableList.value = receivables;
  } catch (error) {
    console.error('加载款项数据失败:', error);
  } finally {
    receivableLoading.value = false;
  }
};

// 查看收款明细
// 查看收款明细 - 跳转到应收详情页
const handleViewReceivablePayments = (row: any) => {
  router.push(`/finance/receivable/payments/${row.id}`);
};

// 格式化金额
function formatAmount(amount: number) {
  if (!amount && amount !== 0) return '0.00';
  return amount.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

// 款项类型标签
function getPaymentTypeLabel(type: string) {
  if (!type) return '-';
  const option = paymentTypeOptions.value.find((item: any) => item.dict_value === type);
  return option ? option.dict_label : type;
}
function getPaymentTypeTheme(type: string) {
  const option = paymentTypeOptions.value.find((item: any) => item.dict_value === type);
  if (option && option.color) {
    const renderConfig = getDictColorRenderConfig(option.color);
    if (renderConfig.mode === 'theme') {
      return renderConfig.theme || 'default';
    }
  }
  return 'default';
}
function getPaymentTypeStyle(type: string) {
  const option = paymentTypeOptions.value.find((item: any) => item.dict_value === type);
  if (option && option.color) {
    const renderConfig = getDictColorRenderConfig(option.color);
    if (renderConfig.mode === 'style' && renderConfig.style) {
      return renderConfig.style;
    }
    return {};
  }
  return {};
}

// 应收账款状态
function getReceivableStatusText(status: string) {
  if (status === 'locked') return '已锁定';
  const option = receivableStatusOptions.value.find((item: any) => item.dict_value === status);
  if (option?.dict_label) return option.dict_label;
  const map: Record<string, string> = {
    pending: '待收款',
    partial: '部分收款',
    paid: '已收款',
    overdue: '逾期',
  };
  return map[status?.toLowerCase()] || status;
}
function getReceivableStatusStyle(status: string) {
  if (status === 'locked') return {};
  const option = receivableStatusOptions.value.find((item: any) => item.dict_value === status) as any;
  if (option?.color) {
    const config = getDictColorRenderConfig(option.color);
    return config.mode === 'style' ? (config.style || {}) : {};
  }
  return {};
}
function getReceivableStatusTheme(status: string) {
  if (status === 'locked') return undefined;
  const option = receivableStatusOptions.value.find((item: any) => item.dict_value === status) as any;
  if (option?.color) {
    return getDictColorRenderConfig(option.color).theme;
  }
  const themeMap: Record<string, string> = { pending: 'default', partial: 'warning', paid: 'success', overdue: 'danger' };
  return themeMap[status?.toLowerCase()] as any;
}

// 收款方式

// 提交表单
const handleSubmit = async () => {
  try {
    // 编辑模式下，先检查数据是否有修改
    if (isEdit.value && !hasFormDataChanged()) {
      MessagePlugin.info('数据未做修改，无需提交');
      dialogVisible.value = false;
      return;
    }

    // 新增模式或数据有修改，进行表单验证
    // 编辑模式下，由于Object.assign不会触发change事件，需要特殊处理
    if (!isEdit.value) {
      // 新增模式：使用TDesign表单验证
      const validateResult = await formRef.value?.validate();
      if (validateResult) {
        MessagePlugin.warning('请检查表单填写是否完整');
        return;
      }
    } else {
      // 编辑模式：手动验证必填字段
      if (!formModel.orderType) {
        MessagePlugin.warning('请选择订单类型');
        return;
      }
      if (!formModel.customerId) {
        MessagePlugin.warning('请选择客户');
        return;
      }
      if (!formModel.status) {
        MessagePlugin.warning('请选择状态');
        return;
      }
    }

    // 将前端驼峰命名转换为后端下划线命名
    const submitData: any = {
      order_type: formModel.orderType,
      customer_id: formModel.customerId,
      total_amount: formModel.totalAmount,
      status: formModel.status,
      items: formModel.items.map((item) => ({
        mold_code: item.mold_code,
        mold_name: item.mold_name,
        amount: item.amount,
      })),
    };

    if (isEdit.value) {
      await updateMoldOrder(submitData);
    } else {
      await createMoldOrder(submitData);
    }

    MessagePlugin.success(isEdit.value ? '更新成功' : '创建成功');
    dialogVisible.value = false;
    getOrderList();
  } catch (error: any) {
    console.error('提交订单失败:', error);
    if (error?.message) {
      MessagePlugin.error(error.message);
    } else {
      MessagePlugin.error(isEdit.value ? '更新失败' : '创建失败');
    }
  }
};

// 关闭对话框
const _handleClose = () => {
  dialogVisible.value = false;
};

// 取消按钮
const handleCancel = () => {
  dialogVisible.value = false;
};

// 对话框完全打开后清除验证（编辑模式下避免误报验证错误）
const handleDialogOpened = async () => {
  await nextTick();
  // 清除表单验证状态，避免编辑模式下打开对话框时误报验证错误
  formRef.value?.clearValidate();
};

// 🔥 基于语义标签的自适应跳转逻辑
// 优势：当字典状态值变更时，无需修改代码，只需配置语义标签即可
const handleViewMoldWorkOrders = (row: any, moldCode?: string) => {
  const code = moldCode || row.mold?.mold_code;
  if (!code) {
    MessagePlugin.warning('该订单没有关联的模具');
    return;
  }

  // 🔥 根据字典的语义标签动态判断，而非硬编码状态值
  const statusItem = orderStatusList.value.find((s) => s.value === row.status);
  const semanticTag = statusItem?.semanticTag || '';

  // 定义需要跳转到模具档案的语义标签
  // 这些标签代表"生产流程已结束"的状态
  const archiveSemanticTags = ['completed', 'delivered'];

  if (archiveSemanticTags.includes(semanticTag)) {
    // 已完成或已交付 → 跳转模具档案页面
    window.location.href = `/mold/index?moldCode=${encodeURIComponent(code)}`;
  } else {
    // 其他状态（未开始、加工中、待试模、已试模、已取消）→ 跳转工单页面
    window.location.href = `/mold/work-order?moldCode=${encodeURIComponent(code)}`;
  }
};

const handleDelete = (row: any) => {
  const userStore = useUserStore();
  const isAdmin = userStore.roles?.includes('超级管理员') || false;

  const dialog = DialogPlugin.confirm({
    header: isAdmin ? '确认删除（超级管理员特权）' : '确认删除',
    body: isAdmin
      ? `超级管理员特权：确定要删除订单 "${row.order_code}" 吗？\n⚠️ 该订单已进入生产阶段，删除后不可恢复！`
      : `确定要删除订单 "${row.order_code}" 吗？此操作不可恢复！`,
    theme: isAdmin ? 'warning' : 'default',
    confirmBtn: isAdmin ? '确定删除' : '确定删除',
    cancelBtn: '取消',
    onConfirm: async () => {
      try {
        await deleteMoldOrder(row.id);
        MessagePlugin.success('删除成功');
        getOrderList();
        dialog.destroy();
      } catch (error: any) {
        MessagePlugin.error(error.message || '删除失败');
        dialog.destroy();
      }
    },
  });
};

const canDeliver = (row: any): boolean => {
  // 后端 status 为字符串（如 "COMPLETED"），前端也可能传数字 5，兼容两种
  return row.status === 5 || String(row.status) === '5' || row.status === 'COMPLETED';
};

// 交付操作
const handleDeliver = (row: any) => {
  currentOrderForDeliver.value = row;
  deliverForm.orderCode = row.order_code || '';
  deliverForm.customerName = row.customer?.cus_name || row.customer?.name || '';
  deliverForm.deliverTime = '';
  deliverForm.remark = '';
  deliverDialogVisible.value = true;
};

// 确认交付
const handleConfirmDeliver = async () => {
  if (!deliverForm.deliverTime) {
    MessagePlugin.warning('请选择交付时间');
    return;
  }

  try {
    await updateMoldOrder({
      id: currentOrderForDeliver.value.id,
      status: 'DELIVERED',
      deliver_time: deliverForm.deliverTime,
    });

    MessagePlugin.success('交付成功');
    deliverDialogVisible.value = false;
    getOrderList();
  } catch (error: any) {
    console.error('交付失败:', error);
    MessagePlugin.error(error?.message || '交付失败');
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
    const res = await getMoldList({ page: 1, page_size: 1000 });
    moldList.value = res.list || [];
  } catch (error) {
    console.error('获取模具列表失败:', error);
  }
};

// 加载款项相关字典（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载订单状态字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

// 加载订单/工单类型字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

// ✨ 优先级字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

// 分页变化
const rehandlePageChange = (curr: any) => {
  pagination.current = curr.current;
  pagination.pageSize = curr.pageSize;
  getOrderList();
};

// 打开列设置对话框
const handleColumnSetting = () => {
  columnSettingRef.value?.open();
};

// 列配置变化时的回调
const handleColumnsChange = (columns: string[]) => {
  selectedColumns.value = columns;
};

// 到期预警选项
const dueSoonOptions = [
  { label: '3天内到期', value: 3 },
  { label: '7天内到期', value: 7 },
  { label: '15天内到期', value: 15 },
  { label: '30天内到期', value: 30 },
];

// 筛选栏配置
const filterConfig: FilterItem[] = [
  { type: 'input', field: 'mold_code', placeholder: '请输入模具编号' },
  { type: 'input', field: 'order_code', placeholder: '请输入订单编号' },
  {
    type: 'select',
    field: 'customer_id',
    placeholder: '请选择客户',
    options: customerList,
    labelKey: 'cus_name',
    valueKey: 'id',
  },
  { type: 'select', field: 'status', placeholder: '请选择状态', options: orderStatusOptions },
  { type: 'dateRangePicker', field: 'createTime', placeholder: '创建时间', span: 8 },
  { type: 'select', field: 'due_soon', placeholder: '到期预警', options: dueSoonOptions, span: 4 },
];

onMounted(() => {
  getOrderList();
  loadCustomerList();
  loadMoldList();
  // 字典已通过 useDictRaw 自动加载
});

// KeepAlive 缓存恢复时自动刷新列表，确保新增/编辑后数据最新
onActivated(() => {
  getOrderList();
});
</script>
<style scoped>
.mold-order-container {
  padding: 16px;
}

.search-card,
.toolbar-card {
  margin-bottom: 16px;
}

.toolbar-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.toolbar-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}

.toolbar-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}

/* 表单分割线样式 */
.section-title {
  font-size: 16px;
  font-weight: bold;
  color: var(--td-text-color-primary);
}

.section-divider-box {
  width: 100%;
  position: relative;
  border-top: 1px solid var(--td-border-level-1-color, #e7e7e7);
}

.section-divider-text {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: white;
  padding: 0 16px;
  color: var(--td-text-color-secondary);
  font-size: 14px;
  white-space: nowrap;
}

/* 模具明细标题和按钮并排布局 */
.mold-items-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.mold-items-header .section-title {
  margin: 0;
  flex: 1;
}

/* 模具编号列表样式 */
.mold-codes-list {
  display: flex;
  flex-direction: column;
}

.mold-code-item {
  padding: 2px 0;
  line-height: 1.5;
  color: var(--td-text-color-primary);

  &:not(:last-child) {
    border-bottom: 1px dashed var(--td-border-level-1-color, #e7e7e7);
  }
}

.mold-code-item .t-link {
  font-weight: 600;
  font-size: 15px;
}

/* 表格行悬停效果 */
:deep(.t-table) {
  .t-table__body tr {
    transition: background-color 0.2s ease;

    &:hover {
      background-color: var(--td-bg-color-container-hover) !important;
    }
  }
}

/* 款项管理弹窗样式 - 类似应收账款详情页 */
.payment-dialog-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.payment-section {
  border-radius: var(--td-radius-medium, 6px);
}

.payment-section :deep(.t-card__header) {
  font-size: 15px;
  font-weight: 600;
  padding: 14px 20px;
  border-bottom: 1px solid var(--td-component-border, #e7e7e7);
}

.payment-section :deep(.t-card__body) {
  padding: 16px 20px;
}

/* 资金汇总 - 横向陈列 */
.summary-stats {
  display: flex;
  gap: 24px;
  justify-content: space-around;
}

.summary-stats .stat-item {
  text-align: center;
  flex: 1;
  min-width: 0;
}

.summary-stats .stat-label {
  font-size: 14px;
  color: var(--td-text-color-secondary, rgb(0 0 0 / 60%));
  margin-bottom: 8px;
}

.summary-stats .stat-value {
  font-size: 22px;
  font-weight: 600;
  color: var(--td-text-color-primary, rgb(0 0 0 / 90%));
  margin-bottom: 4px;
}

.summary-stats .stat-value.success {
  color: var(--td-success-color, #00a870);
}

.summary-stats .stat-value.warning {
  color: var(--td-warning-color, #ed7b2f);
}

/* 文本颜色工具 */
.text-success {
  color: var(--td-success-color, #00a870);
  font-weight: 500;
}

.text-warning {
  color: var(--td-warning-color, #ed7b2f);
  font-weight: 500;
}
</style>

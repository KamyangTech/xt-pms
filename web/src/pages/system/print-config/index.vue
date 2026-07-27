<template>
  <div class="print-config-container">
    <t-card class="header-card">
      <div class="header-content">
        <span class="page-title">打印配置</span>
        <div class="header-actions">
          <t-button variant="outline" @click="handleClearCache">
            <template #icon><delete-icon /></template>
            清除缓存
          </t-button>
        </div>
      </div>
    </t-card>

    <t-card :bordered="false">
      <t-alert theme="info" message="配置将在下次打印时生效" style="margin-bottom: 16px" />

      <t-tabs v-model="activeTab" default-value="production_order" @change="onTabChange">
        <t-tab-panel value="production_order" label="生产通知单">
          <div class="tab-content">
            <t-card
              v-for="group in productionOptionGroups"
              :key="group.groupKey"
              :title="group.label"
              :bordered="false"
              class="setting-card"
            >
              <div class="option-grid">
                <t-checkbox
                  v-for="item in group.options"
                  :key="item.key"
                  :checked="getValue(productionForm, group.groupKey, item.key)"
                  @change="(val: boolean) => setValue(productionForm, group.groupKey, item.key, val)"
                >
                  {{ item.label }}
                </t-checkbox>
              </div>
            </t-card>
          </div>
        </t-tab-panel>

        <t-tab-panel value="delivery_note" label="送货单">
          <div class="tab-content">
            <t-card
              v-for="group in deliveryOptionGroups"
              :key="group.groupKey"
              :title="group.label"
              :bordered="false"
              class="setting-card"
            >
              <div class="option-grid">
                <t-checkbox
                  v-for="item in group.options"
                  :key="item.key"
                  :checked="getValue(deliveryForm, group.groupKey, item.key)"
                  @change="(val: boolean) => setValue(deliveryForm, group.groupKey, item.key, val)"
                >
                  {{ item.label }}
                </t-checkbox>
              </div>
            </t-card>
          </div>
        </t-tab-panel>
      </t-tabs>

      <!-- 操作按钮 -->
      <t-card :bordered="false" class="action-card">
        <t-button theme="primary" @click="handleSave" :loading="saving">
          <template #icon><save-icon /></template>
          保存配置
        </t-button>
        <t-button theme="default" @click="handleReset">
          <template #icon><refresh-icon /></template>
          重置为默认
        </t-button>
      </t-card>
    </t-card>
  </div>
</template>

<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { SaveIcon, RefreshIcon, DeleteIcon } from 'tdesign-icons-vue-next';
import { onMounted, ref, reactive } from 'vue';

import {
  getPrintNoticeConfig,
  savePrintNoticeConfig,
  resetPrintNoticeConfig,
  type PrintNoticeConfig,
  type DeliveryNoteConfig,
} from '@/api/print-notice-config';

defineOptions({
  name: 'PrintNoticeConfig',
});

interface OptionItem {
  label: string;
  key: string;
}

interface OptionGroup {
  label: string;
  groupKey: string;
  options: OptionItem[];
}

const productionOptionGroups: OptionGroup[] = [
  {
    label: '基本信息',
    groupKey: 'basic_info',
    options: [
      { label: '订单号', key: 'order_code' },
      { label: '客户名称', key: 'customer_name' },
      { label: '生产类型', key: 'production_type' },
      { label: '模具来源', key: 'mold_source' },
    ],
  },
  {
    label: '生产信息',
    groupKey: 'production_info',
    options: [
      { label: '计划数量', key: 'plan_qty' },
      { label: '负责人', key: 'worker' },
      { label: '计划开始日期', key: 'plan_start_date' },
      { label: '计划结束日期', key: 'plan_end_date' },
      { label: '完成数量', key: 'completed_qty' },
    ],
  },
  {
    label: '生产明细列',
    groupKey: 'detail_columns',
    options: [
      { label: '产品名称', key: 'product_name' },
      { label: '模具编号', key: 'mold_code' },
      { label: '产品图片', key: 'product_image' },
      { label: '计划数量', key: 'plan_qty' },
      { label: '已完成数量', key: 'completed_qty' },
    ],
  },
];

const deliveryOptionGroups: OptionGroup[] = [
  {
    label: '基本信息',
    groupKey: 'basic_info',
    options: [
      { label: '送货单号', key: 'stock_code' },
      { label: '接收客户', key: 'customer_name' },
      { label: '模具编号', key: 'mold_code' },
      { label: '产品名称', key: 'product_name' },
    ],
  },
  {
    label: '发货信息',
    groupKey: 'delivery_info',
    options: [
      { label: '发货日期', key: 'delivery_date' },
      { label: '出库日期', key: 'out_date' },
      { label: '发货方式', key: 'delivery_method' },
      { label: '发货数量', key: 'quantity' },
      { label: '快递公司', key: 'express_company' },
      { label: '快递单号', key: 'express_no' },
      { label: '司机电话', key: 'driver_phone' },
      { label: '车牌号', key: 'driver_plate' },
    ],
  },
  {
    label: '产品明细列',
    groupKey: 'detail_columns',
    options: [
      { label: '产品图片', key: 'product_image' },
      { label: '模具编号', key: 'mold_code' },
      { label: '产品名称', key: 'product_name' },
      { label: '发货数量', key: 'quantity' },
      { label: '批次号', key: 'batch_no' },
      { label: '备注', key: 'remark' },
    ],
  },
];

const saving = ref(false);
const activeTab = ref('production_order');

// 生产通知单表单数据
const productionForm = reactive<PrintNoticeConfig>({
  basic_info: {
    order_code: true,
    customer_name: true,
    production_type: true,
    mold_source: true,
  },
  production_info: {
    plan_qty: true,
    worker: true,
    plan_start_date: true,
    plan_end_date: true,
    completed_qty: true,
  },
  detail_columns: {
    product_name: true,
    mold_code: true,
    product_image: true,
    plan_qty: true,
    completed_qty: true,
  },
});

// 送货单表单数据
const deliveryForm = reactive<DeliveryNoteConfig>({
  basic_info: {
    stock_code: true,
    customer_name: true,
    mold_code: true,
    product_name: true,
  },
  delivery_info: {
    delivery_date: true,
    out_date: true,
    delivery_method: true,
    quantity: true,
    express_company: true,
    express_no: true,
    driver_phone: true,
    driver_plate: true,
  },
  detail_columns: {
    product_image: true,
    mold_code: true,
    product_name: true,
    quantity: true,
    batch_no: true,
    remark: true,
  },
});

const getValue = (form: Record<string, any>, groupKey: string, key: string): boolean => {
  return form[groupKey]?.[key] ?? false;
};

const setValue = (form: Record<string, any>, groupKey: string, key: string, val: boolean) => {
  if (form[groupKey] && key in form[groupKey]) {
    form[groupKey][key] = val;
  }
};

const handleClearCache = () => {
  MessagePlugin.info('清除缓存功能开发中...');
};

// 获取当前配置类型
const currentConfigType = () => activeTab.value;

const loadConfig = async (configType: string) => {
  try {
    const res = await getPrintNoticeConfig(configType);
    if (!res) return;

    if (configType === 'delivery_note') {
      const data = res.data || res;
      if (data && data.basic_info && data.delivery_info && data.detail_columns) {
        Object.assign(deliveryForm.basic_info, data.basic_info);
        Object.assign(deliveryForm.delivery_info, data.delivery_info);
        Object.assign(deliveryForm.detail_columns, data.detail_columns);
      }
    } else {
      const data = res.data || res;
      if (data && data.basic_info && data.production_info && data.detail_columns) {
        Object.assign(productionForm.basic_info, data.basic_info);
        Object.assign(productionForm.production_info, data.production_info);
        Object.assign(productionForm.detail_columns, data.detail_columns);
      }
    }
  } catch (error) {
    console.error('加载配置失败:', error);
    MessagePlugin.error('加载配置失败');
  }
};

// 保存当前 tab 的配置
const handleSave = async () => {
  saving.value = true;
  try {
    const configType = currentConfigType();
    const configData = configType === 'delivery_note'
      ? JSON.parse(JSON.stringify(deliveryForm))
      : JSON.parse(JSON.stringify(productionForm));

    await savePrintNoticeConfig(configData, configType);
    MessagePlugin.success('保存成功');
  } catch (error) {
    console.error('保存配置失败:', error);
    MessagePlugin.error('保存配置失败');
  } finally {
    saving.value = false;
  }
};

// 重置当前 tab 的配置
const handleReset = async () => {
  try {
    const configType = currentConfigType();
    await resetPrintNoticeConfig(configType);
    MessagePlugin.success('已重置为默认配置');
    await loadConfig(configType);
  } catch (error) {
    console.error('重置配置失败:', error);
    MessagePlugin.error('重置配置失败');
  }
};

// tab 切换时加载对应配置
const onTabChange = (value: string) => {
  loadConfig(value);
};

onMounted(() => {
  loadConfig('production_order');
});
</script>

<style scoped>
.print-config-container {
  margin: var(--td-comp-margin-l);
}

.header-card {
  margin-bottom: var(--td-comp-margin-l);
}

.header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-title {
  font-size: 16px;
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.header-actions {
  display: flex;
  gap: 8px;
}

.tab-content {
  padding-top: 16px;
}

.setting-card {
  margin-bottom: var(--td-comp-margin-l);
}

.option-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 24px;
}
</style>
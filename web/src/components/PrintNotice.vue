<template>
  <PrintTemplate
    v-if="visible && order"
    :visible="visible"
    :title="title"
    :sections="printSections"
    :company-name="companyName"
    :operator="operatorName"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { formatDate } from '@/utils/date';
import type { ProductionOrder } from '@/types/production';
import type { PrintNoticeConfig } from '@/api/print-notice-config';
import { getCompanyInfo } from '@/api/reference';
import { getThumbnailUrl } from '@/api/common/upload';
import PrintTemplate, { type PrintSection } from '@/components/PrintTemplate.vue';

defineOptions({
  name: 'PrintNotice',
});

interface Props {
  visible: boolean;
  order: ProductionOrder | null;
  config?: PrintNoticeConfig | null;
  productionTypeDict?: any[];
  moldSourceDict?: any[];
  employeeDict?: any[];
}

const props = withDefaults(defineProps<Props>(), {
  productionTypeDict: () => [],
  moldSourceDict: () => [],
  employeeDict: () => [],
  config: null,
});

// 公司名称
const companyName = ref('');

// 操作员名称
const operatorName = ref('');

// 标题固定为"生产通知单"
const title = '生产通知单';

const fetchCompanyInfo = async () => {
  try {
    const res = await getCompanyInfo();
    if (res) {
      // 使用公司全称
      companyName.value = res.company_name || '';
    }
  } catch (error) {
    console.error('获取公司信息失败:', error);
  }
};

// 获取当前操作员信息
const fetchOperatorInfo = () => {
  try {
    const userInfoStr = localStorage.getItem('userInfo');
    if (userInfoStr) {
      const userInfo = JSON.parse(userInfoStr);
      operatorName.value = userInfo.real_name || userInfo.username || '';
    }
  } catch (error) {
    console.error('获取操作员信息失败:', error);
  }
};

onMounted(() => {
  fetchCompanyInfo();
  fetchOperatorInfo();
});

// 默认配置
const defaultConfig: PrintNoticeConfig = {
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
};

// 使用配置（如果没有则使用默认配置）
const currentConfig = computed(() => props.config || defaultConfig);

const printSections = computed<PrintSection[]>(() => {
  if (!props.order) return [];
  
  const sections: PrintSection[] = [];
  const o = props.order;
  const c = currentConfig.value;
  
  // 基本信息区块
  if (c.basic_info.order_code || c.basic_info.customer_name || c.basic_info.production_type || c.basic_info.mold_source) {
    const basicFields: Array<{ key: string; label: string; value: string }> = [];
    
    if (c.basic_info.order_code) {
      basicFields.push({ key: 'order_code', label: '订单号', value: o.order_code });
    }
    if (c.basic_info.customer_name) {
      basicFields.push({ key: 'customer_name', label: '客户名称', value: o.customer?.cus_name || '-' });
    }
    if (c.basic_info.production_type) {
      basicFields.push({ key: 'production_type', label: '生产类型', value: getProductionTypeLabel(o.production_type) });
    }
    if (c.basic_info.mold_source) {
      basicFields.push({ key: 'mold_source', label: '模具来源', value: getMoldSourceLabel(o.mold_source_type) });
    }
    
    if (basicFields.length > 0) {
      sections.push({
        type: 'info-table',
        title: '基本信息',
        fields: basicFields,
      });
    }
  }
  
  // 生产信息区块
  if (c.production_info.plan_qty || c.production_info.worker || c.production_info.plan_start_date || c.production_info.plan_end_date || c.production_info.completed_qty) {
    const productionFields: Array<{ key: string; label: string; value: string }> = [];
    
    if (c.production_info.plan_qty) {
      productionFields.push({ key: 'plan_qty', label: '计划数量', value: String(o.plan_qty) });
    }
    if (c.production_info.worker) {
      productionFields.push({ key: 'worker', label: '负责人', value: getWorkerName(o.worker_id) });
    }
    if (c.production_info.plan_start_date) {
      productionFields.push({ key: 'plan_start_date', label: '计划开始日期', value: formatDate(o.plan_start_date) });
    }
    if (c.production_info.plan_end_date) {
      productionFields.push({ key: 'plan_end_date', label: '计划结束日期', value: formatDate(o.plan_end_date) });
    }
    if (c.production_info.completed_qty) {
      productionFields.push({ key: 'completed_qty', label: '完成数量', value: String(o.completed_qty ?? 0) });
    }
    
    if (productionFields.length > 0) {
      sections.push({
        type: 'info-table',
        title: '生产信息',
        fields: productionFields,
      });
    }
  }
  
  // 生产明细区块
  if (o.items && o.items.length > 0) {
    const columns: Array<{ key: string; label: string; width?: string; imageUrl?: string }> = [];
    
    if (c.detail_columns.product_name) {
      columns.push({ key: 'product_name', label: '产品名称', width: '15%' });
    }
    if (c.detail_columns.mold_code) {
      columns.push({ key: 'mold_code', label: '模具编号', width: '18%' });
    }
    if (c.detail_columns.product_image) {
      columns.push({ key: 'product_image', label: '产品图片', width: '42%', imageUrl: 'product_image_url' });
    }
    if (c.detail_columns.plan_qty) {
      columns.push({ key: 'plan_qty', label: '计划数量', width: '12%' });
    }
    if (c.detail_columns.completed_qty) {
      columns.push({ key: 'completed_qty', label: '已完成数量', width: '13%' });
    }
    
    const data = o.items.map((item) => ({
      product_name: item.product_name || item.product?.pro_name || '-',
      mold_code: item.mold?.mold_code || item.external_mold_code || '-',
      product_image_url: getThumbnailUrl(item.product_image_url || ''),
      plan_qty: item.plan_qty,
      completed_qty: item.completed_qty ?? 0,
    }));
    
    sections.push({
      type: 'data-table',
      title: '生产明细',
      columns,
      data,
    });
  }
  
  return sections;
});

const getProductionTypeLabel = (type: string | undefined) => {
  if (!type) return '-';
  const dictItem = props.productionTypeDict.find(
    (item) => item.dict_value === type
  );
  return dictItem?.dict_label || '-';
};

const getMoldSourceLabel = (type: string | undefined) => {
  if (!type) return '-';
  const dictItem = props.moldSourceDict.find(
    (item) => item.dict_value === type
  );
  return dictItem?.dict_label || '-';
};

const getWorkerName = (workerId: string | undefined) => {
  if (!workerId) return '-';
  const dictItem = props.employeeDict.find(
    (item) => item.dict_value === workerId
  );
  return dictItem?.dict_label || workerId;
};
</script>



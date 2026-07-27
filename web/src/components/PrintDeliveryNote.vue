<template>
  <PrintTemplate
    v-if="visible && records.length > 0"
    :visible="visible"
    :title="title"
    :sections="printSections"
    :show-signature="true"
    :company-name="companyName"
    :operator="operatorName"
  />
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { formatDate } from '@/utils/date';
import { getCompanyInfo } from '@/api/reference';
import { getThumbnailUrl } from '@/api/common/upload';
import PrintTemplate, { type PrintSection } from '@/components/PrintTemplate.vue';
import type { DeliveryNoteConfig } from '@/api/print-notice-config';

defineOptions({
  name: 'PrintDeliveryNote',
});

interface Props {
  visible: boolean;
  records: any[];
  customerName?: string;
  config?: DeliveryNoteConfig | null;
  deliveryMethodDict?: any[];
  expressCompanyDict?: any[];
}

const props = withDefaults(defineProps<Props>(), {
  customerName: '',
  config: null,
  deliveryMethodDict: () => [],
  expressCompanyDict: () => [],
});

// 公司名称
const companyName = ref('');

// 操作员名称
const operatorName = ref('');

// 标题固定为"送货单"
const title = '送货单';

const fetchCompanyInfo = async () => {
  try {
    const res = await getCompanyInfo();
    if (res) {
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
const defaultConfig: DeliveryNoteConfig = {
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
};

// 使用配置（如果没有则使用默认配置）
const currentConfig = computed(() => props.config || defaultConfig);

const printSections = computed<PrintSection[]>(() => {
  if (props.records.length === 0) return [];

  const record = props.records[0];
  const sections: PrintSection[] = [];
  const c = currentConfig.value;

  // 基本信息区块
  if (c.basic_info.stock_code || c.basic_info.customer_name || c.basic_info.mold_code || c.basic_info.product_name) {
    const basicFields: Array<{ key: string; label: string; value: string }> = [];

    if (c.basic_info.stock_code) {
      basicFields.push({ key: 'stock_code', label: '送货单号', value: record.stock_code || '-' });
    }
    if (c.basic_info.customer_name) {
      basicFields.push({ key: 'customer_name', label: '接收客户', value: props.customerName || record.customer_name || '-' });
    }
    if (c.basic_info.mold_code) {
      basicFields.push({ key: 'mold_code', label: '模具编号', value: record.mold?.mold_code || record.external_mold_code || '-' });
    }
    if (c.basic_info.product_name) {
      basicFields.push({ key: 'product_name', label: '产品名称', value: record.product_name || '-' });
    }

    if (basicFields.length > 0) {
      sections.push({
        type: 'info-table',
        title: '基本信息',
        fields: basicFields,
      });
    }
  }

  // 发货信息区块（仅在有发货方式时显示）
  if (record.delivery_method) {
    if (c.delivery_info.delivery_date || c.delivery_info.out_date || c.delivery_info.delivery_method || c.delivery_info.quantity || c.delivery_info.express_company || c.delivery_info.express_no || c.delivery_info.driver_phone || c.delivery_info.driver_plate) {
      const deliveryFields: Array<{ key: string; label: string; value: string }> = [];

      if (c.delivery_info.delivery_date) {
        deliveryFields.push({ key: 'delivery_date', label: '发货日期', value: formatDate(record.delivery_date) });
      }
      if (c.delivery_info.out_date) {
        deliveryFields.push({ key: 'stock_out_date', label: '出库日期', value: formatDate(record.stock_out_date) });
      }
      if (c.delivery_info.delivery_method) {
        deliveryFields.push({ key: 'delivery_method', label: '发货方式', value: getDeliveryMethodName(record.delivery_method) });
      }
      if (c.delivery_info.quantity) {
        deliveryFields.push({ key: 'quantity', label: '发货数量', value: String(record.quantity || 0) });
      }

      // 根据发货方式添加额外字段
      if (record.delivery_method === 'express') {
        if (c.delivery_info.express_company) {
          deliveryFields.push({ key: 'express_company', label: '快递公司', value: getExpressCompanyName(record.express_company) });
        }
        if (c.delivery_info.express_no) {
          deliveryFields.push({ key: 'express_no', label: '快递单号', value: record.express_no || '-' });
        }
      } else if (record.delivery_method === 'lalamove') {
        if (c.delivery_info.driver_phone) {
          deliveryFields.push({ key: 'driver_phone', label: '司机电话', value: record.driver_phone || '-' });
        }
        if (c.delivery_info.driver_plate) {
          deliveryFields.push({ key: 'driver_plate', label: '车牌号', value: record.driver_plate || '-' });
        }
      }

      if (deliveryFields.length > 0) {
        sections.push({
          type: 'info-table',
          title: '发货信息',
          fields: deliveryFields,
        });
      }
    }
  }

  // 产品明细表格
  if (c.detail_columns.product_image || c.detail_columns.mold_code || c.detail_columns.product_name || c.detail_columns.quantity || c.detail_columns.batch_no || c.detail_columns.remark) {
    const columns: Array<{ key: string; label: string; width?: string; imageUrl?: string }> = [];
    columns.push({ key: 'index', label: '序号', width: '8%' });

    if (c.detail_columns.product_image) {
      columns.push({ key: 'product_image', label: '产品图片', width: '15%', imageUrl: 'product_image_url' });
    }
    if (c.detail_columns.mold_code) {
      columns.push({ key: 'mold_code', label: '模具编号', width: '15%' });
    }
    if (c.detail_columns.product_name) {
      columns.push({ key: 'product_name', label: '产品名称', width: '20%' });
    }
    if (c.detail_columns.quantity) {
      columns.push({ key: 'quantity', label: '发货数量', width: '12%' });
    }
    if (c.detail_columns.batch_no) {
      columns.push({ key: 'batch_no', label: '批次号', width: '15%' });
    }
    if (c.detail_columns.remark) {
      columns.push({ key: 'remark', label: '备注', width: '15%' });
    }

    sections.push({
      type: 'data-table',
      title: '产品明细',
      columns,
      data: props.records.map((r, index) => {
        // 尝试获取产品图片URL
        let imageUrl = r.product_image_url || '';
        // 如果记录上没有图片URL，尝试从订单明细中查找
        if (!imageUrl && r.order?.items?.length > 0) {
          const foundItem = r.order.items.find((item: any) => item.product_image_url);
          if (foundItem) {
            imageUrl = foundItem.product_image_url;
          }
        }
        return {
          index: index + 1,
          product_image_url: getThumbnailUrl(imageUrl),
          mold_code: r.mold?.mold_code || r.external_mold_code || '-',
          product_name: r.product_name || '-',
          quantity: r.quantity || 0,
          batch_no: r.batch_no || '-',
          remark: r.remark || '-',
        };
      }),
      summary: props.records.length > 1 ? {
        label: '合计:',
        value: props.records.reduce((sum, r) => sum + (r.quantity || 0), 0),
        colspan: 3,
      } : undefined,
    });
  }

  return sections;
});

const getDeliveryMethodName = (value: string) => {
  if (!value) return '-';
  if (!Array.isArray(props.deliveryMethodDict)) return value;
  const item = props.deliveryMethodDict.find((d) => d.dict_value === value);
  return item ? item.dict_label : value;
};

const getExpressCompanyName = (value: string) => {
  if (!value) return '-';
  if (!Array.isArray(props.expressCompanyDict)) return value;
  const item = props.expressCompanyDict.find((d) => d.dict_value === value);
  return item ? item.dict_label : value;
};
</script>

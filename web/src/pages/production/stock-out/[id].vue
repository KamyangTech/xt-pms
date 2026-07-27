<template>
  <div class="stock-out-detail-page">
    <!-- 返回按钮 -->
    <div class="back-button-container">
      <t-button theme="default" @click="handleBack" variant="text">
        <template #icon><t-icon name="chevron-left" /></template>
        返回
      </t-button>
    </div>

    <t-card :bordered="false">
      <div v-if="loading" class="loading-container">
        <PageLoading />
      </div>
      <div v-else-if="detailData">
        <!-- 页面头部 -->
        <div class="page-header">
          <div class="header-left">
            <h2 class="order-code">{{ detailData.stock_code || '送货单详情' }}</h2>
          </div>
          <div class="header-right">
            <t-button theme="default" @click="handlePrint">
              <template #icon><print-icon /></template>
              打印
            </t-button>
            <t-button theme="primary" @click="handleEdit">编辑</t-button>
          </div>
        </div>

        <!-- 基本信息 -->
        <t-card :bordered="false" title="基本信息">
          <t-descriptions :column="3" bordered>
            <t-descriptions-item label="模具编号">{{ detailData.mold_code || '-' }}</t-descriptions-item>
            <t-descriptions-item label="产品名称">{{ detailData.product_name || '-' }}</t-descriptions-item>
            <t-descriptions-item label="出库数量">{{ detailData.quantity }}</t-descriptions-item>
            <t-descriptions-item label="接收客户">{{ detailData.customer_name || '-' }}</t-descriptions-item>
            <t-descriptions-item label="发货日期">{{ formatDateTime(detailData.delivery_date) }}</t-descriptions-item>
            <t-descriptions-item label="出库日期">{{ formatDateTime(detailData.stock_out_date) }}</t-descriptions-item>
          </t-descriptions>
        </t-card>

        <!-- 发货信息 -->
        <t-card title="发货信息" class="container-base-margin-top" :bordered="false">
          <t-descriptions :column="3" bordered>
            <t-descriptions-item label="发货方式">
              <t-tag v-if="detailData.delivery_method" variant="light-outline">
                {{ getDeliveryMethodName(detailData.delivery_method) }}
              </t-tag>
              <span v-else>-</span>
            </t-descriptions-item>
            <t-descriptions-item label="快递公司" v-if="detailData.delivery_method === 'EXPRESS'">
              {{ getExpressCompanyName(detailData.express_company) || '-' }}
            </t-descriptions-item>
            <t-descriptions-item label="快递单号" v-if="detailData.delivery_method === 'EXPRESS'">
              {{ detailData.express_no || '-' }}
            </t-descriptions-item>
            <t-descriptions-item label="司机电话" v-if="detailData.delivery_method === 'LALAMOVE'">
              {{ detailData.driver_phone || '-' }}
            </t-descriptions-item>
            <t-descriptions-item label="车牌号" v-if="detailData.delivery_method === 'LALAMOVE'">
              {{ detailData.driver_plate || '-' }}
            </t-descriptions-item>
          </t-descriptions>
        </t-card>

        <!-- 照片 -->
        <t-card v-if="photos.length > 0" title="发货照片" class="container-base-margin-top" :bordered="false">
          <div class="photos-grid">
            <img
              v-for="(photo, index) in photos"
              :key="index"
              :src="photo"
              class="photo-thumbnail"
              alt="发货照片"
              @click="handlePreviewPhoto(index)"
            />
          </div>
        </t-card>

        <!-- 备注 -->
        <t-card v-if="detailData.remark" title="备注" class="container-base-margin-top" :bordered="false">
          <p style="margin: 0; padding: 16px 0;">{{ detailData.remark }}</p>
        </t-card>
      </div>
      <div v-else class="empty-container">
        <t-empty description="出库单不存在" />
      </div>
    </t-card>

    <!-- 原生图片预览（不使用 TDesign 组件） -->
    <div v-if="imagePreviewVisible" class="image-preview-overlay" @click="imagePreviewVisible = false">
      <div class="image-preview-content" @click.stop>
        <button class="preview-close" @click="imagePreviewVisible = false">×</button>
        <img :src="photos[previewIndex]" class="preview-image" alt="预览" />
      </div>
    </div>

    <!-- 打印组件 - 移除v-if，让组件一直存在于DOM中，和生产通知单保持一致 -->
    <PrintDeliveryNote
      :visible="showPrint"
      :records="printRecords"
      :config="printConfig"
      :customer-name="detailData?.customer_name || ''"
      :delivery-method-dict="deliveryMethodDict"
      :express-company-dict="expressCompanyDict"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed, nextTick } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router';
import { PrintIcon } from 'tdesign-icons-vue-next';
import { ChevronLeftIcon } from 'tdesign-icons-vue-next';

import { getStockOutById } from '@/api/production/stock';
import { getPrintNoticeConfig } from '@/api/reference';
import type { DeliveryNoteConfig } from '@/api/print-notice-config';
import { useDictRaw } from '@/hooks/useDict';
import PrintDeliveryNote from '@/components/PrintDeliveryNote.vue';
import PageLoading from '@/components/PageLoading/index.vue';
import { formatDate, formatDateTime } from '@/utils/date';

defineOptions({
  name: 'StockOutDetail',
});

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const detailData = ref<any>(null);
const showPrint = ref(false);
const printConfig = ref<DeliveryNoteConfig | null>(null);
// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: deliveryMethodDict } = useDictRaw('delivery_method');
const { items: expressCompanyDict } = useDictRaw('express_company');

// 图片预览（原生实现，不使用 TDesign 组件）
const imagePreviewVisible = ref(false);
const previewIndex = ref(0);

// 预览图片
const handlePreviewPhoto = (index: number | string) => {
  previewIndex.value = typeof index === 'string' ? parseInt(index) : index;
  imagePreviewVisible.value = true;
};

const printRecords = computed(() => {
  return detailData.value ? [detailData.value] : [];
});

const photos = computed(() => {
  if (!detailData.value?.photos) return [];
  try {
    return JSON.parse(detailData.value.photos);
  } catch {
    return [];
  }
});

const getDeliveryMethodName = (value: string) => {
  if (!Array.isArray(deliveryMethodDict.value)) return value;
  const item = deliveryMethodDict.value.find((d) => d.dict_value === value);
  return item ? item.dict_label : value;
};

const getExpressCompanyName = (value: string) => {
  if (!Array.isArray(expressCompanyDict.value)) return value;
  const item = expressCompanyDict.value.find((d) => d.dict_value === value);
  return item ? item.dict_label : value;
};

const loadDetail = async () => {
  const id = route.params.id;
  if (!id) {
    MessagePlugin.error('送货单ID不存在');
    return;
  }

  loading.value = true;
  try {
    detailData.value = await getStockOutById(Number(id));
  } catch (error) {
    console.error('加载送货单详情失败:', error);
    MessagePlugin.error('加载送货单详情失败');
  } finally {
    loading.value = false;
  }
};

// 加载字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

const loadPrintConfig = async () => {
  try {
    const res = await getPrintNoticeConfig({ config_type: 'delivery_note' });
    if (res && res.basic_info && res.delivery_info && res.detail_columns) {
      printConfig.value = res;
    } else if (res && res.data) {
      printConfig.value = res.data;
    }
  } catch (error) {
    console.error('[StockOutDetail] 加载送货单打印配置失败:', error);
  }
};

const handlePrint = () => {
  if (!detailData.value) {
    MessagePlugin.warning('数据未加载');
    return;
  }
  showPrint.value = true;
  // 延迟一下确保DOM渲染完成
  setTimeout(() => {
    // 监听打印完成或取消
    const afterPrint = () => {
      setTimeout(() => {
        showPrint.value = false;
      }, 500);
      // 移除监听器
      window.removeEventListener('afterprint', afterPrint);
    };
    
    // 兼容性处理：某些浏览器不支持 afterprint 事件
    if (window.matchMedia) {
      const mediaQueryList = window.matchMedia('print');
      const mediaQueryListener = (mql: MediaQueryList | MediaQueryListEvent) => {
        if (!mql.matches) {
          afterPrint();
          mediaQueryList.removeEventListener('change', mediaQueryListener);
        }
      };
      mediaQueryList.addEventListener('change', mediaQueryListener);
    } else {
      window.addEventListener('afterprint', afterPrint);
    }
    
    window.print();
  }, 100);
};

const handleBack = async () => {
  // 先关闭打印组件
  showPrint.value = false;
  
  // 等待打印组件卸载
  await nextTick();
  
  // 跳转路由
  router.push('/production/stock-out');
};

// 编辑
const handleEdit = () => {
  router.push('/production/stock-out');
};

// 路由离开前的守卫 - 确保组件正确清理
onBeforeRouteLeave(async (to, from, next) => {
  // 关闭打印组件
  showPrint.value = false;
  
  // 等待 DOM 更新
  await nextTick();
  
  // 再等待一个 tick 确保 TDesign 组件清理完成
  await nextTick();
  
  // 允许导航
  next();
});

// 组件卸载时清理
onUnmounted(() => {
  // 清理打印组件状态
  showPrint.value = false;
});

onMounted(async () => {
  // 字典已通过 useDictRaw 自动加载
  loadPrintConfig();
  // 再加载详情数据
  loadDetail();
});
</script>

<style lang="less" scoped>
.stock-out-detail-page {
  padding: 16px;
  background-color: var(--td-bg-color-page, #f5f5f5);
  min-height: 100vh;
}

.loading-container,
.empty-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.back-button-container {
  margin-bottom: 16px;
}

.container-base-margin-top {
  margin-top: 16px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--td-comp-margin-xl);
  padding-bottom: var(--td-comp-paddingTB-l);
  border-bottom: 1px solid var(--td-component-border);
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: var(--td-comp-margin-m);
}

.order-code {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--td-text-color-primary);
}

// 照片网格布局
.photos-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 8px;
  margin-top: 12px;
}

.photo-thumbnail {
  width: 100px;
  height: 100px;
  border-radius: 4px;
  cursor: pointer;
  object-fit: cover;
  transition: opacity 0.2s;

  &:hover {
    opacity: 0.8;
  }
}

// 原生图片预览
.image-preview-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  cursor: pointer;
}

.image-preview-content {
  position: relative;
  max-width: 90vw;
  max-height: 90vh;
  cursor: default;
}

.preview-image {
  max-width: 100%;
  max-height: 90vh;
  display: block;
}

.preview-close {
  position: absolute;
  top: -40px;
  right: 0;
  width: 32px;
  height: 32px;
  background: rgba(255, 255, 255, 0.2);
  border: none;
  border-radius: 50%;
  color: #fff;
  font-size: 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;

  &:hover {
    background: rgba(255, 255, 255, 0.3);
  }
}
</style>

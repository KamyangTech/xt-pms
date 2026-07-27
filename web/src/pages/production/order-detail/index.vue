<template>
  <div class="production-order-detail-page">
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
      <div v-else-if="order">
        <!-- 页面头部 -->
        <div class="page-header">
          <div class="header-left">
            <h2 class="order-code">
              {{ order.order_code }}
              <!-- ✨ 等待模具提示 - 小标签 -->
              <t-tag
                v-if="order.dependency_status === 'WAITING_MOLD'"
                theme="warning"
                variant="light"
                size="medium"
                style="margin-left: 12px; vertical-align: middle"
              >
                <template #icon>
                  <t-icon name="time" />
                </template>
                等待模具试模
              </t-tag>
            </h2>
            <t-space>
              <!-- 订单状态标签 - 使用字典数据 -->
              <t-tag
                shape="round"
                v-if="getOrderStatusDictItem(order.status)"
                v-bind="getOrderStatusTagProps(order.status)"
              >
                {{ getOrderStatusLabel(order.status) }}
              </t-tag>
              <!-- 优先级标签 - 使用字典数据 -->
              <t-tag
                shape="round"
                v-if="getPriorityDictItem(order.priority)"
                v-bind="getPriorityTagProps(order.priority)"
              >
                {{ getPriorityLabel(order.priority) }}
              </t-tag>
            </t-space>
          </div>
          <div class="header-right">
            <t-button theme="default" @click="handlePrint">
              <template #icon><print-icon /></template>
              打印
            </t-button>
            <!-- 报工按钮 - 使用字典语义标签判断 -->
            <t-button v-if="canReportWork" theme="success" @click="handleReportWork">报工</t-button>
            <t-button theme="primary" @click="handleEdit">编辑</t-button>
          </div>
        </div>

        <!-- 基本信息 -->
        <t-card :bordered="false" title="基本信息">
          <t-descriptions :column="2" bordered>
            <t-descriptions-item label="订单号">{{ order.order_code }}</t-descriptions-item>
            <t-descriptions-item label="客户名称">{{ order.customer?.cus_name || '-' }}</t-descriptions-item>
            <t-descriptions-item label="生产类型">
              <t-tag variant="outline" v-bind="getProductionTypeTagProps(order.production_type)">{{ getProductionTypeLabel(order.production_type) }}</t-tag>
            </t-descriptions-item>
            <t-descriptions-item label="模具来源">
              <t-tag variant="outline" v-bind="getMoldSourceTypeTagProps(order.mold_source_type)">{{ getMoldSourceLabel(order.mold_source_type) }}</t-tag>
            </t-descriptions-item>
          </t-descriptions>
        </t-card>

        <!-- 生产信息 -->
        <t-card title="生产信息" class="container-base-margin-top" :bordered="false">
          <t-descriptions :column="3" bordered>
            <t-descriptions-item label="计划数量">{{ order.plan_qty }}</t-descriptions-item>
            <t-descriptions-item label="完成数量">{{ order.completed_qty || 0 }}</t-descriptions-item>
            <t-descriptions-item label="合格数量">{{ order.ok_qty || 0 }}</t-descriptions-item>
            <t-descriptions-item label="不良数量">{{ order.bad_qty || 0 }}</t-descriptions-item>
            <t-descriptions-item label="计划开始日期">{{ formatDate(order.plan_start_date) }}</t-descriptions-item>
            <t-descriptions-item label="计划结束日期">{{ formatDate(order.plan_end_date) }}</t-descriptions-item>
            <t-descriptions-item label="实际开始日期">{{ formatDate(order.actual_start_date) }}</t-descriptions-item>
            <t-descriptions-item label="实际结束日期">{{ formatDate(order.actual_end_date) }}</t-descriptions-item>

            <t-descriptions-item label="负责人">{{ getWorkerName(order.worker_id) }}</t-descriptions-item>
          </t-descriptions>
        </t-card>

        <!-- 外发信息 -->
        <t-card v-if="order.production_type === 'OUTSOURCED'" title="外发信息" class="container-base-margin-top" :bordered="false">
          <t-descriptions :column="3" bordered>
            <t-descriptions-item label="加工商">{{ order.processor_name || '-' }}</t-descriptions-item>
            <t-descriptions-item label="外发单价">¥{{ order.outsource_price?.toFixed(2) || '0.00' }}</t-descriptions-item>
            <t-descriptions-item label="外发总金额">¥{{ order.outsource_total_amount?.toFixed(2) || '0.00' }}</t-descriptions-item>
          </t-descriptions>
        </t-card>

        <!-- 订单明细 -->
        <t-card title="订单明细" class="container-base-margin-top" :bordered="false">
          <t-table
            v-if="order.items && order.items.length > 0"
            :data="order.items"
            :columns="itemColumns"
            row-key="id"
            bordered
          >
            <template #product_name="{ row }">
              {{ row.product_name || row.product?.pro_name || '-' }}
            </template>
            <template #mold_code="{ row }">
              <span class="mold-code-highlight">{{ row.mold?.mold_code || row.external_mold_code || '-' }}</span>
            </template>
            <template #product_image="{ row }">
              <img
                v-if="row.product_image_url"
                :src="getThumbnailUrl(row.product_image_url)"
                class="product-image-thumb"
                @click="handlePreviewImage(getImageUrl(row.product_image_url))"
              />
              <div v-else class="product-image-placeholder">无图片</div>
            </template>
            <template #progress="{ row }">
              <t-progress
                v-if="row.plan_qty > 0"
                :percentage="Math.round((row.completed_qty / row.plan_qty) * 100)"
                :label="true"
              />
              <span v-else>-</span>
            </template>
          </t-table>
          <t-empty v-else description="暂无明细信息" />
        </t-card>

        <!-- 报工记录 -->
        <t-card v-if="reportList.length > 0" title="报工记录" class="container-base-margin-top" :bordered="false">
          <t-table
            :data="reportList"
            :columns="reportColumns"
            row-key="id"
            bordered
          >
            <template #report_date="{ row }">
              {{ formatDate(row.report_date) }}
            </template>
            <template #shift="{ row }">
              <t-tag
                v-if="getShiftDictItem(row.shift)"
                v-bind="getShiftTagProps(row.shift)"
              >
                {{ getShiftLabel(row.shift) }}
              </t-tag>
              <span v-else>-</span>
            </template>
          </t-table>
        </t-card>

        <!-- 成本信息 -->
        <t-card title="成本信息" class="container-base-margin-top" :bordered="false">
          <t-descriptions :column="3" bordered>
            <t-descriptions-item label="材料成本">¥{{ order.material_cost?.toFixed(2) || '0.00' }}</t-descriptions-item>
            <t-descriptions-item label="加工成本">¥{{ order.processing_cost?.toFixed(2) || '0.00' }}</t-descriptions-item>
            <t-descriptions-item label="总成本">¥{{ order.total_cost?.toFixed(2) || '0.00' }}</t-descriptions-item>
            <t-descriptions-item label="销售单价">¥{{ order.unit_price?.toFixed(2) || '0.00' }}</t-descriptions-item>
            <t-descriptions-item label="总金额">¥{{ order.total_amount?.toFixed(2) || '0.00' }}</t-descriptions-item>
          </t-descriptions>
        </t-card>

        <!-- 备注 -->
        <t-card title="备注" class="container-base-margin-top" :bordered="false">
          <p style="margin: 0; padding: 16px 0;">{{ order.remark || '无' }}</p>
        </t-card>

        <!-- 时间信息 -->
        <t-card title="时间信息" class="container-base-margin-top" :bordered="false">
          <t-descriptions :column="2" bordered>
            <t-descriptions-item label="创建时间">{{ formatDateTime(order.create_time) }}</t-descriptions-item>
            <t-descriptions-item label="更新时间">{{ formatDateTime(order.update_time) }}</t-descriptions-item>
          </t-descriptions>
        </t-card>
      </div>
      <div v-else class="empty-container">
        <t-empty description="订单不存在" />
      </div>
    </t-card>



    <!-- 报工对话框 -->
    <t-dialog
      v-model:visible="reportDialogVisible"
      header="生产报工"
      width="600px"
      @confirm="confirmReport"
      @cancel="reportDialogVisible = false"
     :close-on-overlay-click="false">
      <t-form :model="reportForm" :label-width="100" style="margin-top: 16px">
        <t-form-item label="报工日期">
          <t-date-picker
            v-model="reportForm.report_date"
            format="YYYY-MM-DD"
            placeholder="选择日期"
          />
        </t-form-item>
        <t-form-item label="班次">
          <t-select v-model="reportForm.shift" placeholder="请选择班次">
            <t-option
              v-for="item in shiftDict"
              :key="item.dict_value"
              :value="item.dict_value"
              :label="item.dict_label"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="良品数量">
          <t-input-number v-model="reportForm.ok_qty" placeholder="请输入良品数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="不良数量">
          <t-input-number v-model="reportForm.bad_qty" placeholder="请输入不良数量" style="width: 100%" />
        </t-form-item>
        <t-form-item label="报废数量">
          <t-input-number v-model="reportForm.scrap_qty" :min="0" placeholder="请输入报废数量" />
        </t-form-item>
        <t-form-item label="工时(小时)">
          <t-input-number v-model="reportForm.work_hours" :min="0" :decimal-places="2" placeholder="请输入工时" />
        </t-form-item>
        <t-form-item label="设备">
          <t-input v-model="reportForm.device_id" placeholder="请输入使用的设备" />
        </t-form-item>
        <t-form-item label="操作工">
          <t-input v-model="reportForm.worker_id" placeholder="请输入操作工" />
        </t-form-item>
        <t-form-item label="不良原因">
          <t-textarea v-model="reportForm.defect_reason" placeholder="请输入不良原因" :rows="2" />
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="reportForm.remark" placeholder="请输入备注" :rows="2" />
        </t-form-item>
      </t-form>
    </t-dialog>

    <!-- 图片预览 -->
    <t-image-viewer
      v-model:visible="imagePreviewVisible"
      :images="previewImages"
      :default-index="previewIndex"
    />

    <!-- 打印通知单组件 -->
    <PrintNotice
      :visible="isPrinting"
      :order="order"
      :config="printConfig"
      :production-type-dict="productionTypeDict"
      :mold-source-dict="moldSourceDict"
      :employee-dict="employeeDict"
    />
  </div>
</template>

<script setup lang="ts">
import { MessagePlugin } from 'tdesign-vue-next';
import { PrintIcon } from 'tdesign-icons-vue-next';
import dayjs from 'dayjs';
import { computed, nextTick, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { getProductionOrderById, submitReport, getReports } from '@/api/production/order';
import { getPrintNoticeConfig } from '@/api/reference';
import { useDictRaw } from '@/hooks/useDict';
import type { DeliveryNoteConfig, PrintNoticeConfig } from '@/api/print-notice-config';
import { getDictColorTheme, getDictColorRenderConfig, getDictColorRenderConfigForOutline } from '@/utils/dict-color';
import type { ProductionOrder, ProductionReport } from '@/types/production';
import PrintNotice from '@/components/PrintNotice.vue';
import ImagePlaceholder from '@/components/ImagePlaceholder/index.vue';
import PageLoading from '@/components/PageLoading/index.vue';
import { formatDate, formatDateTime } from '@/utils/date';
import { getThumbnailUrl } from '@/api/common/upload';

defineOptions({
  name: 'ProductionOrderDetail',
});

const route = useRoute();
const router = useRouter();

// ✨ orderId 改为计算属性：响应路由参数变化，避免不同卡片之间复用时取到旧值
const orderId = computed(() => Number(route.params.id));

const order = ref<ProductionOrder | null>(null);
const loading = ref(false);
const reportDialogVisible = ref(false);
const reportList = ref<ProductionReport[]>([]);
const isPrinting = ref(false);
const printConfig = ref<PrintNoticeConfig | null>(null);

// 图片预览相关
const imagePreviewVisible = ref(false);
const previewImages = ref<string[]>([]);
const previewIndex = ref(0);

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: productionTypeDict } = useDictRaw('production-type');
const { items: moldSourceDict } = useDictRaw('mold-source');
const { items: employeeDict } = useDictRaw('responsible-employee');
const { items: orderStatusDict } = useDictRaw('production-order-status');
const { items: priorityDict } = useDictRaw('order-priority', (items) =>
  (items as any[]).filter((item: any) => item.status === 1),
);
const { items: shiftDict } = useDictRaw('shift-type');

// 报工表单
const reportForm = ref({
  order_id: orderId.value,
  report_date: dayjs().format('YYYY-MM-DD') as string,
  shift: undefined as string | undefined,
  ok_qty: undefined as number | undefined,
  bad_qty: undefined as number | undefined,
  scrap_qty: undefined as number | undefined,
  work_hours: undefined as number | undefined,
  device_id: undefined as string | undefined,
  worker_id: undefined as string | undefined,
  defect_reason: '',
  remark: '',
});

// 明细列定义
const itemColumns = [
  { title: '产品名称', colKey: 'product_name', width: 150 },
  { title: '模具编号', colKey: 'mold_code', width: 150 },
  { title: '产品图片', colKey: 'product_image', width: 100 },
  { title: '计划数量', colKey: 'plan_qty', width: 100 },
  { title: '完成数量', colKey: 'completed_qty', width: 100 },
  { title: '进度', colKey: 'progress', width: 150 },
];

// 报工记录列定义
const reportColumns = [
  { title: '报工日期', colKey: 'report_date', width: 120 },
  { title: '班次', colKey: 'shift', width: 80 },
  { title: '良品数量', colKey: 'ok_qty', width: 100 },
  { title: '不良数量', colKey: 'bad_qty', width: 100 },
  { title: '报废数量', colKey: 'scrap_qty', width: 100 },
  { title: '工时', colKey: 'work_hours', width: 80 },
  { title: '设备', colKey: 'device_id', width: 120 },
  { title: '操作工', colKey: 'worker_id', width: 100 },
  { title: '备注', colKey: 'remark', ellipsis: true },
];

const loadOrderDetail = async () => {
  loading.value = true;
  try {
    order.value = await getProductionOrderById(orderId.value);
    if (order.value && !['PENDING_SCHEDULE', 'PRODUCING'].includes(order.value.status)) {
      await loadReportList();
    }
  } catch (error) {
    console.error('加载订单详情失败:', error);
    MessagePlugin.error('加载订单详情失败');
  } finally {
    loading.value = false;
  }
};

const loadReportList = async () => {
  try {
    reportList.value = await getReports(orderId.value);
  } catch (error) {
    console.error('加载报工记录失败:', error);
  }
};

// 打开报工对话框
const handleReportWork = () => {
  reportForm.value = {
    order_id: orderId.value,
    report_date: dayjs().format('YYYY-MM-DD'),
    shift: undefined as string | undefined,
    ok_qty: undefined,
    bad_qty: undefined,
    scrap_qty: undefined,
    work_hours: undefined,
    device_id: '',
    worker_id: '',
    defect_reason: '',
    remark: '',
  };
  reportDialogVisible.value = true;
};

// 确认报工
const confirmReport = async () => {
  try {
    if (!reportForm.value.ok_qty && !reportForm.value.bad_qty) {
      MessagePlugin.warning('请至少填写良品或不良数量');
      return;
    }

    await submitReport(reportForm.value);
    MessagePlugin.success('报工成功');
    reportDialogVisible.value = false;
    // 刷新订单详情和报工记录
    await loadOrderDetail();
  } catch (error: any) {
    console.error('报工失败:', error);
    MessagePlugin.error(error.message || '报工失败');
  }
};

// 加载字典数据（已替换为 useDictRaw，自动加载 + 全局缓存）

const getProductionTypeLabel = (type: string | undefined) => {
  if (!type) return '-';
  const dictItem = productionTypeDict.value.find(
    (item) => item.dict_value === type
  );
  return dictItem?.dict_label || '-';
};

const getProductionTypeTheme = (type: string | undefined): 'default' | 'primary' | 'warning' | 'danger' | 'success' => {
  if (!type) return 'default';
  const dictItem = productionTypeDict.value.find(
    (item) => item.dict_value === type
  );
  if (!dictItem || !dictItem.color) return 'default';
  
  // 使用统一工具类获取 TDesign 主题
  return getDictColorTheme(dictItem.color);
};

const getProductionTypeTagProps = (type: string | undefined) => {
  if (!type) return {};
  const dictItem = productionTypeDict.value.find(
    (item) => item.dict_value === type
  );
  if (!dictItem || !dictItem.color) return {};
  
  return getDictColorRenderConfigForOutline(dictItem.color);
};

const getMoldSourceLabel = (type: string | undefined) => {
  if (!type) return '-';
  const dictItem = moldSourceDict.value.find(
    (item) => item.dict_value === type
  );
  return dictItem?.dict_label || '-';
};

const getMoldSourceTypeTheme = (type: string | undefined): 'default' | 'primary' | 'warning' | 'danger' | 'success' => {
  if (!type) return 'default';
  const dictItem = moldSourceDict.value.find(
    (item) => item.dict_value === type
  );
  if (!dictItem || !dictItem.color) return 'default';
  
  // 使用统一工具类获取 TDesign 主题
  return getDictColorTheme(dictItem.color);
};

const getMoldSourceTypeTagProps = (type: string | undefined) => {
  if (!type) return {};
  const dictItem = moldSourceDict.value.find(
    (item) => item.dict_value === type
  );
  if (!dictItem || !dictItem.color) return {};
  
  return getDictColorRenderConfigForOutline(dictItem.color);
};

const getWorkerName = (workerId: string | undefined) => {
  if (!workerId) return '-';
  const dictItem = employeeDict.value.find(
    (item) => item.dict_value === workerId
  );
  return dictItem?.dict_label || workerId;
};

// ==================== 订单状态字典方法 ====================
const getOrderStatusDictItem = (status: string | undefined) => {
  if (!status) return null;
  return orderStatusDict.value.find(
    (item) => item.dict_value === status
  );
};

const getOrderStatusLabel = (status: string | undefined) => {
  const dictItem = getOrderStatusDictItem(status);
  return dictItem?.dict_label || '-';
};

const getOrderStatusTheme = (status: string | undefined): 'default' | 'primary' | 'warning' | 'danger' | 'success' => {
  if (!status) return 'default';
  const dictItem = getOrderStatusDictItem(status);
  if (!dictItem || !dictItem.color) return 'default';
  return getDictColorTheme(dictItem.color);
};

const getOrderStatusTagProps = (status: string | undefined) => {
  if (!status) return {};
  const dictItem = getOrderStatusDictItem(status);
  if (!dictItem || !dictItem.color) return {};
  
  return getDictColorRenderConfig(dictItem.color);
};

const canReportWork = computed(() => {
  if (!order.value || !order.value.status) return false;
  
  // ✨ 检查依赖状态：如果是 WAITING_MOLD，不允许报工
  if (order.value.dependency_status === 'WAITING_MOLD') {
    return false;
  }
  
  const dictItem = getOrderStatusDictItem(order.value.status);
  // 使用语义标签判断，而非硬编码状态值
  return (dictItem as any)?.semantic_tag === 'processing';
});

// ==================== 优先级字典方法 ====================
const getPriorityDictItem = (priority: string | undefined) => {
  if (!priority) return null;
  return priorityDict.value.find(
    (item) => item.dict_value === priority
  );
};

const getPriorityLabel = (priority: string | undefined) => {
  const dictItem = getPriorityDictItem(priority);
  return dictItem?.dict_label || '-';
};

const getPriorityTheme = (priority: string | undefined): 'default' | 'primary' | 'warning' | 'danger' | 'success' => {
  if (!priority) return 'default';
  const dictItem = getPriorityDictItem(priority);
  if (!dictItem || !dictItem.color) return 'default';
  return getDictColorTheme(dictItem.color);
};

const getPriorityTagProps = (priority: string | undefined) => {
  if (!priority) return {};
  const dictItem = getPriorityDictItem(priority);
  if (!dictItem || !dictItem.color) return {};
  
  return getDictColorRenderConfig(dictItem.color);
};

// ==================== 班次字典方法 ====================
const getShiftDictItem = (shift: string | undefined) => {
  if (!shift) return null;
  return shiftDict.value.find(
    (item) => item.dict_value === shift
  );
};

const getShiftLabel = (shift: string | undefined) => {
  const dictItem = getShiftDictItem(shift);
  return dictItem?.dict_label || '-';
};

const getShiftTagProps = (shift: string | undefined) => {
  if (!shift) return {};
  const dictItem = getShiftDictItem(shift);
  if (!dictItem || !dictItem.color) return {};
  
  return getDictColorRenderConfig(dictItem.color);
};


const getImageUrl = (url: string) => {
  if (!url) return '';
  if (url.startsWith('http')) {
    return url;
  }
  const baseUrl = import.meta.env.VITE_API_URL || '';
  return `${baseUrl}${url}`;
};

const handleBack = () => {
  router.back();
};

// 编辑
const handleEdit = () => {
  router.push(`/production/order/edit/${orderId.value}`);
};

const handlePrint = () => {
  isPrinting.value = true;
  // 等待图片加载完成后再打印，但设置超时避免长时间等待
  setTimeout(() => {
    const images = document.querySelectorAll('.print-template-container .print-product-image');

    // 如果没有图片，直接打印
    if (images.length === 0) {
      window.print();
      setTimeout(() => {
        isPrinting.value = false;
      }, 100);
      return;
    }

    const imagePromises = Array.from(images).map((img) => {
      return new Promise((resolve) => {
        const htmlImg = img as HTMLImageElement;
        // 已加载完成且 naturalWidth > 0 才视为成功
        if (htmlImg.complete && htmlImg.naturalWidth > 0) {
          resolve(true);
        } else if (htmlImg.complete && htmlImg.naturalWidth === 0) {
          // 已经加载失败，不再等待
          resolve(false);
        } else {
          htmlImg.onload = () => resolve(true);
          htmlImg.onerror = () => resolve(false);
        }
      });
    });

    // 设置5秒超时，避免图片加载太慢导致长时间等待
    const timeoutPromise = new Promise((resolve) => {
      setTimeout(() => resolve(false), 5000);
    });

    Promise.race([Promise.all(imagePromises), timeoutPromise]).then(() => {
      setTimeout(() => {
        window.print();
        setTimeout(() => {
          isPrinting.value = false;
        }, 100);
      }, 300);
    });
  }, 100);
};

// 图片预览
const handlePreviewImage = (imageUrl: string) => {
  previewImages.value = [imageUrl];
  previewIndex.value = 0;
  imagePreviewVisible.value = true;
};

const loadPrintConfig = async () => {
  try {
    const res = await getPrintNoticeConfig();
    // 响应拦截器已经提取了 data 字段，res 就是配置数据对象
    if (res && res.basic_info && res.production_info && res.detail_columns) {
      printConfig.value = res;
    } else if (res && res.data) {
      // 兼容处理
      printConfig.value = res.data;
    } else {
      console.warn('[OrderDetail] 配置数据格式不正确:', res);
    }
  } catch (error) {
    console.error('[OrderDetail] 加载打印配置失败:', error);
  }
};

onMounted(() => {
  // 字典已通过 useDictRaw 自动加载
  loadOrderDetail();
  loadPrintConfig();

  // 检查URL参数，如果有print=1，则在数据加载完成后自动触发打印
  if (route.query.print === '1') {
    // 等待订单数据加载完成后再打印
    const checkAndPrint = () => {
      if (order.value && !loading.value) {
        // 延迟一点确保DOM渲染完成
        setTimeout(() => {
          handlePrint();
        }, 500);
      } else {
        setTimeout(checkAndPrint, 200);
      }
    };
    checkAndPrint();
  }
});

// ✨ 路由参数变化时彻底重置数据，避免同一组件实例被复用时残留旧数据
watch(
  () => route.params.id,
  (newId) => {
    if (newId && route.name === 'ProductionOrderDetail') {
      order.value = null;
      reportList.value = [];
      printConfig.value = null;
      reportDialogVisible.value = false;
      imagePreviewVisible.value = false;

      // 重新加载数据（使用最新的 orderId.value）
      loadOrderDetail();
      loadPrintConfig();
    }
  },
);
</script>

<style lang="less" scoped>
.production-order-detail-page {
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

.info-card {
  margin-bottom: var(--td-comp-margin-l);

  :deep(.t-card__header) {
    padding-bottom: var(--td-comp-paddingTB-s);
  }

  :deep(.t-card__body) {
    padding-top: var(--td-comp-paddingTB-s);
  }
}

.mold-code-highlight {
  font-weight: 600;
  color: var(--td-text-color-primary);
}

.t-descriptions {
  :deep(.t-descriptions__body) {
    background-color: transparent;
  }
}

/* 产品图片样式 - 统一使用70x70圆角 */
.product-image-thumb {
  width: 70px;
  height: 70px;
  object-fit: cover;
  border-radius: 4px;
  cursor: pointer;
  transition: opacity 0.2s;
}

.product-image-thumb:hover {
  opacity: 0.8;
}

.product-image-placeholder {
  width: 70px;
  height: 70px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
  border-radius: 4px;
  color: #999;
  font-size: 12px;
}
</style>

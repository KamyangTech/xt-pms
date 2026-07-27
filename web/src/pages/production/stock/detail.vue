<template>
  <div class="stock-detail-container">
    <!-- 返回按钮 -->
    <div class="back-button-container">
      <t-button theme="default" @click="handleBack" variant="text">
        <template #icon><t-icon name="chevron-left" /></template>
        返回
      </t-button>
    </div>

    <!-- 产品信息和折线图 -->
    <div class="stats-row">
      <!-- 左侧产品信息卡片 -->
      <div class="card-wrapper">
        <t-card title="产品信息" :bordered="false" class="mold-info-card">
          <div v-if="!infoLoading" class="mold-info-content">
              <!-- 产品名称和模具编号 -->
              <div class="header-row">
                <div class="mold-name" :title="moldInfo.mold_name || '-'">{{ moldInfo.mold_name || '-' }}</div>
                <div class="mold-code">
                  <span class="detail-label">模具编号：</span>
                  <span class="detail-value">{{ moldInfo.mold_code || '-' }}</span>
                </div>
              </div>
              
              <!-- 详细信息双列展示 -->
              <div class="details-grid">
                <div class="detail-item">
                  <span class="detail-label">材质：</span>
                  <span class="detail-value">{{ moldInfo.material || '-' }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">重量：</span>
                  <span class="detail-value">{{ moldInfo.weight || '-' }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">总入库数量：</span>
                  <span class="detail-value highlight">{{ moldInfo.total_stock_in || 0 }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">总出库数量：</span>
                  <span class="detail-value">{{ moldInfo.total_stock_out || 0 }}</span>
                </div>
                <div class="detail-item">
                  <span class="detail-label">当前库存总数：</span>
                  <span class="detail-value highlight">{{ moldInfo.current_stock || 0 }}</span>
                </div>
              </div>
              
              <!-- 产品图片 -->
              <div class="mold-images">
                <template v-if="displayImages.length > 0">
                  <t-image-viewer :images="displayImages">
                    <template #trigger="{ open }">
                      <ImagePlaceholder
                        v-for="(img, index) in displayImages"
                        :key="index"
                        :src="img"
                        :width="80"
                        :height="80"
                        shape="square"
                        @click="open({ index })"
                      />
                    </template>
                  </t-image-viewer>
                </template>
                <ImagePlaceholder v-else :width="80" :height="80" shape="square" />
              </div>
          </div>
        </t-card>
      </div>

      <!-- 右侧折线图 -->
      <div class="card-wrapper">
        <t-card title="出入库趋势" :bordered="false" class="chart-card">
          <template #actions>
            <t-radio-group v-model="trendDays" variant="default-filled" @change="handleTrendChange">
              <t-radio-button :value="7">近7天</t-radio-button>
              <t-radio-button :value="15">近15天</t-radio-button>
              <t-radio-button :value="30">近30天</t-radio-button>
            </t-radio-group>
          </template>
          <div id="trendChart" v-loading="chartLoading" style="width: 100%; height: 300px"></div>
        </t-card>
      </div>
    </div>

    <!-- 出入库记录表格 -->
    <t-card title="出入库记录" :bordered="false" class="table-card">
      <template #actions>
        <t-space :size="12">
          <t-select
            v-model="statusFilters"
            placeholder="筛选状态"
            clearable
            multiple
            style="width: 200px"
          >
            <t-option value="in" label="入库" />
            <t-option value="shipment" label="发货" />
            <t-option
              v-for="item in outReasonDict"
              :key="item.dict_value"
              :value="item.dict_value"
              :label="item.dict_label"
            />
          </t-select>
          <t-radio-group v-model="sortOrder" variant="default-filled" @change="handleSortChange">
            <t-radio-button value="desc">时间倒序</t-radio-button>
            <t-radio-button value="asc">时间正序</t-radio-button>
          </t-radio-group>
          <t-button theme="primary" @click="handleNavigateToStockOut" :loading="infoLoading">
            <template #icon><add-icon /></template>
            发货
          </t-button>
          <t-button theme="default" @click="handleStockOut" :loading="infoLoading" :disabled="infoLoading">
            <template #icon>
              <add-icon v-if="!infoLoading" />
            </template>
            盘库
          </t-button>
        </t-space>
      </template>
      <t-table
        :data="tableData"
        :columns="columns"
        row-key="id"
        :pagination="pagination"
        :loading="tableLoading"
        :scroll="{ x: 1580 }"
        @page-change="handlePageChange"
      >
        <template #type="{ row }">
          <t-tag v-if="row.type === 'in'" theme="primary">入库</t-tag>
          <t-tag v-else-if="row.out_type === 'other'" :theme="getOutReasonTheme(row.out_reason)">
            {{ getOutReasonLabel(row.out_reason) }}
          </t-tag>
          <t-tag v-else theme="success">发货</t-tag>
        </template>
        <template #stock_code="{ row }">
          <span v-if="row.type === 'in'" class="stock-code-text">{{ row.stock_code }}</span>
          <span v-else-if="row.out_type !== 'other'" class="stock-code-text">{{ row.stock_code }}</span>
          <span v-else>-</span>
        </template>
        <template #order_code="{ row }">
          <t-link v-if="row.order" @click="handleViewOrder(row)" class="order-code-link">
            {{ row.order.order_code || '-' }}
          </t-link>
          <span v-else>-</span>
        </template>
        <template #product_name="{ row }">
          {{ row.product_name || row.mold?.mold_name || '-' }}
        </template>
        <template #warehouse_name="{ row }">
          {{ row.warehouse?.wh_name || '-' }}
        </template>
        <template #quantity="{ row }">
          <span :class="row.type === 'in' ? 'text-success' : 'text-warning'">
            {{ row.type === 'in' ? '+' : '-' }}{{ Math.abs(row.quantity) }}
          </span>
        </template>
        <template #date="{ row }">
          {{ formatDateTime(row.record_date) }}
        </template>
        <template #operation="{ row }">
          <t-space>
            <t-link theme="primary" @click="handleEdit(row)">编辑</t-link>
          </t-space>
        </template>
      </t-table>
    </t-card>

    <!-- 编辑对话框 -->
    <t-dialog
      v-model:visible="editDialogVisible"
      :header="editDialogTitle"
      width="600px"
      :confirm-btn="{ content: '确定', theme: 'primary', loading: editLoading }"
      :cancel-btn="{ content: '取消' }"
      @confirm="handleEditConfirm"
      @close="handleEditClose"
     :close-on-overlay-click="false">
      <t-form ref="editFormRef" :data="editFormData" label-width="100px">
        <!-- 入库记录编辑 -->
        <template v-if="editFormData.type === 'in'">
          <t-form-item label="入库单号">
            <t-input v-model="editFormData.stock_code" disabled />
          </t-form-item>
          <t-form-item label="订单编号">
            <t-input :value="editFormData.order?.order_code || '-'" disabled />
          </t-form-item>
          <t-form-item label="产品名称">
            <t-input :value="editFormData.product_name || editFormData.mold?.mold_name || '-'" disabled />
          </t-form-item>
          <t-form-item label="入库数量">
            <t-input-number v-model="editFormData.quantity" :min="1" :step="1" placeholder="请输入入库数量" />
          </t-form-item>
          <t-form-item label="入库日期">
            <t-date-picker 
              v-model="editFormData.stock_in_date" 
              placeholder="请选择入库日期" 
              clearable 
              enable-time-picker 
              format="YYYY-MM-DD HH:mm:ss" 
            />
          </t-form-item>
          <t-form-item label="批次号">
            <t-input v-model="editFormData.batch_no" placeholder="自动生成" disabled />
          </t-form-item>
          <t-form-item label="备注">
            <t-textarea v-model="editFormData.remark" placeholder="请输入备注" :autosize="{ minRows: 3, maxRows: 6 }" />
          </t-form-item>
        </template>
        <!-- 出库记录编辑 -->
        <template v-else>
          <t-form-item label="出库单号">
            <t-input v-model="editFormData.stock_code" disabled />
          </t-form-item>
          <t-form-item label="产品名称">
            <t-input :value="editFormData.product_name || '-'" disabled />
          </t-form-item>
          <t-form-item label="出库类型">
            <t-input :value="getOutTypeLabel(editFormData.out_type)" disabled />
          </t-form-item>
          <t-form-item label="出库原因">
            <t-input :value="getOutReasonLabel(editFormData.out_reason)" disabled />
          </t-form-item>
          <t-form-item label="出库数量">
            <t-input :value="editFormData.quantity" disabled />
          </t-form-item>
          <t-form-item label="仓库名称">
            <t-input :value="editFormData.warehouse?.wh_name || '-'" disabled />
          </t-form-item>
          <t-form-item label="出库日期">
            <t-date-picker 
              v-model="editFormData.stock_out_date" 
              placeholder="请选择出库日期" 
              clearable 
              enable-time-picker 
              format="YYYY-MM-DD HH:mm:ss" 
            />
          </t-form-item>
          <!-- 发货相关字段：仅发货(shipment)类型显示 -->
          <template v-if="editFormData.out_type === 'shipment'">
            <t-form-item label="客户名称">
              <t-input v-model="editFormData.customer_name" placeholder="请输入客户名称" clearable />
            </t-form-item>
            <t-form-item label="发货日期">
              <t-date-picker
                v-model="editFormData.delivery_date"
                placeholder="请选择发货日期"
                clearable
                enable-time-picker
                format="YYYY-MM-DD HH:mm:ss"
              />
            </t-form-item>
            <t-form-item label="发货方式">
              <t-input v-model="editFormData.delivery_method" placeholder="请输入发货方式" clearable />
            </t-form-item>
            <t-form-item label="快递公司">
              <t-input v-model="editFormData.express_company" placeholder="请输入快递公司" clearable />
            </t-form-item>
            <t-form-item label="快递单号">
              <t-input v-model="editFormData.express_no" placeholder="请输入快递单号" clearable />
            </t-form-item>
          </template>
          <t-form-item label="备注">
            <t-textarea v-model="editFormData.remark" placeholder="请输入备注" :autosize="{ minRows: 3, maxRows: 6 }" />
          </t-form-item>
        </template>
      </t-form>
    </t-dialog>

    <!-- 盘库对话框 -->
    <t-dialog
      v-model:visible="stockOutDialogVisible"
      header="盘点库存"
      width="480px"
      :confirm-btn="{ content: '确定', theme: 'primary', loading: stockOutLoading }"
      :cancel-btn="{ content: '取消' }"
      @confirm="handleStockOutConfirm"
      @close="handleStockOutClose"
     :close-on-overlay-click="false">
      <t-alert
        theme="info"
        message="输入实际盘点数量，系统将自动矫正库存并记录差异"
        style="margin-bottom: 16px"
      />
      <t-form ref="stockOutFormRef" :data="stockOutFormData" label-width="80px">
        <t-form-item label="模具编号">
          <t-input :value="moldInfo.mold_code" disabled />
        </t-form-item>
        <t-form-item label="产品名称">
          <t-input :value="moldInfo.mold_name" disabled />
        </t-form-item>
        <t-form-item label="系统库存">
          <t-input :value="String(moldInfo.current_stock)" disabled />
        </t-form-item>
        <t-form-item label="盘点数量" required>
          <t-input-number
            v-model="stockOutFormData.quantity"
            :min="0"
            placeholder="请输入实际盘点数量"
            style="width: 100%"
          />
        </t-form-item>
        <t-form-item v-if="isCheckLoss" label="盘亏原因" required>
          <t-select v-model="stockOutFormData.outReason" placeholder="请选择盘亏原因">
            <t-option
              v-for="item in outReasonDict"
              :key="item.dict_value"
              :label="item.dict_label"
              :value="item.dict_value"
            />
          </t-select>
        </t-form-item>
        <t-form-item label="备注">
          <t-textarea v-model="stockOutFormData.remark" placeholder="请输入备注（可选）" :autosize="{ minRows: 3, maxRows: 6 }" />
        </t-form-item>
      </t-form>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, h, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';
import { DialogPlugin, MessagePlugin } from 'tdesign-vue-next';
import type { FormInstanceFunctions, PrimaryTableCol } from 'tdesign-vue-next';
import { useRoute, useRouter } from 'vue-router';
import { AddIcon, ChevronLeftIcon } from 'tdesign-icons-vue-next';
import * as echarts from 'echarts/core';
import { LineChart } from 'echarts/charts';
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components';
import { CanvasRenderer } from 'echarts/renderers';
import dayjs from 'dayjs';
import 'dayjs/locale/zh-cn';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';

// 配置 dayjs
dayjs.extend(utc);
dayjs.extend(timezone);
dayjs.locale('zh-cn');

import { getStockRecordsByMoldId, getStockRecordsByExternalMoldCode, getStockTrend, updateStockIn, updateStockOut, stockOutOrder, checkStock } from '@/api/production/stock';
import { getMoldById, getWarehouseList, getProductList } from '@/api/reference';
import { useDictRaw } from '@/hooks/useDict';
import { getProductionOrders } from '@/api/production/order';
import { useSettingStore } from '@/store';
import { prefix } from '@/config/global';
import { getLineChartDataSet } from '@/pages/dashboard/base/index';
import { formatDate, formatDateTime } from '@/utils/date';
import ImagePlaceholder from '@/components/ImagePlaceholder/index.vue';

defineOptions({
  name: 'ProductionStockDetail',
});

echarts.use([GridComponent, TooltipComponent, LegendComponent, LineChart, CanvasRenderer]);

const route = useRoute();
const router = useRouter();
const store = useSettingStore();

const isExternalMold = computed(() => !!route.params.externalMoldCode);

// 字典数据（useDictRaw 自动加载 + 全局缓存）
const { items: outTypeDict } = useDictRaw('out_type');
const { items: outReasonDict } = useDictRaw('out_reason');



const getOutTypeLabel = (type: string) => {
  if (type === 'shipment') return '发货';
  const typeItem = outTypeDict.value.find((item: any) => item.dict_value === type);
  return typeItem ? typeItem.dict_label : type || '-';
};

// 获取出库原因标签
const getOutReasonLabel = (reason: string) => {
  const reasonItem = outReasonDict.value.find((item: any) => item.dict_value === reason);
  return reasonItem ? reasonItem.dict_label : '其他';
};

// 获取出库原因主题色
const getOutReasonTheme = (reason: string) => {
  const colorMap: Record<string, any> = {
    'expired': 'danger',
    'damaged': 'warning',
    'unqualified': 'danger',
    'other': 'default',
  };
  return colorMap[reason] || 'default';
};

const infoLoading = ref(true);
const chartLoading = ref(true);

// 模具信息
const moldInfo = ref<any>({
  mold_code: '',
  mold_name: '',
  material: '',
  weight: '',
  total_stock_in: 0,
  total_stock_out: 0,
  current_stock: 0,
  customer_name: '',  // 添加客户名称字段
  images: [],
});

// 显示图片列表
const displayImages = computed(() => {
  if (moldInfo.value.images && moldInfo.value.images.length > 0) {
    return moldInfo.value.images;
  }
  return [];
});

// 趋势数据
const trendDays = ref(7);
let trendChart: echarts.ECharts | null = null;

// 出入库记录排序状态
const sortOrder = ref<'asc' | 'desc'>('desc');

// 状态筛选（多选）
const statusFilters = ref<string[]>([]);

// 表格数据
const tableData = ref<any[]>([]);
const tableLoading = ref(false);

// 分页
const pagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
});

// 表格列定义
const columns = computed<PrimaryTableCol[]>(() => {
  const baseColumns: PrimaryTableCol[] = [
    { title: '类型', colKey: 'type', width: 120, fixed: 'left' },
    { title: '单号', colKey: 'stock_code', width: 180, fixed: 'left' },
    { title: '订单编号', colKey: 'order_code', width: 180 },
    { title: '产品名称', colKey: 'product_name', width: 200 },
    { title: '仓库名称', colKey: 'warehouse_name', width: 150 },
    { title: '数量', colKey: 'quantity', width: 100 },
    { title: '批次号', colKey: 'batch_no', width: 230 },
    { title: '日期', colKey: 'date', width: 160 },
    { title: '备注', colKey: 'remark', width: 280, ellipsis: true },
    { title: '操作', colKey: 'operation', width: 120, fixed: 'right' },
  ];

  return baseColumns;
});

// 缓存的汇总数据（避免重复请求）
const cachedSummary = ref<any>(null);

const fetchMoldInfo = async () => {
  infoLoading.value = true;
  try {
    if (isExternalMold.value) {
      // 外来模具
      const externalMoldCode = route.params.externalMoldCode as string;
      if (!externalMoldCode) {
        infoLoading.value = false;
        return;
      }
      
      const decodedCode = decodeURIComponent(externalMoldCode);
      
      const productId = route.query.product_id;
      const stockParams: any = { page: 1, page_size: 1 };
      if (productId) {
        stockParams.product_id = productId;
      }
      const stockRes = await getStockRecordsByExternalMoldCode(decodedCode, stockParams);
      
      // 缓存汇总数据，供后续使用
      if (stockRes.summary) {
        cachedSummary.value = stockRes.summary;
        const extSummaryImages: string[] = [];
        if (stockRes.summary.product_image_url) {
          extSummaryImages.push(stockRes.summary.product_image_url);
        }
        moldInfo.value = {
            mold_code: stockRes.summary.mold_code || decodedCode,
            mold_name: stockRes.summary.product_name || '-',
            material: '',
            weight: '',
            customer_name: stockRes.summary.customer_name || '',  // 保存客户名称
            images: extSummaryImages,
            total_stock_in: stockRes.summary.total_stock_in || 0,
            total_stock_out: stockRes.summary.total_stock_out || 0,
            current_stock: stockRes.summary.current_stock || 0,
          };
      }
    } else {
      // 自有模具
      const moldId = route.params.id as string;
      if (!moldId) {
        infoLoading.value = false;
        return;
      }
            
      // 并行获取模具详细信息和库存汇总信息
      const productId = route.query.product_id;
      const stockParams: any = { page: 1, page_size: 1 };
      if (productId) {
        stockParams.product_id = productId;
      }
      
      const [moldRes, stockRes] = await Promise.all([
        getMoldById(Number(moldId)),
        getStockRecordsByMoldId(Number(moldId), stockParams),
      ]);
      
      // 缓存汇总数据，供后续使用
      if (stockRes.summary) {
        cachedSummary.value = stockRes.summary;
      }
            
      // 产品库存管理，使用产品图片（来自生产订单项）
      const summaryImage = stockRes.summary?.product_image_url;
      const images = summaryImage ? [summaryImage] : [];

      moldInfo.value = {
        mold_code: moldRes.mold_code || '',
        //  显示产品名称，而不是模具名称
        mold_name: stockRes.summary?.product_name || moldRes.mold_name || '',
        material: moldRes.material || '',
        weight: moldRes.weight || '',
        customer_name: stockRes.summary?.customer_name || '',  // 保存客户名称
        images,
        total_stock_in: stockRes.summary?.total_stock_in || 0,
        total_stock_out: stockRes.summary?.total_stock_out || 0,
        current_stock: stockRes.summary?.current_stock || 0,
      };
    }
  } catch (error) {
    console.error('获取模具信息失败:', error);
    MessagePlugin.error('获取模具信息失败');
  } finally {
    infoLoading.value = false;
  }
};

const fetchTrendData = async () => {
  chartLoading.value = true;
  try {
    const res = await getStockTrend(trendDays.value);
    // 等待 v-loading 状态切换完成后，再渲染图表
    await nextTick();
    setTimeout(() => {
      renderChart(res);
    }, 100);
  } catch (error) {
    console.error('获取趋势数据失败:', error);
  } finally {
    chartLoading.value = false;
  }
};

// 渲染折线图
const renderChart = (data: any[]) => {
  const chartDom = document.getElementById('trendChart');
  if (!chartDom) {
    console.error('trendChart DOM element not found');
    return;
  }
  
  if (chartDom.offsetWidth === 0 || chartDom.offsetHeight === 0) {
    console.error('chartDom has zero dimensions, cannot init chart');
    return;
  }
  
  if (!trendChart) {
    trendChart = echarts.init(chartDom);
  }

  const divideNum = trendDays.value;
  const timeArray: string[] = [];
  const inArray: number[] = [];  // 入库数据
  const outArray: number[] = []; // 出库数据

  for (let i = 0; i < divideNum; i++) {
    const daysAgo = divideNum - 1 - i;
    const date = dayjs().tz('Asia/Shanghai').subtract(daysAgo, 'day');
    const formatted = date.format('YYYY-MM-DD');
    timeArray.push(formatted);
    inArray.push(0);  // 入库数据默认为0
    outArray.push(0); // 出库数据默认为0
  }

  // 将实际的入库和出库数据映射到对应日期
  if (data && data.length > 0) {
    const inDataMap = new Map<string, number>();
    const outDataMap = new Map<string, number>();
    
    data.forEach((item: any) => {
      // 后端返回格式: { date: "2026-04-15", stock_in: 100, stock_out: 50 }
      const dateStr = item.date || '';
      if (dateStr) {
        inDataMap.set(dateStr, item.stock_in || 0);
        outDataMap.set(dateStr, item.stock_out || 0);
      }
    });

    timeArray.forEach((dateKey, index) => {
      inArray[index] = inDataMap.get(dateKey) || 0;
      outArray[index] = outDataMap.get(dateKey) || 0;
    });
  }

  // 使用 dashboard 的模板配置
  const chartColors = store.chartColors;
  const option: any = {
    color: ['#0052d9', '#0594fa'],
    tooltip: {
      trigger: 'axis',
    },
    grid: {
      left: '50px',
      right: '20px',
      top: '20px',
      bottom: '50px',
      containLabel: true,
    },
    legend: {
      left: 'center',
      bottom: '0',
      orient: 'horizontal',
      data: ['入库', '出库'],
      textStyle: {
        fontSize: 12,
        color: chartColors.placeholderColor,
      },
    },
    xAxis: {
      type: 'category',
      data: timeArray.map(d => d.substring(5)), // X轴显示 MM-DD
      boundaryGap: false,
      axisLabel: {
        color: chartColors.placeholderColor,
      },
      axisLine: {
        lineStyle: {
          color: chartColors.borderColor,
          width: 1,
        },
      },
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        color: chartColors.placeholderColor,
      },
      splitLine: {
        lineStyle: {
          color: chartColors.borderColor,
        },
      },
    },
    series: [
      {
        name: '入库',
        data: inArray,
        type: 'line',
        smooth: true,
        showSymbol: true,
        symbol: 'circle',
        symbolSize: 6,
        itemStyle: {
          color: '#0052d9',
        },
        lineStyle: {
          width: 2,
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(0, 82, 217, 0.3)' },
            { offset: 1, color: 'rgba(0, 82, 217, 0.05)' },
          ]),
        },
      },
      {
        name: '出库',
        data: outArray,
        type: 'line',
        smooth: true,
        showSymbol: true,
        symbol: 'circle',
        symbolSize: 6,
        itemStyle: {
          color: '#722ed1', // 紫色
        },
        lineStyle: {
          width: 2,
        },
      },
    ],
  };

  trendChart.setOption(option, true);
};

// 监听状态筛选变化，自动刷新数据
watch(statusFilters, () => {
  pagination.value.current = 1; // 重置到第一页
  fetchStockRecords();
}, { deep: true });

const handleSortChange = () => {
  pagination.value.current = 1; // 重置到第一页
  fetchStockRecords();
};

const fetchStockRecords = async () => {
  tableLoading.value = true;
  try {
    const params: any = {
      page: pagination.value.current,
      page_size: pagination.value.pageSize,
      sort_order: sortOrder.value,
    };

    if (statusFilters.value && statusFilters.value.length > 0) {
      params.status_filters = statusFilters.value.join(',');
    }

    // 从 URL 参数中获取 product_id
    const productId = route.query.product_id;
    if (productId) {
      params.product_id = productId;
    }

    let res;
    if (isExternalMold.value) {
      // 外来模具
      const externalMoldCode = route.params.externalMoldCode as string;
      const decodedCode = decodeURIComponent(externalMoldCode);
      res = await getStockRecordsByExternalMoldCode(decodedCode, params);
    } else {
      // 自有模具
      const moldId = route.params.id as string;
      res = await getStockRecordsByMoldId(Number(moldId), params);
    }
    
    // 缓存汇总数据，供后续使用
    tableData.value = res.list || [];
    pagination.value.total = res.total || 0;
    
    if (res.summary && !cachedSummary.value) {
      cachedSummary.value = res.summary;
      updateMoldInfoFromSummary(res.summary);
    }
  } catch (error) {
    console.error('获取出入库记录失败:', error);
    MessagePlugin.error('获取出入库记录失败');
  } finally {
    tableLoading.value = false;
  }
};

// 从 summary 更新模具信息显示
const updateMoldInfoFromSummary = (summary: any) => {
  if (!summary) return;
  
  moldInfo.value = {
    ...moldInfo.value,
    total_stock_in: summary.total_stock_in || 0,
    total_stock_out: summary.total_stock_out || 0,
    current_stock: summary.current_stock || 0,
  };
};

// 趋势时间范围变化
const handleTrendChange = () => {
  fetchTrendData();
};

// 分页变化
const handlePageChange = (pageInfo: any) => {
  pagination.value.current = pageInfo.current;
  pagination.value.pageSize = pageInfo.pageSize;
  fetchStockRecords();
};

// 查看订单详情
const handleViewOrder = (row: any) => {
  if (row.order_id) {
    router.push(`/production/order/detail/${row.order_id}`);
  }
};

const handleBack = () => {
  router.push('/production/stock');
};

// 窗口resize时更新图表大小
const handleResize = () => {
  trendChart?.resize();
};

// 编辑相关状态
const editDialogVisible = ref(false);
const editLoading = ref(false);
const editFormRef = ref();
const editFormData = ref<any>({
  id: null,
  type: '',
  stock_code: '',
  order: null,
  product_name: '',
  mold: null,
  quantity: 0,
  stock_in_date: '',
  batch_no: '',
  remark: '',
  out_type: '',
  out_reason: '',
  stock_out_date: '',
  warehouse: null,
  customer_name: '',
  delivery_date: '',
  delivery_method: '',
  express_company: '',
  express_no: '',
});

// 编辑对话框标题
const editDialogTitle = computed(() => {
  return editFormData.value.type === 'in' ? '编辑入库记录' : '编辑出库记录';
});

// 盘库相关状态
const stockOutDialogVisible = ref(false);
const stockOutLoading = ref(false);
const stockOutFormRef = ref();
const stockOutFormData = ref({
  quantity: undefined as number | undefined,
  outReason: '',
  remark: '',
});

// 是否盘亏（实际数量 < 系统库存）
const isCheckLoss = computed(() => {
  const actual = stockOutFormData.value.quantity;
  const current = moldInfo.value.current_stock;
  if (actual === undefined || actual === null || current === undefined || current === null)
    return false;
  return Number(actual) < Number(current);
});

// 打开盘库对话框
const handleStockOut = async () => {
  stockOutFormData.value = {
    quantity: undefined,
    outReason: '',
    remark: '',
  };

  stockOutDialogVisible.value = true;
};

// 关闭盘库对话框
const handleStockOutClose = () => {
  stockOutFormRef.value?.clearValidate();
  stockOutFormData.value = {
    quantity: undefined,
    outReason: '',
    remark: '',
  };
};

// 确认盘库
const handleStockOutConfirm = async () => {
  const errors: string[] = [];

  const actualQuantity = stockOutFormData.value.quantity;
  if (actualQuantity === undefined || actualQuantity === null) {
    errors.push('请输入盘点数量');
  } else if (Number(actualQuantity) < 0) {
    errors.push('盘点数量不能为负数');
  }

  // 盘亏时必须选择出库原因
  if (isCheckLoss.value && !stockOutFormData.value.outReason) {
    errors.push('请选择盘亏原因');
  }

  if (errors.length > 0) {
    MessagePlugin.warning(errors.join('；'));
    return;
  }

  const diff = Number(actualQuantity) - (Number(moldInfo.value.current_stock) || 0);

  // 二次确认
  const bodyHtml = `<div style="line-height: 2;">
    <div>模具编号：${moldInfo.value.mold_code || '-'}</div>
    <div>产品名称：${moldInfo.value.mold_name || '-'}</div>
    <div>系统库存：${moldInfo.value.current_stock}</div>
    <div>盘点数量：${actualQuantity}</div>
    ${diff > 0 ? `<div style="color: var(--td-success-color); font-weight: 500;">盘盈：+${diff}</div>` : ''}
    ${diff < 0 ? `<div style="color: var(--td-error-color); font-weight: 500;">盘亏：${diff}</div><div>盘亏原因：${outReasonDict.value.find((item: any) => item.dict_value === stockOutFormData.value.outReason)?.dict_label || '-'}</div>` : ''}
    ${diff === 0 ? '<div>库存数量无变化</div>' : ''}
    <div style="margin-top: 8px; font-weight: 500;">确认提交盘点吗？</div>
  </div>`;

  const confirmDialog = DialogPlugin.confirm({
    header: '确认盘点',
    body: () => h('div', { innerHTML: bodyHtml }),
    onConfirm: async () => {
      try {
        const params: any = {
          warehouse_id: 1,
          actual_quantity: Number(actualQuantity),
          remark: stockOutFormData.value.remark,
        };

        if (isExternalMold.value) {
          params.external_mold_code = route.params.externalMoldCode as string;
        } else {
          params.mold_id = Number(route.params.id);
        }

        const productId = route.query.product_id;
        if (productId) {
          params.product_id = productId;
        }

        // 盘亏时传递出库原因
        if (isCheckLoss.value) {
          params.out_reason = stockOutFormData.value.outReason;
        }

        await checkStock(params);

        MessagePlugin.success('盘点成功');
        stockOutDialogVisible.value = false;
        confirmDialog.hide();

        // 刷新数据
        fetchStockRecords();
        fetchMoldInfo();
      } catch (error: any) {
        console.error('盘点失败:', error);
        const errorMsg = error.response?.data?.message || error.message || '盘点失败';
        MessagePlugin.error(errorMsg);
      }
    },
    onCancel: () => {
      confirmDialog.hide();
    },
  });
};

// 打开编辑对话框
const handleEdit = (row: any) => {
  editFormData.value = {
    id: row.id,
    type: row.type,
    stock_code: row.stock_code,
    order: row.order,
    product_name: row.product_name,
    mold: row.mold,
    quantity: row.quantity,
    stock_in_date: row.stock_in_date || '',
    batch_no: row.batch_no || '',
    remark: row.remark || '',
    out_type: row.out_type || '',
    out_reason: row.out_reason || '',
    stock_out_date: row.stock_out_date || '',
    warehouse: row.warehouse,
    customer_name: row.customer_name || '',
    delivery_date: row.delivery_date || '',
    delivery_method: row.delivery_method || '',
    express_company: row.express_company || '',
    express_no: row.express_no || '',
  };
  editDialogVisible.value = true;
};

// 关闭编辑对话框
const handleEditClose = () => {
  editFormRef.value?.clearValidate();
  editFormData.value = {
    id: null,
    type: '',
    stock_code: '',
    order: null,
    product_name: '',
    mold: null,
    quantity: 0,
    stock_in_date: '',
    batch_no: '',
    remark: '',
    out_type: '',
    out_reason: '',
    stock_out_date: '',
    warehouse: null,
    customer_name: '',
    delivery_date: '',
    delivery_method: '',
    express_company: '',
    express_no: '',
  };
};

// 确认编辑
const handleEditConfirm = async () => {
  editLoading.value = true;
  try {
    // 格式化日期时间为 YYYY-MM-DD HH:mm:ss（使用本地时间，避免时区问题）
    const formatDateTimeForSubmit = (dateValue: any): string | undefined => {
      if (!dateValue) return undefined;
      
      // 如果已经是正确格式的字符串，直接返回
      if (typeof dateValue === 'string' && /^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(dateValue)) {
        return dateValue;
      }
      
      // 如果是 Date 对象，转换为 ISO 字符串后再格式化
      if (dateValue instanceof Date) {
        if (isNaN(dateValue.getTime())) return undefined;
        return formatDateTime(dateValue.toISOString());
      }
      
      // 其他情况尝试使用 formatDateTime
      const result = formatDateTime(String(dateValue));
      return result === '-' ? undefined : result;
    };

    if (editFormData.value.type === 'in') {
      // 编辑入库记录（批次号不允许修改）
      await updateStockIn({
        id: editFormData.value.id,
        quantity: editFormData.value.quantity,
        stock_in_date: formatDateTimeForSubmit(editFormData.value.stock_in_date),
        remark: editFormData.value.remark,
      });
    } else {
      // 编辑出库记录
      await updateStockOut(editFormData.value.id, {
        stock_out_date: formatDateTimeForSubmit(editFormData.value.stock_out_date),
        customer_name: editFormData.value.customer_name || undefined,
        delivery_date: formatDateTimeForSubmit(editFormData.value.delivery_date),
        delivery_method: editFormData.value.delivery_method || undefined,
        express_company: editFormData.value.express_company || undefined,
        express_no: editFormData.value.express_no || undefined,
        remark: editFormData.value.remark,
      });
    }
    MessagePlugin.success('编辑成功');
    editDialogVisible.value = false;
    // 刷新表格数据
    fetchStockRecords();
    // 刷新模具信息（汇总数据可能变化）
    fetchMoldInfo();
  } catch (error) {
    console.error('编辑记录失败:', error);
    MessagePlugin.error('编辑失败');
  } finally {
    editLoading.value = false;
  }
};

// 打开出库对话框
/**
 * 跳转到发货页面（与列表页完全一致）
 */
const handleNavigateToStockOut = () => {
  // 直接使用与列表页相同的逻辑
  const params = new URLSearchParams();
  
  if (isExternalMold.value) {
    params.append('external_mold_code', route.params.externalMoldCode as string);
  } else {
    params.append('mold_id', route.params.id as string);
    if (moldInfo.value.mold_code) {
      params.append('mold_code', moldInfo.value.mold_code);
    }
  }
  
  const productId = route.query.product_id;
  if (productId !== null && productId !== undefined) {
    params.append('product_id', productId as string);
  }
  
  // 添加模具名称（用于显示）
  if (moldInfo.value.mold_name) {
    params.append('mold_name', moldInfo.value.mold_name);
  }
  
  if (moldInfo.value.mold_name) {
    params.append('product_name', moldInfo.value.mold_name);
  }
  
  // 添加客户名称（用于自动填充）
  if (moldInfo.value.customer_name) {
    params.append('customer_name', moldInfo.value.customer_name);
  }
  
  // 添加当前库存（用于自动填充模式下显示库存提示）
  if (moldInfo.value.current_stock !== undefined) {
    params.append('current_stock', String(moldInfo.value.current_stock));
  }
  
  params.append('from', 'stock-detail');
  
  router.push(`/production/stock-out/create?${params.toString()}`);
};

// 监听路由参数变化，同组件不同 id 之间切换时重新加载数据
watch(
  () => [route.params.id, route.params.externalMoldCode],
  (newParams) => {
    // 必须仍是当前页面，避免导航到其他路由时误触发
    if (route.name !== 'ProductionStockDetail') return;
    // 离开详情页（返回列表）时不触发请求
    const [moldId, extCode] = newParams;
    if (!moldId && !extCode) return;

    cachedSummary.value = null;
    pagination.value.current = 1;
    tableData.value = [];
    fetchMoldInfo();
    fetchTrendData();
    fetchStockRecords();
  },
);

onMounted(() => {
  // 字典已通过 useDictRaw 自动加载
  fetchMoldInfo();
  fetchTrendData();
  fetchStockRecords();

  window.addEventListener('resize', handleResize);
});

onUnmounted(() => {
  window.removeEventListener('resize', handleResize);
  trendChart?.dispose();
});
</script>

<style lang="less" scoped>
.stock-detail-container {
  padding: 16px;
  background-color: var(--td-bg-color-page, #f5f5f5);
  min-height: 100vh;
}

.back-button-container {
  margin-bottom: 16px;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

@media (max-width: 1200px) {
  .stats-row {
    grid-template-columns: 1fr;
  }
}

.card-wrapper {
  display: flex;
}

.card-wrapper {
  display: flex;
  min-height: 410px;
}

.mold-info-card {
  width: 100%;
  height: 100%;
  
  :deep(.t-card__body) {
    padding: 0;
    height: 100%;
    box-sizing: border-box;
  }
}

.chart-card {
  width: 100%;
  height: 100%;
  
  :deep(.t-card__body) {
    height: 100%;
  }
}

.mold-info-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xl);
  height: 100%;
  box-sizing: border-box;
  width: 100%;
}

.header-row {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--td-border-level-1-color, #e7e7e7);
  width: 100%;
}

.mold-name {
  font-size: 20px;
  font-weight: 600;
  color: var(--td-text-color-primary, rgba(0, 0, 0, 0.9));
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 100%;
  min-height: 28px;
  padding: 2px 0;
}

.mold-code {
  display: flex;
  align-items: center;
  gap: 6px;
}

.details-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px 16px;
  width: 100%;
}

.detail-item {
  display: flex;
  align-items: center;
  min-height: 24px;
}

.detail-label {
  color: var(--td-text-color-secondary, rgba(0, 0, 0, 0.6));
  font-size: 16px;
  white-space: nowrap;
  flex-shrink: 0;
  line-height: 1.5;
}

.detail-value {
  color: var(--td-text-color-primary, rgba(0, 0, 0, 0.9));
  font-size: 16px;
  font-weight: 500;
  margin-left: 4px;
  line-height: 1.5;
  
  &.highlight {
    font-size: 15px;
    font-weight: 600;
    color: var(--td-brand-color, #0052d9);
  }
}

.mold-images {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding-top: 12px;
  border-top: 1px solid var(--td-border-level-1-color, #e7e7e7);
  width: 100%;
}

.chart-card {
  height: 100%;
  min-height: 410px;
}

.table-card {
  margin-bottom: 16px;
}

.stock-code-text {
  font-weight: 500;
  color: var(--td-text-color-primary);
}

.order-code-link {
  font-size: 15px;
}

.text-success {
  color: var(--td-success-color, #00a870);
  font-weight: 600;
}

.text-warning {
  color: var(--td-warning-color, #ed7b2f);
  font-weight: 600;
}
</style>

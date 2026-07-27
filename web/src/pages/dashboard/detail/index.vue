<template>
  <div class="dashboard-panel-detail">
    <t-card :title="t('pages.dashboardDetail.topPanel.title')" class="dashboard-detail-card" :bordered="false">
      <t-row :gutter="[16, 16]">
        <t-col v-for="(item, index) in PANE_LIST_DATA" :key="index" :xs="6" :xl="3">
          <t-card class="dashboard-list-card" :description="item.title">
            <div class="dashboard-list-card__number">{{ item.number }}</div>
            <div class="dashboard-list-card__text">
              <div class="dashboard-list-card__text-left">
                {{ t('pages.dashboardDetail.topPanel.quarter') }}
                <trend class="icon" :type="item.upTrend ? 'up' : 'down'" :describe="item.upTrend || item.downTrend" />
              </div>
              <t-icon name="chevron-right" />
            </div>
          </t-card>
        </t-col>
      </t-row>
    </t-card>
    <t-card class="dashboard-detail-card row-margin" :title="t('pages.dashboardDetail.procurement.title')" :bordered="false">
      <div id="lineContainer" style="width: 100%; height: 416px" />
    </t-card>
    <t-card
      class="dashboard-detail-card row-margin"
      :title="t('pages.dashboardDetail.delivery.title')"
      :bordered="false"
    >
      <div id="scatterContainer" style="width: 100%; height: 434px" />
    </t-card>
  </div>
</template>
<script setup lang="ts">
import { useWindowSize } from '@vueuse/core';
import { LineChart, ScatterChart } from 'echarts/charts';
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components';
import * as echarts from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { computed, nextTick, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue';

import Trend from '@/components/trend/index.vue';
import { t } from '@/locales';
import { useSettingStore } from '@/store';
import { changeChartsTheme } from '@/utils/color';

import { getDashboardDetailOverview, getDashboardDetailProductionTrend, getDeliveryPunctuality } from '@/api/dashboard';
import type { DashboardDetailOverview, ProductionTrend } from '@/api/dashboard';
import { getPaneListData } from './constants';
import { getFolderLineDataSet, getScatterDataSet } from './index';

defineOptions({
  name: 'DashboardDetail',
});
echarts.use([GridComponent, LegendComponent, TooltipComponent, LineChart, ScatterChart, CanvasRenderer]);

const store = useSettingStore();
const chartColors = computed(() => store.chartColors);

// 统计报表页概览数据（4张卡片）
const detailOverview = ref<DashboardDetailOverview>({
  completedProductionOrders: 0,
  completedProductionOrdersGrowth: 0,
  inProgressProductionOrders: 0,
  inProgressProductionOrdersGrowth: 0,
  inProgressMoldWorkOrders: 0,
  inProgressMoldWorkOrdersGrowth: 0,
  completedMoldWorkOrders: 0,
  completedMoldWorkOrdersGrowth: 0,
});

const PANE_LIST_DATA = computed(() => getPaneListData(detailOverview.value));

const loadDetailOverview = async () => {
  try {
    const res = await getDashboardDetailOverview();
    if (res) {
      detailOverview.value = res;
    }
  } catch (error) {
    console.error('加载统计报表概览数据失败:', error);
  }
};

// lineChart logic
let lineContainer: HTMLElement;
let lineChart: echarts.ECharts;
const trendData = ref<ProductionTrend | null>(null);

const renderLineChart = () => {
  lineContainer = document.getElementById('lineContainer');
  if (!lineChart) {
    lineChart = echarts.init(lineContainer);
  }
  lineChart.setOption(getFolderLineDataSet({ ...chartColors.value, detailProductionTrend: trendData.value }));
};

const loadTrendData = async () => {
  try {
    const res = await getDashboardDetailProductionTrend();
    if (res) {
      trendData.value = res;
      renderLineChart();
    }
  } catch (error) {
    console.error('加载生产执行趋势数据失败:', error);
    renderLineChart();
  }
};

// scatterChart logic — 交付准时率
let scatterContainer: HTMLElement;
let scatterChart: echarts.ECharts;
const deliveryData = ref<ProductionTrend | null>(null);

const renderScatterChart = () => {
  scatterContainer = document.getElementById('scatterContainer');
  if (!scatterChart) {
    scatterChart = echarts.init(scatterContainer);
  }
  scatterChart.setOption(getScatterDataSet({ ...chartColors.value, deliveryPunctuality: deliveryData.value }));
};

const loadDeliveryData = async () => {
  try {
    const res = await getDeliveryPunctuality();
    if (res) {
      deliveryData.value = res;
      renderScatterChart();
    }
  } catch (error) {
    console.error('加载交付准时率数据失败:', error);
  }
};

// chartSize update
const updateContainer = () => {
  lineChart?.resize({
    width: lineContainer.clientWidth,
    height: lineContainer.clientHeight,
  });
  scatterChart?.resize({
    width: scatterContainer.clientWidth,
    height: scatterContainer.clientHeight,
  });
};

onMounted(() => {
  loadDetailOverview();
  loadTrendData();
  loadDeliveryData();
  nextTick(() => {
    updateContainer();
  });
});

const { width, height } = useWindowSize();
watch([width, height], () => {
  updateContainer();
});

onDeactivated(() => {
  storeModeWatch();
  storeBrandThemeWatch();
});

onBeforeUnmount(() => {
  lineChart?.dispose();
  scatterChart?.dispose();
});

const storeModeWatch = watch(
  () => store.mode,
  () => {
    renderLineChart();
    renderScatterChart();
  },
);

const storeBrandThemeWatch = watch(
  () => store.brandTheme,
  () => {
    changeChartsTheme([lineChart, scatterChart]);
  },
);
</script>
<style lang="less" scoped>
.row-margin {
  margin-top: 16px;
}

/* 统一增加8px */
.dashboard-detail-card {
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__header) {
    padding: 0;
  }

  :deep(.t-card__title) {
    font: var(--td-font-title-large);
    font-weight: 400;
  }

  :deep(.t-card__body) {
    padding: 0;
    margin-top: var(--td-comp-margin-xxl);
  }

  :deep(.t-card__actions) {
    display: flex;
    align-items: center;
  }
}

.dashboard-list-card {
  display: flex;
  flex-direction: column;
  flex: 1;
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xl);

  :deep(.t-card__description) {
    margin: 0;
  }

  :deep(.t-card__body) {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    margin-top: var(--td-comp-margin-s);
  }

  &.dark {
    &:hover {
      background: var(--td-gray-color-14);
      cursor: pointer;
    }
  }

  &.light {
    &:hover {
      background: var(--td-gray-color-14);
      cursor: pointer;
    }
  }

  &__number {
    font-size: var(--td-font-size-headline-medium);
    line-height: var(--td-font-size-headline-medium);
    color: var(--td-text-color-primary);
    margin-bottom: var(--td-comp-margin-xxl);
  }

  &__text {
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
    font: var(--td-font-body-medium);
    color: var(--td-text-color-placeholder);
    text-align: left;

    .t-icon {
      font-size: var(--td-comp-size-xxxs);
    }

    &-left {
      display: flex;

      .icon {
        margin: 0 var(--td-comp-margin-s);
      }
    }
  }
}
</style>

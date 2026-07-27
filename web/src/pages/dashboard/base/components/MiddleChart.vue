<template>
  <t-row :gutter="[16, 16]" class="row-container">
    <t-col :xs="12" :xl="9">
      <t-card
        :title="t('pages.dashboardBase.topPanel.analysis.title')"
        :subtitle="currentMonth"
        class="dashboard-chart-card"
        :bordered="false"
      >
        <template #actions>
          <div class="dashboard-chart-title-container">
            <t-date-range-picker
              class="card-date-picker-container"
              theme="primary"
              mode="date"
              :default-value="LAST_7_DAYS"
              @change="(value: any) => onCurrencyChange(value as string[])"
            />
          </div>
        </template>
        <div
          id="monitorContainer"
          class="dashboard-chart-container"
          :style="{ width: '100%', height: `${resizeTime * 326}px` }"
        />
      </t-card>
    </t-col>
    <t-col :xs="12" :xl="3">
      <t-card
        :title="t('pages.dashboardBase.topPanel.analysis.channels')"
        :subtitle="currentMonth"
        class="dashboard-chart-card"
        :bordered="false"
      >
        <div
          id="countContainer"
          class="dashboard-chart-container"
          :style="{ width: `${resizeTime * 326}px`, height: `${resizeTime * 326}px`, margin: '0 auto' }"
        />
      </t-card>
    </t-col>
  </t-row>
</template>
<script setup lang="ts">
import { useWindowSize } from '@vueuse/core';
import { LineChart, PieChart } from 'echarts/charts';
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components';
import * as echarts from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { computed, nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue';

import { t } from '@/locales';
import { useSettingStore } from '@/store';
import { changeChartsTheme } from '@/utils/color';
import { LAST_7_DAYS } from '@/utils/date';
import { getProductionTrend, getMoldStatus } from '@/api/dashboard';

import { getLineChartDataSet, getPieChartDataSet } from '../index';

echarts.use([TooltipComponent, LegendComponent, PieChart, GridComponent, LineChart, CanvasRenderer]);

const getThisMonth = (checkedValues?: string[]) => {
  let date: Date;
  if (!checkedValues || checkedValues.length === 0) {
    date = new Date();
    return `${date.getFullYear()}-${date.getMonth() + 1}`;
  }
  date = new Date(checkedValues[0]);
  const date2 = new Date(checkedValues[1]);

  const startMonth = date.getMonth() + 1 > 9 ? date.getMonth() + 1 : `0${date.getMonth() + 1}`;
  const endMonth = date2.getMonth() + 1 > 9 ? date2.getMonth() + 1 : `0${date2.getMonth() + 1}`;
  return `${date.getFullYear()}-${startMonth}  至  ${date2.getFullYear()}-${endMonth}`;
};

const store = useSettingStore();
const resizeTime = ref(1);

const chartColors = computed(() => store.chartColors);

// monitorChart
let monitorContainer: HTMLElement;
let monitorChart: echarts.ECharts;
const renderMonitorChart = (data?: any) => {
  if (!monitorContainer) {
    monitorContainer = document.getElementById('monitorContainer');
  }
  if (!monitorChart) {
    monitorChart = echarts.init(monitorContainer);
  }
  const hasTwoLines = data?.datasets?.length >= 2 && data?.datasets?.[1]?.data?.length > 0;
  monitorChart.setOption(getLineChartDataSet({ 
    ...chartColors.value,
    dateTime: data?.categories || [],
    lineData: data?.datasets?.[0]?.data || [],
    lastLineData: hasTwoLines ? (data?.datasets?.[1]?.data || []) : [],
    lineName: data?.datasets?.[0]?.name || t('pages.dashboardBase.chart.thisMonth'),
    lastLineName: hasTwoLines ? (data?.datasets?.[1]?.name || t('pages.dashboardBase.chart.lastMonth')) : '',
  }));
};

// countChart (模具状态)
let countContainer: HTMLElement;
let countChart: echarts.ECharts;
const renderCountChart = (data?: any) => {
  if (!countContainer) {
    countContainer = document.getElementById('countContainer');
  }
  if (!countChart) {
    countChart = echarts.init(countContainer);
  }
  
  const pieData = [
    { value: data?.inUse || 0, name: '使用中' },
    { value: data?.idle || 0, name: '闲置' },
    { value: data?.maintenance || 0, name: '维护中' },
    { value: data?.scrap || 0, name: '报废' },
    { value: data?.notStarted || 0, name: '未开始' },
  ].filter(item => item.value > 0);
  
  // 随机选择一个索引，使中心文字每次刷新显示不同状态
  const randomIndex = pieData.length > 0 ? Math.floor(Math.random() * pieData.length) : 0;
  
  countChart.setOption(getPieChartDataSet({ 
    ...chartColors.value,
    pieData: pieData.length > 0 ? pieData : [{ value: 0, name: '无数据' }],
    highlightIndex: randomIndex,
  }));
};

const renderCharts = () => {
  renderMonitorChart();
  renderCountChart();
};

const loadChartData = async (startDate?: string, endDate?: string) => {
  try {
    // 并行加载生产趋势和模具状态
    const period = 'week';
    const [trendRes, statusRes] = await Promise.all([
      getProductionTrend(period, startDate, endDate),
      getMoldStatus(),
    ]);
    
    if (trendRes) {
      renderMonitorChart(trendRes);
    }
    if (statusRes) {
      renderCountChart(statusRes);
    }
  } catch (error) {
    console.error('加载图表数据失败:', error);
  }
};

// chartSize update
const updateContainer = () => {
  if (document.documentElement.clientWidth >= 1400 && document.documentElement.clientWidth < 1920) {
    resizeTime.value = Number((document.documentElement.clientWidth / 2080).toFixed(2));
  } else if (document.documentElement.clientWidth < 1080) {
    resizeTime.value = Number((document.documentElement.clientWidth / 1080).toFixed(2));
  } else {
    resizeTime.value = 1;
  }

  if (monitorChart) {
    monitorChart.resize({
      width: monitorContainer.clientWidth,
      height: resizeTime.value * 326,
    });
  }
  
  if (countChart) {
    countChart.resize({
      width: resizeTime.value * 326,
      height: resizeTime.value * 326,
    });
  }
};

onMounted(() => {
  loadChartData(LAST_7_DAYS[0], LAST_7_DAYS[1]);
  nextTick(() => {
    updateContainer();
  });
});

// ✨ KeepAlive 激活时刷新数据（默认开启 KeepAlive 缓存，否则从其他页面返回仪表盘不会重新查询）
onActivated(() => {
  loadChartData(LAST_7_DAYS[0], LAST_7_DAYS[1]);
});

const { width, height } = useWindowSize();
watch([width, height], () => {
  updateContainer();
});

onDeactivated(() => {
  storeModeWatch();
  storeBrandThemeWatch();
  storeSidebarCompactWatch();
});

onBeforeUnmount(() => {
  monitorChart?.dispose();
  countChart?.dispose();
});

const currentMonth = ref(getThisMonth());

const storeBrandThemeWatch = watch(
  () => store.brandTheme,
  () => {
    changeChartsTheme([monitorChart, countChart]);
  },
);

const storeSidebarCompactWatch = watch(
  () => store.isSidebarCompact,
  () => {
    if (store.isSidebarCompact) {
      nextTick(() => {
        updateContainer();
      });
    } else {
      setTimeout(() => {
        updateContainer();
      }, 180);
    }
  },
);

const storeModeWatch = watch(
  () => store.mode,
  () => {
    [monitorChart, countChart].forEach((item) => {
      item.dispose();
    });

    renderCharts();
  },
);

const onCurrencyChange = async (checkedValues: string[]) => {
  currentMonth.value = getThisMonth(checkedValues);
  if (checkedValues && checkedValues.length >= 2) {
    await loadChartData(checkedValues[0], checkedValues[1]);
  }
};
</script>
<style lang="less" scoped>
.dashboard-chart-card {
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__header) {
    padding: 0;
  }

  :deep(.t-card__body) {
    padding: 0;
    margin-top: var(--td-comp-margin-xxl);
  }

  :deep(.t-card__title) {
    font: var(--td-font-title-large);
    font-weight: 400;
  }
}
</style>

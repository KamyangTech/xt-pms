<template>
  <t-card :bordered="false">
    <t-row>
      <t-col :xs="12" :xl="9">
        <t-card
          :bordered="false"
          :title="t('pages.dashboardBase.outputOverview.title')"
          class="dashboard-overview-card overview-panel"
        >
          <template #actions>
            <t-date-range-picker
              class="card-date-picker-container"
              theme="primary"
              mode="date"
              :default-value="LAST_7_DAYS"
              @change="(value: any) => onDateRangeChange(value as string[])"
            />
          </template>
          <div id="stokeContainer" style="width: 100%; height: 351px" class="dashboard-chart-container"></div>
        </t-card>
      </t-col>
      <t-col :xs="12" :xl="3">
        <t-card :bordered="false" class="dashboard-overview-card export-panel">
          <t-row>
            <t-col :xs="6" :xl="12">
              <t-card
                :bordered="false"
                :subtitle="t('pages.dashboardBase.outputOverview.receivableTotal')"
                class="inner-card"
                style="cursor: pointer;"
                @click="goToReceivable"
              >
                <div class="inner-card__content">
                  <div class="inner-card__content-title">{{ formatMoney(thisMonthReceivable) }}</div>
                  <div class="inner-card__content-footer">
                    {{ t('pages.dashboardBase.outputOverview.vsLastMonth') }}
                    <trend
                      class="trend-tag"
                      :type="thisMonthReceivableGrowth >= 0 ? 'up' : 'down'"
                      :is-reverse-color="false"
                      :describe="Math.abs(thisMonthReceivableGrowth) + '%'"
                    />
                  </div>
                </div>
              </t-card>
            </t-col>
            <t-col :xs="6" :xl="12">
              <t-card
                :bordered="false"
                :subtitle="t('pages.dashboardBase.outputOverview.payableTotal')"
                class="inner-card"
                style="cursor: pointer;"
                @click="goToPayable"
              >
                <div class="inner-card__content">
                  <div class="inner-card__content-title">{{ formatMoney(thisMonthPayable) }}</div>
                  <div class="inner-card__content-footer">
                    {{ t('pages.dashboardBase.outputOverview.vsLastMonth') }}
                    <trend
                      class="trend-tag"
                      :type="thisMonthPayableGrowth >= 0 ? 'up' : 'down'"
                      :is-reverse-color="false"
                      :describe="Math.abs(thisMonthPayableGrowth) + '%'"
                    />
                  </div>
                </div>
              </t-card>
            </t-col>
          </t-row>
        </t-card>
      </t-col>
    </t-row>
  </t-card>
</template>
<script setup lang="ts">
import { useWindowSize } from '@vueuse/core';
import { BarChart, LineChart } from 'echarts/charts';
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components';
import * as echarts from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { computed, nextTick, onActivated, onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';

// 导入样式
import Trend from '@/components/trend/index.vue';
import { t } from '@/locales';
import { useSettingStore } from '@/store';
import { changeChartsTheme } from '@/utils/color';
import { LAST_7_DAYS } from '@/utils/date';

import { getFinanceOverview } from '@/api/dashboard';
import { constructFinanceChartDataset } from '../index';

defineOptions({
  name: 'DashboardBase',
});

echarts.use([TooltipComponent, LegendComponent, GridComponent, BarChart, LineChart, CanvasRenderer]);

const store = useSettingStore();
const resizeTime = ref(1);

const chartColors = computed(() => store.chartColors);
const router = useRouter();

// 跳转到财务页面
const goToReceivable = () => router.push('/finance/receivable');
const goToPayable = () => router.push('/finance/payable');

const thisMonthReceivable = ref(0);
const thisMonthReceivableGrowth = ref(0);
const thisMonthPayable = ref(0);
const thisMonthPayableGrowth = ref(0);

// 格式化金额
const formatMoney = (value: number) => {
  if (value === 0) return '¥0.00';
  return '¥' + value.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

const loadFinanceData = async (startDate: string, endDate: string) => {
  try {
    const res = await getFinanceOverview(startDate, endDate);
    const data = res;
    thisMonthReceivable.value = data.thisMonthReceivable;
    thisMonthReceivableGrowth.value = data.thisMonthReceivableGrowth;
    thisMonthPayable.value = data.thisMonthPayable;
    thisMonthPayableGrowth.value = data.thisMonthPayableGrowth;

    if (stokeChart) {
      stokeChart.setOption(constructFinanceChartDataset({
        chartData: data.chartData,
        ...chartColors.value,
      }));
    }
  } catch (error) {
    console.error('获取财务概览数据失败:', error);
  }
};

// stokeCharts
let stokeContainer: HTMLElement;
let stokeChart: echarts.ECharts;
const renderStokeChart = () => {
  if (!stokeContainer) {
    stokeContainer = document.getElementById('stokeContainer');
  }
  stokeChart = echarts.init(stokeContainer);
  stokeChart.setOption(constructFinanceChartDataset({ chartData: [], ...chartColors.value }));
};

const renderCharts = () => {
  renderStokeChart();
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

  stokeChart.resize({
    width: stokeContainer.clientWidth,
    height: stokeContainer.clientHeight,
  });
};

onMounted(() => {
  renderCharts();
  loadFinanceData(LAST_7_DAYS[0], LAST_7_DAYS[1]);
  nextTick(() => {
    updateContainer();
  });
});

// ✨ KeepAlive 激活时刷新数据
onActivated(() => {
  loadFinanceData(LAST_7_DAYS[0], LAST_7_DAYS[1]);
});

const { width, height } = useWindowSize();
watch([width, height], () => {
  updateContainer();
});

watch(
  () => store.brandTheme,
  () => {
    changeChartsTheme([stokeChart]);
  },
);

watch(
  () => store.mode,
  () => {
    [stokeChart].forEach((item) => {
      item.dispose();
    });

    renderCharts();
    loadFinanceData(LAST_7_DAYS[0], LAST_7_DAYS[1]);
  },
);

const onDateRangeChange = (checkedValues: string[]) => {
  loadFinanceData(checkedValues[0], checkedValues[1]);
};
</script>
<style lang="less" scoped>
:deep(.t-card__body) {
  padding: var(--td-comp-paddingTB-xxl) var(--td-comp-paddingLR-xxl);
}

.dashboard-overview-card {
  :deep(.t-card__header) {
    padding: 0;
  }

  :deep(.t-card__title) {
    font: var(--td-font-title-large);
    font-weight: 400;
  }

  :deep(.t-card__body) {
    margin-top: var(--td-comp-margin-xxl);
    padding: 0;
  }

  &.overview-panel {
    border-right: none;
  }

  &.export-panel {
    border-left: none;
    margin-left: calc(var(--td-comp-margin-xxxl) + var(--td-comp-margin-xxxl));
  }
}

.inner-card {
  margin-top: var(--td-comp-margin-s);
  margin-bottom: var(--td-comp-margin-xxxxl);

  :deep(.t-card__header) {
    padding-bottom: 0;
  }

  :deep(.t-card__body) {
    margin-top: var(--td-comp-margin-s);
  }

  &__content {
    &-title {
      font-size: var(--td-font-size-headline-medium);
      line-height: var(--td-line-height-headline-medium);
    }

    &-footer {
      display: flex;
      align-items: center;
      color: var(--td-text-color-placeholder);
      margin-top: var(--td-comp-margin-xxl);

      .trend-tag {
        margin-left: var(--td-comp-margin-s);
      }
    }
  }
}
</style>

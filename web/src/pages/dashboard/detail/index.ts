import type { TChartColor } from '@/config/color';
import { t } from '@/locales';
import { getChartListColor } from '@/utils/color';
import { getRandomArray } from '@/utils/charts';
import type { ProductionTrend } from '@/api/dashboard';

/**
 * 交付准时率散点图数据
 */
export function getScatterDataSet({
  placeholderColor,
  borderColor,
  deliveryPunctuality,
}: { deliveryPunctuality?: ProductionTrend | null } & TChartColor) {
  let dateArray: string[] = [];
  let earlyData: number[] = [];
  let onTimeData: number[] = [];
  let delayedData: number[] = [];

  if (deliveryPunctuality && deliveryPunctuality.categories) {
    dateArray = deliveryPunctuality.categories;
    for (const ds of deliveryPunctuality.datasets) {
      if (ds.name === 'early') earlyData = ds.data;
      if (ds.name === 'onTime') onTimeData = ds.data;
      if (ds.name === 'delayed') delayedData = ds.data;
    }
  }

  if (dateArray.length === 0) {
    for (let i = 1; i <= 12; i++) {
      dateArray.push(`${i}月`);
    }
  }

  return {
    color: getChartListColor(),
    xAxis: {
      data: dateArray,
      axisLabel: {
        color: placeholderColor,
      },
      splitLine: { show: false },
      axisLine: {
        lineStyle: {
          color: borderColor,
          width: 1,
        },
      },
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        color: placeholderColor,
      },
      nameTextStyle: {
        padding: [0, 0, 0, 60],
      },
      axisTick: {
        show: false,
        axisLine: {
          show: false,
        },
      },
      axisLine: {
        show: false,
      },
      splitLine: {
        lineStyle: {
          color: borderColor,
        },
      },
    },
    tooltip: {
      trigger: 'item',
    },
    grid: {
      top: '5px',
      left: '25px',
      right: '5px',
      bottom: '60px',
    },
    legend: {
      left: 'center',
      bottom: '0',
      orient: 'horizontal',
      data: [
        t('pages.dashboardDetail.delivery.early'),
        t('pages.dashboardDetail.delivery.onTime'),
        t('pages.dashboardDetail.delivery.delayed'),
      ],
      itemHeight: 8,
      itemWidth: 8,
      textStyle: {
        fontSize: 12,
        color: placeholderColor,
      },
    },
    series: [
      {
        name: t('pages.dashboardDetail.delivery.early'),
        symbolSize: 10,
        data: earlyData,
        type: 'scatter',
      },
      {
        name: t('pages.dashboardDetail.delivery.onTime'),
        symbolSize: 10,
        data: onTimeData,
        type: 'scatter',
      },
      {
        name: t('pages.dashboardDetail.delivery.delayed'),
        symbolSize: 10,
        data: delayedData,
        type: 'scatter',
      },
    ],
  };
}

/** 生产执行趋势折线图数据（近12个月月度趋势） */
export function getFolderLineDataSet({
  placeholderColor,
  borderColor,
  detailProductionTrend,
}: { detailProductionTrend?: ProductionTrend | null } & TChartColor) {
  let dateArray: string[] = [];
  let seriesData: Record<string, number[]> = {};

  if (detailProductionTrend && detailProductionTrend.categories) {
    dateArray = detailProductionTrend.categories;
    const datasets = detailProductionTrend.datasets;
    for (const ds of datasets) {
      seriesData[ds.name] = ds.data;
    }
  }

  if (dateArray.length === 0) {
    for (let i = 1; i <= 12; i++) {
      dateArray.push(`${i}月`);
    }
  }

  const legendNames = [
    'completedProductionOrders',
    'inProgressProductionOrders',
    'completedMoldWorkOrders',
    'inProgressMoldWorkOrders',
  ];

  return {
    color: getChartListColor(),
    grid: {
      top: '5%',
      right: '10px',
      left: '30px',
      bottom: '60px',
    },
    legend: {
      left: 'center',
      bottom: '0',
      orient: 'horizontal',
      data: legendNames.map((name) => t(`pages.dashboardDetail.procurement.goods.${name}`)),
      textStyle: {
        fontSize: 12,
        color: placeholderColor,
      },
    },
    xAxis: {
      type: 'category',
      data: dateArray,
      boundaryGap: false,
      axisLabel: {
        color: placeholderColor,
      },
      axisLine: {
        lineStyle: {
          color: borderColor,
          width: 1,
        },
      },
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        color: placeholderColor,
      },
      splitLine: {
        lineStyle: {
          color: borderColor,
        },
      },
    },
    tooltip: {
      trigger: 'item',
    },
    series: legendNames.map((name) => ({
      showSymbol: true,
      symbol: 'circle',
      symbolSize: 8,
      name: t(`pages.dashboardDetail.procurement.goods.${name}`),
      data: seriesData[name] ?? (dateArray.length > 0 ? dateArray.map(() => getRandomArray()) : []),
      type: 'line' as const,
      itemStyle: {
        borderColor,
        borderWidth: 1,
      },
    })),
  };
}

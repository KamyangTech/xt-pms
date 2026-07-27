import dayjs from 'dayjs';
import type { EChartsOption } from 'echarts';

import type { TChartColor } from '@/config/color';
import { t } from '@/locales/index';
import { getRandomArray } from '@/utils/charts';
import { getChartListColor } from '@/utils/color';

/** 首页 dashboard 折线图 */
export function constructInitDashboardDataset(type: string) {
  const dateArray: Array<string> = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];

  const datasetAxis = {
    xAxis: {
      type: 'category',
      show: false,
      data: dateArray,
    },
    yAxis: {
      show: false,
      type: 'value',
    },
    grid: {
      top: 0,
      left: 0,
      right: 0,
      bottom: 0,
    },
  };

  if (type === 'line') {
    const lineDataset = {
      ...datasetAxis,
      color: ['#fff'],
      series: [
        {
          data: [0, 0, 0, 0, 0, 0, 0],
          type,
          showSymbol: true,
          symbol: 'circle',
          symbolSize: 0,
          markPoint: {
            data: [
              { type: 'max', name: '最大值' },
              { type: 'min', name: '最小值' },
            ],
          },
          lineStyle: {
            width: 2,
          },
        },
      ],
    };
    return lineDataset;
  }
  const barDataset = {
    ...datasetAxis,
    color: getChartListColor(),
    series: [
      {
        data: [0, 0, 0, 0, 0, 0, 0],
        type,
        barWidth: 9,
      },
    ],
  };
  return barDataset;
}

/** 柱状图数据源 */
export function constructInitDataset({
  dateTime = [],
  placeholderColor,
  borderColor,
}: { dateTime: Array<string> } & TChartColor) {
  const divideNum = 10;
  const timeArray = [];
  const inArray = [];
  const outArray = [];
  for (let i = 0; i < divideNum; i++) {
    if (dateTime.length > 0) {
      const start = dayjs(dateTime[0]);
      const end = dayjs(dateTime[1]);
      if (start.isValid() && end.isValid()) {
        const diff = end.valueOf() - start.valueOf();
        const point = start.valueOf() + (diff / divideNum) * i;
        timeArray.push(dayjs(point).format('YYYY-MM-DD'));
      } else {
        timeArray.push(
          dayjs()
            .subtract(divideNum - i, 'day')
            .format('YYYY-MM-DD'),
        );
      }
    } else {
      timeArray.push(
        dayjs()
          .subtract(divideNum - i, 'day')
          .format('YYYY-MM-DD'),
      );
    }

    inArray.push(getRandomArray().toString());
    outArray.push(getRandomArray().toString());
  }

  const dataset = {
    color: getChartListColor(),
    tooltip: {
      trigger: 'item',
    },
    xAxis: {
      type: 'category',
      data: timeArray,
      axisLabel: {
        color: placeholderColor,
      },
      axisLine: {
        lineStyle: {
          color: getChartListColor()[1],
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
    grid: {
      top: '5%',
      left: '25px',
      right: 0,
      bottom: '60px',
    },
    legend: {
      icon: 'rect',
      itemWidth: 12,
      itemHeight: 4,
      itemGap: 48,
      textStyle: {
        fontSize: 12,
        color: placeholderColor,
      },
      left: 'center',
      bottom: '0',
      orient: 'horizontal',
      data: [t('pages.dashboardBase.chart.thisMonth'), t('pages.dashboardBase.chart.lastMonth')],
    },
    series: [
      {
        name: t('pages.dashboardBase.chart.thisMonth'),
        data: outArray,
        type: 'bar',
      },
      {
        name: t('pages.dashboardBase.chart.lastMonth'),
        data: inArray,
        type: 'bar',
      },
    ],
  };

  return dataset;
}

/**
 *  线性图表数据源
 *
 * @export
 * @param {Array<string>} [dateTime]
 * @returns {*} dataSet
 */
export function getLineChartDataSet({
  dateTime = [],
  placeholderColor,
  borderColor,
  lineData = [],
  lastLineData = [],
  lineName = t('pages.dashboardBase.chart.thisMonth'),
  lastLineName = t('pages.dashboardBase.chart.lastMonth'),
}: { dateTime?: Array<string>; lineData?: number[]; lastLineData?: number[]; lineName?: string; lastLineName?: string } & TChartColor) {
  // 确定数据长度：使用实际数据长度，至少 7 天
  const dataLen = Math.max(dateTime.length, lineData.length, 7);

  const timeArray: string[] = [];
  if (dateTime.length > 0) {
    // 后端传入了明确标签（如 "周一" 或 "05-09"），直接使用
    for (let i = 0; i < dataLen; i++) {
      timeArray.push(dateTime[i] || '');
    }
  } else {
    // 无标签时按日期倒退生成
    for (let i = dataLen - 1; i >= 0; i--) {
      timeArray.push(dayjs().subtract(i, 'day').format('MM-DD'));
    }
  }

  const inArray: number[] = [];
  const outArray: number[] = [];
  for (let i = 0; i < dataLen; i++) {
    inArray.push(lineData.length > 0 ? (lineData[i] ?? 0) : 0);
    outArray.push(lastLineData.length > 0 ? (lastLineData[i] ?? 0) : 0);
  }

  // 是否显示第二条线
  const hasTwoLines = lastLineData.length > 0 && lastLineName;

  const legendData = hasTwoLines ? [lineName, lastLineName] : [lineName];

  const seriesData: any[] = [
    {
      name: lineName,
      data: inArray,
      type: 'line',
      smooth: false,
      showSymbol: true,
      symbol: 'circle',
      symbolSize: 8,
      itemStyle: {
        borderColor,
        borderWidth: 1,
      },
      areaStyle: {
        opacity: 0.1,
      },
    },
  ];

  if (hasTwoLines) {
    seriesData.push({
      name: lastLineName,
      data: outArray,
      type: 'line',
      smooth: false,
      showSymbol: true,
      symbol: 'circle',
      symbolSize: 8,
      itemStyle: {
        borderColor,
        borderWidth: 1,
      },
    });
  }

  const dataSet = {
    color: getChartListColor(),
    tooltip: {
      trigger: 'item',
    },
    grid: {
      left: '0',
      right: '20px',
      top: '5px',
      bottom: '36px',
      containLabel: true,
    },
    legend: {
      left: 'center',
      bottom: '0',
      orient: 'horizontal',
      data: legendData,
      textStyle: {
        fontSize: 12,
        color: placeholderColor,
      },
    },
    xAxis: {
      type: 'category',
      data: timeArray,
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
    series: seriesData,
  };
  return dataSet;
}

/**
 * 财务概览柱状图数据源（基于后端返回的应收/应付数据）
 *
 * @export
 * @param {object} params
 * @param {Array<{date: string, receivableAmount: number, payableAmount: number}>} params.chartData - 后端返回的图表数据
 * @param {string} params.placeholderColor
 * @param {string} params.borderColor
 * @returns {*} dataSet
 */
export function constructFinanceChartDataset({
  chartData = [],
  placeholderColor,
  borderColor,
}: { chartData: Array<{ date: string; receivableAmount: number; payableAmount: number }> } & TChartColor) {
  const timeArray = chartData.map((item) => item.date);
  const receivableArray = chartData.map((item) => item.receivableAmount);
  const payableArray = chartData.map((item) => item.payableAmount);

  const dataset = {
    color: getChartListColor(),
    tooltip: {
      trigger: 'axis',
    },
    xAxis: {
      type: 'category',
      data: timeArray,
      axisLabel: {
        color: placeholderColor,
      },
      axisLine: {
        lineStyle: {
          color: getChartListColor()[1],
          width: 1,
        },
      },
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        color: placeholderColor,
        formatter: '¥{value}',
      },
      splitLine: {
        lineStyle: {
          color: borderColor,
        },
      },
    },
    grid: {
      top: '5%',
      left: '25px',
      right: 0,
      bottom: '60px',
    },
    legend: {
      icon: 'rect',
      itemWidth: 12,
      itemHeight: 4,
      itemGap: 48,
      textStyle: {
        fontSize: 12,
        color: placeholderColor,
      },
      left: 'center',
      bottom: '0',
      orient: 'horizontal',
      data: [t('pages.dashboardBase.financeOverview.receivable'), t('pages.dashboardBase.financeOverview.payable')],
    },
    series: [
      {
        name: t('pages.dashboardBase.financeOverview.receivable'),
        data: receivableArray,
        type: 'bar',
      },
      {
        name: t('pages.dashboardBase.financeOverview.payable'),
        data: payableArray,
        type: 'bar',
      },
    ],
  };

  return dataset;
}

/**
 * 获取饼图数据
 *
 * @export
 * @param {number} [radius]
 * @returns {*} dataSet
 */
export function getPieChartDataSet({
  radius = 42,
  textColor,
  placeholderColor,
  containerColor,
  pieData = [],
  highlightIndex = 0,
}: { radius?: number; pieData?: Array<{ value: number; name: string }>; highlightIndex?: number } & Record<string, string>): EChartsOption {
  // 如果没有传入数据，使用默认数据
  const data = pieData.length > 0 ? pieData : [
    { value: 0, name: '使用中' },
    { value: 0, name: '闲置' },
  ];

  // 取高亮项用于中心文字显示
  const activeItem = data[highlightIndex] || data[0] || { value: 0, name: '无数据' };

  return {
    color: getChartListColor(),
    tooltip: {
      show: false,
      trigger: 'axis',
      position: null,
    },
    grid: {
      top: '0',
      right: '0',
    },
    legend: {
      selectedMode: false,
      itemWidth: 12,
      itemHeight: 4,
      textStyle: {
        fontSize: 12,
        color: placeholderColor,
      },
      left: 'center',
      bottom: '0',
      orient: 'horizontal', // legend 横向布局。
      data: data.map(item => item.name),
    },
    series: [
      {
        name: '模具状态',
        type: 'pie',
        radius: ['48%', '60%'],
        avoidLabelOverlap: true,
        selectedMode: false,
        silent: true,
        itemStyle: {
          borderColor: 'transparent',
        },
        label: {
          show: true,
          position: 'center',
          formatter: () => `{value|${activeItem.value}}\n{name|${activeItem.name}}`,
          rich: {
            value: {
              color: textColor,
              fontSize: 28,
              fontWeight: 'normal',
              lineHeight: 46,
            },
            name: {
              color: '#909399',
              fontSize: 12,
              lineHeight: 14,
            },
          },
        },
        emphasis: {
          scale: false,
          label: {
            show: false,
          },
        },
        labelLine: {
          show: false,
        },
        data: data.map((item) => ({
          value: item.value,
          name: item.name,
        })),
      },
    ],
  };
}

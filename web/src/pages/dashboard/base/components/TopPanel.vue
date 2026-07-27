<template>
  <t-row :gutter="[16, 16]">
    <t-col v-for="(item, index) in PANE_LIST" :key="item.title" :xs="6" :xl="3">
      <t-card
        :title="t(item.title)"
        :bordered="false"
        class="dashboard-item"
        :class="{ 'dashboard-item--main-color': index === 0 }"
        @click="goToCardPage(index)"
      >
        <div class="dashboard-item-top">
          <span :style="{ fontSize: `${resizeTime * 28}px` }">{{ item.number }}</span>
        </div>
        <div class="dashboard-item-left">
          <span v-if="index === 0" :style="{ marginTop: '-24px' }">
            <layers-icon />
          </span>
          <span v-else-if="index === 1" :style="{ marginTop: '-24px' }">
            <list-icon />
          </span>
          <span v-else-if="index === 2" :style="{ marginTop: '-24px' }">
            <control-platform-icon />
          </span>
          <span v-else :style="{ marginTop: '-24px' }">
            <money-icon />
          </span>
        </div>
        <template #footer>
          <div class="dashboard-item-bottom">
            <div class="dashboard-item-block">
              {{ t('pages.dashboardBase.topPanel.cardTips') }}
              <trend
                class="dashboard-item-trend"
                :type="item.upTrend ? 'up' : 'down'"
                :is-reverse-color="index === 0"
                :describe="item.upTrend || item.downTrend"
              />
            </div>
            <t-icon name="chevron-right" />
          </div>
        </template>
      </t-card>
    </t-col>
  </t-row>
</template>
<script setup lang="ts">
import { useWindowSize } from '@vueuse/core';
import { ControlPlatformIcon, LayersIcon, ListIcon, MoneyIcon } from 'tdesign-icons-vue-next';
import { nextTick, onActivated, onMounted, ref, watch, computed } from 'vue';
import { useRouter } from 'vue-router';

// 数字格式化（保留两位小数，加千分位）
const formatNumber = (num: number): string => {
  if (num === 0) return '0';
  return num.toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
};

// 导入样式
import Trend from '@/components/trend/index.vue';
import { t } from '@/locales';
import { getDashboardOverview } from '@/api/dashboard';

defineOptions({
  name: 'DashboardBase',
});

const resizeTime = ref(1);
const router = useRouter();

// 顶部卡片点击跳转
const goToCardPage = (index: number) => {
  const routes = ['/mold/order', '/production/order/list', '/mold/work-order', '/finance/receivable'];
  router.push(routes[index]);
};

// 仪表盘数据
const overviewData = ref<any>({
  monthlyNewMoldOrders: 0,
  monthlyNewProductionOrders: 0,
  inProductionMolds: 0,
  monthlyNewReceivable: 0,
});

const PANE_LIST = computed(() => [
  {
    title: 'pages.dashboardBase.topPanel.card1',
    number: overviewData.value.monthlyNewMoldOrders || 0,
    upTrend: overviewData.value.monthlyNewMoldOrdersGrowth > 0 ? `${overviewData.value.monthlyNewMoldOrdersGrowth}%` : '',
    downTrend: overviewData.value.monthlyNewMoldOrdersGrowth < 0 ? `${Math.abs(overviewData.value.monthlyNewMoldOrdersGrowth)}%` : '',
    leftType: 'icon-file',
  },
  {
    title: 'pages.dashboardBase.topPanel.card2',
    number: overviewData.value.monthlyNewProductionOrders || 0,
    upTrend: overviewData.value.monthlyNewProductionOrdersGrowth > 0 ? `${overviewData.value.monthlyNewProductionOrdersGrowth}%` : '',
    downTrend: overviewData.value.monthlyNewProductionOrdersGrowth < 0 ? `${Math.abs(overviewData.value.monthlyNewProductionOrdersGrowth)}%` : '',
    leftType: 'icon-file',
  },
  {
    title: 'pages.dashboardBase.topPanel.card3',
    number: overviewData.value.inProductionMolds || 0,
    upTrend: overviewData.value.inProductionMoldsGrowth > 0 ? `${overviewData.value.inProductionMoldsGrowth}%` : '',
    downTrend: overviewData.value.inProductionMoldsGrowth < 0 ? `${Math.abs(overviewData.value.inProductionMoldsGrowth)}%` : '',
    leftType: 'icon-usergroup',
  },
  {
    title: 'pages.dashboardBase.topPanel.card4',
    number: overviewData.value.monthlyNewReceivable ? `¥${formatNumber(overviewData.value.monthlyNewReceivable)}` : '¥0',
    upTrend: overviewData.value.monthlyNewReceivableGrowth > 0 ? `${overviewData.value.monthlyNewReceivableGrowth}%` : '',
    downTrend: overviewData.value.monthlyNewReceivableGrowth < 0 ? `${Math.abs(overviewData.value.monthlyNewReceivableGrowth)}%` : '',
    leftType: 'icon-file-paste',
  },
]);

// chartSize update（仅用于 responsive 字体）
const updateContainer = () => {
  if (document.documentElement.clientWidth >= 1400 && document.documentElement.clientWidth < 1920) {
    resizeTime.value = Number((document.documentElement.clientWidth / 2080).toFixed(2));
  } else if (document.documentElement.clientWidth < 1080) {
    resizeTime.value = Number((document.documentElement.clientWidth / 1080).toFixed(2));
  } else {
    resizeTime.value = 1;
  }
};

const loadOverviewData = async () => {
  try {
    const res = await getDashboardOverview();
    if (res) {
      overviewData.value = res;
    }
  } catch (error) {
    console.error('加载概览数据失败:', error);
  }
};

onMounted(() => {
  loadOverviewData();
  nextTick(() => {
    updateContainer();
  });
});

// ✨ KeepAlive 激活时刷新数据
onActivated(() => {
  loadOverviewData();
});

const { width, height } = useWindowSize();
watch([width, height], () => {
  updateContainer();
});
</script>
<style lang="less" scoped>
.dashboard-item {
  padding: var(--td-comp-paddingTB-xl) var(--td-comp-paddingLR-xxl);

  :deep(.t-card__header) {
    padding: 0;
  }

  :deep(.t-card__footer) {
    padding: 0;
  }

  :deep(.t-card__title) {
    font: var(--td-font-body-medium);
    color: var(--td-text-color-secondary);
  }

  :deep(.t-card__body) {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    flex: 1;
    position: relative;
    padding: 0;
    margin-top: var(--td-comp-margin-s);
    margin-bottom: var(--td-comp-margin-xxl);
  }

  &:hover {
    cursor: pointer;
  }

  &-top {
    display: flex;
    flex-direction: row;
    align-items: flex-start;

    > span {
      display: inline-block;
      color: var(--td-text-color-primary);
      font-size: var(--td-font-size-headline-medium);
      line-height: var(--td-line-height-headline-medium);
    }
  }

  &-bottom {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;

    > .t-icon {
      cursor: pointer;
      font-size: var(--td-comp-size-xxxs);
    }
  }

  &-block {
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--td-text-color-placeholder);
  }

  &-trend {
    margin-left: var(--td-comp-margin-s);
  }

  &-left {
    position: absolute;
    top: 0;
    right: 0;

    > span {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      width: var(--td-comp-size-xxxl);
      height: var(--td-comp-size-xxxl);
      background: var(--td-brand-color-light);
      border-radius: 50%;

      .t-icon {
        font-size: 24px;
        color: var(--td-brand-color);
      }
    }
  }

  /* 针对第一个卡片需要反色处理 */
  &--main-color {
    background: var(--td-brand-color);
    color: var(--td-text-color-primary);

    :deep(.t-card__title),
    .dashboard-item-top span,
    .dashboard-item-bottom {
      color: var(--td-text-color-anti);
    }

    .dashboard-item-block {
      color: var(--td-text-color-anti);
      opacity: 0.6;
    }

    .dashboard-item-bottom {
      color: var(--td-text-color-anti);
    }
  }
}
</style>

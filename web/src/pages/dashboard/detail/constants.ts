import type { DashboardDetailOverview } from '@/api/dashboard';
import { t } from '@/locales';

export const getPaneListData = (overview: DashboardDetailOverview) => [
  {
    title: t('pages.dashboardDetail.topPanel.paneList.totalRequest'),
    number: overview.completedProductionOrders ?? 0,
    upTrend: overview.completedProductionOrdersGrowth > 0 ? `${overview.completedProductionOrdersGrowth}%` : '',
    downTrend: overview.completedProductionOrdersGrowth < 0 ? `${Math.abs(overview.completedProductionOrdersGrowth)}%` : '',
  },
  {
    title: t('pages.dashboardDetail.topPanel.paneList.inProgressProduction'),
    number: overview.inProgressProductionOrders ?? 0,
    upTrend: overview.inProgressProductionOrdersGrowth > 0 ? `${overview.inProgressProductionOrdersGrowth}%` : '',
    downTrend: overview.inProgressProductionOrdersGrowth < 0 ? `${Math.abs(overview.inProgressProductionOrdersGrowth)}%` : '',
  },
  {
    title: t('pages.dashboardDetail.topPanel.paneList.arrivalRate'),
    number: overview.completedMoldWorkOrders ?? 0,
    upTrend: overview.completedMoldWorkOrdersGrowth > 0 ? `${overview.completedMoldWorkOrdersGrowth}%` : '',
    downTrend: overview.completedMoldWorkOrdersGrowth < 0 ? `${Math.abs(overview.completedMoldWorkOrdersGrowth)}%` : '',
  },
  {
    title: t('pages.dashboardDetail.topPanel.paneList.completionRate'),
    number: overview.inProgressMoldWorkOrders ?? 0,
    upTrend: overview.inProgressMoldWorkOrdersGrowth > 0 ? `${overview.inProgressMoldWorkOrdersGrowth}%` : '',
    downTrend: overview.inProgressMoldWorkOrdersGrowth < 0 ? `${Math.abs(overview.inProgressMoldWorkOrdersGrowth)}%` : '',
  },
];



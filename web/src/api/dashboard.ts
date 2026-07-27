import { request } from '@/utils/request';

// 概览数据
export interface DashboardOverview {
  // 顶部四张卡片
  monthlyNewMoldOrders: number;          // 本月新增模具订单数量
  monthlyNewMoldOrdersGrowth: number;    // 环比增长率
  monthlyNewProductionOrders: number;    // 本月新增生产订单数量
  monthlyNewProductionOrdersGrowth: number;
  inProductionMolds: number;             // 正在生产流程的模具数量
  inProductionMoldsGrowth: number;
  monthlyNewReceivable: number;          // 本月新增应收款（营业额）
  monthlyNewReceivableGrowth: number;

  // 原有字段保持兼容
  todayOutput: number;
  todayOutputGrowth: number;
  totalMolds: number;
  moldsGrowth: number;
  stockWarning: number;
  stockWarningChange: number;
  rawMaterialStock: number;
  rawMaterialGrowth: number;
  pendingWorkOrders: number;
  workOrdersGrowth: number;
  totalCustomers: number;
  customersGrowth: number;
  totalSuppliers: number;
  suppliersGrowth: number;
  qualificationRate: number;
  qualificationGrowth: number;
}

// 财务概览数据
export interface FinanceChartItem {
  date: string;
  receivableAmount: number;
  payableAmount: number;
}

export interface FinanceOverview {
  chartData: FinanceChartItem[];
  thisMonthReceivable: number;
  thisMonthReceivableGrowth: number;
  thisMonthPayable: number;
  thisMonthPayableGrowth: number;
}

// 趋势数据集
export interface TrendDataset {
  name: string;
  data: number[];
}

// 生产趋势
export interface ProductionTrend {
  categories: string[];
  datasets: TrendDataset[];
}

// 模具状态
export interface MoldStatus {
  total: number;
  inUse: number;
  maintenance: number;
  idle: number;
  scrap: number;
}

// 排名项
export interface RankingItem {
  rank: number;
  name: string;
  change: number;
  orderCount: number;
}

// 订单排名
export interface OrderRanking {
  items: RankingItem[];
}

// 正在制造的模具项
export interface InProductionMold {
  moldId: number;
  moldCode: string;
  moldName: string;
  workOrderId: number;
  progress: number;
  workOrderStatus: string;
  lastCompletedStep: string;
}

// 正在制造的模具列表
export interface InProductionMolds {
  items: InProductionMold[];
}

// 进行中的模具订单项
export interface InProgressMoldOrderItem {
  order_id: number;
  order_code: string;
  mold_codes: string;
  customer_name: string;
  status: string;
  status_label: string;
  plan_end_date: string;
}

// 进行中的模具订单列表
export interface InProgressMoldOrders {
  items: InProgressMoldOrderItem[];
  total: number;
  page: number;
  page_size: number;
}

// 进行中的生产订单项
export interface InProgressProductionOrderItem {
  order_id: number;
  order_code: string;
  customer_name: string;
  status: string;
  status_label: string;
  plan_end_date: string;
}

// 进行中的生产订单列表
export interface InProgressProductionOrders {
  items: InProgressProductionOrderItem[];
  total: number;
}

// 统计报表页概览数据（4张卡片）
export interface DashboardDetailOverview {
  completedProductionOrders: number;
  completedProductionOrdersGrowth: number;
  inProgressProductionOrders: number;
  inProgressProductionOrdersGrowth: number;
  inProgressMoldWorkOrders: number;
  inProgressMoldWorkOrdersGrowth: number;
  completedMoldWorkOrders: number;
  completedMoldWorkOrdersGrowth: number;
}

const Api = {
  Overview: '/dashboard/overview',
  ProductionTrend: '/dashboard/production-trend',
  MoldStatus: '/dashboard/mold-status',
  OrderRanking: '/dashboard/order-ranking',
  InProductionMolds: '/dashboard/in-production-molds',
  InProgressMoldOrders: '/dashboard/in-progress-mold-orders',
  InProgressProductionOrders: '/dashboard/in-progress-production-orders',
  FinanceOverview: '/dashboard/finance-overview',
  DetailOverview: '/dashboard/detail-overview',
  DetailProductionTrend: '/dashboard/detail-production-trend',
  DeliveryPunctuality: '/dashboard/delivery-punctuality',
};

/**
 * 获取仪表盘概览数据
 */
export function getDashboardOverview() {
  return request.get<DashboardOverview>({
    url: Api.Overview,
  });
}

/**
 * 获取生产趋势数据
 * @param period 时间周期: week/month/year
 * @param startDate 开始日期 (YYYY-MM-DD)
 * @param endDate 结束日期 (YYYY-MM-DD)
 */
export function getProductionTrend(period: string = 'month', startDate?: string, endDate?: string) {
  const params: Record<string, string> = { period };
  if (startDate) params.startDate = startDate;
  if (endDate) params.endDate = endDate;
  return request.get<ProductionTrend>({
    url: Api.ProductionTrend,
    params,
  });
}

/**
 * 获取模具状态统计
 */
export function getMoldStatus() {
  return request.get<MoldStatus>({
    url: Api.MoldStatus,
  });
}

/**
 * 获取订单排名
 * @param type 排名类型: customer(客户)/mold(模具)
 * @param period 时间周期: week/month
 */
export function getOrderRanking(type: string = 'customer', period: string = 'week') {
  return request.get<OrderRanking>({
    url: Api.OrderRanking,
    params: { type, period },
  });
}

/**
 * 获取正在制造的模具列表
 */
export function getInProductionMolds() {
  return request.get<InProductionMolds>({
    url: Api.InProductionMolds,
  });
}

/**
 * 获取进行中的模具订单列表（固定8条最近到期）
 */
export function getInProgressMoldOrders() {
  return request.get<InProgressMoldOrders>({
    url: Api.InProgressMoldOrders,
  });
}

/**
 * 获取进行中的生产订单列表（固定8条最近到期）
 */
export function getInProgressProductionOrders() {
  return request.get<InProgressProductionOrders>({
    url: Api.InProgressProductionOrders,
  });
}

/**
 * 获取财务概览数据
 */
export function getFinanceOverview(startDate: string, endDate: string) {
  return request.get<FinanceOverview>({
    url: Api.FinanceOverview,
    params: { startDate, endDate },
  });
}

/**
 * 获取统计报表页概览数据（4张卡片）
 */
export function getDashboardDetailOverview() {
  return request.get<DashboardDetailOverview>({
    url: Api.DetailOverview,
  });
}

/**
 * 获取统计报表页趋势图数据（近12个月月度趋势，4条折线）
 */
export function getDashboardDetailProductionTrend() {
  return request.get<ProductionTrend>({
    url: Api.DetailProductionTrend,
  });
}

/**
 * 获取交付准时率数据（近12个月月度，3个维度：提前/准时/延迟）
 */
export function getDeliveryPunctuality() {
  return request.get<ProductionTrend>({
    url: Api.DeliveryPunctuality,
  });
}

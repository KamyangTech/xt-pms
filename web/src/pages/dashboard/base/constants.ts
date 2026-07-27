interface TendItem {
  growUp?: number;
  productName: string;
  count: number;
  date: string;
}

export const SALE_TEND_LIST: Array<TendItem> = [
  {
    growUp: 1,
    productName: '东莞精密模具有限公司',
    count: 70,
    date: '2026-04-01',
  },
  {
    growUp: -1,
    productName: '深圳塑胶制品有限公司',
    count: 58,
    date: '2026-04-01',
  },
  {
    growUp: 4,
    productName: '广州注塑科技有限公司',
    count: 45,
    date: '2026-04-01',
  },
  {
    growUp: 3,
    productName: '佛山电器制造有限公司',
    count: 38,
    date: '2026-04-01',
  },
  {
    growUp: -3,
    productName: '惠州电子配件厂',
    count: 32,
    date: '2026-04-01',
  },
  {
    growUp: -3,
    productName: '中山汽车零部件公司',
    count: 20,
    date: '2026-04-12',
  },
];

export const BUY_TEND_LIST: Array<TendItem> = [
  {
    growUp: 1,
    productName: 'MOLD-001 手机外壳模具',
    count: 301,
    date: '2026-04-01',
  },
  {
    growUp: -1,
    productName: 'MOLD-002 电器外壳模具',
    count: 201,
    date: '2026-04-01',
  },
  {
    growUp: 6,
    productName: 'MOLD-003 汽车零件模具',
    count: 181,
    date: '2026-04-11',
  },
  {
    growUp: -3,
    productName: 'MOLD-004 电子配件模具',
    count: 101,
    date: '2026-04-21',
  },
  {
    growUp: -4,
    productName: 'MOLD-005 塑料容器模具',
    count: 44,
    date: '2026-04-19',
  },
  {
    growUp: -3,
    productName: 'MOLD-006 医疗器械模具',
    count: 20,
    date: '2026-04-12',
  },
];

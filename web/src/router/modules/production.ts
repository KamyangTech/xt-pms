import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/production',
    name: 'production',
    component: Layout,
    redirect: '/production/order/list',
    meta: {
      title: { zh_CN: '生产管理', en_US: 'Production Management' },
      icon: 'tools',
      orderNo: 3,
      permission: '',
    },
    children: [
      {
        path: 'order/list',
        name: 'ProductionOrderList',
        component: () => import('@/pages/production/order-list/index.vue'),
        meta: {
          title: { zh_CN: '生产订单', en_US: 'Production Orders' },
          keepAlive: true,
          permission: 'production:order:list',
        },
      },
      {
        path: 'order/create',
        name: 'ProductionOrderCreate',
        component: () => import('@/pages/production/order-form/index.vue'),
        meta: {
          title: { zh_CN: '新建订单', en_US: 'Create Order' },
          hidden: true,
          permission: 'production:order:list',
        },
      },
      {
        path: 'order/edit/:id',
        name: 'ProductionOrderEdit',
        component: () => import('@/pages/production/order-form/index.vue'),
        meta: {
          title: { zh_CN: '编辑订单', en_US: 'Edit Order' },
          hidden: true,
          permission: 'production:order:list',
        },
      },
      {
        path: 'order/detail/:id',
        name: 'ProductionOrderDetail',
        component: () => import('@/pages/production/order-detail/index.vue'),
        meta: {
          title: { zh_CN: '订单详情', en_US: 'Order Detail' },
          hidden: true,
          permission: 'production:order:list',
        },
      },
      {
        path: 'work-report',
        name: 'WorkReport',
        component: () => import('@/pages/production/work-report/index.vue'),
        meta: {
          title: { zh_CN: '生产报工', en_US: 'Work Report' },
          keepAlive: true,
          permission: 'production:work-report:list',
        },
      },
      {
        path: 'stock',
        name: 'ProductionStock',
        component: () => import('@/pages/production/stock/index.vue'),
        meta: {
          title: { zh_CN: '库存管理', en_US: 'Stock Management' },
          keepAlive: true,
          permission: 'production:stock:list',
        },
      },
      {
        path: 'stock/detail/:id',
        name: 'ProductionStockDetail',
        component: () => import('@/pages/production/stock/detail.vue'),
        meta: {
          title: { zh_CN: '入库详情', en_US: 'Stock In Detail' },
          hidden: true,
          permission: 'production:stock:list',
        },
      },
      {
        path: 'stock/detail/ext/:externalMoldCode',
        name: 'ProductionStockExtDetail',
        component: () => import('@/pages/production/stock/detail.vue'),
        meta: {
          title: { zh_CN: '外来模具入库详情', en_US: 'External Mold Stock Detail' },
          hidden: true,
          permission: 'production:stock:list',
        },
      },
      {
        path: 'stock-out',
        name: 'StockOutList',
        component: () => import('@/pages/production/stock-out/index.vue'),
        meta: {
          title: { zh_CN: '发货管理', en_US: 'Delivery' },
          keepAlive: true,
          permission: 'production:stock-out:list',
        },
      },
      {
        path: 'stock-out/create',
        name: 'StockOutCreate',
        component: () => import('@/pages/production/stock-out/form.vue'),
        meta: {
          title: { zh_CN: '创建送货单', en_US: 'Create Delivery Note' },
          hidden: true,
          permission: 'production:stock-out:list',
        },
      },
      {
        path: 'stock-out/:id',
        name: 'StockOutDetail',
        component: () => import('@/pages/production/stock-out/[id].vue'),
        meta: {
          title: { zh_CN: '送货单详情', en_US: 'Delivery Note Detail' },
          hidden: true,
          permission: 'production:stock-out:list',
        },
      },
      {
        path: 'stock-out/edit/:id',
        name: 'StockOutEdit',
        component: () => import('@/pages/production/stock-out/form.vue'),
        meta: {
          title: { zh_CN: '编辑送货单', en_US: 'Edit Delivery Note' },
          hidden: true,
          permission: 'production:stock-out:list',
        },
      },
    ],
  },
];

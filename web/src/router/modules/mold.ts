import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/mold',
    name: 'mold',
    component: Layout,
    redirect: '/mold/index',
    meta: {
      title: { zh_CN: '模具管理', en_US: 'Mold Management' },
      icon: 'chart-3d',
      orderNo: 2,
      permission: '',
    },
    children: [
      {
        path: 'index',
        name: 'MoldIndex',
        component: () => import('@/pages/mold/index.vue'),
        meta: {
          title: { zh_CN: '模具档案', en_US: 'Mold Archive' },
          keepAlive: true,
          permission: 'mold:mold:list',
        },
      },
      {
        path: 'form',
        name: 'MoldForm',
        component: () => import('@/pages/mold/form.vue'),
        meta: {
          title: { zh_CN: '新增模具档案', en_US: 'New Mold Archive' },
          hidden: true,
          keepAlive: false,
          permission: 'mold:mold:list',
        },
      },
      {
        path: 'detail/:id',
        name: 'MoldDetail',
        component: () => import('@/pages/mold/detail.vue'),
        meta: {
          title: { zh_CN: '模具详情', en_US: 'Mold Detail' },
          hidden: true,
          permission: 'mold:mold:list',
        },
      },
      {
        path: 'order',
        name: 'MoldOrder',
        component: () => import('@/pages/mold/order/index.vue'),
        meta: {
          title: { zh_CN: '模具订单', en_US: 'Mold Order' },
          keepAlive: true,
          permission: 'mold:order:list',
        },
      },
      {
        path: 'order/form',
        name: 'MoldOrderForm',
        component: () => import('@/pages/mold/order/form.vue'),
        meta: {
          title: { zh_CN: '新增模具订单', en_US: 'New Mold Order' },
          hidden: true,
          keepAlive: false, // 禁用 KeepAlive，避免编辑数据残留到新增表单
          permission: 'mold:order:list',
        },
      },
      {
        path: 'order/detail/:id',
        name: 'MoldOrderDetail',
        component: () => import('@/pages/mold/order/detail.vue'),
        meta: {
          title: { zh_CN: '模具订单详情', en_US: 'Mold Order Detail' },
          hidden: true,
          permission: 'mold:order:list',
        },
      },
      {
        path: 'order/payment-plan',
        name: 'MoldOrderPaymentPlan',
        component: () => import('@/pages/mold/order/payment-plan.vue'),
        meta: {
          title: { zh_CN: '模具订单款项配置', en_US: 'Mold Order Payment Plan' },
          hidden: true,
          permission: 'mold:order:list',
        },
      },
      {
        path: 'work-order',
        name: 'MoldWorkOrder',
        component: () => import('@/pages/mold/work-order/index.vue'),
        meta: {
          title: { zh_CN: '模具工单', en_US: 'Mold Work Order' },
          keepAlive: true,
          permission: 'mold:work-order:list',
        },
      },
      {
        path: 'work-order/form',
        name: 'MoldWorkOrderForm',
        component: () => import('@/pages/mold/work-order/form.vue'),
        meta: {
          title: { zh_CN: '新增模具工单', en_US: 'New Mold Work Order' },
          hidden: true,
          keepAlive: false, // 禁用 KeepAlive，避免编辑数据残留到新增表单
          permission: 'mold:work-order:list',
        },
      },
      {
        path: 'usage-history',
        name: 'MoldUsageHistory',
        component: () => import('@/pages/mold/usage-history/index.vue'),
        meta: {
          title: { zh_CN: '使用历史', en_US: 'Usage History' },
          keepAlive: true,
          permission: 'mold:usage-history:list',
        },
      },
      {
        path: 'usage-history/:moldId',
        name: 'MoldUsageHistoryDetail',
        component: () => import('@/pages/mold/usage-history/detail.vue'),
        meta: {
          title: { zh_CN: '使用历史详情', en_US: 'Usage History Detail' },
          hidden: true,
          permission: 'mold:usage-history:list',
        },
      },
    ],
  },
];

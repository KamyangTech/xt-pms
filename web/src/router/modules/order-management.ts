import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/order-management',
    name: 'OrderManagement',
    component: Layout,
    redirect: '/order-management/list',
    meta: { 
      title: { zh_CN: '订单中心', en_US: 'Order Center' }, 
      icon: 'view-list', 
      orderNo: 1,
      permission: '',
    },
    children: [
      {
        path: 'list',
        name: 'OrderList',
        component: () => import('@/pages/order-management/index.vue'),
        meta: {
          title: { zh_CN: '订单中心', en_US: 'Order Center' },
          keepAlive: true,
          permission: 'order-management:list',
        },
      },
      {
        path: 'form',
        name: 'CreateMasterOrder',
        component: () => import('@/pages/order-management/form.vue'),
        meta: {
          title: { zh_CN: '综合下单', en_US: 'Combined Order' },
          hidden: true,
          keepAlive: false,
          permission: 'order-management:list',
        },
      },
      {
        path: 'detail/:id',
        name: 'MasterOrderDetail',
        component: () => import('@/pages/order-management/detail.vue'),
        meta: {
          title: { zh_CN: '订单详情', en_US: 'Order Detail' },
          hidden: true,
          permission: 'order-management:list',
        },
      },
    ],
  },
];

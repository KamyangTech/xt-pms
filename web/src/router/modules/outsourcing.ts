import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/outsourcing',
    name: 'outsourcing',
    component: Layout,
    redirect: '/outsourcing/mold',
    meta: {
      title: { zh_CN: '外发管理', en_US: 'Outsourcing Management' },
      icon: 'send',
      orderNo: 5,
      permission: '',
    },
    children: [
      {
        path: 'mold',
        name: 'OutsourcingMold',
        component: () => import('@/pages/outsourcing/mold/index.vue'),
        meta: {
          title: { zh_CN: '模具外发', en_US: 'Mold Outsourcing' },
          keepAlive: true,
          permission: 'outsourcing:mold:list',
        },
      },
      {
        path: 'production',
        name: 'OutsourcingProduction',
        component: () => import('@/pages/outsourcing/production/index.vue'),
        meta: {
          title: { zh_CN: '生产外发', en_US: 'Production Outsourcing' },
          keepAlive: true,
          permission: 'outsourcing:production:list',
        },
      },
    ],
  },
];

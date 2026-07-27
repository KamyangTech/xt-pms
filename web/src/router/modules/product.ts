import { shallowRef } from 'vue';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/product',
    name: 'product',
    component: Layout,
    redirect: '/product/index',
    meta: { title: { zh_CN: '产品管理', en_US: 'Product Management' }, icon: 'layers', orderNo: 4, permission: '' },
    children: [
      {
        path: 'index',
        name: 'ProductIndex',
        component: () => import('@/pages/basic/product/index.vue'),
        meta: {
          title: { zh_CN: '产品列表', en_US: 'Product List' },
          keepAlive: true,
          permission: 'product:list',
        },
      },
    ],
  },
];

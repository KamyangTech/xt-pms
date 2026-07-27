import { shallowRef } from 'vue';
import type { RouteRecordRaw } from 'vue-router';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/dict',
    name: 'dict',
    component: Layout,
    redirect: '/dict/mold-status',
    meta: {
      title: { zh_CN: '系统字典', en_US: 'Dictionary' },
      icon: 'list',
      orderNo: 8,
      permission: '',
    },
    children: [
      {
        path: 'mold-status',
        name: 'DictMoldStatus',
        component: () => import('@/pages/system/status-dict/index.vue'),
        meta: {
          title: { zh_CN: '模具状态字典', en_US: 'Mold Status Dictionary' },
          keepAlive: true,
          permission: 'dict:mold-status:list',
        },
      },
      {
        path: 'production-status-dict',
        name: 'DictProductionStatus',
        component: () => import('@/pages/system/production-status-dict/index.vue'),
        meta: {
          title: { zh_CN: '生产状态字典', en_US: 'Production Status Dictionary' },
          keepAlive: true,
          permission: 'dict:production-status:list',
        },
      },
      {
        path: 'master-order-dict',
        name: 'DictMasterOrder',
        component: () => import('@/pages/system/master-order-dict/index.vue'),
        meta: {
          title: { zh_CN: '综合订单字典', en_US: 'Master Order Dictionary' },
          keepAlive: true,
          permission: 'dict:master-order:list',
        },
      },
      {
        path: 'outsourcing-status',
        name: 'DictOutsourcingStatus',
        component: () => import('@/pages/system/outsourcing-status-dict/index.vue'),
        meta: {
          title: { zh_CN: '外发状态字典', en_US: 'Outsourcing Status Dictionary' },
          keepAlive: true,
          permission: 'dict:outsourcing-status:list',
        },
      },
    ],
  },
] as RouteRecordRaw[];

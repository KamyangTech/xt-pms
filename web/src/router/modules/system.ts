import { shallowRef } from 'vue';
import type { RouteRecordRaw } from 'vue-router';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/system',
    name: 'system',
    component: Layout,
    redirect: '/system/organization',
    meta: {
      title: { zh_CN: '系统设置', en_US: 'System Design' },
      icon: 'setting',
      orderNo: 7,
      permission: '', // 系统管理路由组本身不需要权限，子路由各自控制
    },
    children: [
      {
        path: 'organization',
        name: 'SystemOrganization',
        component: () => import('@/pages/system/organization/index.vue'),
        meta: {
          title: { zh_CN: '组织与权限', en_US: 'Organization & Permissions' },
          keepAlive: true,
          permission: 'system:organization:list',
        },
      },
      {
        path: 'menu',
        name: 'SystemMenu',
        component: () => import('@/pages/system/menu/index.vue'),
        meta: {
          title: { zh_CN: '菜单管理', en_US: 'Menu Management' },
          keepAlive: true,
          permission: 'system:menu:list',
        },
      },
      {
        path: 'basic-data',
        name: 'SystemBasicData',
        component: () => import('@/pages/system/basic-data/index.vue'),
        meta: {
          title: { zh_CN: '基础资料', en_US: 'Basic Data' },
          keepAlive: true,
          permission: 'system:basic-data:list',
        },
      },
      {
        path: 'print-config',
        name: 'SystemPrintConfig',
        component: () => import('@/pages/system/print-config/index.vue'),
        meta: {
          title: { zh_CN: '打印配置', en_US: 'Print Configuration' },
          keepAlive: true,
          permission: 'system:print-config:list',
        },
      },
      {
        path: 'settings',
        name: 'SystemSettings',
        component: () => import('@/pages/system/settings/index.vue'),
        meta: {
          title: { zh_CN: '系统设置', en_US: 'System Settings' },
          keepAlive: true,
          permission: 'system:settings:list',
        },
      },
      {
        path: 'oper-log',
        name: 'SystemOperLog',
        component: () => import('@/pages/system/oper-log/index.vue'),
        meta: {
          title: { zh_CN: '操作日志', en_US: 'Operation Log' },
          keepAlive: true,
          permission: 'system:oper-log:list',
        },
      },
      {
        path: 'business-config',
        name: 'SystemBusinessConfig',
        component: () => import('@/pages/system/business-config/index.vue'),
        meta: {
          title: { zh_CN: '业务配置', en_US: 'Business Configuration' },
          keepAlive: true,
          permission: 'system:business-config:list',
        },
      },
      {
        path: 'monitor',
        name: 'SystemMonitor',
        component: () => import('@/pages/system/monitor/index.vue'),
        meta: {
          title: { zh_CN: '系统监控', en_US: 'System Monitor' },
          keepAlive: false,
          permission: 'system:monitor:view',
        },
      },
    ],
  },
] as RouteRecordRaw[];

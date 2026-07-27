import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/user',
    name: 'user',
    component: Layout,
    redirect: '/user/index',
    meta: { title: { zh_CN: '个人中心', en_US: 'User Center' }, icon: 'user-circle', hidden: true },
    children: [
      {
        path: 'index',
        name: 'UserIndex',
        component: () => import('@/pages/user/index.vue'),
        meta: {
          title: { zh_CN: '个人中心', en_US: 'User Center' },
          keepAlive: false, // 禁用 KeepAlive 缓存，避免响应式数据混乱
        },
      },
    ],
  },
];

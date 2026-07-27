import { NotificationIcon } from 'tdesign-icons-vue-next';
import { shallowRef } from 'vue';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/detail',
    component: Layout,
    redirect: '/detail/secondary',
    name: 'detail',
    meta: {
      title: {
        zh_CN: '通知中心',
        en_US: 'Notifications',
      },
      icon: shallowRef(NotificationIcon),
      orderNo: 0,
      hidden: true,
    },
    children: [
      {
        path: 'secondary',
        name: 'DetailSecondary',
        component: () => import('@/pages/detail/secondary/index.vue'),
        meta: {
          title: {
            zh_CN: '通知中心',
            en_US: 'Notification Center',
          },
        },
      },
    ],
  },
];

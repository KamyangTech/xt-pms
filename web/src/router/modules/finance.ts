import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/finance',
    name: 'Finance',
    component: Layout,
    redirect: '/finance/receivable',
    meta: { 
      title: { zh_CN: '财务管理', en_US: 'Financial Management' }, 
      icon: 'wallet',
      orderNo: 6,
      permission: '',
    },
    children: [
      {
        path: 'receivable',
        name: 'FinanceReceivable',
        component: () => import('@/pages/finance/receivable/index.vue'),
        meta: {
          title: { zh_CN: '应收账款', en_US: 'Accounts Receivable' },
          keepAlive: true,
          permission: 'finance:receivable:list',
        },
      },
      {
        path: 'receivable/payments/:id',
        name: 'FinanceReceivablePayments',
        component: () => import('@/pages/finance/receivable/payments/index.vue'),
        meta: {
          title: { zh_CN: '收款记录', en_US: 'Payment Records' },
          hidden: true,
          permission: 'finance:receivable:list',
        },
      },
      {
        path: 'payable',
        name: 'FinancePayable',
        component: () => import('@/pages/finance/payable/index.vue'),
        meta: {
          title: { zh_CN: '应付账款', en_US: 'Accounts Payable' },
          keepAlive: true,
          permission: 'finance:payable:list',
        },
      },
      {
        path: 'payable/payments/:id',
        name: 'FinancePayablePayments',
        component: () => import('@/pages/finance/payable/payments/index.vue'),
        meta: {
          title: { zh_CN: '付款记录', en_US: 'Payment Records' },
          hidden: true,
          permission: 'finance:payable:list',
        },
      },
      {
        path: 'expense',
        name: 'FinanceExpense',
        component: () => import('@/pages/finance/expense/index.vue'),
        meta: {
          title: { zh_CN: '费用管理', en_US: 'Expense Management' },
          keepAlive: true,
          permission: 'finance:expense:list',
        },
      },
      {
        path: 'receiving-account',
        name: 'FinanceReceivingAccount',
        component: () => import('@/pages/system/receiving-account/index.vue'),
        meta: {
          title: { zh_CN: '收款账户管理', en_US: 'Receiving Account Management' },
          keepAlive: true,
          permission: 'system:receiving-account:list',
        },
      },
    ],
  },
];

import 'nprogress/nprogress.css';

import NProgress from 'nprogress';

import router from '@/router';
import { getPermissionStore, useUserStore } from '@/store';

NProgress.configure({ showSpinner: false });

const permissionStore = getPermissionStore();

router.beforeEach(async (to, from, next) => {
  NProgress.start();

  const { whiteListRouters } = permissionStore;
  const userStore = useUserStore();

  if (userStore.token) {
    if (to.path === '/login') {
      next({ path: '/' });
      return;
    }
    try {
      // 已在错误页面或公开报工页面时跳过 API 请求，直接放行
      if (['/server-error', '/no-permission'].includes(to.path) || to.path.startsWith('/r/')) {
        next();
        return;
      }

      // 仅在首次加载时获取用户信息和权限（后续切换路由不再重复请求）
      if (!userStore.userInfo.name) {
        const sessionValid = await userStore.getUserInfo();
        // 会话恢复失败
        if (!sessionValid) {
          if (userStore.token) {
            // token 仍在但请求失败（后端不可达），跳转全屏服务器错误页面，记录来源
            next({ path: '/server-error', query: { redirect: to.fullPath } });
          } else {
            // token 不存在 → 需要重新登录
            next({
              path: '/login',
              query: { redirect: encodeURIComponent(to.fullPath) },
            });
          }
          NProgress.done();
          return;
        }
      }
      if (permissionStore.permissions.length === 0) {
        await permissionStore.initRoutes();
      }

      // 接口失败（后端不可达），跳转到全屏 500 页面
      if (permissionStore.initFailed) {
        next({ path: '/server-error' });
        return;
      }

      // 没有任何权限时跳转到 /no-permission
      if (permissionStore.permissions.length === 0) {
        next({ path: '/no-permission' });
        return;
      }

      // 单页面权限由 Content.vue 在内容区显示"无权限"提示。
      // 特殊处理：登录后通过 redirect 跳转到目标 URL 时（常见于同一浏览器窗口
      // 先登录管理员访问系统管理菜单，再切换普通账号登录），若目标路由需要权限
      // 而当前用户没有该权限，则直接跳转到首页，避免出现 403 页面。
      if (from.path === '/login') {
        const toPerm = (to.meta as Record<string, unknown>)?.permission as string | undefined;
        if (toPerm && !permissionStore.permissions.includes(toPerm)) {
          next({ path: '/dashboard/base' });
          return;
        }
      }
      next();
    } catch (error) {
      NProgress.done();
      next({
        path: '/login',
        query: { redirect: encodeURIComponent(to.fullPath) },
      });
    }
  } else {
    if (whiteListRouters.includes(to.path) || whiteListRouters.some((route: string) => route.endsWith('/') ? to.path.startsWith(route) : false)) {
      next();
    } else {
      next({
        path: '/login',
        query: { redirect: encodeURIComponent(to.fullPath) },
      });
    }
    NProgress.done();
  }
});

router.afterEach(() => {
  NProgress.done();
});

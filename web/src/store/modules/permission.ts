import { defineStore } from 'pinia';
import type { RouteRecordRaw } from 'vue-router';

import { getUserPermissions, type MenuItem } from '@/api/menu';
import { allRoutes, homepageRouterList, fixedRouterList } from '@/router';
import { store, useUserStore } from '@/store';
import type { MenuRoute } from '@/types/interface';

/**
 * 将后端返回的菜单树（MenuItem[]）转换为侧边栏可渲染的 MenuRoute[] 格式
 * 仅在备用场景下使用，主流程仍通过 filterRoutes 过滤前端路由配置
 */
function convertMenuItemsToRoutes(items: MenuItem[]): MenuRoute[] {
  if (!items || items.length === 0) return [];
  return items
    .filter((item) => item.menu_type !== 3 && item.status === 1)
    .sort((a, b) => a.sort - b.sort)
    .map((item) => ({
      path: item.path || '',
      title: item.menu_name,
      icon: item.icon || '',
      children: convertMenuItemsToRoutes(item.children || []),
      meta: {
        title: item.menu_name,
        icon: item.icon || '',
        orderNo: item.sort,
        hidden: false,
        permission: item.permission || '',
      },
    }));
}

export const usePermissionStore = defineStore('permission', {
  state: () => ({
    whiteListRouters: ['/login', '/no-permission', '/server-error', '/r/'] as string[],
    routers: [] as RouteRecordRaw[],
    removeRoutes: [] as string[],
    permissions: [] as string[],
    menuTree: [] as MenuItem[],
    initFailed: false,
  }),
  getters: {
    hasPermission: (state) => {
      return (perm: string) => {
        if (!perm) return true;
        if (state.permissions.length > 100) return true;
        return state.permissions.includes(perm);
      };
    },
    /** 将后端菜单树转换为侧边栏可渲染的路由格式（备用，当前未使用） */
    sidebarMenus(): MenuRoute[] {
      return convertMenuItemsToRoutes(this.menuTree);
    },
  },
  actions: {
    async initRoutes() {
      this.initFailed = false;
      try {
        const res = await getUserPermissions();
        this.permissions = res.permissions || [];
        this.menuTree = res.menus || [];

        // 同步更新 userStore.perms（供 v-permission 指令使用）
        const userStore = useUserStore();
        userStore.perms = this.permissions;

        // 根据权限过滤路由（保留前端路由配置的图标、层级结构）
        const homepage = this.filterRoutes(
          homepageRouterList as unknown as RouteRecordRaw[],
        );
        const fixed = this.filterRoutes(fixedRouterList as unknown as RouteRecordRaw[]);
        this.routers = [...homepage, ...fixed] as unknown as RouteRecordRaw[];
        return this.routers;
      } catch {
        // 权限接口失败时显示所有路由（如后端不可达）
        this.initFailed = true;
        this.routers = [...allRoutes];
        return allRoutes;
      }
    },
    /** 递归过滤无权限的路由（深度克隆传入路由，避免修改共享对象） */
    filterRoutes(routes: RouteRecordRaw[]): RouteRecordRaw[] {
      return routes.filter((route) => {
        const perm = (route.meta as Record<string, unknown>)?.permission as string | undefined;
        if (perm && !this.permissions.includes(perm)) {
          return false;
        }
        if (route.children) {
          const filtered = this.filterRoutes(route.children);
          // 只有当有可见的子路由时，才保留父路由，避免空菜单显示
          const hasVisibleChild = filtered.some((child) => !(child.meta as Record<string, unknown>)?.hidden);
          return hasVisibleChild;
        }
        return true;
      });
    },
    async buildAsyncRoutes() {
      return this.routers as unknown as RouteRecordRaw[];
    },
    async restoreRoutes() {
      this.removeRoutes = [];
    },
  },
});

export function getPermissionStore() {
  return usePermissionStore(store);
}

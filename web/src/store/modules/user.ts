import { defineStore } from 'pinia';

import { getSecurityConfig } from '@/api/system/settings';
import { getUserProfile, login as apiLogin } from '@/api/user';
import type { UserProfile } from '@/api/user';
import type { UserInfo } from '@/types/interface';
import { getWebSocketClient } from '@/utils/websocket';
import { usePermissionStore } from './permission';

const InitUserInfo: UserInfo = {
  name: '',
  roles: [],
};

export const useUserStore = defineStore('user', {
  state: () => ({
    token: '',
    keepLogin: false,
    userInfo: { ...InitUserInfo },
    perms: [] as string[], // 权限标识列表
  }),
  getters: {
    roles: (state) => {
      return state.userInfo?.roles;
    },
    isLoggedIn: (state) => {
      return !!state.token;
    },
    username: (state) => {
      return state.userInfo.name;
    },
  },
  actions: {
    async login(loginData: Record<string, unknown>) {
      try {
        const keepLogin = !!loginData.keep_login;
        const res = await apiLogin({
          username: (loginData.username || loginData.account) as string,
          password: loginData.password as string,
          captcha_id: loginData.captcha_id as string,
          captcha_val: loginData.captcha_val as string,
          keep_login: keepLogin,
        });

        this.token = res.token;
        this.keepLogin = keepLogin;
        this.userInfo = {
          name: res.userInfo.real_name || res.userInfo.username,
          roles: res.userInfo.roles || ['all'],
        };

        localStorage.setItem('token', res.token);
        localStorage.setItem('userInfo', JSON.stringify(res.userInfo));
        localStorage.setItem('keepLogin', JSON.stringify(keepLogin));

        // 登录后权限获取交由路由守卫统一处理（避免重复请求）
        const wsClient = getWebSocketClient();
        if (wsClient) {
          wsClient.connect();
          // 同步 WebSocket Secure 标志
          try {
            const secCfg = await getSecurityConfig();
            wsClient.setSecure(!!secCfg.websocket_secure);
          } catch {
            // 安全配置加载失败，使用默认值
          }
        }

        return res;
      } catch (error: any) {
        console.error('登录失败:', error);
        throw error;
      }
    },
    async fetchUserInfo() {
      try {
        const res = await getUserProfile();
        this.userInfo = {
          name: res.real_name || res.username,
          roles: this.userInfo.roles,
        };
        const existingInfo = localStorage.getItem('userInfo');
        let userInfoObj: any = {};
        if (existingInfo) {
          userInfoObj = JSON.parse(existingInfo);
        }
        const mergedInfo = {
          ...userInfoObj,
          ...res,
        };
        localStorage.setItem('userInfo', JSON.stringify(mergedInfo));
        return res;
      } catch (error) {
        console.error('获取用户信息失败:', error);
        return null;
      }
    },
    /** 尝试恢复会话。返回 true 表示会话有效，false 表示需要重新登录 */
    async getUserInfo(): Promise<boolean> {
      // keepLogin 仅控制后端 token 过期时间（24h / 系统配置），
      // 不影响前端是否持久化 token，所以总是尝试从 localStorage 恢复
      const savedToken = localStorage.getItem('token');
      const savedKeepLogin = localStorage.getItem('keepLogin');

      if (!savedToken) {
        this.token = '';
        return false;
      }

      this.token = savedToken;
      this.keepLogin = savedKeepLogin === 'true';

      const profile = await this.fetchUserInfo();
      if (!profile) {
        return false;
      }

      // 同步 WebSocket Secure 标志
      try {
        const secCfg = await getSecurityConfig();
        const wsClient = getWebSocketClient();
        if (wsClient) {
          wsClient.setSecure(!!secCfg.websocket_secure);
        }
      } catch {
        // 安全配置加载失败，使用默认值
      }

      return true;
    },
    async logout() {
      this.token = '';
      this.keepLogin = false;
      this.userInfo = { ...InitUserInfo };
      this.perms = [];
      localStorage.removeItem('token');
      localStorage.removeItem('userInfo');
      localStorage.removeItem('keepLogin');

      // 同时重置权限 store，确保重新登录时路由守卫会重新 initRoutes()
      const permissionStore = usePermissionStore();
      permissionStore.$reset();

      // 清除 WebSocket 的 pms_token Cookie（同时清除 Secure 和 Non-Secure 两种版本）
      document.cookie = 'pms_token=; path=/; max-age=0; SameSite=Strict; Secure';
      document.cookie = 'pms_token=; path=/; max-age=0; SameSite=Strict';

      const wsClient = getWebSocketClient();
      if (wsClient) {
        wsClient.disconnect();
      }
    },
  },
  persist: {
    key: 'user',
    paths: ['token', 'keepLogin'],
  },
});

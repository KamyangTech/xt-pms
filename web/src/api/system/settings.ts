import { request } from '@/utils/request';

/** 手动触发孤文件清理 */
export function triggerCleanupOrphanFiles() {
  return request.post({ url: '/system/cleanup-orphan-files' });
}

/** 获取孤文件数量 */
export function getOrphanFileCount(): Promise<{ count: number }> {
  return request.get({ url: '/system/cleanup-orphan-files/count' });
}

/** 获取清理配置 */
export function getCleanupConfig(): Promise<{ retention_days: number }> {
  return request.get({ url: '/system/cleanup-config' });
}

/** 更新清理配置 */
export function updateCleanupConfig(data: { retention_days: number }) {
  return request.put({ url: '/system/cleanup-config', data });
}

// ========== 系统信息配置 ==========

export interface SystemInfo {
  id: number;
  system_name: string;
  system_version: string;
  system_description: string;
  tech_stack: string;
  favicon_url: string;
  create_time: string;
  update_time: string;
}

/** 获取系统信息 */
export function getSystemInfo(): Promise<SystemInfo> {
  return request.get({ url: '/system/system-info' }, { skip500Redirect: true });
}

/** 更新系统信息 */
export function updateSystemInfo(data: {
  system_name: string;
}) {
  return request.put({ url: '/system/system-info', data });
}

// ========== 登录与安全配置 ==========

export interface SecurityConfig {
  id: number;
  token_expire: number;          // Token有效期(秒)
  max_login_attempts: number;    // 最大登录失败次数
  lock_duration: number;         // 账号锁定时长(秒)
  password_min_len: number;      // 密码最小长度
  password_require_upper: boolean;   // 是否要求大写字母
  password_require_number: boolean;  // 是否要求数字
  password_require_special: boolean; // 是否要求特殊字符
  cookie_secure: boolean;        // Cookie Secure: true=仅HTTPS发送
  websocket_secure: boolean;     // WebSocket Secure: true=仅WSS发送
  create_time: string;
  update_time: string;
}

/** 获取安全配置 */
export function getSecurityConfig(): Promise<SecurityConfig> {
  return request.get({ url: '/system/security-config' }, { skip403Redirect: true });
}

/** 更新安全配置 */
export function updateSecurityConfig(data: Partial<SecurityConfig>) {
  return request.put({ url: '/system/security-config', data });
}

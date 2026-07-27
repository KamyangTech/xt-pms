/**
 * 字典更新通知模块（版本号轮询模式）
 * 替代原有的 WebSocket 方案，通过定期轮询版本号检测字典变更
 */

import { MessagePlugin } from 'tdesign-vue-next';
import { getDictVersion } from '@/api/dict';
import { dictUtils } from './dict';

// 轮询间隔（5 分钟），字典数据属于低频变更数据
const POLL_INTERVAL = 5 * 60 * 1000;

// 当前缓存的版本号，用于对比检测变更
let cachedVersion = 0;

// 轮询定时器
let pollTimer: ReturnType<typeof setInterval> | null = null;

// 是否已初始化
let initialized = false;

/**
 * 轮询检查字典版本
 * 如果版本号发生变化，则清除所有字典缓存
 */
async function checkDictVersion() {
  try {
    const res = await getDictVersion();
    const newVersion: number = res.version ?? res.data?.version ?? 0;

    if (cachedVersion > 0 && newVersion !== cachedVersion) {
      // 字典版本变更，清除所有字典缓存
      dictUtils.clearCache();

      // 显示提示消息
      MessagePlugin.info('字典数据已更新，将使用最新数据', 3000);
    }

    cachedVersion = newVersion;
  } catch (error) {
    console.error('[DictNotification] 检查字典版本失败:', error);
  }
}

/**
 * 初始化字典更新通知（版本号轮询模式）
 * 在应用启动时调用一次
 */
export function initDictNotification() {
  if (initialized) {
    console.warn('[DictNotification] 已初始化，请勿重复调用');
    return;
  }

  initialized = true;

  // 立即执行一次检查
  checkDictVersion();

  pollTimer = setInterval(checkDictVersion, POLL_INTERVAL);
}

/**
 * 销毁字典更新通知
 * 在应用卸载时调用
 */
export function destroyDictNotification() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
  initialized = false;
}

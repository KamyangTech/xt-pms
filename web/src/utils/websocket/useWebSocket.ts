/**
 * WebSocket Vue Composable
 * 提供在 Vue 组件中使用 WebSocket 的便捷方法
 */

import { ref, onMounted, onUnmounted } from 'vue';
import { 
  initWebSocket, 
  getWebSocketClient, 
  WSStatus,
  type WSOptions 
} from './client';

/**
 * 使用 WebSocket 的 Composable
 * 
 * @example
 * ```typescript
 * const { status, sendMessage, onMessage } = useWebSocket({
 *   url: '/api/ws',
 *   token: () => localStorage.getItem('token') || '',
 * });
 * 
 * // 监听特定消息
 * onMessage('dict-updated', (data) => {
 *   console.log('字典已更新:', data);
 * });
 * 
 * // 监听连接状态
 * watch(status, (newStatus) => {
 *   console.log('连接状态:', newStatus);
 * });
 * ```
 */
export function useWebSocket(options?: Partial<WSOptions>) {
  const status = ref<WSStatus>(WSStatus.DISCONNECTED);
  const isConnected = ref(false);

  let wsClient = getWebSocketClient();
  let statusUnsubscribe: (() => void) | null = null;
  const messageUnsubscribes: (() => void)[] = [];

  const init = () => {
    if (!wsClient) {
      const defaultOptions: WSOptions = {
        url: options?.url || `/api/ws`,
        token: options?.token || (() => localStorage.getItem('token') || ''),
        autoReconnect: options?.autoReconnect ?? true,
        reconnectInterval: options?.reconnectInterval ?? 3000,
        maxReconnectAttempts: options?.maxReconnectAttempts ?? 10,
        heartbeatInterval: options?.heartbeatInterval ?? 30000,
        secure: options?.secure ?? false,
      };

      wsClient = initWebSocket(defaultOptions);
    }

    // 监听状态变化
    statusUnsubscribe = wsClient.onStatusChange((newStatus) => {
      status.value = newStatus;
      isConnected.value = newStatus === WSStatus.CONNECTED;
    });

    // 连接
    wsClient.connect();
  };

  const sendMessage = (message: any): boolean => {
    if (!wsClient) {
      console.error('[useWebSocket] WebSocket 未初始化');
      return false;
    }
    return wsClient.send(message);
  };

  // 监听特定类型的消息
  const onMessage = (messageType: string, handler: (data: any) => void) => {
    if (!wsClient) {
      console.error('[useWebSocket] WebSocket 未初始化');
      return;
    }

    const unsubscribe = wsClient.on(messageType, handler);
    messageUnsubscribes.push(unsubscribe);
    
    return unsubscribe;
  };

  // 断开连接
  const disconnect = () => {
    if (wsClient) {
      wsClient.disconnect();
    }
  };

  // 组件挂载时初始化
  onMounted(() => {
    init();
  });

  // 组件卸载时清理
  onUnmounted(() => {
    // 取消所有消息监听
    messageUnsubscribes.forEach(unsubscribe => unsubscribe());
    messageUnsubscribes.length = 0;

    // 取消状态监听
    if (statusUnsubscribe) {
      statusUnsubscribe();
      statusUnsubscribe = null;
    }

    // 注意：不在这里断开连接，因为其他组件可能还在使用
    // 如果需要完全断开，可以手动调用 disconnect()
  });

  return {
    status,
    isConnected,
    sendMessage,
    onMessage,
    disconnect,
    init,
  };
}

/**
 * 全局初始化 WebSocket（在 main.ts 中调用）
 */
export function initGlobalWebSocket(options?: Partial<WSOptions>) {
  const defaultOptions: WSOptions = {
    url: options?.url || `/api/ws`,
    token: options?.token || (() => localStorage.getItem('token') || ''),
    autoReconnect: options?.autoReconnect ?? true,
    reconnectInterval: options?.reconnectInterval ?? 3000,
    maxReconnectAttempts: options?.maxReconnectAttempts ?? 10,
    heartbeatInterval: options?.heartbeatInterval ?? 30000,
    secure: options?.secure ?? false,
  };

  return initWebSocket(defaultOptions);
}

// 导出状态枚举
export { WSStatus };

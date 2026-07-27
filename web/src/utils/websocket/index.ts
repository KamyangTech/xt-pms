/**
 * WebSocket 模块统一导出
 */

export { initWebSocket, getWebSocketClient, WSStatus } from './client';
export { useWebSocket, initGlobalWebSocket } from './useWebSocket';
export type { WSOptions, WSMessage } from './client';

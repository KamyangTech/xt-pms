/**
 * WebSocket 客户端核心
 * 提供通用的 WebSocket 连接、重连、心跳等功能
 */

// WebSocket 连接状态
export enum WSStatus {
  CONNECTING = 'connecting',
  CONNECTED = 'connected',
  DISCONNECTED = 'disconnected',
  RECONNECTING = 'reconnecting',
}

// WebSocket 消息类型
export interface WSMessage {
  type: string;
  data: any;
  timestamp: number;
}

// WebSocket 配置选项
export interface WSOptions {
  url: string;                    // WebSocket 地址
  token?: string | (() => string); // 认证 token
  autoReconnect?: boolean;        // 是否自动重连
  reconnectInterval?: number;     // 重连间隔（毫秒）
  maxReconnectAttempts?: number;  // 最大重连次数
  heartbeatInterval?: number;     // 心跳间隔（毫秒）
  secure?: boolean;               // 是否启用 Secure（WSS 时设置 Cookie Secure 标志）
}

// WebSocket 客户端类
class WebSocketClient {
  private ws: WebSocket | null = null;
  private options: Required<WSOptions>;
  private status: WSStatus = WSStatus.DISCONNECTED;
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setTimeout> | null = null;
  private messageHandlers: Map<string, Set<(data: any) => void>> = new Map();
  private statusListeners: Set<(status: WSStatus) => void> = new Set();

  constructor(options: WSOptions) {
    this.options = {
      url: options.url,
      token: options.token || '',
      autoReconnect: options.autoReconnect ?? true,
      reconnectInterval: options.reconnectInterval ?? 3000,
      maxReconnectAttempts: options.maxReconnectAttempts ?? 10,
      heartbeatInterval: options.heartbeatInterval ?? 30000,
      secure: options.secure ?? false,
    };
  }

  /**
   * 连接 WebSocket
   */
  connect(): void {
    if (this.ws && this.status === WSStatus.CONNECTED) {
      console.warn('[WebSocket] 已经连接，无需重复连接');
      return;
    }

    const token = typeof this.options.token === 'function' 
      ? this.options.token() 
      : this.options.token;

    // 如果 token 为空，不建立连接（等待登录后再连接）
    if (!token) {
      this.status = WSStatus.DISCONNECTED;
      this.notifyStatusChange();
      return;
    }

    this.status = WSStatus.CONNECTING;
    this.notifyStatusChange();

    try {
      // 将 token 写入 Cookie（后端优先从 Cookie 读取，避免 token 暴露在 URL 中）
      const secureFlag = this.options.secure ? '; Secure' : '';
      document.cookie = `pms_token=${encodeURIComponent(token)}; path=/; max-age=3600; SameSite=Strict${secureFlag}`;

      // WebSocket URL 中不再携带 token，由 Cookie 负责认证
      const wsUrl = this.options.url;

      this.ws = new WebSocket(wsUrl);

      // 连接成功
      this.ws.onopen = () => {
        this.status = WSStatus.CONNECTED;
        this.reconnectAttempts = 0;
        this.notifyStatusChange();

        // token 已通过 Cookie (pms_token) 传递，无需再发送认证消息

        this.startHeartbeat();
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('[WebSocket] 消息解析失败:', error);
        }
      };

      // 连接关闭
      this.ws.onclose = () => {
        this.status = WSStatus.DISCONNECTED;
        this.notifyStatusChange();
        this.stopHeartbeat();

        // 自动重连
        if (this.options.autoReconnect) {
          this.scheduleReconnect();
        }
      };

      // 连接错误
      this.ws.onerror = (error) => {
        console.error('[WebSocket] 连接错误:', error);
        this.status = WSStatus.DISCONNECTED;
        this.notifyStatusChange();
      };
    } catch (error) {
      console.error('[WebSocket] 创建连接失败:', error);
      this.status = WSStatus.DISCONNECTED;
      this.notifyStatusChange();

      if (this.options.autoReconnect) {
        this.scheduleReconnect();
      }
    }
  }

  /**
   * 断开连接
   */
  disconnect(): void {
    // 禁用自动重连
    this.options.autoReconnect = false;
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.status = WSStatus.DISCONNECTED;
    this.notifyStatusChange();
  }

  /**
   * 设置 Secure 标志（在安全配置加载后调用）
   */
  setSecure(secure: boolean): void {
    this.options.secure = secure;
    // 如果已连接，用新设置刷新一下 Cookie 的 Secure 标志
    if (this.status === WSStatus.CONNECTED) {
      const token = typeof this.options.token === 'function'
        ? this.options.token()
        : this.options.token;
      if (token) {
        const secureFlag = secure ? '; Secure' : '';
        document.cookie = `pms_token=${encodeURIComponent(token)}; path=/; max-age=3600; SameSite=Strict${secureFlag}`;
      }
    }
  }

  /**
   * 发送消息
   */
  send(message: any): boolean {
    if (!this.ws || this.status !== WSStatus.CONNECTED) {
      console.warn('[WebSocket] 连接未建立，无法发送消息');
      return false;
    }

    try {
      this.ws.send(JSON.stringify(message));
      return true;
    } catch (error) {
      console.error('[WebSocket] 发送消息失败:', error);
      return false;
    }
  }

  /**
   * 注册消息处理器
   */
  on(messageType: string, handler: (data: any) => void): () => void {
    if (!this.messageHandlers.has(messageType)) {
      this.messageHandlers.set(messageType, new Set());
    }

    const handlers = this.messageHandlers.get(messageType)!;
    handlers.add(handler);

    return () => {
      handlers.delete(handler);
      if (handlers.size === 0) {
        this.messageHandlers.delete(messageType);
      }
    };
  }

  /**
   * 注册状态变化监听器
   */
  onStatusChange(listener: (status: WSStatus) => void): () => void {
    this.statusListeners.add(listener);

    return () => {
      this.statusListeners.delete(listener);
    };
  }

  /**
   * 获取当前状态
   */
  getStatus(): WSStatus {
    return this.status;
  }

  /**
   * 处理接收到的消息
   */
  private handleMessage(message: WSMessage): void {
    if (message.type === 'connected') {
      return;
    }

    if (message.type === 'token_refresh' && message.data?.token) {
      this.handleTokenRefresh(message.data.token);
      return;
    }

    // 分发消息到对应的处理器
    const handlers = this.messageHandlers.get(message.type);
    if (handlers) {
      handlers.forEach(handler => {
        try {
          handler(message.data);
        } catch (error) {
          console.error(`[WebSocket] 消息处理器执行失败 (${message.type}):`, error);
        }
      });
    }
  }

  /**
   * 处理服务端颁发的续期 Token
   */
  private handleTokenRefresh(newToken: string): void {
    this.refreshToken(newToken);
  }

  /**
   * 更新本地 token 和 WebSocket Cookie（供外部调用，如 HTTP 响应拦截器）
   */
  refreshToken(newToken: string): void {
    localStorage.setItem('token', newToken);
    // 同步更新 Cookie（供后续重连使用）
    const secureFlag = this.options.secure ? '; Secure' : '';
    document.cookie = `pms_token=${encodeURIComponent(newToken)}; path=/; max-age=3600; SameSite=Strict${secureFlag}`;
  }

  /**
   * 安排重连
   */
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
      console.error('[WebSocket] 达到最大重连次数，停止重连');
      return;
    }

    this.status = WSStatus.RECONNECTING;
    this.notifyStatusChange();

    // 指数退避策略
    const delay = this.options.reconnectInterval * Math.pow(2, this.reconnectAttempts);
    
    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++;
      this.connect();
    }, delay);
  }

  /**
   * 启动心跳
   */
  private startHeartbeat(): void {
    this.stopHeartbeat();

    this.heartbeatTimer = setInterval(() => {
      if (this.ws && this.status === WSStatus.CONNECTED) {
        this.send({ type: 'ping' });
      }
    }, this.options.heartbeatInterval);
  }

  /**
   * 停止心跳
   */
  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  /**
   * 通知状态变化
   */
  private notifyStatusChange(): void {
    this.statusListeners.forEach(listener => {
      try {
        listener(this.status);
      } catch (error) {
        console.error('[WebSocket] 状态监听器执行失败:', error);
      }
    });
  }
}

// 单例实例
let wsClient: WebSocketClient | null = null;

/**
 * 初始化 WebSocket 客户端
 */
export function initWebSocket(options: WSOptions): WebSocketClient {
  if (wsClient) {
    console.warn('[WebSocket] 客户端已初始化，请勿重复调用');
    return wsClient;
  }

  wsClient = new WebSocketClient(options);
  return wsClient;
}

/**
 * 获取 WebSocket 客户端实例
 */
export function getWebSocketClient(): WebSocketClient | null {
  return wsClient;
}

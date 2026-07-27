package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/common/middleware"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

// WSClient WebSocket 客户端
type WSClient struct {
	ID     string          // 客户端唯一标识
	Conn   *websocket.Conn // WebSocket 连接
	UserID int64           // 用户 ID
	Send   chan []byte     // 发送消息的通道
}

// WSManager WebSocket 管理器
type WSManager struct {
	Clients    map[string]*WSClient // 所有连接的客户端
	Broadcast  chan []byte          // 广播消息通道
	Register   chan *WSClient       // 注册客户端
	Unregister chan *WSClient       // 注销客户端
	done       chan struct{}        // 关闭信号
	Mutex      sync.RWMutex         // 读写锁
}

// 全局 WebSocket 管理器
var WS *WSManager

func init() {
	WS = &WSManager{
		Clients:    make(map[string]*WSClient),
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *WSClient, 64),   // 带缓冲，避免短时大量连接阻塞
		Unregister: make(chan *WSClient, 64),   // 带缓冲，避免短时大量断开阻塞
		done:       make(chan struct{}),
	}

	go WS.Run()
}

// Run 运行 WebSocket 管理器（启动前先 recover 防止 panic 导致全部泄漏）
func (m *WSManager) Run() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("WebSocket 管理器 panic，重启中",
				zap.Any("recover", r),
			)
			// 重启管理器
			time.Sleep(time.Second)
			go m.Run()
		}
	}()

	for {
		select {
		case <-m.done:
			// 收到关闭信号，断开所有客户端
			logger.Info("WebSocket 管理器正在关闭，断开所有客户端")
			m.Mutex.Lock()
			for _, client := range m.Clients {
				delete(m.Clients, client.ID)
				close(client.Send)
				client.Conn.Close()
			}
			m.Mutex.Unlock()
			logger.Info("WebSocket 管理器已关闭")
			return

		case client := <-m.Register:
			m.Mutex.Lock()
			m.Clients[client.ID] = client
			m.Mutex.Unlock()
			logger.Info("WebSocket 客户端已连接",
				zap.String("client_id", client.ID),
				zap.Int64("user_id", client.UserID),
				zap.Int("total_clients", len(m.Clients)),
			)

		case client := <-m.Unregister:
			// 注销客户端
			m.Mutex.Lock()
			if _, ok := m.Clients[client.ID]; ok {
				delete(m.Clients, client.ID)
				close(client.Send)
				logger.Info("WebSocket 客户端已断开",
					zap.String("client_id", client.ID),
					zap.Int("total_clients", len(m.Clients)),
				)
			}
			m.Mutex.Unlock()

		case message := <-m.Broadcast:
			// 广播消息给所有客户端
			m.Mutex.RLock()
			for _, client := range m.Clients {
				select {
				case client.Send <- message:
				default:
					logger.Warn("WebSocket 客户端缓冲区满，跳过消息",
						zap.String("client_id", client.ID),
					)
				}
			}
			m.Mutex.RUnlock()
		}
	}
}

// Shutdown 优雅关闭 WebSocket 管理器，等待所有客户端处理完成
func (m *WSManager) Shutdown() {
	logger.Info("WebSocket 管理器收到关闭请求")
	close(m.done)
}

// BroadcastMessage 广播消息（供外部调用）
func (m *WSManager) BroadcastMessage(eventType string, data interface{}) {
	message := map[string]interface{}{
		"type":      eventType,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		logger.Error("WebSocket 消息序列化失败", zap.Error(err))
		return
	}

	select {
	case m.Broadcast <- messageJSON:
	default:
		logger.Warn("WebSocket 广播通道已满")
	}

	logger.Info("WebSocket 广播消息", zap.String("event", eventType))
}

// SendToUser 发送消息给指定用户（所有该用户的连接）
func (m *WSManager) SendToUser(userID int64, eventType string, data interface{}) {
	message := map[string]interface{}{
		"type":      eventType,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		logger.Error("WebSocket 消息序列化失败", zap.Error(err))
		return
	}

	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	for _, client := range m.Clients {
		if client.UserID == userID {
			select {
			case client.Send <- messageJSON:
			default:
				logger.Warn("WebSocket 客户端缓冲区满，跳过消息",
					zap.Int64("user_id", userID),
					zap.String("client_id", client.ID),
				)
			}
		}
	}
}

// SendToUsers 发送消息给多个指定用户
func (m *WSManager) SendToUsers(userIDs []int64, eventType string, data interface{}) {
	userSet := make(map[int64]struct{}, len(userIDs))
	for _, id := range userIDs {
		userSet[id] = struct{}{}
	}

	message := map[string]interface{}{
		"type":      eventType,
		"data":      data,
		"timestamp": time.Now().Unix(),
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		logger.Error("WebSocket 消息序列化失败", zap.Error(err))
		return
	}

	m.Mutex.RLock()
	defer m.Mutex.RUnlock()

	for _, client := range m.Clients {
		if _, ok := userSet[client.UserID]; ok {
			select {
			case client.Send <- messageJSON:
			default:
				logger.Warn("WebSocket 客户端缓冲区满，跳过消息",
					zap.Int64("user_id", client.UserID),
					zap.String("client_id", client.ID),
				)
			}
		}
	}
}

// HandleWebSocket 处理 WebSocket 连接
func (c *SystemController) HandleWebSocket(ctx *gin.Context) {
	// 升级 HTTP 连接到 WebSocket
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			// 无 Origin 头（如 Postman、非浏览器客户端），允许连接
			if origin == "" {
				return true
			}
			// 开发/测试模式：放行所有来源，方便 VM 测试部署，无需反复修改配置
			if config.GlobalConfig.App.Mode == "development" {
				return true
			}
			// 生产模式：严格使用 CORS 白名单校验
			for _, allowed := range config.GlobalConfig.CORS.AllowedOrigins {
				if origin == allowed {
					return true
				}
			}
			return false
		},
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logger.Error("WebSocket 升级失败", zap.Error(err))
		return
	}

	var token string
	token, err = ctx.Cookie("pms_token")
	if err != nil || token == "" {
		token = ctx.Query("token")
	}

	if token == "" {
		logger.Warn("WebSocket 连接失败：未提供认证令牌",
			zap.String("client_ip", ctx.ClientIP()),
		)
		conn.Close()
		return
	}

	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(config.GlobalConfig.JWT.Secret), nil
	})

	if err != nil || !jwtToken.Valid {
		logger.Warn("WebSocket Token 验证失败",
			zap.String("client_ip", ctx.ClientIP()),
		)
		conn.Close()
		return
	}

	claims := jwtToken.Claims.(jwt.MapClaims)
	userID := int64(claims["user_id"].(float64))
	username := ""
	if u, ok := claims["username"].(string); ok {
		username = u
	}

	// === Token 滑动续期 ===
	// 每次 WebSocket 连接时自动顺延 token 过期时间，推送给客户端
	var keepLogin bool
	if kl, ok := claims["keep_login"].(bool); ok {
		keepLogin = kl
	}
	pwdVer := 0
	if pv, ok := claims["pwd_ver"].(float64); ok {
		pwdVer = int(pv)
	}

	newToken, err := middleware.GenerateToken(userID, username, keepLogin, pwdVer)
	if err == nil && newToken != "" {
		logger.Info("WebSocket Token 已自动续期",
			zap.Int64("user_id", userID),
		)
		conn.WriteJSON(gin.H{
			"type": "token_refresh",
			"data": gin.H{
				"token": newToken,
			},
		})
	}

	clientID := fmt.Sprintf("ws_user_%d_%d", userID, time.Now().UnixNano())
	client := &WSClient{
		ID:     clientID,
		Conn:   conn,
		UserID: userID,
		Send:   make(chan []byte, 256),
	}

	// 注册客户端（使用 select 防止管理器停止时阻塞）
	select {
	case WS.Register <- client:
	default:
		logger.Warn("WebSocket 管理器已关闭，拒绝新连接",
			zap.Int64("user_id", userID),
		)
		conn.Close()
		return
	}

	conn.WriteJSON(gin.H{
		"type": "connected",
		"data": gin.H{
			"message":   "WebSocket 连接成功",
			"client_id": clientID,
			"timestamp": time.Now().Unix(),
		},
	})

	go client.readPump()
	go client.writePump()
}

// readPump 读取客户端消息
func (c *WSClient) readPump() {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("readPump panic 已恢复", zap.Any("recover", r))
		}
		// 使用 select 防止 Unregister 通道在管理器关闭时阻塞
		select {
		case WS.Unregister <- c:
		default:
		}
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket 读取错误", zap.Error(err))
			}
			break
		}

		logger.Debug("收到 WebSocket 消息",
			zap.String("client_id", c.ID),
			zap.String("message", string(message)),
		)
	}
}

// writePump 向客户端发送消息
func (c *WSClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("writePump panic 已恢复", zap.Any("recover", r))
		}
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func NotifyDictUpdated(dictType string) {
	data := map[string]interface{}{
		"dict_type": dictType,
	}
	WS.BroadcastMessage("dict-updated", data)
}

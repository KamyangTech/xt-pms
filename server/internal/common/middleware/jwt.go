package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/KamyangTech/xt-pms/server/config"
	"github.com/KamyangTech/xt-pms/server/internal/common/response"
	"github.com/KamyangTech/xt-pms/server/internal/dao"
	"github.com/KamyangTech/xt-pms/server/pkg/logger"
	"go.uber.org/zap"
)

// JWT 认证中间件
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, response.UNAUTHORIZED, "未提供认证信息")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Error(c, response.UNAUTHORIZED, "认证格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.GlobalConfig.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			response.Error(c, response.UNAUTHORIZED, "认证信息无效或已过期")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.Error(c, response.UNAUTHORIZED, "无法解析用户信息")
			c.Abort()
			return
		}

		if claims["user_id"] == nil || claims["username"] == nil {
			response.Error(c, response.UNAUTHORIZED, "认证信息不完整")
			c.Abort()
			return
		}

		// 验证密码版本号（确保 Token 在密码修改后失效），带内存级缓存
		userID := int64(claims["user_id"].(float64))
		tokenPwdVer, hasVersion := claims["pwd_ver"].(float64)
		if hasVersion {
			// 优先从缓存读取，避免每次请求查库
			var dbPwdVer int
			if cachedVer, ok := GetPwdVersionFromCache(userID); ok {
				dbPwdVer = cachedVer // 缓存命中
			} else {
				// 缓存未命中，查询数据库并写入缓存
				var user struct {
					PasswordVersion int
				}
				if err := dao.GetDB().Table("sys_user").
					Select("password_version").
					Where("id = ?", userID).
					First(&user).Error; err != nil {
					logger.Warn("JWT 验证失败：查询用户密码版本号出错",
						zap.Int64("user_id", userID),
						zap.Error(err),
					)
					response.Error(c, response.UNAUTHORIZED, "认证信息无效")
					c.Abort()
					return
				}
				dbPwdVer = user.PasswordVersion
				SetPwdVersionCache(userID, dbPwdVer) // 写入缓存
			}

			if int(tokenPwdVer) != dbPwdVer {
				logger.Warn("JWT 验证失败：密码版本号不匹配（密码已修改）",
					zap.Int64("user_id", userID),
					zap.Int("token_version", int(tokenPwdVer)),
					zap.Int("db_version", dbPwdVer),
				)
				response.Error(c, response.UNAUTHORIZED, "认证信息已过期，请重新登录")
				c.Abort()
				return
			}
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims["user_id"])
		c.Set("username", claims["username"])

		// 滑动过期：每次请求自动顺延 token 过期时间
		// 确保闲置从最后一次操作开始计算，用户体验可预测
		keepLogin := false
		if kl, ok := claims["keep_login"].(bool); ok {
			keepLogin = kl
		}
		pwdVer := 0
		if hasVersion {
			pwdVer = int(tokenPwdVer)
		}
		newToken, err := GenerateToken(userID, claims["username"].(string), keepLogin, pwdVer)
		if err == nil {
			c.Header("X-Token-Refresh", newToken)
		}

		c.Next()
	}
}

// GetCurrentUserID 从 Context 获取用户 ID（指针版本）
func GetCurrentUserID(c *gin.Context) *int64 {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil
	}

	// user_id 在 JWT 中存储为 float64（JSON 数字默认类型）
	switch v := userID.(type) {
	case float64:
		id := int64(v)
		return &id
	case int64:
		return &v
	case int:
		id := int64(v)
		return &id
	default:
		return nil
	}
}

// GetUserID 从 Context 获取用户 ID
func GetUserID(c *gin.Context) int64 {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	// JWT claims 中的数字会被解析为 float64
	switch v := userID.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	default:
		return 0
	}
}

// GetUsername 从 Context 获取用户名
func GetUsername(c *gin.Context) string {
	username, exists := c.Get("username")
	if !exists {
		return ""
	}
	if u, ok := username.(string); ok {
		return u
	}
	return ""
}

// GenerateToken 生成 JWT token
// keepLogin: 是否保持登录（true=7天，false=使用配置过期时间）
// passwordVersion: 用户当前的密码版本号，用于 Token 可吊销。为 0 时不做版本校验（向后兼容）
func GenerateToken(userID int64, username string, keepLogin bool, passwordVersion ...int) (string, error) {
	var expire int
	if keepLogin {
		expire = 24 * 3600 // 24 小时
	} else {
		// 优先从动态安全配置读取过期时间，fallback 到 config.yaml
		expire = config.GetJWTExpire()
	}
	claims := jwt.MapClaims{
		"user_id":    userID,
		"username":   username,
		"keep_login": keepLogin,
		"exp":        time.Now().Add(time.Duration(expire) * time.Second).Unix(),
	}

	// 传入密码版本号则写入 claims
	if len(passwordVersion) > 0 && passwordVersion[0] > 0 {
		claims["pwd_ver"] = passwordVersion[0]
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GlobalConfig.JWT.Secret))
}

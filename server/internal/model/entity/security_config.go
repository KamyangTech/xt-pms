package entity

import "time"

// SysSecurityConfig 安全配置（登录与安全）
type SysSecurityConfig struct {
	ID              int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TokenExpire     int       `gorm:"column:token_expire;type:int;not null;default:7200;comment:Token有效期(秒)" json:"token_expire"`
	MaxLoginAttempts int      `gorm:"column:max_login_attempts;type:int;not null;default:5;comment:最大登录失败次数" json:"max_login_attempts"`
	LockDuration    int       `gorm:"column:lock_duration;type:int;not null;default:900;comment:账号锁定时长(秒),0表示需管理员解锁" json:"lock_duration"`
	PasswordMinLen  int       `gorm:"column:password_min_len;type:int;not null;default:6;comment:密码最小长度" json:"password_min_len"`
	PasswordRequireUpper bool `gorm:"column:password_require_upper;type:tinyint(1);not null;default:0;comment:是否要求大写字母" json:"password_require_upper"`
	PasswordRequireNumber bool `gorm:"column:password_require_number;type:tinyint(1);not null;default:1;comment:是否要求数字" json:"password_require_number"`
	PasswordRequireSpecial bool `gorm:"column:password_require_special;type:tinyint(1);not null;default:0;comment:是否要求特殊字符" json:"password_require_special"`
	CookieSecure       bool  `gorm:"column:cookie_secure;type:tinyint(1);not null;default:0;comment:Cookie Secure标志：1=仅HTTPS发送" json:"cookie_secure"`
	WebSocketSecure    bool  `gorm:"column:websocket_secure;type:tinyint(1);not null;default:0;comment:WebSocket Secure标志：1=仅WSS发送" json:"websocket_secure"`
	CreateTime      time.Time `gorm:"column:create_time;autoCreateTime" json:"create_time"`
	UpdateTime      time.Time `gorm:"column:update_time;autoUpdateTime" json:"update_time"`
}

// TableName 指定表名
func (SysSecurityConfig) TableName() string {
	return "sys_security_config"
}

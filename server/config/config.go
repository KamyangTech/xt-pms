package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	CORS     CORSConfig     `yaml:"cors"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	Security SecurityConfig `yaml:"security"`
	App      AppConfig      `yaml:"app"`
	Log      LogConfig      `yaml:"log"`
	Upload   UploadConfig   `yaml:"upload"`
	Auth     AuthConfig     `yaml:"auth"` // 权限认证相关配置
}

type AuthConfig struct {
	SuperAdminUserID  int64 `yaml:"super_admin_user_id"`  // 超级管理员用户ID
	SuperAdminRoleType int8  `yaml:"super_admin_role_type"` // 超级管理员角色类型(0自定义/1系统内置)
}

type ServerConfig struct {
	Port         int  `yaml:"port"`
	CookieSecure bool `yaml:"cookie_secure"` // Cookie Secure 标志：true=仅HTTPS发送，false=HTTP/HTTPS均可
	MaxBodySize  int  `yaml:"max_body_size"` // 请求体最大大小（字节），默认 10MB
}

type CORSConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"` // 允许的跨域来源列表
}

type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	DBName          string        `yaml:"dbname"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
	Expire int    `yaml:"expire"`
}

type SecurityConfig struct {
	Pepper string `yaml:"pepper"`
}

type AppConfig struct {
	Mode             string `yaml:"mode"`               // development 或 production
	UpdateSourceURL  string `yaml:"update_source_url"`  // 系统更新源 URL（默认 Gitee）
	ServiceName      string `yaml:"service_name"`       // systemd 服务名（默认 xt-pms，1Panel 用户需改为面板中对应的服务名）
	UpdateMode       string `yaml:"update_mode"`        // 更新模式：auto(默认自动检测) / systemd / supervisor / manual
}

type LogConfig struct {
	Level  string     `yaml:"level"`  // debug/info/warn/error
	Format string     `yaml:"format"` // json/console
	Output string     `yaml:"output"` // both/file/console
	File   FileConfig `yaml:"file"`
}

type FileConfig struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"max_size"` // MB
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"` // 天
	Compress   bool   `yaml:"compress"`
}

type UploadConfig struct {
	BasePath           string   `yaml:"base_path"`             // 本地存储根目录
	MaxSize            int64    `yaml:"max_size"`              // 最大文件大小（字节）
	AllowedImageTypes  []string `yaml:"allowed_image_types"`   // 允许的图片类型（向后兼容）
	AllowedFileTypes   []string `yaml:"allowed_file_types"`    // 允许的所有文件类型
	ThumbnailMaxWidth  int      `yaml:"thumbnail_max_width"`   // 缩略图最大宽度（像素），默认 200
	ThumbnailMaxHeight int      `yaml:"thumbnail_max_height"`  // 缩略图最大高度（像素），默认 200
}

var GlobalConfig *Config

// jwtExpireOverride 运行时覆盖的 JWT 过期时间（秒），0 表示未设置，使用配置文件值
var jwtExpireOverride int = 0

// GetJWTExpire 获取当前有效的 JWT 过期时间（优先使用运行时动态配置，fallback 到 YAML 配置）
func GetJWTExpire() int {
	if jwtExpireOverride > 0 {
		return jwtExpireOverride
	}
	return GlobalConfig.JWT.Expire
}

// SetJWTExpireOverride 设置运行时的 JWT 过期时间覆盖值（由 SecurityConfigService 调用）
func SetJWTExpireOverride(expire int) {
	jwtExpireOverride = expire
}

// loadEnvFile 从 .env 文件加载环境变量到当前进程（仅设置尚未被 OS 环境变量覆盖的键）
// 优先级：OS 环境变量 > .env 文件 > config.yaml
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // .env 文件不存在，静默跳过
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		// 去除可选引号
		value = strings.Trim(value, `"'`)

		// 仅在 OS 环境变量未设置时才从 .env 注入
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	_ = scanner.Err() // 读取错误容忍，非关键配置不中断启动
}

// envOverrideString 通过环境变量覆盖字符串配置（环境变量非空时生效）
func envOverrideString(target *string, key string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}

// envOverrideInt 通过环境变量覆盖整数配置
func envOverrideInt(target *int, key string) {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*target = n
		}
	}
}

// envOverrideBool 通过环境变量覆盖布尔配置
func envOverrideBool(target *bool, key string) {
	if v := os.Getenv(key); v != "" {
		*target = strings.EqualFold(v, "true") || v == "1"
	}
}

// envOverrides 使用环境变量覆盖敏感配置项（优先级高于 YAML 文件）
//
// 适用场景：
//   - 生产环境 K8s/容器部署，通过 Secrets/ConfigMap 注入
//   - CI/CD 流水线中通过环境变量传入密钥
//   - 本地开发通过 .env 文件或 shell export
//
// 优先级：环境变量 > YAML 文件 > 默认值
func envOverrides() {
	// 数据库连接
	envOverrideString(&GlobalConfig.Database.Host, "PMS_DB_HOST")
	envOverrideInt(&GlobalConfig.Database.Port, "PMS_DB_PORT")
	envOverrideString(&GlobalConfig.Database.User, "PMS_DB_USER")
	envOverrideString(&GlobalConfig.Database.Password, "PMS_DB_PASSWORD")
	envOverrideString(&GlobalConfig.Database.DBName, "PMS_DB_NAME")

	// JWT 密钥
	envOverrideString(&GlobalConfig.JWT.Secret, "PMS_JWT_SECRET")

	// 安全配置（密码加密 Pepper）
	envOverrideString(&GlobalConfig.Security.Pepper, "PMS_SECURITY_PEPPER")

	// 应用模式
	envOverrideString(&GlobalConfig.App.Mode, "PMS_APP_MODE")

	// 系统更新源
	envOverrideString(&GlobalConfig.App.UpdateSourceURL, "PMS_UPDATE_SOURCE_URL")

	// 服务名（1Panel 部署时服务名可能不同）
	envOverrideString(&GlobalConfig.App.ServiceName, "PMS_SERVICE_NAME")

	envOverrideString(&GlobalConfig.App.UpdateMode, "PMS_UPDATE_MODE")

	// 服务器配置
	envOverrideInt(&GlobalConfig.Server.Port, "PMS_SERVER_PORT")
	envOverrideBool(&GlobalConfig.Server.CookieSecure, "PMS_COOKIE_SECURE")

	// 日志配置
	envOverrideString(&GlobalConfig.Log.Level, "PMS_LOG_LEVEL")
}

func Init() error {
	// 支持通过环境变量 PMS_CONFIG_PATH 指定配置文件路径
	configPath := os.Getenv("PMS_CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	// 1. 优先加载 .env 文件（与 config.yaml 同级目录）
	envPath := filepath.Join(filepath.Dir(configPath), ".env")
	loadEnvFile(envPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	GlobalConfig = &Config{}
	if err := yaml.Unmarshal(data, GlobalConfig); err != nil {
		return err
	}

	// 环境变量覆盖（优先级高于 YAML）
	envOverrides()

	return nil
}

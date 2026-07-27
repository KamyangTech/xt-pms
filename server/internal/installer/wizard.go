package installer

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"

	_ "github.com/go-sql-driver/mysql"
)

// ============================================
// 步骤 1：测试数据库连接
// ============================================

// TestDBConnection 测试 MySQL 连接是否可用
func CheckDBIsEmpty(host string, port int, user, password, dbName string) (bool, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbName,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, err
	}
	defer db.Close()

	db.SetConnMaxLifetime(5 * time.Second)
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ?", dbName).Scan(&count); err != nil {
		return false, err
	}
	return count == 0, nil
}

// TestDBConnection 测试 MySQL 连接是否可用
func TestDBConnection(host string, port int, user, password, dbName string) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbName,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("无法打开连接：%w", err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(5 * time.Second)
	if err := db.Ping(); err != nil {
		return fmt.Errorf("连接失败：%w", err)
	}
	return nil
}

// ============================================
// 步骤 2：生成安全配置
// ============================================

// GeneratePepper 生成随机 Pepper（32 字节 hex）
func GeneratePepper() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateJWTSecret 生成随机 JWT 密钥（32 字节 hex）
func GenerateJWTSecret() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ============================================
// 步骤 3：执行安装
// ============================================

type InstallConfig struct {
	// 数据库
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"db_password"`
	DBName     string `json:"db_name"`

	// 安全
	Pepper    string `json:"pepper"`
	JWTSecret string `json:"jwt_secret"`

	// 应用
	AppMode      string `json:"app_mode"`
	ServerPort   int    `json:"server_port"`
	CookieSecure bool   `json:"cookie_secure"`

	// 管理员
	AdminUsername string `json:"admin_username"`
	AdminPassword string `json:"admin_password"`
	AdminRealName string `json:"admin_real_name"`
	AdminEmail    string `json:"admin_email"`
	AdminPhone    string `json:"admin_phone"`
}

type InstallResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Step    string `json:"step"` // 当前步骤，用于进度展示
}

func RunInstall(cfg *InstallConfig) *InstallResult {
	// 0. 检查数据库是否为空
	empty, err := CheckDBIsEmpty(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("检查数据库失败：%v", err), Step: "database"}
	}
	if !empty {
		return &InstallResult{Success: false, Message: "数据库不为空，已有表存在。请使用空库，或手动清理后重试。", Step: "database"}
	}

	// 1. 写入 config.yaml
	if err := writeConfigYAML(cfg); err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("写入配置文件失败：%v", err), Step: "config"}
	}

	// 2. 写入 .env
	if err := writeEnvFile(cfg); err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("写入环境变量失败：%v", err), Step: "env"}
	}

	// 3. 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("连接数据库失败：%v", err), Step: "database"}
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("数据库不可达：%v", err), Step: "database"}
	}

	// 4. 执行建表脚本
	schemaSQL, err := readMigrationFile("init_schema.sql")
	if err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("读取建表脚本失败：%v", err), Step: "schema"}
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("执行建表脚本失败：%v", err), Step: "schema"}
	}

	// 5. 执行种子数据脚本
	seedSQL, err := readMigrationFile("init_seed.sql")
	if err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("读取种子数据失败：%v", err), Step: "seed"}
	}
	if _, err := db.Exec(seedSQL); err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("执行种子数据失败：%v", err), Step: "seed"}
	}

	// 6. 创建管理员账号（使用正确的 pepper 生成密码哈希）
	pepperedPassword := cfg.AdminPassword + cfg.Pepper
	hash, err := bcrypt.GenerateFromPassword([]byte(pepperedPassword), bcrypt.DefaultCost)
	if err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("生成密码哈希失败：%v", err), Step: "admin"}
	}

	_, err = db.Exec(
		`INSERT INTO sys_user (username, password, real_name, email, phone, status) VALUES (?, ?, ?, ?, ?, 1)`,
		cfg.AdminUsername, string(hash), cfg.AdminRealName, cfg.AdminEmail, cfg.AdminPhone,
	)
	if err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("创建管理员失败：%v", err), Step: "admin"}
	}

	// 7. 分配超级管理员角色
	_, err = db.Exec(
		`INSERT INTO sys_user_role (user_id, role_id) SELECT u.id, r.id FROM sys_user u, sys_role r WHERE u.username = ? AND r.role_name = '超级管理员'`,
		cfg.AdminUsername,
	)
	if err != nil {
		return &InstallResult{Success: false, Message: fmt.Sprintf("分配角色失败：%v", err), Step: "admin"}
	}

	// 8. 更新系统信息中的公司名称
	if cfg.AdminRealName != "" {
		db.Exec("UPDATE sys_company SET company_name = ? WHERE id = 1", cfg.AdminRealName+"的管理系统")
	}

	return &InstallResult{Success: true, Message: "安装完成！系统已就绪。", Step: "done"}
}

// ============================================
// 辅助函数
// ============================================

// readMigrationFile 从 migrations 目录读取 SQL 文件
func readMigrationFile(filename string) (string, error) {
	// 尝试多个可能的路径
	candidates := []string{
		filepath.Join("migrations", filename),
		filepath.Join("..", "migrations", filename),
		filepath.Join("..", "..", "migrations", filename),
		filepath.Join(filepath.Dir(os.Args[0]), "migrations", filename),
	}
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("找不到 %s，已尝试：%s", filename, strings.Join(candidates, ", "))
}

// writeConfigYAML 生成 config/config.yaml
func writeConfigYAML(cfg *InstallConfig) error {
	configMap := map[string]interface{}{
		"server": map[string]interface{}{
			"port":          cfg.ServerPort,
			"cookie_secure": cfg.CookieSecure,
		},
		"database": map[string]interface{}{
			"host":     cfg.DBHost,
			"port":     cfg.DBPort,
			"user":     cfg.DBUser,
			"password": cfg.DBPassword,
			"dbname":   cfg.DBName,
		},
		"jwt": map[string]interface{}{
			"secret": cfg.JWTSecret,
			"expire": 7200,
		},
		"security": map[string]interface{}{
			"pepper": cfg.Pepper,
		},
		"app": map[string]interface{}{
			"mode": cfg.AppMode,
		},
		"log": map[string]interface{}{
			"level":  "info",
			"format": "console",
			"output": "both",
			"file": map[string]interface{}{
				"path":        "logs/app.log",
				"max_size":    100,
				"max_backups": 30,
				"max_age":     28,
				"compress":    true,
			},
		},
		"upload": map[string]interface{}{
			"base_path":           "uploads",
			"max_size":            10485760,
			"allowed_image_types": []string{".jpg", ".jpeg", ".png", ".gif", ".svg", ".ico", ".webp"},
			"allowed_file_types":  []string{".jpg", ".jpeg", ".png", ".gif", ".svg", ".ico", ".webp", ".pdf", ".doc", ".docx", ".xls", ".xlsx"},
			"thumbnail_max_width":  200,
			"thumbnail_max_height": 200,
		},
		"cors": map[string]interface{}{
			// 默认同源部署（前后端同一 domain），浏览器不会触发跨域，无需配置 CORS
			// 如需前后端分离部署，在此添加前端域名，如 "https://frontend.your-domain.com"
			"allowed_origins": []string{},
		},
		"auth": map[string]interface{}{
			"super_admin_user_id":   1,
			"super_admin_role_type": 1,
		},
	}

	os.MkdirAll("config", 0755)
	data, err := yaml.Marshal(configMap)
	if err != nil {
		return err
	}

	header := "# Xt-PMS 配置文件\n# 由安装向导自动生成\n# 敏感值通过 config/.env 覆盖，优先级更高\n\n"
	return os.WriteFile("config/config.yaml", append([]byte(header), data...), 0644)
}

// writeEnvFile 生成 config/.env（敏感值通过环境变量覆盖）
func writeEnvFile(cfg *InstallConfig) error {
	content := fmt.Sprintf(`# Xt-PMS 环境变量（敏感配置）
# 由安装向导自动生成
# 优先级：OS 环境变量 > .env 文件 > config.yaml

PMS_DB_HOST=%s
PMS_DB_PORT=%d
PMS_DB_USER=%s
PMS_DB_PASSWORD=%s
PMS_DB_NAME=%s
PMS_JWT_SECRET=%s
PMS_SECURITY_PEPPER=%s
PMS_APP_MODE=%s
PMS_SERVER_PORT=%d
PMS_COOKIE_SECURE=%v
`, cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
		cfg.JWTSecret, cfg.Pepper, cfg.AppMode, cfg.ServerPort, cfg.CookieSecure)

	os.MkdirAll("config", 0755)
	return os.WriteFile("config/.env", []byte(content), 0600)
}

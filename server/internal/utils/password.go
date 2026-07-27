package utils

import (
	"github.com/KamyangTech/xt-pms/server/config"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt + Pepper 加密密码
func HashPassword(password string) (string, error) {
	// 在密码后添加 Pepper
	pepper := config.GlobalConfig.Security.Pepper
	pepperedPassword := password + pepper

	// 使用 bcrypt 加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pepperedPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

// VerifyPassword 验证密码（使用 bcrypt + Pepper）
func VerifyPassword(hash, password string) bool {
	// 在密码后添加 Pepper
	pepper := config.GlobalConfig.Security.Pepper
	pepperedPassword := password + pepper

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pepperedPassword))
	return err == nil
}

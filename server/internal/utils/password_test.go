// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"testing"

	"github.com/KamyangTech/xt-pms/server/config"
)

func init() {
	config.GlobalConfig = &config.Config{
		Security: config.SecurityConfig{Pepper: "test-pepper-123"},
	}
}

func TestHashPasswordAndVerify(t *testing.T) {
	plain := "S3cretP@ss"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hash == plain {
		t.Fatal("hash 不应等于明文")
	}
	if !VerifyPassword(hash, plain) {
		t.Fatal("VerifyPassword 应返回 true（正确密码）")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("VerifyPassword 应返回 false（错误密码）")
	}
}

func TestHashPasswordDifferentSalts(t *testing.T) {
	h1, _ := HashPassword("same-password")
	h2, _ := HashPassword("same-password")
	if h1 == h2 {
		t.Fatal("相同明文两次哈希结果应不同（bcrypt 含随机 salt）")
	}
}

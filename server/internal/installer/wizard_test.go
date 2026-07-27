// SPDX-License-Identifier: AGPL-3.0-or-later

package installer

import (
	"encoding/hex"
	"testing"
)

func TestGeneratePepper(t *testing.T) {
	p1 := GeneratePepper()
	p2 := GeneratePepper()
	if len(p1) != 64 {
		t.Fatalf("Pepper 长度期望 64，得到 %d", len(p1))
	}
	if _, err := hex.DecodeString(p1); err != nil {
		t.Fatalf("Pepper 非合法 hex: %v", err)
	}
	if p1 == p2 {
		t.Fatal("两次生成的 Pepper 不应相同")
	}
}

func TestGenerateJWTSecret(t *testing.T) {
	s1 := GenerateJWTSecret()
	s2 := GenerateJWTSecret()
	if len(s1) != 64 {
		t.Fatalf("JWTSecret 长度期望 64，得到 %d", len(s1))
	}
	if _, err := hex.DecodeString(s1); err != nil {
		t.Fatalf("JWTSecret 非合法 hex: %v", err)
	}
	if s1 == s2 {
		t.Fatal("两次生成的 JWTSecret 不应相同")
	}
}

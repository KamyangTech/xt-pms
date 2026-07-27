// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"errors"
	"testing"
)

func TestEnsureExists(t *testing.T) {
	// 值存在（非 nil 接口）应返回 nil
	if err := EnsureExists(1, nil, "订单"); err != nil {
		t.Fatalf("期望 nil，得到 %v", err)
	}
	if err := EnsureExists(nil, errors.New("db error"), "订单"); err == nil {
		t.Fatal("期望返回错误（err 非 nil）")
	}
	// 接口为 nil（查询未找到）应返回错误
	if err := EnsureExists(nil, nil, "订单"); err == nil {
		t.Fatal("期望返回错误（值不存在）")
	}
}

func TestEnsureExistsWithCustomMsg(t *testing.T) {
	// 错误时应返回自定义消息
	err := EnsureExistsWithCustomMsg(nil, errors.New("x"), "自定义消息")
	if err == nil {
		t.Fatal("期望返回自定义错误")
	} else if err.Error() != "自定义消息" {
		t.Fatalf("期望 '自定义消息'，得到 %q", err.Error())
	}
	// 正常时应返回 nil
	if err := EnsureExistsWithCustomMsg("ok", nil, "自定义消息"); err != nil {
		t.Fatalf("期望 nil，得到 %v", err)
	}
}

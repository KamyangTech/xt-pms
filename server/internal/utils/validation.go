package utils

import (
	"errors"
	"fmt"
)

// EnsureExists 通用存在性检查函数
// value: 查询返回的值（指针类型）
// err: 查询返回的错误
// entityName: 实体名称（如"订单"、"工单"、"物料"等）
// 如果 err != nil 或 value == nil，返回 "entityName不存在" 错误
func EnsureExists(value interface{}, err error, entityName string) error {
	if err != nil {
		return fmt.Errorf("%s不存在", entityName)
	}

	if value == nil {
		return fmt.Errorf("%s不存在", entityName)
	}

	return nil
}

func EnsureExistsWithCustomMsg(value interface{}, err error, customMsg string) error {
	if err != nil || value == nil {
		return errors.New(customMsg)
	}
	return nil
}

package controller

import (
	"errors"
	"fmt"
	"time"
)

// validatePlanDates 校验计划开始和计划结束日期
// 规则：
//  1. 结束日期不能早于开始日期（可同一天）
//  2. 如果开始日期为历史时间，结束日期最早只能是今天（防止产生历史订单）
func validatePlanDates(start, end time.Time) error {
	today := time.Now().Truncate(24 * time.Hour)

	// 规则1：结束日期不能早于开始日期
	if end.Before(start) {
		return errors.New("计划结束日期不能早于计划开始日期")
	}

	// 规则2：如果开始日期为历史时间，结束日期最早只能是今天
	if start.Before(today) && end.Before(today) {
		return errors.New("计划结束日期不能早于今天")
	}

	return nil
}

// validatePlanDateStrings 校验字符串格式的计划开始/结束日期
func validatePlanDateStrings(startStr, endStr *string) error {
	if startStr == nil || endStr == nil || *startStr == "" || *endStr == "" {
		return nil
	}

	start, err := time.ParseInLocation("2006-01-02", *startStr, time.Local)
	if err != nil {
		return fmt.Errorf("计划开始日期格式错误: %v", err)
	}

	end, err := time.ParseInLocation("2006-01-02", *endStr, time.Local)
	if err != nil {
		return fmt.Errorf("计划结束日期格式错误: %v", err)
	}

	return validatePlanDates(start, end)
}

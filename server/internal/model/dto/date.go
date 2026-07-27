package dto

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// ============================================
// 统一时间格式列表（全局共享）
// ============================================
var timeFormats = []string{
	"2006-01-02",                // 仅日期
	"2006-01-02T15:04:05",       // 日期+时间（无时区，前端 TDesign）
	"2006-01-02 15:04:05",       // 日期+时间（空格分隔）
	time.RFC3339,                // "2006-01-02T15:04:05Z07:00"
	"2006-01-02T15:04:05Z07:00", // 显式 RFC3339
}

func parseTimeString(s string) (time.Time, error) {
	for _, format := range timeFormats {
		if t, err := time.ParseInLocation(format, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("无法解析时间: %s", s)
}

// ============================================
// FlexibleTime 灵活时间类型（用于替代所有 *time.Time）
// ============================================

// FlexibleTime 灵活时间类型，支持多种 JSON 日期格式序列化/反序列化
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON 实现 JSON 反序列化，自动尝试所有支持的格式
func (ft *FlexibleTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" || string(b) == `""` {
		ft.Time = time.Time{}
		return nil
	}
	s := string(b[1 : len(b)-1])
	t, err := parseTimeString(s)
	if err != nil {
		return err
	}
	ft.Time = t
	return nil
}

// MarshalJSON 实现 JSON 序列化
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	if ft.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, ft.Time.Format("2006-01-02T15:04:05"))), nil
}

// Value 实现 driver.Valuer 接口（数据库写入）
func (ft FlexibleTime) Value() (driver.Value, error) {
	if ft.Time.IsZero() {
		return nil, nil
	}
	return ft.Time, nil
}

// Scan 实现 sql.Scanner 接口（数据库读取）
func (ft *FlexibleTime) Scan(value interface{}) error {
	if value == nil {
		ft.Time = time.Time{}
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		ft.Time = v
	case []byte:
		t, err := parseTimeString(string(v))
		if err != nil {
			return err
		}
		ft.Time = t
	case string:
		t, err := parseTimeString(v)
		if err != nil {
			return err
		}
		ft.Time = t
	default:
		return fmt.Errorf("不支持的日期类型: %T", value)
	}
	return nil
}

// ============================================
// Date 自定义日期类型（仅日期部分）
// ============================================

// Date 自定义日期类型，支持 "YYYY-MM-DD" 格式
type Date struct {
	time.Time
}

// UnmarshalJSON 实现 JSON 反序列化
func (d *Date) UnmarshalJSON(b []byte) error {
	if string(b) == "null" || string(b) == `""` {
		d.Time = time.Time{}
		return nil
	}

	s := string(b[1 : len(b)-1])

	// 优先纯日期格式，再尝试完整时间格式（取日期部分）
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		t, err = parseTimeString(s)
		if err != nil {
			return fmt.Errorf("无法解析日期: %v", err)
		}
	}

	d.Time = t
	return nil
}

// MarshalJSON 实现 JSON 序列化
func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, d.Time.Format("2006-01-02"))), nil
}

// Value 实现 driver.Valuer 接口
func (d Date) Value() (driver.Value, error) {
	if d.Time.IsZero() {
		return nil, nil
	}
	return d.Time, nil
}

// Scan 实现 sql.Scanner 接口
func (d *Date) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		d.Time = v
	case []byte:
		t, err := time.ParseInLocation("2006-01-02", string(v), time.Local)
		if err != nil {
			return err
		}
		d.Time = t
	case string:
		t, err := time.ParseInLocation("2006-01-02", v, time.Local)
		if err != nil {
			return err
		}
		d.Time = t
	default:
		return fmt.Errorf("不支持的日期类型: %T", value)
	}

	return nil
}

// Ptr 返回指向 Date 的指针
func (d Date) Ptr() *Date {
	return &d
}

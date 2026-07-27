package dto

import (
	"encoding/json"
	"errors"
	"time"
)

// CustomTime 自定义时间类型，支持字符串和 time.Time
type CustomTime struct {
	*time.Time
}

// UnmarshalJSON 实现自定义的 JSON 反序列化
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	// 如果是 null，设置为 nil
	if string(data) == "null" {
		ct.Time = nil
		return nil
	}

	// 尝试解析为字符串格式的时间
	var timeStr string
	if err := json.Unmarshal(data, &timeStr); err != nil {
		return err
	}

	// 如果字符串为空，设置为 nil
	if timeStr == "" {
		ct.Time = nil
		return nil
	}

	// 使用统一的 parseTimeString 解析
	t, err := parseTimeString(timeStr)
	if err != nil {
		return errors.New("无法解析时间格式: " + timeStr)
	}
	ct.Time = &t
	return nil
}

// UserDTO 用户请求 DTO
type UserDTO struct {
	Username string      `json:"username" binding:"required"`
	Password string      `json:"password" binding:"required"`
	RealName string      `json:"real_name"`
	Phone    string      `json:"phone"`
	Email    string      `json:"email"`
	Position string      `json:"position"`
	HireDate *CustomTime `json:"hire_date"`
	DeptID   int64       `json:"dept_id"`
	Status   int8        `json:"status"`
	RoleIDs  []int64     `json:"role_ids"` // 角色ID数组
}

// UserUpdateDTO 用户更新 DTO
type UserUpdateDTO struct {
	ID       int64       `json:"id" binding:"required"`
	Username string      `json:"username"`
	RealName string      `json:"real_name"`
	Phone    string      `json:"phone"`
	Email    string      `json:"email"`
	Position string      `json:"position"`
	HireDate *CustomTime `json:"hire_date"`
	DeptID   int64       `json:"dept_id"`
	Status   int8        `json:"status"`
	RoleIDs  []int64     `json:"role_ids"` // 角色ID数组
}

// LoginDTO 登录请求 DTO
type LoginDTO struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	CaptchaID  string `json:"captcha_id"`
	CaptchaVal string `json:"captcha_val"`
	KeepLogin  bool   `json:"keep_login"`
}

// DeptDTO 部门请求 DTO
type DeptDTO struct {
	DeptName string `json:"dept_name" binding:"required"`
	ParentID int64  `json:"parent_id"`
	Sort     int    `json:"sort"`
	Leader   string `json:"leader"`
	Phone    string `json:"phone"`
}

// ProfileUpdateDTO 个人信息更新 DTO
type ProfileUpdateDTO struct {
	ID       int64       `json:"id"` // 从 JWT 中获取，不需要前端传递
	RealName string      `json:"real_name"`
	Phone    string      `json:"phone"`
	Email    string      `json:"email"`
	Position string      `json:"position"`
	HireDate *CustomTime `json:"hire_date"`
}

// ChangePasswordDTO 修改密码 DTO
type ChangePasswordDTO struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// BatchStatusDTO 批量状态更新 DTO
type BatchStatusDTO struct {
	IDs    []int64 `json:"ids" binding:"required"`
	Status int8    `json:"status" binding:"required"`
}

// ResetPasswordDTO 管理员重置用户密码 DTO
type ResetPasswordDTO struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UserPageQueryDTO struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Status   *int8  `json:"status"`
	DeptID   *int64 `json:"dept_id"`
}

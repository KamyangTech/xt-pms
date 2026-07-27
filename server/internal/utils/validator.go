package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/KamyangTech/xt-pms/server/internal/model/constant"
	"github.com/KamyangTech/xt-pms/server/internal/model/dto"
	"github.com/go-playground/validator/v10"
)

type WorkOrderValidator struct{}

func NewWorkOrderValidator() *WorkOrderValidator {
	return &WorkOrderValidator{}
}

// ValidateCreateMoldWorkOrder 验证创建工单请求
func (v *WorkOrderValidator) ValidateCreateMoldWorkOrder(dto dto.MoldWorkOrderDTO) error {
	// 1. 订单类型验证
	validTypes := map[string]bool{
		constant.MoldOrderTypeNew:          true,
		constant.MoldOrderTypeRepair:       true,
		constant.MoldOrderTypeModification: true,
	}
	if !validTypes[dto.OrderType] {
		return errors.New("订单类型无效，必须为 NEW_MOLD(新模)、REPAIR(修模)或 MODIFICATION(改造)")
	}

	// 2. 根据订单类型进行不同验证
	if dto.OrderType == constant.MoldOrderTypeNew {
		// 新模工单验证
		if err := v.validateNewMoldWorkOrder(dto); err != nil {
			return err
		}
	} else {
		// 修模/改造工单验证
		if err := v.validateRepairMoldWorkOrder(dto); err != nil {
			return err
		}
	}

	// 3. 日期验证
	if dto.PlanStart != nil && dto.PlanEnd != nil {
		if dto.PlanEnd.Time.Before(dto.PlanStart.Time) {
			return errors.New("计划结束时间不能早于计划开始时间")
		}
	}

	// 4. 备注长度验证
	if len(dto.Remark) > 1000 {
		return errors.New("备注长度不能超过1000字符")
	}

	return nil
}

// validateNewMoldWorkOrder 验证新模工单
func (v *WorkOrderValidator) validateNewMoldWorkOrder(dto dto.MoldWorkOrderDTO) error {
	// 模具编号验证
	if strings.TrimSpace(dto.NewMoldCode) == "" {
		return errors.New("新模工单必须填写模具编号")
	}
	if len(dto.NewMoldCode) > 50 {
		return errors.New("模具编号长度不能超过50字符")
	}
	// 只允许字母、数字、下划线、中划线
	if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, dto.NewMoldCode); !matched {
		return errors.New("模具编号只能包含字母、数字、下划线和中划线")
	}

	// 产品名称验证
	if strings.TrimSpace(dto.NewMoldName) == "" {
		return errors.New("新模工单必须填写产品名称")
	}
	if len(dto.NewMoldName) > 100 {
		return errors.New("产品名称长度不能超过100字符")
	}

	// 客户ID验证
	if dto.CustomerID == 0 {
		return errors.New("新模工单必须选择客户")
	}

	return nil
}

// validateRepairMoldWorkOrder 验证修模/改造工单
func (v *WorkOrderValidator) validateRepairMoldWorkOrder(dto dto.MoldWorkOrderDTO) error {
	// 模具ID验证
	if dto.MoldID == nil || *dto.MoldID == 0 {
		return errors.New("修模/改造工单必须选择已有模具")
	}

	return nil
}

// ValidateUpdateMoldWorkOrder 验证更新工单请求
func (v *WorkOrderValidator) ValidateUpdateMoldWorkOrder(id int64, updates map[string]interface{}) error {
	if id <= 0 {
		return errors.New("工单ID无效")
	}

	if progress, ok := updates["progress"]; ok {
		if p, ok := progress.(int); !ok || p < 0 || p > 100 {
			return errors.New("进度必须在0-100之间")
		}
	}

	if status, ok := updates["status"]; ok {
		if s, ok := status.(int8); !ok || s < 1 || s > 10 {
			return errors.New("工单状态无效")
		}
	}

	if remark, ok := updates["remark"]; ok {
		if r, ok := remark.(string); ok && len(r) > 1000 {
			return errors.New("备注长度不能超过1000字符")
		}
	}

	return nil
}

// FormatValidationError 将 Gin/validator 校验错误转为中文友好提示
// 生产环境下 safeMessage 可能脱敏原始错误，通过此函数预先转为用户可理解的提示
func FormatValidationError(err error) string {
	// 尝试解析 validator.ValidationErrors
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		parts := make([]string, 0, len(ve))
		for _, e := range ve {
			field := e.Field()
			tag := e.Tag()
			msg := formatFieldError(field, tag)
			parts = append(parts, msg)
		}
		if len(parts) > 0 {
			return strings.Join(parts, "；")
		}
	}
	// 非 validator 错误（如 JSON 类型不匹配），返回简化提示
	return "请求参数格式不正确，请检查输入内容"
}

func formatFieldError(field, tag string) string {
	switch tag {
	case "required":
		return fmt.Sprintf("%s不能为空", field)
	case "min":
		return fmt.Sprintf("%s长度不足", field)
	case "max":
		return fmt.Sprintf("%s长度超出限制", field)
	case "email":
		return fmt.Sprintf("%s格式不正确（需为邮箱）", field)
	default:
		return fmt.Sprintf("%s校验失败(%s)", field, tag)
	}
}


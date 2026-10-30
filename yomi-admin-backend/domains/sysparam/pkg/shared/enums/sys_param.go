// Package enums 系统参数枚举定义
package enums

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Status 表示系统参数的启停状态
type Status string

const (
	// StatusEnabled 表示启用
	StatusEnabled Status = "ENABLED"
	// StatusDisabled 表示禁用：覆盖值失效，读取方回退默认值
	StatusDisabled Status = "DISABLED"
)

// IsValidStatus 判断状态是否为合法枚举
func IsValidStatus(status Status) bool {
	return status == StatusEnabled || status == StatusDisabled
}

// NormalizeStatus 归一化状态字面量：去空白并转大写。
// 请求体（JSON）与查询参数（query）共用这一个口径，空值原样返回由调用方判断。
func NormalizeStatus(raw string) Status {
	return Status(strings.ToUpper(strings.TrimSpace(raw)))
}

// UnmarshalJSON 实现 json.Unmarshaler，反序列化时归一化并校验
func (s *Status) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	value := NormalizeStatus(raw)
	if value == "" {
		*s = ""
		return nil
	}
	if !IsValidStatus(value) {
		return fmt.Errorf("invalid Status: %s", raw)
	}

	*s = value
	return nil
}

// ParamValueType 表示参数值的类型，决定写入校验与读取转换方式
type ParamValueType string

const (
	// ParamTypeBool 布尔值，取值仅 true / false
	ParamTypeBool ParamValueType = "BOOL"
	// ParamTypeInt 64 位整数
	ParamTypeInt ParamValueType = "INT"
	// ParamTypeString 任意字符串
	ParamTypeString ParamValueType = "STRING"
	// ParamTypeJSON 合法 JSON 文档
	ParamTypeJSON ParamValueType = "JSON"
)

// IsValidParamValueType 判断值类型是否为合法枚举
func IsValidParamValueType(valueType ParamValueType) bool {
	switch valueType {
	case ParamTypeBool, ParamTypeInt, ParamTypeString, ParamTypeJSON:
		return true
	default:
		return false
	}
}

// NormalizeParamValueType 归一化值类型字面量：去空白并转大写。
// 请求体（JSON）与查询参数（query）共用这一个口径，空值原样返回由调用方判断。
func NormalizeParamValueType(raw string) ParamValueType {
	return ParamValueType(strings.ToUpper(strings.TrimSpace(raw)))
}

// UnmarshalJSON 实现 json.Unmarshaler，反序列化时归一化并校验
func (t *ParamValueType) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	value := NormalizeParamValueType(raw)
	if value == "" {
		*t = ""
		return nil
	}
	if !IsValidParamValueType(value) {
		return fmt.Errorf("invalid ParamValueType: %s", raw)
	}

	*t = value
	return nil
}

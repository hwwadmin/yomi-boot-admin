// Package enums 用户域枚举定义
package enums

import (
	"encoding/json"
	"fmt"
	"strings"
)

// UserStatus 表示用户账号状态
type UserStatus string

const (
	// UserStatusEnabled 表示账号正常
	UserStatusEnabled UserStatus = "ENABLED"
	// UserStatusDisabled 表示账号已禁用
	UserStatusDisabled UserStatus = "DISABLED"
)

// IsValidUserStatus 判断用户状态是否为合法枚举
func IsValidUserStatus(status UserStatus) bool {
	return status == UserStatusEnabled || status == UserStatusDisabled
}

// UnmarshalJSON 实现 json.Unmarshaler，反序列化时归一化并校验
func (s *UserStatus) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	value := UserStatus(strings.ToUpper(strings.TrimSpace(raw)))
	if value == "" {
		*s = ""
		return nil
	}
	if !IsValidUserStatus(value) {
		return fmt.Errorf("invalid UserStatus: %s", raw)
	}

	*s = value
	return nil
}

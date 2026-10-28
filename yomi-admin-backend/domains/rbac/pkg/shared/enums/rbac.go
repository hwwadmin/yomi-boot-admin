// Package enums RBAC 域枚举定义
package enums

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Status 表示角色与权限的启停状态
type Status string

const (
	// StatusEnabled 表示启用
	StatusEnabled Status = "ENABLED"
	// StatusDisabled 表示禁用
	StatusDisabled Status = "DISABLED"
)

// IsValidStatus 判断状态是否为合法枚举
func IsValidStatus(status Status) bool {
	return status == StatusEnabled || status == StatusDisabled
}

// UnmarshalJSON 实现 json.Unmarshaler，反序列化时归一化并校验
func (s *Status) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	value := Status(strings.ToUpper(strings.TrimSpace(raw)))
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

// PermissionType 表示权限类型
type PermissionType string

const (
	// PermissionTypeMenu 表示菜单权限
	PermissionTypeMenu PermissionType = "MENU"
	// PermissionTypeAPI 表示接口权限
	PermissionTypeAPI PermissionType = "API"
)

// IsValidPermissionType 判断权限类型是否为合法枚举
func IsValidPermissionType(permissionType PermissionType) bool {
	return permissionType == PermissionTypeMenu || permissionType == PermissionTypeAPI
}

// UnmarshalJSON 实现 json.Unmarshaler，反序列化时归一化并校验
func (t *PermissionType) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	value := PermissionType(strings.ToUpper(strings.TrimSpace(raw)))
	if value == "" {
		*t = ""
		return nil
	}
	if !IsValidPermissionType(value) {
		return fmt.Errorf("invalid PermissionType: %s", raw)
	}

	*t = value
	return nil
}

// RoleCodeSuperAdmin 是内置超级管理员角色编码
const RoleCodeSuperAdmin = "super_admin"

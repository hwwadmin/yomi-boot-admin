// Package enums 用户域枚举定义
package enums

// UserStatus 表示用户账号状态
type UserStatus int8

const (
	// UserStatusDisabled 表示账号已禁用
	UserStatusDisabled UserStatus = 0
	// UserStatusEnabled 表示账号正常
	UserStatusEnabled UserStatus = 1
)

// Valid 判断状态是否合法
func (s UserStatus) Valid() bool {
	return s == UserStatusDisabled || s == UserStatusEnabled
}

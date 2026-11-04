// Package service_contract 用户域服务契约
package service_contract

import "github.com/hwwadmin/yomi-boot/component/core/errors"

var (
	// ErrInvalidCredential 登录凭证不合法（不区分用户不存在与密码错误）
	ErrInvalidCredential = errors.NewBiz("ADMIN_INVALID_CREDENTIAL", "用户名或密码错误")
	// ErrUserDisabled 账号已被禁用
	ErrUserDisabled = errors.NewBiz("ADMIN_USER_DISABLED", "账号已被禁用")
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.NewBiz("ADMIN_USER_NOT_FOUND", "用户不存在")
	// ErrUsernameExists 用户名已存在
	ErrUsernameExists = errors.NewBiz("ADMIN_USERNAME_EXISTS", "用户名已存在")
	// ErrOldPasswordMismatch 修改密码时原密码不正确
	ErrOldPasswordMismatch = errors.NewBiz("ADMIN_OLD_PASSWORD_MISMATCH", "原密码不正确")
	// ErrPasswordUnchanged 新密码与原密码相同
	ErrPasswordUnchanged = errors.NewBiz("ADMIN_PASSWORD_UNCHANGED", "新密码不能与原密码相同")
	// ErrInvalidInput 用户参数不合法
	ErrInvalidInput = errors.NewBiz("ADMIN_INVALID_INPUT", "用户参数不合法")
)

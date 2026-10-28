// Package service_contract 用户域服务契约
package service_contract

import "gitee.com/hq_project/yomi-boot/component/core/errors"

var (
	// ErrInvalidCredential 登录凭证不合法（不区分用户不存在与密码错误）
	ErrInvalidCredential = errors.NewBiz("ADMIN_INVALID_CREDENTIAL", "用户名或密码错误")
	// ErrUserDisabled 账号已被禁用
	ErrUserDisabled = errors.NewBiz("ADMIN_USER_DISABLED", "账号已被禁用")
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.NewBiz("ADMIN_USER_NOT_FOUND", "用户不存在")
	// ErrUsernameExists 用户名已存在
	ErrUsernameExists = errors.NewBiz("ADMIN_USERNAME_EXISTS", "用户名已存在")
	// ErrInvalidInput 用户参数不合法
	ErrInvalidInput = errors.NewBiz("ADMIN_INVALID_INPUT", "用户参数不合法")
	// ErrCannotOperateSelf 不能操作当前登录账号
	ErrCannotOperateSelf = errors.NewBiz("ADMIN_CANNOT_OPERATE_SELF", "不能操作当前登录账号")
)

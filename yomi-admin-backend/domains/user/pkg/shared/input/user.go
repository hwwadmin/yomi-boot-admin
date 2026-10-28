// Package input 用户域服务入参
package input

import "yomi-admin-backend/domains/user/pkg/shared/enums"

// AuthenticateInput 登录鉴权入参
type AuthenticateInput struct {
	Username string
	Password string
}

// PageUsersInput 用户分页查询入参
type PageUsersInput struct {
	Keyword  string
	Page     int
	PageSize int
}

// CreateUserInput 创建用户入参
type CreateUserInput struct {
	Username   string
	Password   string
	Nickname   string
	Email      string
	Phone      string
	Status     enums.UserStatus
	OperatorID int64
}

// UpdateUserInput 更新用户入参；用户名不可变
type UpdateUserInput struct {
	ID         int64
	Nickname   string
	Email      string
	Phone      string
	Status     enums.UserStatus
	OperatorID int64
}

// DeleteUserInput 删除用户入参
type DeleteUserInput struct {
	ID         int64
	OperatorID int64
}

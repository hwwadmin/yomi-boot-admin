// Package output 用户域服务出参
package output

import "yomi-admin-backend/domains/user/pkg/shared/entity"

// AuthenticateOutput 登录鉴权出参（不含 Token）
type AuthenticateOutput struct {
	User *entity.User
}

// PageUsersOutput 用户分页出参
type PageUsersOutput struct {
	Items []*entity.User
	Total int64
}

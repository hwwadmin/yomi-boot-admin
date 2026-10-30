// Package auth 认证相关编排
package auth

import (
	"context"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"
	"yomi-admin-backend/domains/rbac/service_contract"
	"yomi-admin-backend/domains/user/pkg/shared/entity"
)

// MenuView 是登录态返回的菜单视图
type MenuView struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	UIPath string `json:"uiPath"`
}

// RoleView 是登录态返回的角色摘要
type RoleView struct {
	Code   string       `json:"code"`
	Name   string       `json:"name"`
	Status enums.Status `json:"status"`
}

// IdentityView 是登录与当前用户接口共用的身份视图
type IdentityView struct {
	Username    string     `json:"username"`
	Nickname    string     `json:"nickname"`
	Email       string     `json:"email"`
	Phone       string     `json:"phone"`
	Roles       []RoleView `json:"roles"`
	Permissions []string   `json:"permissions"`
	Menus       []MenuView `json:"menus"`
}

// buildIdentityView 聚合用户资料、角色、权限编码与菜单
func buildIdentityView(ctx context.Context, user *entity.User, auth service_contract.RbacAuthService) (*IdentityView, error) {
	permissions, err := auth.ListPermissionCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if permissions == nil {
		permissions = []string{}
	}

	roles, err := auth.ListRoles(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	roleViews := make([]RoleView, 0, len(roles))
	for _, role := range roles {
		roleViews = append(roleViews, RoleView{
			Code:   role.Code,
			Name:   role.Name,
			Status: role.Status,
		})
	}

	menus, err := auth.ListMenus(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	menuViews := make([]MenuView, 0, len(menus))
	for _, menu := range menus {
		menuViews = append(menuViews, MenuView{
			Code:   menu.Code,
			Name:   menu.Name,
			UIPath: menu.UIPath,
		})
	}

	return &IdentityView{
		Username:    user.Username,
		Nickname:    user.Nickname,
		Email:       user.Email,
		Phone:       user.Phone,
		Roles:       roleViews,
		Permissions: permissions,
		Menus:       menuViews,
	}, nil
}

// Package auth 认证相关编排
package auth

import (
	"context"
	"yomi-admin-backend/domains/rbac/service_contract"
	"yomi-admin-backend/domains/user/pkg/shared/entity"
)

// MenuView 是登录态返回的菜单视图
type MenuView struct {
	Code   string `json:"code"`
	Name   string `json:"name"`
	UIPath string `json:"uiPath"`
}

// IdentityView 是登录与当前用户接口共用的身份视图
type IdentityView struct {
	Username    string     `json:"username"`
	Nickname    string     `json:"nickname"`
	Permissions []string   `json:"permissions"`
	Menus       []MenuView `json:"menus"`
}

// buildIdentityView 聚合用户资料、权限编码与菜单
func buildIdentityView(ctx context.Context, user *entity.User, auth service_contract.RbacAuthService) (*IdentityView, error) {
	permissions, err := auth.ListPermissionCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if permissions == nil {
		permissions = []string{}
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
		Permissions: permissions,
		Menus:       menuViews,
	}, nil
}

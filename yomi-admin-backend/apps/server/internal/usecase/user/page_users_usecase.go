// Package user 用户管理编排
package user

import (
	"context"
	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
	userinput "yomi-admin-backend/domains/user/pkg/shared/input"
	"yomi-admin-backend/domains/user/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// PageUsersInput HTTP 分页入参
type PageUsersInput struct {
	dto.PageInput
	Keyword string `json:"keyword"`
}

// RoleBrief 用户列表中的角色摘要
type RoleBrief struct {
	ID     dto.JSONInt64    `json:"id"`
	Code   string           `json:"code"`
	Name   string           `json:"name"`
	Status rbacenums.Status `json:"status"`
}

// UserItem 用户列表项
type UserItem struct {
	ID        dto.JSONInt64    `json:"id"`
	Username  string           `json:"username"`
	Nickname  string           `json:"nickname"`
	Email     string           `json:"email"`
	Phone     string           `json:"phone"`
	Status    enums.UserStatus `json:"status"`
	Roles     []RoleBrief      `json:"roles"`
	CreatedAt int64            `json:"createdAt"` // Unix 毫秒时间戳，由前端按客户端时区展示
}

// PageUsersUsecase 用户分页编排
type PageUsersUsecase struct {
	users service_contract.UserService
	rbac  rbaccontract.RbacAdminService
}

// NewPageUsersUsecase 创建用户分页用例
func NewPageUsersUsecase(users service_contract.UserService, rbac rbaccontract.RbacAdminService) *PageUsersUsecase {
	return &PageUsersUsecase{users: users, rbac: rbac}
}

// Execute 分页查询用户
func (uc *PageUsersUsecase) Execute(ctx context.Context, in *PageUsersInput) (dto.PageOutput[UserItem], error) {
	in.ValiPage()

	out, err := uc.users.Page(ctx, &userinput.PageUsersInput{
		Keyword:  in.Keyword,
		Page:     in.Page,
		PageSize: in.PageSize,
	})
	if err != nil {
		return dto.PageOutput[UserItem]{}, err
	}

	userIDs := make([]int64, 0, len(out.Items))
	for _, user := range out.Items {
		userIDs = append(userIDs, user.ID)
	}
	rolesByUser, err := uc.rbac.ListRolesByUserIDs(ctx, userIDs)
	if err != nil {
		return dto.PageOutput[UserItem]{}, err
	}

	items := make([]UserItem, 0, len(out.Items))
	for _, user := range out.Items {
		roles := make([]RoleBrief, 0)
		for _, role := range rolesByUser[user.ID] {
			roles = append(roles, RoleBrief{
				ID:     dto.JSONInt64(role.ID),
				Code:   role.Code,
				Name:   role.Name,
				Status: role.Status,
			})
		}
		items = append(items, UserItem{
			ID:        dto.JSONInt64(user.ID),
			Username:  user.Username,
			Nickname:  user.Nickname,
			Email:     user.Email,
			Phone:     user.Phone,
			Status:    user.Status,
			Roles:     roles,
			CreatedAt: user.CreateTime,
		})
	}

	return dto.NewPageOutput(items, out.Total), nil
}

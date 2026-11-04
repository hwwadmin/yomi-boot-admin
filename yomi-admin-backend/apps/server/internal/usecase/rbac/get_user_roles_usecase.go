package rbac

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
	usercontract "yomi-admin-backend/domains/user/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// UserRolesOutput 用户角色绑定视图
type UserRolesOutput struct {
	Roles   []RoleItem         `json:"roles"`
	RoleIDs dto.JSONInt64Slice `json:"roleIds"`
}

// GetUserRolesUsecase 用户角色查询编排
type GetUserRolesUsecase struct {
	users usercontract.UserService
	rbac  rbaccontract.RbacAdminService
}

// NewGetUserRolesUsecase 创建用户角色查询用例
func NewGetUserRolesUsecase(users usercontract.UserService, rbac rbaccontract.RbacAdminService) *GetUserRolesUsecase {
	return &GetUserRolesUsecase{users: users, rbac: rbac}
}

// Execute 查询用户角色绑定视图
func (uc *GetUserRolesUsecase) Execute(ctx context.Context, userID int64) (*UserRolesOutput, error) {
	if userID <= 0 {
		return nil, errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	if _, err := uc.users.GetByID(ctx, userID); err != nil {
		return nil, err
	}

	out, err := uc.rbac.ListUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]RoleItem, 0, len(out.Roles))
	for _, role := range out.Roles {
		roles = append(roles, toRoleItem(role))
	}
	return &UserRolesOutput{
		Roles:   roles,
		RoleIDs: support.NonNilIDs(dto.NewJSONInt64Slice(out.RoleIDs)),
	}, nil
}

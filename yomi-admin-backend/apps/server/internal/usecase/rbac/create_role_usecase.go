package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
)

// CreateRoleInput HTTP 创建角色入参
type CreateRoleInput struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Remark     string `json:"remark"`
	OperatorID int64  `json:"-"`
}

// CreateRoleUsecase 创建角色编排
type CreateRoleUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewCreateRoleUsecase 创建角色用例
func NewCreateRoleUsecase(rbac rbaccontract.RbacAdminService) *CreateRoleUsecase {
	return &CreateRoleUsecase{rbac: rbac}
}

// Execute 创建角色
func (uc *CreateRoleUsecase) Execute(ctx context.Context, in *CreateRoleInput) (*RoleItem, error) {
	role, err := uc.rbac.CreateRole(ctx, &rbacinput.CreateRoleInput{
		Code:       in.Code,
		Name:       in.Name,
		Remark:     in.Remark,
		OperatorID: in.OperatorID,
	})
	if err != nil {
		return nil, err
	}
	item := toRoleItem(role)
	return &item, nil
}

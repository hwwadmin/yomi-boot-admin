package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
)

// CreatePermissionGroupInput HTTP 创建权限组入参
type CreatePermissionGroupInput struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Remark     string `json:"remark"`
	OperatorID int64  `json:"-"`
}

// CreatePermissionGroupUsecase 创建权限组编排
type CreatePermissionGroupUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewCreatePermissionGroupUsecase 创建权限组用例
func NewCreatePermissionGroupUsecase(rbac rbaccontract.RbacAdminService) *CreatePermissionGroupUsecase {
	return &CreatePermissionGroupUsecase{rbac: rbac}
}

// Execute 创建权限组
func (uc *CreatePermissionGroupUsecase) Execute(ctx context.Context, in *CreatePermissionGroupInput) (*PermissionGroupItem, error) {
	group, err := uc.rbac.CreatePermissionGroup(ctx, &rbacinput.CreatePermissionGroupInput{
		Code:       in.Code,
		Name:       in.Name,
		Remark:     in.Remark,
		OperatorID: in.OperatorID,
	})
	if err != nil {
		return nil, err
	}
	item := toPermissionGroupItem(group)
	return &item, nil
}

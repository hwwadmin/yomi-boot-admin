package rbac

import (
	"context"

	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"
)

// CreatePermissionInput HTTP 创建权限入参
type CreatePermissionInput struct {
	Code       string                   `json:"code"`
	Name       string                   `json:"name"`
	Type       rbacenums.PermissionType `json:"type"`
	UIPath     string                   `json:"uiPath"`
	HTTPMethod string                   `json:"httpMethod"`
	HTTPPath   string                   `json:"httpPath"`
	Remark     string                   `json:"remark"`
	OperatorID int64                    `json:"-"`
}

// CreatePermissionUsecase 创建权限编排
type CreatePermissionUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewCreatePermissionUsecase 创建权限用例
func NewCreatePermissionUsecase(rbac rbaccontract.RbacAdminService) *CreatePermissionUsecase {
	return &CreatePermissionUsecase{rbac: rbac}
}

// Execute 创建权限
func (uc *CreatePermissionUsecase) Execute(ctx context.Context, in *CreatePermissionInput) (*PermissionItem, error) {
	permission, err := uc.rbac.CreatePermission(ctx, &rbacinput.CreatePermissionInput{
		Code:       in.Code,
		Name:       in.Name,
		Type:       in.Type,
		UIPath:     in.UIPath,
		HTTPMethod: in.HTTPMethod,
		HTTPPath:   in.HTTPPath,
		Remark:     in.Remark,
		OperatorID: in.OperatorID,
	})
	if err != nil {
		return nil, err
	}
	item := toPermissionItem(permission)
	return &item, nil
}

package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// UpdatePermissionInput HTTP 更新权限入参；编码与类型不可变
type UpdatePermissionInput struct {
	ID         dto.JSONInt64 `json:"id"`
	Name       string        `json:"name"`
	UIPath     string        `json:"uiPath"`
	HTTPMethod string        `json:"httpMethod"`
	HTTPPath   string        `json:"httpPath"`
	Remark     string        `json:"remark"`
	OperatorID int64         `json:"-"`
}

// UpdatePermissionUsecase 更新权限编排
type UpdatePermissionUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewUpdatePermissionUsecase 创建更新权限用例
func NewUpdatePermissionUsecase(rbac rbaccontract.RbacAdminService) *UpdatePermissionUsecase {
	return &UpdatePermissionUsecase{rbac: rbac}
}

// Execute 更新权限
func (uc *UpdatePermissionUsecase) Execute(ctx context.Context, in *UpdatePermissionInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.rbac.UpdatePermission(ctx, &rbacinput.UpdatePermissionInput{
		ID:         id,
		Name:       in.Name,
		UIPath:     in.UIPath,
		HTTPMethod: in.HTTPMethod,
		HTTPPath:   in.HTTPPath,
		Remark:     in.Remark,
		OperatorID: in.OperatorID,
	})
}

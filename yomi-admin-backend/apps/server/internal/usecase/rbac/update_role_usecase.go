package rbac

import (
	"context"

	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// UpdateRoleInput HTTP 更新角色入参；编码不可变，启停走独立接口
type UpdateRoleInput struct {
	ID         dto.JSONInt64 `json:"id"`
	Name       string        `json:"name"`
	Remark     string        `json:"remark"`
	OperatorID int64         `json:"-"`
}

// UpdateRoleUsecase 更新角色编排
type UpdateRoleUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewUpdateRoleUsecase 创建更新角色用例
func NewUpdateRoleUsecase(rbac rbaccontract.RbacAdminService) *UpdateRoleUsecase {
	return &UpdateRoleUsecase{rbac: rbac}
}

// Execute 更新角色
func (uc *UpdateRoleUsecase) Execute(ctx context.Context, in *UpdateRoleInput) error {
	id := in.ID.Int64()
	if id <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return uc.rbac.UpdateRole(ctx, &rbacinput.UpdateRoleInput{
		ID:         id,
		Name:       in.Name,
		Remark:     in.Remark,
		OperatorID: in.OperatorID,
	})
}

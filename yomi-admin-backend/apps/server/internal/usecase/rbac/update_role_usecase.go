package rbac

import (
	"context"
	"strings"
	"unicode/utf8"
	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

// UpdateRoleInput HTTP 更新角色入参；编码不可变
type UpdateRoleInput struct {
	Name       string           `json:"name"`
	Remark     string           `json:"remark"`
	Status     rbacenums.Status `json:"status"`
	OperatorID int64            `json:"-"`
}

// Validate 校验入参
func (i *UpdateRoleInput) Validate() error {
	i.Name = strings.TrimSpace(i.Name)
	i.Remark = strings.TrimSpace(i.Remark)

	if i.Name == "" || utf8.RuneCountInString(i.Name) > maxRoleNameRunes {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	if !rbacenums.IsValidStatus(i.Status) {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	if utf8.RuneCountInString(i.Remark) > maxRemarkRunes {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return nil
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
func (uc *UpdateRoleUsecase) Execute(ctx context.Context, id int64, in *UpdateRoleInput) error {
	if id <= 0 {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	if err := in.Validate(); err != nil {
		return err
	}
	return uc.rbac.UpdateRole(ctx, &rbacinput.UpdateRoleInput{
		ID:         id,
		Name:       in.Name,
		Remark:     in.Remark,
		Status:     in.Status,
		OperatorID: in.OperatorID,
	})
}

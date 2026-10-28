package rbac

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
)

const (
	maxRoleCodeLength = 32
	maxRoleNameRunes  = 32
	maxRemarkRunes    = 255
)

// roleCodePattern 约束角色编码为小写字母开头的下划线组合
var roleCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,31}$`)

// CreateRoleInput HTTP 创建角色入参
type CreateRoleInput struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Remark     string `json:"remark"`
	OperatorID int64  `json:"-"`
}

// Validate 校验入参
func (i *CreateRoleInput) Validate() error {
	i.Code = strings.TrimSpace(i.Code)
	i.Name = strings.TrimSpace(i.Name)
	i.Remark = strings.TrimSpace(i.Remark)

	if len(i.Code) > maxRoleCodeLength || !roleCodePattern.MatchString(i.Code) {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	if i.Name == "" || utf8.RuneCountInString(i.Name) > maxRoleNameRunes {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	if utf8.RuneCountInString(i.Remark) > maxRemarkRunes {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return nil
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
	if err := in.Validate(); err != nil {
		return nil, err
	}

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

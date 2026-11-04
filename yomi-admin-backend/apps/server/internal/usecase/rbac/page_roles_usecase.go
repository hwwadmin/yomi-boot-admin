package rbac

import (
	"context"

	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"
)

// PageRolesInput HTTP 角色分页入参
type PageRolesInput struct {
	dto.PageInput
	Keyword string           `json:"keyword"`
	Status  rbacenums.Status `json:"status"`
}

// RoleItem 角色列表项
type RoleItem struct {
	ID         dto.JSONInt64    `json:"id"`
	Code       string           `json:"code"`
	Name       string           `json:"name"`
	Builtin    bool             `json:"builtin"`
	Status     rbacenums.Status `json:"status"`
	Remark     string           `json:"remark"`
	CreateTime int64            `json:"createTime"`
	UpdateTime int64            `json:"updateTime"`
}

// PageRolesUsecase 角色分页编排
type PageRolesUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewPageRolesUsecase 创建角色分页用例
func NewPageRolesUsecase(rbac rbaccontract.RbacAdminService) *PageRolesUsecase {
	return &PageRolesUsecase{rbac: rbac}
}

// Execute 分页查询角色
func (uc *PageRolesUsecase) Execute(ctx context.Context, in *PageRolesInput) (dto.PageOutput[RoleItem], error) {
	in.ValiPage()

	out, err := uc.rbac.PageRoles(ctx, &rbacinput.PageRolesInput{
		Keyword:  in.Keyword,
		Status:   in.Status,
		Page:     in.Page,
		PageSize: in.PageSize,
	})
	if err != nil {
		return dto.PageOutput[RoleItem]{}, err
	}

	items := make([]RoleItem, 0, len(out.Items))
	for _, role := range out.Items {
		items = append(items, toRoleItem(role))
	}
	return dto.NewPageOutput(items, out.Total), nil
}

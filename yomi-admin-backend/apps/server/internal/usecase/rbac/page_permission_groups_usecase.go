package rbac

import (
	"context"

	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// PagePermissionGroupsInput HTTP 权限组分页入参
type PagePermissionGroupsInput struct {
	dto.PageInput
	Keyword string           `json:"keyword"`
	Status  rbacenums.Status `json:"status"`
}

// PermissionGroupItem 权限组列表项
type PermissionGroupItem struct {
	ID         dto.JSONInt64    `json:"id"`
	Code       string           `json:"code"`
	Name       string           `json:"name"`
	Status     rbacenums.Status `json:"status"`
	Remark     string           `json:"remark"`
	CreateTime int64            `json:"createTime"`
	UpdateTime int64            `json:"updateTime"`
}

// PagePermissionGroupsUsecase 权限组分页编排
type PagePermissionGroupsUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewPagePermissionGroupsUsecase 创建权限组分页用例
func NewPagePermissionGroupsUsecase(rbac rbaccontract.RbacAdminService) *PagePermissionGroupsUsecase {
	return &PagePermissionGroupsUsecase{rbac: rbac}
}

// Execute 分页查询权限组
func (uc *PagePermissionGroupsUsecase) Execute(ctx context.Context, in *PagePermissionGroupsInput) (dto.PageOutput[PermissionGroupItem], error) {
	in.ValiPage()

	out, err := uc.rbac.PagePermissionGroups(ctx, &rbacinput.PagePermissionGroupsInput{
		Keyword:  in.Keyword,
		Status:   in.Status,
		Page:     in.Page,
		PageSize: in.PageSize,
	})
	if err != nil {
		return dto.PageOutput[PermissionGroupItem]{}, err
	}

	items := make([]PermissionGroupItem, 0, len(out.Items))
	for _, group := range out.Items {
		items = append(items, toPermissionGroupItem(group))
	}
	return dto.NewPageOutput(items, out.Total), nil
}

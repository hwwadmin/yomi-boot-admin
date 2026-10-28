package rbac

import (
	"context"
	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	rbacinput "yomi-admin-backend/domains/rbac/pkg/shared/input"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// PagePermissionsInput HTTP 权限分页入参
type PagePermissionsInput struct {
	dto.PageInput
	Keyword string                   `json:"keyword"`
	Type    rbacenums.PermissionType `json:"type"`
	Status  rbacenums.Status         `json:"status"`
}

// Validate 校验入参
func (i *PagePermissionsInput) Validate() error {
	i.ValiPage()
	if i.Type != "" && !rbacenums.IsValidPermissionType(i.Type) {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	if i.Status != "" && !rbacenums.IsValidStatus(i.Status) {
		return errors.Sentinel(rbaccontract.ErrInvalidInput)
	}
	return nil
}

// PermissionItem 权限列表项
type PermissionItem struct {
	ID         dto.JSONInt64            `json:"id"`
	Code       string                   `json:"code"`
	Name       string                   `json:"name"`
	Type       rbacenums.PermissionType `json:"type"`
	UIPath     string                   `json:"uiPath,omitempty"`
	HTTPMethod string                   `json:"httpMethod,omitempty"`
	HTTPPath   string                   `json:"httpPath,omitempty"`
	Status     rbacenums.Status         `json:"status"`
	Remark     string                   `json:"remark"`
	CreateTime int64                    `json:"createTime"`
	UpdateTime int64                    `json:"updateTime"`
}

// PagePermissionsUsecase 权限分页编排
type PagePermissionsUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewPagePermissionsUsecase 创建权限分页用例
func NewPagePermissionsUsecase(rbac rbaccontract.RbacAdminService) *PagePermissionsUsecase {
	return &PagePermissionsUsecase{rbac: rbac}
}

// Execute 分页查询权限目录
func (uc *PagePermissionsUsecase) Execute(ctx context.Context, in *PagePermissionsInput) (*dto.PageOutput[PermissionItem], error) {
	if err := in.Validate(); err != nil {
		return nil, err
	}

	out, err := uc.rbac.PagePermissions(ctx, &rbacinput.PagePermissionsInput{
		Keyword:  in.Keyword,
		Type:     in.Type,
		Status:   in.Status,
		Page:     in.Page,
		PageSize: in.PageSize,
	})
	if err != nil {
		return nil, err
	}

	items := make([]PermissionItem, 0, len(out.Items))
	for _, permission := range out.Items {
		items = append(items, PermissionItem{
			ID:         dto.JSONInt64(permission.ID),
			Code:       permission.Code,
			Name:       permission.Name,
			Type:       permission.Type,
			UIPath:     permission.UIPath,
			HTTPMethod: permission.HTTPMethod,
			HTTPPath:   permission.HTTPPath,
			Status:     permission.Status,
			Remark:     permission.Remark,
			CreateTime: permission.CreateTime,
			UpdateTime: permission.UpdateTime,
		})
	}
	return &dto.PageOutput[PermissionItem]{Items: items, Total: out.Total}, nil
}

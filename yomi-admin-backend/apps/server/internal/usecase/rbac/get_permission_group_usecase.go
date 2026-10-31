package rbac

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	rbaccontract "yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
)

// PermissionGroupDetailOutput 权限组详情（含已绑权限 ID）
type PermissionGroupDetailOutput struct {
	PermissionGroupItem
	PermissionIDs dto.JSONInt64Slice `json:"permissionIds"`
}

// GetPermissionGroupUsecase 权限组详情编排
type GetPermissionGroupUsecase struct {
	rbac rbaccontract.RbacAdminService
}

// NewGetPermissionGroupUsecase 创建权限组详情用例
func NewGetPermissionGroupUsecase(rbac rbaccontract.RbacAdminService) *GetPermissionGroupUsecase {
	return &GetPermissionGroupUsecase{rbac: rbac}
}

// Execute 查询权限组详情
func (uc *GetPermissionGroupUsecase) Execute(ctx context.Context, id int64) (*PermissionGroupDetailOutput, error) {
	if id <= 0 {
		return nil, errors.Sentinel(rbaccontract.ErrInvalidInput)
	}

	detail, err := uc.rbac.GetPermissionGroup(ctx, id)
	if err != nil {
		return nil, err
	}
	return &PermissionGroupDetailOutput{
		PermissionGroupItem: toPermissionGroupItem(detail.Group),
		PermissionIDs:       support.NonNilIDs(dto.NewJSONInt64Slice(detail.PermissionIDs)),
	}, nil
}

// toPermissionGroupItem 将权限组实体转换为 HTTP 视图
func toPermissionGroupItem(group *entity.PermissionGroup) PermissionGroupItem {
	return PermissionGroupItem{
		ID:         dto.JSONInt64(group.ID),
		Code:       group.Code,
		Name:       group.Name,
		Status:     group.Status,
		Remark:     group.Remark,
		CreateTime: group.CreateTime,
		UpdateTime: group.UpdateTime,
	}
}

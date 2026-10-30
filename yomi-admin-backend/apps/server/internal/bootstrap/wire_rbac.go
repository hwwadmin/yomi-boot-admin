package bootstrap

import (
	apphandler "yomi-admin-backend/apps/server/internal/handler"
	rbacusecase "yomi-admin-backend/apps/server/internal/usecase/rbac"

	"gitee.com/hq_project/yomi-boot/component/core/router"
)

// newRbacRegistrars 组装 RBAC 模块：角色、权限目录、权限组与鉴权缓存维护
func newRbacRegistrars(c container) []router.Registrar {
	return []router.Registrar{
		apphandler.NewRoleHandler(apphandler.RoleHandlerDeps{
			Page:       rbacusecase.NewPageRolesUsecase(c.rbac.Admin),
			Get:        rbacusecase.NewGetRoleUsecase(c.rbac.Admin),
			Create:     rbacusecase.NewCreateRoleUsecase(c.rbac.Admin),
			Update:     rbacusecase.NewUpdateRoleUsecase(c.rbac.Admin),
			Delete:     rbacusecase.NewDeleteRoleUsecase(c.rbac.Admin, c.tx),
			SetStatus:  rbacusecase.NewSetRoleStatusUsecase(c.rbac.Admin),
			BindGroups: rbacusecase.NewBindRoleGroupsUsecase(c.rbac.Admin, c.tx),
		}),
		apphandler.NewPermissionHandler(apphandler.PermissionHandlerDeps{
			Page:      rbacusecase.NewPagePermissionsUsecase(c.rbac.Admin),
			Create:    rbacusecase.NewCreatePermissionUsecase(c.rbac.Admin),
			Update:    rbacusecase.NewUpdatePermissionUsecase(c.rbac.Admin),
			Delete:    rbacusecase.NewDeletePermissionUsecase(c.rbac.Admin, c.tx),
			SetStatus: rbacusecase.NewSetPermissionStatusUsecase(c.rbac.Admin),
		}),
		apphandler.NewPermissionGroupHandler(apphandler.PermissionGroupHandlerDeps{
			Page:            rbacusecase.NewPagePermissionGroupsUsecase(c.rbac.Admin),
			Get:             rbacusecase.NewGetPermissionGroupUsecase(c.rbac.Admin),
			Create:          rbacusecase.NewCreatePermissionGroupUsecase(c.rbac.Admin),
			Update:          rbacusecase.NewUpdatePermissionGroupUsecase(c.rbac.Admin),
			Delete:          rbacusecase.NewDeletePermissionGroupUsecase(c.rbac.Admin, c.tx),
			SetStatus:       rbacusecase.NewSetPermissionGroupStatusUsecase(c.rbac.Admin),
			BindPermissions: rbacusecase.NewBindGroupPermissionsUsecase(c.rbac.Admin, c.tx),
		}),
		apphandler.NewRbacCacheHandler(rbacusecase.NewRefreshRbacCacheUsecase(c.rbac.Cache)),
	}
}

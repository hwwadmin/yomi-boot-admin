package bootstrap

import (
	apphandler "yomi-admin-backend/apps/server/internal/handler"
	"yomi-admin-backend/apps/server/internal/support/superadmin"
	rbacusecase "yomi-admin-backend/apps/server/internal/usecase/rbac"
	userusecase "yomi-admin-backend/apps/server/internal/usecase/user"

	"gitee.com/hq_project/yomi-boot/component/core/router"
)

// newUserRegistrar 组装系统用户模块：用户 CRUD、重置密码与角色绑定
func newUserRegistrar(c container, guard *superadmin.Guard) router.Registrar {
	return apphandler.NewUserHandler(apphandler.UserHandlerDeps{
		Page:          userusecase.NewPageUsersUsecase(c.users.Users, c.rbac.Admin),
		Create:        userusecase.NewCreateUserUsecase(c.users.Users, c.idem, c.tx),
		Update:        userusecase.NewUpdateUserUsecase(c.users.Users, guard, c.tx, c.token),
		Delete:        userusecase.NewDeleteUserUsecase(c.users.Users, c.rbac.Admin, guard, c.tx, c.token),
		ResetPassword: userusecase.NewResetUserPasswordUsecase(c.users.Users, guard, c.token),
		GetRoles:      rbacusecase.NewGetUserRolesUsecase(c.users.Users, c.rbac.Admin),
		BindRoles:     rbacusecase.NewBindUserRolesUsecase(c.rbac.Admin, guard, c.tx),
		Audit:         c.auditRecorder,
	})
}

// newUserSessionRegistrar 组装用户在线会话模块
func newUserSessionRegistrar(tokens userusecase.SessionTokens, guard *superadmin.Guard) router.Registrar {
	return apphandler.NewUserSessionHandler(
		userusecase.NewListUserSessionsUsecase(tokens),
		userusecase.NewRevokeUserSessionUsecase(tokens, guard),
		userusecase.NewRevokeUserSessionsUsecase(tokens, guard),
	)
}

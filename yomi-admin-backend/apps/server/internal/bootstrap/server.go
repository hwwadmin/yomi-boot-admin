// Package bootstrap 编排层装配入口
package bootstrap

import (
	"sync"
	"yomi-admin-backend/apps/server/internal/middleware"
	"yomi-admin-backend/apps/server/internal/support/superadmin"

	apphandler "yomi-admin-backend/apps/server/internal/handler"

	authusecase "yomi-admin-backend/apps/server/internal/usecase/auth"
	httprecoreusecase "yomi-admin-backend/apps/server/internal/usecase/httprecord"
	rbacusecase "yomi-admin-backend/apps/server/internal/usecase/rbac"
	userusecase "yomi-admin-backend/apps/server/internal/usecase/user"
	httprecordprovider "yomi-admin-backend/domains/httprecord/provider"
	rbacprovider "yomi-admin-backend/domains/rbac/provider"
	userprovider "yomi-admin-backend/domains/user/provider"

	"gitee.com/hq_project/yomi-boot/boot"
	"gitee.com/hq_project/yomi-boot/boot/server"
	"gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/common/distlock"
	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/appctx"
	"gitee.com/hq_project/yomi-boot/component/core/config"
	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/router"
	"gitee.com/hq_project/yomi-boot/plugin"
	appctxplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/appctx"
	dbplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/db"
	redisplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/redis"
	snowflakeplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/snowflake"
	tokenplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/token"

	"github.com/gin-gonic/gin"
)

// envPrefix 是配置环境变量前缀
const envPrefix = "YOMI_ADMIN"

// Options 返回启动选项：中间件链统一在组合根登记
func Options() boot.Options {
	return boot.Options{
		EnvPrefix:   envPrefix,
		Middlewares: Middlewares,
	}
}

// container 保存 Wire 与 Middlewares 共用的装配结果
type container struct {
	users   userprovider.Module
	rbac    rbacprovider.Module
	records httprecordprovider.Module
	tx      tx.Executor
	token   token.TokenService
	appCtx  appctx.AppCtx
	lock    distlock.DistLockFactory
}

var (
	loadOnce     sync.Once
	appContainer container
)

// load 初始化并缓存装配结果；进程内只装配一次
func load(slot *plugin.PluginSlot) container {
	loadOnce.Do(func() {
		db := dbplugin.FromDB(slot)
		txPlugin := dbplugin.FromTx(slot)
		tokenPlugin := tokenplugin.From(slot)
		snowflakePlugin := snowflakeplugin.From(slot)
		appCtxPlugin := appctxplugin.From(slot)
		redisPlugin := redisplugin.From(slot)

		if db.DB() == nil || txPlugin.Executor() == nil || tokenPlugin.Service() == nil ||
			snowflakePlugin.Generator() == nil || appCtxPlugin.AppCtx() == nil ||
			redisPlugin.DistLockFactory() == nil {
			panic("yomi-admin: 需要启用 db、redis、token、snowflake 插件")
		}

		appContainer = container{
			users: userprovider.Provide(userprovider.Deps{
				DB:          db.DB(),
				DBType:      db.Type(),
				IDGenerator: snowflakePlugin.Generator(),
			}),
			rbac: rbacprovider.Provide(rbacprovider.Deps{
				DB:          db.DB(),
				DBType:      db.Type(),
				IDGenerator: snowflakePlugin.Generator(),
			}),
			records: httprecordprovider.Provide(httprecordprovider.Deps{
				DB:          db.DB(),
				IDGenerator: snowflakePlugin.Generator(),
			}),
			tx:     txPlugin.Executor(),
			token:  tokenPlugin.Service(),
			appCtx: appCtxPlugin.AppCtx(),
			lock:   redisPlugin.DistLockFactory(),
		}
	})
	return appContainer
}

// Wire 根据插件插槽装配 HTTP 服务
func Wire(slot *plugin.PluginSlot) []router.Registrar {
	c := load(slot)
	registerHttpRecordHandler(slot, c.records.Records)
	guard := superadmin.New(c.users.Users, c.rbac.Admin, newSuperAdminLocker(c.lock))

	return []router.Registrar{
		corehandler.NewHealthHandler("/v1/health", "/v1/health/ops", c.appCtx),
		apphandler.NewAuthHandler(
			authusecase.NewLoginUsecase(c.users.Users, c.rbac.Auth, c.token),
			authusecase.NewProfileUsecase(c.users.Users, c.rbac.Auth),
			authusecase.NewLogoutUsecase(c.token),
		),
		apphandler.NewUserHandler(
			userusecase.NewPageUsersUsecase(c.users.Users, c.rbac.Admin),
			userusecase.NewCreateUserUsecase(c.users.Users, c.tx),
			userusecase.NewUpdateUserUsecase(c.users.Users, guard, c.tx, c.token),
			userusecase.NewDeleteUserUsecase(c.users.Users, guard, c.tx, c.token),
		),
		apphandler.NewRbacHandler(
			rbacusecase.NewPageRolesUsecase(c.rbac.Admin),
			rbacusecase.NewGetRoleUsecase(c.rbac.Admin),
			rbacusecase.NewCreateRoleUsecase(c.rbac.Admin),
			rbacusecase.NewUpdateRoleUsecase(c.rbac.Admin),
			rbacusecase.NewBindRoleGroupsUsecase(c.rbac.Admin, c.tx),
			rbacusecase.NewPagePermissionsUsecase(c.rbac.Admin),
			rbacusecase.NewSetPermissionStatusUsecase(c.rbac.Admin),
			rbacusecase.NewGetUserRolesUsecase(c.users.Users, c.rbac.Admin),
			rbacusecase.NewBindUserRolesUsecase(c.rbac.Admin, guard, c.tx),
		),
		apphandler.NewPermissionGroupHandler(
			rbacusecase.NewPagePermissionGroupsUsecase(c.rbac.Admin),
			rbacusecase.NewGetPermissionGroupUsecase(c.rbac.Admin),
			rbacusecase.NewCreatePermissionGroupUsecase(c.rbac.Admin),
			rbacusecase.NewUpdatePermissionGroupUsecase(c.rbac.Admin),
			rbacusecase.NewBindGroupPermissionsUsecase(c.rbac.Admin, c.tx),
		),
		apphandler.NewHttpRecordHandler(
			httprecoreusecase.NewPageRecordsUsecase(c.records.Records),
			httprecoreusecase.NewGetRecordUsecase(c.records.Records),
		),
	}
}

// Middlewares 在默认中间件链后追加 RBAC 鉴权
func Middlewares(slot *plugin.PluginSlot) []gin.HandlerFunc {
	c := load(slot)

	tokenCfg := config.MustSection[tokenplugin.BootConfig](slot.Config, tokenplugin.SectionKey).ToConfig()
	skip := make(map[string]struct{}, len(tokenCfg.SkipValidatePaths)+2)
	for _, path := range tokenCfg.SkipValidatePaths {
		skip[path] = struct{}{}
	}
	skip["/api/auth/profile"] = struct{}{}
	skip["/api/auth/logout"] = struct{}{}

	middlewares := server.DefaultMiddlewares(slot)
	return append(middlewares, middleware.RBAC(c.rbac.Auth, skip))
}

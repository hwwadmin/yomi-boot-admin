// Package bootstrap 编排层装配入口
package bootstrap

import (
	"sync"
	svcconfig "yomi-admin-backend/apps/server/internal/config"
	"yomi-admin-backend/apps/server/internal/middleware"
	"yomi-admin-backend/apps/server/internal/support/loginguard"
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
	"gitee.com/hq_project/yomi-boot/component/common/captcha"
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
	redisratelimitplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/redisratelimit"
	snowflakeplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/snowflake"
	tokenplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/token"

	"github.com/gin-gonic/gin"
)

// envPrefix 是配置环境变量前缀
const envPrefix = "YOMI_ADMIN"

// Options 返回启动选项：中间件链与业务配置段统一在组合根登记
func Options() boot.Options {
	var pluginOpts plugin.PluginOptions
	pluginOpts.AddConfigSection(svcconfig.ConfigSection())

	return boot.Options{
		EnvPrefix:   envPrefix,
		Middlewares: Middlewares,
		Plugin:      pluginOpts,
	}
}

// container 保存 Wire 与 Middlewares 共用的装配结果
type container struct {
	users      userprovider.Module
	rbac       rbacprovider.Module
	records    httprecordprovider.Module
	tx         tx.Executor
	token      token.TokenService
	captcha    *authusecase.CaptchaUsecase
	loginGuard *loginguard.Guard
	appCtx     appctx.AppCtx
	lock       distlock.DistLockFactory
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
		rateLimitPlugin := redisratelimitplugin.From(slot)

		if db.DB() == nil || txPlugin.Executor() == nil || tokenPlugin.Service() == nil ||
			snowflakePlugin.Generator() == nil || appCtxPlugin.AppCtx() == nil ||
			redisPlugin.DistLockFactory() == nil {
			panic("yomi-admin: 需要启用 db、redis、token、snowflake 插件")
		}

		bizCfg := config.MustSection[svcconfig.BootConfig](slot.Config, svcconfig.SectionKey)
		rateLimitCfg := config.MustSection[redisratelimitplugin.BootConfig](slot.Config, redisratelimitplugin.SectionKey)
		if bizCfg.Captcha.RefreshInterval() > 0 && !rateLimitCfg.EnabledOrDefault() {
			panic("yomi-admin: biz.captcha 刷新限流需要启用 ratelimit 插件（ratelimit.enabled=true）")
		}
		if bizCfg.LoginGuard.EnabledOrDefault() && !rateLimitCfg.EnabledOrDefault() {
			panic("yomi-admin: biz.login_guard 需要启用 ratelimit 插件（ratelimit.enabled=true）")
		}

		captchaStore, err := captcha.NewRedisStore(redisPlugin.Service(), captcha.RedisStoreConfig{
			KeyPrefix: bizCfg.Captcha.KeyPrefixOrDefault(),
		})
		if err != nil {
			panic(err)
		}

		rbacModule, err := rbacprovider.Provide(rbacprovider.Deps{
			DB:          db.DB(),
			DBType:      db.Type(),
			IDGenerator: snowflakePlugin.Generator(),
			Redis:       redisPlugin.Service(),
			Locks:       redisPlugin.DistLockFactory(),
			Tx:          txPlugin.Executor(),
			Cache:       toRbacCacheOptions(bizCfg.Rbac.Cache),
		})
		if err != nil {
			panic(err)
		}

		appContainer = container{
			users: userprovider.Provide(userprovider.Deps{
				DB:          db.DB(),
				DBType:      db.Type(),
				IDGenerator: snowflakePlugin.Generator(),
			}),
			rbac: rbacModule,
			records: httprecordprovider.Provide(httprecordprovider.Deps{
				DB:          db.DB(),
				IDGenerator: snowflakePlugin.Generator(),
			}),
			tx:    txPlugin.Executor(),
			token: tokenPlugin.Service(),
			captcha: authusecase.NewCaptchaUsecase(
				captcha.MustNew(captchaStore, bizCfg.Captcha.ToConfig()),
				rateLimitPlugin.Limiter(),
				bizCfg.Captcha.RefreshInterval(),
				bizCfg.Captcha.EnabledOrDefault(),
			),
			loginGuard: loginguard.New(
				redisPlugin.Service(),
				rateLimitPlugin.Limiter(),
				bizCfg.LoginGuard.ToPolicy(),
			),
			appCtx: appCtxPlugin.AppCtx(),
			lock:   redisPlugin.DistLockFactory(),
		}

		// 启动期刷新与定时刷新挂在运行时 context 上，进程退出时自动停止
		appContainer.rbac.StartBackground(slot.RuntimeContext())
	})
	return appContainer
}

// toRbacCacheOptions 将业务配置段转换为 RBAC 缓存的装配参数
func toRbacCacheOptions(cfg svcconfig.RbacCacheConfig) rbacprovider.CacheOptions {
	return rbacprovider.CacheOptions{
		Enabled:         cfg.EnabledOrDefault(),
		Key:             cfg.KeyOrDefault(),
		LockKey:         cfg.LockKeyOrDefault(),
		MaxBytes:        cfg.MaxBytesOrDefault(),
		LockWait:        cfg.LockWait(),
		BootstrapWarmup: cfg.BootstrapWarmupOrDefault(),
		RefreshInterval: cfg.RefreshInterval(),
	}
}

// Wire 根据插件插槽装配 HTTP 服务
func Wire(slot *plugin.PluginSlot) []router.Registrar {
	c := load(slot)
	registerHttpRecordHandler(slot, c.records.Records)
	guard := superadmin.New(c.users.Users, c.rbac.Admin, newSuperAdminLocker(c.lock))

	return []router.Registrar{
		corehandler.NewHealthHandler("/v1/health", "/v1/health/ops", c.appCtx),
		apphandler.NewAuthHandler(
			authusecase.NewLoginUsecase(c.users.Users, c.rbac.Auth, c.token, c.captcha, c.loginGuard),
			c.captcha,
			authusecase.NewProfileUsecase(c.users.Users, c.rbac.Auth),
			authusecase.NewLogoutUsecase(c.token),
		),
		apphandler.NewUserHandler(
			userusecase.NewPageUsersUsecase(c.users.Users, c.rbac.Admin),
			userusecase.NewCreateUserUsecase(c.users.Users, c.tx),
			userusecase.NewUpdateUserUsecase(c.users.Users, guard, c.tx, c.token),
			userusecase.NewDeleteUserUsecase(c.users.Users, c.rbac.Admin, guard, c.tx, c.token),
			rbacusecase.NewGetUserRolesUsecase(c.users.Users, c.rbac.Admin),
			rbacusecase.NewBindUserRolesUsecase(c.rbac.Admin, guard, c.tx),
		),
		apphandler.NewRoleHandler(
			rbacusecase.NewPageRolesUsecase(c.rbac.Admin),
			rbacusecase.NewGetRoleUsecase(c.rbac.Admin),
			rbacusecase.NewCreateRoleUsecase(c.rbac.Admin),
			rbacusecase.NewUpdateRoleUsecase(c.rbac.Admin),
			rbacusecase.NewDeleteRoleUsecase(c.rbac.Admin, c.tx),
			rbacusecase.NewSetRoleStatusUsecase(c.rbac.Admin),
			rbacusecase.NewBindRoleGroupsUsecase(c.rbac.Admin, c.tx),
		),
		apphandler.NewPermissionHandler(
			rbacusecase.NewPagePermissionsUsecase(c.rbac.Admin),
			rbacusecase.NewCreatePermissionUsecase(c.rbac.Admin),
			rbacusecase.NewUpdatePermissionUsecase(c.rbac.Admin),
			rbacusecase.NewDeletePermissionUsecase(c.rbac.Admin, c.tx),
			rbacusecase.NewSetPermissionStatusUsecase(c.rbac.Admin),
		),
		apphandler.NewPermissionGroupHandler(
			rbacusecase.NewPagePermissionGroupsUsecase(c.rbac.Admin),
			rbacusecase.NewGetPermissionGroupUsecase(c.rbac.Admin),
			rbacusecase.NewCreatePermissionGroupUsecase(c.rbac.Admin),
			rbacusecase.NewUpdatePermissionGroupUsecase(c.rbac.Admin),
			rbacusecase.NewDeletePermissionGroupUsecase(c.rbac.Admin, c.tx),
			rbacusecase.NewSetPermissionGroupStatusUsecase(c.rbac.Admin),
			rbacusecase.NewBindGroupPermissionsUsecase(c.rbac.Admin, c.tx),
		),
		apphandler.NewHttpRecordHandler(
			httprecoreusecase.NewPageRecordsUsecase(c.records.Records),
			httprecoreusecase.NewGetRecordUsecase(c.records.Records),
		),
		apphandler.NewRbacCacheHandler(
			rbacusecase.NewRefreshRbacCacheUsecase(c.rbac.Cache),
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

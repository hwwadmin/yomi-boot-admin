// Package bootstrap 编排层装配入口
package bootstrap

import (
	"sync"
	svcconfig "yomi-admin-backend/apps/server/internal/config"
	"yomi-admin-backend/apps/server/internal/middleware"
	"yomi-admin-backend/apps/server/internal/support/audit"
	"yomi-admin-backend/apps/server/internal/support/loginguard"
	authusecase "yomi-admin-backend/apps/server/internal/usecase/auth"
	auditlogprovider "yomi-admin-backend/domains/auditlog/provider"
	httprecordprovider "yomi-admin-backend/domains/httprecord/provider"
	rbacprovider "yomi-admin-backend/domains/rbac/provider"
	sysparamprovider "yomi-admin-backend/domains/sysparam/provider"
	userprovider "yomi-admin-backend/domains/user/provider"

	"gitee.com/hq_project/yomi-boot/boot"
	"gitee.com/hq_project/yomi-boot/boot/server"
	"gitee.com/hq_project/yomi-boot/component/common/captcha"
	"gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/common/distlock"
	"gitee.com/hq_project/yomi-boot/component/common/token"
	"gitee.com/hq_project/yomi-boot/component/core/appctx"
	"gitee.com/hq_project/yomi-boot/component/core/config"
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
	users         userprovider.Module
	rbac          rbacprovider.Module
	records       httprecordprovider.Module
	audit         auditlogprovider.Module
	auditRecorder audit.Recorder
	sysparam      sysparamprovider.Module
	tx            tx.Executor
	token         token.TokenService
	captcha       *authusecase.CaptchaUsecase
	loginGuard    *loginguard.Guard
	loginConfig   *authusecase.LoginConfigUsecase
	appCtx        appctx.AppCtx
	lock          distlock.DistLockFactory
	// rbacSkip 是免接口权限校验的路由集合，键为 "GET /api/auth/login"，由 biz.auth.public_paths 生成
	rbacSkip map[string]struct{}
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

		// 公开路由只在 biz.auth.public_paths 声明（path 为完整 URI）：这里把它同时注入
		// 令牌中间件的跳过清单，并生成 RBAC 的跳过集合，保证"哪些路由免鉴权"只有一个来源。
		rbacSkip, anonymousPaths := splitPublicRoutes(bizCfg.Auth.PublicRoutes())
		config.MustSection[tokenplugin.BootConfig](slot.Config, tokenplugin.SectionKey).SkipValidatePaths = anonymousPaths

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

		auditModule := auditlogprovider.Provide(auditlogprovider.Deps{
			DB:          db.DB(),
			IDGenerator: snowflakePlugin.Generator(),
		})
		var auditRecorder audit.Recorder
		if bizCfg.Audit.EnabledOrDefault() {
			recorder := audit.NewTwoPhaseRecorder(auditModule.Writer, snowflakePlugin.Generator())
			// 异步落库：启动消费协程，进程退出时冲刷队列后停止
			recorder.StartBackground(slot.RuntimeContext())
			auditRecorder = recorder
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
			audit:         auditModule,
			auditRecorder: auditRecorder,
			sysparam: sysparamprovider.Provide(sysparamprovider.Deps{
				DB:          db.DB(),
				DBType:      db.Type(),
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
			loginConfig: authusecase.NewLoginConfigUsecase(bizCfg.Captcha.EnabledOrDefault()),
			appCtx:      appCtxPlugin.AppCtx(),
			lock:        redisPlugin.DistLockFactory(),

			rbacSkip: rbacSkip,
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
		MemorySync:      cfg.MemorySyncOrDefault(),
	}
}

// Middlewares 在默认中间件链后追加 RBAC 鉴权
func Middlewares(slot *plugin.PluginSlot) []gin.HandlerFunc {
	c := load(slot)

	middlewares := server.DefaultMiddlewares(slot)
	return append(middlewares, middleware.RBAC(c.rbac.Auth, c.rbacSkip))
}

// splitPublicRoutes 把公开路由拆成 RBAC 跳过集合与免令牌路径清单。
//
// RBAC 跳过键为"方法 + 完整路径"；免令牌清单只取 token=false 的条目，
// 因为令牌中间件只能按路径跳过。
func splitPublicRoutes(routes []svcconfig.PublicRoute) (map[string]struct{}, []string) {
	skip := make(map[string]struct{}, len(routes))
	anonymous := make([]string, 0, len(routes))
	for _, route := range routes {
		skip[route.Method+" "+route.Path] = struct{}{}
		if !route.Token {
			anonymous = append(anonymous, route.Path)
		}
	}
	return skip, anonymous
}

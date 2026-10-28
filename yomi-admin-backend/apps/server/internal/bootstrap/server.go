// Package bootstrap 编排层装配入口
package bootstrap

import (
	apphandler "yomi-admin-backend/apps/server/internal/handler"
	"yomi-admin-backend/apps/server/internal/usecase"
	userprovider "yomi-admin-backend/domains/user/provider"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/router"
	"gitee.com/hq_project/yomi-boot/plugin"
	appctxplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/appctx"
	dbplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/db"
	tokenplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/token"
)

// Wire 根据插件插槽装配 HTTP 服务
func Wire(slot *plugin.PluginSlot) []router.Registrar {
	db := dbplugin.FromDB(slot)
	txPlugin := dbplugin.FromTx(slot)
	tokenPlugin := tokenplugin.From(slot)
	appCtxPlugin := appctxplugin.From(slot)

	if db.DB() == nil || txPlugin.Executor() == nil || tokenPlugin.Service() == nil || appCtxPlugin.AppCtx() == nil {
		panic("yomi-admin: 需要启用 db、redis、token 插件")
	}

	users := userprovider.Provide(userprovider.Deps{
		DB:     db.DB(),
		DBType: db.Type(),
	})

	return []router.Registrar{
		corehandler.NewHealthHandler("/v1/health", "/v1/health/ops", appCtxPlugin.AppCtx()),
		apphandler.NewAuthHandler(
			usecase.NewLoginUsecase(users.Users, tokenPlugin.Service()),
			usecase.NewProfileUsecase(users.Users),
			usecase.NewLogoutUsecase(tokenPlugin.Service()),
		),
		apphandler.NewUserHandler(
			usecase.NewPageUsersUsecase(users.Users),
			usecase.NewCreateUserUsecase(users.Users, txPlugin.Executor()),
			usecase.NewUpdateUserUsecase(users.Users, txPlugin.Executor(), tokenPlugin.Service()),
			usecase.NewDeleteUserUsecase(users.Users, txPlugin.Executor(), tokenPlugin.Service()),
		),
	}
}

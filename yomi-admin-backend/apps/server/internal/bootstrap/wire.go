// Package bootstrap 编排层装配入口
package bootstrap

import (
	"yomi-admin-backend/apps/server/internal/support/superadmin"
	superadminredislock "yomi-admin-backend/apps/server/internal/support/superadmin/redislock"
	userusecase "yomi-admin-backend/apps/server/internal/usecase/user"

	"gitee.com/hq_project/yomi-boot/component/common/token"
	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/router"
	"gitee.com/hq_project/yomi-boot/plugin"
)

// Wire 根据插件插槽装配 HTTP 服务。
//
// 这里只列模块与跨模块依赖：各上下文的 Handler 构造放在 wire_<context>.go，
// 接口增减只影响对应上下文的 wire 文件，中心函数不随接口数增长。
func Wire(slot *plugin.PluginSlot) []router.Registrar {
	c := load(slot)
	registerHttpRecordConsumer(slot, c)
	guard := superadmin.New(c.users.Users, c.rbac.Admin, superadminredislock.New(c.lock))
	sessionTokens := mustSessionTokens(c.token)

	registrars := []router.Registrar{
		corehandler.NewHealthHandler("/api/v1/health", "/api/v1/health/ops", c.appCtx),
		newAuthRegistrar(c),
		newUserRegistrar(c, guard),
		newUserSessionRegistrar(sessionTokens, guard),
	}
	registrars = append(registrars, newRbacRegistrars(c)...)
	registrars = append(registrars,
		newHttpRecordRegistrar(c),
		newSysParamRegistrar(c),
		newAuditLogRegistrar(c),
	)
	registrars = append(registrars, newDocsRegistrars(slot)...)
	return registrars
}

// mustSessionTokens 桥接基础库尚未纳入 TokenService 接口的按用户列举能力。
func mustSessionTokens(t token.TokenService) userusecase.SessionTokens {
	sessionTokens, ok := t.(userusecase.SessionTokens)
	if !ok {
		panic("yomi-admin: 当前 yomi-boot token 服务未提供 ListByUserID，无法启用用户会话管理")
	}
	return sessionTokens
}

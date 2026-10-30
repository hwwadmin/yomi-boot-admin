package bootstrap

import (
	apphandler "yomi-admin-backend/apps/server/internal/handler"
	auditlogusecase "yomi-admin-backend/apps/server/internal/usecase/auditlog"

	"gitee.com/hq_project/yomi-boot/component/core/router"
)

// newAuditLogRegistrar 组装操作审计日志查询模块
func newAuditLogRegistrar(c container) router.Registrar {
	return apphandler.NewAuditLogHandler(apphandler.AuditLogHandlerDeps{
		Page: auditlogusecase.NewPageLogsUsecase(c.audit.Reader, c.users.Users),
	})
}

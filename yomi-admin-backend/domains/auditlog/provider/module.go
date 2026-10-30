// Package provider 操作审计日志域装配入口（仅 apps/server/bootstrap 使用）
package provider

import (
	"yomi-admin-backend/domains/auditlog/internal/repo"
	"yomi-admin-backend/domains/auditlog/internal/service"
	"yomi-admin-backend/domains/auditlog/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/snowflake"

	"gorm.io/gorm"
)

// Deps 装配依赖
type Deps struct {
	DB          *gorm.DB
	IDGenerator snowflake.Generator
}

// Module 导出本域能力
type Module struct {
	Writer service_contract.AuditLogWriter
	Reader service_contract.AuditLogReader
}

// Provide 装配 auditlog 域
func Provide(deps Deps) Module {
	logRepo := repo.NewAuditLogRepo(deps.DB)
	svc := service.NewAuditLogService(logRepo, deps.IDGenerator)
	return Module{Writer: svc, Reader: svc}
}

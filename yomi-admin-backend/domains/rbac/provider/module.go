// Package provider RBAC 域装配入口（仅 apps/server/bootstrap 使用）
package provider

import (
	"yomi-admin-backend/domains/rbac/internal/repo"
	"yomi-admin-backend/domains/rbac/internal/service"
	"yomi-admin-backend/domains/rbac/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/snowflake"
	"gitee.com/hq_project/yomi-boot/component/core/shared/enums"

	"gorm.io/gorm"
)

// Deps 装配依赖
type Deps struct {
	DB          *gorm.DB
	DBType      enums.DbType
	IDGenerator snowflake.Generator
}

// Module 导出本域能力
type Module struct {
	Admin service_contract.RbacAdminService
	Auth  service_contract.RbacAuthService
}

// Provide 装配 rbac 域
func Provide(deps Deps) Module {
	rbacRepo := repo.NewRbacRepo(deps.DB, deps.DBType)
	return Module{
		Admin: service.NewRbacAdminService(rbacRepo, deps.IDGenerator),
		Auth:  service.NewRbacAuthService(rbacRepo),
	}
}

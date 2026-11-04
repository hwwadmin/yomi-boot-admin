// Package provider 系统参数域装配入口（仅 apps/server/bootstrap 使用）
package provider

import (
	"yomi-admin-backend/domains/sysparam/internal/repo"
	"yomi-admin-backend/domains/sysparam/internal/service"
	"yomi-admin-backend/domains/sysparam/service_contract"

	"github.com/hwwadmin/yomi-boot/component/common/snowflake"
	sharedenums "github.com/hwwadmin/yomi-boot/component/core/shared/enums"

	"gorm.io/gorm"
)

// Deps 装配依赖
type Deps struct {
	DB          *gorm.DB
	DBType      sharedenums.DbType
	IDGenerator snowflake.Generator
}

// Module 导出本域能力
type Module struct {
	Admin  service_contract.SysParamAdminService
	Reader service_contract.SysParamReader
}

// Provide 装配 sysparam 域
func Provide(deps Deps) Module {
	paramRepo := repo.NewSysParamRepo(deps.DB, deps.DBType)
	return Module{
		Admin:  service.NewSysParamAdminService(paramRepo, deps.IDGenerator),
		Reader: service.NewSysParamReaderService(paramRepo),
	}
}

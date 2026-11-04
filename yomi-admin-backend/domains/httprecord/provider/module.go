// Package provider HTTP 访问记录域装配入口（仅 apps/server/bootstrap 使用）
package provider

import (
	"yomi-admin-backend/domains/httprecord/internal/repo"
	"yomi-admin-backend/domains/httprecord/internal/service"
	"yomi-admin-backend/domains/httprecord/service_contract"

	"github.com/hwwadmin/yomi-boot/component/common/snowflake"

	"gorm.io/gorm"
)

// Deps 装配依赖
type Deps struct {
	DB          *gorm.DB
	IDGenerator snowflake.Generator
}

// Module 导出本域能力
type Module struct {
	Records service_contract.RecordService
}

// Provide 装配 httprecord 域
func Provide(deps Deps) Module {
	recordRepo := repo.NewRecordRepo(deps.DB)
	return Module{
		Records: service.NewRecordService(recordRepo, deps.IDGenerator),
	}
}

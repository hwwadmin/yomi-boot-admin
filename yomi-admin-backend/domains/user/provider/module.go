// Package provider 用户域装配入口（仅 apps/server/bootstrap 使用）
package provider

import (
	"yomi-admin-backend/domains/user/internal/repo"
	"yomi-admin-backend/domains/user/internal/service"
	"yomi-admin-backend/domains/user/service_contract"

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

// Module 导出本域能力（均为 service_contract 接口）
type Module struct {
	Users service_contract.UserService
}

// Provide 装配 user 域
func Provide(deps Deps) Module {
	userRepo := repo.NewUserRepo(deps.DB, deps.DBType)
	return Module{Users: service.NewUserService(userRepo, deps.IDGenerator)}
}

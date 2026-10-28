// Package main 管理后台编排层入口
package main

import (
	"yomi-admin-backend/apps/server/internal/bootstrap"

	"gitee.com/hq_project/yomi-boot/boot"
)

// envPrefix 是配置环境变量前缀
const envPrefix = "YOMI_ADMIN"

func main() {
	boot.RunWithOptions(boot.Options{
		EnvPrefix:   envPrefix,
		Middlewares: bootstrap.Middlewares,
	}, bootstrap.Wire)
}

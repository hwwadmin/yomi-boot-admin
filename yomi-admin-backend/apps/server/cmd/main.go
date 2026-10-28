// Package main 管理后台编排层入口
package main

import (
	"yomi-admin-backend/apps/server/internal/bootstrap"

	"gitee.com/hq_project/yomi-boot/boot"
	"gitee.com/hq_project/yomi-boot/plugin"
)

// envPrefix 是配置环境变量前缀
const envPrefix = "YOMI_ADMIN"

func main() {
	var pluginOpts plugin.PluginOptions
	pluginOpts.AddCustomPlugin(bootstrap.NewHttpRecordConsumerPlugin())

	boot.RunWithOptions(boot.Options{
		EnvPrefix:   envPrefix,
		Plugin:      pluginOpts,
		Middlewares: bootstrap.Middlewares,
	}, bootstrap.Wire)
}

// Package bootstrap 编排层装配入口
package bootstrap

import (
	apphandler "yomi-admin-backend/apps/server/internal/handler"
	httprecoreusecase "yomi-admin-backend/apps/server/internal/usecase/httprecord"

	"gitee.com/hq_project/yomi-boot/plugin"
	httprecordplugin "gitee.com/hq_project/yomi-boot/plugin/builtin/httprecord"
)

// registerHttpRecordHandler 向 httprecord 插件注册访问记录消费适配器
// 须在插件 OnReady 启动消费者之前调用，wire 阶段即满足该时序
func registerHttpRecordHandler(slot *plugin.PluginSlot, save *httprecoreusecase.SaveRecordUsecase) {
	recordPlugin := httprecordplugin.From(slot)
	if recordPlugin == nil {
		return
	}
	recordPlugin.MustRegisterHandler(apphandler.NewHttpRecordConsumer(save))
}

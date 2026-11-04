package bootstrap

import (
	apphandler "yomi-admin-backend/apps/server/internal/handler"
	httprecoreusecase "yomi-admin-backend/apps/server/internal/usecase/httprecord"

	"github.com/hwwadmin/yomi-boot/component/core/router"
	"github.com/hwwadmin/yomi-boot/plugin"
	httprecordplugin "github.com/hwwadmin/yomi-boot/plugin/builtin/httprecord"
)

// registerHttpRecordConsumer 向 httprecord 插件注册消费适配器。
// 须在插件 OnReady 启动消费者之前调用，wire 阶段即满足该时序。
func registerHttpRecordConsumer(slot *plugin.PluginSlot, c container) {
	recordPlugin := httprecordplugin.From(slot)
	if recordPlugin == nil {
		return
	}
	recordPlugin.MustRegisterHandler(apphandler.NewHttpRecordConsumer(
		httprecoreusecase.NewSaveRecordUsecase(c.records.Records),
	))
}

// newHttpRecordRegistrar 组装 HTTP 访问记录查询模块
func newHttpRecordRegistrar(c container) router.Registrar {
	return apphandler.NewHttpRecordHandler(
		httprecoreusecase.NewPageRecordsUsecase(c.records.Records),
		httprecoreusecase.NewGetRecordUsecase(c.records.Records),
	)
}

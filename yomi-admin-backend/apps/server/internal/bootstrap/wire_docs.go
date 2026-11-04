package bootstrap

import (
	"fmt"
	svcconfig "yomi-admin-backend/apps/server/internal/config"

	apphandler "yomi-admin-backend/apps/server/internal/handler"

	"github.com/hwwadmin/yomi-boot/component/core/config"
	"github.com/hwwadmin/yomi-boot/component/core/logger"
	"github.com/hwwadmin/yomi-boot/component/core/router"
	"github.com/hwwadmin/yomi-boot/plugin"
	routerplugin "github.com/hwwadmin/yomi-boot/plugin/builtin/router"
)

// newDocsRegistrars 按 biz.docs.enabled 决定是否挂载 API 文档；生产默认关闭。
func newDocsRegistrars(slot *plugin.PluginSlot) []router.Registrar {
	bizCfg := config.MustSection[svcconfig.BootConfig](slot.Config, svcconfig.SectionKey)
	if !bizCfg.Docs.EnabledOrDefault() {
		return nil
	}
	httpCfg := config.MustSection[routerplugin.BootConfig](slot.Config, routerplugin.SectionKey).ToConfig()
	logger.Info(slot.RuntimeContext(), "API 文档已启用",
		"url", fmt.Sprintf("http://127.0.0.1:%d%s/api/system/api-doc", httpCfg.Port, httpCfg.RouterPrefixClean()))
	return []router.Registrar{apphandler.NewApiDocHandler()}
}

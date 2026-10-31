package bootstrap

import (
	apphandler "yomi-admin-backend/apps/server/internal/handler"
	sysparamusecase "yomi-admin-backend/apps/server/internal/usecase/sysparam"

	"gitee.com/hq_project/yomi-boot/component/core/router"
)

// newSysParamRegistrar 组装系统参数模块
func newSysParamRegistrar(c container) router.Registrar {
	return apphandler.NewSysParamHandler(apphandler.SysParamHandlerDeps{
		Page:      sysparamusecase.NewPageParamsUsecase(c.sysparam.Admin),
		Get:       sysparamusecase.NewGetParamUsecase(c.sysparam.Admin),
		Create:    sysparamusecase.NewCreateParamUsecase(c.sysparam.Admin, c.idem, c.tx),
		Update:    sysparamusecase.NewUpdateParamUsecase(c.sysparam.Admin),
		Delete:    sysparamusecase.NewDeleteParamUsecase(c.sysparam.Admin),
		SetStatus: sysparamusecase.NewSetParamStatusUsecase(c.sysparam.Admin),
	})
}

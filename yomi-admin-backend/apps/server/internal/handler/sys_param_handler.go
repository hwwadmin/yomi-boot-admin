// Package handler HTTP 适配层
package handler

import (
	"context"
	sysparamusecase "yomi-admin-backend/apps/server/internal/usecase/sysparam"
	sysparamenums "yomi-admin-backend/domains/sysparam/pkg/shared/enums"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"

	"github.com/gin-gonic/gin"
)

// SysParamHandler 系统参数管理路由
type SysParamHandler struct {
	pageUC      *sysparamusecase.PageParamsUsecase
	getUC       *sysparamusecase.GetParamUsecase
	createUC    *sysparamusecase.CreateParamUsecase
	updateUC    *sysparamusecase.UpdateParamUsecase
	deleteUC    *sysparamusecase.DeleteParamUsecase
	setStatusUC *sysparamusecase.SetParamStatusUsecase
}

// NewSysParamHandler 创建系统参数 Handler
func NewSysParamHandler(
	pageUC *sysparamusecase.PageParamsUsecase,
	getUC *sysparamusecase.GetParamUsecase,
	createUC *sysparamusecase.CreateParamUsecase,
	updateUC *sysparamusecase.UpdateParamUsecase,
	deleteUC *sysparamusecase.DeleteParamUsecase,
	setStatusUC *sysparamusecase.SetParamStatusUsecase,
) *SysParamHandler {
	return &SysParamHandler{
		pageUC:      pageUC,
		getUC:       getUC,
		createUC:    createUC,
		updateUC:    updateUC,
		deleteUC:    deleteUC,
		setStatusUC: setStatusUC,
	}
}

// Routes 注册系统参数管理路由
func (h *SysParamHandler) Routes(router *gin.RouterGroup) {
	router.GET("/api/system/param", h.page)
	router.GET("/api/system/param/detail", h.get)
	router.POST("/api/system/param/create", h.create)
	router.POST("/api/system/param/update", h.update)
	router.POST("/api/system/param/delete", h.delete)
	router.POST("/api/system/param/set-status", h.setStatus)
}

func (h *SysParamHandler) page(c *gin.Context) {
	in := &sysparamusecase.PageParamsInput{
		PageInput: dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:   c.Query("keyword"),
		Status:    sysparamenums.NormalizeStatus(c.Query("status")),
		ValueType: sysparamenums.NormalizeParamValueType(c.Query("valueType")),
	}
	corehandler.Handle(c, func(ctx context.Context) (dto.PageOutput[sysparamusecase.ParamItem], error) {
		return h.pageUC.Execute(ctx, in)
	})
}

func (h *SysParamHandler) get(c *gin.Context) {
	id, ok := queryID(c, "id")
	if !ok {
		return
	}
	corehandler.Handle(c, func(ctx context.Context) (*sysparamusecase.ParamItem, error) {
		return h.getUC.Execute(ctx, id)
	})
}

func (h *SysParamHandler) create(c *gin.Context) {
	var in sysparamusecase.CreateParamInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.Handle(c, func(ctx context.Context) (*sysparamusecase.ParamItem, error) {
		return h.createUC.Execute(ctx, &in)
	})
}

func (h *SysParamHandler) update(c *gin.Context) {
	var in sysparamusecase.UpdateParamInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.updateUC.Execute(ctx, &in)
	})
}

func (h *SysParamHandler) delete(c *gin.Context) {
	var in sysparamusecase.DeleteParamInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.deleteUC.Execute(ctx, &in)
	})
}

func (h *SysParamHandler) setStatus(c *gin.Context) {
	var in sysparamusecase.SetParamStatusInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.setStatusUC.Execute(ctx, &in)
	})
}

// Package handler HTTP 适配层
package handler

import (
	"context"
	rbacusecase "yomi-admin-backend/apps/server/internal/usecase/rbac"
	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"

	"github.com/gin-gonic/gin"
)

// PermissionHandler 权限目录管理路由
type PermissionHandler struct {
	pageUC      *rbacusecase.PagePermissionsUsecase
	createUC    *rbacusecase.CreatePermissionUsecase
	updateUC    *rbacusecase.UpdatePermissionUsecase
	deleteUC    *rbacusecase.DeletePermissionUsecase
	setStatusUC *rbacusecase.SetPermissionStatusUsecase
}

// NewPermissionHandler 创建权限 Handler
func NewPermissionHandler(
	pageUC *rbacusecase.PagePermissionsUsecase,
	createUC *rbacusecase.CreatePermissionUsecase,
	updateUC *rbacusecase.UpdatePermissionUsecase,
	deleteUC *rbacusecase.DeletePermissionUsecase,
	setStatusUC *rbacusecase.SetPermissionStatusUsecase,
) *PermissionHandler {
	return &PermissionHandler{
		pageUC:      pageUC,
		createUC:    createUC,
		updateUC:    updateUC,
		deleteUC:    deleteUC,
		setStatusUC: setStatusUC,
	}
}

// Routes 注册权限目录管理路由
func (h *PermissionHandler) Routes(router *gin.RouterGroup) {
	router.GET("/api/system/permission", h.page)
	router.POST("/api/system/permission/create", h.create)
	router.POST("/api/system/permission/update", h.update)
	router.POST("/api/system/permission/delete", h.delete)
	router.POST("/api/system/permission/set-status", h.setStatus)
}

func (h *PermissionHandler) page(c *gin.Context) {
	in := &rbacusecase.PagePermissionsInput{
		PageInput: dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:   c.Query("keyword"),
		Type:      rbacenums.PermissionType(c.Query("type")),
		Status:    rbacenums.Status(c.Query("status")),
	}
	corehandler.Handle(c, func(ctx context.Context) (dto.PageOutput[rbacusecase.PermissionItem], error) {
		return h.pageUC.Execute(ctx, in)
	})
}

func (h *PermissionHandler) create(c *gin.Context) {
	var in rbacusecase.CreatePermissionInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.Handle(c, func(ctx context.Context) (*rbacusecase.PermissionItem, error) {
		return h.createUC.Execute(ctx, &in)
	})
}

func (h *PermissionHandler) update(c *gin.Context) {
	var in rbacusecase.UpdatePermissionInput
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

func (h *PermissionHandler) delete(c *gin.Context) {
	var in rbacusecase.DeletePermissionInput
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

func (h *PermissionHandler) setStatus(c *gin.Context) {
	var in rbacusecase.SetPermissionStatusInput
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

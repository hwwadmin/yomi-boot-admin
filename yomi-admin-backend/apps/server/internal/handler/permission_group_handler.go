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

// PermissionGroupHandler 权限组管理路由
type PermissionGroupHandler struct {
	pageUC            *rbacusecase.PagePermissionGroupsUsecase
	getUC             *rbacusecase.GetPermissionGroupUsecase
	createUC          *rbacusecase.CreatePermissionGroupUsecase
	updateUC          *rbacusecase.UpdatePermissionGroupUsecase
	deleteUC          *rbacusecase.DeletePermissionGroupUsecase
	setStatusUC       *rbacusecase.SetPermissionGroupStatusUsecase
	bindPermissionsUC *rbacusecase.BindGroupPermissionsUsecase
}

// NewPermissionGroupHandler 创建权限组 Handler
func NewPermissionGroupHandler(
	pageUC *rbacusecase.PagePermissionGroupsUsecase,
	getUC *rbacusecase.GetPermissionGroupUsecase,
	createUC *rbacusecase.CreatePermissionGroupUsecase,
	updateUC *rbacusecase.UpdatePermissionGroupUsecase,
	deleteUC *rbacusecase.DeletePermissionGroupUsecase,
	setStatusUC *rbacusecase.SetPermissionGroupStatusUsecase,
	bindPermissionsUC *rbacusecase.BindGroupPermissionsUsecase,
) *PermissionGroupHandler {
	return &PermissionGroupHandler{
		pageUC:            pageUC,
		getUC:             getUC,
		createUC:          createUC,
		updateUC:          updateUC,
		deleteUC:          deleteUC,
		setStatusUC:       setStatusUC,
		bindPermissionsUC: bindPermissionsUC,
	}
}

// Routes 注册权限组管理路由
func (h *PermissionGroupHandler) Routes(router *gin.RouterGroup) {
	router.GET("/api/system/permission-group", h.page)
	router.GET("/api/system/permission-group/detail", h.get)
	router.POST("/api/system/permission-group/create", h.create)
	router.POST("/api/system/permission-group/update", h.update)
	router.POST("/api/system/permission-group/delete", h.delete)
	router.POST("/api/system/permission-group/set-status", h.setStatus)
	router.POST("/api/system/permission-group/bind-permission", h.bindPermissions)
}

func (h *PermissionGroupHandler) page(c *gin.Context) {
	in := &rbacusecase.PagePermissionGroupsInput{
		PageInput: dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:   c.Query("keyword"),
		Status:    rbacenums.Status(c.Query("status")),
	}
	handle(c, func(ctx context.Context) (dto.PageOutput[rbacusecase.PermissionGroupItem], error) {
		return h.pageUC.Execute(ctx, in)
	})
}

func (h *PermissionGroupHandler) get(c *gin.Context) {
	id, ok := queryID(c, "id")
	if !ok {
		return
	}
	handle(c, func(ctx context.Context) (*rbacusecase.PermissionGroupDetailOutput, error) {
		return h.getUC.Execute(ctx, id)
	})
}

func (h *PermissionGroupHandler) create(c *gin.Context) {
	var in rbacusecase.CreatePermissionGroupInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	handle(c, func(ctx context.Context) (*rbacusecase.PermissionGroupItem, error) {
		return h.createUC.Execute(ctx, &in)
	})
}

func (h *PermissionGroupHandler) update(c *gin.Context) {
	var in rbacusecase.UpdatePermissionGroupInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	handleVoid(c, func(ctx context.Context) error {
		return h.updateUC.Execute(ctx, &in)
	})
}

func (h *PermissionGroupHandler) delete(c *gin.Context) {
	var in rbacusecase.DeletePermissionGroupInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	handleVoid(c, func(ctx context.Context) error {
		return h.deleteUC.Execute(ctx, &in)
	})
}

func (h *PermissionGroupHandler) setStatus(c *gin.Context) {
	var in rbacusecase.SetPermissionGroupStatusInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	handleVoid(c, func(ctx context.Context) error {
		return h.setStatusUC.Execute(ctx, &in)
	})
}

func (h *PermissionGroupHandler) bindPermissions(c *gin.Context) {
	var in rbacusecase.BindGroupPermissionsInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	handleVoid(c, func(ctx context.Context) error {
		return h.bindPermissionsUC.Execute(ctx, &in)
	})
}

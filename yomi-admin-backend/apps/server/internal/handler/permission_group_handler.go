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
	bindPermissionsUC *rbacusecase.BindGroupPermissionsUsecase
}

// NewPermissionGroupHandler 创建权限组 Handler
func NewPermissionGroupHandler(
	pageUC *rbacusecase.PagePermissionGroupsUsecase,
	getUC *rbacusecase.GetPermissionGroupUsecase,
	createUC *rbacusecase.CreatePermissionGroupUsecase,
	updateUC *rbacusecase.UpdatePermissionGroupUsecase,
	bindPermissionsUC *rbacusecase.BindGroupPermissionsUsecase,
) *PermissionGroupHandler {
	return &PermissionGroupHandler{
		pageUC:            pageUC,
		getUC:             getUC,
		createUC:          createUC,
		updateUC:          updateUC,
		bindPermissionsUC: bindPermissionsUC,
	}
}

// Routes 注册权限组管理路由
func (h *PermissionGroupHandler) Routes(router *gin.RouterGroup) {
	router.GET("/system/permission-group", h.page)
	router.GET("/system/permission-group/detail", h.get)
	router.POST("/system/permission-group/create", h.create)
	router.POST("/system/permission-group/update", h.update)
	router.POST("/system/permission-group/bind-permission", h.bindPermissions)
}

func (h *PermissionGroupHandler) page(c *gin.Context) {
	in := &rbacusecase.PagePermissionGroupsInput{
		PageInput: dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:   c.Query("keyword"),
		Status:    rbacenums.Status(c.Query("status")),
	}
	corehandler.HandlePage(c, func(ctx context.Context) (*dto.PageOutput[rbacusecase.PermissionGroupItem], error) {
		return h.pageUC.Execute(ctx, in)
	})
}

func (h *PermissionGroupHandler) get(c *gin.Context) {
	id, ok := queryID(c, "id")
	if !ok {
		return
	}
	corehandler.Handle(c, func(ctx context.Context) (*rbacusecase.PermissionGroupDetailOutput, error) {
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
	corehandler.Handle(c, func(ctx context.Context) (*rbacusecase.PermissionGroupItem, error) {
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
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.updateUC.Execute(ctx, &in)
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
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.bindPermissionsUC.Execute(ctx, &in)
	})
}

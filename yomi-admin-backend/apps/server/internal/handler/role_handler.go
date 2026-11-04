// Package handler 入口适配层：HTTP 路由与插件消费回调
package handler

import (
	"context"
	rbacusecase "yomi-admin-backend/apps/server/internal/usecase/rbac"
	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"

	corehandler "github.com/hwwadmin/yomi-boot/component/core/handler"
	"github.com/hwwadmin/yomi-boot/component/core/shared/dto"

	"github.com/gin-gonic/gin"
)

// RoleHandler 角色管理路由
type RoleHandler struct {
	pageUC       *rbacusecase.PageRolesUsecase
	getUC        *rbacusecase.GetRoleUsecase
	createUC     *rbacusecase.CreateRoleUsecase
	updateUC     *rbacusecase.UpdateRoleUsecase
	deleteUC     *rbacusecase.DeleteRoleUsecase
	setStatusUC  *rbacusecase.SetRoleStatusUsecase
	bindGroupsUC *rbacusecase.BindRoleGroupsUsecase
}

// RoleHandlerDeps 角色管理 Handler 依赖
type RoleHandlerDeps struct {
	Page       *rbacusecase.PageRolesUsecase
	Get        *rbacusecase.GetRoleUsecase
	Create     *rbacusecase.CreateRoleUsecase
	Update     *rbacusecase.UpdateRoleUsecase
	Delete     *rbacusecase.DeleteRoleUsecase
	SetStatus  *rbacusecase.SetRoleStatusUsecase
	BindGroups *rbacusecase.BindRoleGroupsUsecase
}

// NewRoleHandler 创建角色 Handler
func NewRoleHandler(deps RoleHandlerDeps) *RoleHandler {
	return &RoleHandler{
		pageUC:       deps.Page,
		getUC:        deps.Get,
		createUC:     deps.Create,
		updateUC:     deps.Update,
		deleteUC:     deps.Delete,
		setStatusUC:  deps.SetStatus,
		bindGroupsUC: deps.BindGroups,
	}
}

// Routes 注册角色管理路由
func (h *RoleHandler) Routes(router *gin.RouterGroup) {
	router.GET("/api/system/role", h.page)
	router.GET("/api/system/role/detail", h.get)
	router.POST("/api/system/role/create", h.create)
	router.POST("/api/system/role/update", h.update)
	router.POST("/api/system/role/delete", h.delete)
	router.POST("/api/system/role/set-status", h.setStatus)
	router.POST("/api/system/role/bind-group", h.bindGroups)
}

func (h *RoleHandler) page(c *gin.Context) {
	in := &rbacusecase.PageRolesInput{
		PageInput: dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:   c.Query("keyword"),
		Status:    rbacenums.Status(c.Query("status")),
	}
	handle(c, func(ctx context.Context) (dto.PageOutput[rbacusecase.RoleItem], error) {
		return h.pageUC.Execute(ctx, in)
	})
}

func (h *RoleHandler) get(c *gin.Context) {
	id, ok := queryID(c, "id")
	if !ok {
		return
	}
	handle(c, func(ctx context.Context) (*rbacusecase.RoleDetailOutput, error) {
		return h.getUC.Execute(ctx, id)
	})
}

func (h *RoleHandler) create(c *gin.Context) {
	var in rbacusecase.CreateRoleInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	in.IdempotencyKey = idempotencyKey(c)
	handle(c, func(ctx context.Context) (*rbacusecase.RoleItem, error) {
		return h.createUC.Execute(ctx, &in)
	})
}

func (h *RoleHandler) update(c *gin.Context) {
	var in rbacusecase.UpdateRoleInput
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

func (h *RoleHandler) delete(c *gin.Context) {
	var in rbacusecase.DeleteRoleInput
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

func (h *RoleHandler) setStatus(c *gin.Context) {
	var in rbacusecase.SetRoleStatusInput
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

func (h *RoleHandler) bindGroups(c *gin.Context) {
	var in rbacusecase.BindRoleGroupsInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	handleVoid(c, func(ctx context.Context) error {
		return h.bindGroupsUC.Execute(ctx, &in)
	})
}

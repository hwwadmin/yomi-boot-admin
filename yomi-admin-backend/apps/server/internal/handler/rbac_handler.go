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

// RbacHandler 角色与权限管理路由
type RbacHandler struct {
	pageRolesUC           *rbacusecase.PageRolesUsecase
	getRoleUC             *rbacusecase.GetRoleUsecase
	createRoleUC          *rbacusecase.CreateRoleUsecase
	updateRoleUC          *rbacusecase.UpdateRoleUsecase
	bindGroupsUC          *rbacusecase.BindRoleGroupsUsecase
	pagePermissionsUC     *rbacusecase.PagePermissionsUsecase
	setPermissionStatusUC *rbacusecase.SetPermissionStatusUsecase
	getUserRolesUC        *rbacusecase.GetUserRolesUsecase
	bindUserRolesUC       *rbacusecase.BindUserRolesUsecase
}

// NewRbacHandler 创建 RBAC Handler
func NewRbacHandler(
	pageRolesUC *rbacusecase.PageRolesUsecase,
	getRoleUC *rbacusecase.GetRoleUsecase,
	createRoleUC *rbacusecase.CreateRoleUsecase,
	updateRoleUC *rbacusecase.UpdateRoleUsecase,
	bindGroupsUC *rbacusecase.BindRoleGroupsUsecase,
	pagePermissionsUC *rbacusecase.PagePermissionsUsecase,
	setPermissionStatusUC *rbacusecase.SetPermissionStatusUsecase,
	getUserRolesUC *rbacusecase.GetUserRolesUsecase,
	bindUserRolesUC *rbacusecase.BindUserRolesUsecase,
) *RbacHandler {
	return &RbacHandler{
		pageRolesUC:           pageRolesUC,
		getRoleUC:             getRoleUC,
		createRoleUC:          createRoleUC,
		updateRoleUC:          updateRoleUC,
		bindGroupsUC:          bindGroupsUC,
		pagePermissionsUC:     pagePermissionsUC,
		setPermissionStatusUC: setPermissionStatusUC,
		getUserRolesUC:        getUserRolesUC,
		bindUserRolesUC:       bindUserRolesUC,
	}
}

// Routes 注册角色与权限管理路由
func (h *RbacHandler) Routes(router *gin.RouterGroup) {
	router.GET("/system/role", h.pageRoles)
	router.GET("/system/role/detail", h.getRole)
	router.POST("/system/role/create", h.createRole)
	router.POST("/system/role/update", h.updateRole)
	router.POST("/system/role/bind-group", h.bindRoleGroups)

	router.GET("/system/permission", h.pagePermissions)
	router.POST("/system/permission/set-status", h.setPermissionStatus)

	router.GET("/system/user/role", h.getUserRoles)
	router.POST("/system/user/bind-role", h.bindUserRoles)
}

func (h *RbacHandler) pageRoles(c *gin.Context) {
	in := &rbacusecase.PageRolesInput{
		PageInput: dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:   c.Query("keyword"),
		Status:    rbacenums.Status(c.Query("status")),
	}
	corehandler.HandlePage(c, func(ctx context.Context) (*dto.PageOutput[rbacusecase.RoleItem], error) {
		return h.pageRolesUC.Execute(ctx, in)
	})
}

func (h *RbacHandler) getRole(c *gin.Context) {
	id, ok := queryID(c, "id")
	if !ok {
		return
	}
	corehandler.Handle(c, func(ctx context.Context) (*rbacusecase.RoleDetailOutput, error) {
		return h.getRoleUC.Execute(ctx, id)
	})
}

func (h *RbacHandler) createRole(c *gin.Context) {
	var in rbacusecase.CreateRoleInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.Handle(c, func(ctx context.Context) (*rbacusecase.RoleItem, error) {
		return h.createRoleUC.Execute(ctx, &in)
	})
}

func (h *RbacHandler) updateRole(c *gin.Context) {
	var in rbacusecase.UpdateRoleInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.updateRoleUC.Execute(ctx, &in)
	})
}

func (h *RbacHandler) bindRoleGroups(c *gin.Context) {
	var in rbacusecase.BindRoleGroupsInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.bindGroupsUC.Execute(ctx, &in)
	})
}

func (h *RbacHandler) pagePermissions(c *gin.Context) {
	in := &rbacusecase.PagePermissionsInput{
		PageInput: dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:   c.Query("keyword"),
		Type:      rbacenums.PermissionType(c.Query("type")),
		Status:    rbacenums.Status(c.Query("status")),
	}
	corehandler.HandlePage(c, func(ctx context.Context) (*dto.PageOutput[rbacusecase.PermissionItem], error) {
		return h.pagePermissionsUC.Execute(ctx, in)
	})
}

func (h *RbacHandler) setPermissionStatus(c *gin.Context) {
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
		return h.setPermissionStatusUC.Execute(ctx, &in)
	})
}

func (h *RbacHandler) getUserRoles(c *gin.Context) {
	userID, ok := queryID(c, "userId")
	if !ok {
		return
	}
	corehandler.Handle(c, func(ctx context.Context) (*rbacusecase.UserRolesOutput, error) {
		return h.getUserRolesUC.Execute(ctx, userID)
	})
}

func (h *RbacHandler) bindUserRoles(c *gin.Context) {
	var in rbacusecase.BindUserRolesInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.bindUserRolesUC.Execute(ctx, &in)
	})
}

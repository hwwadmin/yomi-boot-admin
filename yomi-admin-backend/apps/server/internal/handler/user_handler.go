// Package handler HTTP 适配层
package handler

import (
	"context"
	rbacusecase "yomi-admin-backend/apps/server/internal/usecase/rbac"
	userusecase "yomi-admin-backend/apps/server/internal/usecase/user"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"

	"github.com/gin-gonic/gin"
)

// UserHandler 系统用户路由
type UserHandler struct {
	pageUC      *userusecase.PageUsersUsecase
	createUC    *userusecase.CreateUserUsecase
	updateUC    *userusecase.UpdateUserUsecase
	deleteUC    *userusecase.DeleteUserUsecase
	resetPwdUC  *userusecase.ResetUserPasswordUsecase
	getRolesUC  *rbacusecase.GetUserRolesUsecase
	bindRolesUC *rbacusecase.BindUserRolesUsecase
}

// NewUserHandler 创建系统用户 Handler
func NewUserHandler(
	pageUC *userusecase.PageUsersUsecase,
	createUC *userusecase.CreateUserUsecase,
	updateUC *userusecase.UpdateUserUsecase,
	deleteUC *userusecase.DeleteUserUsecase,
	resetPwdUC *userusecase.ResetUserPasswordUsecase,
	getRolesUC *rbacusecase.GetUserRolesUsecase,
	bindRolesUC *rbacusecase.BindUserRolesUsecase,
) *UserHandler {
	return &UserHandler{
		pageUC:      pageUC,
		createUC:    createUC,
		updateUC:    updateUC,
		deleteUC:    deleteUC,
		resetPwdUC:  resetPwdUC,
		getRolesUC:  getRolesUC,
		bindRolesUC: bindRolesUC,
	}
}

// Routes 注册系统用户路由
func (h *UserHandler) Routes(router *gin.RouterGroup) {
	router.GET("/api/system/user", h.page)
	router.POST("/api/system/user/create", h.create)
	router.POST("/api/system/user/update", h.update)
	router.POST("/api/system/user/delete", h.delete)
	router.POST("/api/system/user/reset-password", h.resetPassword)
	router.GET("/api/system/user/role", h.getRoles)
	router.POST("/api/system/user/bind-role", h.bindRoles)
}

func (h *UserHandler) page(c *gin.Context) {
	in := &userusecase.PageUsersInput{
		PageInput: dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:   c.Query("keyword"),
	}
	corehandler.Handle(c, func(ctx context.Context) (dto.PageOutput[userusecase.UserItem], error) {
		return h.pageUC.Execute(ctx, in)
	})
}

func (h *UserHandler) create(c *gin.Context) {
	var in userusecase.CreateUserInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.createUC.Execute(ctx, &in)
	})
}

func (h *UserHandler) update(c *gin.Context) {
	var in userusecase.UpdateUserInput
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

func (h *UserHandler) delete(c *gin.Context) {
	var in userusecase.DeleteUserInput
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

func (h *UserHandler) resetPassword(c *gin.Context) {
	var in userusecase.ResetUserPasswordInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.resetPwdUC.Execute(ctx, &in)
	})
}

func (h *UserHandler) getRoles(c *gin.Context) {
	userID, ok := queryID(c, "userId")
	if !ok {
		return
	}
	corehandler.Handle(c, func(ctx context.Context) (*rbacusecase.UserRolesOutput, error) {
		return h.getRolesUC.Execute(ctx, userID)
	})
}

func (h *UserHandler) bindRoles(c *gin.Context) {
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
		return h.bindRolesUC.Execute(ctx, &in)
	})
}

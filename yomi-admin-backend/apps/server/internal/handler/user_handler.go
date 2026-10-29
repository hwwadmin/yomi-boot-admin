// Package handler HTTP 适配层
package handler

import (
	"context"
	userusecase "yomi-admin-backend/apps/server/internal/usecase/user"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"

	"github.com/gin-gonic/gin"
)

// UserHandler 系统用户路由
type UserHandler struct {
	pageUC   *userusecase.PageUsersUsecase
	createUC *userusecase.CreateUserUsecase
	updateUC *userusecase.UpdateUserUsecase
	deleteUC *userusecase.DeleteUserUsecase
}

// NewUserHandler 创建系统用户 Handler
func NewUserHandler(
	pageUC *userusecase.PageUsersUsecase,
	createUC *userusecase.CreateUserUsecase,
	updateUC *userusecase.UpdateUserUsecase,
	deleteUC *userusecase.DeleteUserUsecase,
) *UserHandler {
	return &UserHandler{pageUC: pageUC, createUC: createUC, updateUC: updateUC, deleteUC: deleteUC}
}

// Routes 注册系统用户路由
func (h *UserHandler) Routes(router *gin.RouterGroup) {
	router.GET("/system/user", h.page)
	router.POST("/system/user/create", h.create)
	router.POST("/system/user/update", h.update)
	router.POST("/system/user/delete", h.delete)
}

func (h *UserHandler) page(c *gin.Context) {
	in := &userusecase.PageUsersInput{
		PageInput: dto.PageInput{Page: queryInt(c, "page", 1), PageSize: queryInt(c, "pageSize", 10)},
		Keyword:   c.Query("keyword"),
	}
	corehandler.HandlePage(c, func(ctx context.Context) (*dto.PageOutput[userusecase.UserItem], error) {
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

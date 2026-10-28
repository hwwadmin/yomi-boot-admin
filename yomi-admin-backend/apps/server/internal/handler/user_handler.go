// Package handler HTTP 适配层
package handler

import (
	"context"
	"strconv"
	userusecase "yomi-admin-backend/apps/server/internal/usecase/user"

	"gitee.com/hq_project/yomi-boot/component/core/errors"
	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/shared/dto"
	sharederr "gitee.com/hq_project/yomi-boot/component/core/shared/error"

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
	router.GET("/system/users", h.page)
	router.POST("/system/users", h.create)
	router.PUT("/system/users/:id", h.update)
	router.DELETE("/system/users/:id", h.delete)
}

func (h *UserHandler) page(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	in := &userusecase.PageUsersInput{
		PageInput: dto.PageInput{Page: page, PageSize: pageSize},
		Keyword:   c.Query("keyword"),
	}
	corehandler.HandlePage(c, func(ctx context.Context) (*dto.PageOutput[userusecase.UserItem], error) {
		return h.pageUC.Execute(ctx, in)
	})
}

func (h *UserHandler) create(c *gin.Context) {
	corehandler.BindAndHandleVoid(c, h.createUC.Execute)
}

func (h *UserHandler) update(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}

	var in userusecase.UpdateUserInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.updateUC.Execute(ctx, id, &in)
	})
}

func (h *UserHandler) delete(c *gin.Context) {
	id, ok := parseUserID(c)
	if !ok {
		return
	}
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.deleteUC.Execute(ctx, id)
	})
}

// parseUserID 解析路径参数中的用户 ID
func parseUserID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		corehandler.WriteBusinessError(c, errors.Sentinel(sharederr.ErrInvalidInput))
		return 0, false
	}
	return id, true
}

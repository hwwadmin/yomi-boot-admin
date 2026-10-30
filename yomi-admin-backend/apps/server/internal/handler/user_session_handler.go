// Package handler 入口适配层：HTTP 路由与插件消费回调
package handler

import (
	"context"
	userusecase "yomi-admin-backend/apps/server/internal/usecase/user"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"

	"github.com/gin-gonic/gin"
)

// UserSessionHandler 用户在线会话路由
type UserSessionHandler struct {
	listUC      *userusecase.ListUserSessionsUsecase
	revokeUC    *userusecase.RevokeUserSessionUsecase
	revokeAllUC *userusecase.RevokeUserSessionsUsecase
}

// NewUserSessionHandler 创建用户会话 Handler
func NewUserSessionHandler(
	listUC *userusecase.ListUserSessionsUsecase,
	revokeUC *userusecase.RevokeUserSessionUsecase,
	revokeAllUC *userusecase.RevokeUserSessionsUsecase,
) *UserSessionHandler {
	return &UserSessionHandler{listUC: listUC, revokeUC: revokeUC, revokeAllUC: revokeAllUC}
}

// Routes 注册用户会话路由
func (h *UserSessionHandler) Routes(router *gin.RouterGroup) {
	router.GET("/api/system/user/session", h.list)
	router.POST("/api/system/user/session/revoke", h.revoke)
	router.POST("/api/system/user/session/revoke-all", h.revokeAll)
}

func (h *UserSessionHandler) list(c *gin.Context) {
	userID, ok := queryID(c, "userId")
	if !ok {
		return
	}
	handle(c, func(ctx context.Context) ([]userusecase.UserSessionItem, error) {
		return h.listUC.Execute(ctx, userID)
	})
}

func (h *UserSessionHandler) revoke(c *gin.Context) {
	var in userusecase.RevokeUserSessionInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	handleVoid(c, func(ctx context.Context) error {
		return h.revokeUC.Execute(ctx, &in)
	})
}

func (h *UserSessionHandler) revokeAll(c *gin.Context) {
	var in userusecase.RevokeUserSessionsInput
	if !corehandler.BindJSON(c, &in) {
		return
	}
	opID, ok := operatorID(c)
	if !ok {
		return
	}
	in.OperatorID = opID
	handleVoid(c, func(ctx context.Context) error {
		return h.revokeAllUC.Execute(ctx, &in)
	})
}

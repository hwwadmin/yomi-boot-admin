// Package handler HTTP 适配层
package handler

import (
	"context"
	rbacusecase "yomi-admin-backend/apps/server/internal/usecase/rbac"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"

	"github.com/gin-gonic/gin"
)

// RbacCacheHandler 鉴权缓存维护路由
type RbacCacheHandler struct {
	refreshUC *rbacusecase.RefreshRbacCacheUsecase
}

// NewRbacCacheHandler 创建鉴权缓存 Handler
func NewRbacCacheHandler(refreshUC *rbacusecase.RefreshRbacCacheUsecase) *RbacCacheHandler {
	return &RbacCacheHandler{refreshUC: refreshUC}
}

// Routes 注册鉴权缓存维护路由
func (h *RbacCacheHandler) Routes(router *gin.RouterGroup) {
	router.POST("/system/rbac-cache/refresh", h.refresh)
}

func (h *RbacCacheHandler) refresh(c *gin.Context) {
	corehandler.Handle(c, func(ctx context.Context) (*rbacusecase.CacheRefreshOutput, error) {
		return h.refreshUC.Execute(ctx)
	})
}

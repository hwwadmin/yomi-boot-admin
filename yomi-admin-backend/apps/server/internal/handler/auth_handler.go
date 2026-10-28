// Package handler HTTP 适配层
package handler

import (
	"yomi-admin-backend/apps/server/internal/usecase"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证相关路由
type AuthHandler struct {
	loginUC   *usecase.LoginUsecase
	profileUC *usecase.ProfileUsecase
	logoutUC  *usecase.LogoutUsecase
}

// NewAuthHandler 创建认证 Handler
func NewAuthHandler(loginUC *usecase.LoginUsecase, profileUC *usecase.ProfileUsecase, logoutUC *usecase.LogoutUsecase) *AuthHandler {
	return &AuthHandler{loginUC: loginUC, profileUC: profileUC, logoutUC: logoutUC}
}

// Routes 注册认证路由
func (h *AuthHandler) Routes(router *gin.RouterGroup) {
	router.POST("/auth/login", h.login)
	router.GET("/auth/profile", h.profile)
	router.POST("/auth/logout", h.logout)
}

func (h *AuthHandler) login(c *gin.Context) {
	corehandler.BindAndHandle(c, h.loginUC.Execute)
}

func (h *AuthHandler) profile(c *gin.Context) {
	corehandler.Handle(c, h.profileUC.Execute)
}

func (h *AuthHandler) logout(c *gin.Context) {
	corehandler.HandleVoid(c, h.logoutUC.Execute)
}

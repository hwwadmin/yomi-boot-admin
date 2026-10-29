// Package handler HTTP 适配层
package handler

import (
	"context"
	authusecase "yomi-admin-backend/apps/server/internal/usecase/auth"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/utils/ginutil"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证相关路由
type AuthHandler struct {
	loginUC    *authusecase.LoginUsecase
	captchaUC  *authusecase.CaptchaUsecase
	profileUC  *authusecase.ProfileUsecase
	logoutUC   *authusecase.LogoutUsecase
	passwordUC *authusecase.ChangePasswordUsecase
}

// NewAuthHandler 创建认证 Handler
func NewAuthHandler(
	loginUC *authusecase.LoginUsecase,
	captchaUC *authusecase.CaptchaUsecase,
	profileUC *authusecase.ProfileUsecase,
	logoutUC *authusecase.LogoutUsecase,
	passwordUC *authusecase.ChangePasswordUsecase,
) *AuthHandler {
	return &AuthHandler{
		loginUC:    loginUC,
		captchaUC:  captchaUC,
		profileUC:  profileUC,
		logoutUC:   logoutUC,
		passwordUC: passwordUC,
	}
}

// Routes 注册认证路由
func (h *AuthHandler) Routes(router *gin.RouterGroup) {
	router.GET("/auth/captcha", h.captcha)
	router.POST("/auth/login", h.login)
	router.GET("/auth/profile", h.profile)
	router.POST("/auth/logout", h.logout)
	router.POST("/auth/password", h.changePassword)
}

func (h *AuthHandler) captcha(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	corehandler.Handle(c, func(ctx context.Context) (*authusecase.CaptchaOutput, error) {
		return h.captchaUC.Generate(ctx, ginutil.ClientIP(c))
	})
}

func (h *AuthHandler) login(c *gin.Context) {
	in := new(authusecase.LoginInput)
	if !corehandler.BindJSON(c, in) {
		return
	}
	clientIP := ginutil.ClientIP(c)
	corehandler.Handle(c, func(ctx context.Context) (*authusecase.LoginOutput, error) {
		return h.loginUC.Execute(ctx, in, clientIP)
	})
}

func (h *AuthHandler) profile(c *gin.Context) {
	userID, ok := operatorID(c)
	if !ok {
		return
	}
	corehandler.Handle(c, func(ctx context.Context) (*authusecase.IdentityView, error) {
		return h.profileUC.Execute(ctx, userID)
	})
}

func (h *AuthHandler) logout(c *gin.Context) {
	corehandler.HandleVoid(c, h.logoutUC.Execute)
}

func (h *AuthHandler) changePassword(c *gin.Context) {
	in := new(authusecase.ChangePasswordInput)
	if !corehandler.BindJSON(c, in) {
		return
	}
	operatorID, ok := operatorID(c)
	if !ok {
		return
	}
	corehandler.HandleVoid(c, func(ctx context.Context) error {
		return h.passwordUC.Execute(ctx, in, operatorID)
	})
}

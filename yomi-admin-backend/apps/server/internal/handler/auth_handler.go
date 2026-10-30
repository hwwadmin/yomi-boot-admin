// Package handler 入口适配层：HTTP 路由与插件消费回调
package handler

import (
	"context"
	"yomi-admin-backend/apps/server/internal/support/audit"
	authusecase "yomi-admin-backend/apps/server/internal/usecase/auth"

	corehandler "gitee.com/hq_project/yomi-boot/component/core/handler"
	"gitee.com/hq_project/yomi-boot/component/core/utils/ginutil"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证相关路由
type AuthHandler struct {
	loginUC         *authusecase.LoginUsecase
	captchaUC       *authusecase.CaptchaUsecase
	profileUC       *authusecase.ProfileUsecase
	updateProfileUC *authusecase.UpdateProfileUsecase
	logoutUC        *authusecase.LogoutUsecase
	passwordUC      *authusecase.ChangePasswordUsecase
	configUC        *authusecase.LoginConfigUsecase
	audit           audit.Recorder
}

// AuthHandlerDeps 认证 Handler 依赖
type AuthHandlerDeps struct {
	Login          *authusecase.LoginUsecase
	Captcha        *authusecase.CaptchaUsecase
	Profile        *authusecase.ProfileUsecase
	UpdateProfile  *authusecase.UpdateProfileUsecase
	Logout         *authusecase.LogoutUsecase
	ChangePassword *authusecase.ChangePasswordUsecase
	LoginConfig    *authusecase.LoginConfigUsecase
	Audit          audit.Recorder
}

// NewAuthHandler 创建认证 Handler
func NewAuthHandler(deps AuthHandlerDeps) *AuthHandler {
	return &AuthHandler{
		loginUC:         deps.Login,
		captchaUC:       deps.Captcha,
		profileUC:       deps.Profile,
		updateProfileUC: deps.UpdateProfile,
		logoutUC:        deps.Logout,
		passwordUC:      deps.ChangePassword,
		configUC:        deps.LoginConfig,
		audit:           deps.Audit,
	}
}

// Routes 注册认证路由
func (h *AuthHandler) Routes(router *gin.RouterGroup) {
	router.GET("/api/auth/captcha", h.captcha)
	router.GET("/api/auth/config", h.loginConfig)
	router.POST("/api/auth/login", h.login)
	router.GET("/api/auth/profile", h.profile)
	router.POST("/api/auth/profile", h.updateProfile)
	router.POST("/api/auth/logout", h.logout)
	router.POST("/api/auth/password", h.changePassword)
}

// loginConfig 下发登录页公开配置（验证码开关等），免认证
func (h *AuthHandler) loginConfig(c *gin.Context) {
	handle(c, func(ctx context.Context) (*authusecase.LoginConfigOutput, error) {
		return h.configUC.Execute(), nil
	})
}

func (h *AuthHandler) captcha(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	handle(c, func(ctx context.Context) (*authusecase.CaptchaOutput, error) {
		return h.captchaUC.Generate(ctx, ginutil.ClientIP(c))
	})
}

func (h *AuthHandler) login(c *gin.Context) {
	in := new(authusecase.LoginInput)
	if !corehandler.BindJSON(c, in) {
		return
	}
	clientIP := ginutil.ClientIP(c)
	userAgent := c.GetHeader("User-Agent")
	start := audit.Spec{
		Event:    audit.EventLogin,
		Operator: audit.Operator{Name: in.Username},
		ClientIP: clientIP,
	}
	finish := func(out *authusecase.LoginOutput, err error) audit.Spec {
		operator := audit.Operator{Name: in.Username}
		if out != nil && out.User != nil {
			operator = audit.Operator{ID: out.User.ID.Int64(), Name: out.User.Username}
		}
		return audit.Spec{Event: audit.EventLogin, Operator: operator, ClientIP: clientIP}
	}
	handle(c, func(ctx context.Context) (*authusecase.LoginOutput, error) {
		return auditRun(c, h.audit, start, finish, func(ctx context.Context) (*authusecase.LoginOutput, error) {
			return h.loginUC.Execute(ctx, in, clientIP, userAgent)
		})
	})
}

func (h *AuthHandler) profile(c *gin.Context) {
	userID, ok := operatorID(c)
	if !ok {
		return
	}
	handle(c, func(ctx context.Context) (*authusecase.IdentityView, error) {
		return h.profileUC.Execute(ctx, userID)
	})
}

func (h *AuthHandler) updateProfile(c *gin.Context) {
	in := new(authusecase.UpdateProfileInput)
	if !corehandler.BindJSON(c, in) {
		return
	}
	userID, ok := operatorID(c)
	if !ok {
		return
	}
	handle(c, func(ctx context.Context) (*authusecase.IdentityView, error) {
		return h.updateProfileUC.Execute(ctx, in, userID)
	})
}

func (h *AuthHandler) logout(c *gin.Context) {
	handleVoid(c, h.logoutUC.Execute)
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
	start := audit.Spec{
		Event:    audit.EventChangePassword,
		Operator: auditOperator(c.Request.Context()),
		ClientIP: ginutil.ClientIP(c),
	}
	handleVoid(c, func(ctx context.Context) error {
		return auditVoid(c, h.audit, start, func(error) audit.Spec {
			return start
		}, func(ctx context.Context) error {
			return h.passwordUC.Execute(ctx, in, operatorID)
		})
	})
}

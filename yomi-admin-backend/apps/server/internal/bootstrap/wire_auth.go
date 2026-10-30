package bootstrap

import (
	apphandler "yomi-admin-backend/apps/server/internal/handler"
	authusecase "yomi-admin-backend/apps/server/internal/usecase/auth"

	"gitee.com/hq_project/yomi-boot/component/core/router"
)

// newAuthRegistrar 组装认证模块：验证码、登录、个人资料与改密
func newAuthRegistrar(c container) router.Registrar {
	return apphandler.NewAuthHandler(apphandler.AuthHandlerDeps{
		Login:          authusecase.NewLoginUsecase(c.users.Users, c.rbac.Auth, c.token, c.captcha, c.loginGuard),
		Captcha:        c.captcha,
		Profile:        authusecase.NewProfileUsecase(c.users.Users, c.rbac.Auth),
		UpdateProfile:  authusecase.NewUpdateProfileUsecase(c.users.Users, c.rbac.Auth),
		Logout:         authusecase.NewLogoutUsecase(c.token),
		ChangePassword: authusecase.NewChangePasswordUsecase(c.users.Users, c.token),
		LoginConfig:    c.loginConfig,
	})
}

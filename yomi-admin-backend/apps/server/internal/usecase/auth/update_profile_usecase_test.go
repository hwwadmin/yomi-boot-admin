package auth

import (
	"context"
	"testing"

	rbacentity "yomi-admin-backend/domains/rbac/pkg/shared/entity"
	rbacenums "yomi-admin-backend/domains/rbac/pkg/shared/enums"
	userentity "yomi-admin-backend/domains/user/pkg/shared/entity"
	userinput "yomi-admin-backend/domains/user/pkg/shared/input"
	usercontract "yomi-admin-backend/domains/user/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

// fakeUpdateProfileUser 只实现编排会调用到的更新资料契约方法
type fakeUpdateProfileUser struct {
	usercontract.UserService
	got *userinput.UpdateProfileInput
	out *userentity.User
	err error
}

func (f *fakeUpdateProfileUser) UpdateProfile(_ context.Context, in *userinput.UpdateProfileInput) (*userentity.User, error) {
	f.got = in
	return f.out, f.err
}

func TestUpdateProfileUsesOperatorAndReturnsIdentity(t *testing.T) {
	users := &fakeUpdateProfileUser{out: &userentity.User{
		ID:       7,
		Username: "admin",
		Nickname: "新昵称",
		Email:    "admin@example.com",
		Phone:    "13900139000",
	}}
	auth := &fakeAuthService{
		permissions: []string{"*"},
		roles:       []*rbacentity.Role{{Code: "super_admin", Name: "超级管理员", Status: rbacenums.StatusEnabled}},
		menus:       []*rbacentity.Permission{{Code: "system:user", Name: "系统用户", UIPath: "/system/user"}},
	}
	uc := NewUpdateProfileUsecase(users, auth)

	view, err := uc.Execute(context.Background(), &UpdateProfileInput{
		Nickname: "新昵称",
		Email:    "admin@example.com",
		Phone:    "13900139000",
	}, 7)
	if err != nil {
		t.Fatalf("更新资料失败: %v", err)
	}
	if users.got == nil || users.got.ID != 7 || users.got.OperatorID != 7 ||
		users.got.Nickname != "新昵称" || users.got.Email != "admin@example.com" || users.got.Phone != "13900139000" {
		t.Fatalf("更新资料入参未透传操作者: %+v", users.got)
	}
	if view.Username != "admin" || view.Nickname != "新昵称" ||
		view.Email != "admin@example.com" || view.Phone != "13900139000" {
		t.Fatalf("返回的身份视图未包含更新后的资料: %+v", view)
	}
	if len(view.Roles) != 1 || view.Roles[0].Name != "超级管理员" || view.Roles[0].Status != rbacenums.StatusEnabled {
		t.Fatalf("返回的身份视图缺少角色: %+v", view)
	}
	if len(view.Permissions) != 1 || len(view.Menus) != 1 {
		t.Fatalf("返回的身份视图缺少权限或菜单: %+v", view)
	}
}

func TestUpdateProfileDomainErrorIsReturned(t *testing.T) {
	users := &fakeUpdateProfileUser{err: errors.Sentinel(usercontract.ErrInvalidInput)}
	uc := NewUpdateProfileUsecase(users, &fakeAuthService{})

	_, err := uc.Execute(context.Background(), &UpdateProfileInput{Nickname: ""}, 7)
	if !errors.Is(err, usercontract.ErrInvalidInput) {
		t.Fatalf("期望 ErrInvalidInput，实际 %v", err)
	}
}

func TestUpdateProfileRejectsMissingOperator(t *testing.T) {
	users := &fakeUpdateProfileUser{}
	uc := NewUpdateProfileUsecase(users, &fakeAuthService{})

	if _, err := uc.Execute(context.Background(), &UpdateProfileInput{Nickname: "新昵称"}, 0); err == nil {
		t.Fatal("缺少操作者应报错")
	}
	if users.got != nil {
		t.Fatalf("缺少操作者不应触达领域服务: %+v", users.got)
	}
}

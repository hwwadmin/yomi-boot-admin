package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
	"yomi-admin-backend/domains/user/internal/repo_contract"
	"yomi-admin-backend/domains/user/pkg/shared/entity"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
	"yomi-admin-backend/domains/user/pkg/shared/input"
	"yomi-admin-backend/domains/user/service_contract"

	"github.com/hwwadmin/yomi-boot/component/core/bc/passwd"
	"github.com/hwwadmin/yomi-boot/component/core/errors"
)

type fakeUserRepo struct {
	users       map[int64]*entity.User
	credentials map[string]repo_contract.UserCredential
	nextID      int64
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{
		users:       make(map[int64]*entity.User),
		credentials: make(map[string]repo_contract.UserCredential),
		nextID:      1,
	}
}

func (r *fakeUserRepo) FindByID(_ context.Context, id int64) (*entity.User, bool, error) {
	user, ok := r.users[id]
	return user, ok, nil
}

func (r *fakeUserRepo) FindUsersByIDs(_ context.Context, ids []int64) ([]*entity.User, error) {
	users := make([]*entity.User, 0, len(ids))
	for _, id := range ids {
		if user, ok := r.users[id]; ok {
			users = append(users, user)
		}
	}
	return users, nil
}

func (r *fakeUserRepo) FindCredentialByUsername(_ context.Context, username string) (*repo_contract.UserCredential, bool, error) {
	cred, ok := r.credentials[username]
	if !ok {
		return nil, false, nil
	}
	return &cred, true, nil
}

func (r *fakeUserRepo) FindCredentialByID(_ context.Context, id int64) (*repo_contract.UserCredential, bool, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, false, nil
	}
	cred, ok := r.credentials[user.Username]
	if !ok {
		return nil, false, nil
	}
	return &cred, true, nil
}

func (r *fakeUserRepo) ExistsByUsername(_ context.Context, username string) (bool, error) {
	_, ok := r.credentials[username]
	return ok, nil
}

func (r *fakeUserRepo) Page(_ context.Context, keyword string, page, pageSize int) ([]*entity.User, int64, error) {
	items := make([]*entity.User, 0, len(r.users))
	for _, user := range r.users {
		if keyword == "" || strings.Contains(user.Username, keyword) {
			items = append(items, user)
		}
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []*entity.User{}, int64(len(items)), nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], int64(len(items)), nil
}

func (r *fakeUserRepo) Create(_ context.Context, user *entity.User, passwordHash string, operatorID int64) error {
	user.CreateTime = time.Now().UnixMilli()
	r.users[user.ID] = user
	r.credentials[user.Username] = repo_contract.UserCredential{User: user, PasswordHash: passwordHash}
	return nil
}

func (r *fakeUserRepo) Update(_ context.Context, user *entity.User, operatorID int64) error {
	r.users[user.ID] = user
	r.credentials[user.Username] = repo_contract.UserCredential{User: user, PasswordHash: r.credentials[user.Username].PasswordHash}
	return nil
}

func (r *fakeUserRepo) SoftDelete(_ context.Context, id int64, operatorID int64) error {
	delete(r.users, id)
	return nil
}

func (r *fakeUserRepo) UpdatePassword(_ context.Context, id int64, passwordHash string, operatorID int64) error {
	user, ok := r.users[id]
	if !ok {
		return fmt.Errorf("user %d not found", id)
	}
	r.credentials[user.Username] = repo_contract.UserCredential{User: user, PasswordHash: passwordHash}
	return nil
}

type seqIDGenerator struct {
	next int64
}

func (g *seqIDGenerator) NextID() (int64, error) {
	g.next++
	return g.next, nil
}

func newTestUserService(repo *fakeUserRepo) *UserService {
	return NewUserService(repo, &seqIDGenerator{next: 1000})
}

func seedUser(t *testing.T, repo *fakeUserRepo, username, password string, status enums.UserStatus) *entity.User {
	t.Helper()

	hash, err := passwd.HashWithCost(password, 4)
	if err != nil {
		t.Fatalf("生成密码哈希失败: %v", err)
	}
	user := &entity.User{
		ID:         repo.nextID,
		Username:   username,
		Nickname:   username,
		Status:     status,
		CreateTime: time.Now().UnixMilli(),
	}
	repo.nextID++
	repo.users[user.ID] = user
	repo.credentials[username] = repo_contract.UserCredential{User: user, PasswordHash: hash}
	return user
}

func TestAuthenticate(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	seedUser(t, repo, "admin", "admin123", enums.UserStatusEnabled)
	seedUser(t, repo, "disabled", "admin123", enums.UserStatusDisabled)
	svc := newTestUserService(repo)

	out, err := svc.Authenticate(ctx, &input.AuthenticateInput{Username: "admin", Password: "admin123"})
	if err != nil {
		t.Fatalf("正确密码应通过: %v", err)
	}
	if out.User.Username != "admin" {
		t.Fatalf("返回用户不匹配: %+v", out.User)
	}

	if _, err := svc.Authenticate(ctx, &input.AuthenticateInput{Username: "admin", Password: "wrong"}); !errors.Is(err, service_contract.ErrInvalidCredential) {
		t.Fatalf("错误密码期望 ErrInvalidCredential，实际 %v", err)
	}
	if _, err := svc.Authenticate(ctx, &input.AuthenticateInput{Username: "missing", Password: "admin123"}); !errors.Is(err, service_contract.ErrInvalidCredential) {
		t.Fatalf("不存在用户期望 ErrInvalidCredential，实际 %v", err)
	}
	if _, err := svc.Authenticate(ctx, &input.AuthenticateInput{Username: "disabled", Password: "admin123"}); !errors.Is(err, service_contract.ErrUserDisabled) {
		t.Fatalf("禁用用户期望 ErrUserDisabled，实际 %v", err)
	}
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	svc := newTestUserService(repo)

	created, err := svc.Create(ctx, &input.CreateUserInput{
		Username:   "test_user",
		Password:   "test1234",
		Nickname:   "测试用户",
		Email:      "test@example.com",
		Phone:      "13800138000",
		Status:     enums.UserStatusEnabled,
		OperatorID: 1,
	})
	if err != nil {
		t.Fatalf("创建合法用户失败: %v", err)
	}
	if created.ID == 0 || created.Username != "test_user" {
		t.Fatalf("创建结果不符合预期: %+v", created)
	}

	cases := []struct {
		name string
		in   *input.CreateUserInput
		want error
	}{
		{
			name: "重复用户名",
			in:   &input.CreateUserInput{Username: "test_user", Password: "test1234", Nickname: "x", Status: enums.UserStatusEnabled},
			want: service_contract.ErrUsernameExists,
		},
		{
			name: "非法用户名",
			in:   &input.CreateUserInput{Username: "a!", Password: "test1234", Nickname: "x", Status: enums.UserStatusEnabled},
			want: service_contract.ErrInvalidInput,
		},
		{
			name: "密码过短",
			in:   &input.CreateUserInput{Username: "new_user", Password: "123", Nickname: "x", Status: enums.UserStatusEnabled},
			want: service_contract.ErrInvalidInput,
		},
		{
			name: "非法邮箱",
			in:   &input.CreateUserInput{Username: "new_user", Password: "test1234", Nickname: "x", Email: "bad", Status: enums.UserStatusEnabled},
			want: service_contract.ErrInvalidInput,
		},
		{
			name: "非法手机号",
			in:   &input.CreateUserInput{Username: "new_user", Password: "test1234", Nickname: "x", Phone: "123", Status: enums.UserStatusEnabled},
			want: service_contract.ErrInvalidInput,
		},
		{
			name: "非法状态",
			in:   &input.CreateUserInput{Username: "new_user", Password: "test1234", Nickname: "x", Status: enums.UserStatus("UNKNOWN")},
			want: service_contract.ErrInvalidInput,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := svc.Create(ctx, tc.in); !errors.Is(err, tc.want) {
				t.Fatalf("期望 %v，实际 %v", tc.want, err)
			}
		})
	}
}

func TestUpdateAndDeleteUser(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	admin := seedUser(t, repo, "admin", "admin123", enums.UserStatusEnabled)
	svc := newTestUserService(repo)

	if err := svc.Update(ctx, &input.UpdateUserInput{
		ID:         admin.ID,
		Nickname:   "新昵称",
		Email:      "admin@example.com",
		Phone:      "13900139000",
		Status:     enums.UserStatusEnabled,
		OperatorID: 999,
	}); err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}
	if admin.Nickname != "新昵称" || admin.Email != "admin@example.com" {
		t.Fatalf("更新未生效: %+v", admin)
	}

	err := svc.Update(ctx, &input.UpdateUserInput{ID: 999, Nickname: "x", Status: enums.UserStatusEnabled, OperatorID: 1})
	if !errors.Is(err, service_contract.ErrUserNotFound) {
		t.Fatalf("更新不存在用户期望 ErrUserNotFound，实际 %v", err)
	}

	if err := svc.Delete(ctx, &input.DeleteUserInput{ID: admin.ID, OperatorID: 999}); err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}
	if _, ok := repo.users[admin.ID]; ok {
		t.Fatal("用户未被删除")
	}
}

func TestUpdateProfileKeepsUsernameAndStatus(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	admin := seedUser(t, repo, "admin", "admin123", enums.UserStatusEnabled)
	svc := newTestUserService(repo)

	user, err := svc.UpdateProfile(ctx, &input.UpdateProfileInput{
		ID:         admin.ID,
		Nickname:   " 新昵称 ",
		Email:      " admin@example.com ",
		Phone:      "13900139000",
		OperatorID: 999,
	})
	if err != nil {
		t.Fatalf("更新自身资料失败: %v", err)
	}
	if user.Nickname != "新昵称" || user.Email != "admin@example.com" || user.Phone != "13900139000" {
		t.Fatalf("资料未按预期裁剪/更新: %+v", user)
	}
	if user.Username != "admin" || user.Status != enums.UserStatusEnabled {
		t.Fatalf("自身资料更新不应改变用户名或状态: %+v", user)
	}

	if _, err := svc.UpdateProfile(ctx, &input.UpdateProfileInput{
		ID: admin.ID, Nickname: "x", Email: "not-an-email",
	}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("非法邮箱期望 ErrInvalidInput，实际 %v", err)
	}
	if _, err := svc.UpdateProfile(ctx, &input.UpdateProfileInput{ID: 999, Nickname: "x"}); !errors.Is(err, service_contract.ErrUserNotFound) {
		t.Fatalf("更新不存在用户期望 ErrUserNotFound，实际 %v", err)
	}
}

func TestChangePassword(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	admin := seedUser(t, repo, "admin", "admin123", enums.UserStatusEnabled)
	svc := newTestUserService(repo)

	cases := []struct {
		name string
		in   *input.ChangePasswordInput
		want error
	}{
		{name: "原密码错误", in: &input.ChangePasswordInput{ID: admin.ID, OldPassword: "wrong", NewPassword: "newpass1"}, want: service_contract.ErrOldPasswordMismatch},
		{name: "新密码过短", in: &input.ChangePasswordInput{ID: admin.ID, OldPassword: "admin123", NewPassword: "123"}, want: service_contract.ErrInvalidInput},
		{name: "新密码与原密码相同", in: &input.ChangePasswordInput{ID: admin.ID, OldPassword: "admin123", NewPassword: "admin123"}, want: service_contract.ErrPasswordUnchanged},
		{name: "用户不存在", in: &input.ChangePasswordInput{ID: 999, OldPassword: "admin123", NewPassword: "newpass1"}, want: service_contract.ErrUserNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := svc.ChangePassword(ctx, tc.in); !errors.Is(err, tc.want) {
				t.Fatalf("期望 %v，实际 %v", tc.want, err)
			}
		})
	}

	// 成功路径：新密码可登录，原密码失效
	if err := svc.ChangePassword(ctx, &input.ChangePasswordInput{ID: admin.ID, OldPassword: "admin123", NewPassword: "newpass1"}); err != nil {
		t.Fatalf("修改密码失败: %v", err)
	}
	if _, err := svc.Authenticate(ctx, &input.AuthenticateInput{Username: "admin", Password: "newpass1"}); err != nil {
		t.Fatalf("新密码应可登录: %v", err)
	}
	if _, err := svc.Authenticate(ctx, &input.AuthenticateInput{Username: "admin", Password: "admin123"}); !errors.Is(err, service_contract.ErrInvalidCredential) {
		t.Fatalf("原密码应失效: %v", err)
	}
}

func TestResetPassword(t *testing.T) {
	ctx := context.Background()
	repo := newFakeUserRepo()
	admin := seedUser(t, repo, "admin", "admin123", enums.UserStatusEnabled)
	svc := newTestUserService(repo)

	if err := svc.ResetPassword(ctx, &input.ResetPasswordInput{ID: admin.ID, NewPassword: "reset123", OperatorID: 999}); err != nil {
		t.Fatalf("重置密码失败: %v", err)
	}
	if _, err := svc.Authenticate(ctx, &input.AuthenticateInput{Username: "admin", Password: "reset123"}); err != nil {
		t.Fatalf("重置后的密码应可登录: %v", err)
	}

	if err := svc.ResetPassword(ctx, &input.ResetPasswordInput{ID: 999, NewPassword: "reset123", OperatorID: 999}); !errors.Is(err, service_contract.ErrUserNotFound) {
		t.Fatalf("重置不存在用户期望 ErrUserNotFound，实际 %v", err)
	}
	if err := svc.ResetPassword(ctx, &input.ResetPasswordInput{ID: admin.ID, NewPassword: "123", OperatorID: 999}); !errors.Is(err, service_contract.ErrInvalidInput) {
		t.Fatalf("重置密码过短期望 ErrInvalidInput，实际 %v", err)
	}
}

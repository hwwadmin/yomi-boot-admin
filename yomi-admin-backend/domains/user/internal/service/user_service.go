// Package service 用户域服务实现
package service

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"
	"yomi-admin-backend/domains/user/internal/repo_contract"
	"yomi-admin-backend/domains/user/pkg/shared/entity"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
	"yomi-admin-backend/domains/user/pkg/shared/input"
	"yomi-admin-backend/domains/user/pkg/shared/output"
	"yomi-admin-backend/domains/user/service_contract"

	"gitee.com/hq_project/yomi-boot/component/common/snowflake"
	"gitee.com/hq_project/yomi-boot/component/core/bc/passwd"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/utils/validateutil"
)

const (
	minPasswordBytes = 6
	maxNicknameRunes = 32
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{4,32}$`)

// UserService 用户领域服务实现
type UserService struct {
	users       repo_contract.UserRepo
	idGenerator snowflake.Generator
}

var _ service_contract.UserService = (*UserService)(nil)

// NewUserService 创建用户领域服务
func NewUserService(users repo_contract.UserRepo, idGenerator snowflake.Generator) *UserService {
	return &UserService{users: users, idGenerator: idGenerator}
}

// Authenticate 校验账号密码
func (s *UserService) Authenticate(ctx context.Context, in *input.AuthenticateInput) (*output.AuthenticateOutput, error) {
	username := strings.TrimSpace(in.Username)
	if username == "" || in.Password == "" {
		return nil, errors.Sentinel(service_contract.ErrInvalidCredential)
	}

	cred, found, err := s.users.FindCredentialByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Sentinel(service_contract.ErrInvalidCredential)
	}

	if err := passwd.Verify(cred.PasswordHash, in.Password); err != nil {
		if errors.Is(err, passwd.ErrMismatch) || errors.Is(err, passwd.ErrPasswordTooLong) {
			return nil, errors.Sentinel(service_contract.ErrInvalidCredential)
		}
		return nil, errors.Wrap(err, "校验密码失败")
	}

	if cred.User.Status != enums.UserStatusEnabled {
		return nil, errors.Sentinel(service_contract.ErrUserDisabled)
	}

	return &output.AuthenticateOutput{User: cred.User}, nil
}

// GetByID 按用户 ID 查询
func (s *UserService) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	if id <= 0 {
		return nil, errors.Sentinel(service_contract.ErrUserNotFound)
	}

	user, found, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Sentinel(service_contract.ErrUserNotFound)
	}
	return user, nil
}

// Page 分页查询用户
func (s *UserService) Page(ctx context.Context, in *input.PageUsersInput) (*output.PageUsersOutput, error) {
	items, total, err := s.users.Page(ctx, in.Keyword, in.Page, in.PageSize)
	if err != nil {
		return nil, err
	}
	return &output.PageUsersOutput{Items: items, Total: total}, nil
}

// Create 创建用户
func (s *UserService) Create(ctx context.Context, in *input.CreateUserInput) (*entity.User, error) {
	username := strings.TrimSpace(in.Username)
	if !usernamePattern.MatchString(username) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if err := validatePassword(in.Password); err != nil {
		return nil, err
	}

	nickname, email, phone, err := normalizeProfile(in.Nickname, in.Email, in.Phone)
	if err != nil {
		return nil, err
	}

	if !enums.IsValidUserStatus(in.Status) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	exists, err := s.users.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Sentinel(service_contract.ErrUsernameExists)
	}

	passwordHash, err := hashPassword(in.Password)
	if err != nil {
		return nil, err
	}

	id, err := s.idGenerator.NextID()
	if err != nil {
		return nil, errors.Wrap(err, "生成用户 ID 失败")
	}

	user := &entity.User{
		ID:       id,
		Username: username,
		Nickname: nickname,
		Email:    email,
		Phone:    phone,
		Status:   in.Status,
	}
	if err := s.users.Create(ctx, user, passwordHash, in.OperatorID); err != nil {
		if errors.Is(err, repo_contract.ErrDuplicatedKey) {
			return nil, errors.Sentinel(service_contract.ErrUsernameExists)
		}
		return nil, err
	}
	return user, nil
}

// Update 更新用户资料与状态；调用方负责最后一名超管与自操作校验
func (s *UserService) Update(ctx context.Context, in *input.UpdateUserInput) error {
	if in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrUserNotFound)
	}
	if !enums.IsValidUserStatus(in.Status) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	nickname, email, phone, err := normalizeProfile(in.Nickname, in.Email, in.Phone)
	if err != nil {
		return err
	}

	user, found, err := s.users.FindByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrUserNotFound)
	}

	user.Nickname = nickname
	user.Email = email
	user.Phone = phone
	user.Status = in.Status
	return s.users.Update(ctx, user, in.OperatorID)
}

// UpdateProfile 更新用户自身资料，不改变用户名与状态
func (s *UserService) UpdateProfile(ctx context.Context, in *input.UpdateProfileInput) (*entity.User, error) {
	if in.ID <= 0 {
		return nil, errors.Sentinel(service_contract.ErrUserNotFound)
	}

	nickname, email, phone, err := normalizeProfile(in.Nickname, in.Email, in.Phone)
	if err != nil {
		return nil, err
	}

	user, found, err := s.users.FindByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.Sentinel(service_contract.ErrUserNotFound)
	}

	user.Nickname = nickname
	user.Email = email
	user.Phone = phone
	if err := s.users.Update(ctx, user, in.OperatorID); err != nil {
		return nil, err
	}
	return user, nil
}

// Delete 软删除用户；调用方负责最后一名超管与自操作校验
func (s *UserService) Delete(ctx context.Context, in *input.DeleteUserInput) error {
	if in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrUserNotFound)
	}

	_, found, err := s.users.FindByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrUserNotFound)
	}
	return s.users.SoftDelete(ctx, in.ID, in.OperatorID)
}

// ChangePassword 校验原密码后修改密码；调用方负责撤销该用户的登录令牌
func (s *UserService) ChangePassword(ctx context.Context, in *input.ChangePasswordInput) error {
	if in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrUserNotFound)
	}
	if err := validatePassword(in.NewPassword); err != nil {
		return err
	}

	cred, found, err := s.users.FindCredentialByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrUserNotFound)
	}

	// 原密码错误与超长都统一按“原密码不正确”返回，不额外暴露密码长度规则
	if err := passwd.Verify(cred.PasswordHash, in.OldPassword); err != nil {
		if errors.Is(err, passwd.ErrMismatch) || errors.Is(err, passwd.ErrPasswordTooLong) {
			return errors.Sentinel(service_contract.ErrOldPasswordMismatch)
		}
		return errors.Wrap(err, "校验原密码失败")
	}
	if in.NewPassword == in.OldPassword {
		return errors.Sentinel(service_contract.ErrPasswordUnchanged)
	}

	passwordHash, err := hashPassword(in.NewPassword)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, in.ID, passwordHash, in.ID)
}

// ResetPassword 直接重置密码；调用方负责超管权限校验与撤销该用户的登录令牌
func (s *UserService) ResetPassword(ctx context.Context, in *input.ResetPasswordInput) error {
	if in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrUserNotFound)
	}
	if err := validatePassword(in.NewPassword); err != nil {
		return err
	}

	_, found, err := s.users.FindByID(ctx, in.ID)
	if err != nil {
		return err
	}
	if !found {
		return errors.Sentinel(service_contract.ErrUserNotFound)
	}

	passwordHash, err := hashPassword(in.NewPassword)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, in.ID, passwordHash, in.OperatorID)
}

// normalizeProfile 统一裁剪并校验昵称、邮箱与手机号
func normalizeProfile(nickname, email, phone string) (string, string, string, error) {
	nickname = strings.TrimSpace(nickname)
	if nickname == "" || utf8.RuneCountInString(nickname) > maxNicknameRunes {
		return "", "", "", errors.Sentinel(service_contract.ErrInvalidInput)
	}

	email = strings.TrimSpace(email)
	if email != "" && !validateutil.IsValidEmail(email) {
		return "", "", "", errors.Sentinel(service_contract.ErrInvalidInput)
	}

	phone = strings.TrimSpace(phone)
	if phone != "" && !validateutil.IsValidPhone(phone) {
		return "", "", "", errors.Sentinel(service_contract.ErrInvalidInput)
	}

	return nickname, email, phone, nil
}

// validatePassword 校验密码长度：下限按业务设定，上限取自 bcrypt 的 72 字节
func validatePassword(password string) error {
	if len([]byte(password)) < minPasswordBytes || len([]byte(password)) > passwd.MaxPasswordBytes {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	return nil
}

// hashPassword 生成密码哈希；超长按非法入参处理
func hashPassword(password string) (string, error) {
	passwordHash, err := passwd.Hash(password)
	if err != nil {
		if errors.Is(err, passwd.ErrPasswordTooLong) {
			return "", errors.Sentinel(service_contract.ErrInvalidInput)
		}
		return "", errors.Wrap(err, "生成密码哈希失败")
	}
	return passwordHash, nil
}

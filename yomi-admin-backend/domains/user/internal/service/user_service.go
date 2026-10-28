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

	"gitee.com/hq_project/yomi-boot/component/core/bc/passwd"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	"gitee.com/hq_project/yomi-boot/component/core/shared"
	"gitee.com/hq_project/yomi-boot/component/core/utils/validateutil"
)

const (
	minPasswordBytes = 6
	maxNicknameRunes = 32
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{4,32}$`)

// UserService 用户领域服务实现
type UserService struct {
	users repo_contract.UserRepo
}

var _ service_contract.UserService = (*UserService)(nil)

// NewUserService 创建用户领域服务
func NewUserService(users repo_contract.UserRepo) *UserService {
	return &UserService{users: users}
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
	page, pageSize := normalizePage(in.Page, in.PageSize)

	items, total, err := s.users.Page(ctx, in.Keyword, page, pageSize)
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
	if len([]byte(in.Password)) < minPasswordBytes || len([]byte(in.Password)) > passwd.MaxPasswordBytes {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	nickname := strings.TrimSpace(in.Nickname)
	if nickname == "" || utf8.RuneCountInString(nickname) > maxNicknameRunes {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	email := strings.TrimSpace(in.Email)
	if email != "" && !validateutil.IsValidEmail(email) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	phone := strings.TrimSpace(in.Phone)
	if phone != "" && !validateutil.IsValidPhone(phone) {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	if !in.Status.Valid() {
		return nil, errors.Sentinel(service_contract.ErrInvalidInput)
	}

	exists, err := s.users.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Sentinel(service_contract.ErrUsernameExists)
	}

	passwordHash, err := passwd.Hash(in.Password)
	if err != nil {
		if errors.Is(err, passwd.ErrPasswordTooLong) {
			return nil, errors.Sentinel(service_contract.ErrInvalidInput)
		}
		return nil, errors.Wrap(err, "生成密码哈希失败")
	}

	user := &entity.User{
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

// Update 更新用户资料与状态
func (s *UserService) Update(ctx context.Context, in *input.UpdateUserInput) error {
	if in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrUserNotFound)
	}
	if !in.Status.Valid() {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}
	if in.ID == in.OperatorID && in.Status == enums.UserStatusDisabled {
		return errors.Sentinel(service_contract.ErrCannotOperateSelf)
	}

	nickname := strings.TrimSpace(in.Nickname)
	if nickname == "" || utf8.RuneCountInString(nickname) > maxNicknameRunes {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	email := strings.TrimSpace(in.Email)
	if email != "" && !validateutil.IsValidEmail(email) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
	}

	phone := strings.TrimSpace(in.Phone)
	if phone != "" && !validateutil.IsValidPhone(phone) {
		return errors.Sentinel(service_contract.ErrInvalidInput)
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

// Delete 软删除用户
func (s *UserService) Delete(ctx context.Context, in *input.DeleteUserInput) error {
	if in.ID <= 0 {
		return errors.Sentinel(service_contract.ErrUserNotFound)
	}
	if in.ID == in.OperatorID {
		return errors.Sentinel(service_contract.ErrCannotOperateSelf)
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

// normalizePage 填充分页默认值并限制每页上限
func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = shared.DefaultPage
	}
	if pageSize <= 0 {
		pageSize = shared.DefaultPageSize
	}
	if pageSize > shared.MaxPageSize {
		pageSize = shared.MaxPageSize
	}
	return page, pageSize
}

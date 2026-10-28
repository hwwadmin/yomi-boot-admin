package repo

import (
	"context"
	"testing"
	"yomi-admin-backend/domains/user/internal/repo_contract"
	"yomi-admin-backend/domains/user/pkg/shared/entity"
	"yomi-admin-backend/domains/user/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/common/db/dbtest"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	sharedenums "gitee.com/hq_project/yomi-boot/component/core/shared/enums"
)

func TestUserRepoCRUD(t *testing.T) {
	gormDB, cleanup := dbtest.New(t, dbtest.WithTables(&userModel{}))
	defer cleanup()

	repo := NewUserRepo(gormDB, sharedenums.MySQL)
	ctx := context.Background()

	user := &entity.User{
		ID:       100000000000000001,
		Username: "admin",
		Nickname: "管理员",
		Email:    "admin@example.com",
		Phone:    "13800138000",
		Status:   enums.UserStatusEnabled,
	}
	if err := repo.Create(ctx, user, "hash", 100); err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	if user.ID != 100000000000000001 {
		t.Fatalf("雪花 ID 未写入: %d", user.ID)
	}
	if user.CreateTime == 0 {
		t.Fatal("创建后未回填创建时间")
	}

	found, ok, err := repo.FindByID(ctx, user.ID)
	if err != nil || !ok {
		t.Fatalf("按 ID 查询失败: ok=%v err=%v", ok, err)
	}
	if found.Username != "admin" {
		t.Fatalf("查询结果不匹配: %+v", found)
	}

	cred, ok, err := repo.FindCredentialByUsername(ctx, "admin")
	if err != nil || !ok {
		t.Fatalf("按用户名查询凭证失败: ok=%v err=%v", ok, err)
	}
	if cred.PasswordHash != "hash" {
		t.Fatalf("密码哈希不匹配: %s", cred.PasswordHash)
	}

	items, total, err := repo.Page(ctx, "adm", 1, 10)
	if err != nil {
		t.Fatalf("分页查询失败: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("分页结果不符合预期: total=%d items=%d", total, len(items))
	}

	user.Nickname = "新昵称"
	user.Status = enums.UserStatusDisabled
	if err := repo.Update(ctx, user, 100); err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}
	updated, ok, err := repo.FindByID(ctx, user.ID)
	if err != nil || !ok {
		t.Fatalf("更新后查询失败: ok=%v err=%v", ok, err)
	}
	if updated.Nickname != "新昵称" || updated.Status != enums.UserStatusDisabled {
		t.Fatalf("更新未生效: %+v", updated)
	}

	duplicate := &entity.User{ID: 100000000000000002, Username: "admin", Nickname: "重复", Status: enums.UserStatusEnabled}
	err = repo.Create(ctx, duplicate, "hash", 100)
	if !errors.Is(err, repo_contract.ErrDuplicatedKey) {
		t.Fatalf("重复用户名期望 ErrDuplicatedKey，实际 %v", err)
	}

	if err := repo.SoftDelete(ctx, user.ID, 100); err != nil {
		t.Fatalf("软删除失败: %v", err)
	}
	if _, ok, err := repo.FindByID(ctx, user.ID); err != nil || ok {
		t.Fatalf("软删除后不应查到: ok=%v err=%v", ok, err)
	}

	exists, err := repo.ExistsByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("检查用户名失败: %v", err)
	}
	if !exists {
		t.Fatal("软删除后用户名仍应占用")
	}
}

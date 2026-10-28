// Package repo 用户域仓储实现
package repo

import (
	"yomi-admin-backend/domains/user/pkg/shared/entity"

	"gitee.com/hq_project/yomi-boot/component/core/utils/ptrutil"
)

// toUserEntity 将持久化模型转换为领域实体
func toUserEntity(row *userModel) *entity.User {
	if row == nil {
		return nil
	}
	return &entity.User{
		ID:         row.ID,
		Username:   row.Username,
		Nickname:   row.Nickname,
		Email:      row.Email,
		Phone:      row.Phone,
		Status:     row.Status,
		CreateTime: row.CreateTime.UnixMilli(),
		UpdateTime: row.UpdateTime.UnixMilli(),
	}
}

// toUserModel 将领域实体转换为持久化模型
func toUserModel(user *entity.User, passwordHash string, operatorID int64) *userModel {
	row := &userModel{
		Username:     user.Username,
		PasswordHash: passwordHash,
		Nickname:     user.Nickname,
		Email:        user.Email,
		Phone:        user.Phone,
		Status:       user.Status,
	}
	row.ID = user.ID
	row.CreateBy = ptrutil.PtrFromInt64(operatorID)
	row.UpdateBy = ptrutil.PtrFromInt64(operatorID)
	return row
}

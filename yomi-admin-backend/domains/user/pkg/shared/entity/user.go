// Package entity 用户域实体
package entity

import (
	"time"
	"yomi-admin-backend/domains/user/pkg/shared/enums"
)

// User 表示系统用户（非持久化模型）
type User struct {
	ID        int64
	Username  string
	Nickname  string
	Email     string
	Phone     string
	Status    enums.UserStatus
	CreatedAt time.Time
}

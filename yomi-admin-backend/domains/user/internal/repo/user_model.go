// Package repo 用户域仓储实现
package repo

import (
	"yomi-admin-backend/domains/user/pkg/shared/enums"

	"gitee.com/hq_project/yomi-boot/component/common/db/orm"
)

// tableName 是用户表名
const tableName = "t_user"

// userModel 是 t_user 的持久化模型
type userModel struct {
	Username     string           `gorm:"column:username;size:64;uniqueIndex:uk_username"`
	PasswordHash string           `gorm:"column:password_hash;size:255"`
	Nickname     string           `gorm:"column:nickname;size:64"`
	Email        string           `gorm:"column:email;size:128"`
	Phone        string           `gorm:"column:phone;size:32"`
	Status       enums.UserStatus `gorm:"column:status;size:20"`
	orm.BaseModel
}

// TableName 返回表名
func (userModel) TableName() string {
	return tableName
}

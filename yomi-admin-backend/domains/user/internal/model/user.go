// Package model 用户域持久化模型
package model

import "gitee.com/hq_project/yomi-boot/component/common/db/orm"

// User 系统用户表映射
type User struct {
	orm.BaseModel
	Username     string `gorm:"column:username;uniqueIndex:uk_sys_user_username"`
	PasswordHash string `gorm:"column:password_hash"`
	Nickname     string `gorm:"column:nickname"`
	Email        string `gorm:"column:email"`
	Phone        string `gorm:"column:phone"`
	Status       int8   `gorm:"column:status"`
}

// TableName 返回表名
func (User) TableName() string {
	return "sys_user"
}

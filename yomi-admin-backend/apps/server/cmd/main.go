// Package main 管理后台编排层入口
package main

import (
	"yomi-admin-backend/apps/server/internal/bootstrap"

	"gitee.com/hq_project/yomi-boot/boot"
)

func main() {
	boot.RunWithOptions(bootstrap.Options(), bootstrap.Wire)
}

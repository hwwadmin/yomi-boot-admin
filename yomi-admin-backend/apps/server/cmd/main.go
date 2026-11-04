// Package main 管理后台编排层入口
package main

import (
	"yomi-admin-backend/apps/server/internal/bootstrap"

	"github.com/hwwadmin/yomi-boot/boot"
)

func main() {
	boot.RunWithOptions(bootstrap.Options(), bootstrap.Wire)
}

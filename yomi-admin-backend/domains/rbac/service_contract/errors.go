// Package service_contract RBAC 域服务契约
package service_contract

import "gitee.com/hq_project/yomi-boot/component/core/errors"

var (
	// ErrPermissionDenied 没有访问权限
	ErrPermissionDenied = errors.NewBiz("ADMIN_PERMISSION_DENIED", "没有访问权限")
	// ErrRoleNotFound 角色不存在
	ErrRoleNotFound = errors.NewBiz("ADMIN_ROLE_NOT_FOUND", "角色不存在")
	// ErrPermissionNotFound 权限不存在
	ErrPermissionNotFound = errors.NewBiz("ADMIN_PERMISSION_NOT_FOUND", "权限不存在")
	// ErrPermissionGroupNotFound 权限组不存在
	ErrPermissionGroupNotFound = errors.NewBiz("ADMIN_PERMISSION_GROUP_NOT_FOUND", "权限组不存在")
	// ErrRoleCodeExists 角色编码已存在
	ErrRoleCodeExists = errors.NewBiz("ADMIN_ROLE_CODE_EXISTS", "角色编码已存在")
	// ErrPermissionGroupCodeExists 权限组编码已存在
	ErrPermissionGroupCodeExists = errors.NewBiz("ADMIN_PERMISSION_GROUP_CODE_EXISTS", "权限组编码已存在")
	// ErrPermissionCodeExists 权限编码已存在
	ErrPermissionCodeExists = errors.NewBiz("ADMIN_PERMISSION_CODE_EXISTS", "权限编码已存在")
	// ErrPermissionResourceExists 权限资源已存在
	ErrPermissionResourceExists = errors.NewBiz("ADMIN_PERMISSION_RESOURCE_EXISTS", "权限资源已存在")
	// ErrRoleBoundToUsers 角色仍被用户绑定
	ErrRoleBoundToUsers = errors.NewBiz("ADMIN_ROLE_BOUND_TO_USERS", "角色已绑定用户，无法删除")
	// ErrPermissionGroupBoundToRoles 权限组仍被角色绑定
	ErrPermissionGroupBoundToRoles = errors.NewBiz("ADMIN_PERMISSION_GROUP_BOUND_TO_ROLES", "权限组已绑定角色，无法删除")
	// ErrBuiltinRoleImmutable 内置角色不允许变更或删除
	ErrBuiltinRoleImmutable = errors.NewBiz("ADMIN_BUILTIN_ROLE_IMMUTABLE", "内置角色不允许变更或删除")
	// ErrRelatedDisabled 关联的角色、权限组或权限已停用
	ErrRelatedDisabled = errors.NewBiz("ADMIN_RELATED_DISABLED", "关联的角色、权限组或权限已停用")
	// ErrInvalidInput 参数不合法
	ErrInvalidInput = errors.NewBiz("ADMIN_INVALID_INPUT", "参数不合法")
)

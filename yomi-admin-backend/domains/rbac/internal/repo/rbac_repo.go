// Package repo RBAC 域仓储实现
package repo

import (
	"context"
	"strings"
	"yomi-admin-backend/domains/rbac/internal/repo_contract"
	"yomi-admin-backend/domains/rbac/pkg/shared/entity"
	"yomi-admin-backend/domains/rbac/pkg/shared/enums"

	yomidb "gitee.com/hq_project/yomi-boot/component/common/db"
	"gitee.com/hq_project/yomi-boot/component/common/db/orm"
	"gitee.com/hq_project/yomi-boot/component/common/db/tx"
	"gitee.com/hq_project/yomi-boot/component/core/errors"
	sharedenums "gitee.com/hq_project/yomi-boot/component/core/shared/enums"
	"gitee.com/hq_project/yomi-boot/component/core/utils/ptrutil"

	"gorm.io/gorm"
)

// 鉴权查询 SQL；用户 → 角色 → 权限组 → 权限 的链路用原生 SQL 收口，参数全部占位符
const (
	// sqlUserHasEnabledSuperAdmin 判断用户是否拥有启用态超管角色
	sqlUserHasEnabledSuperAdmin = `
SELECT 1
FROM t_rbac_user_role_rel ur
JOIN t_rbac_role r ON r.id = ur.role_id AND r.delete_time IS NULL
WHERE ur.user_id = ? AND r.code = ? AND r.status = ?
LIMIT 1`

	// sqlListSuperAdminUserIDs 列出拥有启用态超管角色的用户 ID
	sqlListSuperAdminUserIDs = `
SELECT DISTINCT ur.user_id
FROM t_rbac_user_role_rel ur
JOIN t_rbac_role r ON r.id = ur.role_id AND r.delete_time IS NULL
WHERE r.code = ? AND r.status = ?
ORDER BY ur.user_id`

	// sqlExistsRouteAccess 判断用户是否可通过启用角色 → 启用权限组 → 启用权限访问接口
	sqlExistsRouteAccess = `
SELECT 1
FROM t_rbac_user_role_rel ur
JOIN t_rbac_role r ON r.id = ur.role_id AND r.delete_time IS NULL
JOIN t_rbac_role_group_rel rg ON rg.role_id = r.id
JOIN t_rbac_permission_group g ON g.id = rg.group_id AND g.delete_time IS NULL
JOIN t_rbac_group_permission_rel gp ON gp.group_id = g.id
JOIN t_rbac_permission p ON p.id = gp.permission_id AND p.delete_time IS NULL
WHERE ur.user_id = ?
  AND r.status = ?
  AND g.status = ?
  AND p.status = ?
  AND p.permission_type = ?
  AND p.http_method = ?
  AND p.http_path = ?
LIMIT 1`

	// sqlListMenusByUserID 查询用户可见的启用菜单
	sqlListMenusByUserID = `
SELECT DISTINCT p.*
FROM t_rbac_user_role_rel ur
JOIN t_rbac_role r ON r.id = ur.role_id AND r.delete_time IS NULL
JOIN t_rbac_role_group_rel rg ON rg.role_id = r.id
JOIN t_rbac_permission_group g ON g.id = rg.group_id AND g.delete_time IS NULL
JOIN t_rbac_group_permission_rel gp ON gp.group_id = g.id
JOIN t_rbac_permission p ON p.id = gp.permission_id AND p.delete_time IS NULL
WHERE ur.user_id = ?
  AND r.status = ?
  AND g.status = ?
  AND p.status = ?
  AND p.permission_type = ?
ORDER BY p.id`

	// sqlListPermissionCodesByUserID 查询用户拥有的启用权限编码
	sqlListPermissionCodesByUserID = `
SELECT DISTINCT p.code
FROM t_rbac_user_role_rel ur
JOIN t_rbac_role r ON r.id = ur.role_id AND r.delete_time IS NULL
JOIN t_rbac_role_group_rel rg ON rg.role_id = r.id
JOIN t_rbac_permission_group g ON g.id = rg.group_id AND g.delete_time IS NULL
JOIN t_rbac_group_permission_rel gp ON gp.group_id = g.id
JOIN t_rbac_permission p ON p.id = gp.permission_id AND p.delete_time IS NULL
WHERE ur.user_id = ?
  AND r.status = ?
  AND g.status = ?
  AND p.status = ?
ORDER BY p.code`

	// sqlListPermissionIDsByRoleID 查询角色所绑权限组的权限 ID（去重聚合，不看启停）
	sqlListPermissionIDsByRoleID = `
SELECT DISTINCT gp.permission_id
FROM t_rbac_role_group_rel rg
JOIN t_rbac_group_permission_rel gp ON gp.group_id = rg.group_id
JOIN t_rbac_permission p ON p.id = gp.permission_id AND p.delete_time IS NULL
WHERE rg.role_id = ?
ORDER BY gp.permission_id`
)

// RbacRepo 角色、权限组、权限与关系仓储实现
type RbacRepo struct {
	orm.BaseRepository
	dbType sharedenums.DbType
}

var _ repo_contract.RbacRepo = (*RbacRepo)(nil)

// NewRbacRepo 创建 RBAC 仓储
func NewRbacRepo(d *gorm.DB, dbType sharedenums.DbType) *RbacRepo {
	return &RbacRepo{
		BaseRepository: orm.BaseRepository{DB: d},
		dbType:         dbType,
	}
}

// FindRoleByID 按主键查询角色
func (r *RbacRepo) FindRoleByID(ctx context.Context, id int64) (*entity.Role, bool, error) {
	var row roleModel
	found, err := r.FirstByID(ctx, id, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按主键查询角色失败")
	}
	if !found {
		return nil, false, nil
	}
	return toRoleEntity(&row), true, nil
}

// FindRoleByCode 按编码查询角色
func (r *RbacRepo) FindRoleByCode(ctx context.Context, code string) (*entity.Role, bool, error) {
	var row roleModel
	found, err := r.FirstByField(ctx, "code", code, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按编码查询角色失败")
	}
	if !found {
		return nil, false, nil
	}
	return toRoleEntity(&row), true, nil
}

// FindRolesByIDs 按主键批量查询角色
func (r *RbacRepo) FindRolesByIDs(ctx context.Context, ids []int64) ([]*entity.Role, error) {
	var rows []*roleModel
	if err := r.FindByIDs(ctx, ids, &roleModel{}, &rows); err != nil {
		return nil, errors.Wrap(err, "批量查询角色失败")
	}
	items := make([]*entity.Role, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRoleEntity(row))
	}
	return items, nil
}

// ListAllRoles 列出全部角色（含禁用）
func (r *RbacRepo) ListAllRoles(ctx context.Context) ([]*entity.Role, error) {
	var rows []*roleModel
	if err := tx.Exec(ctx, r.DB).Model(&roleModel{}).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, errors.Wrap(err, "查询全部角色失败")
	}
	items := make([]*entity.Role, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRoleEntity(row))
	}
	return items, nil
}

// PageRoles 分页查询角色
func (r *RbacRepo) PageRoles(ctx context.Context, keyword string, status enums.Status, page, pageSize int) ([]*entity.Role, int64, error) {
	query := tx.Exec(ctx, r.DB).Model(&roleModel{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := orm.LikeFull(keyword)
		query = query.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var rows []*roleModel
	total, err := r.CountPage(ctx, query, page, pageSize, &rows, orm.OrderAsc("id"))
	if err != nil {
		return nil, 0, errors.Wrap(err, "分页查询角色失败")
	}

	items := make([]*entity.Role, 0, len(rows))
	for _, row := range rows {
		items = append(items, toRoleEntity(row))
	}
	return items, total, nil
}

// CreateRole 创建角色
func (r *RbacRepo) CreateRole(ctx context.Context, role *entity.Role, operatorID int64) error {
	if err := r.Create(ctx, toRoleModel(role, operatorID)); err != nil {
		if yomidb.IsDuplicateKey(err, r.dbType) {
			return errors.Sentinel(repo_contract.ErrDuplicatedKey)
		}
		return err
	}
	return nil
}

// UpdateRole 更新角色资料
func (r *RbacRepo) UpdateRole(ctx context.Context, role *entity.Role, operatorID int64) error {
	updates := map[string]any{
		"name":      role.Name,
		"remark":    role.Remark,
		"update_by": ptrutil.PtrFromInt64(operatorID),
	}
	if err := r.UpdateByID(ctx, role.ID, &roleModel{}, updates); err != nil {
		return errors.Wrap(err, "更新角色失败")
	}
	return nil
}

// UpdateRoleStatus 更新角色启停
func (r *RbacRepo) UpdateRoleStatus(ctx context.Context, id int64, status enums.Status, operatorID int64) error {
	updates := map[string]any{
		"status":    status,
		"update_by": ptrutil.PtrFromInt64(operatorID),
	}
	if err := r.UpdateByID(ctx, id, &roleModel{}, updates); err != nil {
		return errors.Wrap(err, "更新角色状态失败")
	}
	return nil
}

// DeleteRole 硬删除角色
func (r *RbacRepo) DeleteRole(ctx context.Context, id int64) error {
	if err := r.DeleteByID(ctx, id, &roleModel{}); err != nil {
		return errors.Wrap(err, "删除角色失败")
	}
	return nil
}

// FindPermissionGroupByID 按主键查询权限组
func (r *RbacRepo) FindPermissionGroupByID(ctx context.Context, id int64) (*entity.PermissionGroup, bool, error) {
	var row permissionGroupModel
	found, err := r.FirstByID(ctx, id, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按主键查询权限组失败")
	}
	if !found {
		return nil, false, nil
	}
	return toPermissionGroupEntity(&row), true, nil
}

// FindPermissionGroupByCode 按编码查询权限组
func (r *RbacRepo) FindPermissionGroupByCode(ctx context.Context, code string) (*entity.PermissionGroup, bool, error) {
	var row permissionGroupModel
	found, err := r.FirstByField(ctx, "code", code, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按编码查询权限组失败")
	}
	if !found {
		return nil, false, nil
	}
	return toPermissionGroupEntity(&row), true, nil
}

// FindPermissionGroupsByIDs 按主键批量查询权限组
func (r *RbacRepo) FindPermissionGroupsByIDs(ctx context.Context, ids []int64) ([]*entity.PermissionGroup, error) {
	var rows []*permissionGroupModel
	if err := r.FindByIDs(ctx, ids, &permissionGroupModel{}, &rows); err != nil {
		return nil, errors.Wrap(err, "批量查询权限组失败")
	}
	items := make([]*entity.PermissionGroup, 0, len(rows))
	for _, row := range rows {
		items = append(items, toPermissionGroupEntity(row))
	}
	return items, nil
}

// PagePermissionGroups 分页查询权限组
func (r *RbacRepo) PagePermissionGroups(ctx context.Context, keyword string, status enums.Status, page, pageSize int) ([]*entity.PermissionGroup, int64, error) {
	query := tx.Exec(ctx, r.DB).Model(&permissionGroupModel{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := orm.LikeFull(keyword)
		query = query.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var rows []*permissionGroupModel
	total, err := r.CountPage(ctx, query, page, pageSize, &rows, orm.OrderAsc("id"))
	if err != nil {
		return nil, 0, errors.Wrap(err, "分页查询权限组失败")
	}

	items := make([]*entity.PermissionGroup, 0, len(rows))
	for _, row := range rows {
		items = append(items, toPermissionGroupEntity(row))
	}
	return items, total, nil
}

// CreatePermissionGroup 创建权限组
func (r *RbacRepo) CreatePermissionGroup(ctx context.Context, group *entity.PermissionGroup, operatorID int64) error {
	if err := r.Create(ctx, toPermissionGroupModel(group, operatorID)); err != nil {
		if yomidb.IsDuplicateKey(err, r.dbType) {
			return errors.Sentinel(repo_contract.ErrDuplicatedKey)
		}
		return err
	}
	return nil
}

// UpdatePermissionGroup 更新权限组资料
func (r *RbacRepo) UpdatePermissionGroup(ctx context.Context, group *entity.PermissionGroup, operatorID int64) error {
	updates := map[string]any{
		"name":      group.Name,
		"remark":    group.Remark,
		"update_by": ptrutil.PtrFromInt64(operatorID),
	}
	if err := r.UpdateByID(ctx, group.ID, &permissionGroupModel{}, updates); err != nil {
		return errors.Wrap(err, "更新权限组失败")
	}
	return nil
}

// UpdatePermissionGroupStatus 更新权限组启停
func (r *RbacRepo) UpdatePermissionGroupStatus(ctx context.Context, id int64, status enums.Status, operatorID int64) error {
	updates := map[string]any{
		"status":    status,
		"update_by": ptrutil.PtrFromInt64(operatorID),
	}
	if err := r.UpdateByID(ctx, id, &permissionGroupModel{}, updates); err != nil {
		return errors.Wrap(err, "更新权限组状态失败")
	}
	return nil
}

// DeletePermissionGroup 硬删除权限组
func (r *RbacRepo) DeletePermissionGroup(ctx context.Context, id int64) error {
	if err := r.DeleteByID(ctx, id, &permissionGroupModel{}); err != nil {
		return errors.Wrap(err, "删除权限组失败")
	}
	return nil
}

// FindPermissionByID 按主键查询权限
func (r *RbacRepo) FindPermissionByID(ctx context.Context, id int64) (*entity.Permission, bool, error) {
	var row permissionModel
	found, err := r.FirstByID(ctx, id, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按主键查询权限失败")
	}
	if !found {
		return nil, false, nil
	}
	return toPermissionEntity(&row), true, nil
}

// FindPermissionByCode 按编码查询权限
func (r *RbacRepo) FindPermissionByCode(ctx context.Context, code string) (*entity.Permission, bool, error) {
	var row permissionModel
	found, err := r.FirstByField(ctx, "code", code, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按编码查询权限失败")
	}
	if !found {
		return nil, false, nil
	}
	return toPermissionEntity(&row), true, nil
}

// FindPermissionByUIPath 按菜单路径查询权限
func (r *RbacRepo) FindPermissionByUIPath(ctx context.Context, uiPath string) (*entity.Permission, bool, error) {
	var row permissionModel
	found, err := r.FirstByField(ctx, "ui_path", uiPath, &row)
	if err != nil {
		return nil, false, errors.Wrap(err, "按菜单路径查询权限失败")
	}
	if !found {
		return nil, false, nil
	}
	return toPermissionEntity(&row), true, nil
}

// FindPermissionByHTTPRoute 按 HTTP 方法 + 路由模板查询权限
func (r *RbacRepo) FindPermissionByHTTPRoute(ctx context.Context, method, path string) (*entity.Permission, bool, error) {
	var row permissionModel
	err := tx.Exec(ctx, r.DB).
		Where("http_method = ? AND http_path = ?", method, path).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, errors.Wrap(err, "按 HTTP 路由查询权限失败")
	}
	return toPermissionEntity(&row), true, nil
}

// FindPermissionsByIDs 按主键批量查询权限
func (r *RbacRepo) FindPermissionsByIDs(ctx context.Context, ids []int64) ([]*entity.Permission, error) {
	var rows []*permissionModel
	if err := r.FindByIDs(ctx, ids, &permissionModel{}, &rows); err != nil {
		return nil, errors.Wrap(err, "批量查询权限失败")
	}
	items := make([]*entity.Permission, 0, len(rows))
	for _, row := range rows {
		items = append(items, toPermissionEntity(row))
	}
	return items, nil
}

// PagePermissions 分页查询权限目录
func (r *RbacRepo) PagePermissions(ctx context.Context, keyword string, permissionType enums.PermissionType, status enums.Status, page, pageSize int) ([]*entity.Permission, int64, error) {
	query := tx.Exec(ctx, r.DB).Model(&permissionModel{})
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := orm.LikeFull(keyword)
		query = query.Where("code LIKE ? OR name LIKE ?", like, like)
	}
	if permissionType != "" {
		query = query.Where("permission_type = ?", permissionType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var rows []*permissionModel
	total, err := r.CountPage(ctx, query, page, pageSize, &rows, orm.OrderAsc("id"))
	if err != nil {
		return nil, 0, errors.Wrap(err, "分页查询权限失败")
	}

	items := make([]*entity.Permission, 0, len(rows))
	for _, row := range rows {
		items = append(items, toPermissionEntity(row))
	}
	return items, total, nil
}

// UpdatePermissionStatus 更新权限启停
func (r *RbacRepo) UpdatePermissionStatus(ctx context.Context, id int64, status enums.Status, operatorID int64) error {
	updates := map[string]any{
		"status":    status,
		"update_by": ptrutil.PtrFromInt64(operatorID),
	}
	if err := r.UpdateByID(ctx, id, &permissionModel{}, updates); err != nil {
		return errors.Wrap(err, "更新权限状态失败")
	}
	return nil
}

// CreatePermission 创建权限
func (r *RbacRepo) CreatePermission(ctx context.Context, permission *entity.Permission, operatorID int64) error {
	if err := r.Create(ctx, toPermissionModel(permission, operatorID)); err != nil {
		if yomidb.IsDuplicateKey(err, r.dbType) {
			return errors.Sentinel(repo_contract.ErrDuplicatedKey)
		}
		return err
	}
	return nil
}

// UpdatePermission 更新权限资料
func (r *RbacRepo) UpdatePermission(ctx context.Context, permission *entity.Permission, operatorID int64) error {
	updates := map[string]any{
		"name":        permission.Name,
		"ui_path":     ptrutil.PtrFromString(permission.UIPath),
		"http_method": ptrutil.PtrFromString(permission.HTTPMethod),
		"http_path":   ptrutil.PtrFromString(permission.HTTPPath),
		"remark":      permission.Remark,
		"update_by":   ptrutil.PtrFromInt64(operatorID),
	}
	if err := r.UpdateByID(ctx, permission.ID, &permissionModel{}, updates); err != nil {
		if yomidb.IsDuplicateKey(err, r.dbType) {
			return errors.Sentinel(repo_contract.ErrDuplicatedKey)
		}
		return errors.Wrap(err, "更新权限失败")
	}
	return nil
}

// DeletePermission 硬删除权限
func (r *RbacRepo) DeletePermission(ctx context.Context, id int64) error {
	if err := r.DeleteByID(ctx, id, &permissionModel{}); err != nil {
		return errors.Wrap(err, "删除权限失败")
	}
	return nil
}

// ListGroupIDsByRoleID 列出角色已绑权限组 ID
func (r *RbacRepo) ListGroupIDsByRoleID(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	if err := tx.Exec(ctx, r.DB).
		Model(&roleGroupRelModel{}).
		Where("role_id = ?", roleID).
		Order("group_id ASC").
		Pluck("group_id", &ids).Error; err != nil {
		return nil, errors.Wrap(err, "查询角色权限组失败")
	}
	return ids, nil
}

// CountUserRolesByRoleID 统计绑定了角色的用户数
func (r *RbacRepo) CountUserRolesByRoleID(ctx context.Context, roleID int64) (int64, error) {
	var total int64
	if err := tx.Exec(ctx, r.DB).
		Model(&userRoleRelModel{}).
		Where("role_id = ?", roleID).
		Count(&total).Error; err != nil {
		return 0, errors.Wrap(err, "统计角色用户绑定失败")
	}
	return total, nil
}

// DeleteUserRolesByUserID 硬删除用户全部角色关系
func (r *RbacRepo) DeleteUserRolesByUserID(ctx context.Context, userID int64) error {
	if err := tx.Exec(ctx, r.DB).
		Unscoped().
		Where("user_id = ?", userID).
		Delete(&userRoleRelModel{}).Error; err != nil {
		return errors.Wrap(err, "清理用户角色失败")
	}
	return nil
}

// ReplaceRoleGroups 覆盖角色权限组绑定
func (r *RbacRepo) ReplaceRoleGroups(ctx context.Context, roleID int64, rels []*entity.RoleGroupRel, operatorID int64) error {
	if err := tx.Exec(ctx, r.DB).
		Unscoped().
		Where("role_id = ?", roleID).
		Delete(&roleGroupRelModel{}).Error; err != nil {
		return errors.Wrap(err, "清理角色权限组失败")
	}
	if len(rels) == 0 {
		return nil
	}
	if err := r.CreateInBatches(ctx, toRoleGroupRelModels(rels, operatorID), 100); err != nil {
		return errors.Wrap(err, "写入角色权限组失败")
	}
	return nil
}

// ListPermissionIDsByRoleID 列出角色所绑权限组的权限 ID（去重聚合）
func (r *RbacRepo) ListPermissionIDsByRoleID(ctx context.Context, roleID int64) ([]int64, error) {
	var ids []int64
	if err := tx.Exec(ctx, r.DB).
		Raw(sqlListPermissionIDsByRoleID, roleID).
		Scan(&ids).Error; err != nil {
		return nil, errors.Wrap(err, "查询角色聚合权限失败")
	}
	return ids, nil
}

// DeleteRoleGroupsByRoleID 硬删除角色全部权限组关系
func (r *RbacRepo) DeleteRoleGroupsByRoleID(ctx context.Context, roleID int64) error {
	if err := tx.Exec(ctx, r.DB).
		Unscoped().
		Where("role_id = ?", roleID).
		Delete(&roleGroupRelModel{}).Error; err != nil {
		return errors.Wrap(err, "清理角色权限组失败")
	}
	return nil
}

// CountRoleGroupsByGroupID 统计绑定了权限组的角色数
func (r *RbacRepo) CountRoleGroupsByGroupID(ctx context.Context, groupID int64) (int64, error) {
	var total int64
	if err := tx.Exec(ctx, r.DB).
		Model(&roleGroupRelModel{}).
		Where("group_id = ?", groupID).
		Count(&total).Error; err != nil {
		return 0, errors.Wrap(err, "统计权限组角色绑定失败")
	}
	return total, nil
}

// DeleteRoleGroupsByGroupID 硬删除权限组全部角色关系
func (r *RbacRepo) DeleteRoleGroupsByGroupID(ctx context.Context, groupID int64) error {
	if err := tx.Exec(ctx, r.DB).
		Unscoped().
		Where("group_id = ?", groupID).
		Delete(&roleGroupRelModel{}).Error; err != nil {
		return errors.Wrap(err, "清理权限组角色关系失败")
	}
	return nil
}

// ListPermissionIDsByGroupID 列出权限组已绑权限 ID
func (r *RbacRepo) ListPermissionIDsByGroupID(ctx context.Context, groupID int64) ([]int64, error) {
	var ids []int64
	if err := tx.Exec(ctx, r.DB).
		Model(&groupPermissionRelModel{}).
		Where("group_id = ?", groupID).
		Order("permission_id ASC").
		Pluck("permission_id", &ids).Error; err != nil {
		return nil, errors.Wrap(err, "查询权限组权限失败")
	}
	return ids, nil
}

// ReplaceGroupPermissions 覆盖权限组权限绑定
func (r *RbacRepo) ReplaceGroupPermissions(ctx context.Context, groupID int64, rels []*entity.GroupPermissionRel, operatorID int64) error {
	if err := tx.Exec(ctx, r.DB).
		Unscoped().
		Where("group_id = ?", groupID).
		Delete(&groupPermissionRelModel{}).Error; err != nil {
		return errors.Wrap(err, "清理权限组权限失败")
	}
	if len(rels) == 0 {
		return nil
	}
	if err := r.CreateInBatches(ctx, toGroupPermissionRelModels(rels, operatorID), 100); err != nil {
		return errors.Wrap(err, "写入权限组权限失败")
	}
	return nil
}

// DeleteGroupPermissionsByGroupID 硬删除权限组全部权限关系
func (r *RbacRepo) DeleteGroupPermissionsByGroupID(ctx context.Context, groupID int64) error {
	if err := tx.Exec(ctx, r.DB).
		Unscoped().
		Where("group_id = ?", groupID).
		Delete(&groupPermissionRelModel{}).Error; err != nil {
		return errors.Wrap(err, "清理权限组权限失败")
	}
	return nil
}

// DeleteGroupPermissionsByPermissionID 硬删除权限被权限组引用的全部关系
func (r *RbacRepo) DeleteGroupPermissionsByPermissionID(ctx context.Context, permissionID int64) error {
	if err := tx.Exec(ctx, r.DB).
		Unscoped().
		Where("permission_id = ?", permissionID).
		Delete(&groupPermissionRelModel{}).Error; err != nil {
		return errors.Wrap(err, "清理权限的权限组关系失败")
	}
	return nil
}

// ListRoleIDsByUserID 列出用户已绑角色 ID
func (r *RbacRepo) ListRoleIDsByUserID(ctx context.Context, userID int64) ([]int64, error) {
	var ids []int64
	if err := tx.Exec(ctx, r.DB).
		Model(&userRoleRelModel{}).
		Where("user_id = ?", userID).
		Order("role_id ASC").
		Pluck("role_id", &ids).Error; err != nil {
		return nil, errors.Wrap(err, "查询用户角色失败")
	}
	return ids, nil
}

// ListRolesByUserIDs 按用户批量查询已绑角色
func (r *RbacRepo) ListRolesByUserIDs(ctx context.Context, userIDs []int64) (map[int64][]*entity.Role, error) {
	result := make(map[int64][]*entity.Role, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	type relationRow struct {
		UserID int64 `gorm:"column:user_id"`
		RoleID int64 `gorm:"column:role_id"`
	}
	var relations []relationRow
	if err := tx.Exec(ctx, r.DB).
		Model(&userRoleRelModel{}).
		Select("user_id, role_id").
		Where("user_id IN ?", userIDs).
		Order("user_id ASC, role_id ASC").
		Find(&relations).Error; err != nil {
		return nil, errors.Wrap(err, "查询用户角色关系失败")
	}

	roleIDs := make([]int64, 0, len(relations))
	for _, relation := range relations {
		roleIDs = append(roleIDs, relation.RoleID)
	}
	roles, err := r.FindRolesByIDs(ctx, uniqueInt64s(roleIDs))
	if err != nil {
		return nil, err
	}
	roleByID := make(map[int64]*entity.Role, len(roles))
	for _, role := range roles {
		roleByID[role.ID] = role
	}
	for _, relation := range relations {
		if role := roleByID[relation.RoleID]; role != nil {
			result[relation.UserID] = append(result[relation.UserID], role)
		}
	}
	return result, nil
}

// ReplaceUserRoles 覆盖用户角色绑定
func (r *RbacRepo) ReplaceUserRoles(ctx context.Context, userID int64, rels []*entity.UserRoleRel, operatorID int64) error {
	if err := tx.Exec(ctx, r.DB).
		Unscoped().
		Where("user_id = ?", userID).
		Delete(&userRoleRelModel{}).Error; err != nil {
		return errors.Wrap(err, "清理用户角色失败")
	}
	if len(rels) == 0 {
		return nil
	}
	if err := r.CreateInBatches(ctx, toUserRoleRelModels(rels, operatorID), 100); err != nil {
		return errors.Wrap(err, "写入用户角色失败")
	}
	return nil
}

// UserHasEnabledSuperAdmin 判断用户是否拥有启用态超管角色
func (r *RbacRepo) UserHasEnabledSuperAdmin(ctx context.Context, userID int64) (bool, error) {
	var matched int64
	if err := tx.Exec(ctx, r.DB).
		Raw(sqlUserHasEnabledSuperAdmin, userID, enums.RoleCodeSuperAdmin, enums.StatusEnabled).
		Scan(&matched).Error; err != nil {
		return false, errors.Wrap(err, "查询超管角色失败")
	}
	return matched > 0, nil
}

// ListUserIDsWithEnabledSuperAdmin 列出拥有启用态超管角色的用户 ID
func (r *RbacRepo) ListUserIDsWithEnabledSuperAdmin(ctx context.Context) ([]int64, error) {
	var ids []int64
	if err := tx.Exec(ctx, r.DB).
		Raw(sqlListSuperAdminUserIDs, enums.RoleCodeSuperAdmin, enums.StatusEnabled).
		Scan(&ids).Error; err != nil {
		return nil, errors.Wrap(err, "查询超管用户失败")
	}
	return ids, nil
}

// ExistsRouteAccess 判断用户是否可访问指定 HTTP 方法与路由模板
func (r *RbacRepo) ExistsRouteAccess(ctx context.Context, userID int64, method, route string) (bool, error) {
	var matched int64
	if err := tx.Exec(ctx, r.DB).
		Raw(sqlExistsRouteAccess, userID, enums.StatusEnabled, enums.StatusEnabled, enums.StatusEnabled, enums.PermissionTypeAPI, method, route).
		Scan(&matched).Error; err != nil {
		return false, errors.Wrap(err, "查询接口权限失败")
	}
	return matched > 0, nil
}

// ListMenusByUserID 列出用户可见的启用菜单
func (r *RbacRepo) ListMenusByUserID(ctx context.Context, userID int64) ([]*entity.Permission, error) {
	var rows []*permissionModel
	if err := tx.Exec(ctx, r.DB).
		Raw(sqlListMenusByUserID, userID, enums.StatusEnabled, enums.StatusEnabled, enums.StatusEnabled, enums.PermissionTypeMenu).
		Scan(&rows).Error; err != nil {
		return nil, errors.Wrap(err, "查询用户菜单失败")
	}
	items := make([]*entity.Permission, 0, len(rows))
	for _, row := range rows {
		items = append(items, toPermissionEntity(row))
	}
	return items, nil
}

// ListEnabledMenus 列出全部启用菜单
func (r *RbacRepo) ListEnabledMenus(ctx context.Context) ([]*entity.Permission, error) {
	var rows []*permissionModel
	if err := tx.Exec(ctx, r.DB).
		Model(&permissionModel{}).
		Where("permission_type = ? AND status = ?", enums.PermissionTypeMenu, enums.StatusEnabled).
		Order("id ASC").
		Find(&rows).Error; err != nil {
		return nil, errors.Wrap(err, "查询启用菜单失败")
	}
	items := make([]*entity.Permission, 0, len(rows))
	for _, row := range rows {
		items = append(items, toPermissionEntity(row))
	}
	return items, nil
}

// ListPermissionCodesByUserID 列出用户拥有的启用权限编码
func (r *RbacRepo) ListPermissionCodesByUserID(ctx context.Context, userID int64) ([]string, error) {
	var codes []string
	if err := tx.Exec(ctx, r.DB).
		Raw(sqlListPermissionCodesByUserID, userID, enums.StatusEnabled, enums.StatusEnabled, enums.StatusEnabled).
		Scan(&codes).Error; err != nil {
		return nil, errors.Wrap(err, "查询用户权限编码失败")
	}
	return codes, nil
}

// ListEnabledPermissionCodes 列出全部启用权限编码
func (r *RbacRepo) ListEnabledPermissionCodes(ctx context.Context) ([]string, error) {
	var codes []string
	if err := tx.Exec(ctx, r.DB).
		Model(&permissionModel{}).
		Where("status = ?", enums.StatusEnabled).
		Order("code ASC").
		Pluck("code", &codes).Error; err != nil {
		return nil, errors.Wrap(err, "查询启用权限编码失败")
	}
	return codes, nil
}

// uniqueInt64s 按首次出现顺序去重
func uniqueInt64s(ids []int64) []int64 {
	if len(ids) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

/** 后端统一响应体（yomi-boot 默认协议） */
export interface ApiResponse<T = unknown> {
  code: string
  msg: string
  timestamp: number
  data: T
  /** 分页接口在顶层返回总数 */
  total?: number
}

/** 分页查询参数 */
export interface PageQuery {
  page: number
  pageSize: number
}

/** 分页响应 */
export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

/** 启停状态，与后端 Status 枚举一致 */
export type Status = 'ENABLED' | 'DISABLED'

/** 登录参数 */
export interface LoginParams {
  username: string
  password: string
}

/** 登录结果 */
export interface LoginResult {
  token: string
  user: UserInfo
}

/** 菜单视图 */
export interface MenuItem {
  code: string
  name: string
  uiPath: string
}

/** 当前登录用户信息 */
export interface UserInfo {
  username: string
  nickname: string
  /** 权限编码列表，'*' 表示全部权限 */
  permissions: string[]
  /** 可见菜单，用于过滤侧栏与页面访问 */
  menus: MenuItem[]
}

/** 用户状态 */
export type UserStatus = Status

/** 用户列表中的角色摘要 */
export interface RoleBrief {
  id: string
  code: string
  name: string
  status: Status
}

export interface UserItem {
  /** 雪花 ID，后端以字符串返回，避免 JS 精度丢失 */
  id: string
  username: string
  nickname: string
  email: string
  phone: string
  status: UserStatus
  roles: RoleBrief[]
  /** Unix 毫秒时间戳，前端按客户端时区展示 */
  createdAt: number
}

/** 编辑用户时的请求体 */
export type UserPayload = Omit<UserItem, 'id' | 'createdAt' | 'roles'>

/** 新增用户时的请求体（后端要求设置初始密码） */
export type CreateUserPayload = UserPayload & { password: string }

/** 权限类型 */
export type PermissionType = 'MENU' | 'API'

/** 角色列表项 */
export interface RoleItem {
  id: string
  code: string
  name: string
  builtin: boolean
  status: Status
  remark: string
  createTime: number
  updateTime: number
}

/** 角色详情（含已绑权限 ID） */
export interface RoleDetail extends RoleItem {
  permissionIds: string[]
}

/** 角色创建/更新请求体 */
export interface RolePayload {
  code?: string
  name: string
  remark: string
  status: Status
}

/** 权限目录项 */
export interface PermissionItem {
  id: string
  code: string
  name: string
  type: PermissionType
  uiPath?: string
  httpMethod?: string
  httpPath?: string
  status: Status
  remark: string
  createTime: number
  updateTime: number
}

/** 用户角色绑定视图 */
export interface UserRoleBinding {
  roles: RoleItem[]
  roleIds: string[]
}

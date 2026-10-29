/** 后端统一响应体（yomi-boot 默认协议） */
export interface ApiResponse<T = unknown> {
  code: string
  msg: string
  timestamp: number
  data: T
}

/** 分页接口的 data 载荷：当前页数组 + 总条数 */
export interface PageData<T> {
  items: T[]
  total: number
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

/** 登录图像验证码 */
export interface CaptchaChallenge {
  /** 校验时原样回传的标识 */
  id: string
  /** data:image/png;base64,...，可直接作为 img src */
  image: string
  /** 剩余有效期（秒） */
  expiresIn: number
}

/** 登录页公开配置 */
export interface LoginConfig {
  /** 登录是否需要图形验证码（由 biz.captcha.enabled 决定） */
  captchaRequired: boolean
}

/** 登录参数 */
export interface LoginParams {
  username: string
  password: string
  captchaId: string
  captchaCode: string
}

/** 登录结果 */
export interface LoginResult {
  token: string
  user: UserInfo
}

/** 修改自身密码请求体 */
export interface ChangePasswordPayload {
  oldPassword: string
  newPassword: string
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

/** 角色详情（含已绑权限组、聚合权限与生效权限 ID） */
export interface RoleDetail extends RoleItem {
  /** 已绑权限组 ID */
  groupIds: string[]
  /** 权限组内权限的去重聚合（只读），含停用权限组与停用权限 */
  permissionIds: string[]
  /** 实际生效的权限 ID：仅启用权限组下的启用权限，用于标注未生效项 */
  effectivePermissionIds: string[]
}

/** 角色创建/更新请求体；启停走独立接口 */
export interface RolePayload {
  code?: string
  name: string
  remark: string
}

/** 鉴权缓存刷新结果 */
export interface RbacCacheRefreshResult {
  /** 投影版本号 */
  version: string
  /** 投影构建时间（Unix 毫秒） */
  builtAt: number
  roleCount: number
  permissionCount: number
  permissionGroupCount: number
  /** 用户-角色 / 角色-权限组 / 权限组-权限 关系的条目总数 */
  relationCount: number
  /** 刷新耗时（毫秒） */
  durationMs: number
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

/** 权限组列表项 */
export interface PermissionGroupItem {
  id: string
  code: string
  name: string
  status: Status
  remark: string
  createTime: number
  updateTime: number
}

/** 权限组详情（含已绑权限 ID） */
export interface PermissionGroupDetail extends PermissionGroupItem {
  permissionIds: string[]
}

/** 权限组创建/更新请求体；启停走独立接口 */
export interface PermissionGroupPayload {
  code?: string
  name: string
  remark: string
}

/** 权限创建/更新请求体；编码与类型不可变，启停走独立接口 */
export interface PermissionPayload {
  code?: string
  name: string
  type: PermissionType
  uiPath?: string
  httpMethod?: string
  httpPath?: string
  remark: string
}

/** 用户角色绑定视图 */
export interface UserRoleBinding {
  roles: RoleItem[]
  roleIds: string[]
}

/** 访问记录可筛选的 HTTP 方法 */
export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

/** HTTP 状态码分段 */
export type HttpStatusClass = '2xx' | '3xx' | '4xx' | '5xx'

/** 访问记录查询参数 */
export interface HttpRecordQuery extends PageQuery {
  keyword?: string
  method?: HttpMethod | ''
  statusClass?: HttpStatusClass | ''
  /** Unix 毫秒时间戳 */
  startTime?: number
  /** Unix 毫秒时间戳 */
  endTime?: number
}

/** 访问记录列表项 */
export interface HttpRecordItem {
  /** 雪花 ID，后端以字符串返回 */
  id: string
  serviceType: string
  serviceVersion: string
  traceId: string
  /** Unix 毫秒时间戳 */
  requestAt: number
  durationMs: number
  method: string
  path: string
  route: string
  clientIp: string
  /** 当前登录用户 ID，未登录为 '0' */
  userId: string
  clientType: string
  statusCode: number
  bizCode: string
  bizMsg: string
}

/** 访问记录详情（含正文与请求头） */
export interface HttpRecordDetail extends HttpRecordItem {
  query: string
  userAgent: string
  /** 已脱敏的请求头 JSON 字符串 */
  requestHeaders: string
  requestBody: string
  requestBodyTruncated: boolean
  responseBody: string
  responseBodyTruncated: boolean
  createTime: number
}

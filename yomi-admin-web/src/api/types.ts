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

export interface LoginParams {
  username: string
  password: string
}

export interface LoginResult {
  token: string
  user: UserInfo
}

export interface UserInfo {
  username: string
  nickname: string
  /** 权限标识列表，'*' 表示全部权限 */
  permissions: string[]
}

export type UserStatus = 0 | 1

export interface UserItem {
  id: number
  username: string
  nickname: string
  email: string
  phone: string
  status: UserStatus
  /** Unix 毫秒时间戳，前端按客户端时区展示 */
  createdAt: number
}

/** 编辑用户时的请求体 */
export type UserPayload = Omit<UserItem, 'id' | 'createdAt'>

/** 新增用户时的请求体（后端要求设置初始密码） */
export type CreateUserPayload = UserPayload & { password: string }

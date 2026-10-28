/** 后端统一响应体（所有接口的约定，见 README） */
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
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
  createdAt: string
}

/** 新增/编辑用户时的请求体 */
export type UserPayload = Omit<UserItem, 'id' | 'createdAt'>

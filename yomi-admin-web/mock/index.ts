/**
 * 极简 dev mock：以 vite 中间件方式拦截 /api 请求，不引入额外依赖。
 * - 仅在 `apply: 'serve'`（dev）下生效，且由 vite.config.ts 传入的 enabled 控制
 * - 数据保存在内存中，重启 dev server 即重置
 * - 后端就绪后把 .env.development 里 VITE_USE_MOCK 改为 false 即可切换到真实接口
 */
import type { Connect, Plugin } from 'vite'
import type { ServerResponse } from 'node:http'

type Status = 'ENABLED' | 'DISABLED'
type PermissionType = 'MENU' | 'API'

/** 与 src/api/types.ts 的 ApiResponse 保持一致 */
interface MockResponse<T = unknown> {
  code: string
  msg: string
  timestamp: number
  data: T
}

interface MockPermission {
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

interface MockRole {
  id: string
  code: string
  name: string
  builtin: boolean
  status: Status
  remark: string
  groupIds: string[]
  createTime: number
  updateTime: number
}

interface MockPermissionGroup {
  id: string
  code: string
  name: string
  status: Status
  remark: string
  permissionIds: string[]
  createTime: number
  updateTime: number
}

interface MockUser {
  id: string
  username: string
  nickname: string
  email: string
  phone: string
  status: Status
  roles: MockRole[]
  createdAt: number
}

interface MockHttpRecord {
  id: string
  serviceType: string
  serviceVersion: string
  traceId: string
  requestAt: number
  durationMs: number
  method: string
  path: string
  route: string
  query: string
  clientIp: string
  userAgent: string
  requestHeaders: string
  requestBody: string
  requestBodyTruncated: boolean
  userId: string
  clientType: string
  statusCode: number
  responseBody: string
  responseBodyTruncated: boolean
  bizCode: string
  bizMsg: string
  createTime: number
}

const ok = <T>(data: T): MockResponse<T> => ({
  code: 'OK',
  msg: 'success',
  timestamp: Date.now(),
  data,
})
const fail = (msg: string, code = 'ADMIN_ERROR'): MockResponse<null> => ({
  code,
  msg,
  timestamp: Date.now(),
  data: null,
})
const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))
const now = () => Date.now()

// ———————— 权限目录 ————————
const permissions: MockPermission[] = [
  { id: 'p1', code: 'system:user', name: '系统用户', type: 'MENU', uiPath: '/system/user', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p2', code: 'system:role', name: '角色', type: 'MENU', uiPath: '/system/role', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p3', code: 'system:permission', name: '权限', type: 'MENU', uiPath: '/system/permission', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p4', code: 'system:http-record', name: '访问记录', type: 'MENU', uiPath: '/system/http-record', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p5', code: 'system:permission-group', name: '权限组', type: 'MENU', uiPath: '/system/permission-group', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p7', code: 'system:audit-log', name: '审计日志', type: 'MENU', uiPath: '/system/audit-log', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p91', code: 'system:audit-log:list', name: '审计日志列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/audit-log', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p11', code: 'system:user:list', name: '用户列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/users', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p12', code: 'system:user:create', name: '创建用户', type: 'API', httpMethod: 'POST', httpPath: '/api/system/users', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p13', code: 'system:user:update', name: '更新用户', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/users/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p14', code: 'system:user:delete', name: '删除用户', type: 'API', httpMethod: 'DELETE', httpPath: '/api/system/users/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p16', code: 'system:user:session:list', name: '用户会话列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/user/session', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p17', code: 'system:user:session:revoke', name: '强制下线会话', type: 'API', httpMethod: 'POST', httpPath: '/api/system/user/session/revoke', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p18', code: 'system:user:session:revoke-all', name: '强制下线用户全部会话', type: 'API', httpMethod: 'POST', httpPath: '/api/system/user/session/revoke-all', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p21', code: 'system:role:list', name: '角色列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/role', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p22', code: 'system:role:detail', name: '角色详情', type: 'API', httpMethod: 'GET', httpPath: '/api/system/role/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p23', code: 'system:role:create', name: '创建角色', type: 'API', httpMethod: 'POST', httpPath: '/api/system/role', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p24', code: 'system:role:update', name: '更新角色', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/role/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p25', code: 'system:role:bind-group', name: '角色绑定权限组', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/role/:id/group', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p26', code: 'system:rbac_cache:refresh', name: '刷新权限缓存', type: 'API', httpMethod: 'POST', httpPath: '/api/system/rbac-cache/refresh', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p31', code: 'system:permission:list', name: '权限列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/permission', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p32', code: 'system:permission:status', name: '权限启停', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/permission/:id/status', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p41', code: 'system:user:role:list', name: '用户角色查询', type: 'API', httpMethod: 'GET', httpPath: '/api/system/user/:id/role', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p42', code: 'system:user:role:bind', name: '用户角色绑定', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/user/:id/role', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p51', code: 'system:http-record:list', name: '访问记录列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/http-records', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p52', code: 'system:http-record:detail', name: '访问记录详情', type: 'API', httpMethod: 'GET', httpPath: '/api/system/http-records/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p61', code: 'system:permission-group:list', name: '权限组列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/permission-group', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p62', code: 'system:permission-group:detail', name: '权限组详情', type: 'API', httpMethod: 'GET', httpPath: '/api/system/permission-group/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p63', code: 'system:permission-group:create', name: '创建权限组', type: 'API', httpMethod: 'POST', httpPath: '/api/system/permission-group', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p64', code: 'system:permission-group:update', name: '更新权限组', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/permission-group/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p65', code: 'system:permission-group:bind-permission', name: '权限组绑定权限', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/permission-group/:id/permission', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
]

// ———————— 权限组 ————————
const permissionGroups: MockPermissionGroup[] = [
  { id: 'g1', code: 'system_user', name: '用户管理', status: 'ENABLED', remark: '用户菜单、用户 CRUD、用户角色绑定与会话管理', permissionIds: ['p1', 'p11', 'p12', 'p13', 'p14', 'p41', 'p42', 'p16', 'p17', 'p18'], createTime: now(), updateTime: now() },
  { id: 'g2', code: 'system_role', name: '角色管理', status: 'ENABLED', remark: '角色菜单、角色 CRUD 与角色绑定权限组', permissionIds: ['p2', 'p21', 'p22', 'p23', 'p24', 'p25', 'p26'], createTime: now(), updateTime: now() },
  { id: 'g3', code: 'system_permission', name: '权限管理', status: 'ENABLED', remark: '权限目录与权限组的查看、维护', permissionIds: ['p3', 'p5', 'p31', 'p32', 'p61', 'p62', 'p63', 'p64', 'p65'], createTime: now(), updateTime: now() },
  { id: 'g4', code: 'system_http_record', name: '访问记录', status: 'ENABLED', remark: '访问记录菜单、列表与详情', permissionIds: ['p4', 'p51', 'p52'], createTime: now(), updateTime: now() },
]
let nextGroupId = permissionGroups.length + 1

// ———————— 角色 ————————
const roles: MockRole[] = [
  {
    id: 'r1',
    code: 'super_admin',
    name: '超级管理员',
    builtin: true,
    status: 'ENABLED',
    remark: '内置角色，拥有全部权限',
    groupIds: [],
    createTime: now(),
    updateTime: now(),
  },
  {
    id: 'r2',
    code: 'viewer',
    name: '只读角色',
    builtin: false,
    status: 'ENABLED',
    remark: '仅可查看用户列表',
    groupIds: ['g1'],
    createTime: now(),
    updateTime: now(),
  },
]
const superRole = roles[0]!
const viewerRole = roles[1]!
let nextRoleId = roles.length + 1

// ———————— 用户 ————————
const users: MockUser[] = Array.from({ length: 28 }, (_, index) => {
  const id = index + 1
  return {
    id: String(100000000000000000n + BigInt(id)),
    username: id === 1 ? 'admin' : `user${String(id).padStart(3, '0')}`,
    nickname: id === 1 ? '超级管理员' : `用户${id}`,
    email: `user${id}@yomi.dev`,
    phone: `138${String(10000000 + id * 7).slice(0, 8)}`,
    status: id % 7 === 0 ? 'DISABLED' : 'ENABLED',
    roles: id === 1 ? [superRole] : id === 2 ? [viewerRole] : [],
    createdAt: now() - (28 - id) * 86_400_000,
  }
})
let nextUserId = users.length + 1

// ———————— 在线会话（mock） ————————
interface MockUserSession {
  sessionId: string
  userId: string
  username: string
  nickname: string
  clientType: string
  loginIp: string
  userAgent: string
  loginAt: number
  expireAt: number
  current: boolean
}

const userSessions: MockUserSession[] = [
  {
    sessionId: 'mock-session-current',
    userId: users[0]!.id,
    username: users[0]!.username,
    nickname: users[0]!.nickname,
    clientType: 'admin',
    loginIp: '127.0.0.1',
    userAgent:
      'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0 Safari/537.36',
    loginAt: now() - 15 * 60_000,
    expireAt: now() + 105 * 60_000,
    current: true,
  },
  {
    sessionId: 'mock-session-phone',
    userId: users[0]!.id,
    username: users[0]!.username,
    nickname: users[0]!.nickname,
    clientType: 'admin',
    loginIp: '192.168.1.24',
    userAgent:
      'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148',
    loginAt: now() - 40 * 60_000,
    expireAt: now() + 80 * 60_000,
    current: false,
  },
  {
    sessionId: 'mock-session-user002',
    userId: users[1]!.id,
    username: users[1]!.username,
    nickname: users[1]!.nickname,
    clientType: 'admin',
    loginIp: '10.0.0.8',
    userAgent:
      'Mozilla/5.0 (Macintosh; Intel Mac OS X 14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0 Safari/537.36',
    loginAt: now() - 5 * 60_000,
    expireAt: now() + 115 * 60_000,
    current: false,
  },
]

// ———————— 访问记录 ————————
const httpRecordSeeds = [
  { method: 'GET', path: '/api/system/users', statusCode: 200, bizCode: 'OK' },
  { method: 'GET', path: '/api/auth/profile', statusCode: 200, bizCode: 'OK' },
  { method: 'POST', path: '/api/system/users', statusCode: 200, bizCode: 'ADMIN_USERNAME_EXISTS' },
  { method: 'PUT', path: '/api/system/role/:id', statusCode: 200, bizCode: 'OK' },
  { method: 'DELETE', path: '/api/system/users/:id', statusCode: 200, bizCode: 'OK' },
  { method: 'GET', path: '/api/system/permission', statusCode: 404, bizCode: 'NOT_FOUND' },
  { method: 'GET', path: '/api/system/http-records', statusCode: 200, bizCode: 'OK' },
  { method: 'POST', path: '/api/auth/login', statusCode: 200, bizCode: 'OK' },
  { method: 'GET', path: '/api/system/http-records/:id', statusCode: 500, bizCode: 'INTERNAL_ERROR' },
  { method: 'GET', path: '/api/v1/health', statusCode: 200, bizCode: '' },
]

const httpRecords: MockHttpRecord[] = httpRecordSeeds.map((seed, index) => {
  const requestAt = now() - index * 60_000
  return {
    id: String(100000000000010000n + BigInt(index + 1)),
    serviceType: 'yomi-admin-backend',
    serviceVersion: 'v0.1.0',
    traceId: `mock-trace-${index + 1}`,
    requestAt,
    durationMs: 8 + index * 3,
    method: seed.method,
    path: seed.path,
    route: seed.path,
    query: seed.method === 'GET' ? 'page=1&pageSize=10' : '',
    clientIp: '127.0.0.1',
    userAgent: 'Mozilla/5.0 (mock)',
    requestHeaders: JSON.stringify({ Authorization: '***', 'Content-Type': 'application/json' }),
    requestBody:
      seed.method === 'GET' ? '' : JSON.stringify({ username: 'admin', password: '***' }, null, 2),
    requestBodyTruncated: false,
    userId: index === 7 ? '0' : '100000000000000001',
    clientType: 'admin',
    statusCode: seed.statusCode,
    responseBody:
      seed.statusCode === 200
        ? JSON.stringify({ code: seed.bizCode, msg: 'success', data: {} }, null, 2)
        : '',
    responseBodyTruncated: false,
    bizCode: seed.bizCode,
    bizMsg: seed.statusCode === 200 ? 'success' : 'mock error',
    createTime: requestAt,
  }
})

const tokenToUsername = new Map<string, string>()
const captchaStore = new Map<string, { code: string; expiresAt: number }>()

/** 与后端一致：同一 IP 生成验证码的最小间隔 */
const CAPTCHA_REFRESH_INTERVAL_MS = 10_000
let lastCaptchaAt = 0

/** 生成 mock 验证码：SVG 里直接画出答案，方便本地手动登录 */
function createCaptcha() {
  for (const [key, value] of captchaStore) {
    if (value.expiresAt < now()) captchaStore.delete(key)
  }
  const code = String(Math.floor(1000 + Math.random() * 9000))
  const id = `c${Date.now().toString(36)}${Math.random().toString(36).slice(2, 8)}`
  const expiresIn = 180
  captchaStore.set(id, { code, expiresAt: now() + expiresIn * 1000 })
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="160" height="60" viewBox="0 0 160 60"><rect width="160" height="60" fill="#f2f3f5"/><text x="80" y="42" text-anchor="middle" font-family="monospace" font-size="32" font-weight="700" letter-spacing="8" fill="#303133">${code}</text></svg>`
  return { id, image: `data:image/svg+xml;charset=utf-8,${encodeURIComponent(svg)}`, expiresIn }
}

// ———————— 工具 ————————
function readBody(req: Connect.IncomingMessage): Promise<Record<string, unknown>> {
  return new Promise((resolve) => {
    let raw = ''
    req.on('data', (chunk) => (raw += chunk))
    req.on('end', () => {
      if (!raw) return resolve({})
      try {
        resolve(JSON.parse(raw) as Record<string, unknown>)
      } catch {
        resolve({})
      }
    })
  })
}

function sendJson(res: ServerResponse, body: MockResponse) {
  res.statusCode = 200
  res.setHeader('Content-Type', 'application/json; charset=utf-8')
  res.end(JSON.stringify(body))
}

function bearerToken(req: Connect.IncomingMessage): string | null {
  const header = req.headers.authorization
  if (!header?.startsWith('Bearer ')) return null
  return header.slice('Bearer '.length)
}

function currentUser(req: Connect.IncomingMessage): MockUser | null {
  const token = bearerToken(req)
  const username = token ? tokenToUsername.get(token) : null
  return username ? (users.find((item) => item.username === username) ?? null) : null
}

function isSuper(user: MockUser): boolean {
  return user.roles.some((role) => role.code === 'super_admin' && role.status === 'ENABLED')
}

/** 角色所绑权限组的权限 ID（结构聚合，不看启停） */
function aggregatedPermissionIdsOfRole(role: MockRole): string[] {
  const ids = new Set<string>()
  for (const groupId of role.groupIds) {
    const group = permissionGroups.find((item) => item.id === groupId)
    if (!group) continue
    for (const permissionId of group.permissionIds) ids.add(permissionId)
  }
  return [...ids]
}

/** 用户经启用角色与启用权限组聚合出的启用权限 */
function effectivePermissionsOfUser(user: MockUser): MockPermission[] {
  const ids = new Set<string>()
  for (const role of user.roles) {
    if (role.status !== 'ENABLED') continue
    for (const groupId of role.groupIds) {
      const group = permissionGroups.find((item) => item.id === groupId)
      if (!group || group.status !== 'ENABLED') continue
      for (const permissionId of group.permissionIds) ids.add(permissionId)
    }
  }
  return permissions.filter((item) => item.status === 'ENABLED' && ids.has(item.id))
}

function visibleMenus(user: MockUser): MockPermission[] {
  if (isSuper(user)) {
    return permissions.filter((item) => item.type === 'MENU' && item.status === 'ENABLED')
  }
  return effectivePermissionsOfUser(user).filter((item) => item.type === 'MENU')
}

function grantedCodes(user: MockUser): string[] {
  if (isSuper(user)) return ['*']
  return effectivePermissionsOfUser(user).map((item) => item.code)
}

function identity(user: MockUser) {
  return {
    id: user.id,
    username: user.username,
    nickname: user.nickname,
    email: user.email,
    phone: user.phone,
    roles: user.roles.map((role) => ({
      code: role.code,
      name: role.name,
      status: role.status,
    })),
    permissions: grantedCodes(user),
    menus: visibleMenus(user).map((item) => ({
      code: item.code,
      name: item.name,
      uiPath: item.uiPath ?? '',
    })),
  }
}

function userView(user: MockUser) {
  return {
    id: user.id,
    username: user.username,
    nickname: user.nickname,
    email: user.email,
    phone: user.phone,
    status: user.status,
    roles: user.roles.map((role) => ({
      id: role.id,
      code: role.code,
      name: role.name,
      status: role.status,
    })),
    createdAt: user.createdAt,
  }
}

/** 按页码切分，按 yomi-boot 默认协议返回 data: { items, total } */
function paginate<T>(items: T[], query: URLSearchParams): MockResponse<{ items: T[]; total: number }> {
  const page = Math.max(1, Number(query.get('page')) || 1)
  const pageSize = Math.min(200, Math.max(1, Number(query.get('pageSize')) || 10))
  const start = (page - 1) * pageSize
  return ok({ items: items.slice(start, start + pageSize), total: items.length })
}

type MockHandler = (ctx: {
  req: Connect.IncomingMessage
  res: ServerResponse
  params: string[]
  query: URLSearchParams
}) => Promise<void>

interface MockRoute {
  method: string
  pattern: RegExp
  handler: MockHandler
}

interface MockAuditLog {
  id: string
  auditId: string
  phase: 'STARTED' | 'FINISHED'
  eventCode: string
  operatorId: string
  operatorName: string
  clientIp: string
  targetType: string
  targetId: string
  targetName: string
  result: 'PENDING' | 'SUCCESS' | 'FAILURE'
  bizCode: string
  createTime: number
}

const auditLogs: MockAuditLog[] = [
  { id: 'a12', auditId: '9006', phase: 'STARTED', eventCode: 'auth.reset_password', operatorId: '100000000000000001', operatorName: 'admin', clientIp: '127.0.0.1', targetType: 'user', targetId: 'u2', targetName: '', result: 'PENDING', bizCode: '', createTime: now() - 8 * 60 * 1000 - 1000 },
  { id: 'a6', auditId: '9006', phase: 'FINISHED', eventCode: 'auth.reset_password', operatorId: '100000000000000001', operatorName: 'admin', clientIp: '127.0.0.1', targetType: 'user', targetId: 'u2', targetName: 'test', result: 'SUCCESS', bizCode: 'OK', createTime: now() - 8 * 60 * 1000 },
  { id: 'a11', auditId: '9005', phase: 'STARTED', eventCode: 'auth.change_password', operatorId: 'u2', operatorName: 'test', clientIp: '127.0.0.1', targetType: '', targetId: '0', targetName: '', result: 'PENDING', bizCode: '', createTime: now() - 20 * 60 * 1000 - 1000 },
  { id: 'a5', auditId: '9005', phase: 'FINISHED', eventCode: 'auth.change_password', operatorId: 'u2', operatorName: 'test', clientIp: '127.0.0.1', targetType: '', targetId: '0', targetName: '', result: 'SUCCESS', bizCode: 'OK', createTime: now() - 20 * 60 * 1000 },
  { id: 'a10', auditId: '9004', phase: 'STARTED', eventCode: 'auth.change_password', operatorId: 'u2', operatorName: 'test', clientIp: '127.0.0.1', targetType: '', targetId: '0', targetName: '', result: 'PENDING', bizCode: '', createTime: now() - 21 * 60 * 1000 - 1000 },
  { id: 'a4', auditId: '9004', phase: 'FINISHED', eventCode: 'auth.change_password', operatorId: 'u2', operatorName: 'test', clientIp: '127.0.0.1', targetType: '', targetId: '0', targetName: '', result: 'FAILURE', bizCode: 'ADMIN_OLD_PASSWORD_MISMATCH', createTime: now() - 21 * 60 * 1000 },
  { id: 'a9', auditId: '9003', phase: 'STARTED', eventCode: 'auth.login', operatorId: '0', operatorName: 'test', clientIp: '127.0.0.1', targetType: '', targetId: '0', targetName: '', result: 'PENDING', bizCode: '', createTime: now() - 22 * 60 * 1000 - 1000 },
  { id: 'a3', auditId: '9003', phase: 'FINISHED', eventCode: 'auth.login', operatorId: 'u2', operatorName: 'test', clientIp: '127.0.0.1', targetType: '', targetId: '0', targetName: '', result: 'SUCCESS', bizCode: 'OK', createTime: now() - 22 * 60 * 1000 },
  { id: 'a8', auditId: '9002', phase: 'STARTED', eventCode: 'auth.login', operatorId: '0', operatorName: 'ghost', clientIp: '10.0.0.8', targetType: '', targetId: '0', targetName: '', result: 'PENDING', bizCode: '', createTime: now() - 30 * 60 * 1000 - 1000 },
  { id: 'a2', auditId: '9002', phase: 'FINISHED', eventCode: 'auth.login', operatorId: '0', operatorName: 'ghost', clientIp: '10.0.0.8', targetType: '', targetId: '0', targetName: '', result: 'FAILURE', bizCode: 'ADMIN_INVALID_CREDENTIAL', createTime: now() - 30 * 60 * 1000 },
  { id: 'a7', auditId: '9001', phase: 'STARTED', eventCode: 'auth.login', operatorId: '0', operatorName: 'admin', clientIp: '127.0.0.1', targetType: '', targetId: '0', targetName: '', result: 'PENDING', bizCode: '', createTime: now() - 60 * 60 * 1000 - 1000 },
  { id: 'a1', auditId: '9001', phase: 'FINISHED', eventCode: 'auth.login', operatorId: '100000000000000001', operatorName: 'admin', clientIp: '127.0.0.1', targetType: '', targetId: '0', targetName: '', result: 'SUCCESS', bizCode: 'OK', createTime: now() - 60 * 60 * 1000 },
]

const routes: MockRoute[] = [
  {
    method: 'GET',
    pattern: /^\/api\/auth\/config$/,
    handler: async ({ res }) => {
      await delay(50)
      // mock 演示开启验证码，与后端默认部署一致
      sendJson(res, ok({ captchaRequired: true }))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/auth\/captcha$/,
    handler: async ({ res }) => {
      await delay(80)
      if (now() - lastCaptchaAt < CAPTCHA_REFRESH_INTERVAL_MS) {
        return sendJson(res, fail('验证码刷新过于频繁，请 10 秒后重试', 'ADMIN_CAPTCHA_TOO_FREQUENT'))
      }
      lastCaptchaAt = now()
      sendJson(res, ok(createCaptcha()))
    },
  },
  {
    method: 'POST',
    pattern: /^\/api\/auth\/login$/,
    handler: async ({ req, res }) => {
      const body = await readBody(req)
      const username = String(body.username ?? '')
      const password = String(body.password ?? '')
      const captchaId = String(body.captchaId ?? '')
      const captchaCode = String(body.captchaCode ?? '').trim()
      await delay(200)

      // 与后端一致：先校验并消费验证码
      const issued = captchaStore.get(captchaId)
      captchaStore.delete(captchaId)
      if (!issued || issued.expiresAt < now()) {
        return sendJson(res, fail('验证码已失效，请刷新后重试', 'ADMIN_CAPTCHA_INVALID'))
      }
      if (issued.code !== captchaCode) {
        return sendJson(res, fail('验证码错误', 'ADMIN_CAPTCHA_MISMATCH'))
      }

      const user = users.find((item) => item.username === username)
      const valid = user && (password === 'admin123' || password === 'test1234')
      if (!user || !valid) {
        return sendJson(res, fail('用户名或密码错误', 'ADMIN_INVALID_CREDENTIAL'))
      }
      if (user.status === 'DISABLED') {
        return sendJson(res, fail('账号已被禁用', 'ADMIN_USER_DISABLED'))
      }
      const token = `mock-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`
      tokenToUsername.set(token, username)
      sendJson(res, ok({ token, user: identity(user) }))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/auth\/profile$/,
    handler: async ({ req, res }) => {
      await delay(80)
      const user = currentUser(req)
      if (!user) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      sendJson(res, ok(identity(user)))
    },
  },
  {
    method: 'POST',
    pattern: /^\/api\/auth\/profile$/,
    handler: async ({ req, res }) => {
      await delay(120)
      const user = currentUser(req)
      if (!user) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const nickname = String(body.nickname ?? '').trim()
      const email = String(body.email ?? '').trim()
      const phone = String(body.phone ?? '').trim()
      if (!nickname || nickname.length > 32) {
        return sendJson(res, fail('用户参数不合法', 'ADMIN_INVALID_INPUT'))
      }
      if (email && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
        return sendJson(res, fail('用户参数不合法', 'ADMIN_INVALID_INPUT'))
      }
      if (phone && !/^1\d{10}$/.test(phone)) {
        return sendJson(res, fail('用户参数不合法', 'ADMIN_INVALID_INPUT'))
      }
      user.nickname = nickname
      user.email = email
      user.phone = phone
      sendJson(res, ok(identity(user)))
    },
  },
  {
    method: 'POST',
    pattern: /^\/api\/auth\/logout$/,
    handler: async ({ req, res }) => {
      const token = bearerToken(req)
      if (token) tokenToUsername.delete(token)
      sendJson(res, ok(null))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/user$/,
    handler: async ({ req, res, query }) => {
      await delay(80)
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const keyword = (query.get('keyword') ?? '').trim().toLowerCase()
      const filtered = users.filter(
        (user) =>
          !keyword ||
          user.username.toLowerCase().includes(keyword) ||
          user.nickname.toLowerCase().includes(keyword),
      )
      sendJson(res, paginate(filtered.map(userView), query))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/user\/session$/,
    handler: async ({ req, res, query }) => {
      await delay(80)
      const current = currentUser(req)
      if (!current) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const userId = (query.get('userId') ?? '').trim()
      const items = userSessions
        .filter((session) => session.userId === userId)
        .map((session) => ({ ...session, current: session.current && current.id === userId }))
      sendJson(res, ok(items))
    },
  },
  {
    method: 'POST',
    pattern: /^\/api\/system\/user\/session\/revoke$/,
    handler: async ({ req, res }) => {
      await delay(120)
      const current = currentUser(req)
      if (!current) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const userId = String(body.userId ?? '')
      const sessionId = String(body.sessionId ?? '')
      const index = userSessions.findIndex(
        (session) => session.userId === userId && session.sessionId === sessionId,
      )
      if (index === -1) return sendJson(res, ok(null))
      if (current.id === userId && userSessions[index]!.current) {
        return sendJson(res, fail('不能操作当前登录账号', 'ADMIN_CANNOT_OPERATE_SELF'))
      }
      userSessions.splice(index, 1)
      sendJson(res, ok(null))
    },
  },
  {
    method: 'POST',
    pattern: /^\/api\/system\/user\/session\/revoke-all$/,
    handler: async ({ req, res }) => {
      await delay(120)
      const current = currentUser(req)
      if (!current) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const userId = String(body.userId ?? '')
      if (userId === current.id) {
        return sendJson(res, fail('不能操作当前登录账号', 'ADMIN_CANNOT_OPERATE_SELF'))
      }
      for (let i = userSessions.length - 1; i >= 0; i -= 1) {
        if (userSessions[i]!.userId === userId) userSessions.splice(i, 1)
      }
      sendJson(res, ok(null))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/users$/,
    handler: async ({ req, res, query }) => {
      await delay(120)
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const keyword = (query.get('keyword') ?? '').trim().toLowerCase()
      const filtered = users.filter(
        (user) =>
          !keyword ||
          user.username.toLowerCase().includes(keyword) ||
          user.nickname.toLowerCase().includes(keyword),
      )
      sendJson(res, paginate(filtered.map(userView), query))
    },
  },
  {
    method: 'POST',
    pattern: /^\/api\/system\/users$/,
    handler: async ({ req, res }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const username = String(body.username ?? '').trim()
      if (!username) return sendJson(res, fail('用户名不能为空', 'ADMIN_INVALID_INPUT'))
      if (users.some((user) => user.username === username)) {
        return sendJson(res, fail('用户名已存在', 'ADMIN_USERNAME_EXISTS'))
      }
      users.unshift({
        id: String(100000000000000000n + BigInt(nextUserId++)),
        username,
        nickname: String(body.nickname ?? username),
        email: String(body.email ?? ''),
        phone: String(body.phone ?? ''),
        status: body.status === 'DISABLED' ? 'DISABLED' : 'ENABLED',
        roles: [],
        createdAt: now(),
      })
      sendJson(res, ok(null))
    },
  },
  {
    method: 'PUT',
    pattern: /^\/api\/system\/users\/([0-9]+)$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const target = users.find((user) => user.id === params[0])
      if (!target) return sendJson(res, fail('用户不存在', 'ADMIN_USER_NOT_FOUND'))
      if (body.nickname !== undefined) target.nickname = String(body.nickname)
      if (body.email !== undefined) target.email = String(body.email)
      if (body.phone !== undefined) target.phone = String(body.phone)
      if (body.status !== undefined) {
        target.status = body.status === 'DISABLED' ? 'DISABLED' : 'ENABLED'
      }
      sendJson(res, ok(null))
    },
  },
  {
    method: 'DELETE',
    pattern: /^\/api\/system\/users\/([0-9]+)$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const index = users.findIndex((user) => user.id === params[0])
      if (index === -1) return sendJson(res, fail('用户不存在', 'ADMIN_USER_NOT_FOUND'))
      users.splice(index, 1)
      sendJson(res, ok(null))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/role$/,
    handler: async ({ req, res, query }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const keyword = (query.get('keyword') ?? '').trim().toLowerCase()
      const status = query.get('status') ?? ''
      const filtered = roles.filter((role) => {
        const matchKeyword =
          !keyword ||
          role.code.toLowerCase().includes(keyword) ||
          role.name.toLowerCase().includes(keyword)
        return matchKeyword && (!status || role.status === status)
      })
      sendJson(res, paginate(filtered, query))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/role\/([^/]+)$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const role = roles.find((item) => item.id === params[0])
      if (!role) return sendJson(res, fail('角色不存在', 'ADMIN_ROLE_NOT_FOUND'))
      sendJson(res, ok({ ...role, permissionIds: aggregatedPermissionIdsOfRole(role) }))
    },
  },
  {
    method: 'POST',
    pattern: /^\/api\/system\/role$/,
    handler: async ({ req, res }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const code = String(body.code ?? '').trim()
      if (roles.some((role) => role.code === code)) {
        return sendJson(res, fail('角色编码已存在', 'ADMIN_ROLE_CODE_EXISTS'))
      }
      const role: MockRole = {
        id: `r${nextRoleId++}`,
        code,
        name: String(body.name ?? code),
        builtin: false,
        status: 'ENABLED',
        remark: String(body.remark ?? ''),
        groupIds: [],
        createTime: now(),
        updateTime: now(),
      }
      roles.push(role)
      sendJson(res, ok(role))
    },
  },
  {
    method: 'PUT',
    pattern: /^\/api\/system\/role\/([^/]+)$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const role = roles.find((item) => item.id === params[0])
      if (!role) return sendJson(res, fail('角色不存在', 'ADMIN_ROLE_NOT_FOUND'))
      if (role.builtin && body.status === 'DISABLED') {
        return sendJson(res, fail('内置角色不允许变更', 'ADMIN_BUILTIN_ROLE_IMMUTABLE'))
      }
      if (body.name !== undefined) role.name = String(body.name)
      if (body.remark !== undefined) role.remark = String(body.remark)
      if (body.status !== undefined) {
        role.status = body.status === 'DISABLED' ? 'DISABLED' : 'ENABLED'
      }
      role.updateTime = now()
      sendJson(res, ok(null))
    },
  },
  {
    method: 'PUT',
    pattern: /^\/api\/system\/role\/([^/]+)\/group$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const role = roles.find((item) => item.id === params[0])
      if (!role) return sendJson(res, fail('角色不存在', 'ADMIN_ROLE_NOT_FOUND'))
      if (role.builtin) return sendJson(res, fail('内置角色不允许变更', 'ADMIN_BUILTIN_ROLE_IMMUTABLE'))
      const groupIds = Array.isArray(body.groupIds) ? (body.groupIds as string[]) : []
      const unknown = groupIds.find((id) => !permissionGroups.some((group) => group.id === id))
      if (unknown) return sendJson(res, fail('权限组不存在', 'ADMIN_PERMISSION_GROUP_NOT_FOUND'))
      const disabled = groupIds.find(
        (id) => permissionGroups.find((group) => group.id === id)?.status !== 'ENABLED',
      )
      if (disabled) return sendJson(res, fail('关联的角色、权限组或权限已停用', 'ADMIN_RELATED_DISABLED'))
      role.groupIds = [...new Set(groupIds)]
      role.updateTime = now()
      sendJson(res, ok(null))
    },
  },
  {
    method: 'POST',
    pattern: /^\/api\/system\/rbac-cache\/refresh$/,
    handler: async ({ req, res }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      await delay(120)
      const relationCount =
        users.reduce((total, user) => total + user.roles.length, 0) +
        roles.reduce((total, role) => total + role.groupIds.length, 0) +
        permissionGroups.reduce((total, group) => total + group.permissionIds.length, 0)
      sendJson(
        res,
        ok({
          version: String(now()),
          builtAt: now(),
          roleCount: roles.length,
          permissionCount: permissions.length,
          permissionGroupCount: permissionGroups.length,
          relationCount,
          durationMs: 120,
        }),
      )
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/permission-group$/,
    handler: async ({ req, res, query }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const keyword = (query.get('keyword') ?? '').trim().toLowerCase()
      const status = query.get('status') ?? ''
      const filtered = permissionGroups.filter((group) => {
        const matchKeyword =
          !keyword ||
          group.code.toLowerCase().includes(keyword) ||
          group.name.toLowerCase().includes(keyword)
        return matchKeyword && (!status || group.status === status)
      })
      sendJson(res, paginate(filtered, query))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/permission-group\/([^/]+)$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const group = permissionGroups.find((item) => item.id === params[0])
      if (!group) return sendJson(res, fail('权限组不存在', 'ADMIN_PERMISSION_GROUP_NOT_FOUND'))
      sendJson(res, ok(group))
    },
  },
  {
    method: 'POST',
    pattern: /^\/api\/system\/permission-group$/,
    handler: async ({ req, res }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const code = String(body.code ?? '').trim()
      if (permissionGroups.some((group) => group.code === code)) {
        return sendJson(res, fail('权限组编码已存在', 'ADMIN_PERMISSION_GROUP_CODE_EXISTS'))
      }
      const group: MockPermissionGroup = {
        id: `g${nextGroupId++}`,
        code,
        name: String(body.name ?? code),
        status: 'ENABLED',
        remark: String(body.remark ?? ''),
        permissionIds: [],
        createTime: now(),
        updateTime: now(),
      }
      permissionGroups.push(group)
      sendJson(res, ok(group))
    },
  },
  {
    method: 'PUT',
    pattern: /^\/api\/system\/permission-group\/([^/]+)$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const group = permissionGroups.find((item) => item.id === params[0])
      if (!group) return sendJson(res, fail('权限组不存在', 'ADMIN_PERMISSION_GROUP_NOT_FOUND'))
      if (body.name !== undefined) group.name = String(body.name)
      if (body.remark !== undefined) group.remark = String(body.remark)
      if (body.status !== undefined) {
        group.status = body.status === 'DISABLED' ? 'DISABLED' : 'ENABLED'
      }
      group.updateTime = now()
      sendJson(res, ok(null))
    },
  },
  {
    method: 'PUT',
    pattern: /^\/api\/system\/permission-group\/([^/]+)\/permission$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const group = permissionGroups.find((item) => item.id === params[0])
      if (!group) return sendJson(res, fail('权限组不存在', 'ADMIN_PERMISSION_GROUP_NOT_FOUND'))
      const permissionIds = Array.isArray(body.permissionIds)
        ? (body.permissionIds as string[])
        : []
      const unknown = permissionIds.find((id) => !permissions.some((item) => item.id === id))
      if (unknown) return sendJson(res, fail('权限不存在', 'ADMIN_PERMISSION_NOT_FOUND'))
      const disabled = permissionIds.find(
        (id) => permissions.find((item) => item.id === id)?.status !== 'ENABLED',
      )
      if (disabled) return sendJson(res, fail('关联的角色、权限组或权限已停用', 'ADMIN_RELATED_DISABLED'))
      group.permissionIds = [...new Set(permissionIds)]
      group.updateTime = now()
      sendJson(res, ok(null))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/permission$/,
    handler: async ({ req, res, query }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const keyword = (query.get('keyword') ?? '').trim().toLowerCase()
      const type = query.get('type') ?? ''
      const status = query.get('status') ?? ''
      const filtered = permissions.filter((item) => {
        const matchKeyword =
          !keyword ||
          item.code.toLowerCase().includes(keyword) ||
          item.name.toLowerCase().includes(keyword)
        return matchKeyword && (!type || item.type === type) && (!status || item.status === status)
      })
      sendJson(res, paginate(filtered, query))
    },
  },
  {
    method: 'PUT',
    pattern: /^\/api\/system\/permission\/([^/]+)\/status$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const permission = permissions.find((item) => item.id === params[0])
      if (!permission) return sendJson(res, fail('权限不存在', 'ADMIN_PERMISSION_NOT_FOUND'))
      permission.status = body.status === 'DISABLED' ? 'DISABLED' : 'ENABLED'
      permission.updateTime = now()
      sendJson(res, ok(null))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/http-records$/,
    handler: async ({ req, res, query }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const keyword = (query.get('keyword') ?? '').trim().toLowerCase()
      const method = query.get('method') ?? ''
      const statusClass = query.get('statusClass') ?? ''
      const startTime = Number(query.get('startTime')) || 0
      const endTime = Number(query.get('endTime')) || 0

      const filtered = httpRecords.filter((record) => {
        const matchKeyword =
          !keyword ||
          record.path.toLowerCase().includes(keyword) ||
          record.traceId.toLowerCase().includes(keyword) ||
          record.clientIp.toLowerCase().includes(keyword)
        const matchStatusClass =
          !statusClass || String(Math.floor(record.statusCode / 100)) === statusClass.charAt(0)
        const matchTime =
          (!startTime || record.requestAt >= startTime) && (!endTime || record.requestAt <= endTime)
        return matchKeyword && (!method || record.method === method) && matchStatusClass && matchTime
      })
      sendJson(res, paginate(filtered, query))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/http-records\/([^/]+)$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const record = httpRecords.find((item) => item.id === params[0])
      if (!record) return sendJson(res, fail('访问记录不存在', 'ADMIN_HTTP_RECORD_NOT_FOUND'))
      sendJson(res, ok(record))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/user\/([0-9]+)\/role$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const user = users.find((item) => item.id === params[0])
      if (!user) return sendJson(res, fail('用户不存在', 'ADMIN_USER_NOT_FOUND'))
      sendJson(
        res,
        ok({
          roles,
          roleIds: user.roles.map((role) => role.id),
        }),
      )
    },
  },
  {
    method: 'PUT',
    pattern: /^\/api\/system\/user\/([0-9]+)\/role$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const user = users.find((item) => item.id === params[0])
      if (!user) return sendJson(res, fail('用户不存在', 'ADMIN_USER_NOT_FOUND'))
      const roleIds = Array.isArray(body.roleIds) ? (body.roleIds as string[]) : []
      user.roles = roles.filter((role) => roleIds.includes(role.id))
      sendJson(res, ok(null))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/system\/audit-log$/,
    handler: async ({ req, res, query }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const keyword = (query.get('keyword') ?? '').trim().toLowerCase()
      const eventCode = query.get('eventCode') ?? ''
      const phase = query.get('phase') ?? ''
      const result = query.get('result') ?? ''
      const auditId = (query.get('auditId') ?? '').trim()
      const startTime = Number(query.get('startTime')) || 0
      const endTime = Number(query.get('endTime')) || 0

      const filtered = auditLogs.filter((item) => {
        const matchKeyword =
          !keyword ||
          item.operatorName.toLowerCase().includes(keyword) ||
          item.clientIp.includes(keyword)
        const matchTime =
          (!startTime || item.createTime >= startTime) && (!endTime || item.createTime <= endTime)
        return (
          matchKeyword &&
          (!eventCode || item.eventCode === eventCode) &&
          (!phase || item.phase === phase) &&
          (!result || item.result === result) &&
          (!auditId || item.auditId === auditId) &&
          matchTime
        )
      })
      sendJson(res, paginate(filtered, query))
    },
  },
]

export function mockPlugin(enabled: boolean): Plugin {
  return {
    name: 'yomi-mock',
    apply: 'serve',
    configureServer(server) {
      if (!enabled) return
      server.middlewares.use((req, res, next) => {
        const url = req.url ?? ''
        const method = (req.method ?? 'GET').toUpperCase()
        const path = url.split('?')[0] ?? '/'
        const route = routes.find((item) => item.method === method && item.pattern.test(path))
        if (!route) return next()
        const match = route.pattern.exec(path)
        const query = new URL(url, 'http://localhost').searchParams
        void route
          .handler({
            req,
            res,
            params: match ? match.slice(1).filter((item): item is string => item !== undefined) : [],
            query,
          })
          .catch(() => sendJson(res, fail('mock 内部错误')))
      })
    },
  }
}

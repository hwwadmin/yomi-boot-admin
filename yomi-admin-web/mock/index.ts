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
  total?: number
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

const ok = <T>(data: T, total?: number): MockResponse<T> => ({
  code: 'OK',
  msg: 'success',
  timestamp: Date.now(),
  data,
  total,
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
  { id: 'p2', code: 'system:role', name: '角色管理', type: 'MENU', uiPath: '/system/role', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p3', code: 'system:permission', name: '权限管理', type: 'MENU', uiPath: '/system/permission', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p4', code: 'system:http-record', name: '访问记录', type: 'MENU', uiPath: '/system/http-record', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p11', code: 'system:user:list', name: '用户列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/users', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p12', code: 'system:user:create', name: '创建用户', type: 'API', httpMethod: 'POST', httpPath: '/api/system/users', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p13', code: 'system:user:update', name: '更新用户', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/users/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p14', code: 'system:user:delete', name: '删除用户', type: 'API', httpMethod: 'DELETE', httpPath: '/api/system/users/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p21', code: 'system:role:list', name: '角色列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/role', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p22', code: 'system:role:detail', name: '角色详情', type: 'API', httpMethod: 'GET', httpPath: '/api/system/role/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p23', code: 'system:role:create', name: '创建角色', type: 'API', httpMethod: 'POST', httpPath: '/api/system/role', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p24', code: 'system:role:update', name: '更新角色', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/role/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p25', code: 'system:role:bind-permission', name: '角色绑定权限', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/role/:id/permission', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p31', code: 'system:permission:list', name: '权限列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/permission', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p32', code: 'system:permission:status', name: '权限启停', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/permission/:id/status', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p41', code: 'system:user:role:list', name: '用户角色查询', type: 'API', httpMethod: 'GET', httpPath: '/api/system/user/:id/role', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p42', code: 'system:user:role:bind', name: '用户角色绑定', type: 'API', httpMethod: 'PUT', httpPath: '/api/system/user/:id/role', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p51', code: 'system:http-record:list', name: '访问记录列表', type: 'API', httpMethod: 'GET', httpPath: '/api/system/http-records', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
  { id: 'p52', code: 'system:http-record:detail', name: '访问记录详情', type: 'API', httpMethod: 'GET', httpPath: '/api/system/http-records/:id', status: 'ENABLED', remark: '', createTime: now(), updateTime: now() },
]

// ———————— 角色 ————————
const roles: MockRole[] = [
  {
    id: 'r1',
    code: 'super_admin',
    name: '超级管理员',
    builtin: true,
    status: 'ENABLED',
    remark: '内置角色，拥有全部权限',
    permissionIds: [],
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
    permissionIds: ['p1', 'p11'],
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

function visibleMenus(user: MockUser): MockPermission[] {
  if (isSuper(user)) {
    return permissions.filter((item) => item.type === 'MENU' && item.status === 'ENABLED')
  }
  const ids = new Set(user.roles.flatMap((role) => role.permissionIds))
  return permissions.filter(
    (item) => item.type === 'MENU' && item.status === 'ENABLED' && ids.has(item.id),
  )
}

function grantedCodes(user: MockUser): string[] {
  if (isSuper(user)) return ['*']
  const ids = new Set(user.roles.flatMap((role) => role.permissionIds))
  return permissions
    .filter((item) => item.status === 'ENABLED' && ids.has(item.id))
    .map((item) => item.code)
}

function identity(user: MockUser) {
  return {
    username: user.username,
    nickname: user.nickname,
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

/** 按页码切分并返回 data + total 顶层分页结构 */
function paginate<T>(items: T[], query: URLSearchParams): MockResponse<T[]> {
  const page = Math.max(1, Number(query.get('page')) || 1)
  const pageSize = Math.min(200, Math.max(1, Number(query.get('pageSize')) || 10))
  const start = (page - 1) * pageSize
  return ok(items.slice(start, start + pageSize), items.length)
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

const routes: MockRoute[] = [
  {
    method: 'POST',
    pattern: /^\/api\/auth\/login$/,
    handler: async ({ req, res }) => {
      const body = await readBody(req)
      const username = String(body.username ?? '')
      const password = String(body.password ?? '')
      await delay(200)
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
    pattern: /^\/api\/auth\/logout$/,
    handler: async ({ req, res }) => {
      const token = bearerToken(req)
      if (token) tokenToUsername.delete(token)
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
      sendJson(res, ok(role))
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
        permissionIds: [],
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
    pattern: /^\/api\/system\/role\/([^/]+)\/permission$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail('登录已过期，请重新登录', 'TOKEN_REQUIRED'))
      const body = await readBody(req)
      const role = roles.find((item) => item.id === params[0])
      if (!role) return sendJson(res, fail('角色不存在', 'ADMIN_ROLE_NOT_FOUND'))
      if (role.builtin) return sendJson(res, fail('内置角色不允许变更', 'ADMIN_BUILTIN_ROLE_IMMUTABLE'))
      role.permissionIds = Array.isArray(body.permissionIds)
        ? (body.permissionIds as string[])
        : []
      role.updateTime = now()
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

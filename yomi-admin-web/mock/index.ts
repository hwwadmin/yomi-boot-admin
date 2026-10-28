/**
 * 极简 dev mock：以 vite 中间件方式拦截 /api 请求，不引入额外依赖。
 * - 仅在 `apply: 'serve'`（dev）下生效，且由 vite.config.ts 传入的 enabled 控制
 * - 数据保存在内存中，重启 dev server 即重置
 * - 后端就绪后把 .env.development 里 VITE_USE_MOCK 改为 false 即可切换到真实接口
 */
import type { Connect, Plugin } from 'vite'
import type { ServerResponse } from 'node:http'

/** 与 src/api/types.ts 的 ApiResponse 保持一致 */
interface MockResponse<T = unknown> {
  code: number
  message: string
  data: T
}

interface MockUser {
  id: number
  username: string
  nickname: string
  email: string
  phone: string
  status: 0 | 1
  createdAt: string
}

const ok = <T>(data: T): MockResponse<T> => ({ code: 200, message: 'ok', data })
const fail = (code: number, message: string): MockResponse<null> => ({ code, message, data: null })
const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

// ———————— 内存数据 ————————
const users: MockUser[] = Array.from({ length: 28 }, (_, index) => {
  const id = index + 1
  return {
    id,
    username: `user${String(id).padStart(3, '0')}`,
    nickname: `用户${id}`,
    email: `user${id}@yomi.dev`,
    phone: `138${String(10000000 + id * 7).slice(0, 8)}`,
    status: id % 7 === 0 ? 0 : 1,
    createdAt: `2026-09-${String((id % 28) + 1).padStart(2, '0')} 10:${String((id * 3) % 60).padStart(2, '0')}:00`,
  }
})
let nextId = users.length + 1

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

function currentUser(req: Connect.IncomingMessage): string | null {
  const token = bearerToken(req)
  return token ? (tokenToUsername.get(token) ?? null) : null
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

// ———————— 路由 ————————
const routes: MockRoute[] = [
  {
    method: 'POST',
    pattern: /^\/api\/auth\/login$/,
    handler: async ({ req, res }) => {
      const body = await readBody(req)
      const username = String(body.username ?? '')
      const password = String(body.password ?? '')
      await delay(300)
      if (username !== 'admin' || password !== 'admin123') {
        return sendJson(res, fail(400, '用户名或密码错误'))
      }
      const token = `mock-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 8)}`
      tokenToUsername.set(token, username)
      sendJson(res, ok({ token, user: { username: 'admin', nickname: '超级管理员', permissions: ['*'] } }))
    },
  },
  {
    method: 'GET',
    pattern: /^\/api\/auth\/profile$/,
    handler: async ({ req, res }) => {
      await delay(100)
      const username = currentUser(req)
      if (!username) return sendJson(res, fail(401, '登录已过期，请重新登录'))
      sendJson(res, ok({ username, nickname: '超级管理员', permissions: ['*'] }))
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
      await delay(200)
      if (!currentUser(req)) return sendJson(res, fail(401, '登录已过期，请重新登录'))
      const keyword = (query.get('keyword') ?? '').trim().toLowerCase()
      const page = Math.max(1, Number(query.get('page')) || 1)
      const pageSize = Math.min(100, Math.max(1, Number(query.get('pageSize')) || 10))
      const filtered = users.filter(
        (user) =>
          !keyword ||
          user.username.toLowerCase().includes(keyword) ||
          user.nickname.toLowerCase().includes(keyword),
      )
      const start = (page - 1) * pageSize
      sendJson(res, ok({ list: filtered.slice(start, start + pageSize), total: filtered.length, page, pageSize }))
    },
  },
  {
    method: 'POST',
    pattern: /^\/api\/system\/users$/,
    handler: async ({ req, res }) => {
      if (!currentUser(req)) return sendJson(res, fail(401, '登录已过期，请重新登录'))
      const body = await readBody(req)
      const username = String(body.username ?? '').trim()
      if (!username) return sendJson(res, fail(400, '用户名不能为空'))
      if (users.some((user) => user.username === username)) {
        return sendJson(res, fail(400, `用户名 ${username} 已存在`))
      }
      const now = new Date()
      const pad = (n: number) => String(n).padStart(2, '0')
      users.unshift({
        id: nextId++,
        username,
        nickname: String(body.nickname ?? username),
        email: String(body.email ?? ''),
        phone: String(body.phone ?? ''),
        status: body.status === 0 ? 0 : 1,
        createdAt: `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`,
      })
      sendJson(res, ok(null))
    },
  },
  {
    method: 'PUT',
    pattern: /^\/api\/system\/users\/(\d+)$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail(401, '登录已过期，请重新登录'))
      const body = await readBody(req)
      const target = users.find((user) => user.id === Number(params[0]))
      if (!target) return sendJson(res, fail(404, '用户不存在'))
      if (body.nickname !== undefined) target.nickname = String(body.nickname)
      if (body.email !== undefined) target.email = String(body.email)
      if (body.phone !== undefined) target.phone = String(body.phone)
      if (body.status !== undefined) target.status = body.status === 1 ? 1 : 0
      sendJson(res, ok(null))
    },
  },
  {
    method: 'DELETE',
    pattern: /^\/api\/system\/users\/(\d+)$/,
    handler: async ({ req, res, params }) => {
      if (!currentUser(req)) return sendJson(res, fail(401, '登录已过期，请重新登录'))
      const index = users.findIndex((user) => user.id === Number(params[0]))
      if (index === -1) return sendJson(res, fail(404, '用户不存在'))
      users.splice(index, 1)
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
          .catch(() => sendJson(res, fail(500, 'mock 内部错误')))
      })
    },
  }
}

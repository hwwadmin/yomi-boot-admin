import { existsSync, readdirSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const frontendRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const backendRoot = resolve(frontendRoot, '..', 'yomi-admin-backend')
const outputFile = resolve(frontendRoot, 'src/api/errorCodes.ts')
const errorDictionaryFiles = [
  resolve(frontendRoot, 'src/i18n/locales/zh-CN/errors.json'),
  resolve(frontendRoot, 'src/i18n/locales/en-US/errors.json'),
]

/** 框架层可能出现在响应里的通用 code，不在管理端源码里定义 */
const FRAMEWORK_CODES = [
  'OK',
  'INTERNAL_ERROR',
  'INVALID_BODY',
  'INVALID_INPUT',
  'REQUEST_TIMEOUT',
  'REQUEST_BODY_TOO_LARGE',
  'TOKEN_REQUIRED',
  'TOKEN_VALIDATE_FAILED',
  'TOKEN_INVALID',
]

const SKIP_DIRS = new Set(['.git', '.tmp', 'node_modules', 'vendor', 'dist'])

function collectGoFiles(dir, files = []) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    if (entry.isDirectory()) {
      if (!SKIP_DIRS.has(entry.name)) collectGoFiles(join(dir, entry.name), files)
      continue
    }
    if (entry.isFile() && entry.name.endsWith('.go') && !entry.name.endsWith('_test.go')) {
      files.push(join(dir, entry.name))
    }
  }
  return files
}

function extractBackendCodes() {
  if (!existsSync(backendRoot)) {
    throw new Error(`未找到后端源码目录：${backendRoot}`)
  }

  const codes = new Set(FRAMEWORK_CODES)
  const pattern = /NewBizf?\("([A-Z][A-Z0-9_]+)"/g
  for (const file of collectGoFiles(backendRoot)) {
    const source = readFileSync(file, 'utf8')
    for (const match of source.matchAll(pattern)) {
      codes.add(match[1])
    }
  }
  return [...codes].sort()
}

function render(codes) {
  const entries = codes.map((code) => `  '${code}',`).join('\n')
  return `// 由 scripts/check-error-codes.mjs 从后端源码生成，请勿手改。\n// 后端新增错误码后运行 pnpm check:error-codes 重新生成。\nexport const ERROR_CODES = [\n${entries}\n] as const\n\nexport type ErrorCode = (typeof ERROR_CODES)[number]\n`
}

function verifyDictionaries(codes) {
  let failed = false
  for (const file of errorDictionaryFiles) {
    const messages = JSON.parse(readFileSync(file, 'utf8'))
    const keys = Object.keys(messages.codes ?? {}).sort()
    const missing = codes.filter((code) => !keys.includes(code))
    const extra = keys.filter((code) => !codes.includes(code))

    if (!missing.length && !extra.length) continue
    failed = true
    console.error(`${relative(frontendRoot, file)} 的错误码字典与后端不一致`)
    if (missing.length) console.error(`  缺少：${missing.join(', ')}`)
    if (extra.length) console.error(`  多余：${extra.join(', ')}`)
  }
  return !failed
}

const codes = extractBackendCodes()
const content = render(codes)

if (process.argv.includes('--check')) {
  const current = existsSync(outputFile) ? readFileSync(outputFile, 'utf8') : ''
  if (current !== content) {
    console.error('src/api/errorCodes.ts 与后端源码不一致，请运行 pnpm check:error-codes 重新生成。')
    process.exit(1)
  }
} else {
  writeFileSync(outputFile, content)
}

if (!verifyDictionaries(codes)) {
  process.exit(1)
}

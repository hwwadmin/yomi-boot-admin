import { readdirSync, readFileSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const frontendRoot = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const srcRoot = resolve(frontendRoot, 'src')
const localesRoot = resolve(srcRoot, 'i18n/locales')
const localeCodes = ['zh-CN', 'en-US']
const baseLocale = 'zh-CN'

const CJK = /[\u4e00-\u9fa5]/
const issues = []

function walkFiles(dir, extensions, files = []) {
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const fullPath = join(dir, entry.name)
    if (entry.isDirectory()) {
      walkFiles(fullPath, extensions, files)
      continue
    }
    if (extensions.some((extension) => entry.name.endsWith(extension))) {
      files.push(fullPath)
    }
  }
  return files
}

function flattenKeys(value, prefix = '') {
  const keys = []
  for (const [key, child] of Object.entries(value)) {
    const path = prefix ? `${prefix}.${key}` : key
    if (child !== null && typeof child === 'object' && !Array.isArray(child)) {
      keys.push(...flattenKeys(child, path))
      continue
    }
    keys.push(path)
  }
  return keys
}

/** locale 文件在消息树中的命名空间：system/user.json -> system.user */
function namespaceOf(file) {
  const path = file.replace(/\.json$/, '').replaceAll('\\', '/')
  return path.startsWith('system/') ? path.replace('/', '.') : path
}

function listLocaleFiles(locale) {
  const root = resolve(localesRoot, locale)
  return walkFiles(root, ['.json'])
    .map((file) => relative(root, file).replaceAll('\\', '/'))
    .sort()
}

function checkLocaleParity() {
  const baseFiles = listLocaleFiles(baseLocale)

  for (const locale of localeCodes) {
    if (locale === baseLocale) continue
    const files = listLocaleFiles(locale)
    const missingFiles = baseFiles.filter((file) => !files.includes(file))
    const extraFiles = files.filter((file) => !baseFiles.includes(file))

    if (missingFiles.length || extraFiles.length) {
      issues.push(
        `${locale} 语言包文件与 ${baseLocale} 不一致` +
          (missingFiles.length ? `；缺少：${missingFiles.join(', ')}` : '') +
          (extraFiles.length ? `；多余：${extraFiles.join(', ')}` : ''),
      )
    }
  }

  for (const file of baseFiles) {
    const base = JSON.parse(readFileSync(resolve(localesRoot, baseLocale, file), 'utf8'))
    const baseKeys = new Set(flattenKeys(base))

    for (const locale of localeCodes) {
      if (locale === baseLocale) continue
      const targetPath = resolve(localesRoot, locale, file)
      let target
      try {
        target = JSON.parse(readFileSync(targetPath, 'utf8'))
      } catch {
        continue
      }
      const targetKeys = new Set(flattenKeys(target))
      const missing = [...baseKeys].filter((key) => !targetKeys.has(key))
      const extra = [...targetKeys].filter((key) => !baseKeys.has(key))

      if (missing.length) issues.push(`${locale}/${file} 缺少键：${missing.join(', ')}`)
      if (extra.length) issues.push(`${locale}/${file} 多余键：${extra.join(', ')}`)
    }
  }
}

function lineOf(source, index) {
  return source.slice(0, index).split('\n').length
}

function checkTemplateLiterals() {
  for (const file of walkFiles(srcRoot, ['.vue'])) {
    const source = readFileSync(file, 'utf8')
    const start = source.indexOf('<template>')
    const end = source.lastIndexOf('</template>')
    if (start === -1 || end === -1) continue

    const template = source
      .slice(start + '<template>'.length, end)
      .replace(/<!--[\s\S]*?-->/g, '')
    const matchIndex = template.search(CJK)
    if (matchIndex === -1) continue

    const absoluteIndex = start + '<template>'.length + matchIndex
    issues.push(`${relative(frontendRoot, file)}:${lineOf(source, absoluteIndex)} 模板中仍有中文文案`)
  }
}

function checkScriptLiterals() {
  const patterns = [
    /ElMessage(?:Box)?\.\w+\(\s*['"`][^'"`]*[\u4e00-\u9fa5]/,
    /(?:message|title|description|placeholder|label)\s*:\s*['"`][^'"`]*[\u4e00-\u9fa5]/,
  ]

  for (const file of walkFiles(srcRoot, ['.ts', '.vue'])) {
    if (file.includes(`${join('src', 'i18n')}`)) continue
    const source = readFileSync(file, 'utf8')
    for (const pattern of patterns) {
      const match = pattern.exec(source)
      if (!match) continue
      issues.push(`${relative(frontendRoot, file)}:${lineOf(source, match.index)} 代码中仍有中文提示文案`)
      break
    }
  }
}

function checkTranslationKeys() {
  const keys = new Set()
  for (const file of listLocaleFiles(baseLocale)) {
    const messages = JSON.parse(readFileSync(resolve(localesRoot, baseLocale, file), 'utf8'))
    const namespace = namespaceOf(file)
    for (const key of flattenKeys(messages)) keys.add(`${namespace}.${key}`)
  }

  const patterns = [/\bt\(\s*['"]([\w.-]+)['"]/g, /titleKey:\s*['"]([\w.-]+)['"]/g]
  for (const file of walkFiles(srcRoot, ['.ts', '.vue'])) {
    if (file.includes(`${join('src', 'i18n')}`)) continue
    const source = readFileSync(file, 'utf8')
    for (const pattern of patterns) {
      for (const match of source.matchAll(pattern)) {
        if (!keys.has(match[1])) {
          issues.push(`${relative(frontendRoot, file)}:${lineOf(source, match.index)} 未定义的语言键 ${match[1]}`)
        }
      }
    }
  }
}

checkLocaleParity()
checkTemplateLiterals()
checkScriptLiterals()
checkTranslationKeys()

if (issues.length) {
  console.error('i18n 检查未通过：')
  for (const issue of issues) console.error(`  - ${issue}`)
  process.exit(1)
}

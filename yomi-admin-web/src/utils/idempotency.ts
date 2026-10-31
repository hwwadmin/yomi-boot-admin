/** 幂等请求头名，与后端 support.IdempotencyKeyHeader 保持一致 */
export const IDEMPOTENCY_KEY_HEADER = 'X-Idempotency-Key'

/**
 * 生成一次提交意图的幂等键。
 *
 * 同一次提交（含超时重试、失败后重试）必须复用同一个值；提交成功后应清空，
 * 下一次提交重新生成，避免把上一次的成功结果回放给新的表单内容。
 */
export function newIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  // crypto.randomUUID 只在安全上下文可用，局域网 http 访问时退回时间戳 + 随机串
  return `k-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`
}

/** 构造幂等请求头；键为空时返回 undefined（后端不做幂等保护） */
export function idempotencyHeaders(key: string): Record<string, string> | undefined {
  return key ? { [IDEMPOTENCY_KEY_HEADER]: key } : undefined
}

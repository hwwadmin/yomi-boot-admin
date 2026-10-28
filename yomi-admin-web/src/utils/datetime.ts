/**
 * 把后端返回的 Unix 毫秒时间戳格式化为客户端本地时区的 yyyy-MM-dd HH:mm:ss。
 * 后端统一使用 UTC，时区转换只在前端发生。
 */
export function formatDateTime(timestamp: number): string {
  if (!timestamp) return '-'

  const date = new Date(timestamp)
  const pad = (value: number) => String(value).padStart(2, '0')

  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
}

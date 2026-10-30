export const SUPPORTED_LOCALES = ['zh-CN', 'en-US'] as const
export type LocaleCode = (typeof SUPPORTED_LOCALES)[number]

export function isLocaleCode(value: unknown): value is LocaleCode {
  return typeof value === 'string' && (SUPPORTED_LOCALES as readonly string[]).includes(value)
}

import { i18n } from '@/i18n'

interface ConfirmActionOptions {
  message: string
  title?: string
  confirmButtonText?: string
  type?: 'success' | 'warning' | 'info' | 'error'
}

/**
 * 危险操作二次确认：用户取消或关闭弹窗时返回 false，不抛错。
 */
export async function confirmAction(options: ConfirmActionOptions): Promise<boolean> {
  try {
    await ElMessageBox.confirm(
      options.message,
      options.title ?? i18n.global.t('common.confirmTitle'),
      {
        type: options.type ?? 'warning',
        confirmButtonText: options.confirmButtonText ?? i18n.global.t('common.action.confirm'),
        cancelButtonText: i18n.global.t('common.action.cancel'),
      },
    )
    return true
  } catch {
    return false
  }
}

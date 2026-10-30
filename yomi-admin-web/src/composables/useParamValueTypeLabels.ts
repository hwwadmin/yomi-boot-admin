import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import type { ParamValueType } from '@/api/types'

/** 参数值类型展示名；按展示顺序声明（对象键序遍历，字符串放第一位） */
export function useParamValueTypeLabels() {
  const { t } = useI18n()

  return computed<Record<ParamValueType, string>>(() => ({
    STRING: t('system.param.valueType.string'),
    BOOL: t('system.param.valueType.bool'),
    INT: t('system.param.valueType.int'),
    JSON: t('system.param.valueType.json'),
  }))
}

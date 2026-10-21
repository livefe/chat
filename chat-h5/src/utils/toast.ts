import { showFailToast, showToast } from 'vant'

import { getErrorMessage, isReportedByRequestLayer } from '@/utils/apiError'

export function showCatchError(err: unknown, fallback = '操作失败', useFail = false) {
  if (isReportedByRequestLayer(err)) return
  const message = getErrorMessage(err, fallback)
  if (useFail) showFailToast(message)
  else showToast(message)
}

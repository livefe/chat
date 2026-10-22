import { ElMessage, ElMessageBox } from 'element-plus'

import { showApiError } from '@/utils/apiError'

export function isCancelAction(e: unknown) {
  return e === 'cancel' || e === 'close'
}

export async function confirmAndRun(
  message: string,
  action: () => Promise<void>,
  successMessage = '操作成功',
) {
  try {
    await ElMessageBox.confirm(message, '提示', { type: 'warning' })
  } catch {
    return
  }
  try {
    await action()
    ElMessage.success(successMessage)
  } catch (e: unknown) {
    showApiError(e, '操作失败')
  }
}

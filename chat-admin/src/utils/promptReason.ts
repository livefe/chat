import { ElMessageBox } from 'element-plus'

export async function confirmWithReason(
  message: string,
  options?: {
    title?: string
    required?: boolean
    inputPlaceholder?: string
  },
): Promise<string | null> {
  try {
    const { value } = await ElMessageBox.prompt(message, options?.title ?? '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPlaceholder: options?.inputPlaceholder ?? '请输入原因（选填）',
      inputValidator: options?.required ? (val) => (val?.trim() ? true : '请填写原因') : undefined,
    })
    return value?.trim() ?? ''
  } catch {
    return null
  }
}

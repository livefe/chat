import { ElMessage } from 'element-plus'

export function showApiError(e: unknown, fallback: string) {
  const message = e instanceof Error && e.message ? e.message : fallback
  ElMessage.error(message)
}

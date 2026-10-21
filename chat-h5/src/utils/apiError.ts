interface ApiRejectPayload {
  code?: number
  message?: string
  data?: unknown
}

function isApiReject(err: unknown): err is ApiRejectPayload {
  return (
    !!err &&
    typeof err === 'object' &&
    'code' in err &&
    typeof (err as ApiRejectPayload).code === 'number'
  )
}

function isAxiosLike(err: unknown): err is { response?: { data?: { message?: string } } } {
  return !!err && typeof err === 'object' && 'response' in err
}

export function getErrorMessage(err: unknown, fallback = '操作失败'): string {
  if (isApiReject(err) && err.message) return err.message
  if (err instanceof Error && err.message) return err.message
  if (isAxiosLike(err) && err.response?.data?.message) return err.response.data.message
  return fallback
}

export function isReportedByRequestLayer(err: unknown): boolean {
  if (isApiReject(err)) return true
  return isAxiosLike(err) && !!err.response
}

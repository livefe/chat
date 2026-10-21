import type { CheckInResultPayload, PostCheckInResult } from '@/api/post'
import { formatDateTimeChinese } from '@/utils/datetime'

export function buildCheckInResultPayload(
  kind: CheckInResultPayload['kind'],
  data?: PostCheckInResult,
  message = '',
): CheckInResultPayload {
  if (kind === 'fail') {
    return { kind, message }
  }
  return {
    kind,
    postId: data?.postId,
    postTitle: data?.postTitle || '',
    userId: data?.userId,
    userNickname: data?.userNickname || '',
    userAvatarUrl: data?.userAvatarUrl || '',
    signupAt: data?.signupAt || '',
    checkedInAtText: data?.checkedInAtText || '',
    operatorNickname: data?.operatorNickname || '',
    operatorAvatarUrl: data?.operatorAvatarUrl || '',
  }
}

export function encodeCheckInResultPayload(payload: CheckInResultPayload): string {
  return encodeURIComponent(JSON.stringify(payload))
}

export function parseCheckInResultPayload(
  raw: string | null | undefined,
): CheckInResultPayload | null {
  if (!raw) return null
  try {
    return JSON.parse(decodeURIComponent(raw)) as CheckInResultPayload
  } catch {
    return null
  }
}

export function formatCheckInDisplayTime(v: string): string {
  return formatDateTimeChinese(v, true) || v
}

export function checkInStatusText(kind: string): string {
  if (kind === 'fail') return '核验失败'
  if (kind === 'already') return '已核验过'
  return '核验成功'
}

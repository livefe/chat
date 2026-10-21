import { isSameDay, pad, toUnixMs } from '@/utils/datetime'

const MESSAGE_TIME_DIVIDER_GAP_MS = 5 * 60 * 1000

export function formatChatBubbleTime(ts: number): string {
  if (!ts) return ''
  const d = new Date(toUnixMs(ts))
  const now = new Date()
  const hm = `${pad(d.getHours())}:${pad(d.getMinutes())}`

  if (isSameDay(d, now)) return hm

  const yesterday = new Date(now)
  yesterday.setDate(yesterday.getDate() - 1)
  if (isSameDay(d, yesterday)) return `昨天 ${hm}`

  if (d.getFullYear() === now.getFullYear()) {
    return `${d.getMonth() + 1}月${d.getDate()}日 ${hm}`
  }

  return `${d.getFullYear()}/${d.getMonth() + 1}/${d.getDate()} ${hm}`
}

function isValidTs(ts: number | undefined | null): ts is number {
  return typeof ts === 'number' && ts > 0
}

export function shouldShowTimeDivider(
  prevTs: number | undefined | null,
  currTs: number,
  opts?: { isFirst?: boolean; gapMs?: number },
): boolean {
  if (!isValidTs(currTs)) return false
  if (opts?.isFirst) return true

  const gapMs = opts?.gapMs ?? MESSAGE_TIME_DIVIDER_GAP_MS
  if (!isValidTs(prevTs)) return false
  const prev = toUnixMs(prevTs)
  const curr = toUnixMs(currTs)
  if (curr - prev >= gapMs) return true

  return !isSameDay(new Date(prev), new Date(curr))
}

export function formatConversationTime(ts: number): string {
  if (!ts) return ''
  const diff = Date.now() - toUnixMs(ts)
  if (diff < 60_000) return '刚刚'
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}分钟前`
  return formatChatBubbleTime(ts)
}

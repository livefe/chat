import type { AdminMessageItem } from '@/api/message'
import { MessageContentType } from '@/constants/message'

export function formatMessageLocation(content: string | undefined) {
  if (!content) return '[位置]'
  try {
    const raw = JSON.parse(content) as { placeName?: string; latitude?: number; longitude?: number }
    const name = String(raw.placeName || '').trim() || '位置'
    if (typeof raw.latitude === 'number' && typeof raw.longitude === 'number') {
      return `[位置] ${name} (${raw.latitude.toFixed(5)}, ${raw.longitude.toFixed(5)})`
    }
    return `[位置] ${name}`
  } catch {
    return '[位置]'
  }
}

export function formatMessagePost(content: string | undefined) {
  if (!content) return '[动态]'
  try {
    const raw = JSON.parse(content) as { title?: string; postId?: number }
    const title = String(raw.title || '').trim()
    return title ? `[动态] ${title}` : '[动态]'
  } catch {
    return '[动态]'
  }
}

export function formatMessageVoice(content: string | undefined) {
  if (!content) return '[语音]'
  try {
    const raw = JSON.parse(content) as { durationMs?: number }
    const ms = typeof raw.durationMs === 'number' ? Math.round(raw.durationMs) : 0
    if (ms > 0) {
      const sec = Math.max(1, Math.round(ms / 1000))
      return `[语音] ${sec}"`
    }
    return '[语音]'
  } catch {
    return '[语音]'
  }
}

export function parseMessageMediaUrl(content: string | undefined): string {
  if (!content) return ''
  try {
    const raw = JSON.parse(content) as { url?: string }
    return String(raw.url || '').trim()
  } catch {
    return ''
  }
}

export function formatMessageFile(content: string | undefined) {
  if (!content) return '[文件]'
  try {
    const raw = JSON.parse(content) as { name?: string }
    const name = String(raw.name || '').trim()
    return name ? `[文件] ${name}` : '[文件]'
  } catch {
    return '[文件]'
  }
}

export function formatMessageSummary(contentType: number, content: string) {
  switch (contentType) {
    case MessageContentType.Post:
      return formatMessagePost(content)
    case MessageContentType.Location:
      return formatMessageLocation(content)
    case MessageContentType.File:
      return formatMessageFile(content)
    case MessageContentType.Video:
      return '[视频]'
    case MessageContentType.Voice:
      return formatMessageVoice(content)
    case MessageContentType.Image:
      return '[图片]'
    default:
      return content
  }
}

export function adminReplyLabel(item: AdminMessageItem) {
  if (!item.replyToId) return ''
  const c = item.replyToContent || ''
  if (!c) return item.replyToDeleted ? '原消息已删除' : ''
  if (
    item.replyToContentType === MessageContentType.Image ||
    item.replyToContentType === MessageContentType.Voice ||
    item.replyToContentType === MessageContentType.Video ||
    item.replyToContentType === MessageContentType.File ||
    item.replyToContentType === MessageContentType.Location ||
    item.replyToContentType === MessageContentType.Post
  ) {
    return formatMessageSummary(item.replyToContentType, c)
  }
  const text = c.trim()
  return text.length > 80 ? `${text.slice(0, 80)}…` : text
}

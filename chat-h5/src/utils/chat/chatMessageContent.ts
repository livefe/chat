import { MessageContentTypes } from '@/constants/chat'
import { parseChatMedia } from '@/utils/chat/chatMedia'
import { chatPostQuoteLabel, parseChatPost } from '@/utils/chat/chatPost'
import { voiceContentLabel } from '@/utils/chat/chatVoice'

export function messagePreview(
  contentType: number | undefined,
  content: string | undefined,
): string {
  if (contentType === MessageContentTypes.Image) return '[图片]'
  if (contentType === MessageContentTypes.Voice) return voiceContentLabel(content)
  if (contentType === MessageContentTypes.Video) return '[视频]'
  if (contentType === MessageContentTypes.Location) return '[位置]'
  if (contentType === MessageContentTypes.Post) return '[动态]'
  if (contentType === MessageContentTypes.File) {
    const name = chatFileDisplayName(content)
    return name === '文件' ? '[文件]' : `[文件] ${truncateFileDisplayName(name)}`
  }
  return (content || '').trim()
}

export function replyQuoteText(
  contentType: number | undefined,
  content: string | undefined,
): string {
  if (isImageMessage(contentType) || isVideoMessage(contentType)) {
    return messagePreview(contentType, content)
  }
  if (isLocationMessage(contentType)) return (content || '').trim()
  if (isVoiceMessage(contentType)) return voiceContentLabel(content)
  if (isPostMessage(contentType)) return chatPostQuoteLabel(parseChatPost(content))
  const preview = messagePreview(contentType, content)
  return preview.length > 80 ? `${preview.slice(0, 80)}…` : preview
}

export function messageClipboardText(
  contentType: number | undefined,
  content: string | undefined,
): string {
  if (contentType === MessageContentTypes.Location) {
    try {
      const loc = JSON.parse(content || '') as { placeName?: string }
      const name = (loc.placeName || '').trim()
      if (name) return name
    } catch {}
    return '[位置]'
  }
  if (isVoiceMessage(contentType)) return voiceContentLabel(content)
  if (isPostMessage(contentType)) return chatPostQuoteLabel(parseChatPost(content))
  if (isFileMessage(contentType)) {
    const name = chatFileDisplayName(content)
    return name === '文件' ? '[文件]' : `[文件] ${name}`
  }
  if (isImageMessage(contentType) || isVideoMessage(contentType)) {
    return messagePreview(contentType, content)
  }
  return (content || '').trim()
}

export function isImageMessage(contentType: number | undefined): boolean {
  return contentType === MessageContentTypes.Image
}

export function isVoiceMessage(contentType: number | undefined): boolean {
  return contentType === MessageContentTypes.Voice
}

export function isVideoMessage(contentType: number | undefined): boolean {
  return contentType === MessageContentTypes.Video
}

export function isLocationMessage(contentType: number | undefined): boolean {
  return contentType === MessageContentTypes.Location
}

export function isPostMessage(contentType: number | undefined): boolean {
  return contentType === MessageContentTypes.Post
}

export function isFileMessage(contentType: number | undefined): boolean {
  return contentType === MessageContentTypes.File
}

export function chatFileDisplayName(raw: string | undefined): string {
  const t = (raw || '').trim()
  if (!t) return '文件'
  try {
    const body = JSON.parse(t) as { name?: string; url?: string }
    const name = (body.name || '').trim()
    if (name) return name
    const fromUrl = (body.url || '').split('/').pop()?.split('?')[0] || ''
    if (fromUrl) return fromUrl
  } catch {}
  return '文件'
}

export function truncateFileDisplayName(name: string, maxLen = 28): string {
  const n = (name || '').trim() || '文件'
  if (n.length <= maxLen) return n
  const dot = n.lastIndexOf('.')
  const ext = dot > 0 && n.length - dot <= 9 ? n.slice(dot) : ''
  const budget = Math.max(8, maxLen - ext.length - 1)
  return `${n.slice(0, budget)}…${ext}`
}

export function replyQuoteMediaUrl(
  contentType: number | undefined,
  content: string | undefined,
): string {
  if (!isImageMessage(contentType) && !isVideoMessage(contentType)) return ''
  const media = parseChatMedia(content)
  return media?.url || ''
}

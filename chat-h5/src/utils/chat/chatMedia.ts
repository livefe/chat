export type ChatMediaContent = {
  mediaId: number
  url: string
}

export function parseChatMedia(raw?: string | null): ChatMediaContent | null {
  const t = (raw || '').trim()
  if (!t || !t.startsWith('{')) return null
  try {
    const data = JSON.parse(t) as { mediaId?: unknown; url?: unknown }
    const url = typeof data.url === 'string' ? data.url.trim() : ''
    if (
      typeof data.mediaId !== 'number' ||
      !Number.isInteger(data.mediaId) ||
      data.mediaId < 1 ||
      !url
    ) {
      return null
    }
    return { mediaId: data.mediaId, url }
  } catch {
    return null
  }
}

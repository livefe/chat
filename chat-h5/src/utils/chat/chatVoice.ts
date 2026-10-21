export const VOICE_MAX_DURATION_MS = 60_000
export const VOICE_MIN_DURATION_MS = 500

export type ChatVoice = {
  mediaId: number
  url: string
  durationMs: number
}

export function parseChatVoice(raw?: string | null): ChatVoice | null {
  const t = (raw || '').trim()
  if (!t) return null
  try {
    const data = JSON.parse(t) as { mediaId?: unknown; url?: unknown; durationMs?: unknown }
    if (typeof data.mediaId !== 'number' || !Number.isInteger(data.mediaId) || data.mediaId < 1) {
      return null
    }
    const url = typeof data.url === 'string' ? data.url.trim() : ''
    const durationMs = typeof data.durationMs === 'number' ? Math.round(data.durationMs) : 0
    if (durationMs < 1) return null
    return { mediaId: data.mediaId, url, durationMs }
  } catch {
    return null
  }
}

export function encodeChatVoice(mediaId: number, durationMs: number): string {
  return JSON.stringify({
    mediaId,
    durationMs: Math.max(1, Math.min(VOICE_MAX_DURATION_MS, Math.round(durationMs))),
  })
}

export function formatVoiceDuration(durationMs: number): string {
  const sec = Math.max(1, Math.round(durationMs / 1000))
  return `${sec}"`
}

function voiceQuoteLabel(durationMs: number): string {
  return `[语音] ${formatVoiceDuration(durationMs)}`
}

export function voiceContentLabel(raw?: string | null): string {
  const voice = parseChatVoice(raw)
  return voice ? voiceQuoteLabel(voice.durationMs) : '[语音]'
}

export function pickAudioMimeType(): string {
  if (typeof MediaRecorder === 'undefined') return ''
  const candidates = ['audio/webm;codecs=opus', 'audio/webm', 'audio/mp4', 'audio/ogg;codecs=opus']
  return candidates.find((t) => MediaRecorder.isTypeSupported(t)) || ''
}

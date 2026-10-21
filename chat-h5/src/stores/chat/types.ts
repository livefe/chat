import type { MessageItem } from '@/api/message'
import { CHAT_PAGE_LIMIT } from '@/constants/chat'

export type PostSubReason = 'feed'

export interface HistoryState {
  hasMore: boolean
  loading: boolean
}

export function emptyHistory(): HistoryState {
  return { hasMore: false, loading: false }
}

export function historyFromBatch(batch: MessageItem[]): HistoryState {
  return { hasMore: batch.length >= CHAT_PAGE_LIMIT, loading: false }
}

export type SendFailReason = 'invalid' | 'noLocation' | 'connectFailed' | 'sendFailed'

export type SendResult = { ok: true } | { ok: false; reason: SendFailReason }

export interface SendOptions {
  onConnecting?: () => void
}

export type ChatNotice =
  | { type: 'sendTimeout' }
  | { type: 'serverError'; message: string }
  | { type: 'syncFailed'; error: unknown }
  | { type: 'unreadFailed'; error: unknown }

export type ChatLocationSource = () => { lat: number; lng: number } | null

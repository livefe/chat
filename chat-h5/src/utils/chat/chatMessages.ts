import type { MessageItem } from '@/api/message'
import { messagePreview } from '@/utils/chat/chatMessageContent'

const MESSAGE_KEEP = 200

export function messageStableKey(m: MessageItem): string | number {
  if (m.id > 0) return m.id
  if (m.clientMsgId) return `c:${m.clientMsgId}`
  return `${m.createdAt}-${m.fromUserId}`
}

function compareMessage(a: MessageItem, b: MessageItem): number {
  if (a.createdAt !== b.createdAt) return a.createdAt - b.createdAt
  return a.id - b.id
}

function sortMessagesAsc(list: MessageItem[]) {
  list.sort(compareMessage)
}

export function normalizeMessages(list: MessageItem[]) {
  return list.slice().sort(compareMessage)
}

function trimMessageList(list: MessageItem[]) {
  if (list.length > MESSAGE_KEEP) list.splice(0, list.length - MESSAGE_KEEP)
}

function needsSortAround(list: MessageItem[], idx: number): boolean {
  const m = list[idx]
  if (!m) return false
  const prev = list[idx - 1]
  if (prev && compareMessage(prev, m) > 0) return true
  const next = list[idx + 1]
  return !!(next && compareMessage(m, next) > 0)
}

export function mergeMessageLists(existing: MessageItem[], incoming: MessageItem[]): MessageItem[] {
  const seen = new Set(incoming.map((m) => messageStableKey(m)))
  const merged = [...incoming]

  for (const m of existing) {
    const key = messageStableKey(m)
    if (seen.has(key)) continue
    merged.push(m)
    seen.add(key)
  }

  merged.sort(compareMessage)
  trimMessageList(merged)
  return merged
}

export function prependMessages(list: MessageItem[], older: MessageItem[]) {
  if (older.length === 0) return
  const existing = new Set(list.filter((m) => m.id > 0).map((m) => m.id))
  const toAdd = older.filter((m) => m.id > 0 && !existing.has(m.id))
  if (toAdd.length === 0) return
  list.unshift(...toAdd)
  sortMessagesAsc(list)
}

export function ensureMessageList(map: Map<number, MessageItem[]>, key: number): MessageItem[] {
  const existing = map.get(key)
  if (existing) return existing
  map.set(key, [])
  return map.get(key) ?? []
}

export function appendMessage(list: MessageItem[], msg: MessageItem) {
  if (msg.id > 0 && list.some((x) => x.id === msg.id)) return
  const prev = list[list.length - 1]
  list.push(msg)
  if (prev && compareMessage(prev, msg) > 0) sortMessagesAsc(list)
  trimMessageList(list)
}

export function replaceOptimisticOrAppend(
  list: MessageItem[],
  msg: MessageItem,
  selfId: number | undefined,
  clientMsgId?: string,
) {
  if (clientMsgId) {
    const byId = list.findIndex((m) => m.clientMsgId === clientMsgId && m.id <= 0)
    if (byId >= 0) {
      list[byId] = { ...msg, clientMsgId: undefined, sendStatus: undefined }
      if (needsSortAround(list, byId)) sortMessagesAsc(list)
      trimMessageList(list)
      return
    }
  }
  if (selfId && msg.fromUserId === selfId && msg.id > 0) {
    const idx = list.findIndex(
      (m) =>
        m.id <= 0 &&
        m.sendStatus === 'sending' &&
        m.fromUserId === selfId &&
        m.content === msg.content,
    )
    if (idx >= 0) {
      list[idx] = msg
      if (needsSortAround(list, idx)) sortMessagesAsc(list)
      trimMessageList(list)
      return
    }
  }
  appendMessage(list, msg)
}

export function applyAckToList(list: MessageItem[], clientMsgId: string, id: number) {
  const idx = list.findIndex((m) => m.clientMsgId === clientMsgId && m.id <= 0)
  const prev = list[idx]
  if (!prev) return false
  list[idx] = { ...prev, id, sendStatus: undefined, clientMsgId: undefined }
  if (needsSortAround(list, idx)) sortMessagesAsc(list)
  return true
}

export function markOptimisticFailed(list: MessageItem[], clientMsgId: string, reason?: string) {
  const idx = list.findIndex((m) => m.clientMsgId === clientMsgId && m.id <= 0)
  const prev = list[idx]
  if (!prev) return
  list[idx] = {
    ...prev,
    sendStatus: 'failed',
    sendFailReason: reason || undefined,
  }
}

export function markAllOptimisticFailed(list: MessageItem[]) {
  for (const [i, m] of list.entries()) {
    if (m.id <= 0 && m.sendStatus === 'sending') {
      list[i] = { ...m, sendStatus: 'failed' }
    }
  }
}

export function applyRecallToList(list: MessageItem[], recalledId: number) {
  if (recalledId <= 0) return
  for (const [i, m] of list.entries()) {
    if (m.id === recalledId) {
      if (!m.recalled) list[i] = { ...m, recalled: true }
    } else if (m.replyToId === recalledId && m.replyToContent !== '原消息已撤回') {
      list[i] = { ...m, replyToContent: '原消息已撤回', replyToContentType: 0 }
    }
  }
}

export function lastVisiblePreview(list: MessageItem[]): string {
  for (let i = list.length - 1; i >= 0; i--) {
    const m = list[i]
    if (m && m.id > 0 && !m.recalled) return messagePreview(m.contentType, m.content)
  }
  return '撤回了一条消息'
}

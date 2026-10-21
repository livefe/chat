import { computed, ref } from 'vue'

import {
  conversationUnreadSummary,
  markConversationRead,
  type ConversationItem,
} from '@/api/conversation'
import { READ_ACK_DEBOUNCE_MS } from '@/constants/chat'

import type { ChatNotice } from './types'

export function createUnreadHelpers(deps: { notify: (notice: ChatNotice) => void }) {
  const unreadByPeer = ref<Record<number, number>>({})
  const readAckTimers = new Map<number, ReturnType<typeof setTimeout>>()

  let fetchGeneration = 0

  function getUnread(peerId: number): number {
    return unreadByPeer.value[peerId] ?? 0
  }

  function setUnread(peerId: number, count: number) {
    if (count <= 0) {
      if (!(peerId in unreadByPeer.value)) return
      const next = { ...unreadByPeer.value }
      delete next[peerId]
      unreadByPeer.value = next
      return
    }
    unreadByPeer.value = { ...unreadByPeer.value, [peerId]: count }
  }

  function markPeerRead(peerId: number) {
    setUnread(peerId, 0)
  }

  function clearReadAckTimer(peerId: number) {
    const timer = readAckTimers.get(peerId)
    if (!timer) return
    clearTimeout(timer)
    readAckTimers.delete(peerId)
  }

  function clearAllReadAckTimers() {
    for (const timer of readAckTimers.values()) {
      clearTimeout(timer)
    }
    readAckTimers.clear()
  }

  function flushPeerRead(peerId: number) {
    if (peerId <= 0) return
    clearReadAckTimer(peerId)
    markPeerRead(peerId)
    void markConversationRead(peerId).catch(() => {})
  }

  function acknowledgePeerRead(peerId: number) {
    if (peerId <= 0) return
    markPeerRead(peerId)
    clearReadAckTimer(peerId)
    readAckTimers.set(
      peerId,
      setTimeout(() => {
        readAckTimers.delete(peerId)
        void markConversationRead(peerId).catch(() => {})
      }, READ_ACK_DEBOUNCE_MS),
    )
  }

  function incrPeerUnread(peerId: number) {
    setUnread(peerId, getUnread(peerId) + 1)
  }

  function applyUnreadSummary(items: { peerId: number; count: number }[]) {
    const next: Record<number, number> = {}
    for (const item of items) {
      if (item.count > 0) next[item.peerId] = item.count
    }
    unreadByPeer.value = next
  }

  function applyConversationUnread(list: ConversationItem[]) {
    const next = { ...unreadByPeer.value }
    for (const c of list) {
      if (c.unreadCount > 0) {
        next[c.peerId] = c.unreadCount
      } else {
        delete next[c.peerId]
      }
    }
    unreadByPeer.value = next
  }

  function clearAllUnread() {
    fetchGeneration += 1
    unreadByPeer.value = {}
  }

  async function refreshUnreadSummary() {
    const gen = fetchGeneration
    try {
      const res = await conversationUnreadSummary()
      if (gen !== fetchGeneration) return
      applyUnreadSummary(res.data.items ?? [])
    } catch (e) {
      if (gen !== fetchGeneration) return
      deps.notify({ type: 'unreadFailed', error: e })
    }
  }

  const totalUnread = computed(() =>
    Object.values(unreadByPeer.value).reduce((sum, n) => sum + (n > 0 ? n : 0), 0),
  )

  const tabUnreadBadge = computed(() => {
    const n = totalUnread.value
    if (n <= 0) return undefined
    return n > 99 ? '99+' : n
  })

  return {
    getUnread,
    markPeerRead,
    flushPeerRead,
    acknowledgePeerRead,
    incrPeerUnread,
    applyConversationUnread,
    clearAllUnread,
    clearAllReadAckTimers,
    refreshUnreadSummary,
    tabUnreadBadge,
  }
}

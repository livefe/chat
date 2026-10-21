import { ref, watch } from 'vue'

import { conversationList, type ConversationItem } from '@/api/conversation'
import { CONVERSATION_PAGE_SIZE } from '@/constants/nearby'
import { useChatStore } from '@/stores/chat'
import { sanitizeNickname } from '@/utils/user'

export function sortConversations(items: ConversationItem[]) {
  return [...items].sort((a, b) => {
    const ap = a.pinned ? 1 : 0
    const bp = b.pinned ? 1 : 0
    if (ap !== bp) return bp - ap
    return (b.lastActiveAt || 0) - (a.lastActiveAt || 0)
  })
}

function dedupeByPeer(items: ConversationItem[]) {
  const seen = new Set<number>()
  return items.filter((c) => {
    if (seen.has(c.peerId)) return false
    seen.add(c.peerId)
    return true
  })
}

export function useConversationList(opts: {
  live?: boolean
  onError: (error: unknown, firstPage: boolean) => void
}) {
  const chatStore = useChatStore()
  const live = !!opts.live

  const list = ref<ConversationItem[]>([])
  const loading = ref(false)
  const refreshing = ref(false)
  const finished = ref(false)
  const page = ref(1)
  let total = 0
  let loadSeq = 0

  function order(items: ConversationItem[]) {
    return live ? sortConversations(items) : items
  }

  function normalize(c: ConversationItem): ConversationItem {
    const next = { ...c, peerNickname: sanitizeNickname(c.peerNickname) }
    if (live) next.pinned = !!c.pinned
    return next
  }

  function applyPreview(preview: ConversationItem) {
    const existing = list.value.find((c) => c.peerId === preview.peerId)
    const lastActiveAt = Math.max(preview.lastActiveAt || 0, existing?.lastActiveAt || 0)
    const updated: ConversationItem = {
      peerId: preview.peerId,
      peerNickname:
        sanitizeNickname(preview.peerNickname) || sanitizeNickname(existing?.peerNickname) || '',
      peerAvatar: preview.peerAvatar || existing?.peerAvatar || '',
      lastContent: preview.lastContent || existing?.lastContent || '',
      lastActiveAt,
      unreadCount: existing?.unreadCount ?? 0,
      pinned: existing?.pinned ?? !!preview.pinned,
    }
    list.value = sortConversations([
      updated,
      ...list.value.filter((c) => c.peerId !== preview.peerId),
    ])
  }

  function applyAllPreviews() {
    for (const preview of Object.values(chatStore.conversationPreviews)) {
      applyPreview(preview)
    }
  }

  if (live) {
    watch(() => chatStore.conversationPreviews, applyAllPreviews, { deep: true })
  }

  async function loadMore() {
    const seq = ++loadSeq
    loading.value = true
    try {
      const res = await conversationList(page.value, CONVERSATION_PAGE_SIZE)
      if (seq !== loadSeq) return
      const items = (res.data.list || []).map(normalize)
      total = res.data.total
      list.value = page.value === 1 ? order(items) : order(dedupeByPeer([...list.value, ...items]))
      if (live) {
        chatStore.applyConversationUnread(items)
        applyAllPreviews()
      }
      if (list.value.length >= total || items.length < CONVERSATION_PAGE_SIZE) {
        finished.value = true
      } else {
        page.value += 1
      }
    } catch (e: unknown) {
      if (seq === loadSeq) opts.onError(e, page.value === 1)
    } finally {
      if (seq === loadSeq) {
        loading.value = false
        refreshing.value = false
      }
    }
  }

  function refresh() {
    page.value = 1
    finished.value = false
    return loadMore()
  }

  function reset() {
    list.value = []
    return refresh()
  }

  return { list, loading, refreshing, finished, loadMore, refresh, reset }
}

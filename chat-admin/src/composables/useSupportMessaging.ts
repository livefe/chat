import { nextTick, onUnmounted, ref, watch, type Ref } from 'vue'

import {
  getSupportMessages,
  readSupport,
  sendSupportMessage,
  type SupportAdminSession,
  type SupportMessageItem,
} from '@/api/support'
import { imMessagePreview } from '@/components/im/media'
import { useImConnection } from '@/composables/useImConnection'
import type { SupportSessionList } from '@/composables/useSupportSessions'
import { showApiError } from '@/utils/apiError'
import { usePageVisible } from '@/utils/attention'
import { newClientMsgId, uploadImMessageContent } from '@/utils/imMessage'
import type { ImWSMessage } from '@/utils/imWs'

const MSG_LIMIT = 50
const SOFT_REFRESH_TOAST_COOLDOWN_MS = 30_000
const LIST_REFRESH_DEBOUNCE_MS = 400
export const IM_FALLBACK_POLL_MS = 15_000

type UnreadMode = 'keep' | 'clear' | 'increment'

export function useSupportMessaging(
  active: Ref<SupportAdminSession | null>,
  options: {
    unreadList: SupportSessionList
    previewLists: SupportSessionList[]
    refreshLists: () => Promise<unknown>
    canLoad?: (s: SupportAdminSession) => boolean
    canSend?: (s: SupportAdminSession) => boolean
    canUpload?: () => boolean
    /** 只有当前坐席本人才需要标记已读 */
    canMarkRead?: (s: SupportAdminSession) => boolean
    onPreviewBumped?: () => void
    /** 收到别人发来的消息（不含自己的回显） */
    onIncoming?: (item: SupportMessageItem) => void
    onQueueChanged?: () => void
  },
) {
  const messages = ref<SupportMessageItem[]>([])
  const messagesLoading = ref(false)
  const messagesLoadingMore = ref(false)
  const messagesHasMore = ref(false)
  const threadRef = ref<{ scrollBottom: () => void } | null>(null)
  const draft = ref('')
  const sending = ref(false)
  const uploading = ref(false)
  const pageVisible = usePageVisible()

  let softRefreshFailToastAt = 0
  let pollTimer: number | null = null
  let listRefreshTimer: number | null = null
  let loadSeq = 0

  const { wsState, connect, reconnect, isOpen } = useImConnection({
    onMessage: handleWs,
    onOpen: () => {
      scheduleListRefresh()
      void softRefreshMessages()
    },
  })

  function loadable(s: SupportAdminSession | null): s is SupportAdminSession {
    return !!s && (!options.canLoad || options.canLoad(s))
  }

  function sendable(s: SupportAdminSession | null): s is SupportAdminSession {
    return !!s && (!options.canSend || options.canSend(s))
  }

  function scrollBottom() {
    threadRef.value?.scrollBottom()
  }

  function findSession(list: SupportSessionList, convId: number) {
    return list.list.find((x) => Number(x.id) === convId)
  }

  function bumpPreview(convId: number, item: SupportMessageItem, unread: UnreadMode) {
    let found = false
    const own = findSession(options.unreadList, convId)
    if (own) {
      own.lastContent = imMessagePreview(item)
      own.lastActiveAt = item.createdAt
      if (unread === 'clear') own.unreadCount = 0
      else if (unread === 'increment') own.unreadCount = (own.unreadCount || 0) + 1
      found = true
    }
    if (unread === 'keep') return found
    for (const list of options.previewLists) {
      const cur = findSession(list, convId)
      if (!cur) continue
      cur.lastContent = imMessagePreview(item)
      cur.lastActiveAt = item.createdAt
      found = true
    }
    return found
  }

  /** 页面可见且是本人负责的会话时才上报已读，避免后台挂着页面把消息“读掉”。 */
  async function markActiveRead() {
    const conv = active.value
    if (!conv || !pageVisible.value) return
    if (options.canMarkRead && !options.canMarkRead(conv)) return
    const own = findSession(options.unreadList, conv.id)
    const lastId = messages.value[messages.value.length - 1]?.id
    try {
      await readSupport(conv.id, lastId)
      if (own) own.unreadCount = 0
    } catch {
      return
    }
  }

  async function loadMessages() {
    if (!loadable(active.value)) return
    const convId = active.value.id
    const seq = ++loadSeq
    messagesLoading.value = true
    messagesHasMore.value = false
    try {
      const { data } = await getSupportMessages({ conversationId: convId, limit: MSG_LIMIT })
      if (seq !== loadSeq) return
      messages.value = data.list || []
      messagesHasMore.value = messages.value.length >= MSG_LIMIT
      await nextTick()
      scrollBottom()
      await markActiveRead()
    } catch (e) {
      if (seq === loadSeq) showApiError(e, '加载消息失败')
    } finally {
      if (seq === loadSeq) messagesLoading.value = false
    }
  }

  async function loadMoreMessages() {
    if (!active.value || messagesLoadingMore.value || !messagesHasMore.value) return
    const firstId = messages.value[0]?.id
    if (!firstId) return
    const convId = active.value.id
    messagesLoadingMore.value = true
    try {
      const { data } = await getSupportMessages({
        conversationId: convId,
        beforeId: firstId,
        limit: MSG_LIMIT,
      })
      if (active.value?.id !== convId) return
      const older = data.list || []
      messagesHasMore.value = older.length >= MSG_LIMIT
      const seen = new Set(messages.value.map((m) => m.id))
      const fresh = older.filter((m) => !seen.has(m.id))
      if (fresh.length) messages.value = [...fresh, ...messages.value]
    } catch (e) {
      showApiError(e, '加载更早消息失败')
    } finally {
      messagesLoadingMore.value = false
    }
  }

  /** 拉取最新一页并合并新消息，用于断线兜底与总览旁听轮询。 */
  async function softRefreshMessages() {
    if (!loadable(active.value) || messagesLoading.value) return
    const convId = active.value.id
    try {
      const { data } = await getSupportMessages({ conversationId: convId, limit: MSG_LIMIT })
      if (active.value?.id !== convId) return
      const byId = new Map(messages.value.map((m) => [m.id, m]))
      let added = false
      for (const m of data.list || []) {
        const cur = byId.get(m.id)
        if (!cur) {
          messages.value.push(m)
          added = true
        } else if (m.recalled && !cur.recalled) {
          cur.recalled = true
          cur.content = ''
        }
      }
      if (added) {
        messages.value.sort((a, b) => a.id - b.id)
        await nextTick()
        scrollBottom()
        await markActiveRead()
      }
    } catch (e) {
      const now = Date.now()
      if (now - softRefreshFailToastAt >= SOFT_REFRESH_TOAST_COOLDOWN_MS) {
        softRefreshFailToastAt = now
        showApiError(e, '消息刷新失败')
      }
    }
  }

  function pushMessage(m: SupportMessageItem) {
    if (messages.value.some((x) => x.id === m.id)) return false
    messages.value.push(m)
    void nextTick().then(scrollBottom)
    return true
  }

  function clearMessages() {
    loadSeq++
    messages.value = []
    messagesHasMore.value = false
    messagesLoading.value = false
  }

  function appendSent(m: SupportMessageItem) {
    const convId = Number(m.conversationId || 0)
    if (!convId || convId !== Number(active.value?.id || 0)) {
      void options.unreadList.load({ silent: true })
      return
    }
    if (!pushMessage(m)) return
    bumpPreview(convId, m, 'keep')
  }

  async function sendContent(conversationId: number, contentType: number, content: string) {
    const { data } = await sendSupportMessage({
      conversationId,
      contentType,
      content,
      clientMsgId: newClientMsgId('s'),
    })
    appendSent(data)
  }

  async function onSend() {
    const conv = active.value
    if (!sendable(conv) || sending.value) return
    const text = draft.value.trim()
    if (!text) return
    sending.value = true
    try {
      await sendContent(conv.id, 1, text)
      draft.value = ''
    } catch (e) {
      showApiError(e, '发送失败')
    } finally {
      sending.value = false
    }
  }

  async function onUpload(contentType: number, file: File) {
    const conv = active.value
    if (!sendable(conv) || (options.canUpload && !options.canUpload())) return
    uploading.value = true
    try {
      const content = await uploadImMessageContent(contentType, file)
      await sendContent(conv.id, contentType, content)
    } catch (e) {
      showApiError(e, '上传失败')
    } finally {
      uploading.value = false
    }
  }

  function handleWs(msg: ImWSMessage) {
    if (msg.type === 'support.queue') {
      options.onQueueChanged?.()
      scheduleListRefresh()
      return
    }
    if (msg.type !== 'message' || !msg.data) return

    const item = msg.data as SupportMessageItem
    const convId = Number(item.conversationId || 0)
    if (!convId) return
    if (item.senderKind === 1) options.onIncoming?.(item)

    const cur = active.value
    if (loadable(cur) && Number(cur.id) === convId) {
      pushMessage(item)
      bumpPreview(convId, item, pageVisible.value ? 'clear' : 'increment')
      void markActiveRead()
      return
    }

    if (bumpPreview(convId, item, 'increment')) {
      options.onPreviewBumped?.()
      return
    }
    scheduleListRefresh()
  }

  function scheduleListRefresh() {
    if (listRefreshTimer) window.clearTimeout(listRefreshTimer)
    listRefreshTimer = window.setTimeout(() => {
      listRefreshTimer = null
      void options.refreshLists()
    }, LIST_REFRESH_DEBOUNCE_MS)
  }

  watch(pageVisible, (v) => {
    if (v && messages.value.length) void markActiveRead()
  })

  function start() {
    connect()
    pollTimer = window.setInterval(() => {
      if (isOpen()) return
      void options.refreshLists()
      void softRefreshMessages()
    }, IM_FALLBACK_POLL_MS)
  }

  onUnmounted(() => {
    if (pollTimer) window.clearInterval(pollTimer)
    if (listRefreshTimer) window.clearTimeout(listRefreshTimer)
  })

  return {
    messages,
    messagesLoading,
    messagesLoadingMore,
    messagesHasMore,
    threadRef,
    draft,
    sending,
    uploading,
    wsState,
    reconnect,
    start,
    loadMessages,
    loadMoreMessages,
    softRefreshMessages,
    clearMessages,
    onSend,
    onUpload,
  }
}

import { computed, nextTick, ref, watch } from 'vue'

import {
  getImContacts,
  getImConversations,
  getImMessages,
  openImDM,
  readImConversation,
  type ImContactItem,
  type ImConversationItem,
  type ImMessageItem,
} from '@/api/im'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'
import { showApiError } from '@/utils/apiError'
import { usePageVisible } from '@/utils/attention'

const MSG_LIMIT = 30
const SESSION_PAGE_SIZE = 30
const CONTACT_PAGE_SIZE = 50
const SOFT_REFRESH_TOAST_COOLDOWN_MS = 30_000

type StaffChatTab = 'sessions' | 'contacts'
type LoadOptions = { silent?: boolean; append?: boolean }

export type StaffThreadMessage = ImMessageItem & { sendState?: 'sending' | 'failed' }

export function useStaffChatSessions() {
  const { can } = usePermission()
  const canOpen = computed(() => can(PERM.ImChatOpen))
  const canRead = computed(() => can(PERM.ImChatRead))
  const pageVisible = usePageVisible()

  const sideTab = ref<StaffChatTab>('sessions')
  const sessionKeyword = ref('')
  const sessions = ref<ImConversationItem[]>([])
  const sessionsTotal = ref(0)
  const sessionsLoading = ref(false)
  const sessionsLoadingMore = ref(false)
  const contactKeyword = ref('')
  const contacts = ref<ImContactItem[]>([])
  const contactsTotal = ref(0)
  const contactsLoading = ref(false)
  const contactsLoadingMore = ref(false)
  const active = ref<ImConversationItem | null>(null)
  const messages = ref<StaffThreadMessage[]>([])
  const messagesLoading = ref(false)
  const loadingMore = ref(false)
  const hasMore = ref(false)
  const threadRef = ref<{ scrollBottom: () => void } | null>(null)

  let softRefreshFailToastAt = 0
  let sessionSeq = 0
  let messageSeq = 0

  const unreadTotal = computed(() => sessions.value.reduce((n, s) => n + (s.unreadCount || 0), 0))
  const tabs = computed(() => [
    { key: 'sessions', label: '会话', count: sessionsTotal.value, unread: unreadTotal.value },
    { key: 'contacts', label: '通讯录' },
  ])

  function scrollBottom() {
    threadRef.value?.scrollBottom()
  }

  async function loadContacts(opts?: { append?: boolean }) {
    const append = opts?.append === true
    if (append) contactsLoadingMore.value = true
    else contactsLoading.value = true
    try {
      const page = append ? Math.floor(contacts.value.length / CONTACT_PAGE_SIZE) + 1 : 1
      const { data } = await getImContacts({
        page,
        pageSize: CONTACT_PAGE_SIZE,
        keyword: contactKeyword.value,
      })
      contactsTotal.value = data.total
      if (append) {
        const seen = new Set(contacts.value.map((c) => c.id))
        contacts.value.push(...(data.list || []).filter((c) => !seen.has(c.id)))
      } else {
        contacts.value = data.list || []
      }
    } catch (e) {
      showApiError(e, '加载通讯录失败')
    } finally {
      contactsLoading.value = false
      contactsLoadingMore.value = false
    }
  }

  async function loadSessions(opts?: LoadOptions) {
    const silent = opts?.silent === true
    const append = opts?.append === true
    const seq = ++sessionSeq
    if (append) sessionsLoadingMore.value = true
    else if (!silent) sessionsLoading.value = true
    try {
      const page = append ? Math.floor(sessions.value.length / SESSION_PAGE_SIZE) + 1 : 1
      const { data } = await getImConversations({
        page,
        pageSize: SESSION_PAGE_SIZE,
        keyword: sessionKeyword.value,
      })
      if (seq !== sessionSeq && !append) return
      const fetched = data.list || []
      sessionsTotal.value = data.total
      if (append) {
        const seen = new Set(sessions.value.map((s) => s.id))
        sessions.value.push(...fetched.filter((s) => !seen.has(s.id)))
        return
      }
      const byId = new Map(sessions.value.map((s) => [s.id, s]))
      const head = fetched.map((item) => {
        const cur = byId.get(item.id)
        if (!cur) return item
        Object.assign(cur, item)
        return cur
      })
      const headIds = new Set(head.map((s) => s.id))
      const tail = silent
        ? sessions.value.slice(SESSION_PAGE_SIZE).filter((s) => !headIds.has(s.id))
        : []
      sessions.value = [...head, ...tail]
      const cur = active.value
      if (cur) {
        const fresh = sessions.value.find((s) => s.id === cur.id)
        if (fresh && fresh !== cur) active.value = fresh
      }
    } catch (e) {
      if (!silent) showApiError(e, '加载会话失败')
    } finally {
      if (append) sessionsLoadingMore.value = false
      else if (!silent) sessionsLoading.value = false
    }
  }

  async function startChat(c: ImContactItem) {
    if (!canOpen.value) return
    try {
      const { data } = await openImDM(c.id)
      sideTab.value = 'sessions'
      sessionKeyword.value = ''
      await loadSessions({ silent: true })
      active.value = sessions.value.find((s) => s.id === data.id) || data
      await loadMessages()
    } catch (e) {
      showApiError(e, '打开会话失败')
    }
  }

  async function selectSession(s: ImConversationItem) {
    active.value = s
    await loadMessages()
  }

  async function loadMessages() {
    if (!active.value) return
    const convId = active.value.id
    const seq = ++messageSeq
    messagesLoading.value = true
    messages.value = []
    hasMore.value = false
    try {
      const { data } = await getImMessages({ conversationId: convId, limit: MSG_LIMIT })
      if (seq !== messageSeq) return
      messages.value = data.list || []
      hasMore.value = messages.value.length >= MSG_LIMIT
      await nextTick()
      scrollBottom()
      await markRead()
    } catch (e) {
      if (seq === messageSeq) showApiError(e, '加载消息失败')
    } finally {
      if (seq === messageSeq) messagesLoading.value = false
    }
  }

  async function loadMore() {
    const convId = active.value?.id
    const beforeId = messages.value.find((m) => m.id > 0)?.id
    if (!convId || !beforeId || loadingMore.value) return
    loadingMore.value = true
    try {
      const { data } = await getImMessages({ conversationId: convId, beforeId, limit: MSG_LIMIT })
      if (active.value?.id !== convId) return
      const older = data.list || []
      hasMore.value = older.length >= MSG_LIMIT
      const seen = new Set(messages.value.map((m) => m.id))
      const fresh = older.filter((m) => !seen.has(m.id))
      if (fresh.length) messages.value = [...fresh, ...messages.value]
    } catch (e) {
      showApiError(e, '加载失败')
    } finally {
      loadingMore.value = false
    }
  }

  async function markRead() {
    const conv = active.value
    if (!conv || !canRead.value || !pageVisible.value) return
    const last = [...messages.value].reverse().find((m) => m.id > 0)
    try {
      await readImConversation({ conversationId: conv.id, lastReadMsgId: last?.id })
      conv.unreadCount = 0
    } catch {
      return
    }
  }

  async function softRefreshMessages() {
    if (!active.value || messagesLoading.value) return
    const convId = active.value.id
    try {
      const { data } = await getImMessages({ conversationId: convId, limit: MSG_LIMIT })
      if (active.value?.id !== convId) return
      const byId = new Map(messages.value.map((m) => [m.id, m]))
      let added = false
      for (const m of data.list || []) {
        const cur = byId.get(m.id)
        if (cur) {
          if (m.recalled && !cur.recalled) {
            cur.recalled = true
            cur.content = ''
          }
          continue
        }
        const pendingIdx = m.clientMsgId
          ? messages.value.findIndex((x) => x.id <= 0 && x.clientMsgId === m.clientMsgId)
          : -1
        if (pendingIdx >= 0) messages.value[pendingIdx] = m
        else messages.value.push(m)
        added = true
      }
      if (added) {
        messages.value.sort((a, b) => (a.id > 0 ? a.id : Infinity) - (b.id > 0 ? b.id : Infinity))
        await nextTick()
        scrollBottom()
        await markRead()
      }
    } catch (e) {
      const now = Date.now()
      if (now - softRefreshFailToastAt >= SOFT_REFRESH_TOAST_COOLDOWN_MS) {
        softRefreshFailToastAt = now
        showApiError(e, '消息刷新失败')
      }
    }
  }

  watch(pageVisible, (v) => {
    if (v && active.value && messages.value.length) void markRead()
  })

  return {
    canOpen,
    sideTab,
    tabs,
    unreadTotal,
    sessionKeyword,
    sessions,
    sessionsTotal,
    sessionsLoading,
    sessionsLoadingMore,
    contactKeyword,
    contacts,
    contactsTotal,
    contactsLoading,
    contactsLoadingMore,
    active,
    messages,
    messagesLoading,
    loadingMore,
    hasMore,
    threadRef,
    scrollBottom,
    loadContacts,
    loadSessions,
    startChat,
    selectSession,
    loadMore,
    markRead,
    softRefreshMessages,
  }
}

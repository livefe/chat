import { computed, nextTick, ref } from 'vue'

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

const MSG_LIMIT = 30
const SOFT_REFRESH_TOAST_COOLDOWN_MS = 30_000

type StaffChatTab = 'sessions' | 'contacts'

export function useStaffChatSessions() {
  const { can } = usePermission()
  const canOpen = computed(() => can(PERM.ImChatOpen))
  const canRead = computed(() => can(PERM.ImChatRead))

  const sideTab = ref<StaffChatTab>('sessions')
  const contactKeyword = ref('')
  const contacts = ref<ImContactItem[]>([])
  const contactsLoading = ref(false)
  const sessions = ref<ImConversationItem[]>([])
  const sessionsLoading = ref(false)
  const active = ref<ImConversationItem | null>(null)
  const messages = ref<ImMessageItem[]>([])
  const messagesLoading = ref(false)
  const loadingMore = ref(false)
  const hasMore = ref(false)
  const threadRef = ref<{ scrollBottom: () => void } | null>(null)

  let softRefreshFailToastAt = 0

  const tabs = computed(() => [
    {
      key: 'sessions',
      label: '会话',
      count: sessions.value.reduce((n, s) => n + (s.unreadCount || 0), 0) || undefined,
    },
    { key: 'contacts', label: '通讯录' },
  ])

  function scrollBottom() {
    threadRef.value?.scrollBottom()
  }

  async function loadContacts() {
    contactsLoading.value = true
    try {
      const { data } = await getImContacts({
        page: 1,
        pageSize: 200,
        keyword: contactKeyword.value.trim() || undefined,
      })
      contacts.value = data.list || []
    } catch (e) {
      showApiError(e, '加载通讯录失败')
    } finally {
      contactsLoading.value = false
    }
  }

  async function loadSessions() {
    sessionsLoading.value = true
    try {
      const { data } = await getImConversations({ page: 1, pageSize: 50 })
      sessions.value = data.list || []
    } catch (e) {
      showApiError(e, '加载会话失败')
    } finally {
      sessionsLoading.value = false
    }
  }

  async function startChat(c: ImContactItem) {
    if (!canOpen.value) return
    try {
      const { data } = await openImDM(c.id)
      sideTab.value = 'sessions'
      await loadSessions()
      const found = sessions.value.find((s) => s.id === data.id)
      active.value = found || data
      await loadMessages(true)
    } catch (e) {
      showApiError(e, '打开会话失败')
    }
  }

  async function selectSession(s: ImConversationItem) {
    active.value = s
    await loadMessages(true)
  }

  async function loadMessages(reset: boolean) {
    if (!active.value) return
    if (reset) {
      messagesLoading.value = true
      messages.value = []
      hasMore.value = false
    }
    try {
      const { data } = await getImMessages({
        conversationId: active.value.id,
        limit: MSG_LIMIT,
      })
      messages.value = data.list || []
      hasMore.value = (data.list || []).length >= MSG_LIMIT
      await nextTick()
      scrollBottom()
      await markRead()
    } catch (e) {
      showApiError(e, '加载消息失败')
    } finally {
      messagesLoading.value = false
    }
  }

  async function loadMore() {
    if (!active.value || messages.value.length === 0) return
    loadingMore.value = true
    try {
      const beforeId = messages.value[0]?.id
      const { data } = await getImMessages({
        conversationId: active.value.id,
        beforeId,
        limit: MSG_LIMIT,
      })
      const older = data.list || []
      hasMore.value = older.length >= MSG_LIMIT
      messages.value = [...older, ...messages.value]
    } catch (e) {
      showApiError(e, '加载失败')
    } finally {
      loadingMore.value = false
    }
  }

  async function markRead() {
    if (!active.value || !canRead.value) return
    const last = messages.value[messages.value.length - 1]
    try {
      await readImConversation({
        conversationId: active.value.id,
        lastReadMsgId: last?.id && last.id > 0 ? last.id : undefined,
      })
      const s = sessions.value.find((x) => x.id === active.value?.id)
      if (s) s.unreadCount = 0
    } catch {
      return
    }
  }

  async function softRefreshMessages() {
    if (!active.value) return
    const convId = active.value.id
    try {
      const { data } = await getImMessages({
        conversationId: convId,
        limit: MSG_LIMIT,
      })
      const list = data.list || []
      if (active.value?.id !== convId) return
      const prevIds = new Set(messages.value.map((m) => m.id))
      let added = false
      for (const m of list) {
        if (m.id > 0 && !prevIds.has(m.id)) {
          messages.value.push(m)
          added = true
        }
      }
      if (added) {
        messages.value.sort((a, b) => a.id - b.id)
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

  return {
    canOpen,
    sideTab,
    tabs,
    contactKeyword,
    contacts,
    contactsLoading,
    sessions,
    sessionsLoading,
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

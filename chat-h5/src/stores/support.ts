import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import {
  getSupportCurrent,
  getSupportMessages,
  getSupportProfile,
  openSupportSession,
  readSupportSession,
  sendSupportMessage,
  type SupportMessage,
  type SupportProfile,
  type SupportSession,
} from '@/api/support'
import type { WSMessage } from '@/utils/chat/chatWs'
import { messagePreview } from '@/utils/chat/chatMessageContent'

export const useSupportStore = defineStore('support', () => {
  const profile = ref<SupportProfile | null>(null)
  const session = ref<SupportSession | null>(null)
  const messages = ref<SupportMessage[]>([])
  const viewing = ref(false)
  const previewUnread = ref(0)
  const previewContent = ref('')
  const previewActiveAt = ref(0)

  const brandName = computed(
    () => profile.value?.brandName || session.value?.brandName || '官方客服',
  )
  const brandAvatar = computed(
    () => profile.value?.brandAvatarUrl || session.value?.brandAvatarUrl || '',
  )

  async function loadProfile(silent = false) {
    const res = await getSupportProfile(silent)
    profile.value = res.data
  }

  async function refreshSession(silent = false) {
    const res = await getSupportCurrent(silent)
    session.value = res.data || null
    if (session.value) {
      previewContent.value = session.value.lastContent || previewContent.value
      previewActiveAt.value = session.value.lastActiveAt || previewActiveAt.value
      if (!viewing.value) {
        previewUnread.value = session.value.unreadCount || 0
      }
    }
    return session.value
  }

  async function open() {
    const res = await openSupportSession('h5')
    session.value = res.data
    previewContent.value = res.data.lastContent || ''
    previewActiveAt.value = res.data.lastActiveAt || Date.now() / 1000
    return res.data
  }

  async function loadMessages() {
    if (!session.value) return
    const res = await getSupportMessages({ conversationId: session.value.id, limit: 50 })
    messages.value = res.data.list || []
  }

  async function enter() {
    viewing.value = true
    await refreshSession()
    await loadMessages()
    previewUnread.value = 0
    const last = messages.value[messages.value.length - 1]
    if (session.value) {
      await readSupportSession(session.value.id, last?.id, true).catch(() => {})
    }
  }

  function leave() {
    viewing.value = false
  }

  async function ensureOpenSession() {
    if (!session.value || session.value.status === 2) {
      await open()
      await loadMessages()
    }
    if (!session.value) throw new Error('无法打开客服会话')
    return session.value
  }

  async function sendText(text: string) {
    const cur = await ensureOpenSession()
    const res = await sendSupportMessage({
      conversationId: cur.id,
      contentType: 1,
      content: text,
      clientMsgId: `u-${Date.now()}`,
    })
    appendLocal(res.data)
    return res.data
  }

  async function sendMedia(contentType: number, content: string) {
    const cur = await ensureOpenSession()
    const res = await sendSupportMessage({
      conversationId: cur.id,
      contentType,
      content,
      clientMsgId: `u-${Date.now()}`,
    })
    appendLocal(res.data)
    return res.data
  }

  function previewOf(m: SupportMessage) {
    return messagePreview(m.contentType, m.content)
  }

  function appendLocal(m: SupportMessage) {
    if (messages.value.some((x) => x.id === m.id)) return
    messages.value.push(m)
    previewContent.value = previewOf(m)
    previewActiveAt.value = m.createdAt
    if (session.value) {
      session.value.lastContent = previewContent.value
      session.value.lastActiveAt = m.createdAt
      session.value.lastMessageId = m.id
    }
  }

  function handleWs(env: WSMessage) {
    if (env.type === 'support.session' && env.data) {
      session.value = env.data as SupportSession
      previewContent.value = session.value.lastContent || previewContent.value
      previewActiveAt.value = session.value.lastActiveAt || previewActiveAt.value
      return
    }
    if (env.type === 'message' && env.scope === 'support' && env.data) {
      const m = env.data as SupportMessage
      if (session.value && m.conversationId !== session.value.id) return
      if (viewing.value) {
        appendLocal(m)
        if (session.value) void readSupportSession(session.value.id, m.id)
      } else {
        previewContent.value = previewOf(m)
        previewActiveAt.value = m.createdAt
        previewUnread.value += 1
      }
    }
  }

  return {
    profile,
    session,
    messages,
    viewing,
    previewUnread,
    previewContent,
    previewActiveAt,
    brandName,
    brandAvatar,
    loadProfile,
    refreshSession,
    open,
    enter,
    leave,
    sendText,
    sendMedia,
    appendLocal,
    handleWs,
  }
})

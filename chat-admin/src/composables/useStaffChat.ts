import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'

import { recallImMessage, sendImMessage, type ImMessageItem } from '@/api/im'
import { imMessagePreview } from '@/components/im/media'
import type { ImThreadMessage } from '@/components/im/types'
import { PERM } from '@/constants/permission'
import { useImConnection } from '@/composables/useImConnection'
import { usePermission } from '@/composables/usePermission'
import { IM_FALLBACK_POLL_MS } from '@/composables/useSupportMessaging'
import { useStaffChatSessions, type StaffThreadMessage } from '@/composables/useStaffChatSessions'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { alertAttention, usePageVisible, useTitleCount } from '@/utils/attention'
import { newClientMsgId, uploadImMessageContent } from '@/utils/imMessage'
import type { ImWSMessage } from '@/utils/imWs'

const ACK_TIMEOUT_MS = 10_000

export function useStaffChat() {
  const employeeStore = useEmployeeStore()
  const { can } = usePermission()
  const pageVisible = usePageVisible()
  const selfId = computed(() => employeeStore.employeeInfo?.id || 0)
  const canSend = computed(() => can(PERM.ImChatSend))
  const canRecall = computed(() => can(PERM.ImChatRecall))
  const canUpload = computed(() => can(PERM.OpsUpload))

  const chat = useStaffChatSessions()
  const { sessions, active, messages, scrollBottom, loadContacts, loadSessions, markRead } = chat

  const draft = ref('')
  const sending = ref(false)
  const uploading = ref(false)
  const ackTimers = new Map<string, number>()
  let pollTimer = 0

  const {
    wsState,
    connect,
    reconnect,
    isOpen,
    send: wsSend,
  } = useImConnection({
    onMessage: handleWsMessage,
    onOpen: () => {
      void loadSessions({ silent: true })
      void chat.softRefreshMessages()
    },
  })

  useTitleCount('staff-chat', () => chat.unreadTotal.value)

  function findPending(clientMsgId: string) {
    return messages.value.find((m) => m.clientMsgId === clientMsgId && m.id <= 0)
  }

  function clearAckTimer(clientMsgId: string) {
    const t = ackTimers.get(clientMsgId)
    if (t) window.clearTimeout(t)
    ackTimers.delete(clientMsgId)
  }

  function markFailed(clientMsgId: string) {
    clearAckTimer(clientMsgId)
    const m = findPending(clientMsgId)
    if (m) m.sendState = 'failed'
  }

  function watchAck(clientMsgId: string) {
    clearAckTimer(clientMsgId)
    ackTimers.set(
      clientMsgId,
      window.setTimeout(() => markFailed(clientMsgId), ACK_TIMEOUT_MS),
    )
  }

  function buildPendingText(conversationId: number, clientMsgId: string, text: string) {
    const info = employeeStore.employeeInfo
    const msg: StaffThreadMessage = {
      id: -Date.now(),
      conversationId,
      senderKind: 2,
      senderId: selfId.value,
      senderName: info?.name || info?.mobile || '',
      senderAvatarUrl: '',
      contentType: 1,
      content: text,
      createdAt: Math.floor(Date.now() / 1000),
      recalled: false,
      clientMsgId,
      sendState: 'sending',
    }
    return msg
  }

  /** 优先走 WebSocket（等待 ack，超时标记失败）；未连接时走 HTTP。clientMsgId 保证重发幂等。 */
  async function deliverText(conversationId: number, clientMsgId: string, text: string) {
    const payload = { conversationId, contentType: 1, content: text, clientMsgId }
    if (wsSend({ type: 'send', ...payload })) {
      watchAck(clientMsgId)
      return
    }
    try {
      const { data } = await sendImMessage(payload)
      appendMessage(data)
    } catch (e) {
      markFailed(clientMsgId)
      showApiError(e, '发送失败')
    }
  }

  async function onSend() {
    const conv = active.value
    if (!conv || !canSend.value || sending.value) return
    const text = draft.value.trim()
    if (!text) return
    sending.value = true
    const clientMsgId = newClientMsgId('c')
    messages.value.push(buildPendingText(conv.id, clientMsgId, text))
    draft.value = ''
    void nextTick().then(scrollBottom)
    try {
      await deliverText(conv.id, clientMsgId, text)
    } finally {
      sending.value = false
    }
  }

  async function onResend(m: ImThreadMessage) {
    const conv = active.value
    if (!conv || !m.clientMsgId || m.sendState !== 'failed') return
    m.sendState = 'sending'
    await deliverText(conv.id, m.clientMsgId, m.content)
  }

  async function onUpload(contentType: number, file: File) {
    const conv = active.value
    if (!conv || !canSend.value || !canUpload.value) return
    uploading.value = true
    try {
      const content = await uploadImMessageContent(contentType, file)
      const { data: msg } = await sendImMessage({
        conversationId: conv.id,
        contentType,
        content,
        clientMsgId: newClientMsgId('c'),
      })
      appendMessage(msg)
    } catch (e) {
      showApiError(e, '上传失败')
    } finally {
      uploading.value = false
    }
  }

  async function onRecall(m: ImThreadMessage) {
    if (!canRecall.value || m.id <= 0) return
    try {
      await recallImMessage(m.id)
      m.recalled = true
      m.content = ''
    } catch (e) {
      showApiError(e, '撤回失败')
    }
  }

  /** 更新左侧会话条目；找不到时静默刷新列表（不出现加载遮罩）。 */
  function bumpSession(m: ImMessageItem, incoming: boolean) {
    const s = sessions.value.find((x) => x.id === m.conversationId)
    if (!s) {
      void loadSessions({ silent: true })
      return
    }
    s.lastContent = imMessagePreview(m)
    s.lastActiveAt = m.createdAt
    s.lastMessageId = m.id
    const viewing = active.value?.id === m.conversationId && pageVisible.value
    if (incoming && !viewing) s.unreadCount = (s.unreadCount || 0) + 1
    const idx = sessions.value.indexOf(s)
    if (idx > 0 && !s.pinned) {
      sessions.value.splice(idx, 1)
      const firstUnpinned = sessions.value.findIndex((x) => !x.pinned)
      sessions.value.splice(firstUnpinned < 0 ? sessions.value.length : firstUnpinned, 0, s)
    }
  }

  function appendMessage(m: ImMessageItem) {
    const incoming = m.senderId !== selfId.value
    bumpSession(m, incoming)
    if (!active.value || m.conversationId !== active.value.id) return
    if (m.clientMsgId) {
      clearAckTimer(m.clientMsgId)
      const idx = messages.value.findIndex((x) => x.clientMsgId === m.clientMsgId)
      if (idx >= 0) {
        messages.value[idx] = m
        return
      }
    }
    if (messages.value.some((x) => x.id === m.id)) return
    messages.value.push(m)
    void nextTick().then(scrollBottom)
    if (incoming) void markRead()
  }

  function notifyIncoming(m: ImMessageItem) {
    if (m.senderId === selfId.value) return
    if (pageVisible.value && active.value?.id === m.conversationId) return
    const s = sessions.value.find((x) => x.id === m.conversationId)
    const name = s?.peerName || m.senderName || '同事'
    alertAttention(`${name} 发来新消息`, imMessagePreview(m), `staff-${m.conversationId}`)
  }

  function handleWsMessage(msg: ImWSMessage) {
    if (msg.type === 'message' && msg.data) {
      const item = msg.data as ImMessageItem
      notifyIncoming(item)
      appendMessage(item)
      return
    }
    if (msg.type === 'ack' && msg.clientMsgId && msg.id) {
      clearAckTimer(msg.clientMsgId)
      const cur = findPending(msg.clientMsgId)
      if (cur) {
        cur.id = msg.id
        cur.sendState = undefined
      }
      return
    }
    if (msg.type === 'recall' && msg.id && msg.conversationId) {
      if (active.value?.id === msg.conversationId) {
        const m = messages.value.find((x) => x.id === msg.id)
        if (m) {
          m.recalled = true
          m.content = ''
        }
      }
      return
    }
    if (msg.type === 'error') {
      if (msg.clientMsgId) markFailed(msg.clientMsgId)
      showApiError(new Error(msg.content || '发送失败'), '发送失败')
    }
  }

  onMounted(async () => {
    await Promise.all([loadContacts(), loadSessions()])
    connect()
    pollTimer = window.setInterval(() => {
      if (isOpen()) return
      void loadSessions({ silent: true })
      void chat.softRefreshMessages()
    }, IM_FALLBACK_POLL_MS)
  })

  onUnmounted(() => {
    window.clearInterval(pollTimer)
    ackTimers.forEach((t) => window.clearTimeout(t))
    ackTimers.clear()
  })

  return {
    ...chat,
    selfId,
    canSend,
    canRecall,
    canUpload,
    draft,
    sending,
    uploading,
    wsState,
    reconnect,
    onSend,
    onResend,
    onUpload,
    onRecall,
  }
}

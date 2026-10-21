import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'

import { recallImMessage, sendImMessage, type ImMessageItem } from '@/api/im'
import { imMessagePreview } from '@/components/im/media'
import type { ImThreadMessage } from '@/components/im/types'
import { PERM } from '@/constants/permission'
import { useImConnection } from '@/composables/useImConnection'
import { usePermission } from '@/composables/usePermission'
import { useStaffChatSessions } from '@/composables/useStaffChatSessions'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { newClientMsgId, uploadImMessageContent } from '@/utils/imMessage'
import type { ImWSMessage } from '@/utils/imWs'

export function useStaffChat() {
  const employeeStore = useEmployeeStore()
  const { can } = usePermission()
  const selfId = computed(() => employeeStore.employeeInfo?.id || 0)
  const canSend = computed(() => can(PERM.ImChatSend))
  const canRecall = computed(() => can(PERM.ImChatRecall))
  const canUpload = computed(() => can(PERM.OpsUpload))

  const chat = useStaffChatSessions()
  const { sessions, active, messages, scrollBottom, loadContacts, loadSessions, markRead } = chat

  const draft = ref('')
  const sending = ref(false)
  const uploading = ref(false)
  const pendingSendText = new Map<string, string>()

  const {
    wsState,
    connect,
    reconnect,
    send: wsSend,
  } = useImConnection({
    onMessage: handleWsMessage,
    onOpen: () => {
      void loadSessions()
      if (active.value) void chat.softRefreshMessages()
    },
  })

  function buildOptimisticText(
    conversationId: number,
    clientMsgId: string,
    text: string,
  ): ImMessageItem {
    const info = employeeStore.employeeInfo
    return {
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
    }
  }

  function failOptimistic(clientMsgId: string, fallbackText?: string) {
    const text = pendingSendText.get(clientMsgId) ?? fallbackText ?? ''
    pendingSendText.delete(clientMsgId)
    messages.value = messages.value.filter((m) => m.clientMsgId !== clientMsgId)
    if (text && !draft.value.trim()) {
      draft.value = text
    }
  }

  async function onSend() {
    const conv = active.value
    if (!conv || !canSend.value || sending.value) return
    const text = draft.value.trim()
    if (!text) return
    sending.value = true
    const clientMsgId = newClientMsgId('c')
    messages.value.push(buildOptimisticText(conv.id, clientMsgId, text))
    pendingSendText.set(clientMsgId, text)
    draft.value = ''
    void nextTick().then(scrollBottom)

    try {
      const sentViaWs = wsSend({
        type: 'send',
        conversationId: conv.id,
        contentType: 1,
        content: text,
        clientMsgId,
      })
      if (!sentViaWs) {
        const { data } = await sendImMessage({
          conversationId: conv.id,
          contentType: 1,
          content: text,
          clientMsgId,
        })
        pendingSendText.delete(clientMsgId)
        appendMessage(data)
      }
    } catch (e) {
      failOptimistic(clientMsgId, text)
      showApiError(e, '发送失败')
    } finally {
      sending.value = false
    }
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
    if (!canRecall.value) return
    try {
      await recallImMessage(m.id)
      m.recalled = true
      m.content = ''
    } catch (e) {
      showApiError(e, '撤回失败')
    }
  }

  function afterMessageShown(m: ImMessageItem) {
    void nextTick().then(scrollBottom)
    void markRead()
    const s = sessions.value.find((x) => x.id === m.conversationId)
    if (s) {
      s.lastContent = imMessagePreview(m)
      s.lastActiveAt = m.createdAt
      s.lastMessageId = m.id
    } else {
      void loadSessions()
    }
  }

  function appendMessage(m: ImMessageItem) {
    if (!active.value || m.conversationId !== active.value.id) {
      void loadSessions()
      return
    }
    if (m.clientMsgId) {
      pendingSendText.delete(m.clientMsgId)
      const idx = messages.value.findIndex((x) => x.clientMsgId === m.clientMsgId)
      if (idx >= 0) {
        messages.value[idx] = m
        afterMessageShown(m)
        return
      }
    }
    if (m.id > 0 && messages.value.some((x) => x.id === m.id)) return
    messages.value.push(m)
    afterMessageShown(m)
  }

  function handleWsMessage(msg: ImWSMessage) {
    if (msg.type === 'message' && msg.data) {
      const item = msg.data as ImMessageItem
      appendMessage(item)
      if (
        item.senderId !== selfId.value &&
        (!active.value || item.conversationId !== active.value.id)
      ) {
        const s = sessions.value.find((x) => x.id === item.conversationId)
        if (s) s.unreadCount = (s.unreadCount || 0) + 1
        else void loadSessions()
      }
      return
    }
    if (msg.type === 'ack' && msg.clientMsgId && msg.id) {
      const idx = messages.value.findIndex((x) => x.clientMsgId === msg.clientMsgId)
      const cur = messages.value[idx]
      if (idx >= 0 && cur && cur.id <= 0) {
        messages.value[idx] = { ...cur, id: msg.id }
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
      if (msg.clientMsgId) {
        failOptimistic(msg.clientMsgId)
      }
      showApiError(new Error(msg.content || '发送失败'), '发送失败')
    }
  }

  onMounted(async () => {
    await Promise.all([loadContacts(), loadSessions()])
    connect()
  })

  onUnmounted(() => {
    pendingSendText.clear()
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
    onUpload,
    onRecall,
  }
}

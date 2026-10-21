import type { Ref } from 'vue'

import { privateMessages, type MessageItem } from '@/api/message'
import { deleteConversation as deleteConversationApi } from '@/api/conversation'
import { CHAT_PAGE_LIMIT, type ActiveChatChannel } from '@/constants/chat'
import { mergeMessageLists, normalizeMessages } from '@/utils/chat/chatMessages'

import { historyFromBatch, type HistoryState } from './types'

export interface PrivateChatSessionDeps {
  privateChatMsgsMap: Ref<Map<number, MessageItem[]>>
  privateChatHistory: Ref<Record<number, HistoryState>>
  activeChannel: Ref<ActiveChatChannel>
  currentPeerId: Ref<number>
  setPrivateClearedBefore: (peerId: number, clearedBeforeId: number) => void
  markPeerRead: (peerId: number) => void
  flushPeerRead: (peerId: number) => void
  clearConversationPreview: (peerId: number) => void
  refreshUnreadSummary: () => Promise<unknown> | unknown
}

export function createPrivateChatHelpers(deps: PrivateChatSessionDeps) {
  let enterGeneration = 0

  function getPrivateChatMessages(peerId: number) {
    return deps.privateChatMsgsMap.value.get(peerId) || []
  }

  async function refreshPrivateChatMessages(peerId: number) {
    const res = await privateMessages(peerId, 0, CHAT_PAGE_LIMIT)
    const clearedBefore = res.data.clearedBeforeId
    deps.setPrivateClearedBefore(peerId, clearedBefore)
    const batch = normalizeMessages(res.data.list).filter(
      (m) => !(m.id > 0 && m.id <= clearedBefore),
    )
    const existing = deps.privateChatMsgsMap.value.get(peerId) ?? []
    deps.privateChatMsgsMap.value.set(peerId, mergeMessageLists(existing, batch))
    deps.privateChatMsgsMap.value = new Map(deps.privateChatMsgsMap.value)
    deps.privateChatHistory.value = {
      ...deps.privateChatHistory.value,
      [peerId]: historyFromBatch(batch),
    }
    deps.markPeerRead(peerId)
    return batch
  }

  async function enterPrivateChat(peerId: number) {
    const gen = ++enterGeneration
    const prevChannel = deps.activeChannel.value
    const prevPeerId = deps.currentPeerId.value

    deps.activeChannel.value = 'private'
    deps.currentPeerId.value = peerId

    try {
      await refreshPrivateChatMessages(peerId)
      if (gen !== enterGeneration) return
    } catch (e) {
      if (gen !== enterGeneration) return
      deps.activeChannel.value = prevChannel
      deps.currentPeerId.value = prevPeerId
      void deps.refreshUnreadSummary()
      throw e
    }
  }

  function leavePrivateChat() {
    enterGeneration += 1
    const peerId = deps.currentPeerId.value
    deps.flushPeerRead(peerId)
    deps.currentPeerId.value = 0

    if (deps.activeChannel.value === 'private') {
      deps.activeChannel.value = 'nearby'
    }
  }

  function removeConversation(peerId: number, clearedBeforeId = 0) {
    deps.setPrivateClearedBefore(peerId, clearedBeforeId)
    deps.markPeerRead(peerId)
    deps.clearConversationPreview(peerId)
    if (deps.privateChatMsgsMap.value.has(peerId)) {
      deps.privateChatMsgsMap.value.delete(peerId)
      deps.privateChatMsgsMap.value = new Map(deps.privateChatMsgsMap.value)
    }
    if (deps.privateChatHistory.value[peerId]) {
      const next = { ...deps.privateChatHistory.value }
      delete next[peerId]
      deps.privateChatHistory.value = next
    }
    if (deps.activeChannel.value === 'private' && deps.currentPeerId.value === peerId) {
      leavePrivateChat()
    }
  }

  async function deleteConversation(peerId: number) {
    const res = await deleteConversationApi(peerId)
    removeConversation(peerId, res.data.clearedBeforeId)
  }

  return {
    getPrivateChatMessages,
    refreshPrivateChatMessages,
    enterPrivateChat,
    leavePrivateChat,
    deleteConversation,
  }
}

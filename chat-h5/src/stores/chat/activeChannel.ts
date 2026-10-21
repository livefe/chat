import { computed, type Ref } from 'vue'

import type { MessageItem } from '@/api/message'
import type { ActiveChatChannel } from '@/constants/chat'
import { ensureMessageList } from '@/utils/chat/chatMessages'

import { emptyHistory, type HistoryState } from './types'

export function createActiveChannelHelpers(deps: {
  activeChannel: Ref<ActiveChatChannel>
  currentRoomId: Ref<number>
  currentPeerId: Ref<number>
  nearbyMsgs: Ref<MessageItem[]>
  roomMsgsMap: Ref<Map<number, MessageItem[]>>
  privateChatMsgsMap: Ref<Map<number, MessageItem[]>>
  nearbyHistory: Ref<HistoryState>
  roomHistory: Ref<Record<number, HistoryState>>
  privateChatHistory: Ref<Record<number, HistoryState>>
}) {
  const currentMessages = computed<MessageItem[]>(() => {
    if (deps.activeChannel.value === 'nearby') return deps.nearbyMsgs.value
    if (deps.activeChannel.value === 'room') {
      return deps.roomMsgsMap.value.get(deps.currentRoomId.value) || []
    }
    if (deps.activeChannel.value === 'private') {
      return deps.privateChatMsgsMap.value.get(deps.currentPeerId.value) || []
    }
    return []
  })

  function currentHistoryState(): HistoryState {
    if (deps.activeChannel.value === 'nearby') return deps.nearbyHistory.value
    if (deps.activeChannel.value === 'room') {
      return deps.roomHistory.value[deps.currentRoomId.value] ?? emptyHistory()
    }
    if (deps.activeChannel.value === 'private') {
      return deps.privateChatHistory.value[deps.currentPeerId.value] ?? emptyHistory()
    }
    return emptyHistory()
  }

  function setCurrentHistoryState(state: HistoryState) {
    if (deps.activeChannel.value === 'nearby') {
      deps.nearbyHistory.value = state
      return
    }
    if (deps.activeChannel.value === 'room' && deps.currentRoomId.value > 0) {
      deps.roomHistory.value = {
        ...deps.roomHistory.value,
        [deps.currentRoomId.value]: state,
      }
      return
    }
    if (deps.activeChannel.value === 'private' && deps.currentPeerId.value > 0) {
      deps.privateChatHistory.value = {
        ...deps.privateChatHistory.value,
        [deps.currentPeerId.value]: state,
      }
    }
  }

  function currentMessageList(): MessageItem[] | null {
    if (deps.activeChannel.value === 'nearby') return deps.nearbyMsgs.value
    if (deps.activeChannel.value === 'room') {
      const rid = deps.currentRoomId.value
      if (!rid) return null
      return ensureMessageList(deps.roomMsgsMap.value, rid)
    }
    if (deps.activeChannel.value === 'private') {
      const peer = deps.currentPeerId.value
      if (!peer) return null
      return ensureMessageList(deps.privateChatMsgsMap.value, peer)
    }
    return null
  }

  const historyLoading = computed(() => currentHistoryState().loading)
  const historyHasMore = computed(() => currentHistoryState().hasMore)

  return {
    currentMessages,
    currentHistoryState,
    setCurrentHistoryState,
    currentMessageList,
    historyLoading,
    historyHasMore,
  }
}

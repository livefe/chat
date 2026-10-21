import type { ChatState } from './state'
import { emptyHistory } from './types'

export function clearChatSessionState(state: ChatState) {
  state.activeChannel.value = 'nearby'
  state.currentRoomId.value = 0
  state.currentPeerId.value = 0
  state.nearbyMsgs.value = []
  state.roomMsgsMap.value = new Map()
  state.postMsgsMap.value = new Map()
  state.privateChatMsgsMap.value = new Map()
  state.rooms.value = []
  state.roomMetaById.value = {}
  state.roomOnlineById.value = {}
  state.postOnlineById.value = {}
  state.nearbyOnlineCount.value = 0
  state.nearbyHistory.value = emptyHistory()
  state.roomHistory.value = {}
  state.postHistory.value = {}
  state.privateChatHistory.value = {}
  state.conversationPreviews.value = {}
  state.messagesLoading.value = false
  state.subscribedRooms.value = new Set()
  state.postSubReasons.value = new Map()
}

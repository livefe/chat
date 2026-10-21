import type { Ref } from 'vue'

import { nearbyBootstrap, type NearbyRoomItem } from '@/api/room'
import type { MessageItem } from '@/api/message'
import { CHAT_PAGE_LIMIT } from '@/constants/chat'
import { NEARBY_RADIUS_METERS } from '@/constants/nearby'
import { mergeMessageLists, normalizeMessages } from '@/utils/chat/chatMessages'

import { emptyHistory, historyFromBatch, type ChatLocationSource, type HistoryState } from './types'

export interface NearbyHelpersDeps {
  rooms: Ref<NearbyRoomItem[]>
  nearbyMsgs: Ref<MessageItem[]>
  nearbyHistory: Ref<HistoryState>
  nearbyOnlineCount: Ref<number>
  messagesLoading: Ref<boolean>
  getLocation: ChatLocationSource
  reportLocation: () => void
}

export function createNearbyHelpers(deps: NearbyHelpersDeps) {
  let generation = 0

  async function refreshNearby(opts?: {
    includeMessages?: boolean
    replaceMessages?: boolean
    syncServerLocation?: boolean
  }) {
    const loc = deps.getLocation()
    if (!loc) return

    const includeMessages = opts?.includeMessages !== false
    const replaceMessages = opts?.replaceMessages === true

    if (opts?.syncServerLocation !== false) {
      deps.reportLocation()
    }

    const gen = ++generation
    const { lat, lng } = loc
    if (includeMessages) deps.messagesLoading.value = true
    try {
      const res = await nearbyBootstrap(
        lat,
        lng,
        CHAT_PAGE_LIMIT,
        includeMessages,
        NEARBY_RADIUS_METERS,
      )
      if (gen !== generation) return

      deps.rooms.value = res.data.rooms
      deps.nearbyOnlineCount.value = res.data.onlineCount

      if (includeMessages) {
        const batch = normalizeMessages(res.data.messages ?? [])
        deps.nearbyMsgs.value = replaceMessages
          ? batch
          : mergeMessageLists(deps.nearbyMsgs.value, batch)
        deps.nearbyHistory.value = historyFromBatch(batch)
      }
    } finally {
      if (includeMessages && gen === generation) deps.messagesLoading.value = false
    }
  }

  function clearNearbyMessages() {
    deps.nearbyMsgs.value = []
    deps.nearbyHistory.value = emptyHistory()
  }

  return {
    refreshNearby,
    clearNearbyMessages,
  }
}

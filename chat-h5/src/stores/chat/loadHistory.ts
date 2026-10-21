import type { Ref } from 'vue'

import { nearbyMessages, privateMessages, roomMessages, type MessageItem } from '@/api/message'
import { CHAT_PAGE_LIMIT, type ActiveChatChannel } from '@/constants/chat'
import { NEARBY_RADIUS_METERS } from '@/constants/nearby'
import { normalizeMessages } from '@/utils/chat/chatMessages'

import { loadOlderMessages } from './history'
import type { ChatLocationSource, HistoryState } from './types'

export async function loadMoreActiveHistory(opts: {
  activeChannel: Ref<ActiveChatChannel>
  currentRoomId: Ref<number>
  currentPeerId: Ref<number>
  getLocation: ChatLocationSource
  getState: () => HistoryState
  setState: (state: HistoryState) => void
  list: MessageItem[] | null
  setPrivateClearedBefore: (peerId: number, clearedBeforeId: number) => void
}): Promise<boolean> {
  return loadOlderMessages({
    getState: opts.getState,
    setState: opts.setState,
    list: opts.list,
    fetch: async (beforeId) => {
      if (opts.activeChannel.value === 'nearby') {
        const loc = opts.getLocation()
        if (!loc) return null
        const res = await nearbyMessages(
          loc.lat,
          loc.lng,
          beforeId,
          CHAT_PAGE_LIMIT,
          NEARBY_RADIUS_METERS,
        )
        return normalizeMessages(res.data.list)
      }
      if (opts.activeChannel.value === 'room') {
        const loc = opts.getLocation()
        const res = await roomMessages(
          opts.currentRoomId.value,
          beforeId,
          CHAT_PAGE_LIMIT,
          loc?.lat,
          loc?.lng,
        )
        return normalizeMessages(res.data.list)
      }
      const peerId = opts.currentPeerId.value
      const res = await privateMessages(peerId, beforeId, CHAT_PAGE_LIMIT)
      opts.setPrivateClearedBefore(peerId, res.data.clearedBeforeId)
      return normalizeMessages(res.data.list)
    },
  })
}

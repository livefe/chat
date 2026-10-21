import { ref } from 'vue'

import type { NearbyRoomItem, RoomDetail } from '@/api/room'
import type { MessageItem } from '@/api/message'
import type { ConversationItem } from '@/api/conversation'
import type { ActiveChatChannel } from '@/constants/chat'

import { emptyHistory, type HistoryState, type PostSubReason } from './types'

export function createChatState() {
  return {
    activeChannel: ref<ActiveChatChannel>('nearby'),
    currentRoomId: ref<number>(0),
    currentPeerId: ref<number>(0),

    nearbyMsgs: ref<MessageItem[]>([]),
    roomMsgsMap: ref<Map<number, MessageItem[]>>(new Map()),
    postMsgsMap: ref<Map<number, MessageItem[]>>(new Map()),
    privateChatMsgsMap: ref<Map<number, MessageItem[]>>(new Map()),

    rooms: ref<NearbyRoomItem[]>([]),
    roomMetaById: ref<Record<number, Pick<RoomDetail, 'id' | 'name' | 'requireGeofence'>>>({}),
    roomOnlineById: ref<Record<number, number>>({}),
    postOnlineById: ref<Record<number, number>>({}),
    nearbyOnlineCount: ref(0),

    nearbyHistory: ref<HistoryState>(emptyHistory()),
    roomHistory: ref<Record<number, HistoryState>>({}),
    postHistory: ref<Record<number, HistoryState>>({}),
    privateChatHistory: ref<Record<number, HistoryState>>({}),
    conversationPreviews: ref<Record<number, ConversationItem>>({}),

    connected: ref(false),
    messagesLoading: ref(false),
    subscribedRooms: ref(new Set<number>()),
    postSubReasons: ref(new Map<number, Set<PostSubReason>>()),
  }
}

export type ChatState = ReturnType<typeof createChatState>

import type { Ref } from 'vue'

import { roomBootstrap, roomDetail, type NearbyRoomItem, type RoomDetail } from '@/api/room'
import type { MessageItem } from '@/api/message'
import { CHAT_PAGE_LIMIT, type ActiveChatChannel } from '@/constants/chat'
import { normalizeMessages } from '@/utils/chat/chatMessages'

import type { ChatLocationSource, PostSubReason } from './types'

export interface RoomSessionDeps {
  activeChannel: Ref<ActiveChatChannel>
  currentRoomId: Ref<number>
  rooms: Ref<NearbyRoomItem[]>
  roomMetaById: Ref<Record<number, Pick<RoomDetail, 'id' | 'name' | 'requireGeofence'>>>
  messagesLoading: Ref<boolean>
  connect: () => void
  getLocation: ChatLocationSource
  waitRoomSubscribe: (roomId: number) => Promise<void>
  rejectRoomSubscribe: (roomId: number, error: Error) => void
  acquireRoomSubscription: (roomId: number) => boolean
  releaseRoomSubscription: (roomId: number) => void
  sendRoomSubscribe: (roomId: number) => void
  applyRoomMessages: (roomId: number, batch: MessageItem[]) => void
  setRoomOnlineCount: (roomId: number, onlineCount: number) => void
  waitPostSubscribe: (postId: number) => Promise<void>
  postSubscribeMatches: (postId: number) => boolean
  rejectPostSubscribe: (postId: number, error: Error) => void
  acquirePostSubscription: (postId: number, reason: PostSubReason) => boolean
  releasePostSubscription: (postId: number, reason: PostSubReason) => void
  sendPostSubscribe: (postId: number) => void
}

export function createRoomSessionHelpers(deps: RoomSessionDeps) {
  let enterGeneration = 0
  let enteringRoomId = 0

  function roomName(roomId: number): string {
    if (roomId <= 0) return ''
    const nearby = deps.rooms.value.find((r) => r.id === roomId)
    if (nearby?.name) return nearby.name
    return deps.roomMetaById.value[roomId]?.name ?? ''
  }

  function rememberRoomMeta(detail: Pick<RoomDetail, 'id' | 'name' | 'requireGeofence'>) {
    deps.roomMetaById.value = {
      ...deps.roomMetaById.value,
      [detail.id]: {
        id: detail.id,
        name: detail.name,
        requireGeofence: detail.requireGeofence,
      },
    }
  }

  async function enterRoom(roomId: number): Promise<RoomDetail | null> {
    if (deps.activeChannel.value === 'room' && deps.currentRoomId.value === roomId) {
      const res = await roomDetail(roomId)
      rememberRoomMeta(res.data)
      return res.data
    }

    const gen = ++enterGeneration
    enteringRoomId = roomId
    const prevChannel = deps.activeChannel.value
    const prevRoomId = deps.currentRoomId.value
    let newlySubscribed = false

    if (deps.currentRoomId.value > 0 && deps.currentRoomId.value !== roomId) {
      deps.releaseRoomSubscription(deps.currentRoomId.value)
    }

    try {
      deps.activeChannel.value = 'room'
      deps.currentRoomId.value = roomId
      deps.messagesLoading.value = true
      deps.connect()

      const loc = deps.getLocation()
      const waitPromise = deps.waitRoomSubscribe(roomId)
      newlySubscribed = deps.acquireRoomSubscription(roomId)
      if (!newlySubscribed) {
        deps.sendRoomSubscribe(roomId)
      }

      const bootPromise = roomBootstrap(roomId, CHAT_PAGE_LIMIT, true, loc?.lat, loc?.lng)

      const [bootRes] = await Promise.all([bootPromise, waitPromise])
      if (gen !== enterGeneration) {
        if (newlySubscribed) deps.releaseRoomSubscription(roomId)
        return null
      }

      rememberRoomMeta(bootRes.data.room)
      deps.setRoomOnlineCount(roomId, bootRes.data.onlineCount)
      deps.applyRoomMessages(roomId, normalizeMessages(bootRes.data.messages ?? []))
      return bootRes.data.room
    } catch (e) {
      if (gen !== enterGeneration) return null
      if (newlySubscribed) {
        deps.releaseRoomSubscription(roomId)
      }
      deps.activeChannel.value = prevChannel
      deps.currentRoomId.value = prevRoomId
      if (prevChannel === 'room' && prevRoomId > 0) {
        deps.acquireRoomSubscription(prevRoomId)
      }
      throw e
    } finally {
      if (gen === enterGeneration) {
        deps.messagesLoading.value = false
        enteringRoomId = 0
      }
    }
  }

  function leaveRoom() {
    enterGeneration += 1
    const leavingId = deps.currentRoomId.value || enteringRoomId
    enteringRoomId = 0
    if (leavingId > 0) {
      deps.rejectRoomSubscribe(leavingId, new Error('已离开聊天区'))
      deps.releaseRoomSubscription(leavingId)
    }
    deps.currentRoomId.value = 0
    if (deps.activeChannel.value === 'room') {
      deps.activeChannel.value = 'nearby'
    }
  }

  async function subscribePostFeed(postId: number) {
    if (postId <= 0) return
    deps.connect()
    const waitPromise = deps.waitPostSubscribe(postId)
    const newlySubscribed = deps.acquirePostSubscription(postId, 'feed')
    if (!newlySubscribed) {
      deps.sendPostSubscribe(postId)
    }
    try {
      await waitPromise
    } catch (e) {
      if (newlySubscribed) {
        deps.releasePostSubscription(postId, 'feed')
      }
      throw e
    }
  }

  function unsubscribePostFeed(postId: number) {
    if (postId <= 0) return
    if (deps.postSubscribeMatches(postId)) {
      deps.rejectPostSubscribe(postId, new Error('已取消订阅'))
    }
    deps.releasePostSubscription(postId, 'feed')
  }

  return {
    roomName,
    enterRoom,
    leaveRoom,
    subscribePostFeed,
    unsubscribePostFeed,
  }
}

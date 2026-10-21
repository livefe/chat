import type { Ref } from 'vue'

import type { WSClient, WSMessage } from '@/utils/chat/chatWs'

import type { PostSubReason } from './types'

export interface SubscriptionDeps {
  subscribedRooms: Ref<Set<number>>
  postSubReasons: Ref<Map<number, Set<PostSubReason>>>
  getClient: () => WSClient | null
  getActiveLocation: () => { lat: number; lng: number } | null
}

export function createSubscriptionHelpers(deps: SubscriptionDeps) {
  function sendRoomSubscribe(roomId: number) {
    const payload: WSMessage = { type: 'subscribe', scope: 'room', roomId }
    const loc = deps.getActiveLocation()
    if (loc) {
      payload.lat = loc.lat
      payload.lng = loc.lng
    }
    deps.getClient()?.send(payload)
  }

  function sendRoomUnsubscribe(roomId: number) {
    deps.getClient()?.send({ type: 'unsubscribe', scope: 'room', roomId })
  }

  function sendPostSubscribe(postId: number) {
    deps.getClient()?.send({ type: 'subscribe', scope: 'post', postId })
  }

  function sendPostUnsubscribe(postId: number) {
    deps.getClient()?.send({ type: 'unsubscribe', scope: 'post', postId })
  }

  function acquireRoomSubscription(roomId: number): boolean {
    if (roomId <= 0) return false
    const wasActive = deps.subscribedRooms.value.has(roomId)
    deps.subscribedRooms.value = new Set(deps.subscribedRooms.value).add(roomId)
    if (!wasActive) {
      sendRoomSubscribe(roomId)
      return true
    }
    return false
  }

  function releaseRoomSubscription(roomId: number) {
    if (roomId <= 0 || !deps.subscribedRooms.value.has(roomId)) return
    const next = new Set(deps.subscribedRooms.value)
    next.delete(roomId)
    deps.subscribedRooms.value = next
    sendRoomUnsubscribe(roomId)
  }

  function acquirePostSubscription(postId: number, reason: PostSubReason): boolean {
    if (postId <= 0) return false
    const next = new Map(deps.postSubReasons.value)
    const reasons = new Set(next.get(postId) ?? [])
    const wasActive = reasons.size > 0
    reasons.add(reason)
    next.set(postId, reasons)
    deps.postSubReasons.value = next
    if (!wasActive) {
      sendPostSubscribe(postId)
      return true
    }
    return false
  }

  function releasePostSubscription(postId: number, reason: PostSubReason) {
    if (postId <= 0) return
    const reasons = deps.postSubReasons.value.get(postId)
    if (!reasons?.has(reason)) return
    const next = new Map(deps.postSubReasons.value)
    const updated = new Set(reasons)
    updated.delete(reason)
    if (updated.size === 0) {
      next.delete(postId)
      sendPostUnsubscribe(postId)
    } else {
      next.set(postId, updated)
    }
    deps.postSubReasons.value = next
  }

  function resyncRoomSubscriptions() {
    for (const roomId of deps.subscribedRooms.value) {
      sendRoomSubscribe(roomId)
    }
  }

  function resyncPostSubscriptions() {
    for (const postId of deps.postSubReasons.value.keys()) {
      sendPostSubscribe(postId)
    }
  }

  function feedPostIds() {
    const ids: number[] = []
    for (const [postId, set] of deps.postSubReasons.value) {
      if (set.has('feed')) ids.push(postId)
    }
    return ids
  }

  return {
    sendRoomSubscribe,
    sendPostSubscribe,
    acquireRoomSubscription,
    releaseRoomSubscription,
    acquirePostSubscription,
    releasePostSubscription,
    resyncRoomSubscriptions,
    resyncPostSubscriptions,
    feedPostIds,
  }
}

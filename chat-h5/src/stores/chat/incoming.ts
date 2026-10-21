import type { Ref } from 'vue'

import type { MessageItem } from '@/api/message'
import { MessageScope, type ActiveChatChannel } from '@/constants/chat'
import {
  applyRecallToList,
  ensureMessageList,
  lastVisiblePreview,
  replaceOptimisticOrAppend,
} from '@/utils/chat/chatMessages'
import { messagePreview } from '@/utils/chat/chatMessageContent'
import type { WSMessage } from '@/utils/chat/chatWs'

import type { ChatNotice } from './types'

export interface IncomingDeps {
  nearbyMsgs: Ref<MessageItem[]>
  roomMsgsMap: Ref<Map<number, MessageItem[]>>
  postMsgsMap: Ref<Map<number, MessageItem[]>>
  privateChatMsgsMap: Ref<Map<number, MessageItem[]>>
  activeChannel: Ref<ActiveChatChannel>
  currentPeerId: Ref<number>
  getSelfId: () => number | undefined
  clearPendingSend: (clientMsgId: string) => void
  applyAckEverywhere: (clientMsgId: string, id: number) => void
  failOptimisticEverywhere: (clientMsgId: string, reason?: string) => void
  roomSubscribeMatches: (roomId: number) => boolean
  rejectRoomSubscribe: (roomId: number, err: Error) => void
  resolveRoomSubscribe: (roomId: number) => void
  postSubscribeMatches: (postId: number) => boolean
  rejectPostSubscribe: (postId: number, err: Error) => void
  resolvePostSubscribe: (postId: number) => void
  setRoomOnlineCount: (roomId: number, onlineCount: number) => void
  setPostOnlineCount: (postId: number, onlineCount: number) => void
  isClearedMessage: (peerId: number, messageId: number) => boolean
  upsertConversationPreview: (opts: {
    peerId: number
    lastContent: string
    lastActiveAt: number
    peerNickname?: string
    peerAvatar?: string
  }) => void
  acknowledgePeerRead: (peerId: number) => void
  incrPeerUnread: (peerId: number) => void
  notify: (notice: ChatNotice) => void
  forwardSupportEvent: (env: WSMessage) => void
}

export function createIncomingHandler(deps: IncomingDeps) {
  function handleIncoming(env: WSMessage) {
    if (env.type === 'ack') {
      const clientMsgId = env.clientMsgId
      const id = env.id
      if (clientMsgId) {
        deps.clearPendingSend(clientMsgId)
        if (id && id > 0) {
          deps.applyAckEverywhere(clientMsgId, id)
        }
      }
      return
    }
    if (env.type === 'error') {
      const text = env.content || '操作失败'
      const roomId = env.roomId ?? 0
      if (roomId > 0 && deps.roomSubscribeMatches(roomId)) {
        deps.rejectRoomSubscribe(roomId, new Error(text))
      }
      const postId = env.postId ?? 0
      if (postId > 0 && deps.postSubscribeMatches(postId)) {
        deps.rejectPostSubscribe(postId, new Error(text))
      }
      const clientMsgId = env.clientMsgId
      if (clientMsgId) {
        deps.clearPendingSend(clientMsgId)
        deps.failOptimisticEverywhere(clientMsgId, text)
      }
      deps.notify({ type: 'serverError', message: text })
      return
    }
    if (env.type === 'roomPresence') {
      const roomId = env.roomId ?? 0
      if (roomId > 0 && env.onlineCount !== undefined) {
        deps.setRoomOnlineCount(roomId, env.onlineCount)
      }
      if (roomId > 0 && deps.roomSubscribeMatches(roomId)) {
        deps.resolveRoomSubscribe(roomId)
      }
      return
    }
    if (env.type === 'postPresence') {
      const postId = env.postId ?? 0
      if (postId > 0 && env.onlineCount !== undefined) {
        deps.setPostOnlineCount(postId, env.onlineCount)
      }
      if (postId > 0 && deps.postSubscribeMatches(postId)) {
        deps.resolvePostSubscribe(postId)
      }
      return
    }
    if (env.type === 'support.session') {
      deps.forwardSupportEvent(env)
      return
    }
    if (env.type === 'message' && env.data) {
      if (env.scope === 'support') {
        deps.forwardSupportEvent(env)
        return
      }
      const payload = env.data as MessageItem
      const item: MessageItem = {
        id: payload.id,
        scope: payload.scope,
        roomId: payload.roomId,
        postId: payload.postId,
        fromUserId: payload.fromUserId,
        toUserId: payload.toUserId,
        fromName: payload.fromName,
        fromAvatar: payload.fromAvatar,
        contentType: payload.contentType,
        content: payload.content,
        replyToId: payload.replyToId,
        replyToFromName: payload.replyToFromName,
        replyToContentType: payload.replyToContentType,
        replyToContent: payload.replyToContent,
        latitude: payload.latitude,
        longitude: payload.longitude,
        createdAt: payload.createdAt,
        recalled: payload.recalled,
      }
      const selfId = deps.getSelfId()
      const clientMsgId = env.clientMsgId
      if (clientMsgId) deps.clearPendingSend(clientMsgId)
      if (env.scope === 'nearby') {
        replaceOptimisticOrAppend(deps.nearbyMsgs.value, item, selfId, clientMsgId)
      } else if (env.scope === 'room') {
        const list = ensureMessageList(deps.roomMsgsMap.value, item.roomId)
        replaceOptimisticOrAppend(list, item, selfId, clientMsgId)
      } else if (env.scope === 'post') {
        const pid = item.postId
        if (!pid) return
        const list = ensureMessageList(deps.postMsgsMap.value, pid)
        replaceOptimisticOrAppend(list, item, selfId, clientMsgId)
      } else if (env.scope === 'private') {
        const peer = item.fromUserId === selfId ? item.toUserId : item.fromUserId
        if (deps.isClearedMessage(peer, item.id)) return
        const list = ensureMessageList(deps.privateChatMsgsMap.value, peer)
        replaceOptimisticOrAppend(list, item, selfId, clientMsgId)

        const fromOther = !!(selfId && item.fromUserId !== selfId)
        deps.upsertConversationPreview({
          peerId: peer,
          lastContent: messagePreview(item.contentType, item.content),
          lastActiveAt: item.createdAt,
          peerNickname: fromOther ? item.fromName : undefined,
          peerAvatar: fromOther ? item.fromAvatar : undefined,
        })
        if (fromOther) {
          const viewingPeer =
            deps.activeChannel.value === 'private' && deps.currentPeerId.value === peer
          if (viewingPeer) {
            deps.acknowledgePeerRead(peer)
          } else {
            deps.incrPeerUnread(peer)
          }
        }
      }
      return
    }
    if (env.type === 'recall' && env.data) {
      const payload = env.data as MessageItem
      applyRecallEverywhere(payload.id, payload)
    }
  }

  function applyRecallEverywhere(id: number, item?: MessageItem) {
    if (id <= 0) return
    applyRecallToList(deps.nearbyMsgs.value, id)
    for (const list of deps.roomMsgsMap.value.values()) applyRecallToList(list, id)
    for (const list of deps.postMsgsMap.value.values()) applyRecallToList(list, id)

    const selfId = deps.getSelfId()
    let patchedPrivate = false
    for (const [peer, list] of deps.privateChatMsgsMap.value) {
      if (!list.some((m) => m.id === id)) continue
      applyRecallToList(list, id)
      patchedPrivate = true
      deps.upsertConversationPreview({
        peerId: peer,
        lastContent: lastVisiblePreview(list),
        lastActiveAt: 0,
      })
    }
    if (patchedPrivate || !item) return
    const isPrivate = item.scope === MessageScope.Private
    if (!isPrivate) return
    const peer = item.fromUserId === selfId ? item.toUserId : item.fromUserId
    if (peer <= 0 || deps.isClearedMessage(peer, id)) return
    deps.upsertConversationPreview({
      peerId: peer,
      lastContent: '撤回了一条消息',
      lastActiveAt: 0,
    })
  }

  return { handleIncoming, applyRecallEverywhere }
}

import type { Ref } from 'vue'

import type { MessageItem, MessageReplyQuote } from '@/api/message'
import {
  appendMessage,
  applyAckToList,
  ensureMessageList,
  markAllOptimisticFailed,
  markOptimisticFailed,
} from '@/utils/chat/chatMessages'
import { messageScopeToChannel } from '@/constants/chat'
import { messagePreview } from '@/utils/chat/chatMessageContent'

type OptimisticLoc =
  | { kind: 'nearby' }
  | { kind: 'room'; roomId: number }
  | { kind: 'post'; postId: number }
  | { kind: 'private'; peerId: number }

export interface OptimisticMessageDeps {
  nearbyMsgs: Ref<MessageItem[]>
  roomMsgsMap: Ref<Map<number, MessageItem[]>>
  postMsgsMap: Ref<Map<number, MessageItem[]>>
  privateChatMsgsMap: Ref<Map<number, MessageItem[]>>
  currentPeerId: Ref<number>
  getSelfProfile: () => { id?: number; nickname?: string; avatarUrl?: string } | null | undefined
  upsertConversationPreview: (opts: {
    peerId: number
    lastContent: string
    lastActiveAt: number
    peerNickname?: string
    peerAvatar?: string
  }) => void
}

export function createOptimisticHelpers(deps: OptimisticMessageDeps) {
  const byClientMsgId = new Map<string, OptimisticLoc>()

  function remember(clientMsgId: string | undefined, loc: OptimisticLoc) {
    if (clientMsgId) byClientMsgId.set(clientMsgId, loc)
  }

  function listFor(loc: OptimisticLoc): MessageItem[] | undefined {
    switch (loc.kind) {
      case 'nearby':
        return deps.nearbyMsgs.value
      case 'room':
        return deps.roomMsgsMap.value.get(loc.roomId)
      case 'post':
        return deps.postMsgsMap.value.get(loc.postId)
      case 'private':
        return deps.privateChatMsgsMap.value.get(loc.peerId)
    }
  }

  function applyAckEverywhere(clientMsgId: string, id: number) {
    const loc = byClientMsgId.get(clientMsgId)
    if (loc) {
      byClientMsgId.delete(clientMsgId)
      const list = listFor(loc)
      if (list && applyAckToList(list, clientMsgId, id)) return
    }
    if (applyAckToList(deps.nearbyMsgs.value, clientMsgId, id)) return
    for (const [, list] of deps.roomMsgsMap.value) {
      if (applyAckToList(list, clientMsgId, id)) return
    }
    for (const [, list] of deps.postMsgsMap.value) {
      if (applyAckToList(list, clientMsgId, id)) return
    }
    for (const [, list] of deps.privateChatMsgsMap.value) {
      if (applyAckToList(list, clientMsgId, id)) return
    }
  }

  function failOptimisticEverywhere(clientMsgId: string, reason?: string) {
    const loc = byClientMsgId.get(clientMsgId)
    if (loc) {
      byClientMsgId.delete(clientMsgId)
      const list = listFor(loc)
      if (list) {
        markOptimisticFailed(list, clientMsgId, reason)
        return
      }
    }
    markOptimisticFailed(deps.nearbyMsgs.value, clientMsgId, reason)
    for (const [, list] of deps.roomMsgsMap.value) markOptimisticFailed(list, clientMsgId, reason)
    for (const [, list] of deps.postMsgsMap.value) markOptimisticFailed(list, clientMsgId, reason)
    for (const [, list] of deps.privateChatMsgsMap.value)
      markOptimisticFailed(list, clientMsgId, reason)
  }

  function failAllOptimisticEverywhere() {
    byClientMsgId.clear()
    markAllOptimisticFailed(deps.nearbyMsgs.value)
    for (const [, list] of deps.roomMsgsMap.value) markAllOptimisticFailed(list)
    for (const [, list] of deps.postMsgsMap.value) markAllOptimisticFailed(list)
    for (const [, list] of deps.privateChatMsgsMap.value) markAllOptimisticFailed(list)
  }

  function insertOptimistic(item: MessageItem) {
    const channel = messageScopeToChannel(item.scope)
    if (channel === 'nearby') {
      appendMessage(deps.nearbyMsgs.value, item)
      remember(item.clientMsgId, { kind: 'nearby' })
      return
    }
    if (channel === 'room') {
      const rid = item.roomId
      if (!rid) return
      appendMessage(ensureMessageList(deps.roomMsgsMap.value, rid), item)
      remember(item.clientMsgId, { kind: 'room', roomId: rid })
      return
    }
    if (channel === 'private') {
      const peer = item.toUserId || deps.currentPeerId.value
      if (!peer) return
      appendMessage(ensureMessageList(deps.privateChatMsgsMap.value, peer), item)
      remember(item.clientMsgId, { kind: 'private', peerId: peer })
      deps.upsertConversationPreview({
        peerId: peer,
        lastContent: messagePreview(item.contentType, item.content),
        lastActiveAt: item.createdAt,
      })
    }
  }

  function insertOptimisticPost(postId: number, item: MessageItem) {
    appendMessage(ensureMessageList(deps.postMsgsMap.value, postId), item)
    remember(item.clientMsgId, { kind: 'post', postId })
  }

  function buildOptimistic(
    content: string,
    clientMsgId: string,
    extra: Pick<MessageItem, 'scope' | 'roomId' | 'postId' | 'toUserId'>,
    contentType = 1,
    reply?: MessageReplyQuote | null,
  ): MessageItem {
    const u = deps.getSelfProfile()
    return {
      id: 0,
      scope: extra.scope,
      roomId: extra.roomId,
      postId: extra.postId,
      fromUserId: u?.id ?? 0,
      toUserId: extra.toUserId,
      fromName: u?.nickname || '',
      fromAvatar: u?.avatarUrl || '',
      contentType,
      content,
      replyToId: reply?.replyToId,
      replyToFromName: reply?.replyToFromName,
      replyToContentType: reply?.replyToContentType,
      replyToContent: reply?.replyToContent,
      createdAt: Math.floor(Date.now() / 1000),
      clientMsgId,
      sendStatus: 'sending',
    }
  }

  return {
    applyAckEverywhere,
    failOptimisticEverywhere,
    failAllOptimisticEverywhere,
    insertOptimistic,
    insertOptimisticPost,
    buildOptimistic,
  }
}

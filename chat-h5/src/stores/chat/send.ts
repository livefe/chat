import type { Ref } from 'vue'

import type { MessageItem, MessageReplyQuote } from '@/api/message'
import {
  channelToMessageScope,
  MessageContentTypes,
  PENDING_SEND_TIMEOUT_MS,
  type ActiveChatChannel,
  type ChatChannel,
} from '@/constants/chat'
import type { WSMessage } from '@/utils/chat/chatWs'

import type { ChatLocationSource, ChatNotice, SendOptions, SendResult } from './types'

export interface SendHelpersDeps {
  activeChannel: Ref<ActiveChatChannel>
  currentRoomId: Ref<number>
  currentPeerId: Ref<number>
  getLocation: ChatLocationSource
  reportLocation: () => void
  getClient: () => { isOpen: () => boolean; send: (msg: WSMessage) => boolean } | null
  connect: () => void
  waitForOpen: (ms: number) => Promise<boolean>
  insertOptimistic: (item: MessageItem) => void
  insertOptimisticPost: (postId: number, item: MessageItem) => void
  buildOptimistic: (
    content: string,
    clientMsgId: string,
    extra: Pick<MessageItem, 'scope' | 'roomId' | 'postId' | 'toUserId'>,
    contentType?: number,
    reply?: MessageReplyQuote | null,
  ) => MessageItem
  failOptimisticEverywhere: (clientMsgId: string, reason?: string) => void
  notify: (notice: ChatNotice) => void
}

interface SendTarget {
  channel: ChatChannel
  roomId?: number
  postId?: number
  toUserId?: number
  coords?: { lat: number; lng: number } | null
  attachCurrentLocation?: boolean
}

const OK: SendResult = { ok: true }

function fail(reason: Exclude<SendResult, { ok: true }>['reason']): SendResult {
  return { ok: false, reason }
}

export function createSendHelpers(deps: SendHelpersDeps) {
  const pendingSendIds = new Set<string>()
  const pendingSendTimers = new Map<string, ReturnType<typeof setTimeout>>()

  function clearPendingSend(clientMsgId: string) {
    pendingSendIds.delete(clientMsgId)
    const timer = pendingSendTimers.get(clientMsgId)
    if (timer) {
      clearTimeout(timer)
    }
    pendingSendTimers.delete(clientMsgId)
  }

  function clearAllPendingSends() {
    pendingSendIds.clear()
    for (const timer of pendingSendTimers.values()) {
      clearTimeout(timer)
    }
    pendingSendTimers.clear()
  }

  function trackPendingSend(clientMsgId: string) {
    if (!clientMsgId) return
    pendingSendIds.add(clientMsgId)
    const timer = setTimeout(() => {
      if (pendingSendIds.has(clientMsgId)) {
        clearPendingSend(clientMsgId)
        deps.failOptimisticEverywhere(clientMsgId)
        deps.notify({ type: 'sendTimeout' })
      }
    }, PENDING_SEND_TIMEOUT_MS)
    pendingSendTimers.set(clientMsgId, timer)
  }

  function buildPayload(
    target: SendTarget,
    content: string,
    contentType: number,
    clientMsgId: string,
    reply?: MessageReplyQuote | null,
  ): WSMessage {
    const payload: WSMessage = {
      type: 'send',
      scope: target.channel,
      content,
      contentType,
      clientMsgId,
    }
    if (target.channel === 'room') payload.roomId = target.roomId
    if (target.channel === 'post') payload.postId = target.postId
    if (target.channel === 'private') payload.toUserId = target.toUserId
    const coords = target.attachCurrentLocation ? deps.getLocation() : target.coords
    if (coords) {
      payload.lat = coords.lat
      payload.lng = coords.lng
    }
    if (reply?.replyToId) payload.replyToId = reply.replyToId
    return payload
  }

  async function dispatch(
    target: SendTarget,
    content: string,
    contentType: number,
    reply: MessageReplyQuote | null | undefined,
    opts: SendOptions | undefined,
  ): Promise<SendResult> {
    if (!deps.getClient()) {
      deps.connect()
    }
    if (!deps.getClient()?.isOpen()) {
      opts?.onConnecting?.()
      const opened = await deps.waitForOpen(5000)
      if (!opened) return fail('connectFailed')
    }
    const client = deps.getClient()
    if (!client?.isOpen()) return fail('connectFailed')
    const clientMsgId = `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
    const payload = buildPayload(target, content, contentType, clientMsgId, reply)
    const opt = deps.buildOptimistic(
      content,
      clientMsgId,
      {
        scope: channelToMessageScope(target.channel),
        roomId: target.roomId ?? 0,
        postId: target.postId ?? 0,
        toUserId: target.toUserId ?? 0,
      },
      contentType,
      reply,
    )
    if (target.channel === 'post' && opt.postId > 0) {
      deps.insertOptimisticPost(opt.postId, opt)
    } else {
      deps.insertOptimistic(opt)
    }
    if (client.send(payload)) {
      trackPendingSend(clientMsgId)
      return OK
    }
    if (opt.clientMsgId) deps.failOptimisticEverywhere(opt.clientMsgId)
    return fail('sendFailed')
  }

  function sendToNearby(
    content: string,
    contentType: number = MessageContentTypes.Text,
    reply?: MessageReplyQuote | null,
    opts?: SendOptions,
  ): Promise<SendResult> {
    if (!content) return Promise.resolve(fail('invalid'))
    const coords = deps.getLocation()
    if (!coords) return Promise.resolve(fail('noLocation'))
    deps.reportLocation()
    return dispatch({ channel: 'nearby', coords }, content, contentType, reply, opts)
  }

  function roomTarget(roomId: number): SendTarget {
    return { channel: 'room', roomId, attachCurrentLocation: true }
  }

  function send(
    content: string,
    contentType: number = MessageContentTypes.Text,
    reply?: MessageReplyQuote | null,
    opts?: SendOptions,
  ): Promise<SendResult> {
    if (!content) return Promise.resolve(fail('invalid'))
    const channel = deps.activeChannel.value
    if (channel === 'nearby') return sendToNearby(content, contentType, reply, opts)
    if (channel === 'room') {
      return dispatch(roomTarget(deps.currentRoomId.value), content, contentType, reply, opts)
    }
    return dispatch(
      { channel: 'private', toUserId: deps.currentPeerId.value },
      content,
      contentType,
      reply,
      opts,
    )
  }

  function sendToRoom(
    roomId: number,
    content: string,
    contentType: number = MessageContentTypes.Text,
    reply?: MessageReplyQuote | null,
    opts?: SendOptions,
  ): Promise<SendResult> {
    if (!content || roomId <= 0) return Promise.resolve(fail('invalid'))
    return dispatch(roomTarget(roomId), content, contentType, reply, opts)
  }

  function sendToPrivateChat(
    toUserId: number,
    content: string,
    contentType: number = MessageContentTypes.Text,
    reply?: MessageReplyQuote | null,
    opts?: SendOptions,
  ): Promise<SendResult> {
    if (!content || toUserId <= 0) return Promise.resolve(fail('invalid'))
    return dispatch({ channel: 'private', toUserId }, content, contentType, reply, opts)
  }

  function sendToPost(
    postId: number,
    content: string,
    contentType: number = MessageContentTypes.Text,
    reply?: MessageReplyQuote | null,
    opts?: SendOptions,
  ): Promise<SendResult> {
    if (!content || postId <= 0) return Promise.resolve(fail('invalid'))
    return dispatch({ channel: 'post', postId }, content, contentType, reply, opts)
  }

  return {
    send,
    sendToNearby,
    sendToRoom,
    sendToPrivateChat,
    sendToPost,
    clearPendingSend,
    clearAllPendingSends,
  }
}

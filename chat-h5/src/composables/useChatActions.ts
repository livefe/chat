import { showToast } from 'vant'

import type { MessageItem, MessageReplyQuote } from '@/api/message'
import { NEARBY_COPY } from '@/constants/nearby'
import { useChatStore, type SendFailReason, type SendResult } from '@/stores/chat'
import { showCatchError } from '@/utils/toast'

const SEND_FAIL_TEXT: Record<SendFailReason, string> = {
  invalid: '',
  noLocation: NEARBY_COPY.chooseLocation,
  connectFailed: '连接失败，请稍后再试',
  sendFailed: '发送失败，请稍后再试',
}

const sendOptions = {
  onConnecting: () => showToast('尚未连接服务器，正在连接…'),
}

function reportSendResult(result: SendResult): boolean {
  if (result.ok) return true
  const text = SEND_FAIL_TEXT[result.reason]
  if (text) showToast(text)
  return false
}

export function useChatActions() {
  const chatStore = useChatStore()

  async function send(content: string, contentType?: number, reply?: MessageReplyQuote | null) {
    return reportSendResult(await chatStore.send(content, contentType, reply, sendOptions))
  }

  async function sendToNearby(content: string, contentType?: number) {
    return reportSendResult(
      await chatStore.sendToNearby(content, contentType, undefined, sendOptions),
    )
  }

  async function sendToRoom(roomId: number, content: string, contentType?: number) {
    return reportSendResult(
      await chatStore.sendToRoom(roomId, content, contentType, undefined, sendOptions),
    )
  }

  async function sendToPrivateChat(peerId: number, content: string, contentType?: number) {
    return reportSendResult(
      await chatStore.sendToPrivateChat(peerId, content, contentType, undefined, sendOptions),
    )
  }

  async function sendToPost(
    postId: number,
    content: string,
    contentType?: number,
    reply?: MessageReplyQuote | null,
  ) {
    return reportSendResult(
      await chatStore.sendToPost(postId, content, contentType, reply, sendOptions),
    )
  }

  async function loadMoreHistory(): Promise<boolean> {
    try {
      return await chatStore.loadMoreHistory()
    } catch (e) {
      showCatchError(e, '加载历史消息失败')
      return false
    }
  }

  async function loadMorePostMessages(postId: number): Promise<boolean> {
    try {
      return await chatStore.loadMorePostMessages(postId)
    } catch (e) {
      showCatchError(e, '加载历史消息失败')
      return false
    }
  }

  async function recall(msg: MessageItem) {
    if (!msg.id || msg.recalled) return
    try {
      await chatStore.recallMessage(msg.id)
    } catch (e) {
      showCatchError(e, '撤回失败', true)
    }
  }

  return {
    send,
    sendToNearby,
    sendToRoom,
    sendToPrivateChat,
    sendToPost,
    loadMoreHistory,
    loadMorePostMessages,
    recall,
  }
}

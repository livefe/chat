import type { Router } from 'vue-router'

import { encodeChatPost } from '@/utils/chat/chatPost'
import { MessageContentTypes } from '@/constants/chat'

export type ForwardPayload = {
  contentType: number
  content: string
  pageTitle?: string
  postId?: number
  title?: string
  cover?: string
  shareUrl?: string
  shareContent?: string
}

let pending: ForwardPayload | null = null

function setForwardPayload(payload: ForwardPayload) {
  pending = payload
}

export function consumeForwardPayload(): ForwardPayload | null {
  const fromState = (window.history.state as { forwardPayload?: ForwardPayload } | null)
    ?.forwardPayload
  if (fromState?.content?.trim()) return fromState
  const p = pending
  pending = null
  return p
}

function openForward(router: Router, payload: ForwardPayload) {
  if (!payload.content.trim()) return
  setForwardPayload(payload)
  router.push({
    name: 'Forward',
    state: { forwardPayload: payload },
  })
}

export function openForwardMessage(router: Router, msg: { contentType: number; content: string }) {
  openForward(router, {
    contentType: msg.contentType,
    content: msg.content,
    pageTitle: '转发',
  })
}

export function openForwardPost(
  router: Router,
  opts: {
    postId: number
    title: string
    cover: string
    shareUrl: string
    shareContent: string
  },
) {
  openForward(router, {
    contentType: MessageContentTypes.Post,
    content: encodeChatPost({
      postId: opts.postId,
      title: opts.title,
      cover: opts.cover,
    }),
    pageTitle: '分享动态',
    postId: opts.postId,
    title: opts.title,
    cover: opts.cover,
    shareUrl: opts.shareUrl,
    shareContent: opts.shareContent,
  })
}

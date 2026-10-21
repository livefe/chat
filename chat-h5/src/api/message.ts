import { get, post } from './request'

export interface MessageItem {
  id: number
  scope: number
  roomId: number
  postId: number
  fromUserId: number
  toUserId: number
  fromName: string
  fromAvatar: string
  contentType: number
  content: string
  replyToId?: number
  replyToFromName?: string
  replyToContentType?: number
  replyToContent?: string
  createdAt: number
  recalled?: boolean

  clientMsgId?: string

  sendStatus?: 'sending' | 'failed'
  sendFailReason?: string
}

export type MessageReplyQuote = {
  replyToId: number
  replyToFromName: string
  replyToContentType: number
  replyToContent: string
}

export function roomMessages(
  roomId: number,
  beforeId: number,
  limit: number,
  lat?: number,
  lng?: number,
) {
  const params: Record<string, number> = { roomId, beforeId, limit }
  if (lat !== undefined && lat !== 0) params.lat = lat
  if (lng !== undefined && lng !== 0) params.lng = lng
  return get<{ list: MessageItem[] }>('/api/message/room', params)
}

export function postMessages(postId: number, beforeId: number, limit: number) {
  return get<{ list: MessageItem[] }>('/api/message/post', { postId, beforeId, limit })
}

export function nearbyMessages(
  lat: number,
  lng: number,
  beforeId: number,
  limit: number,
  radius: number,
) {
  return get<{ list: MessageItem[] }>('/api/message/nearby', {
    lat,
    lng,
    radius,
    beforeId,
    limit,
  })
}

export function privateMessages(peerId: number, beforeId: number, limit: number) {
  return get<{ list: MessageItem[]; clearedBeforeId: number }>('/api/message/private', {
    peerId,
    beforeId,
    limit,
  })
}

export function recallMessage(id: number) {
  return post<null>('/api/message/recall', { id })
}

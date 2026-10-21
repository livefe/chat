import { get, post } from './request'

export interface ConversationItem {
  peerId: number
  peerNickname: string
  peerAvatar: string
  lastContent: string
  lastActiveAt: number
  unreadCount: number
  pinned?: boolean
}

export interface ConversationUnreadSummary {
  total: number
  items: { peerId: number; count: number }[]
}

export function conversationList(page: number, pageSize: number) {
  return get<{ total: number; list: ConversationItem[] }>('/api/conversation/list', {
    page,
    pageSize,
  })
}

export function conversationUnreadSummary() {
  return get<ConversationUnreadSummary>('/api/conversation/unread', undefined, { silent: true })
}

export function deleteConversation(peerId: number) {
  return post<{ clearedBeforeId: number }>('/api/conversation/delete', { peerId })
}

export function pinConversation(peerId: number, pinned: boolean) {
  return post<null>('/api/conversation/pin', { peerId, pinned })
}

export function markConversationRead(peerId: number) {
  return post<null>('/api/conversation/read', { peerId }, { silent: true })
}

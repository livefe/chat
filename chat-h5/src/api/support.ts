import { get, post } from './request'

export interface SupportProfile {
  brandName: string
  brandAvatarUrl: string
  welcomeText: string
  offlineText: string
  enabled: boolean
}

export interface SupportSession {
  id: number
  status: number
  statusText: string
  lastMessageId: number
  lastContent: string
  lastActiveAt: number
  unreadCount: number
  assigneeEmployeeId?: number
  assigneeName?: string
  queueAt?: number
  brandName: string
  brandAvatarUrl: string
}

export interface SupportMessage {
  id: number
  conversationId: number
  senderKind: number
  senderId: number
  senderName: string
  senderAvatarUrl: string
  contentType: number
  content: string
  createdAt: number
  recalled: boolean
  clientMsgId?: string
}

export function getSupportProfile(silent = false) {
  return get<SupportProfile>(
    '/api/support/profile',
    undefined,
    silent ? { silent: true } : undefined,
  )
}

export function getSupportCurrent(silent = false) {
  return get<SupportSession | null>(
    '/api/support/session/current',
    undefined,
    silent ? { silent: true } : undefined,
  )
}

export function openSupportSession(source = 'h5') {
  return post<SupportSession>('/api/support/session/open', { source })
}

export function getSupportMessages(params: {
  conversationId: number
  beforeId?: number
  limit: number
}) {
  return get<{ list: SupportMessage[] }>('/api/support/messages', params)
}

export function sendSupportMessage(data: {
  conversationId: number
  contentType?: number
  content: string
  clientMsgId?: string
}) {
  return post<SupportMessage>('/api/support/messages', data)
}

export function readSupportSession(conversationId: number, lastReadMsgId?: number, silent = false) {
  return post(
    '/api/support/read',
    { conversationId, lastReadMsgId },
    silent ? { silent: true } : undefined,
  )
}

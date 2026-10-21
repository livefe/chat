import { request } from './http'

export interface SupportAdminSession {
  id: number
  status: number
  statusText: string
  userId: number
  userNickname: string
  userAvatarUrl: string
  userMobile: string
  assigneeEmployeeId: number
  assigneeName: string
  lastMessageId: number
  lastContent: string
  lastActiveAt: number
  unreadCount: number
  queueAt: number
  acceptAt: number
  closeAt?: number
  queueWaitSec: number
}

export interface SupportMessageItem {
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

export interface SupportConfig {
  brandName: string
  brandAvatarUrl: string
  welcomeText: string
  offlineText: string
  enabled: boolean
}

export const getSupportQueue = (params: { page: number; pageSize: number }) =>
  request.get<{ total: number; list: SupportAdminSession[] }>('/api/admin/support/queue', params)

export const getSupportMine = (params: { page: number; pageSize: number }) =>
  request.get<{ total: number; list: SupportAdminSession[] }>('/api/admin/support/mine', params)

export const getSupportClosed = (params: { page: number; pageSize: number }) =>
  request.get<{ total: number; list: SupportAdminSession[] }>('/api/admin/support/closed', params)

export const getSupportActiveAll = (params: {
  page: number
  pageSize: number
  assigneeEmployeeId?: number
  keyword?: string
}) =>
  request.get<{ total: number; list: SupportAdminSession[] }>(
    '/api/admin/support/active-all',
    params,
  )

export const getSupportClosedAll = (params: {
  page: number
  pageSize: number
  assigneeEmployeeId?: number
  keyword?: string
}) =>
  request.get<{ total: number; list: SupportAdminSession[] }>(
    '/api/admin/support/closed-all',
    params,
  )

export const getSupportOverviewSummary = () =>
  request.get<{
    queueCount: number
    activeCount: number
    longWaitQueueCount: number
  }>('/api/admin/support/overview/summary')

export const acceptSupport = (conversationId: number) =>
  request.post<SupportAdminSession>('/api/admin/support/accept', { conversationId })

export const transferSupport = (conversationId: number, toEmployeeId: number) =>
  request.post('/api/admin/support/transfer', { conversationId, toEmployeeId })

export const closeSupport = (conversationId: number, reason?: string) =>
  request.post('/api/admin/support/close', { conversationId, reason })

export const getSupportMessages = (params: {
  conversationId: number
  beforeId?: number
  limit: number
}) => request.get<{ list: SupportMessageItem[] }>('/api/admin/support/messages', params)

export const sendSupportMessage = (data: {
  conversationId: number
  contentType?: number
  content: string
  clientMsgId?: string
}) => request.post<SupportMessageItem>('/api/admin/support/messages', data)

export const readSupport = (conversationId: number, lastReadMsgId?: number) =>
  request.post('/api/admin/support/read', { conversationId, lastReadMsgId })

export const getSupportConfig = () => request.get<SupportConfig>('/api/admin/support/config')

export const updateSupportConfig = (
  data: Partial<Omit<SupportConfig, 'enabled'>> & { enabled?: boolean | number },
) => request.post<SupportConfig>('/api/admin/support/config', data)

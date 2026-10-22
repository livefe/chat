import { request } from './http'

export interface SupportQuickReply {
  title: string
  content: string
}

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
  closeReason: string
  closeReasonText: string
  closeNote: string
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
  quickReplies: SupportQuickReply[]
}

export interface SupportUserSession {
  id: number
  status: number
  statusText: string
  assigneeName: string
  createdAt: string
  acceptedAt: string
  closedAt: string
  closeReason: string
  closeReasonText: string
  closeNote: string
  lastMessage: string
  lastMessageAt: string
}

export interface SupportAgent {
  id: number
  name: string
  avatarUrl: string
  departmentName: string
  online: boolean
}

type SessionPage = { total: number; list: SupportAdminSession[] }
type UserSessionPage = { total: number; list: SupportUserSession[] }

export interface SupportOverviewParams {
  page: number
  pageSize: number
  assigneeEmployeeId?: number
  keyword?: string
  startDate?: string
  endDate?: string
}

export const getSupportQueue = (params: { page: number; pageSize: number; keyword?: string }) =>
  request.get<SessionPage>('/api/admin/support/queue', params)

export const getSupportMine = (params: { page: number; pageSize: number }) =>
  request.get<SessionPage>('/api/admin/support/mine', params)

export const getSupportClosed = (params: { page: number; pageSize: number }) =>
  request.get<SessionPage>('/api/admin/support/closed', params)

export const getSupportActiveAll = (params: SupportOverviewParams) =>
  request.get<SessionPage>('/api/admin/support/active-all', params)

export const getSupportClosedAll = (params: SupportOverviewParams) =>
  request.get<SessionPage>('/api/admin/support/closed-all', params)

export const getSupportSession = (conversationId: number) =>
  request.get<SupportAdminSession>('/api/admin/support/session', { conversationId })

export const getSupportOverviewSummary = () =>
  request.get<{
    queueCount: number
    activeCount: number
    longWaitQueueCount: number
  }>('/api/admin/support/overview/summary')

export const getUserSupportSessions = (params: {
  userId: number
  page: number
  pageSize: number
}) => request.get<UserSessionPage>('/api/admin/support/by-user', params)

export const getSupportHistory = (params: {
  conversationId: number
  page: number
  pageSize: number
}) => request.get<UserSessionPage>('/api/admin/support/history', params)

export const getSupportAgents = () =>
  request.get<{ list: SupportAgent[] }>('/api/admin/support/agents')

export const getSupportQuickReplies = () =>
  request.get<{ list: SupportQuickReply[] }>('/api/admin/support/quick-replies')

export const acceptSupport = (conversationId: number) =>
  request.post<SupportAdminSession>('/api/admin/support/accept', { conversationId })

export const transferSupport = (conversationId: number, toEmployeeId: number) =>
  request.post('/api/admin/support/transfer', { conversationId, toEmployeeId })

export const closeSupport = (data: { conversationId: number; reason?: string; note?: string }) =>
  request.post('/api/admin/support/close', data)

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

export const updateSupportConfig = (data: Omit<SupportConfig, 'enabled'> & { enabled: number }) =>
  request.post<SupportConfig>('/api/admin/support/config', data)

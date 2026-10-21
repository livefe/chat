import { request, uploadForm } from './http'

export interface ImContactItem {
  id: number
  name: string
  avatarUrl: string
  departmentId: number
  departmentName: string
  online: boolean
}

export interface ImConversationItem {
  id: number
  type: number
  title: string
  peerId: number
  peerName: string
  peerAvatarUrl: string
  lastMessageId: number
  lastContent: string
  lastActiveAt: number
  unreadCount: number
  pinned: boolean
}

export interface ImMessageItem {
  id: number
  conversationId: number
  senderKind: number
  senderId: number
  senderName: string
  senderAvatarUrl: string
  contentType: number
  content: string
  replyToId?: number
  replyToSenderName?: string
  replyToContentType?: number
  replyToContent?: string
  createdAt: number
  recalled: boolean
  clientMsgId?: string
}

export const getImContacts = (params: {
  page: number
  pageSize: number
  keyword?: string
  departmentId?: number
}) => request.get<{ total: number; list: ImContactItem[] }>('/api/admin/im/contacts', params)

export const openImDM = (peerEmployeeId: number) =>
  request.post<ImConversationItem>('/api/admin/im/conversations/dm', { peerEmployeeId })

export const getImConversations = (params: { page: number; pageSize: number }) =>
  request.get<{ total: number; list: ImConversationItem[] }>('/api/admin/im/conversations', params)

export const getImMessages = (params: {
  conversationId: number
  beforeId?: number
  limit: number
}) => request.get<{ list: ImMessageItem[] }>('/api/admin/im/messages', params)

export const sendImMessage = (data: {
  conversationId: number
  contentType?: number
  content: string
  replyToId?: number
  clientMsgId?: string
}) => request.post<ImMessageItem>('/api/admin/im/messages', data)

export const readImConversation = (data: { conversationId: number; lastReadMsgId?: number }) =>
  request.post('/api/admin/im/read', data)

export const recallImMessage = (messageId: number) =>
  request.post('/api/admin/im/recall', { messageId })

export const uploadImImage = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return uploadForm<{ id: number; url: string }>('/api/admin/upload/image', form)
}

export const uploadImVideo = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return uploadForm<{ id: number; url: string }>('/api/admin/upload/video', form, 180000)
}

export const uploadImFile = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return uploadForm<{ id: number; url: string; mimeType?: string; sizeBytes?: number }>(
    '/api/admin/upload/file',
    form,
    180000,
  )
}

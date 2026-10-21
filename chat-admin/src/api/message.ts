import { request } from './http'

export interface AdminMessageItem {
  id: number
  scope: number
  scopeText: string
  roomId: number
  roomName: string
  postId?: number
  fromUserId: number
  toUserId: number
  fromName: string
  fromAvatar: string
  toName: string
  contentType: number
  content: string
  replyToId?: number
  replyToFromName?: string
  replyToContentType?: number
  replyToContent?: string
  replyToRecalled?: boolean
  replyToDeleted?: boolean
  latitude: number
  longitude: number
  createdAt: number
  recalled?: boolean
  deleted: boolean
  deletedAt: string
}

export interface AdminMessageListRequest {
  page: number
  pageSize: number
  userId?: number
  peerUserId?: number
  roomId?: number
  postId?: number
  scope?: number
  keyword?: string
  startTime?: string
  endTime?: string
}

export interface AdminMessageListResponse {
  total: number
  list: AdminMessageItem[]
}

export const getMessageList = (params: AdminMessageListRequest) =>
  request.get<AdminMessageListResponse>('/api/admin/message/list', params)

export const getMessageContext = (id: number, limit = 20) =>
  request.get<{ list: AdminMessageItem[] }>('/api/admin/message/context', { id, limit })

export const deleteMessage = (id: number) => request.post('/api/admin/message/delete', { id })

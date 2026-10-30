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
  fromMobile: string
  fromAvatar: string
  toName: string
  toMobile: string
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
  id?: number
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

export const getMessageContext = (
  id: number,
  options: { limit?: number; beforeId?: number; afterId?: number } = {},
) =>
  request.get<{ list: AdminMessageItem[] }>('/api/admin/message/context', {
    id,
    limit: options.limit ?? 20,
    beforeId: options.beforeId,
    afterId: options.afterId,
  })

export interface DeleteMessagesRequest {
  ids: number[]
  reason: string
  muteDuration: number
}

export interface DeleteMessagesResult {
  deleted: number
  skipped: number
  muted: number
}

export const deleteMessages = (data: DeleteMessagesRequest) =>
  request.post<DeleteMessagesResult>('/api/admin/message/batch/delete', data)

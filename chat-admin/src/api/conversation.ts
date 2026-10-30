import { request } from './http'

export interface AdminConversationItem {
  id: number
  userMinId: number
  userMaxId: number
  userMinName: string
  userMaxName: string
  userMinMobile: string
  userMaxMobile: string
  lastContent: string
  lastActiveAt: number
  createdAt: string
}

export interface AdminConversationListRequest {
  page: number
  pageSize: number
  userId?: number
  peerUserId?: number
  startTime?: string
  endTime?: string
}

export interface AdminConversationListResponse {
  total: number
  list: AdminConversationItem[]
}

export const getConversationList = (params: AdminConversationListRequest) =>
  request.get<AdminConversationListResponse>('/api/admin/conversation/list', params)

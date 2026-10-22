import { request } from './http'

export interface FeedbackItem {
  id: number
  userId: number
  userNickname?: string
  userMobile?: string
  type: number
  typeText: string
  content: string
  contact: string
  images: string[]
  status: number
  statusText: string
  handleResult: string
  employeeId: number
  employeeName?: string
  handledAt: string
  createdAt: string
  updatedAt: string
}

export interface FeedbackListRequest {
  page: number
  pageSize: number
  status?: number
  type?: number
  userId?: number
  keyword?: string
  startTime?: string
  endTime?: string
}

export interface FeedbackListResponse {
  total: number
  list: FeedbackItem[]
}

export interface UpdateFeedbackStatusRequest {
  id: number
  status: number
  handleResult: string
}

export const getFeedbackList = (params: FeedbackListRequest) =>
  request.get<FeedbackListResponse>('/api/admin/feedback/list', params)

export const getFeedbackDetail = (id: number) =>
  request.get<FeedbackItem>('/api/admin/feedback/detail', { id })

export const updateFeedbackStatus = (data: UpdateFeedbackStatusRequest) =>
  request.post('/api/admin/feedback/update/status', data)

export const batchCloseFeedback = (ids: number[], handleResult: string) =>
  request.post<{ updated: number; skipped: number }>('/api/admin/feedback/batch/close', {
    ids,
    handleResult,
  })

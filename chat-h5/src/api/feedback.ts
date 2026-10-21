import { get, post } from './request'

export { feedbackTypeOptions } from '@/constants/feedback'

export interface FeedbackItem {
  id: number
  type: number
  typeText: string
  content: string
  contact: string
  images: string[]
  status: number
  statusText: string
  reply: string
  handledAt: string
  createdAt: string
}

export interface FeedbackListResponse {
  total: number
  list: FeedbackItem[]
}

export interface SubmitFeedbackRequest {
  type: number
  content: string
  contact?: string
  images?: number[]
}

export function submitFeedback(data: SubmitFeedbackRequest) {
  return post('/api/feedback/submit', data)
}

export function getMyFeedbackList(params: { page: number; pageSize: number }) {
  return get<FeedbackListResponse>('/api/feedback/my/list', params)
}

export function getMyFeedbackDetail(id: number) {
  return get<FeedbackItem>('/api/feedback/my/detail', { id })
}

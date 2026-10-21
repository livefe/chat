import { request } from './http'

export interface ReportContextSnapshot {
  nickname?: string
  avatar?: string
  postTitle?: string
  postContent?: string
  postImages?: string[]
  messageContent?: string
  messageType?: number
  scope?: number
  roomId?: number
  peerId?: number
  sentAt?: string
}

export interface ReportItem {
  id: number
  reporterId: number
  reporterNickname?: string
  reporterMobile?: string
  targetType: number
  targetTypeText: string
  targetId: number
  targetUserId: number
  targetUserName?: string
  reason: number
  reasonText: string
  content: string
  images: string[]
  context: ReportContextSnapshot
  status: number
  statusText: string
  handleResult: string
  handleAction: number
  handleActionText: string
  employeeId: number
  employeeName?: string
  handledAt: string
  targetReportCount?: number
  createdAt: string
  updatedAt: string
}

export interface ReportListRequest {
  page: number
  pageSize: number
  status?: number
  targetType?: number
  reason?: number
  keyword?: string
}

export interface ReportListResponse {
  total: number
  list: ReportItem[]
}

export interface HandleReportRequest {
  id: number
  status: number
  handleResult?: string
  action?: number
  muteHours?: number
}

export const getReportList = (params: ReportListRequest) =>
  request.get<ReportListResponse>('/api/admin/report/list', params)

export const getReportDetail = (id: number) =>
  request.get<ReportItem>('/api/admin/report/detail', { id })

export const handleReport = (data: HandleReportRequest) =>
  request.post('/api/admin/report/handle', data)

import { request } from './http'
import type { UserGovernanceSummary } from './user'

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
  postId?: number
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
  targetUserMobile?: string
  /** 仅详情接口返回。 */
  targetUser?: UserGovernanceSummary
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

/** status 为待处理时按提交时间正序返回（先进先出）。 */
export interface ReportListRequest {
  page: number
  pageSize: number
  status?: number
  targetType?: number
  targetId?: number
  targetUserId?: number
  reporterId?: number
  reason?: number
  keyword?: string
  startTime?: string
  endTime?: string
}

export interface ReportListResponse {
  total: number
  list: ReportItem[]
}

export interface HandleReportRequest {
  id: number
  status: number
  handleResult: string
  action: number
  /** 禁言小时数，-1 永久；仅禁言动作使用。 */
  muteDuration: number
  /** 同对象的其它待处理举报按相同结果一并结案。 */
  resolveSameTarget: boolean
}

export const getReportList = (params: ReportListRequest) =>
  request.get<ReportListResponse>('/api/admin/report/list', params)

export const getReportDetail = (id: number) =>
  request.get<ReportItem>('/api/admin/report/detail', { id })

export const handleReport = (data: HandleReportRequest) =>
  request.post<{ resolvedCount: number }>('/api/admin/report/handle', data)

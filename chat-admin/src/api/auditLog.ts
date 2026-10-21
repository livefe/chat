import { request } from './http'

export interface AuditModuleOption {
  value: string
  label: string
}

export interface AuditLogItem {
  id: number
  employeeId: number
  employeeName: string
  employeeMobile: string
  module: string
  moduleText: string
  action: string
  actionText: string
  targetType: string
  targetId: number
  targetLabel: string
  requestPath: string
  requestMethod: string
  success: boolean
  successText: string
  detail: string
  ip: string
  userAgent: string
  createdAt: string
}

export interface AuditLogListRequest {
  page: number
  pageSize: number
  employeeId?: number
  module?: string
  action?: string
  keyword?: string
  startTime?: string
  endTime?: string
}

export interface AuditActionOption {
  value: string
  label: string
}

export interface AuditLogListResponse {
  total: number
  list: AuditLogItem[]
  modules: AuditModuleOption[]
  actions: AuditActionOption[]
}

export const getAuditLogList = (params: AuditLogListRequest) =>
  request.get<AuditLogListResponse>('/api/admin/audit/list', params)

export const getAuditLogDetail = (id: number) =>
  request.get<AuditLogItem>('/api/admin/audit/detail', { id })

import { request } from './http'

export interface AuditOption {
  value: string
  label: string
}

export interface AuditActionOption extends AuditOption {
  module: string
}

export interface AuditOptions {
  modules: AuditOption[]
  actions: AuditActionOption[]
  targetTypes: AuditOption[]
}

export interface AuditEmployeeOption {
  id: number
  name: string
  mobile: string
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
  targetTypeText: string
  targetId: number
  targetLabel: string
  requestPath: string
  requestMethod: string
  success: boolean
  successText: string
  resultCode: number
  resultMessage: string
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
  targetType?: string
  targetId?: number
  requestPath?: string
  success?: boolean
  keyword?: string
  startTime?: string
  endTime?: string
}

export interface AuditLogListResponse {
  total: number
  list: AuditLogItem[]
}

export const getAuditLogList = (params: AuditLogListRequest) =>
  request.get<AuditLogListResponse>('/api/admin/audit/list', params)

export const getAuditLogOptions = () => request.get<AuditOptions>('/api/admin/audit/options')

export const searchAuditEmployees = (params: { keyword?: string; id?: number }) =>
  request.get<AuditEmployeeOption[]>('/api/admin/audit/employees', params)

export const getAuditLogDetail = (id: number) =>
  request.get<AuditLogItem>('/api/admin/audit/detail', { id })

import { request } from './http'

export interface UserItem {
  id: number
  mobile: string
  nickname: string
  avatarUrl: string
  status: number
  statusText: string
  mutedUntil: number
  muteText: string
  deactivated: boolean
  createdAt: string
  updatedAt: string
}

export interface UserListRequest {
  page: number
  pageSize: number
  id?: number
  mobile?: string
  nickname?: string
  status?: number
  muted?: boolean
}

export interface UserListResponse {
  total: number
  list: UserItem[]
}

export const getUserList = (params: UserListRequest) =>
  request.get<UserListResponse>('/api/admin/user/list', params)

export const createUser = (data: { mobile: string; nickname?: string; status: number }) =>
  request.post('/api/admin/user/create', data)

export const updateUser = (data: {
  id: number
  mobile: string
  nickname?: string
  status: number
}) => request.post('/api/admin/user/update', data)

export const updateUserStatus = (id: number, status: number) =>
  request.post('/api/admin/user/update/status', { id, status })

export const deactivateUser = (id: number, reason?: string) =>
  request.post('/api/admin/user/deactivate', { id, reason: reason || '' })

export interface LoginRecordItem {
  id: number
  loginType: string
  ip: string
  status: number
  statusText: string
  remark: string
  createdAt: string
}

export interface UserDeactivationLogItem {
  id: number
  mobile: string
  source: number
  sourceText: string
  employeeId: number
  reason: string
  createdAt: string
}

export interface UserRoomSummary {
  roomId: number
  roomName: string
  messageCount: number
  lastMessageAt: string
}

export interface AdminUserDetail extends UserItem {
  loginRecords: LoginRecordItem[]
  deactivationLogs: UserDeactivationLogItem[]
  rooms: UserRoomSummary[]
}

export const getUserDetail = (id: number) =>
  request.get<AdminUserDetail>('/api/admin/user/detail', { id })

export const muteUser = (id: number, duration: number, reason?: string) =>
  request.post('/api/admin/user/update/mute', { id, duration, reason: reason || '' })

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
  lastLoginAt: string
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
  startTime?: string
  endTime?: string
}

export interface PageResponse<T> {
  total: number
  list: T[]
}

export interface UserSubListRequest {
  userId: number
  page: number
  pageSize: number
}

export interface BatchUserResult {
  updated: number
  skipped: number
}

export const getUserList = (params: UserListRequest) =>
  request.get<PageResponse<UserItem>>('/api/admin/user/list', params)

export const createUser = (data: { mobile: string; nickname: string; status: number }) =>
  request.post('/api/admin/user/create', data)

/** mobile 为空表示不修改手机号。 */
export const updateUser = (data: { id: number; mobile?: string; nickname: string }) =>
  request.post('/api/admin/user/update', data)

export const updateUserStatus = (id: number, status: number, reason: string) =>
  request.post('/api/admin/user/update/status', { id, status, reason })

export const batchUpdateUserStatus = (ids: number[], status: number, reason: string) =>
  request.post<BatchUserResult>('/api/admin/user/batch/status', { ids, status, reason })

/** duration：小时数，-1 永久，0 解除禁言。 */
export const muteUser = (id: number, duration: number, reason: string) =>
  request.post('/api/admin/user/update/mute', { id, duration, reason })

export const batchMuteUser = (ids: number[], duration: number, reason: string) =>
  request.post<BatchUserResult>('/api/admin/user/batch/mute', { ids, duration, reason })

export const deactivateUser = (id: number, reason: string) =>
  request.post('/api/admin/user/deactivate', { id, reason })

export const revealUserMobile = (id: number) =>
  request.post<{ mobile: string }>('/api/admin/user/mobile/reveal', { id })

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
  deactivationLogs: UserDeactivationLogItem[]
  rooms: UserRoomSummary[]
}

export const getUserDetail = (id: number) =>
  request.get<AdminUserDetail>('/api/admin/user/detail', { id })

export interface LoginRecordItem {
  id: number
  loginType: string
  ip: string
  userAgent: string
  status: number
  statusText: string
  remark: string
  createdAt: string
}

export const getUserLoginRecords = (params: UserSubListRequest) =>
  request.get<PageResponse<LoginRecordItem>>('/api/admin/user/login-records', params)

export interface UserPunishmentItem {
  id: number
  userId: number
  action: number
  actionText: string
  muteUntil: number
  muteText: string
  targetId: number
  reason: string
  employeeId: number
  employeeName: string
  reportId: number
  createdAt: string
}

export const getUserPunishments = (params: UserSubListRequest) =>
  request.get<PageResponse<UserPunishmentItem>>('/api/admin/user/punishments', params)

export interface UserDeviceItem {
  id: number
  deviceId: string
  platform: string
  vendor: string
  appVersion: string
  osVersion: string
  pushEnabled: boolean
  lastActiveAt: number
  createdAt: string
}

export const getUserDevices = (params: UserSubListRequest) =>
  request.get<PageResponse<UserDeviceItem>>('/api/admin/user/devices', params)

/** 举报审核中展示的被举报用户现状。 */
export interface UserGovernanceSummary {
  userId: number
  nickname: string
  mobile: string
  avatarUrl: string
  status: number
  statusText: string
  muteText: string
  deactivated: boolean
  reportCount: number
  punishmentCount: number
}

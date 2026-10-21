import { get, post } from './request'

export interface UserInfo {
  id: number
  mobile: string
  nickname: string
  avatarUrl: string
  status: number
}

export interface LoginResult {
  accessToken: string
  refreshToken: string
  isFirstLogin: boolean
  user: UserInfo
}

export interface RefreshTokenResult {
  accessToken: string
  refreshToken: string
}

export interface PublicUserProfile {
  id: number
  nickname: string
  avatarUrl: string
  createdAt: string
  isSelf?: boolean
  deactivated?: boolean
  blockedByMe?: boolean
  canMessage?: boolean
}

export interface BlockListItem {
  userId: number
  nickname: string
  avatarUrl: string
  createdAt: string
}

export interface BlockListResponse {
  total: number
  list: BlockListItem[]
}

export function login(mobile: string, code: string, source = 'app_login') {
  return post<LoginResult>('/api/auth/login', { mobile, code, source })
}

export function refreshTokens(refreshToken: string) {
  return post<RefreshTokenResult>('/api/auth/refresh', { refreshToken })
}

export function getUserDetail() {
  return get<UserInfo>('/api/user/me')
}

export function getUserProfile(id: number) {
  return get<PublicUserProfile>('/api/user/profile', { id })
}

export function updateProfile(payload: { nickname?: string; avatarMediaId?: number | null }) {
  return post<null>('/api/user/profile/update', payload)
}

export function sendChangeMobileOldCode() {
  return post<null>('/api/user/mobile/send/old-code')
}

export function verifyChangeMobileOldCode(code: string) {
  return post<{ ticket: string }>('/api/user/mobile/verify/old-code', { code })
}

export function sendChangeMobileNewCode(mobile: string, ticket: string) {
  return post<null>('/api/user/mobile/send/new-code', { mobile, ticket })
}

export function changeMobile(payload: { ticket: string; newMobile: string; newCode: string }) {
  return post<null>('/api/user/mobile/change', payload)
}

export function sendLoginCode(mobile: string) {
  return post<null>('/api/auth/sms/code', { mobile })
}

export function sendDeactivationCode() {
  return post<null>('/api/user/deactivation/send/code')
}

export function deactivateAccount(code: string) {
  return post<null>('/api/user/deactivate', { code })
}

export function blockUser(userId: number) {
  return post<null>('/api/user/block', { userId })
}

export function unblockUser(userId: number) {
  return post<null>('/api/user/unblock', { userId })
}

export function getBlockList(params: { page: number; pageSize: number }) {
  return get<BlockListResponse>('/api/user/blocks', params)
}

import { request } from './http'
import type { EmployeeInfo } from './employee'

interface LoginResponse {
  accessToken: string
  refreshToken: string
  employee: EmployeeInfo
}

interface RefreshTokenResponse {
  accessToken: string
  refreshToken: string
}

export const login = (mobile: string, code: string) =>
  request.post<LoginResponse>('/api/admin/auth/login', { mobile, code })

export const refreshTokenApi = (refreshToken: string) =>
  request.post<RefreshTokenResponse>('/api/admin/auth/refresh', { refreshToken })

import { request } from './http'

export const sendLoginCode = (mobile: string) =>
  request.post<null>('/api/admin/auth/sms/code', { mobile })

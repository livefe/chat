import { request } from './http'

export interface SensitiveWordItem {
  id: number
  word: string
  category: string
  categoryText: string
  level: number
  levelText: string
  action: number
  actionText: string
  status: number
  statusText: string
  remark: string
  createdAt: string
  updatedAt: string
}

export interface SensitiveWordListRequest {
  page: number
  pageSize: number
  word?: string
  category?: string
  status?: number
}

export interface SensitiveWordListResponse {
  total: number
  list: SensitiveWordItem[]
}

export const getSensitiveWordList = (params: SensitiveWordListRequest) =>
  request.get<SensitiveWordListResponse>('/api/admin/sensitive-word/list', params)

export const createSensitiveWord = (data: {
  word: string
  category: string
  level: number
  status: number
  remark?: string
}) => request.post('/api/admin/sensitive-word/create', data)

export const updateSensitiveWord = (data: {
  id: number
  word: string
  category: string
  level: number
  status: number
  remark?: string
}) => request.post('/api/admin/sensitive-word/update', data)

export const updateSensitiveWordStatus = (id: number, status: number) =>
  request.post('/api/admin/sensitive-word/update/status', { id, status })

export const deleteSensitiveWord = (id: number) =>
  request.post('/api/admin/sensitive-word/delete', { id })

export const importSensitiveWords = (data: { words: string; category?: string; level?: number }) =>
  request.post<{ created: number; skipped: number }>('/api/admin/sensitive-word/import', data)

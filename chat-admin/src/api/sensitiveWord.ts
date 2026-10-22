import { request } from './http'

export interface SensitiveWordItem {
  id: number
  word: string
  category: string
  categoryText: string
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

export interface SensitiveWordPayload {
  word: string
  category: string
  status: number
  remark?: string
}

export interface ImportSensitiveWordsResult {
  created: number
  existing: string[]
  tooLong: string[]
  duplicated: string[]
  invalid: string[]
}

export const getSensitiveWordList = (params: SensitiveWordListRequest) =>
  request.get<SensitiveWordListResponse>('/api/admin/sensitive-word/list', params)

export const createSensitiveWord = (data: SensitiveWordPayload) =>
  request.post('/api/admin/sensitive-word/create', data)

export const updateSensitiveWord = (data: SensitiveWordPayload & { id: number }) =>
  request.post('/api/admin/sensitive-word/update', data)

export const updateSensitiveWordStatus = (id: number, status: number) =>
  request.post('/api/admin/sensitive-word/update/status', { id, status })

export const deleteSensitiveWord = (id: number) =>
  request.post('/api/admin/sensitive-word/delete', { id })

export const batchDeleteSensitiveWords = (ids: number[]) =>
  request.post('/api/admin/sensitive-word/batch/delete', { ids })

export const batchUpdateSensitiveWordStatus = (ids: number[], status: number) =>
  request.post('/api/admin/sensitive-word/batch/status', { ids, status })

export const importSensitiveWords = (data: { words: string; category: string }) =>
  request.post<ImportSensitiveWordsResult>('/api/admin/sensitive-word/import', data)

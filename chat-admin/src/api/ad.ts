import { request, uploadForm } from './http'

export interface AdCampaignItem {
  id: number
  name: string
  advertiserName: string
  title: string
  subtitle: string
  imageMediaId: number
  imageUrl: string
  landingType: string
  landingValue: string
  status: number
  statusText: string
  priority: number
  startAt: number
  endAt: number
  cpdPrice: number
  hasGeo: boolean
  latitude: number
  longitude: number
  radiusM: number
  dailyImpressionLimit: number
  rejectRemark: string
  landingInvalid: boolean
  todayImpressions: number
  todayClicks: number
  todayCtr: number
  lowCtrWarning: boolean
  createdAt: string
  updatedAt: string
}

export interface AdCampaignListRequest {
  page: number
  pageSize: number
  keyword?: string
  status?: number
  advertiserName?: string
  startDate?: string
  endDate?: string
}

export interface AdCampaignListResponse {
  total: number
  list: AdCampaignItem[]
}

export interface AdCampaignSaveRequest {
  name?: string
  advertiserName: string
  title: string
  subtitle?: string
  imageMediaId: number
  landingType: string
  landingValue: string
  priority?: number
  startAt: number
  endAt: number
  cpdPrice?: number
  hasGeo?: boolean
  latitude?: number
  longitude?: number
  radiusM?: number
  dailyImpressionLimit?: number
  submitAudit?: boolean
}

export interface AdDailyStatItem {
  statDate: string
  impressions: number
  clicks: number
  ctr: number
}

export interface UploadMediaResponse {
  id: number
  url: string
  key: string
  kind: string
  mimeType: string
  sizeBytes: number
}

export const getAdCampaignList = (params: AdCampaignListRequest) =>
  request.get<AdCampaignListResponse>('/api/admin/ad/campaign/list', params)

export const getAdAdvertisers = () =>
  request.get<{ list: string[] }>('/api/admin/ad/campaign/advertisers')

export const getAdCampaignDetail = (id: number) =>
  request.get<AdCampaignItem>('/api/admin/ad/campaign/detail', { id })

export const saveAdCampaign = (data: AdCampaignSaveRequest & { id?: number }) =>
  request.post<{ id: number }>('/api/admin/ad/campaign/save', data)

/** status：1 提交审核、2 恢复投放（仅已暂停）、4 暂停、5 结束 */
export const updateAdCampaignStatus = (data: { id: number; status: number }) =>
  request.post('/api/admin/ad/campaign/update/status', data)

export const auditAdCampaign = (data: {
  id: number
  action: 'approve' | 'reject'
  rejectRemark?: string
}) => request.post('/api/admin/ad/campaign/audit', data)

export const batchPauseAdCampaigns = (ids: number[]) =>
  request.post<{ paused: number; skipped: number }>('/api/admin/ad/campaign/batch/pause', { ids })

export const copyAdCampaign = (id: number) =>
  request.post<{ id: number }>('/api/admin/ad/campaign/copy', { id })

export const getAdCampaignDailyStats = (params: {
  campaignId: number
  fromDate?: string
  toDate?: string
}) =>
  request.get<{ fromDate: string; toDate: string; list: AdDailyStatItem[] }>(
    '/api/admin/ad/campaign/daily-stats',
    params,
  )

export const uploadAdImage = (file: File) => {
  const form = new FormData()
  form.append('file', file)
  return uploadForm<UploadMediaResponse>('/api/admin/upload/image', form)
}

export interface AdSetting {
  enabled: boolean
  firstSlotAfter: number
  slotInterval: number
  maxActivePerAdvertiser: number
  lowCtrThreshold: number
}

export const getAdSetting = () => request.get<AdSetting>('/api/admin/ad/setting')

export const updateAdSetting = (data: AdSetting) => request.post('/api/admin/ad/setting', data)

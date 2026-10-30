import { request } from './http'

export interface RegionItem {
  code: string
  name: string
  shortName: string
  parentCode: string
  level: number
  fullPath: string
  tiandituGb: string
  amapAdcode: string
}

export interface RegionAdminItem extends RegionItem {
  levelText: string
  status: number
  statusText: string
  dataYear: string
}

export interface RegionListResponse {
  list: RegionItem[]
}

export interface RegionSearchResponse {
  list: RegionItem[]
  truncated: boolean
}

export const getRegionChildren = (parentCode = '') =>
  request.get<RegionListResponse>('/api/admin/region/children', { parentCode })

export const searchRegion = (params: { keyword: string; level?: number; limit: number }) =>
  request.get<RegionSearchResponse>('/api/admin/region/search', params)

export const getRegionTrail = (code: string) =>
  request.get<RegionListResponse>('/api/admin/region/trail', { code })

export const getRegionDetail = (code: string) =>
  request.get<RegionAdminItem>('/api/admin/region/detail', { code })

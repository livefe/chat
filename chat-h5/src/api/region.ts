import { get } from './request'

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

export interface RegionListResponse {
  list: RegionItem[]
}

export function getRegionChildren(parentCode = '') {
  return get<RegionListResponse>('/api/region/children', { parentCode })
}

export function searchRegion(params: { keyword: string; level?: number; limit?: number }) {
  return get<RegionListResponse>('/api/region/search', params)
}

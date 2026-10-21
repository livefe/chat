import { get } from './request'

export interface AppReleasePublicItem {
  platform: string
  platformText: string
  version: string
  downloadUrl: string
}

export interface AppReleasePublicResult {
  list: AppReleasePublicItem[]
}

export function getAppReleaseCurrent() {
  return get<AppReleasePublicResult>('/api/app/release/current')
}

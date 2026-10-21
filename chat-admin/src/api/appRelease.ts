import type { AppPlatform } from '@/constants/appRelease'

import { request, uploadForm } from './http'

export interface AppReleaseItem {
  id: number
  platform: string
  platformText: string
  version: string
  downloadUrl: string
  fileSize: number
  changelog: string
  isCurrent: number
  createdAt: string
}

export interface AppReleaseListResponse {
  list: AppReleaseItem[]
}

export interface AppUpdatePolicyItem {
  platform: string
  platformText: string
  minSupportedVersion: string
  latestVersion: string
  downloadUrl: string
  title: string
  message: string
  updatedAt: string
}

export const getAppReleaseList = (platform?: string) =>
  request.get<AppReleaseListResponse>(
    '/api/admin/app/release/list',
    platform ? { platform } : undefined,
  )

export const uploadAppRelease = (payload: {
  platform: Exclude<AppPlatform, 'h5'>
  version: string
  changelog?: string
  file: File
}) => {
  const form = new FormData()
  form.append('platform', payload.platform)
  form.append('version', payload.version)
  form.append('changelog', payload.changelog || '')
  form.append('file', payload.file)
  return uploadForm<AppReleaseItem>('/api/admin/app/release/upload', form, 180000)
}

export const publishAppReleaseMeta = (payload: {
  platform: 'h5'
  version: string
  downloadUrl: string
  changelog?: string
}) => request.post<AppReleaseItem>('/api/admin/app/release/publish-meta', payload)

export const enableAppRelease = (id: number) =>
  request.post('/api/admin/app/release/set-current', { id })

export const disableAppRelease = (id: number) =>
  request.post('/api/admin/app/release/unset-current', { id })

export const getAppUpdatePolicyList = () =>
  request.get<{ list: AppUpdatePolicyItem[] }>('/api/admin/app/release/policy/list')

export const updateAppUpdatePolicy = (data: {
  platform: AppPlatform
  minSupportedVersion: string
  title: string
  message: string
}) => request.post('/api/admin/app/release/policy/save', data)

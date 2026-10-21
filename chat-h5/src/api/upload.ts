import { MessageContentTypes } from '@/constants/chat'

import service, { type ApiResponse } from './request'

export interface UploadMediaResult {
  id: number
  url: string
  key: string
  kind: string
  mimeType: string
  sizeBytes: number
}

export type UploadOutcome<T = UploadMediaResult> =
  { ok: true; media: T } | { ok: false; message: string | null }

export interface ChatMediaUpload {
  mediaId: number
  url: string
  contentType: number
}

const IMAGE_MAX_BYTES = 10 * 1024 * 1024
const VIDEO_MAX_BYTES = 50 * 1024 * 1024
const AUDIO_MAX_BYTES = 5 * 1024 * 1024
const FILE_MAX_BYTES = 20 * 1024 * 1024

function postUpload(path: string, file: File, timeout: number) {
  const form = new FormData()
  form.append('file', file)
  return service.post<unknown, ApiResponse<UploadMediaResult>>(path, form, { timeout })
}

async function runUpload(
  path: string,
  file: File,
  timeout: number,
  failMessage: string,
): Promise<UploadOutcome> {
  try {
    const res = await postUpload(path, file, timeout)
    const data = res.data
    if (!data?.id || !data.url?.trim()) return { ok: false, message: failMessage }
    return { ok: true, media: data }
  } catch {
    return { ok: false, message: null }
  }
}

function rejected(message: string): UploadOutcome<never> {
  return { ok: false, message }
}

function isVideoFile(file: File) {
  if (file.type.startsWith('video/')) return true
  return /\.(mp4|mov|webm|3gp|m4v)$/i.test(file.name || '')
}

export function uploadImageMedia(file: File, failMessage = '上传失败'): Promise<UploadOutcome> {
  return runUpload('/api/upload/image', file, 60000, failMessage)
}

export async function uploadAudioMedia(
  file: File,
  failMessage = '上传失败',
): Promise<UploadOutcome> {
  if (file.size > AUDIO_MAX_BYTES) return rejected('语音不能超过 5MB')
  return runUpload('/api/upload/audio', file, 60000, failMessage)
}

export async function uploadFileMedia(
  file: File,
  failMessage = '上传失败',
): Promise<UploadOutcome> {
  if (file.size > FILE_MAX_BYTES) return rejected('文件不能超过 20MB')
  return runUpload('/api/upload/file', file, 180000, failMessage)
}

export async function uploadChatMedia(file: File): Promise<UploadOutcome<ChatMediaUpload>> {
  if (isVideoFile(file)) {
    if (file.size > VIDEO_MAX_BYTES) return rejected('视频不能超过 50MB')
    const res = await runUpload('/api/upload/video', file, 180000, '视频上传失败')
    if (!res.ok) return res
    return {
      ok: true,
      media: { mediaId: res.media.id, url: res.media.url, contentType: MessageContentTypes.Video },
    }
  }
  const isImage = file.type.startsWith('image/') || /\.(jpe?g|png|gif|webp)$/i.test(file.name || '')
  if (!isImage) return rejected('请选择图片或视频')
  if (file.size > IMAGE_MAX_BYTES) return rejected('图片不能超过 10MB')
  const res = await uploadImageMedia(file, '图片上传失败')
  if (!res.ok) return res
  return {
    ok: true,
    media: { mediaId: res.media.id, url: res.media.url, contentType: MessageContentTypes.Image },
  }
}

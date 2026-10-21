import { uploadImFile, uploadImImage, uploadImVideo } from '@/api/im'
import { MessageContentType } from '@/constants/message'

export function newClientMsgId(prefix: string) {
  return `${prefix}-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
}

export async function uploadImMessageContent(contentType: number, file: File) {
  if (contentType === MessageContentType.Image) {
    const { data } = await uploadImImage(file)
    return String(data.id)
  }
  if (contentType === MessageContentType.Video) {
    const { data } = await uploadImVideo(file)
    return String(data.id)
  }
  const { data } = await uploadImFile(file)
  return JSON.stringify({ mediaId: data.id, name: file.name })
}

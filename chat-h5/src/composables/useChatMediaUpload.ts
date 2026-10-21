import { ref } from 'vue'
import { showToast } from 'vant'

import {
  uploadAudioMedia,
  uploadChatMedia,
  uploadFileMedia,
  type UploadOutcome,
} from '@/api/upload'
import { MessageContentTypes } from '@/constants/chat'
import { encodeChatVoice, VOICE_MAX_DURATION_MS } from '@/utils/chat/chatVoice'

export interface ChatMediaPayload {
  content: string
  contentType: number
}

function uploaded<T>(res: UploadOutcome<T>): T | null {
  if (res.ok) return res.media
  if (res.message) showToast(res.message)
  return null
}

export function useChatMediaUpload() {
  const uploadingMedia = ref(false)
  const uploadingVoice = ref(false)

  async function uploadMedia(file: File): Promise<ChatMediaPayload | null> {
    uploadingMedia.value = true
    try {
      const media = uploaded(await uploadChatMedia(file))
      if (!media) return null
      return { content: String(media.mediaId), contentType: media.contentType }
    } finally {
      uploadingMedia.value = false
    }
  }

  async function uploadFile(file: File): Promise<ChatMediaPayload | null> {
    uploadingMedia.value = true
    try {
      const media = uploaded(await uploadFileMedia(file, '文件上传失败'))
      if (!media) return null
      return {
        content: JSON.stringify({ mediaId: media.id, name: file.name }),
        contentType: MessageContentTypes.File,
      }
    } finally {
      uploadingMedia.value = false
    }
  }

  async function uploadVoice(
    blob: Blob,
    mime: string,
    durationMs: number,
  ): Promise<ChatMediaPayload | null> {
    uploadingVoice.value = true
    try {
      const ext = mime.includes('mp4') ? 'm4a' : mime.includes('ogg') ? 'ogg' : 'webm'
      const file = new File([blob], `voice.${ext}`, { type: blob.type || mime || 'audio/webm' })
      const media = uploaded(await uploadAudioMedia(file, '语音上传失败'))
      if (!media) return null
      return {
        content: encodeChatVoice(media.id, Math.min(durationMs, VOICE_MAX_DURATION_MS)),
        contentType: MessageContentTypes.Voice,
      }
    } finally {
      uploadingVoice.value = false
    }
  }

  return { uploadingMedia, uploadingVoice, uploadMedia, uploadFile, uploadVoice }
}

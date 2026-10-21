import { ref } from 'vue'
import { showToast, type UploaderFileListItem } from 'vant'

import { uploadImageMedia } from '@/api/upload'

export type VanUploaderMediaItem = UploaderFileListItem & { mediaId?: number }

async function uploadVanItem(fileItem: VanUploaderMediaItem): Promise<number | null> {
  if (!fileItem.file) return null
  const res = await uploadImageMedia(fileItem.file)
  if (!res.ok) {
    if (res.message) showToast(res.message)
    fileItem.status = 'failed'
    return null
  }
  fileItem.url = res.media.url.trim()
  fileItem.mediaId = res.media.id
  fileItem.status = 'done'
  return res.media.id
}

export function useVanImageUpload(options?: { onSuccess?: (mediaId: number) => void }) {
  const uploading = ref(false)

  async function onAfterRead(item: VanUploaderMediaItem | VanUploaderMediaItem[]) {
    const targets = (Array.isArray(item) ? item : [item]).filter((fileItem) => fileItem?.file)
    if (targets.length === 0) return

    uploading.value = true
    try {
      const ids = await Promise.all(targets.map(uploadVanItem))
      const uploaded = ids.filter((id): id is number => !!id)
      const mediaId = uploaded[uploaded.length - 1]
      if (mediaId) options?.onSuccess?.(mediaId)
    } finally {
      uploading.value = false
    }
  }

  return { uploading, onAfterRead }
}

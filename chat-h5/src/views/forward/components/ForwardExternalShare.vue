<script setup lang="ts">
import { ref } from 'vue'
import { closeToast, showFailToast, showLoadingToast, showSuccessToast } from 'vant'

import type { ForwardPayload } from '@/utils/chat/chatForwardNav'
import { buildPostShareCardBlob } from '@/utils/post/postShareCard'
import { copyText, saveBlobAsFile, shareImageFile } from '@/utils/share/shareCore'
import { showCatchError } from '@/utils/toast'

const props = defineProps<{
  payload: ForwardPayload
}>()

const emit = defineEmits<{
  shared: []
}>()

const canSystemShare = typeof navigator.share === 'function'
const shareBusy = ref(false)
let shareBlob: Blob | null = null

function shareFileName() {
  return props.payload.postId ? `post-${props.payload.postId}.jpg` : 'share.jpg'
}

async function ensureShareBlob(): Promise<Blob | null> {
  if (shareBlob) return shareBlob
  const p = props.payload
  if (!p.shareUrl) return null
  showLoadingToast({ message: '正在生成分享图…', forbidClick: true, duration: 0 })
  try {
    shareBlob = await buildPostShareCardBlob({
      title: p.title || '动态',
      url: p.shareUrl,
      content: p.shareContent,
      coverUrl: p.cover,
    })
    return shareBlob
  } catch (e: unknown) {
    showCatchError(e, '分享失败，请稍后重试')
    return null
  } finally {
    closeToast()
  }
}

async function onCopyLink() {
  const url = props.payload.shareUrl
  if (!url) return
  try {
    await copyText(url)
    showSuccessToast('链接已复制')
  } catch {
    showFailToast('复制失败')
  }
}

async function onSaveImage() {
  if (shareBusy.value) return
  shareBusy.value = true
  try {
    const blob = await ensureShareBlob()
    if (!blob) return
    saveBlobAsFile(blob, shareFileName())
    showSuccessToast('已开始保存')
  } finally {
    shareBusy.value = false
  }
}

async function onSystemShare() {
  if (shareBusy.value) return
  if (!canSystemShare) {
    await onCopyLink()
    return
  }
  shareBusy.value = true
  try {
    const blob = await ensureShareBlob()
    if (!blob) return
    const file = new File([blob], shareFileName(), { type: 'image/jpeg' })
    const title = props.payload.title || '聚点'
    const ok = await shareImageFile({
      file,
      title,
      text: `${title}\n${props.payload.shareUrl || ''}`,
    })
    if (ok) emit('shared')
    else showFailToast('当前环境不支持系统分享')
  } catch (e: unknown) {
    if (e instanceof Error && e.name === 'AbortError') return
    showFailToast('分享失败，请稍后重试')
  } finally {
    shareBusy.value = false
  }
}
</script>

<template>
  <div class="forward-external">
    <p class="forward-external__title">分享到站外</p>
    <div class="forward-external__row">
      <button type="button" class="forward-ext-btn" :disabled="shareBusy" @click="onSaveImage">
        <van-icon name="photo-o" size="20" />
        <span>保存图片</span>
      </button>
      <button type="button" class="forward-ext-btn" @click="onCopyLink">
        <van-icon name="link-o" size="20" />
        <span>复制链接</span>
      </button>
      <button type="button" class="forward-ext-btn" :disabled="shareBusy" @click="onSystemShare">
        <van-icon name="share-o" size="20" />
        <span>{{ canSystemShare ? '更多' : '复制分享' }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.forward-external {
  flex-shrink: 0;
  padding: 10px 16px calc(10px + env(safe-area-inset-bottom));
  border-top: 1px solid var(--border-color, rgba(0, 0, 0, 0.06));
  background: var(--card-bg, #fff);
}

.forward-external__title {
  margin: 0 0 8px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.forward-external__row {
  display: flex;
  gap: 8px;
}

.forward-ext-btn {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 10px 4px;
  border: none;
  border-radius: 12px;
  background: var(--page-bg);
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
}

.forward-ext-btn:active:not(:disabled) {
  background: rgba(0, 0, 0, 0.04);
}

.forward-ext-btn:disabled {
  opacity: 0.55;
}
</style>

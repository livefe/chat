<script setup lang="ts">
import { computed } from 'vue'
import { showConfirmDialog, showToast } from 'vant'

import { parseChatMedia } from '@/utils/chat/chatMedia'
import { chatFileDisplayName, truncateFileDisplayName } from '@/utils/chat/chatMessageContent'
import { resolveMediaUrl } from '@/utils/media'
import FileClipIcon from '../FileClipIcon.vue'
import MessageBubbleGestureArea from './MessageBubbleGestureArea.vue'
import type { MessageBubbleContentProps } from './types'

const props = defineProps<MessageBubbleContentProps>()

const fileSrc = computed(() => resolveMediaUrl(parseChatMedia(props.message.content)?.url || ''))
const fileName = computed(() => chatFileDisplayName(props.message.content))
const fileNameLabel = computed(() => truncateFileDisplayName(fileName.value))

async function onFileClick(e: Event) {
  e.preventDefault()
  if (props.gestures.consumeClickIfLongPressed(e)) return
  if (!fileSrc.value) {
    showToast('文件地址无效')
    return
  }
  try {
    await showConfirmDialog({
      title: '下载文件',
      message: `确定下载「${fileNameLabel.value}」？`,
      confirmButtonText: '下载',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  const a = document.createElement('a')
  a.href = fileSrc.value
  a.download = fileName.value || 'file'
  a.target = '_blank'
  a.rel = 'noopener'
  document.body.appendChild(a)
  a.click()
  a.remove()
}
</script>

<template>
  <MessageBubbleGestureArea
    tag="button"
    class="file-msg"
    :send-status="message.sendStatus"
    :gestures="gestures"
    @click="onFileClick"
  >
    <span class="file-msg__icon" aria-hidden="true">
      <FileClipIcon :size="22" />
    </span>
    <span class="file-msg__body">
      <span class="file-msg__name" :title="fileName">{{ fileNameLabel }}</span>
      <span class="file-msg__hint">点击查看</span>
    </span>
  </MessageBubbleGestureArea>
</template>

<style scoped>
.file-msg {
  display: flex;
  align-items: center;
  gap: 10px;
  width: max-content;
  max-width: min(240px, 70vw);
  min-width: 168px;
  overflow: hidden;
  padding: 10px 12px;
  border: none;
  border-radius: 4px;
  background: #fff;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  text-decoration: none;
  text-align: left;
  color: inherit;
  font: inherit;
  cursor: pointer;
  box-sizing: border-box;
}

.file-msg.is-failed {
  opacity: 1;
  outline: 1px solid rgba(229, 77, 66, 0.35);
}

.file-msg__icon {
  flex-shrink: 0;
  display: flex;
  color: var(--text-secondary);
  line-height: 1;
}

.file-msg__body {
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.file-msg__name {
  display: -webkit-box;
  overflow: hidden;
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  word-break: break-all;
  overflow-wrap: anywhere;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.file-msg__hint {
  font-size: 11px;
  color: var(--text-secondary);
}
</style>

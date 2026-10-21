<script setup lang="ts">
import { computed } from 'vue'
import { showImagePreview } from 'vant'

import { openChatLocationMap, parseChatLocation } from '@/utils/chat/chatLocation'
import {
  isLocationMessage,
  isVideoMessage,
  replyQuoteMediaUrl,
} from '@/utils/chat/chatMessageContent'
import { resolveMediaUrl } from '@/utils/media'

const props = defineProps<{
  fromName: string
  contentType?: number
  content?: string
  text: string
}>()

const emit = defineEmits<{
  previewVideo: [src: string]
}>()

const thumbSrc = computed(() =>
  resolveMediaUrl(replyQuoteMediaUrl(props.contentType, props.content)),
)
const thumbIsVideo = computed(() => isVideoMessage(props.contentType) && !!thumbSrc.value)
const location = computed(() =>
  isLocationMessage(props.contentType) ? parseChatLocation(props.content) : null,
)
const label = computed(() => location.value?.placeName ?? props.text)

function previewImage() {
  if (!thumbSrc.value || thumbIsVideo.value) return
  showImagePreview({ images: [thumbSrc.value], startPosition: 0 })
}

function previewVideo() {
  if (!thumbSrc.value || !thumbIsVideo.value) return
  emit('previewVideo', thumbSrc.value)
}

function openLocation() {
  if (!location.value) return
  openChatLocationMap(location.value)
}
</script>

<template>
  <span class="reply-preview">
    <span class="reply-preview__name">{{ fromName }}：</span>
    <button
      v-if="thumbSrc && !thumbIsVideo"
      type="button"
      class="reply-preview__thumb-btn"
      aria-label="查看图片"
      @click="previewImage"
    >
      <img class="reply-preview__thumb" :src="thumbSrc" alt="图片" />
    </button>
    <button
      v-else-if="thumbSrc"
      type="button"
      class="reply-preview__thumb-btn reply-preview__video"
      aria-label="播放视频"
      @click="previewVideo"
    >
      <video class="reply-preview__thumb" :src="thumbSrc" muted playsinline preload="metadata" />
      <van-icon name="play-circle-o" class="reply-preview__play" size="14" />
    </button>
    <button v-else-if="location" type="button" class="reply-preview__loc" @click="openLocation">
      {{ label }}
    </button>
    <span v-else class="reply-preview__text">{{ label }}</span>
  </span>
</template>

<style scoped>
.reply-preview {
  display: flex;
  align-items: center;
  gap: 4px;
  min-width: 0;
}

.reply-preview__name {
  flex-shrink: 0;
}

.reply-preview__text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reply-preview__loc {
  min-width: 0;
  max-width: 100%;
  padding: 0;
  border: none;
  background: transparent;
  font: inherit;
  color: inherit;
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}

.reply-preview__thumb-btn {
  flex-shrink: 0;
  padding: 0;
  border: none;
  background: transparent;
  line-height: 0;
  cursor: pointer;
}

.reply-preview__thumb {
  display: block;
  width: 28px;
  height: 28px;
  border-radius: 3px;
  object-fit: cover;
  background: var(--page-bg);
  pointer-events: none;
}

.reply-preview__video {
  position: relative;
  width: 28px;
  height: 28px;
}

.reply-preview__play {
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  color: #fff;
  pointer-events: none;
}
</style>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { parseChatMedia } from '@/utils/chat/chatMedia'
import {
  fitChatVideoEl,
  formatVideoDuration,
  unlockVideoCoverFrame,
} from '@/utils/chat/chatVideoCover'
import { resolveMediaUrl } from '@/utils/media'
import MessageBubbleGestureArea from './MessageBubbleGestureArea.vue'
import type { MessageBubbleContentProps } from './types'

const props = defineProps<MessageBubbleContentProps>()

const emit = defineEmits<{
  preview: [src: string]
}>()

const videoSrc = computed(() => resolveMediaUrl(parseChatMedia(props.message.content)?.url || ''))
const videoDurationText = ref('')
const videoCoverReady = ref(false)
const videoPosterUrl = ref('')
const videoFrameStyle = ref<Record<string, string>>({})
let coverToken = 0

watch(videoSrc, () => {
  videoDurationText.value = ''
  videoCoverReady.value = false
  videoPosterUrl.value = ''
  videoFrameStyle.value = {}
  coverToken += 1
})

function applyVideoFrameSize(el: HTMLVideoElement) {
  fitChatVideoEl(el)
  if (el.style.width && el.style.height) {
    videoFrameStyle.value = {
      width: el.style.width,
      height: el.style.height,
    }
  }
}

async function prepareVideoCover(e: Event) {
  const el = e.target as HTMLVideoElement
  if (!el || el.dataset.coverBusy === '1' || el.dataset.coverReady === '1') return
  el.dataset.coverBusy = '1'
  const token = coverToken
  if (Number.isFinite(el.duration) && el.duration > 0) {
    videoDurationText.value = formatVideoDuration(el.duration)
  }
  applyVideoFrameSize(el)
  try {
    const poster = await unlockVideoCoverFrame(el)
    if (token !== coverToken) return
    if (poster) videoPosterUrl.value = poster
    if (Number.isFinite(el.duration) && el.duration > 0) {
      videoDurationText.value = formatVideoDuration(el.duration)
    }
    applyVideoFrameSize(el)
  } finally {
    if (token === coverToken) {
      el.dataset.coverReady = '1'
      el.dataset.coverBusy = '0'
      videoCoverReady.value = true
    }
  }
}

function previewVideo(e?: Event) {
  if (props.gestures.consumeClickIfLongPressed(e)) return
  if (!videoSrc.value) return
  emit('preview', videoSrc.value)
}

function onVideoRelease(e: TouchEvent) {
  e.preventDefault()
  if (props.gestures.consumeClickIfLongPressed()) return
  previewVideo()
}
</script>

<template>
  <MessageBubbleGestureArea
    class="video-msg"
    :class="{ 'video-msg--ready': videoCoverReady }"
    role="button"
    tabindex="0"
    aria-label="播放视频"
    :send-status="message.sendStatus"
    :gestures="gestures"
    @click="previewVideo"
    @release="onVideoRelease"
  >
    <img
      v-if="videoPosterUrl"
      class="video video--poster"
      :src="videoPosterUrl"
      :style="videoFrameStyle"
      alt=""
      draggable="false"
    />
    <video
      v-show="!videoPosterUrl"
      :key="videoSrc"
      class="video"
      :src="videoSrc"
      :style="videoFrameStyle"
      muted
      defaultMuted
      playsinline
      webkit-playsinline
      x5-playsinline
      x5-video-player-type="h5"
      preload="auto"
      @loadedmetadata="prepareVideoCover"
      @loadeddata="prepareVideoCover"
      @canplay="prepareVideoCover"
    />
    <span class="video-play" aria-hidden="true">
      <van-icon name="play-circle-o" size="44" />
    </span>
    <span v-if="videoDurationText" class="video-duration">{{ videoDurationText }}</span>
  </MessageBubbleGestureArea>
</template>

<style scoped>
.video-msg {
  position: relative;
  display: inline-block;
  max-width: 100%;
  line-height: 0;
  cursor: pointer;
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  user-select: none;
}

.video {
  display: block;
  width: auto;
  height: auto;
  max-width: min(220px, 60vw);
  max-height: 220px;
  min-width: 140px;
  min-height: 100px;
  border-radius: 4px;
  object-fit: contain;
  background: #1a1a1a;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.15s ease;
}

.video--poster {
  opacity: 1;
}

.video-msg--ready .video {
  opacity: 1;
}

.video-play {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.92);
  pointer-events: none;
  text-shadow: 0 1px 8px rgba(0, 0, 0, 0.35);
}

.video-duration {
  position: absolute;
  right: 8px;
  bottom: 8px;
  color: #fff;
  font-size: 11px;
  line-height: 1.2;
  pointer-events: none;
  text-shadow: 0 1px 4px rgba(0, 0, 0, 0.45);
}
</style>

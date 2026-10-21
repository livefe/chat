<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { showToast } from 'vant'

import { formatVoiceDuration, type ChatVoice } from '@/utils/chat/chatVoice'
import {
  isVoicePlaying,
  stopVoicePlayback,
  subscribeVoicePlayer,
  toggleVoicePlayback,
} from '@/utils/chat/chatVoicePlayer'
import { resolveMediaUrl } from '@/utils/media'
import MessageBubbleGestureArea from './MessageBubbleGestureArea.vue'
import type { MessageBubbleContentProps } from './types'

const props = defineProps<
  MessageBubbleContentProps & {
    voice: ChatVoice
    self: boolean
  }
>()

const voiceSrc = computed(() => resolveMediaUrl(props.voice.url))
const voiceKey = computed(
  () => `msg-${props.message.id || props.message.clientMsgId || voiceSrc.value}`,
)
const voicePlaying = ref(false)

function syncVoicePlaying() {
  voicePlaying.value = isVoicePlaying(voiceKey.value)
}

onMounted(() => {
  syncVoicePlaying()
})
const unsubVoice = subscribeVoicePlayer(syncVoicePlaying)
onBeforeUnmount(() => {
  unsubVoice()
  if (isVoicePlaying(voiceKey.value)) stopVoicePlayback()
})

async function onVoiceClick() {
  if (props.gestures.consumeClickIfLongPressed()) return
  if (!voiceSrc.value) return
  try {
    await toggleVoicePlayback(voiceKey.value, voiceSrc.value)
  } catch {
    showToast('播放失败')
  }
}
</script>

<template>
  <MessageBubbleGestureArea
    tag="button"
    class="voice-msg"
    :class="{
      'voice-msg--self': self,
      'voice-msg--playing': voicePlaying,
    }"
    :style="{ width: `${Math.min(180, 72 + Math.round(voice.durationMs / 1000) * 4)}px` }"
    aria-label="播放语音"
    :send-status="message.sendStatus"
    :gestures="gestures"
    @click="onVoiceClick"
  >
    <span class="voice-msg__wave" aria-hidden="true"> <i /><i /><i /> </span>
    <span class="voice-msg__dur">{{ formatVoiceDuration(voice.durationMs) }}</span>
  </MessageBubbleGestureArea>
</template>

<style scoped>
.voice-msg {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-width: 72px;
  max-width: 180px;
  height: 40px;
  padding: 0 12px;
  border: none;
  border-radius: 6px;
  background: #fff;
  color: var(--text-primary);
  box-shadow: 0 1px 1px rgba(0, 0, 0, 0.04);
  cursor: pointer;
  -webkit-tap-highlight-color: transparent;
  touch-action: manipulation;
}

.voice-msg--self {
  background: var(--brand-color);
  color: #fff;
  flex-direction: row-reverse;
}

.voice-msg__dur {
  font-size: 14px;
  font-variant-numeric: tabular-nums;
  line-height: 1;
}

.voice-msg__wave {
  display: inline-flex;
  align-items: flex-end;
  gap: 3px;
  height: 14px;
}

.voice-msg__wave i {
  display: block;
  width: 3px;
  border-radius: 1px;
  background: currentColor;
  opacity: 0.85;
}

.voice-msg__wave i:nth-child(1) {
  height: 6px;
}
.voice-msg__wave i:nth-child(2) {
  height: 12px;
}
.voice-msg__wave i:nth-child(3) {
  height: 8px;
}

.voice-msg--playing .voice-msg__wave i {
  animation: voice-wave 0.9s ease-in-out infinite;
}
.voice-msg--playing .voice-msg__wave i:nth-child(2) {
  animation-delay: 0.15s;
}
.voice-msg--playing .voice-msg__wave i:nth-child(3) {
  animation-delay: 0.3s;
}

@keyframes voice-wave {
  0%,
  100% {
    transform: scaleY(0.55);
  }
  50% {
    transform: scaleY(1);
  }
}
</style>

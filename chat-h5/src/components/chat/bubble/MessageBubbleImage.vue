<script setup lang="ts">
import { computed } from 'vue'
import { showImagePreview } from 'vant'

import { parseChatMedia } from '@/utils/chat/chatMedia'
import { resolveMediaUrl } from '@/utils/media'
import MessageBubbleGestureArea from './MessageBubbleGestureArea.vue'
import type { MessageBubbleContentProps } from './types'

const props = defineProps<MessageBubbleContentProps>()

const imageSrc = computed(() => resolveMediaUrl(parseChatMedia(props.message.content)?.url || ''))

function previewImage(e?: Event) {
  if (props.gestures.consumeClickIfLongPressed(e)) return
  if (!imageSrc.value) return
  showImagePreview({ images: [imageSrc.value], startPosition: 0 })
}
</script>

<template>
  <MessageBubbleGestureArea
    tag="button"
    class="image-msg"
    :send-status="message.sendStatus"
    :gestures="gestures"
    @click="previewImage"
  >
    <img class="image" :src="imageSrc" alt="图片" draggable="false" />
  </MessageBubbleGestureArea>
</template>

<style scoped>
.image-msg {
  display: block;
  padding: 0;
  border: none;
  background: transparent;
  cursor: pointer;
  line-height: 0;
  max-width: 100%;
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  user-select: none;
}

.image {
  display: block;
  width: auto;
  height: auto;
  max-width: min(220px, 60vw);
  max-height: 220px;
  border-radius: 4px;
  object-fit: contain;
  background: var(--page-bg);
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  user-select: none;
  pointer-events: none;
}
</style>

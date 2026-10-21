<script setup lang="ts">
import type { MessageBubbleGestures } from '@/composables/useMessageBubbleMenu'
import { openChatLocationMap, type ChatLocationContent } from '@/utils/chat/chatLocation'

const props = defineProps<{
  location: ChatLocationContent
  gestures: MessageBubbleGestures
}>()

function openLocation(e?: Event) {
  if (props.gestures.consumeClickIfLongPressed(e)) return
  openChatLocationMap(props.location)
}
</script>

<template>
  <button type="button" class="location-btn" @click="openLocation">
    <van-icon name="location-o" size="20" class="location-btn__icon" />
    <span class="location-btn__body">
      <span class="location-btn__name">{{ location.placeName }}</span>
      <span class="location-btn__hint">点击查看地图</span>
    </span>
  </button>
</template>

<style scoped>
.location-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 168px;
  max-width: min(220px, 60vw);
  padding: 10px 12px;
  border: none;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.location-btn__icon {
  flex-shrink: 0;
  color: var(--brand-color);
}

.location-btn__body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.location-btn__name {
  font-size: 16px;
  font-weight: 500;
  line-height: 1.35;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.location-btn__hint {
  font-size: 11px;
  line-height: 1.2;
  color: var(--text-secondary);
}
</style>

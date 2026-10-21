<template>
  <el-image
    v-if="item.contentType === MessageContentType.Image && mediaUrl"
    :src="mediaUrl"
    :preview-src-list="[mediaUrl]"
    fit="cover"
    class="msg-thumb"
    :class="mutedClass"
  />
  <audio
    v-else-if="item.contentType === MessageContentType.Voice && mediaUrl"
    :src="mediaUrl"
    class="msg-audio"
    :class="mutedClass"
    controls
    preload="metadata"
  />
  <video
    v-else-if="item.contentType === MessageContentType.Video && mediaUrl"
    :src="mediaUrl"
    class="msg-video"
    :class="mutedClass"
    controls
    preload="metadata"
  />
  <span v-else :class="mutedClass">{{ text }}</span>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { AdminMessageItem } from '@/api/message'
import { MessageContentType } from '@/constants/message'
import { formatMessageSummary, parseMessageMediaUrl } from '@/utils/messageDisplay'

const props = defineProps<{
  item: Pick<AdminMessageItem, 'contentType' | 'content'>
  muted?: boolean
}>()

const mutedClass = computed(() => ({ 'muted-content': props.muted }))
const mediaUrl = computed(() => {
  const { contentType, content } = props.item
  if (
    contentType === MessageContentType.Image ||
    contentType === MessageContentType.Video ||
    contentType === MessageContentType.Voice
  ) {
    return parseMessageMediaUrl(content)
  }
  return ''
})
const text = computed(() => formatMessageSummary(props.item.contentType, props.item.content))
</script>

<style scoped>
.muted-content {
  color: var(--el-text-color-secondary);
  text-decoration: line-through;
}
.msg-thumb {
  width: 72px;
  height: 72px;
  border-radius: 6px;
}
.msg-video {
  width: 220px;
  max-width: 100%;
  max-height: 140px;
  border-radius: 6px;
  background: #000;
}
.msg-audio {
  width: 220px;
  max-width: 100%;
  height: 36px;
}
</style>

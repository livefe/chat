<script setup lang="ts">
import { computed } from 'vue'

import type { MessageItem } from '@/api/message'
import { isVoiceMessage } from '@/utils/chat/chatMessageContent'
import { voiceContentLabel } from '@/utils/chat/chatVoice'
import ReplyQuotePreview from '../ReplyQuotePreview.vue'

const props = defineProps<{
  message: MessageItem
}>()

const emit = defineEmits<{
  previewVideo: [src: string]
}>()

const replyText = computed(() => {
  if (isVoiceMessage(props.message.replyToContentType)) {
    return voiceContentLabel(props.message.replyToContent)
  }
  return props.message.replyToContent || '原消息已删除'
})
</script>

<template>
  <div class="reply-quote">
    <ReplyQuotePreview
      :from-name="message.replyToFromName || '消息'"
      :content-type="message.replyToContentType"
      :content="message.replyToContent"
      :text="replyText"
      @preview-video="emit('previewVideo', $event)"
    />
  </div>
</template>

<style scoped>
.reply-quote {
  display: flex;
  align-items: center;
  box-sizing: border-box;
  padding: 5px 8px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.05);
  font-size: 11px;
  line-height: 1.35;
  color: var(--text-muted);
  min-width: 0;
}
</style>

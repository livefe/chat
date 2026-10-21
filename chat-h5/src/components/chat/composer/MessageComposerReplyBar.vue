<script setup lang="ts">
import type { MessageReplyQuote } from '@/api/message'
import ReplyQuotePreview from '../ReplyQuotePreview.vue'

defineProps<{
  reply: MessageReplyQuote
}>()

const emit = defineEmits<{
  close: []
  previewVideo: [src: string]
}>()
</script>

<template>
  <div class="reply-bar">
    <ReplyQuotePreview
      class="reply-bar__body"
      :from-name="reply.replyToFromName"
      :content-type="reply.replyToContentType"
      :content="reply.replyToContent"
      :text="reply.replyToContent"
      @preview-video="emit('previewVideo', $event)"
    />
    <button type="button" class="reply-bar__close" aria-label="取消引用" @click="emit('close')">
      <van-icon name="cross" size="14" />
    </button>
  </div>
</template>

<style scoped>
.reply-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0 12px 6px;
  padding: 6px 8px;
  border-radius: 4px;
  background: rgba(0, 0, 0, 0.05);
}

.reply-bar__body {
  flex: 1;
  font-size: 12px;
  line-height: 1.35;
  color: var(--text-muted);
}

.reply-bar__close {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: var(--text-tertiary);
}
</style>

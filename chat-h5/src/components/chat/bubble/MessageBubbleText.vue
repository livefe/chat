<script setup lang="ts">
import MessageBubbleGestureArea from './MessageBubbleGestureArea.vue'
import type { MessageBubbleContentProps } from './types'

defineProps<
  MessageBubbleContentProps & {
    self: boolean
    card?: 'location' | 'post'
  }
>()
</script>

<template>
  <MessageBubbleGestureArea
    class="bubble"
    :class="{
      'bubble--self': self,
      'bubble--location': card === 'location',
      'bubble--post': card === 'post',
    }"
    :send-status="message.sendStatus"
    :gestures="gestures"
  >
    <slot>
      <div class="content">{{ message.content }}</div>
    </slot>
    <div v-if="message.sendStatus === 'failed'" class="send-hint">
      {{ message.sendFailReason || '发送失败' }}
    </div>
  </MessageBubbleGestureArea>
</template>

<style scoped>
.bubble {
  position: relative;
  display: inline-block;
  width: auto;
  max-width: 100%;
  background: #ffffff;
  border-radius: 4px;
  padding: 6px 10px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  box-sizing: border-box;

  --bubble-tail-top: 12px;
}

.bubble::after {
  content: '';
  position: absolute;
  top: var(--bubble-tail-top);
  width: 0;
  height: 0;
  border-style: solid;
  pointer-events: none;
}

.bubble:not(.bubble--self)::after {
  left: -5px;
  border-width: 5px 6px 5px 0;
  border-color: transparent #ffffff transparent transparent;
}

.bubble--self {
  background: var(--brand-color);
  color: #fff;
  border-radius: 4px;
  box-shadow: none;
}

.bubble--self::after {
  right: -5px;
  left: auto;
  border-width: 5px 0 5px 6px;
  border-color: transparent transparent transparent var(--brand-color);
}

.bubble--location,
.bubble--post {
  padding: 0;
  background: #ffffff;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
  overflow: visible;
}

.bubble--location.bubble--self,
.bubble--post.bubble--self {
  background: #ffffff;
  color: inherit;
}

.bubble--location.bubble--self::after,
.bubble--post.bubble--self::after {
  border-left-color: #ffffff;
}

.bubble.is-failed.bubble--self:not(.bubble--location):not(.bubble--post)::after {
  border-left-color: #888888;
}

.content {
  font-size: 16px;
  line-height: 1.4;
  white-space: pre-wrap;
  overflow-wrap: break-word;
  word-break: normal;
}

.bubble.is-failed.bubble--self:not(.bubble--location):not(.bubble--post) {
  background: #888888;
  box-shadow: none;
}

.send-hint {
  margin-top: 4px;
  padding: 0 4px 2px;
  font-size: 11px;
  opacity: 0.9;
}

.bubble--location .send-hint,
.bubble--post .send-hint {
  color: var(--text-secondary);
}
</style>

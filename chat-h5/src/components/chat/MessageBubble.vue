<script setup lang="ts">
import { computed, ref, type ComponentPublicInstance } from 'vue'

import UserAvatar from '@/components/UserAvatar.vue'
import ChatVideoPreview from './ChatVideoPreview.vue'
import MessageBubbleActionMenu from './bubble/MessageBubbleActionMenu.vue'
import MessageBubbleFile from './bubble/MessageBubbleFile.vue'
import MessageBubbleImage from './bubble/MessageBubbleImage.vue'
import MessageBubbleLocation from './bubble/MessageBubbleLocation.vue'
import MessageBubblePost from './bubble/MessageBubblePost.vue'
import MessageBubbleReplyQuote from './bubble/MessageBubbleReplyQuote.vue'
import MessageBubbleText from './bubble/MessageBubbleText.vue'
import MessageBubbleVideo from './bubble/MessageBubbleVideo.vue'
import MessageBubbleVoice from './bubble/MessageBubbleVoice.vue'
import type { MessageItem } from '@/api/message'
import {
  useMessageBubbleMenu,
  type MessageBubbleGestures,
} from '@/composables/useMessageBubbleMenu'
import { MessageScope } from '@/constants/chat'
import { parseChatLocation } from '@/utils/chat/chatLocation'
import {
  isFileMessage,
  isImageMessage,
  isLocationMessage,
  isPostMessage,
  isVideoMessage,
  isVoiceMessage,
} from '@/utils/chat/chatMessageContent'
import { parseChatPost } from '@/utils/chat/chatPost'
import { parseChatVoice } from '@/utils/chat/chatVoice'

const props = withDefaults(
  defineProps<{
    message: MessageItem
    self: boolean
    hideName?: boolean
    enableActions?: boolean
  }>(),
  { hideName: false, enableActions: true },
)

const emit = defineEmits<{
  avatarClick: [msg: MessageItem]
  reply: [msg: MessageItem]
  forward: [msg: MessageItem]
  report: [msg: MessageItem]
  recall: [msg: MessageItem]
}>()

const {
  menuOpen,
  menuStyle,
  stackRef,
  closeMenu,
  onTouchStart,
  onTouchEnd,
  onTouchMove,
  consumeClickIfLongPressed,
  onCopy,
  onQuote,
  onForward,
  onReport,
  onRecall,
  canRecall,
} = useMessageBubbleMenu(
  () => props.message,
  () => props.self,
  {
    reply: (msg) => emit('reply', msg),
    forward: (msg) => emit('forward', msg),
    report: (msg) => emit('report', msg),
    recall: (msg) => emit('recall', msg),
  },
)

const gestures = computed<MessageBubbleGestures>(() => ({
  enabled: props.enableActions,
  onTouchStart,
  onTouchEnd,
  onTouchMove,
  consumeClickIfLongPressed,
}))

function setStackRef(el: Element | ComponentPublicInstance | null) {
  stackRef.value = el && '$el' in el ? (el.$el as HTMLElement) : (el as HTMLElement | null)
}

const showSenderName = computed(() => !props.hideName && !props.self)
const recallHint = computed(() => {
  if (props.self) return '你撤回了一条消息'
  if (props.hideName) return '对方撤回了一条消息'
  const name = props.message.fromName?.trim()
  return name ? `${name}撤回了一条消息` : '撤回了一条消息'
})
const isSystemHint = computed(() => props.message.scope === MessageScope.SystemHint)
const hasReply = computed(() => (props.message.replyToId ?? 0) > 0)
const image = computed(() => isImageMessage(props.message.contentType))
const voice = computed(() =>
  isVoiceMessage(props.message.contentType) ? parseChatVoice(props.message.content) : null,
)
const video = computed(() => isVideoMessage(props.message.contentType))
const file = computed(() => isFileMessage(props.message.contentType))
const location = computed(() =>
  isLocationMessage(props.message.contentType) ? parseChatLocation(props.message.content) : null,
)
const post = computed(() =>
  isPostMessage(props.message.contentType) ? parseChatPost(props.message.content) : null,
)
const plainFailHint = computed(
  () =>
    props.message.sendStatus === 'failed' &&
    (image.value || !!voice.value || video.value || file.value),
)

const videoPreview = ref('')
</script>

<template>
  <div v-if="isSystemHint" class="recall-hint">{{ message.content }}</div>
  <div v-else-if="message.recalled" class="recall-hint">{{ recallHint }}</div>
  <div v-else class="message" :class="{ self, 'message--private': hideName }">
    <MessageBubbleReplyQuote
      v-if="hasReply"
      class="reply-quote"
      :message="message"
      @preview-video="videoPreview = $event"
    />
    <div class="body">
      <UserAvatar
        class="avatar"
        :src="message.fromAvatar"
        :size="40"
        tone="chat"
        @click="emit('avatarClick', message)"
      />
      <div class="stack">
        <div v-if="showSenderName" class="sender-name">{{ message.fromName }}</div>

        <MessageBubbleImage
          v-if="image"
          :ref="setStackRef"
          :message="message"
          :gestures="gestures"
        />
        <MessageBubbleVoice
          v-else-if="voice"
          :ref="setStackRef"
          :message="message"
          :voice="voice"
          :self="self"
          :gestures="gestures"
        />
        <MessageBubbleVideo
          v-else-if="video"
          :ref="setStackRef"
          :message="message"
          :gestures="gestures"
          @preview="videoPreview = $event"
        />
        <MessageBubbleFile
          v-else-if="file"
          :ref="setStackRef"
          :message="message"
          :gestures="gestures"
        />
        <MessageBubbleText
          v-else
          :ref="setStackRef"
          :message="message"
          :self="self"
          :card="post ? 'post' : location ? 'location' : undefined"
          :gestures="gestures"
        >
          <MessageBubblePost v-if="post" :post="post" :gestures="gestures" />
          <MessageBubbleLocation v-else-if="location" :location="location" :gestures="gestures" />
        </MessageBubbleText>
        <div v-if="plainFailHint" class="send-hint">{{ message.sendFailReason || '发送失败' }}</div>
      </div>
    </div>
    <MessageBubbleActionMenu
      v-if="enableActions"
      :open="menuOpen"
      :menu-style="menuStyle"
      :self="self"
      :can-recall="canRecall"
      @close="closeMenu"
      @copy="onCopy"
      @forward="onForward"
      @quote="onQuote"
      @recall="onRecall"
      @report="onReport"
    />
    <ChatVideoPreview v-if="videoPreview" :src="videoPreview" @close="videoPreview = ''" />
  </div>
</template>

<style scoped>
.recall-hint {
  align-self: center;
  width: 100%;
  padding: 6px 16px;
  font-size: 12px;
  line-height: 1.4;
  color: var(--text-muted);
  text-align: center;
}

.message {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  width: 100%;
  -webkit-user-select: none;
  user-select: none;
  -webkit-touch-callout: none;
}

.message.self {
  align-items: flex-end;
}

.body {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  align-self: stretch;
  width: 100%;
  max-width: 100%;
  box-sizing: border-box;
}

.message.self .body {
  flex-direction: row-reverse;
}

.message--private .body {
  gap: 10px;
}

.avatar {
  flex-shrink: 0;
}

.stack {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  max-width: 72%;
  min-width: 0;
}

.message.self .stack {
  align-items: flex-end;
}

.sender-name {
  padding: 0 2px;
  font-size: 11px;
  font-weight: 500;
  line-height: 1.2;
  color: var(--text-secondary);
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.reply-quote {
  margin-left: calc(40px + 8px);
  max-width: min(72%, calc(100% - 48px));
}

.message--private .reply-quote {
  margin-left: calc(40px + 10px);
}

.message.self .reply-quote {
  margin-left: 0;
  margin-right: calc(40px + 8px);
}

.message--private.self .reply-quote {
  margin-right: calc(40px + 10px);
}

.send-hint {
  margin-top: 4px;
  padding: 0 4px 2px;
  font-size: 11px;
  opacity: 0.9;
  color: var(--text-secondary);
}
</style>

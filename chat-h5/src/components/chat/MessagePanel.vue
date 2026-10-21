<script setup lang="ts">
import { nextTick, onActivated, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import type { MessageItem, MessageReplyQuote } from '@/api/message'
import { useLocationStore, type LocationShareTarget } from '@/stores/location'
import { MessageContentTypes } from '@/constants/chat'
import { replyQuoteText } from '@/utils/chat/chatMessageContent'
import { encodeChatLocation } from '@/utils/chat/chatLocation'
import { locationPickRoute } from '@/utils/map/locationPick'
import MessageComposer from './MessageComposer.vue'
import MessageList from './MessageList.vue'
import { openForwardMessage } from '@/utils/chat/chatForwardNav'

const SHARE_STICK_LOCK_MS = 2500
const INITIAL_STICK_LOCK_MS = 2000

const props = withDefaults(
  defineProps<{
    messages: MessageItem[]
    showBack?: boolean
    backLabel?: string

    initialLoading?: boolean
    loadingMore?: boolean
    hasMore?: boolean

    hideSenderName?: boolean

    emptyText?: string

    canCompose?: boolean
    enableActions?: boolean
    enableVoice?: boolean
    enableLocation?: boolean
    enableFile?: boolean
    placeholder?: string

    locationShare?: LocationShareTarget
  }>(),
  {
    backLabel: '返回',
    initialLoading: false,
    loadingMore: false,
    hasMore: false,
    hideSenderName: false,
    emptyText: '暂无消息',
    canCompose: true,
    enableActions: true,
    enableVoice: true,
    enableLocation: true,
    enableFile: true,
    placeholder: '说点什么…',
  },
)

const emit = defineEmits<{
  send: [content: string, contentType?: number, reply?: MessageReplyQuote | null]
  back: []
  avatarClick: [msg: MessageItem]
  messageLongPress: [msg: MessageItem]
  recall: [msg: MessageItem]
  loadMore: []
}>()

const locationStore = useLocationStore()
const router = useRouter()

const listRef = ref<InstanceType<typeof MessageList> | null>(null)
const composerRef = ref<InstanceType<typeof MessageComposer> | null>(null)
const moreOpen = ref(false)
const replyTo = ref<MessageReplyQuote | null>(null)

function takeReply(): MessageReplyQuote | null {
  const q = replyTo.value
  replyTo.value = null
  return q
}

function trySendShareLocation(): boolean {
  if (props.initialLoading) return false
  const pending = locationStore.consumeSharePending()
  if (!pending) return false

  emit('send', encodeChatLocation(pending.loc), MessageContentTypes.Location, takeReply())

  listRef.value?.scrollToBottom({ lockMs: SHARE_STICK_LOCK_MS })
  return true
}

function pickLocation() {
  locationStore.beginShare(props.locationShare ?? { type: 'nearby' })
  router.push(locationPickRoute('share'))
}

onMounted(() => {
  void nextTick(() => {
    trySendShareLocation()
  })
})

onActivated(() => {
  void nextTick(() => {
    const list = listRef.value
    const sentShare = trySendShareLocation()
    list?.syncAfterActivate(
      sentShare
        ? {
            forceStick: true,
            lockMs: Math.max(SHARE_STICK_LOCK_MS, list.stickLockUntil() - Date.now()),
          }
        : undefined,
    )
  })
})

watch(
  () => props.messages.find((m) => m.id === replyTo.value?.replyToId)?.recalled,
  (recalled) => {
    if (recalled) replyTo.value = null
  },
)

watch(
  () => props.initialLoading,
  (loading, wasLoading) => {
    if (loading) return
    const sentShare = trySendShareLocation()

    if (sentShare || wasLoading || props.messages.length > 0) {
      listRef.value?.scrollToBottom({
        lockMs: sentShare ? SHARE_STICK_LOCK_MS : INITIAL_STICK_LOCK_MS,
      })
    }
  },
)

function quoteFromMessage(msg: MessageItem): MessageReplyQuote {
  return {
    replyToId: msg.id,
    replyToFromName: msg.fromName || '消息',
    replyToContentType: msg.contentType,
    replyToContent: replyQuoteText(msg.contentType, msg.content) || '原消息已删除',
  }
}

function onBubbleReply(msg: MessageItem) {
  if (!props.enableActions) return
  if (!msg.id || msg.recalled) return
  replyTo.value = quoteFromMessage(msg)
  moreOpen.value = false
  composerRef.value?.focus()
}

function onBubbleForward(msg: MessageItem) {
  if (!props.enableActions) return
  if (!msg.id || msg.recalled || !msg.content?.trim()) return
  moreOpen.value = false
  openForwardMessage(router, msg)
}

function onBubbleRecall(msg: MessageItem) {
  if (!props.enableActions) return
  if (!msg.id || msg.recalled) return
  emit('recall', msg)
}

function onDismiss() {
  composerRef.value?.dismiss()
}

function onComposerStick(force: boolean) {
  if (force) listRef.value?.scrollToBottom()
  else listRef.value?.requestStick()
}

function onComposerSend(content: string, contentType?: number, reply?: MessageReplyQuote | null) {
  emit('send', content, contentType, reply)
}
</script>

<template>
  <div class="msg-panel">
    <div v-if="showBack" class="toolbar">
      <button class="back-btn" type="button" @click="emit('back')">
        <van-icon name="arrow-left" size="22" color="var(--text-primary)" />
        <span>{{ backLabel }}</span>
      </button>
    </div>

    <div class="msg-panel__main">
      <MessageList
        ref="listRef"
        :messages="messages"
        :initial-loading="initialLoading"
        :loading-more="loadingMore"
        :has-more="hasMore"
        :hide-sender-name="hideSenderName"
        :empty-text="emptyText"
        :enable-actions="enableActions"
        @avatar-click="emit('avatarClick', $event)"
        @reply="onBubbleReply"
        @forward="onBubbleForward"
        @report="emit('messageLongPress', $event)"
        @recall="onBubbleRecall"
        @load-more="emit('loadMore')"
        @dismiss="onDismiss"
      />
      <div
        v-if="moreOpen"
        class="msg-panel__dismiss"
        aria-hidden="true"
        @pointerdown.prevent="moreOpen = false"
      />
    </div>

    <MessageComposer
      ref="composerRef"
      v-model:more-open="moreOpen"
      v-model:reply-to="replyTo"
      :can-compose="canCompose"
      :enable-voice="enableVoice"
      :enable-location="enableLocation"
      :enable-file="enableFile"
      :placeholder="placeholder"
      @send="onComposerSend"
      @pick-location="pickLocation"
      @stick="onComposerStick"
    />
  </div>
</template>

<style scoped>
.msg-panel {
  position: relative;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  height: 100%;
  border: none;
  border-radius: 0;
  background: var(--page-bg);
  overflow: hidden;
}

.msg-panel__main {
  position: relative;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.msg-panel__dismiss {
  position: absolute;
  inset: 0;
  z-index: 5;
}

.msg-panel :deep(.chat-composer-wrap) {
  position: relative;
  z-index: 6;
}

.toolbar {
  position: relative;
  z-index: 6;
  display: flex;
  align-items: center;
  min-height: var(--app-navbar-height);
  padding: 0 12px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  min-height: 44px;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: var(--font-page-title);
  font-weight: 600;
  padding: 0;
}
</style>

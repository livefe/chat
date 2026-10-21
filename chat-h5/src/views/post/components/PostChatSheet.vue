<script setup lang="ts">
import { useRouter } from 'vue-router'

import MessagePanel from '@/components/chat/MessagePanel.vue'
import type { MessageItem, MessageReplyQuote } from '@/api/message'
import { openMessageReportSheet } from '@/composables/useReportAction'
import { openUserProfile } from '@/utils/user'

const props = defineProps<{
  show: boolean
  title?: string
  postId?: number
  onlineLabel?: string
  messages: MessageItem[]
  loadingMore?: boolean
  hasMore?: boolean
  opening?: boolean
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  send: [content: string, contentType?: number, reply?: MessageReplyQuote | null]
  recall: [msg: MessageItem]
  loadMore: []
}>()

const router = useRouter()

function onClose() {
  emit('update:show', false)
}

function onAvatarClick(msg: MessageItem) {
  if (!msg.fromUserId) return
  emit('update:show', false)
  openUserProfile(router, msg.fromUserId, {
    from: 'post',
    postId: props.postId,
    name: msg.fromName || '',
    avatarUrl: msg.fromAvatar || '',
  })
}

function onMessageLongPress(msg: MessageItem) {
  openMessageReportSheet(router, msg)
}
</script>

<template>
  <van-popup
    :show="show"
    position="bottom"
    round
    safe-area-inset-bottom
    :style="{ height: '52%' }"
    @update:show="emit('update:show', $event)"
  >
    <div class="post-chat-sheet">
      <div class="post-chat-sheet__header">
        <div class="app-sheet__handle" />
        <div class="post-chat-sheet__title-row">
          <div class="post-chat-sheet__title-block">
            <h2 class="post-chat-sheet__title">{{ title || '讨论' }}</h2>
            <span v-if="onlineLabel" class="post-chat-sheet__online">{{ onlineLabel }}</span>
          </div>
          <button type="button" class="post-chat-sheet__close" aria-label="关闭" @click="onClose">
            <van-icon name="cross" size="18" />
          </button>
        </div>
      </div>

      <div v-if="opening" class="post-chat-sheet__loading">
        <van-loading size="24px">加载中…</van-loading>
      </div>

      <MessagePanel
        v-else
        class="post-chat-sheet__panel"
        :messages="messages"
        :loading-more="loadingMore"
        :has-more="hasMore"
        :location-share="postId ? { type: 'post', postId } : { type: 'nearby' }"
        empty-text="暂无讨论，来发第一条吧"
        @send="(content, contentType, reply) => emit('send', content, contentType, reply)"
        @recall="emit('recall', $event)"
        @load-more="emit('loadMore')"
        @avatar-click="onAvatarClick"
        @message-long-press="onMessageLongPress"
      />
    </div>
  </van-popup>
</template>

<style scoped>
.post-chat-sheet {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  background: var(--page-bg);
}

.post-chat-sheet__header {
  flex-shrink: 0;
  padding: var(--sheet-pad-top) 12px 8px;
  border-bottom: 1px solid var(--border-color);
}

.post-chat-sheet__title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 32px;
}

.post-chat-sheet__title-block {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 8px;
}

.post-chat-sheet__title {
  margin: 0;
  font-size: var(--sheet-title-size);
  font-weight: var(--sheet-title-weight);
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.post-chat-sheet__online {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  line-height: 1.3;
  font-weight: 400;
  color: var(--text-muted);
}

.post-chat-sheet__online::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--online-color);
}

.post-chat-sheet__close {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: var(--radius-sm);
  background: rgba(0, 0, 0, 0.06);
  color: var(--text-secondary);
}

.post-chat-sheet__loading {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.post-chat-sheet__panel {
  flex: 1;
  min-height: 0;
}
</style>

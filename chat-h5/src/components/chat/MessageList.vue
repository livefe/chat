<script setup lang="ts">
import { computed, ref, toRef } from 'vue'

import MessageBubble from './MessageBubble.vue'
import type { MessageItem } from '@/api/message'
import { useMessageListScroll } from '@/composables/useMessageListScroll'
import {
  useVirtualMessageList,
  virtualRowKey,
  VIRTUAL_HEADER_HEIGHT,
} from '@/composables/useVirtualMessageList'
import { useUserStore } from '@/stores/user'
import { formatChatBubbleTime, shouldShowTimeDivider } from '@/utils/chat/chatTime'

const props = withDefaults(
  defineProps<{
    messages: MessageItem[]
    initialLoading?: boolean
    loadingMore?: boolean
    hasMore?: boolean
    hideSenderName?: boolean
    emptyText?: string
    enableActions?: boolean
  }>(),
  {
    initialLoading: false,
    loadingMore: false,
    hasMore: false,
    hideSenderName: false,
    emptyText: '暂无消息',
    enableActions: true,
  },
)

const emit = defineEmits<{
  avatarClick: [msg: MessageItem]
  reply: [msg: MessageItem]
  forward: [msg: MessageItem]
  report: [msg: MessageItem]
  recall: [msg: MessageItem]
  loadMore: []
  dismiss: []
}>()

const userStore = useUserStore()
const listRef = ref<HTMLDivElement | null>(null)

const messagesRef = toRef(props, 'messages')
const compactRows = computed(() => props.hideSenderName)
const {
  enabled: virtualEnabled,
  visibleRows: virtualVisibleRows,
  layout: virtualLayout,
  bindRowRef: setVirtualRowRef,
  onScroll: onVirtualScroll,
  scrollToBottom: virtualScrollToBottom,
  preserveScrollAfterPrepend,
  getTotalHeight,
  rowTop,
} = useVirtualMessageList(listRef, messagesRef, { compactRows })

const {
  messageKey,
  onScroll,
  onPointerDown,
  scrollToBottom,
  syncAfterActivate,
  requestStick,
  isStickLocked,
  stickLockUntil,
} = useMessageListScroll({
  listRef,
  messages: () => props.messages,
  hasMore: () => props.hasMore,
  loadingMore: () => props.loadingMore,
  virtualEnabled,
  virtualTotal: () => virtualLayout.value.total,
  onVirtualScroll,
  virtualScrollToBottom,
  preserveScrollAfterPrepend,
  getTotalHeight,
  onLoadMore: () => emit('loadMore'),
  onDismiss: () => emit('dismiss'),
})

const showLoadHint = computed(
  () => props.loadingMore || (props.hasMore && props.messages.length > 0),
)

function isSelf(m: MessageItem) {
  return m.fromUserId === userStore.info?.id
}

defineExpose({
  scrollToBottom,
  syncAfterActivate,
  requestStick,
  isStickLocked,
  stickLockUntil,
})
</script>

<template>
  <div class="list-scroll" ref="listRef" @scroll="onScroll" @pointerdown="onPointerDown">
    <div v-if="initialLoading && messages.length === 0" class="tab-page__status">
      <van-loading size="24px" vertical>加载中…</van-loading>
    </div>
    <van-empty v-else-if="messages.length === 0" :description="emptyText" />

    <div
      v-else-if="!virtualEnabled"
      class="list-inner"
      :class="{ 'list-inner--private': hideSenderName }"
    >
      <div v-if="loadingMore" class="load-hint">加载中…</div>
      <div v-else-if="hasMore" class="load-hint muted">上滑加载更多</div>
      <template v-for="(m, i) in messages" :key="messageKey(m)">
        <div
          v-if="
            shouldShowTimeDivider(messages[i - 1]?.createdAt, m.createdAt, { isFirst: i === 0 })
          "
          class="time-divider"
        >
          {{ formatChatBubbleTime(m.createdAt) }}
        </div>
        <MessageBubble
          :message="m"
          :self="isSelf(m)"
          :hide-name="hideSenderName"
          :enable-actions="enableActions"
          @avatar-click="emit('avatarClick', $event)"
          @reply="emit('reply', $event)"
          @forward="emit('forward', $event)"
          @report="emit('report', $event)"
          @recall="emit('recall', $event)"
        />
      </template>
    </div>

    <div v-else class="virtual-body" :style="{ height: `${virtualLayout.total}px` }">
      <div
        v-if="showLoadHint"
        class="load-hint virtual-header"
        :style="{ height: `${VIRTUAL_HEADER_HEIGHT}px` }"
      >
        <span v-if="loadingMore">加载中…</span>
        <span v-else class="muted">上滑加载更多</span>
      </div>

      <div
        v-for="{ row, index } in virtualVisibleRows"
        :key="virtualRowKey(row)"
        class="virtual-row"
        :class="{
          'virtual-row--divider': row.kind === 'divider',
        }"
        :style="{ transform: `translateY(${rowTop(index)}px)` }"
        :ref="(el) => setVirtualRowRef(virtualRowKey(row), row, el)"
      >
        <div v-if="row.kind === 'divider'" class="time-divider">
          {{ formatChatBubbleTime(row.ts) }}
        </div>
        <MessageBubble
          v-else
          :message="row.message"
          :self="isSelf(row.message)"
          :hide-name="hideSenderName"
          :enable-actions="enableActions"
          @avatar-click="emit('avatarClick', $event)"
          @reply="emit('reply', $event)"
          @forward="emit('forward', $event)"
          @report="emit('report', $event)"
          @recall="emit('recall', $event)"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.list-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  -webkit-overflow-scrolling: touch;
  overscroll-behavior-y: none;
  touch-action: pan-y;
}

.list-inner {
  min-height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  gap: 10px;
  padding: 8px 12px 12px;
  box-sizing: border-box;
}

.list-inner--private {
  gap: 4px;
  padding: 8px 12px 14px;
}

.time-divider {
  align-self: center;
  margin: 12px 0 8px;
  padding: 3px 10px;
  font-size: 11px;
  line-height: 1.2;
  color: #888888;
  background: #e5e5e5;
  border-radius: 999px;
  font-variant-numeric: tabular-nums;
  user-select: none;
}

.virtual-body {
  position: relative;
  width: 100%;
}

.virtual-row {
  position: absolute;
  left: 0;
  right: 0;
  padding: 0 12px;
  box-sizing: border-box;
  overflow: visible;
}

.virtual-row--divider {
  display: flex;
  justify-content: center;
  padding: 10px 12px 6px;
  box-sizing: border-box;
}

.virtual-row--divider .time-divider {
  margin: 0;
}

.virtual-header {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.load-hint {
  text-align: center;
  color: var(--text-muted);
  font-size: 12px;
  padding: 4px 0 8px;
}
.load-hint .muted,
.load-hint.muted {
  color: var(--text-tertiary);
}
</style>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showToast } from 'vant'

import MessagePanel from '@/components/chat/MessagePanel.vue'
import SubPageLayout from '@/components/SubPageLayout.vue'
import type { RoomDetail } from '@/api/room'
import type { MessageItem, MessageReplyQuote } from '@/api/message'
import { useChatStore } from '@/stores/chat'
import { useChatActions } from '@/composables/useChatActions'
import { openMessageReportSheet } from '@/composables/useReportAction'
import { openUserProfile } from '@/utils/user'
import { showCatchError } from '@/utils/toast'
import { goBackOr } from '@/utils/nav'
import { isLocationPickNavigation } from '@/utils/map/locationPick'

const props = defineProps<{ id: string }>()

const router = useRouter()
const chatStore = useChatStore()
const chatActions = useChatActions()

const roomId = computed(() => Number(props.id) || 0)
const locationShare = computed(() => ({ type: 'room' as const, roomId: roomId.value }))
const room = ref<RoomDetail | null>(null)
const pageLoading = ref(true)

const title = computed(() => room.value?.name || chatStore.roomName(roomId.value) || '聊天区')
const onlineLabel = computed(() => {
  const n = chatStore.onlineCountForRoom(roomId.value)
  return `${n} 人在线`
})

async function enterCurrentRoom() {
  if (!roomId.value) {
    showToast('无效的聊天区')
    router.replace('/chat')
    return
  }
  chatStore.connect()
  pageLoading.value = true
  try {
    const detail = await chatStore.enterRoom(roomId.value)
    if (!detail) return
    room.value = detail
  } catch (e: unknown) {
    showCatchError(e, '进入聊天区失败', true)
    router.replace('/chat')
  } finally {
    pageLoading.value = false
  }
}

onMounted(() => {
  void enterCurrentRoom()
})

watch(roomId, (id, prev) => {
  if (!id || id === prev) return
  room.value = null
  void enterCurrentRoom()
})

onUnmounted(() => {
  if (isLocationPickNavigation(router.currentRoute.value)) return
  chatStore.leaveRoom()
})

function onSend(content: string, contentType?: number, reply?: MessageReplyQuote | null) {
  void chatActions.send(content, contentType, reply)
}

function onAvatarClick(msg: MessageItem) {
  if (!msg.fromUserId) return
  openUserProfile(router, msg.fromUserId, {
    from: 'room',
    name: msg.fromName || '',
    avatarUrl: msg.fromAvatar || '',
  })
}

function onMessageLongPress(msg: MessageItem) {
  openMessageReportSheet(router, msg)
}

function onBack() {
  chatStore.leaveRoom()
  goBackOr(router, '/chat')
}
</script>

<template>
  <SubPageLayout class="room-layout" full-height :on-back="onBack">
    <template #title>
      <div class="room-nav-title">
        <span class="room-nav-title__name">{{ title }}</span>
        <span class="room-nav-title__sub">{{ onlineLabel }}</span>
      </div>
    </template>

    <van-loading v-if="pageLoading" class="tab-page__status" />
    <div v-else class="room-pane">
      <MessagePanel
        :messages="chatStore.currentMessages"
        :initial-loading="chatStore.messagesLoading"
        :loading-more="chatStore.historyLoading"
        :has-more="chatStore.historyHasMore"
        :location-share="locationShare"
        empty-text="暂无消息"
        @send="onSend"
        @avatar-click="onAvatarClick"
        @message-long-press="onMessageLongPress"
        @recall="chatActions.recall"
        @load-more="chatActions.loadMoreHistory"
      />
    </div>
  </SubPageLayout>
</template>

<style scoped>
.room-nav-title {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.room-nav-title__name {
  flex: 0 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-page-title);
  font-weight: 600;
  color: var(--text-primary);
}
.room-nav-title__sub {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  font-weight: 400;
  color: var(--text-muted);
}
.room-nav-title__sub::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--online-color);
}
.room-pane {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: var(--page-bg);
}
</style>

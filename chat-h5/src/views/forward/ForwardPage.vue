<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showFailToast, showSuccessToast } from 'vant'

import SubPageLayout from '@/components/SubPageLayout.vue'
import type { NearbyRoomItem } from '@/api/room'
import type { ConversationItem } from '@/api/conversation'
import { NEARBY_COPY } from '@/constants/nearby'
import { useChatActions } from '@/composables/useChatActions'
import { useConversationList } from '@/composables/useConversationList'
import { useNearbyChatBootstrap } from '@/composables/useNearbyChatBootstrap'
import { useChatStore } from '@/stores/chat'
import { consumeForwardPayload, type ForwardPayload } from '@/utils/chat/chatForwardNav'
import { sortRoomsByDistance } from '@/utils/chat/chatRooms'
import { goBackOr } from '@/utils/nav'
import { showCatchError } from '@/utils/toast'
import ForwardChatTargets from './components/ForwardChatTargets.vue'
import ForwardExternalShare from './components/ForwardExternalShare.vue'
import ForwardRoomTargets from './components/ForwardRoomTargets.vue'

type Tab = 'rooms' | 'chats'

const router = useRouter()
const chatStore = useChatStore()
const chatActions = useChatActions()
const { hasLocation, refreshRooms } = useNearbyChatBootstrap()
const {
  list: chats,
  loading: chatsLoading,
  finished: chatsFinished,
  loadMore: loadChats,
  reset: resetChats,
} = useConversationList({
  onError: (e) => showCatchError(e, '加载失败'),
})

const payload = ref<ForwardPayload | null>(null)
const tab = ref<Tab>('chats')
const sending = ref(false)
const roomsLoading = ref(false)

const pageTitle = computed(
  () => payload.value?.pageTitle || (payload.value?.shareUrl ? '分享动态' : '转发'),
)
const canForward = computed(() => !!payload.value?.content?.trim())
const rooms = computed(() => sortRoomsByDistance(chatStore.rooms))

function handleBack() {
  goBackOr(router, '/chat')
}

async function forwardTo(send: (content: string, contentType?: number) => Promise<boolean>) {
  const p = payload.value
  if (sending.value || !canForward.value || !p) return
  sending.value = true
  try {
    if (await send(p.content, p.contentType)) {
      showSuccessToast('已转发')
      handleBack()
    }
  } finally {
    sending.value = false
  }
}

function onNearby() {
  if (sending.value || !canForward.value || !payload.value) return
  if (!hasLocation.value) {
    showFailToast(NEARBY_COPY.chooseLocation)
    return
  }
  void forwardTo(chatActions.sendToNearby)
}

function onRoom(room: NearbyRoomItem) {
  if (room.id <= 0) return
  void forwardTo((content, contentType) => chatActions.sendToRoom(room.id, content, contentType))
}

function onChat(c: ConversationItem) {
  if (c.peerId <= 0) return
  void forwardTo((content, contentType) =>
    chatActions.sendToPrivateChat(c.peerId, content, contentType),
  )
}

async function loadRooms() {
  if (!hasLocation.value) return
  roomsLoading.value = true
  try {
    await refreshRooms()
  } catch (e: unknown) {
    showCatchError(e, '加载聊天区失败')
  } finally {
    roomsLoading.value = false
  }
}

watch(tab, (t) => {
  if (t === 'rooms') void loadRooms()
  else void resetChats()
})

onMounted(() => {
  const p = consumeForwardPayload()
  if (!p?.content?.trim()) {
    showFailToast('转发内容无效')
    handleBack()
    return
  }
  payload.value = p
  void resetChats()
})
</script>

<template>
  <SubPageLayout :title="pageTitle" full-height :on-back="handleBack">
    <div v-if="payload" class="forward-page">
      <button type="button" class="forward-nearby" :disabled="sending" @click="onNearby">
        <span class="forward-nearby__icon">
          <van-icon name="chat-o" size="20" />
        </span>
        <span class="forward-nearby__title">附近聊天</span>
        <van-icon name="arrow" class="forward-nearby__arrow" />
      </button>

      <section class="forward-sheet">
        <div class="forward-tabs">
          <button
            type="button"
            class="forward-tabs__item"
            :class="{ 'forward-tabs__item--on': tab === 'chats' }"
            @click="tab = 'chats'"
          >
            消息
          </button>
          <button
            type="button"
            class="forward-tabs__item"
            :class="{ 'forward-tabs__item--on': tab === 'rooms' }"
            @click="tab = 'rooms'"
          >
            聊天区
          </button>
        </div>

        <ForwardChatTargets
          v-if="tab === 'chats'"
          v-model:loading="chatsLoading"
          :chats="chats"
          :finished="chatsFinished"
          :sending="sending"
          @load="loadChats"
          @select="onChat"
        />
        <ForwardRoomTargets
          v-else
          :rooms="rooms"
          :loading="roomsLoading"
          :has-location="hasLocation"
          :sending="sending"
          @select="onRoom"
        />
      </section>

      <ForwardExternalShare v-if="payload.shareUrl" :payload="payload" @shared="handleBack" />
    </div>
  </SubPageLayout>
</template>

<style scoped>
.forward-page {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: var(--page-bg);
}

.forward-nearby {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  margin: 0;
  padding: 14px 16px;
  border: none;
  border-radius: 0;
  background: var(--card-bg, #fff);
  text-align: left;
  cursor: pointer;
}

.forward-nearby:active:not(:disabled) {
  background: rgba(0, 0, 0, 0.03);
}

.forward-nearby:disabled {
  opacity: 0.55;
}

.forward-nearby__icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 12px;
  color: var(--text-secondary);
  background: var(--page-bg);
}

.forward-nearby__title {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.forward-nearby__arrow {
  color: #c0c0c0;
  font-size: 20px;
}

.forward-sheet {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  margin-top: 8px;
  background: var(--card-bg, #fff);
  border-radius: 0;
}

.forward-tabs {
  display: flex;
  gap: 24px;
  flex-shrink: 0;
  padding: 0 16px;
  border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.06));
}

.forward-tabs__item {
  position: relative;
  padding: 12px 0;
  border: none;
  background: transparent;
  font-size: 15px;
  color: var(--text-secondary);
  cursor: pointer;
}

.forward-tabs__item--on {
  font-weight: 600;
  color: var(--text-primary);
}

.forward-tabs__item--on::after {
  content: '';
  position: absolute;
  left: 50%;
  bottom: 0;
  width: 28px;
  height: 2px;
  border-radius: 1px;
  background: var(--brand-color, #4c8dff);
  transform: translateX(-50%);
}
</style>

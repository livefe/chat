<script setup lang="ts">
import { computed, onActivated, onDeactivated, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { showFailToast } from 'vant'

import MessagePanel from '@/components/chat/MessagePanel.vue'
import NearbyRoomMarquee from './components/NearbyRoomMarquee.vue'
import { ROOM_MARQUEE_LIMIT } from '@/constants/chat'
import { NEARBY_COPY } from '@/constants/nearby'
import { useChatStore } from '@/stores/chat'
import type { NearbyRoomItem } from '@/api/room'
import type { MessageItem, MessageReplyQuote } from '@/api/message'
import { useChatActions } from '@/composables/useChatActions'
import { useNearbyChatBootstrap } from '@/composables/useNearbyChatBootstrap'
import { openMessageReportSheet } from '@/composables/useReportAction'
import { sortRoomsByDistance } from '@/utils/chat/chatRooms'
import { openUserProfile } from '@/utils/user'
import { locationPickRoute } from '@/utils/map/locationPick'
import { showCatchError } from '@/utils/toast'

const router = useRouter()
const chatStore = useChatStore()
const chatActions = useChatActions()
const nearby = useNearbyChatBootstrap()
const { hasLocation, isBrowsing, locationLabel } = nearby
const chatBootstrapping = ref(false)
const showMarquee = ref(false)
const locationSheetVisible = ref(false)

const nearbyStats = computed(() => {
  if (chatStore.activeChannel !== 'nearby') return ''
  return `${chatStore.nearbyOnlineCount} ${NEARBY_COPY.onlineSuffix}`
})

const marqueeRooms = computed(() =>
  sortRoomsByDistance(chatStore.rooms).slice(0, ROOM_MARQUEE_LIMIT),
)
const marqueeKey = computed(() => marqueeRooms.value.map((r) => r.id).join(','))

watch(
  [marqueeKey, hasLocation],
  () => {
    showMarquee.value = hasLocation.value && marqueeRooms.value.length > 0
  },
  { immediate: true },
)

function onMarqueeFinished() {
  showMarquee.value = false
}

function openNearbyRooms() {
  router.push('/chat/rooms')
}

onActivated(async () => {
  chatBootstrapping.value = true
  nearby.startPresencePolling()
  try {
    await nearby.activate()
  } catch (e: unknown) {
    showCatchError(e, '加载附近聊天失败', true)
  } finally {
    chatBootstrapping.value = false
  }
})

onDeactivated(() => {
  nearby.stopPresencePolling()
})

function onRoomClick(room: NearbyRoomItem) {
  router.push(`/room/${room.id}`)
}

function onSend(content: string, contentType?: number, reply?: MessageReplyQuote | null) {
  void chatActions.send(content, contentType, reply)
}

function onAvatarClick(msg: MessageItem) {
  if (!msg.fromUserId) return
  openUserProfile(router, msg.fromUserId, {
    from: 'chat',
    name: msg.fromName || '',
    avatarUrl: msg.fromAvatar || '',
  })
}

function onMessageLongPress(msg: MessageItem) {
  openMessageReportSheet(router, msg)
}

function openLocationSheet() {
  locationSheetVisible.value = true
}

function onPickLocation() {
  locationSheetVisible.value = false
  router.push(locationPickRoute('chat'))
}

async function onUseGps() {
  locationSheetVisible.value = false
  try {
    if (!(await nearby.relocateToGps())) showFailToast('定位失败')
  } catch (e: unknown) {
    showCatchError(e, '定位失败', true)
  }
}
</script>

<template>
  <div class="chat-page">
    <header class="page-header">
      <div class="page-header__top">
        <div class="page-header__main">
          <button type="button" class="location-switcher" @click="openLocationSheet">
            <span class="location-switcher__icon" aria-hidden="true">
              <svg width="18" height="18" viewBox="0 0 18 18" fill="none">
                <circle cx="9" cy="9" r="7.25" stroke="currentColor" stroke-width="1.5" />
                <circle cx="9" cy="9" r="2.25" fill="currentColor" />
              </svg>
            </span>
            <span class="location-switcher__text">{{ locationLabel }}</span>
            <span class="location-switcher__arrow" aria-hidden="true">
              <svg width="10" height="6" viewBox="0 0 10 6" fill="currentColor">
                <path d="M1 1.2 5 4.8 9 1.2" stroke="currentColor" stroke-width="1.4" fill="none" />
              </svg>
            </span>
          </button>
          <span v-if="nearbyStats" class="header-online">{{ nearbyStats }}</span>
        </div>

        <button type="button" class="rooms-entry-btn" aria-label="聊天区" @click="openNearbyRooms">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path
              d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"
            />
          </svg>
        </button>
      </div>
    </header>

    <div class="chat-body">
      <MessagePanel
        :messages="chatStore.currentMessages"
        :initial-loading="chatBootstrapping || chatStore.messagesLoading"
        :loading-more="chatStore.historyLoading"
        :has-more="chatStore.historyHasMore"
        :empty-text="NEARBY_COPY.emptyMessages"
        :can-compose="hasLocation"
        :location-share="{ type: 'nearby' }"
        @send="onSend"
        @avatar-click="onAvatarClick"
        @message-long-press="onMessageLongPress"
        @recall="chatActions.recall"
        @load-more="chatActions.loadMoreHistory"
      />
      <NearbyRoomMarquee
        v-if="showMarquee"
        :key="marqueeKey"
        class="nearby-marquee"
        :rooms="marqueeRooms"
        @select="onRoomClick"
        @finished="onMarqueeFinished"
      />
    </div>

    <van-popup v-model:show="locationSheetVisible" position="bottom" round safe-area-inset-bottom>
      <div class="app-sheet">
        <div class="app-sheet__handle" />
        <button type="button" class="app-sheet__item" @click="onUseGps">
          <van-icon name="aim" class="location-sheet__icon" />
          <span class="location-sheet__label">重新定位</span>
          <van-icon
            v-if="hasLocation && !isBrowsing"
            name="success"
            class="location-sheet__check"
          />
        </button>
        <button type="button" class="app-sheet__item" @click="onPickLocation">
          <van-icon name="location-o" class="location-sheet__icon" />
          <span class="location-sheet__label">地图选点</span>
          <van-icon v-if="hasLocation && isBrowsing" name="success" class="location-sheet__check" />
        </button>
      </div>
    </van-popup>
  </div>
</template>

<style scoped>
.chat-page {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  height: 100%;
  background: var(--page-bg);
}

.page-header {
  flex-shrink: 0;
  padding: calc(14px + env(safe-area-inset-top, 0px)) 16px 8px;
  background: var(--page-bg);
}

.page-header__top {
  display: flex;
  align-items: center;
  gap: 10px;
}

.page-header__main {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
}

.rooms-entry-btn {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-primary);
  cursor: pointer;
  outline: none;
  -webkit-tap-highlight-color: transparent;
}

.location-switcher {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 0 1 auto;
  min-width: 0;
  max-width: 50%;
  padding: 0;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: var(--font-page-title);
  font-weight: 600;
  line-height: 1.3;
  cursor: pointer;
}

.location-switcher__icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-primary);
}

.location-switcher__text {
  flex: 0 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.location-switcher__arrow {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  margin-top: 2px;
  color: var(--text-primary);
}

.header-online {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
}

.header-online::before {
  content: '';
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--online-color);
}

.chat-body {
  position: relative;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.nearby-marquee {
  position: absolute;
  top: 8px;
  left: 0;
  right: 0;
  z-index: 2;
}

.location-sheet__icon {
  font-size: 20px;
  color: var(--text-primary);
}

.location-sheet__label {
  flex: 1;
  font-size: var(--sheet-item-size);
  font-weight: var(--sheet-item-weight);
  color: var(--text-primary);
}

.location-sheet__check {
  font-size: 18px;
  color: var(--brand-color);
}
</style>

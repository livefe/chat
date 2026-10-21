<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import SubPageLayout from '@/components/SubPageLayout.vue'
import RoomPlaceIcon from '@/components/chat/RoomPlaceIcon.vue'
import type { NearbyRoomItem } from '@/api/room'
import { useChatStore } from '@/stores/chat'
import { useNearbyChatBootstrap } from '@/composables/useNearbyChatBootstrap'
import { formatDistance } from '@/utils/map/geo'
import { roomPlaceTypeLabel } from '@/utils/chat/chatRoomPlaceType'
import { sortRoomsByDistance } from '@/utils/chat/chatRooms'
import { showCatchError } from '@/utils/toast'
import { goBackOr } from '@/utils/nav'

const router = useRouter()
const chatStore = useChatStore()
const { hasLocation, refreshRooms: refreshNearbyRooms } = useNearbyChatBootstrap()
const refreshing = ref(false)
const loading = ref(false)

const sortedRooms = computed(() => sortRoomsByDistance(chatStore.rooms))

function handleBack() {
  goBackOr(router, '/chat')
}

async function refreshRooms(opts?: { silent?: boolean }) {
  if (!opts?.silent) loading.value = true
  try {
    await refreshNearbyRooms()
  } catch (e: unknown) {
    showCatchError(e, '刷新聊天区失败', true)
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

function onRoomClick(room: NearbyRoomItem) {
  router.push(`/room/${room.id}`)
}

function openCreateRoom() {
  router.push('/chat/room/create')
}

onMounted(() => {
  if (!hasLocation.value) return
  if (sortedRooms.value.length === 0) {
    void refreshRooms()
  }
})
</script>

<template>
  <SubPageLayout title="聊天区" full-height :on-back="handleBack">
    <template #right>
      <button type="button" class="rooms-nav-add" aria-label="新建聊天区" @click="openCreateRoom">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
          <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
        </svg>
      </button>
    </template>

    <van-pull-refresh
      v-model="refreshing"
      class="rooms-page sub-page__scroll"
      @refresh="refreshRooms({ silent: true })"
    >
      <div v-if="loading && sortedRooms.length === 0" class="tab-page__status">
        <van-loading size="24px" vertical>加载中…</van-loading>
      </div>
      <van-empty v-else-if="!hasLocation" description="选择位置" />
      <van-empty v-else-if="sortedRooms.length === 0" description="暂无聊天区" />
      <div v-else class="room-list">
        <button
          v-for="room in sortedRooms"
          :key="room.id"
          type="button"
          class="room-item"
          @click="onRoomClick(room)"
        >
          <RoomPlaceIcon :place-type="room.placeType" />
          <span class="room-item__body">
            <span class="room-item__name">{{ room.name || '聊天区' }}</span>
            <span class="room-item__meta">
              <template v-if="room.placeTypeText?.trim() || room.placeType">
                {{ room.placeTypeText?.trim() || roomPlaceTypeLabel(room.placeType) }} ·
              </template>
              距离 {{ formatDistance(room.distance) }}
            </span>
          </span>
          <van-icon name="arrow" class="room-item__arrow" />
        </button>
      </div>
    </van-pull-refresh>
  </SubPageLayout>
</template>

<style scoped>
.rooms-nav-add {
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
}

.rooms-page {
  flex: 1;
  min-height: 0;
  height: 100%;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  background: var(--page-bg);
}

.room-list {
  padding: 8px 12px 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.room-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 14px;
  border: none;
  border-radius: var(--radius-md);
  background: var(--card-bg);
  text-align: left;
  cursor: pointer;
}

.room-item:active {
  background: rgba(0, 0, 0, 0.02);
}

.room-item__body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.room-item__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.room-item__meta {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.room-item__arrow {
  flex-shrink: 0;
  color: #c0c0c0;
  font-size: 20px;
}
</style>

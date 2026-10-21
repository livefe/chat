<script setup lang="ts">
import { computed, ref } from 'vue'

import RoomPlaceIcon from '@/components/chat/RoomPlaceIcon.vue'
import type { NearbyRoomItem } from '@/api/room'
import { NEARBY_COPY } from '@/constants/nearby'
import { roomPlaceTypeLabel } from '@/utils/chat/chatRoomPlaceType'
import ForwardTargetItem from './ForwardTargetItem.vue'
import ForwardTargetPane from './ForwardTargetPane.vue'

const props = defineProps<{
  rooms: NearbyRoomItem[]
  loading: boolean
  hasLocation: boolean
  sending: boolean
}>()

const emit = defineEmits<{
  select: [room: NearbyRoomItem]
}>()

const query = ref('')

const filteredRooms = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return props.rooms
  return props.rooms.filter((room) => {
    const name = (room.name || '').toLowerCase()
    const type = (room.placeTypeText || roomPlaceTypeLabel(room.placeType) || '').toLowerCase()
    return name.includes(q) || type.includes(q)
  })
})
</script>

<template>
  <ForwardTargetPane v-model:query="query" placeholder="搜索聊天区" :disabled="sending">
    <van-loading v-if="loading && rooms.length === 0" size="24px" />
    <van-empty v-else-if="!hasLocation" :description="NEARBY_COPY.chooseLocation" />
    <van-empty v-else-if="rooms.length === 0" description="暂无聊天区" />
    <van-empty v-else-if="filteredRooms.length === 0" description="没有匹配的聊天区" />
    <div v-else class="forward-list">
      <ForwardTargetItem
        v-for="room in filteredRooms"
        :key="room.id"
        :name="room.name || '聊天区'"
        :disabled="sending"
        @select="emit('select', room)"
      >
        <template #icon>
          <RoomPlaceIcon :place-type="room.placeType" />
        </template>
      </ForwardTargetItem>
    </div>
  </ForwardTargetPane>
</template>

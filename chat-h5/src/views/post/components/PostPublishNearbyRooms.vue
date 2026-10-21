<script setup lang="ts">
import type { NearbyRoomItem } from '@/api/room'

defineProps<{
  rooms: NearbyRoomItem[]
  selectedId: number
  loading: boolean
}>()

const emit = defineEmits<{
  select: [room: NearbyRoomItem]
}>()
</script>

<template>
  <div class="location-nearby">
    <span class="location-nearby__label">附近</span>
    <div class="location-nearby__list">
      <button
        v-for="room in rooms"
        :key="room.id"
        type="button"
        class="location-nearby__item"
        :class="{ 'location-nearby__item--on': selectedId === room.id }"
        @click="emit('select', room)"
      >
        {{ room.name || '聊天区' }}
      </button>
      <span v-if="loading && !rooms.length" class="location-nearby__status">查找中…</span>
    </div>
  </div>
</template>

<style scoped>
.location-nearby {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 16px 12px;
}
.location-nearby__label {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--text-tertiary);
}
.location-nearby__list {
  display: flex;
  flex: 1;
  min-width: 0;
  gap: 8px;
  overflow-x: auto;
  overscroll-behavior-x: contain;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
}
.location-nearby__list::-webkit-scrollbar {
  display: none;
}
.location-nearby__item {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  max-width: 120px;
  height: 28px;
  padding: 0 10px;
  border: none;
  border-radius: 6px;
  background: var(--page-bg);
  font: inherit;
  font-size: 12px;
  line-height: 1;
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
  -webkit-appearance: none;
  appearance: none;
}
.location-nearby__item--on {
  background: var(--brand-soft);
  color: var(--text-primary);
  font-weight: 500;
}
.location-nearby__status {
  flex-shrink: 0;
  font-size: 12px;
  color: var(--text-tertiary);
}
</style>

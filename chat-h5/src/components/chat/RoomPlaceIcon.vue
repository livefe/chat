<script setup lang="ts">
import { roomPlaceTypeMark } from '@/utils/chat/chatRoomPlaceType'

defineProps<{ placeType?: number | null }>()
</script>

<template>
  <span class="room-place-icon" :data-type="placeType || 7">
    {{ roomPlaceTypeMark(placeType) }}
  </span>
</template>

<style scoped>
.room-place-icon {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-sm);
  font-size: 15px;
  font-weight: 600;
  line-height: 1;
  color: var(--brand-color-dark);
  background: var(--brand-soft);
}

.room-place-icon[data-type='1'] {
  color: var(--brand-color-dark);
  background: var(--brand-soft);
}
.room-place-icon[data-type='2'] {
  color: #0f766e;
  background: #eeeeee;
}
.room-place-icon[data-type='3'] {
  color: #c2410c;
  background: #eeeeee;
}
.room-place-icon[data-type='4'] {
  color: #0369a1;
  background: #eeeeee;
}
.room-place-icon[data-type='5'] {
  color: #4d7c0f;
  background: #eeeeee;
}
.room-place-icon[data-type='6'] {
  color: #555555;
  background: #f0f0f0;
}
.room-place-icon[data-type='7'] {
  color: #6b6b6b;
  background: #f5f5f5;
}
</style>

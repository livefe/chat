<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import type { NearbyRoomItem } from '@/api/room'

const props = defineProps<{
  rooms: NearbyRoomItem[]
}>()

const emit = defineEmits<{
  select: [room: NearbyRoomItem]
  finished: []
}>()

const rootRef = ref<HTMLElement | null>(null)
const trackRef = ref<HTMLElement | null>(null)
const fading = ref(false)
const entered = ref(false)

let paused = false
let rafId = 0
let finished = false

function stopLoop() {
  if (rafId) {
    cancelAnimationFrame(rafId)
    rafId = 0
  }
}

function finish() {
  if (finished) return
  finished = true
  stopLoop()
  fading.value = true
  window.setTimeout(() => emit('finished'), 420)
}

async function start() {
  finished = false
  fading.value = false
  entered.value = false
  stopLoop()
  await nextTick()
  const root = rootRef.value
  const track = trackRef.value
  if (!root || !track || props.rooms.length === 0) {
    emit('finished')
    return
  }
  const pad = root.clientWidth
  track.style.paddingLeft = `${pad}px`
  track.style.paddingRight = `${pad}px`
  root.scrollLeft = 0
  entered.value = true
  await new Promise((r) => setTimeout(r, 520))
  if (finished) return

  const step = () => {
    if (finished) return
    const el = rootRef.value
    if (!el) {
      finish()
      return
    }
    if (!paused) {
      const max = el.scrollWidth - el.clientWidth
      if (max <= 0 || el.scrollLeft >= max - 0.5) {
        finish()
        return
      }
      el.scrollLeft += 1
    }
    rafId = requestAnimationFrame(step)
  }
  rafId = requestAnimationFrame(step)
}

function onPointerDown() {
  paused = true
}

function onPointerUp() {
  paused = false
}

function onSelect(room: NearbyRoomItem) {
  emit('select', room)
}

onMounted(() => {
  void start()
})

watch(
  () => props.rooms.map((r) => r.id).join(','),
  () => {
    void start()
  },
)

onBeforeUnmount(() => {
  stopLoop()
})
</script>

<template>
  <div
    ref="rootRef"
    class="room-marquee"
    :class="{ 'room-marquee--entered': entered, 'room-marquee--fading': fading }"
    @pointerdown="onPointerDown"
    @pointerup="onPointerUp"
    @pointercancel="onPointerUp"
    @pointerleave="onPointerUp"
  >
    <div ref="trackRef" class="room-marquee__track">
      <button
        v-for="room in rooms"
        :key="room.id"
        type="button"
        class="room-marquee__chip"
        @click="onSelect(room)"
      >
        {{ room.name || '聊天区' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.room-marquee {
  width: 100%;
  height: 37px;
  overflow-x: auto;
  overflow-y: hidden;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: none;
  opacity: 0;
  transition: opacity 0.36s cubic-bezier(0.4, 0, 0.2, 1);

  pointer-events: auto;
}

.room-marquee::-webkit-scrollbar {
  display: none;
}

.room-marquee--entered {
  opacity: 1;
}

.room-marquee--fading {
  opacity: 0;
  transition-duration: 0.42s;
  pointer-events: none;
}

.room-marquee__track {
  display: flex;
  align-items: center;
  gap: 14px;
  height: 100%;
  width: max-content;
  box-sizing: content-box;
}

.room-marquee__chip {
  flex-shrink: 0;
  height: 37px;
  padding: 0 12px;
  border: none;
  border-radius: var(--radius-bubble);
  background: var(--card-bg);

  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.06);
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
  line-height: 21px;
  white-space: nowrap;
  cursor: pointer;
}

.room-marquee__chip:active {
  opacity: 0.88;
}
</style>

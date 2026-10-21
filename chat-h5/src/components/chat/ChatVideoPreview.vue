<script setup lang="ts">
import { nextTick, onMounted, onUnmounted, ref, watch } from 'vue'

const props = defineProps<{
  src: string
}>()

const emit = defineEmits<{
  close: []
}>()

const videoRef = ref<HTMLVideoElement | null>(null)
const errorText = ref('')

function onClose() {
  const el = videoRef.value
  if (el) el.pause()
  emit('close')
}

async function tryPlay() {
  const el = videoRef.value
  if (!el) return
  el.setAttribute('playsinline', '')
  el.setAttribute('webkit-playsinline', '')
  el.setAttribute('x5-playsinline', 'true')
  el.setAttribute('x5-video-player-type', 'h5')
  el.playsInline = true
  try {
    await el.play()
    errorText.value = ''
  } catch {
    errorText.value = '点击下方控件播放'
  }
}

onMounted(async () => {
  await nextTick()
  await tryPlay()
})

watch(
  () => props.src,
  async () => {
    errorText.value = ''
    await nextTick()
    await tryPlay()
  },
)

onUnmounted(() => {
  const el = videoRef.value
  if (el) el.pause()
})
</script>

<template>
  <Teleport to="body">
    <div class="video-preview" @click.self="onClose">
      <video
        ref="videoRef"
        class="video-preview__player"
        :src="props.src"
        controls
        playsinline
        webkit-playsinline
        x5-playsinline
        x5-video-player-type="h5"
        preload="auto"
        @click.stop
      />
      <div v-if="errorText" class="video-preview__hint" @click.stop="tryPlay">{{ errorText }}</div>
      <button type="button" class="video-preview__close" aria-label="关闭" @click="onClose">
        <van-icon name="cross" size="22" />
      </button>
    </div>
  </Teleport>
</template>

<style scoped>
.video-preview {
  position: fixed;
  inset: 0;
  z-index: 3000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.92);
}

.video-preview__player {
  width: 100%;
  height: 100%;
  max-width: 100vw;
  max-height: 100dvh;
  background: #000;
  object-fit: contain;
}

.video-preview__hint {
  position: absolute;
  left: 50%;
  bottom: 72px;
  transform: translateX(-50%);
  padding: 8px 14px;
  border-radius: 999px;
  background: rgba(0, 0, 0, 0.55);
  color: #fff;
  font-size: 13px;
}

.video-preview__close {
  position: absolute;
  top: 12px;
  right: 12px;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  padding: 0;
  border: none;
  background: transparent;
  color: #fff;
}
</style>

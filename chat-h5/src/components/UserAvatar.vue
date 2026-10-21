<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import { resolveMediaUrl } from '@/utils/media'

const props = withDefaults(
  defineProps<{
    src?: string
    size?: number
    tone?: 'brand' | 'profile' | 'chat'
  }>(),
  { size: 32, tone: 'brand' },
)

const emit = defineEmits<{ click: [] }>()

const loadFailed = ref(false)

const resolvedSrc = computed(() => resolveMediaUrl(props.src))

watch(
  () => props.src,
  () => {
    loadFailed.value = false
  },
)

const showImage = computed(() => !!resolvedSrc.value && !loadFailed.value)
</script>

<template>
  <div
    class="user-avatar"
    :class="`tone-${tone}`"
    :style="{ width: `${size}px`, height: `${size}px` }"
    @click="emit('click')"
  >
    <img v-if="showImage" :src="resolvedSrc" alt="" @error="loadFailed = true" />
  </div>
</template>

<style scoped>
.user-avatar {
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  flex-shrink: 0;
}
.user-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}
.tone-brand {
  background: #e8e8e8;
}
.tone-profile {
  background: #e8e8e8;
}
.tone-chat {
  background: #e8e8e8;
}
</style>

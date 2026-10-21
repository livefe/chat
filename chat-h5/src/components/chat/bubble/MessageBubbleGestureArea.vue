<script setup lang="ts">
import type { MessageItem } from '@/api/message'
import type { MessageBubbleGestures } from '@/composables/useMessageBubbleMenu'

const props = withDefaults(
  defineProps<{
    sendStatus: MessageItem['sendStatus']
    gestures: MessageBubbleGestures
    tag?: 'div' | 'button'
  }>(),
  { tag: 'div' },
)

const emit = defineEmits<{
  release: [e: TouchEvent]
}>()

function onTouchStart(e: TouchEvent) {
  if (props.gestures.enabled) props.gestures.onTouchStart(e)
}

function onTouchMove(e: TouchEvent) {
  if (props.gestures.enabled) props.gestures.onTouchMove(e)
}

function onTouchEnd(e: TouchEvent) {
  if (!props.gestures.enabled) return
  props.gestures.onTouchEnd()
  emit('release', e)
}

function onTouchCancel() {
  if (props.gestures.enabled) props.gestures.onTouchEnd()
}
</script>

<template>
  <component
    :is="tag"
    :type="tag === 'button' ? 'button' : undefined"
    :class="{
      'bubble-press': true,
      'is-sending': sendStatus === 'sending',
      'is-failed': sendStatus === 'failed',
    }"
    @touchstart.passive="onTouchStart"
    @touchend="onTouchEnd"
    @touchmove="onTouchMove"
    @touchcancel="onTouchCancel"
    @contextmenu.prevent
  >
    <slot />
  </component>
</template>

<style scoped>
.bubble-press {
  -webkit-touch-callout: none;
  -webkit-user-select: none;
  user-select: none;
}

.is-sending {
  opacity: 0.72;
}

.is-failed {
  opacity: 0.9;
}
</style>

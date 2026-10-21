<script setup lang="ts">
import { ref } from 'vue'

import type { NavMoreAction } from '@/types/nav'

defineProps<{
  actions: NavMoreAction[]
}>()

const emit = defineEmits<{
  select: [action: NavMoreAction]
}>()

const visible = ref(false)

function open() {
  visible.value = true
}

function onSelect(action: NavMoreAction) {
  emit('select', action)
}
</script>

<template>
  <button type="button" class="nav-more-btn" aria-label="更多" @click.stop="open">
    <van-icon name="ellipsis" size="22" />
  </button>
  <van-action-sheet
    v-model:show="visible"
    :actions="actions"
    cancel-text="取消"
    close-on-click-action
    @select="onSelect"
  />
</template>

<style scoped>
.nav-more-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border: none;
  background: transparent;
  padding: 0;
  margin: 0;
  color: inherit;
  cursor: pointer;
}
</style>

<script setup lang="ts">
import type { ImWsState } from '@/utils/imWs'

defineProps<{
  state: ImWsState
}>()

const emit = defineEmits<{
  reconnect: []
}>()
</script>

<template>
  <div
    v-if="state === 'connecting' || state === 'failed'"
    class="im-conn-banner"
    :class="{ 'is-failed': state === 'failed' }"
  >
    <span v-if="state === 'connecting'">实时连接中…</span>
    <template v-else>
      <span>实时连接已断开，新消息可能延迟</span>
      <el-button link type="primary" @click="emit('reconnect')">重新连接</el-button>
    </template>
  </div>
</template>

<style scoped>
.im-conn-banner {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  flex-shrink: 0;
  padding: 6px 12px;
  font-size: 12px;
  color: var(--el-color-warning-dark-2, #b88230);
  background: var(--el-color-warning-light-9, #fdf6ec);
  border-bottom: 1px solid var(--el-color-warning-light-7, #f5dab1);
}

.im-conn-banner.is-failed {
  color: var(--el-color-danger);
  background: var(--el-color-danger-light-9, #fef0f0);
  border-bottom-color: var(--el-color-danger-light-7, #fde2e2);
}
</style>

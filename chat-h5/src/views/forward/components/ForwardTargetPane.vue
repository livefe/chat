<script setup lang="ts">
const query = defineModel<string>('query', { required: true })

defineProps<{
  placeholder: string
  disabled: boolean
}>()
</script>

<template>
  <div class="forward-target-pane">
    <van-search v-model="query" :placeholder="placeholder" shape="round" :disabled="disabled" />
    <div class="forward-pane">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.forward-target-pane {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.forward-target-pane :deep(.van-search) {
  flex-shrink: 0;
  padding: 8px 12px;
  background: transparent;
}

.forward-target-pane :deep(.van-search__content) {
  background: var(--page-bg);
}

.forward-pane {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  display: flex;
  flex-direction: column;
}

.forward-pane :slotted(.van-loading),
.forward-pane :slotted(.van-empty) {
  margin: auto;
  padding: 24px 0;
}

.forward-pane :slotted(.forward-list) {
  padding: 0;
  display: flex;
  flex-direction: column;
}
</style>

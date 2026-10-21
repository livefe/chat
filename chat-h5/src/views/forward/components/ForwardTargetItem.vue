<script setup lang="ts">
defineProps<{
  name: string
  disabled: boolean
}>()

const emit = defineEmits<{
  select: []
}>()
</script>

<template>
  <button type="button" class="target-item" :disabled="disabled" @click="emit('select')">
    <slot name="icon" />
    <span class="target-item__body">
      <span class="target-item__name">{{ name }}</span>
    </span>
    <van-icon name="arrow" class="target-item__arrow" />
  </button>
</template>

<style scoped>
.target-item {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 12px 16px;
  border: none;
  border-radius: 0;
  border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.06));
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.target-item:last-of-type {
  border-bottom: none;
}

.target-item:active:not(:disabled) {
  background: rgba(0, 0, 0, 0.03);
}

.target-item:disabled {
  opacity: 0.55;
}

.target-item__body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.target-item__name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.target-item__arrow {
  flex-shrink: 0;
  color: #c0c0c0;
  font-size: 20px;
}
</style>

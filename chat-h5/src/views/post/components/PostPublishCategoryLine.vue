<script setup lang="ts">
defineProps<{
  label?: string
  hint?: string
  locked: boolean
}>()

const emit = defineEmits<{
  open: []
}>()
</script>

<template>
  <button
    type="button"
    class="category-line"
    :class="{ 'category-line--static': locked }"
    :disabled="locked"
    @click="emit('open')"
  >
    <span class="category-line__main">
      <span class="category-line__label">{{ label }}</span>
      <span class="category-line__hint">{{ hint }}</span>
    </span>
    <van-icon v-if="!locked" name="arrow" class="category-line__arrow" />
  </button>
</template>

<style scoped>
.category-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
  padding: 14px 16px;
  border: none;
  background: var(--card-bg);
  text-align: left;
  cursor: pointer;
  font: inherit;
  color: inherit;
  -webkit-appearance: none;
  appearance: none;
}
.category-line:disabled {
  opacity: 1;
}
.category-line--static {
  cursor: default;
}
.category-line__main {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}
.category-line__label {
  flex-shrink: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}
.category-line__hint {
  font-size: 12px;
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.category-line__arrow {
  flex-shrink: 0;
  font-size: 14px;
  color: var(--text-tertiary);
}
</style>

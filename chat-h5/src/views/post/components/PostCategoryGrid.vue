<script setup lang="ts">
import { POST_CATEGORY_OPTIONS, type PostCategoryValue } from '@/utils/post/postCategory'

defineProps<{
  active?: string
  sheet?: boolean
}>()

const emit = defineEmits<{
  select: [value: PostCategoryValue]
}>()
</script>

<template>
  <div class="pick-grid" :class="{ 'pick-grid--sheet': sheet }" :role="sheet ? undefined : 'list'">
    <button
      v-for="opt in POST_CATEGORY_OPTIONS"
      :key="opt.value"
      type="button"
      class="pick-card"
      :class="{ 'pick-card--active': active === opt.value }"
      :role="sheet ? undefined : 'listitem'"
      @click="emit('select', opt.value)"
    >
      <span class="pick-card__label">{{ opt.label }}</span>
      <span class="pick-card__hint">{{ opt.hint }}</span>
    </button>
  </div>
</template>

<style scoped>
.pick-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}
.pick-grid--sheet {
  padding: 0 16px 8px;
}
.pick-card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 6px;
  padding: 18px 16px;
  border: 1px solid var(--border-color);
  border-radius: 12px;
  background: var(--card-bg);
  text-align: left;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    background 0.15s ease;
}
.pick-card:active {
  background: var(--page-bg);
}
.pick-card--active {
  border-color: var(--brand-soft-strong);
  background: var(--brand-soft);
}
.pick-card__label {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.2;
}
.pick-card__hint {
  font-size: 12px;
  color: var(--text-tertiary);
  line-height: 1.3;
}
.pick-card--active .pick-card__label {
  color: var(--brand-color-dark);
}
.pick-card--active .pick-card__hint {
  color: var(--brand-color);
}
</style>

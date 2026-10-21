<script setup lang="ts">
import PostListActionBar, { type PostListAction } from './PostListActionBar.vue'

defineProps<{
  showCheckInCode: boolean
  actions: PostListAction[]
}>()

const emit = defineEmits<{
  checkInCode: []
}>()
</script>

<template>
  <div class="joined-actions" @click.stop>
    <button
      v-if="showCheckInCode"
      type="button"
      class="joined-qr"
      aria-label="报名码"
      @click="emit('checkInCode')"
    >
      <svg class="joined-qr__icon" viewBox="0 0 24 24" aria-hidden="true">
        <path
          fill="currentColor"
          d="M3 3h8v8H3V3zm2 2v4h4V5H5zm8-2h8v8h-8V3zm2 2v4h4V5h-4zM3 13h8v8H3v-8zm2 2v4h4v-4H5zm13-2h3v2h-3v-2zm-5 0h3v2h-3v-2zm0 5h2v3h-3v-2h1v-1zm5 0h3v5h-5v-3h2v-2zm-3 3h2v2h-2v-2z"
        />
      </svg>
    </button>
    <PostListActionBar v-if="actions.length" :actions="actions" />
  </div>
</template>

<style scoped>
.joined-actions {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.joined-qr {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: var(--line-meta);
  margin: 0;
  padding: 0;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  -webkit-appearance: none;
  appearance: none;
}
.joined-qr::before {
  content: '';
  position: absolute;
  inset: -10px -6px;
}
.joined-qr:active {
  background: rgba(0, 0, 0, 0.05);
}
.joined-qr__icon {
  display: block;
  width: 18px;
  height: 18px;
}
</style>

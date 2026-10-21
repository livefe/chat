<script setup lang="ts">
defineProps<{
  label: string
  disabled: boolean
  hint: string
}>()

const emit = defineEmits<{
  submit: []
}>()
</script>

<template>
  <div class="submit">
    <button type="button" class="primary-pill-btn" :disabled="disabled" @click="emit('submit')">
      {{ label }}
    </button>
    <p v-if="hint" class="submit-hint">{{ hint }}</p>
  </div>
</template>

<style scoped>
.submit {
  flex-shrink: 0;
  padding: 12px 16px calc(8px + env(safe-area-inset-bottom, 0px));
  background: var(--page-bg);
}
.submit .primary-pill-btn {
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
}
.submit-hint {
  margin: 8px 0 0;
  text-align: center;
  font-size: 12px;
  color: var(--text-tertiary);
  line-height: 1.4;
}
</style>

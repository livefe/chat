<template>
  <div class="table-empty">
    <template v-if="error">
      <span>加载失败</span>
      <el-button link type="primary" @click="emit('retry')">重试</el-button>
    </template>
    <span v-else>{{ text }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{ error?: boolean; text: string }>()
const emit = defineEmits<{ retry: [] }>()
</script>

<style scoped>
.table-empty {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--el-text-color-secondary);
}
</style>

<template>
  <slot />
  <div v-if="hasMore" class="im-sidebar-more">
    <el-button link type="primary" :loading="loadingMore" @click="emit('load-more')">
      加载更多
    </el-button>
  </div>
  <div v-if="empty" class="im-sidebar-empty">
    <el-empty :description="emptyText" :image-size="72" />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  hasMore: boolean
  loadingMore: boolean
  empty: boolean
  emptyText: string
}>()

const emit = defineEmits<{ 'load-more': [] }>()
</script>

<style scoped>
.im-sidebar-more {
  display: flex;
  justify-content: center;
  padding: 8px 0 12px;
}
</style>

<script setup lang="ts">
import type { TiandituPoi } from '@/utils/map/tianditu'

defineProps<{
  loading: boolean
  results: TiandituPoi[]
}>()

const emit = defineEmits<{
  select: [poi: TiandituPoi]
}>()
</script>

<template>
  <div class="search-panel">
    <van-loading v-if="loading" size="20px" class="search-panel__loading">搜索中…</van-loading>
    <van-empty v-else-if="!results.length" description="未找到相关地点" />
    <button
      v-for="poi in results"
      :key="`${poi.lng}-${poi.lat}-${poi.name}`"
      type="button"
      class="search-item"
      @click="emit('select', poi)"
    >
      <span class="search-item__name">{{ poi.name }}</span>
      <span v-if="poi.address" class="search-item__addr">{{ poi.address }}</span>
    </button>
  </div>
</template>

<style scoped>
.search-panel {
  max-height: 38vh;
  overflow-y: auto;
  background: var(--card-bg);
  border-top: 1px solid rgba(0, 0, 0, 0.06);
}

.search-panel__loading {
  display: flex;
  justify-content: center;
  padding: 20px;
}

.search-item {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  width: 100%;
  padding: 12px 16px;
  border: none;
  border-bottom: 1px solid rgba(0, 0, 0, 0.06);
  background: transparent;
  text-align: left;
}

.search-item__name {
  font-size: 15px;
  color: #1f1f1f;
}

.search-item__addr {
  font-size: 12px;
  color: #999999;
}
</style>

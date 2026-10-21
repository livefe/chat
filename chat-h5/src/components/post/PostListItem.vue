<script setup lang="ts">
import type { PostItem } from '@/api/post'
import PostListItemRow from './PostListItemRow.vue'
import PostListItemWaterfall from './PostListItemWaterfall.vue'

withDefaults(
  defineProps<{
    post: PostItem
    variant?: 'list' | 'waterfall'

    badgeMode?: 'owner' | 'joined'
  }>(),
  {
    variant: 'list',
    badgeMode: 'owner',
  },
)
</script>

<template>
  <article class="post-item" :class="{ 'post-item--waterfall': variant === 'waterfall' }">
    <PostListItemWaterfall v-if="variant === 'waterfall'" :post="post" />
    <PostListItemRow v-else :post="post" :badge-mode="badgeMode">
      <template v-if="$slots.actions" #actions>
        <slot name="actions" />
      </template>
    </PostListItemRow>
  </article>
</template>

<style scoped>
.post-item {
  padding: 14px 16px 12px;
  border-bottom: 1px solid var(--border-color);
}

.post-item:active {
  background: rgba(0, 0, 0, 0.02);
}

.post-item--waterfall {
  display: block;
  position: relative;
  padding: 0;
  border-bottom: none;
  border-radius: 8px;
  overflow: hidden;
  background: var(--card-bg, #fff);
}

.post-item--waterfall:active {
  background: var(--card-bg, #fff);
  opacity: 0.96;
}
</style>

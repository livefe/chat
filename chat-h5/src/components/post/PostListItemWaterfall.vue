<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import UserAvatar from '@/components/UserAvatar.vue'
import type { PostItem } from '@/api/post'
import { resolveMediaUrl } from '@/utils/media'

const RATIO_MIN = 0.66
const RATIO_MAX = 3 / 4
const RATIO_DEFAULT = 3 / 4

const props = defineProps<{
  post: PostItem
}>()

const coverUrl = computed(() => resolveMediaUrl(props.post.images?.[0]))
const contentPreview = computed(() => (props.post.content || '').trim().replace(/\s+/g, ' '))
const discussCount = computed(() => Number(props.post.discussMessageCount) || 0)

const coverRatio = ref(RATIO_DEFAULT)
const coverReady = ref(false)

watch(coverUrl, () => {
  coverRatio.value = RATIO_DEFAULT
  coverReady.value = false
})

function onCoverLoad(e: Event) {
  const img = e.target as HTMLImageElement
  const w = img.naturalWidth
  const h = img.naturalHeight
  if (w > 0 && h > 0) {
    coverRatio.value = Math.min(RATIO_MAX, Math.max(RATIO_MIN, w / h))
  }
  coverReady.value = true
}
</script>

<template>
  <div v-if="coverUrl" class="post-item__cover" :style="{ aspectRatio: String(coverRatio) }">
    <img
      :src="coverUrl"
      alt=""
      class="post-item__cover-img"
      :class="{ 'post-item__cover-img--ready': coverReady }"
      loading="lazy"
      decoding="async"
      @load="onCoverLoad"
    />
  </div>
  <div v-else class="post-item__text-cover">
    <div class="post-item__text-cover-inner">
      <p class="post-item__text-cover-body">{{ contentPreview || post.title }}</p>
    </div>
  </div>
  <div class="post-item__caption">
    <h3 class="post-item__title">{{ post.title }}</h3>
    <div v-if="post.publisher || discussCount > 0" class="post-item__meta-row">
      <div v-if="post.publisher" class="post-item__author">
        <UserAvatar :src="post.publisher.avatarUrl" :size="18" tone="profile" />
        <span class="post-item__author-name">{{ post.publisher.nickname }}</span>
      </div>
      <span v-else class="post-item__author-spacer" />
      <span v-if="discussCount > 0" class="post-item__discuss" aria-label="讨论数">
        <svg class="post-item__discuss-icon" viewBox="0 0 24 24" aria-hidden="true">
          <path
            fill="currentColor"
            d="M20 2H4c-1.1 0-2 .9-2 2v18l4-4h14c1.1 0 2-.9 2-2V4c0-1.1-.9-2-2-2zm0 14H6l-2 2V4h16v12z"
          />
        </svg>
        <span class="post-item__discuss-num">{{ discussCount }}</span>
      </span>
    </div>
  </div>
</template>

<style scoped>
.post-item__cover {
  position: relative;
  margin-bottom: 0;
  border-radius: 0;
  overflow: hidden;
  background: linear-gradient(180deg, #f0f0f0 0%, #e5e5e5 100%);
}

.post-item__cover-img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
  opacity: 0;
  transition: opacity 0.22s ease;
}

.post-item__cover-img--ready {
  opacity: 1;
}

.post-item__caption {
  padding: 8px 10px 10px;
}

.post-item__title {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.4;
  color: var(--text-primary);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.post-item__meta-row {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
  margin-top: 6px;
}

.post-item__author {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  flex: 1;
}

.post-item__author-spacer {
  flex: 1;
  min-width: 0;
}

.post-item__author-name {
  font-size: 11px;
  line-height: 1.3;
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.post-item__discuss {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  flex-shrink: 0;
  color: var(--text-tertiary);
}

.post-item__discuss-icon {
  width: 15px;
  height: 15px;
  display: block;
}

.post-item__discuss-num {
  font-size: 12px;
  line-height: 1.2;
  font-variant-numeric: tabular-nums;
}

.post-item__text-cover {
  position: relative;
  box-sizing: border-box;
  aspect-ratio: 3 / 4;
  padding: 20px 18px;
  background: linear-gradient(165deg, #f0f0f0 0%, #e8e8e8 52%, #d8d8d8 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.post-item__text-cover-inner {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  width: 100%;
  min-height: 0;
}

.post-item__text-cover-body {
  margin: 0;
  width: 100%;
  font-size: 16px;
  line-height: 1.58;
  font-weight: 500;
  letter-spacing: 0.01em;
  color: #2a2a2a;
  text-align: center;
  font-family: 'PingFang SC', 'Hiragino Sans GB', 'Noto Sans SC', 'Microsoft YaHei', sans-serif;
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 6;
  overflow: hidden;
  word-break: break-word;
}
</style>

<script setup lang="ts">
import { computed, useSlots } from 'vue'

import type { PostItem } from '@/api/post'
import { PostStatus } from '@/constants/post'
import { resolveMediaUrl } from '@/utils/media'
import {
  formatPostCreatedAt,
  formatPostListFactGroups,
  joinedListCornerBadge,
  postStatusText,
} from '@/utils/post/post'
import { postCategoryOption } from '@/utils/post/postCategory'

const STATUS_TONE: Record<number, string> = {
  [PostStatus.Pending]: 'pending',
  [PostStatus.Published]: 'published',
  [PostStatus.Rejected]: 'rejected',
  [PostStatus.Disabled]: 'disabled',
}

const props = defineProps<{
  post: PostItem
  badgeMode: 'owner' | 'joined'
}>()

const slots = useSlots()
const coverUrl = computed(() => resolveMediaUrl(props.post.images?.[0]))
const imageCount = computed(() => props.post.images?.length ?? 0)
const categoryLabel = computed(() => postCategoryOption(props.post.category)?.label || '')
const timeLabel = computed(() => formatPostCreatedAt(props.post.createdAt))
const placeLabel = computed(() => (props.post.placeName || '').trim())
const hasListMeta = computed(() => !!(placeLabel.value || timeLabel.value))
const hasActions = computed(() => !!slots.actions)

const cornerBadge = computed(() => {
  if (props.badgeMode === 'joined') {
    return joinedListCornerBadge(props.post)
  }
  const status = props.post.status
  const statusLabel = postStatusText(status)
  const statusBadge = { text: statusLabel, tone: STATUS_TONE[status] }
  const nonPublished =
    status === PostStatus.Pending ||
    status === PostStatus.Rejected ||
    status === PostStatus.Disabled
  if (nonPublished && statusLabel) return statusBadge
  const phase = formatPostListFactGroups(props.post).phase
  if (phase) return { text: phase.text, tone: phase.tone }
  if (status === PostStatus.Published && statusLabel) return statusBadge
  return null
})
</script>

<template>
  <div class="post-item__head">
    <div v-if="coverUrl" class="post-item__thumb">
      <img :src="coverUrl" alt="" class="post-item__thumb-img" loading="lazy" decoding="async" />
      <span v-if="imageCount > 1" class="post-item__thumb-count">
        <span>{{ imageCount }}</span>
      </span>
    </div>
    <div v-else class="post-item__thumb post-item__thumb--empty">
      <span class="post-item__thumb-fallback">{{ categoryLabel.slice(0, 1) || '帖' }}</span>
    </div>

    <div class="post-item__content">
      <div class="post-item__top-row">
        <h3 class="post-item__title">{{ post.title }}</h3>
        <span
          v-if="cornerBadge"
          class="post-item__chip"
          :class="cornerBadge.tone ? `post-item__chip--${cornerBadge.tone}` : ''"
        >
          {{ cornerBadge.text }}
        </span>
      </div>
      <div v-if="hasListMeta || hasActions" class="post-item__bottom-row">
        <p v-if="hasListMeta" class="post-item__meta">
          <span v-if="placeLabel" class="post-item__meta-place">{{ placeLabel }}</span>
          <span v-if="placeLabel && timeLabel" class="post-item__meta-sep"> · </span>
          <span v-if="timeLabel" class="post-item__meta-time">{{ timeLabel }}</span>
        </p>
        <div v-if="hasActions" class="post-item__actions" @click.stop>
          <slot name="actions" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.post-item__head {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.post-item__thumb {
  position: relative;
  flex-shrink: 0;
  width: 88px;
  height: 88px;
  border-radius: 6px;
  overflow: hidden;
  background: #f0f0f0;
}

.post-item__thumb-count {
  position: absolute;
  right: 4px;
  bottom: 4px;
  box-sizing: border-box;
  width: 20px;
  height: 20px;
  padding: 0;
  border-radius: 50%;
  display: grid;
  place-items: center;
  background: rgba(0, 0, 0, 0.58);
  color: #fff;
  font-size: 11px;
  font-weight: 500;
}

.post-item__thumb-count > span {
  display: block;
  height: 11px;
  line-height: 11px;
  font-size: 11px;
  font-variant-numeric: tabular-nums;
}

.post-item__thumb--empty {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f0f0;
}

.post-item__thumb-fallback {
  font-size: 18px;
  font-weight: 600;
  color: var(--brand-color-dark);
  line-height: 1;
}

.post-item__thumb-img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.post-item__content {
  flex: 1;
  min-width: 0;
  min-height: 88px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 8px;
}

.post-item__top-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.post-item__bottom-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  min-width: 0;
}

.post-item__actions {
  flex-shrink: 0;
  margin-left: auto;
}

.post-item__title {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: var(--font-title);
  font-weight: 600;
  line-height: 1.35;
  color: var(--text-primary);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
  overflow: hidden;
}

.post-item__meta {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: var(--font-meta);
  line-height: var(--line-meta);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.post-item__meta-place {
  color: var(--text-secondary);
}

.post-item__meta-sep,
.post-item__meta-time {
  color: var(--text-tertiary);
}

.post-item__chip {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: var(--font-caption);
  line-height: 1.35;
  font-weight: 500;
  white-space: nowrap;
  background: rgba(0, 0, 0, 0.06);
  color: var(--text-tertiary);
}

.post-item__chip--pending {
  background: rgba(245, 158, 11, 0.12);
  color: #b45309;
}

.post-item__chip--rejected {
  background: rgba(224, 90, 82, 0.12);
  color: #c24740;
}

.post-item__chip--disabled,
.post-item__chip--phase-done {
  background: rgba(0, 0, 0, 0.06);
  color: var(--text-tertiary);
}

.post-item__chip--published {
  background: rgba(76, 141, 255, 0.12);
  color: var(--brand-color);
}

.post-item__chip--joined {
  background: rgba(76, 141, 255, 0.12);
  color: var(--brand-color);
}

.post-item__chip--waitlisted {
  background: rgba(245, 158, 11, 0.12);
  color: #b45309;
}

.post-item__chip--check-in-pending {
  background: rgba(76, 141, 255, 0.12);
  color: var(--brand-color-dark);
}

.post-item__chip--check-in-missed {
  background: rgba(0, 0, 0, 0.06);
  color: var(--text-tertiary);
}

.post-item__chip--checked-in {
  background: rgba(16, 185, 129, 0.12);
  color: #047857;
}

.post-item__chip--phase-pending {
  background: rgba(76, 141, 255, 0.12);
  color: var(--brand-color-dark);
}

.post-item__chip--phase-live {
  background: rgba(16, 185, 129, 0.12);
  color: #047857;
}
</style>

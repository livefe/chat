<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import type { MessageBubbleGestures } from '@/composables/useMessageBubbleMenu'
import type { ChatPostContent } from '@/utils/chat/chatPost'
import { resolveMediaUrl } from '@/utils/media'

const props = defineProps<{
  post: ChatPostContent
  gestures: MessageBubbleGestures
}>()

const router = useRouter()
const route = useRoute()

const postCoverSrc = computed(() => resolveMediaUrl(props.post.cover || ''))

function openPost(e?: Event) {
  if (props.gestures.consumeClickIfLongPressed(e)) return
  const id = props.post.postId
  if (!id) return
  if (route.name === 'PostDetail' && Number(route.params.id) === id) return
  void router.push(`/post/${id}`)
}
</script>

<template>
  <button type="button" class="post-btn" @click="openPost">
    <img v-if="postCoverSrc" class="post-btn__cover" :src="postCoverSrc" alt="" />
    <span v-else class="post-btn__cover post-btn__cover--empty" aria-hidden="true">动</span>
    <span class="post-btn__body">
      <span class="post-btn__name">{{ post.title }}</span>
      <span class="post-btn__hint">点击查看动态</span>
    </span>
  </button>
</template>

<style scoped>
.post-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  width: min(240px, 68vw);
  min-width: 0;
  padding: 10px 12px;
  border: none;
  background: transparent;
  text-align: left;
  cursor: pointer;
}

.post-btn__cover {
  flex-shrink: 0;
  width: 48px;
  height: 48px;
  border-radius: 6px;
  object-fit: cover;
  background: var(--page-bg);
}

.post-btn__cover--empty {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 600;
  color: var(--brand-color);
  background: var(--brand-soft);
}

.post-btn__body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.post-btn__name {
  font-size: 15px;
  font-weight: 500;
  line-height: 1.35;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  white-space: normal;
}

.post-btn__hint {
  font-size: 11px;
  line-height: 1.2;
  color: var(--text-secondary);
}
</style>

<script setup lang="ts">
import { computed } from 'vue'

import type { PostDetail } from '@/api/post'
import { PostStatus } from '@/constants/post'
import { canDiscussPost, isEventEnded, isSignupDeadlinePassed } from '@/utils/post/post'

const props = defineProps<{
  detail: PostDetail
  hasJoined: boolean
  isWaitlisted: boolean
  acting: boolean
  chatOpening: boolean
  chatPeekText: string
}>()

const emit = defineEmits<{
  openChat: []
  share: []
  signup: []
  cancelSignup: []
}>()

const needSignup = computed(() => !!props.detail.extras?.needSignup)
const isWithdrawn = computed(() => props.detail.status === PostStatus.Disabled)
const canDiscuss = computed(() => canDiscussPost(props.detail))
const cancelSignupLabel = computed(() => (props.isWaitlisted ? '取消候补' : '取消报名'))
const signupClosed = computed(
  () => needSignup.value && isSignupDeadlinePassed(props.detail.extras?.signupDeadlineAt || 0),
)
const eventEnded = computed(() => isEventEnded(props.detail.endAt || 0))
const signupUnavailableLabel = computed(() => {
  if (signupClosed.value) return '报名已截止'
  if (eventEnded.value) return '已结束'
  return ''
})
</script>

<template>
  <div class="footer-wrap">
    <div v-if="isWithdrawn" class="footer footer--withdrawn">动态已下架，无法报名或讨论</div>
    <template v-else>
      <button
        v-if="canDiscuss && chatPeekText"
        type="button"
        class="chat-peek"
        @click="emit('openChat')"
      >
        <span class="chat-peek__tag">讨论</span>
        <span class="chat-peek__text">{{ chatPeekText }}</span>
        <van-icon name="arrow-up" size="14" />
      </button>

      <div class="footer">
        <button type="button" class="share-btn" @click="emit('share')">
          <svg
            class="share-btn__icon"
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="currentColor"
            aria-hidden="true"
          >
            <path d="M10 9V5l-7 7 7 7v-4.1c5 0 8.5 1.6 11 5.1C19.8 14.8 15.8 9.8 10 9z" />
          </svg>
          <span>分享</span>
        </button>
        <div class="footer-actions">
          <button
            v-if="needSignup && hasJoined"
            type="button"
            class="action-btn action-btn--ghost"
            :disabled="acting"
            @click="emit('cancelSignup')"
          >
            {{ cancelSignupLabel }}
          </button>
          <button
            v-else-if="needSignup && signupUnavailableLabel"
            type="button"
            class="action-btn action-btn--ghost"
            disabled
          >
            {{ signupUnavailableLabel }}
          </button>
          <button
            v-else-if="needSignup"
            type="button"
            class="action-btn"
            :disabled="acting"
            @click="emit('signup')"
          >
            报名
          </button>
          <button
            v-if="canDiscuss"
            type="button"
            class="action-btn"
            :class="{ 'action-btn--ghost': needSignup }"
            :disabled="chatOpening"
            @click="emit('openChat')"
          >
            讨论
          </button>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.footer-wrap {
  flex-shrink: 0;
  background: var(--card-bg);
  border-top: 1px solid var(--border-color);
}

.chat-peek {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 10px 16px;
  border: none;
  border-bottom: 1px solid var(--border-color);
  background: rgba(0, 0, 0, 0.03);
  text-align: left;
}

.chat-peek__tag {
  flex-shrink: 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
  background: rgba(0, 0, 0, 0.06);
}

.chat-peek__text {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  line-height: 1.4;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 16px calc(12px + env(safe-area-inset-bottom, 0px));
}

.footer--withdrawn {
  justify-content: center;
  color: var(--text-tertiary);
  font-size: 13px;
}

.share-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  border: none;
  background: transparent;
  color: var(--text-primary);
  font-size: 14px;
  line-height: 1;
  padding: 10px 4px;
  flex-shrink: 0;
}

.share-btn__icon {
  display: block;
  flex-shrink: 0;
  transform: scaleX(-1);
}

.footer-actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
  min-width: 0;
}

.action-btn {
  border: none;
  background: var(--brand-color);
  color: #fff;
  font-size: 14px;
  line-height: 1;
  padding: 10px 14px;
  border-radius: var(--radius-lg);
  white-space: nowrap;
}

.action-btn--ghost {
  background: transparent;
  color: var(--text-primary);
  border: 1px solid var(--border-color);
}

.action-btn:disabled {
  opacity: 0.5;
}
</style>

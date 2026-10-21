<script setup lang="ts">
import { computed } from 'vue'

import type { PostDetail } from '@/api/post'
import {
  formatPostFeeBrief,
  formatPostScheduleRange,
  formatSignupPeopleDetail,
  formatUnixTime,
  hasPostFee,
} from '@/utils/post/post'
import { PostCategory } from '@/utils/post/postCategory'

const props = defineProps<{
  detail: PostDetail
}>()

const extras = computed(() => props.detail.extras)
const scheduleRange = computed(() =>
  formatPostScheduleRange(props.detail.startAt, props.detail.endAt),
)
const hasSchedule = computed(() => !!(scheduleRange.value?.start || scheduleRange.value?.end))
const needSignup = computed(() => !!extras.value?.needSignup)
const hasFee = computed(() => !!extras.value && hasPostFee(extras.value))
const isReward = computed(() => extras.value?.feeType === 2)
const feeLabel = computed(() => (extras.value ? formatPostFeeBrief(extras.value) : ''))

const idlePriceLabel = computed(() => {
  if (props.detail.category !== PostCategory.Idle) return ''
  const price = Number(extras.value?.amount) || 0
  if (price <= 0 && !extras.value?.negotiable) return ''
  let text = price > 0 ? `¥${price}` : '面议'
  if (extras.value?.negotiable && price > 0) text += ' · 可议价'
  return text
})

const signupPeopleText = computed(() => {
  const ex = extras.value
  if (!ex?.needSignup) return ''
  return formatSignupPeopleDetail(ex, props.detail.signupSummary)
})

const signupLimitText = computed(() => {
  const ex = extras.value
  if (!ex?.maxParticipants) return ''
  return ex.allowWaitlist ? `限 ${ex.maxParticipants} 人，可候补` : `限 ${ex.maxParticipants} 人`
})

const signupDeadlineText = computed(() => {
  const deadline = extras.value?.signupDeadlineAt || 0
  return deadline > 0 ? formatUnixTime(deadline) : ''
})

const hasExtraSection = computed(
  () => hasSchedule.value || hasFee.value || needSignup.value || !!idlePriceLabel.value,
)
</script>

<template>
  <section v-if="hasExtraSection" class="detail-card">
    <div class="extra-group">
      <div v-if="idlePriceLabel" class="extra-block">
        <div class="extra-row">
          <span class="extra-label">价格</span>
          <span class="extra-value">{{ idlePriceLabel }}</span>
        </div>
      </div>

      <div v-if="hasSchedule && scheduleRange" class="extra-block">
        <div v-if="scheduleRange.start" class="extra-row">
          <span class="extra-label">{{ scheduleRange.end ? '开始' : '时间' }}</span>
          <span class="extra-value">{{ scheduleRange.start }}</span>
        </div>
        <div v-if="scheduleRange.end" class="extra-row">
          <span class="extra-label">{{ scheduleRange.start ? '结束' : '时间' }}</span>
          <span class="extra-value">{{ scheduleRange.end }}</span>
        </div>
      </div>

      <div v-if="needSignup" class="extra-block">
        <div v-if="signupLimitText" class="extra-row">
          <span class="extra-label">名额</span>
          <span class="extra-value">{{ signupLimitText }}</span>
        </div>
        <div v-if="signupDeadlineText" class="extra-row">
          <span class="extra-label">截止</span>
          <span class="extra-value">{{ signupDeadlineText }}</span>
        </div>
        <div class="extra-row">
          <span class="extra-label">报名</span>
          <span class="extra-value">{{ signupPeopleText }}</span>
        </div>
      </div>

      <div v-if="hasFee && extras" class="extra-block">
        <div class="extra-row">
          <span class="extra-label">金额</span>
          <span
            class="extra-value extra-value--strong"
            :class="isReward ? 'extra-value--reward' : 'extra-value--charge'"
          >
            {{ feeLabel }}
          </span>
        </div>
        <p v-if="extras.feeDescription" class="extra-note">
          {{ extras.feeDescription }}
        </p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.detail-card {
  padding: 16px;
  border-radius: var(--radius-md);
  background: var(--card-bg);
}

.extra-group {
  display: flex;
  flex-direction: column;
}

.extra-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 0;
}

.extra-block:first-child {
  padding-top: 0;
}

.extra-block:last-child {
  padding-bottom: 0;
}

.extra-block + .extra-block {
  border-top: 1px solid var(--border-color);
}

.extra-row {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.extra-label {
  flex-shrink: 0;
  width: 40px;
  padding-top: 1px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--text-tertiary);
}

.extra-value {
  flex: 1;
  min-width: 0;
  font-size: 15px;
  line-height: 1.5;
  color: var(--text-primary);
  word-break: break-word;
}

.extra-value--strong {
  font-weight: 600;
}

.extra-value--charge {
  color: var(--fee-color);
}

.extra-value--reward {
  color: var(--reward-color);
}

.extra-note {
  margin: 0 0 0 52px;
  font-size: 13px;
  line-height: 1.55;
  color: var(--text-secondary);
  word-break: break-word;
}
</style>

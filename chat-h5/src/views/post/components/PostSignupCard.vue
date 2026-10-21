<script setup lang="ts">
import { computed } from 'vue'

import type { PostSignupItem } from '@/api/post'
import { SignupStatus } from '@/constants/post'
import { hasCheckInTime, signupCheckInText, signupListStatusText } from '@/utils/post/post'

const props = defineProps<{
  item: PostSignupItem
  needCheckIn: boolean
  canCheckIn: boolean
  checking: boolean
}>()

const emit = defineEmits<{
  editRemark: []
  checkIn: []
}>()

const confirmedWithCheckIn = computed(
  () => props.needCheckIn && props.item.signupStatus === SignupStatus.Confirmed,
)

const statusClass = computed(() => {
  const status = props.item.signupStatus
  if (status === SignupStatus.Waitlisted) return 'waitlisted'
  if (status === SignupStatus.Cancelled) return 'cancelled'
  if (confirmedWithCheckIn.value) return hasCheckInTime(props.item.checkedInAt) ? 'arrived' : 'none'
  if (status === SignupStatus.Confirmed) return 'confirmed'
  return ''
})

const statusLabel = computed(() =>
  confirmedWithCheckIn.value
    ? signupCheckInText(props.item.checkedInAt)
    : signupListStatusText(props.item),
)

const remark = computed(() => (props.item.remark || '').trim())
</script>

<template>
  <div class="signup-card">
    <div class="signup-card__head">
      <span class="signup-card__name">{{ item.userNickname }}</span>
      <span v-if="statusLabel" class="signup-card__status" :class="statusClass">
        {{ statusLabel }}
      </span>
    </div>
    <div class="signup-card__meta">
      <span>报名 {{ item.createdAt }}</span>
      <span v-if="needCheckIn && item.checkedInAt">核验 {{ item.checkedInAt }}</span>
      <span v-if="item.cancelledAt">取消 {{ item.cancelledAt }}</span>
    </div>
    <button
      type="button"
      class="signup-card__remark"
      :class="{ 'signup-card__remark--empty': !remark }"
      @click="emit('editRemark')"
    >
      <template v-if="remark">备注 {{ item.remark }}</template>
      <template v-else>添加备注</template>
    </button>
    <div v-if="canCheckIn" class="signup-card__actions">
      <button
        type="button"
        class="confirm-btn confirm-btn--on"
        :disabled="checking"
        @click="emit('checkIn')"
      >
        {{ checking ? '核验中' : '手动核验' }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.signup-card {
  padding: 14px 16px;
  border-radius: 12px;
  background: var(--card-bg);
  box-shadow: none;
}

.signup-card__head {
  display: flex;
  align-items: center;
  gap: 12px;
}

.signup-card__name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
}

.signup-card__status {
  flex-shrink: 0;
  font-size: 12px;
  line-height: 1.4;
  padding: 2px 8px;
  border-radius: 999px;
}

.signup-card__status.confirmed,
.signup-card__status.arrived {
  color: #047857;
  background: rgba(16, 185, 129, 0.12);
}

.signup-card__status.none,
.signup-card__status.cancelled {
  color: var(--text-tertiary);
  background: rgba(0, 0, 0, 0.08);
}

.signup-card__status.waitlisted {
  color: #c2410c;
  background: rgba(255, 151, 106, 0.14);
}

.signup-card__meta {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--text-tertiary);
}

.signup-card__remark {
  display: block;
  width: 100%;
  margin-top: 8px;
  border: none;
  background: transparent;
  padding: 0;
  text-align: left;
  font: inherit;
  font-size: 12px;
  line-height: 1.45;
  color: var(--text-tertiary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}

.signup-card__remark--empty {
  color: var(--text-muted, var(--text-tertiary));
}

.signup-card__actions {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

.confirm-btn {
  border: 1px solid var(--border-color);
  border-radius: 999px;
  padding: 6px 14px;
  background: transparent;
  color: var(--text-tertiary);
  font: inherit;
  font-size: 13px;
  line-height: 1.2;
}

.confirm-btn--on {
  border: none;
  background: var(--brand-color);
  color: #fff;
}

.confirm-btn:disabled {
  opacity: 0.55;
}
</style>

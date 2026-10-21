<script setup lang="ts">
import { computed } from 'vue'

import type { PostSignupListCounts } from '@/api/post'
import { SignupCheckIn, SignupStatus } from '@/constants/post'

const props = defineProps<{
  counts: PostSignupListCounts
  status: number | undefined
  checkIn: number | undefined
  needCheckIn: boolean
}>()

const emit = defineEmits<{
  change: [status: number | undefined, checkIn: number | undefined]
}>()

interface SignupTab {
  key: string
  label: string
  count: number
  status?: number
  checkIn?: number
}

const statusTabs = computed<SignupTab[]>(() => {
  const c = props.counts
  return [
    { key: 'all', label: '全部', count: c.total },
    { key: 'confirmed', label: '已报名', count: c.confirmed, status: SignupStatus.Confirmed },
    { key: 'waitlisted', label: '候补', count: c.waitlisted, status: SignupStatus.Waitlisted },
    { key: 'cancelled', label: '已取消', count: c.cancelled, status: SignupStatus.Cancelled },
  ]
})

const checkInTabs = computed<SignupTab[]>(() => {
  const c = props.counts
  return [
    {
      key: 'notChecked',
      label: '未核验',
      count: c.notCheckedIn,
      checkIn: SignupCheckIn.NotChecked,
    },
    { key: 'checked', label: '已核验', count: c.checkedIn, checkIn: SignupCheckIn.Checked },
  ]
})

const showCheckInTabs = computed(() => props.needCheckIn && props.status === SignupStatus.Confirmed)

function onStatusClick(tab: SignupTab) {
  if (props.status === tab.status && props.checkIn === undefined) return
  emit('change', tab.status, undefined)
}

function onCheckInClick(tab: SignupTab) {
  const next = props.checkIn === tab.checkIn ? undefined : tab.checkIn
  emit('change', SignupStatus.Confirmed, next)
}
</script>

<template>
  <div class="signup-filters">
    <div class="signup-tabs">
      <button
        v-for="tab in statusTabs"
        :key="tab.key"
        type="button"
        class="signup-tab"
        :class="{ 'signup-tab--on': status === tab.status }"
        @click="onStatusClick(tab)"
      >
        <span class="signup-tab__name">{{ tab.label }}</span>
        <span class="signup-tab__count">{{ tab.count }}</span>
      </button>
    </div>
    <div v-if="showCheckInTabs" class="signup-checkin">
      <button
        v-for="tab in checkInTabs"
        :key="tab.key"
        type="button"
        class="signup-checkin__chip"
        :class="{ 'signup-checkin__chip--on': checkIn === tab.checkIn }"
        @click="onCheckInClick(tab)"
      >
        {{ tab.label }}
        <span>{{ tab.count }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.signup-filters {
  flex-shrink: 0;
  background: var(--card-bg);
}

.signup-tabs {
  display: flex;
  align-items: stretch;
  padding: 0 4px;
  border-bottom: 1px solid var(--border-color);
}

.signup-tab {
  flex: 1;
  position: relative;
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 4px;
  border: none;
  background: transparent;
  padding: 12px 0 10px;
  font: inherit;
  font-size: 15px;
  line-height: 1.2;
  color: var(--text-muted);
}

.signup-tab--on {
  font-weight: 600;
  color: var(--text-primary);
}

.signup-tab--on::after {
  content: '';
  position: absolute;
  left: 50%;
  bottom: 0;
  width: 18px;
  height: 3px;
  border-radius: 3px;
  background: var(--brand-color);
  transform: translateX(-50%);
}

.signup-tab__count {
  font-size: 12px;
  font-weight: 400;
  color: var(--text-tertiary);
}

.signup-tab--on .signup-tab__count {
  color: var(--brand-color);
}

.signup-checkin {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  overflow-x: auto;
}

.signup-checkin__chip {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  border: 1px solid var(--border-color-soft);
  border-radius: 999px;
  padding: 5px 10px;
  background: transparent;
  font: inherit;
  font-size: 12px;
  line-height: 1.2;
  color: var(--text-secondary);
}

.signup-checkin__chip span {
  color: var(--text-tertiary);
}

.signup-checkin__chip--on {
  border-color: var(--brand-color);
  background: rgba(76, 141, 255, 0.08);
  color: var(--brand-color-dark);
  font-weight: 600;
}

.signup-checkin__chip--on span {
  color: var(--brand-color);
}
</style>

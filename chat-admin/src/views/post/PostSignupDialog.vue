<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="1080px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <div class="signup-filters">
      <div class="signup-tabs">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          type="button"
          class="signup-tab"
          :class="{ 'is-on': isStatusOn(tab) }"
          @click="onStatusClick(tab)"
        >
          <span>{{ tab.label }}</span>
          <span class="signup-tab__count">{{ tab.count }}</span>
        </button>
      </div>
      <div v-if="showCheckInTabs" class="signup-checkin">
        <button
          v-for="tab in checkInTabs"
          :key="tab.key"
          type="button"
          class="signup-checkin__chip"
          :class="{ 'is-on': isCheckInOn(tab) }"
          @click="onCheckInClick(tab)"
        >
          {{ tab.label }} {{ tab.count }}
        </button>
      </div>
    </div>
    <el-table :data="list" v-loading="loading" max-height="420">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column label="用户" min-width="160">
        <template #default="{ row }">
          <div>{{ row.userNickname || '-' }}</div>
          <div class="sub">{{ row.userMobile || `ID ${row.userId}` }}</div>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="signupStatusTag(row.signupStatus)">{{ row.signupStatusText }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="报名时间" width="170" />
      <el-table-column v-if="needCheckIn" label="核验" width="90">
        <template #default="{ row }">
          <el-tag :type="signupCheckInTag(row as PostSignupItem)">{{
            signupCheckInText(row as PostSignupItem)
          }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column v-if="needCheckIn" prop="checkedInAt" label="核验时间" width="170">
        <template #default="{ row }">{{ row.checkedInAt || '-' }}</template>
      </el-table-column>
      <el-table-column prop="cancelledAt" label="取消时间" width="170">
        <template #default="{ row }">{{ row.cancelledAt || '-' }}</template>
      </el-table-column>
      <el-table-column prop="remark" label="备注" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ row.remark || '-' }}</template>
      </el-table-column>
    </el-table>
    <div class="pagination">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="PAGE_SIZES"
        :layout="PAGINATION_LAYOUT"
        @size-change="emit('fetch')"
        @current-change="emit('fetch')"
      />
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'

import type { PostSignupItem, PostSignupListCounts } from '@/api/post'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { SignupCheckIn, SignupStatus } from '@/constants/post'
import {
  emptySignupListCounts,
  signupCheckInTag,
  signupCheckInText,
  signupStatusTag,
} from '@/utils/postDisplay'

interface SignupTab {
  key: string
  label: string
  count: number
  status?: number
  checkIn?: number
}

const visible = defineModel<boolean>({ required: true })
const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })
const status = defineModel<number>('status')
const checkIn = defineModel<number>('checkIn')

const props = defineProps<{
  title: string
  list: PostSignupItem[]
  loading: boolean
  total: number
  needCheckIn?: boolean
  counts?: PostSignupListCounts
}>()

const emit = defineEmits<{ fetch: [] }>()

const tabs = computed<SignupTab[]>(() => {
  const c = props.counts ?? emptySignupListCounts()
  return [
    { key: 'all', label: '全部', count: c.total },
    { key: 'confirmed', label: '已报名', count: c.confirmed, status: SignupStatus.Confirmed },
    { key: 'waitlisted', label: '候补', count: c.waitlisted, status: SignupStatus.Waitlisted },
    { key: 'cancelled', label: '已取消', count: c.cancelled, status: SignupStatus.Cancelled },
  ]
})

const checkInTabs = computed<SignupTab[]>(() => {
  const c = props.counts ?? emptySignupListCounts()
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

const showCheckInTabs = computed(
  () => !!props.needCheckIn && status.value === SignupStatus.Confirmed,
)

function isStatusOn(tab: SignupTab) {
  return status.value === tab.status
}

function isCheckInOn(tab: SignupTab) {
  return checkIn.value === tab.checkIn
}

function onStatusClick(tab: SignupTab) {
  if (status.value === tab.status && checkIn.value === undefined) return
  status.value = tab.status
  checkIn.value = undefined
  page.value = 1
  emit('fetch')
}

function onCheckInClick(tab: SignupTab) {
  const next = checkIn.value === tab.checkIn ? undefined : tab.checkIn
  status.value = SignupStatus.Confirmed
  checkIn.value = next
  page.value = 1
  emit('fetch')
}
</script>

<style scoped>
.signup-filters {
  margin-bottom: 12px;
}

.signup-tabs {
  display: flex;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.signup-tab {
  flex: 1;
  position: relative;
  display: inline-flex;
  align-items: baseline;
  justify-content: center;
  gap: 4px;
  border: none;
  background: transparent;
  padding: 10px 0 12px;
  font: inherit;
  font-size: 14px;
  color: var(--el-text-color-secondary);
  cursor: pointer;
}

.signup-tab.is-on {
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.signup-tab.is-on::after {
  content: '';
  position: absolute;
  left: 50%;
  bottom: 0;
  width: 18px;
  height: 3px;
  border-radius: 3px;
  background: var(--el-color-primary);
  transform: translateX(-50%);
}

.signup-tab__count {
  font-size: 12px;
  font-weight: 400;
  color: var(--el-text-color-placeholder);
}

.signup-tab.is-on .signup-tab__count {
  color: var(--el-color-primary);
}

.signup-checkin {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-top: 10px;
}

.signup-checkin__chip {
  border: 1px solid var(--el-border-color);
  border-radius: 999px;
  padding: 4px 10px;
  background: transparent;
  font: inherit;
  font-size: 12px;
  color: var(--el-text-color-regular);
  cursor: pointer;
}

.signup-checkin__chip.is-on {
  border-color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-weight: 600;
}
</style>

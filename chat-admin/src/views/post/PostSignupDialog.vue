<template>
  <el-dialog
    v-model="visible"
    :title="post ? `报名名单 - ${post.title}` : '报名名单'"
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
      <div class="signup-toolbar">
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
        <div class="signup-search">
          <el-input
            v-model="userIdText"
            placeholder="用户ID"
            clearable
            style="width: 110px"
            @keyup.enter="handleSearch"
          />
          <el-input
            v-model="query.keyword"
            placeholder="昵称/手机号"
            clearable
            style="width: 150px"
            @keyup.enter="handleSearch"
          />
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button :loading="exporting" :disabled="!total" @click="exportCsv">导出</el-button>
        </div>
      </div>
    </div>
    <el-table :data="list" v-loading="loading" max-height="420" @sort-change="onSortChange">
      <template #empty>
        <TableEmpty :error="error" text="暂无报名记录" @retry="fetchList" />
      </template>
      <el-table-column label="用户" min-width="180">
        <template #default="{ row }">
          <UserCell :user-id="row.userId" :nickname="row.userNickname" :mobile="row.userMobile" />
        </template>
      </el-table-column>
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="signupStatusTag(row.signupStatus)">{{ row.signupStatusText }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="报名时间" width="170" sortable="custom" />
      <el-table-column v-if="needCheckIn" label="核验" width="90">
        <template #default="{ row }">
          <el-tag :type="signupCheckInTag(row as PostSignupItem)">{{
            signupCheckInText(row as PostSignupItem)
          }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column
        v-if="needCheckIn"
        prop="checkedInAt"
        label="核验时间"
        width="170"
        sortable="custom"
      >
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
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        :page-sizes="PAGE_SIZES"
        :layout="PAGINATION_LAYOUT"
        @size-change="fetchList"
        @current-change="fetchList"
      />
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'

import {
  getPostSignupList,
  type PostItem,
  type PostSignupItem,
  type PostSignupListCounts,
  type PostSignupListRequest,
  type SignupSortField,
} from '@/api/post'
import TableEmpty from '@/components/common/TableEmpty.vue'
import UserCell from '@/components/common/UserCell.vue'
import { useListQuery } from '@/composables/useListQuery'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { SignupCheckIn, SignupStatus } from '@/constants/post'
import { showApiError } from '@/utils/apiError'
import {
  downloadCsv,
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

const EXPORT_PAGE_SIZE = 100
const EXPORT_MAX_ROWS = 10000

const visible = defineModel<boolean>({ required: true })
const props = defineProps<{ post: Pick<PostItem, 'id' | 'title'> | null }>()

const needCheckIn = ref(false)
const counts = ref<PostSignupListCounts>(emptySignupListCounts())
const exporting = ref(false)

const defaults = {
  page: 1,
  pageSize: 20,
  status: undefined as number | undefined,
  checkIn: undefined as number | undefined,
  userId: undefined as number | undefined,
  keyword: '',
  sortField: undefined as SignupSortField | undefined,
  sortOrder: undefined as 'asc' | 'desc' | undefined,
}

function filterParams(q: typeof defaults): Omit<PostSignupListRequest, 'page' | 'pageSize'> {
  return {
    postId: props.post?.id ?? 0,
    status: q.status,
    checkIn: q.checkIn,
    userId: q.userId,
    keyword: q.keyword,
    sortField: q.sortField,
    sortOrder: q.sortOrder,
  }
}

const { query, list, total, loading, error, fetchList, handleSearch } = useListQuery({
  fetch: (q) =>
    getPostSignupList({ ...filterParams(q), page: q.page, pageSize: q.pageSize }).then((res) => {
      needCheckIn.value = !!res.data.needCheckIn
      counts.value = res.data.counts ?? emptySignupListCounts()
      return res
    }),
  defaults,
  errorMessage: '获取报名名单失败',
  syncUrl: false,
  immediate: false,
})

const userIdText = computed({
  get: () => (query.userId ? String(query.userId) : ''),
  set: (v: string) => {
    const n = Number(v.trim())
    query.userId = Number.isInteger(n) && n > 0 ? n : undefined
  },
})

watch(
  () => [visible.value, props.post?.id] as const,
  ([open, postId]) => {
    if (!open || !postId) return
    Object.assign(query, { ...defaults, pageSize: query.pageSize })
    counts.value = emptySignupListCounts()
    void fetchList()
  },
  { immediate: true },
)

const tabs = computed<SignupTab[]>(() => {
  const c = counts.value
  return [
    { key: 'all', label: '全部', count: c.total },
    { key: 'confirmed', label: '已报名', count: c.confirmed, status: SignupStatus.Confirmed },
    { key: 'waitlisted', label: '候补', count: c.waitlisted, status: SignupStatus.Waitlisted },
    { key: 'cancelled', label: '已取消', count: c.cancelled, status: SignupStatus.Cancelled },
  ]
})

const checkInTabs = computed<SignupTab[]>(() => {
  const c = counts.value
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

const showCheckInTabs = computed(() => needCheckIn.value && query.status === SignupStatus.Confirmed)

function isStatusOn(tab: SignupTab) {
  return query.status === tab.status
}

function isCheckInOn(tab: SignupTab) {
  return query.checkIn === tab.checkIn
}

function onStatusClick(tab: SignupTab) {
  if (query.status === tab.status && query.checkIn === undefined) return
  query.status = tab.status
  query.checkIn = undefined
  void handleSearch()
}

function onCheckInClick(tab: SignupTab) {
  query.status = SignupStatus.Confirmed
  query.checkIn = query.checkIn === tab.checkIn ? undefined : tab.checkIn
  void handleSearch()
}

function onSortChange({ prop, order }: { prop: string | null; order: string | null }) {
  if (!prop || !order) {
    query.sortField = undefined
    query.sortOrder = undefined
  } else {
    query.sortField = prop as SignupSortField
    query.sortOrder = order === 'ascending' ? 'asc' : 'desc'
  }
  void handleSearch()
}

function timestamp() {
  const d = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}${pad(d.getHours())}${pad(d.getMinutes())}`
}

async function exportCsv() {
  if (!props.post) return
  exporting.value = true
  try {
    const params = filterParams(query)
    const rows: PostSignupItem[] = []
    for (let page = 1; rows.length < EXPORT_MAX_ROWS; page++) {
      const { data } = await getPostSignupList({ ...params, page, pageSize: EXPORT_PAGE_SIZE })
      rows.push(...data.list)
      if (data.list.length < EXPORT_PAGE_SIZE || rows.length >= data.total) break
    }
    downloadCsv(
      `报名名单_${props.post.id}_${timestamp()}.csv`,
      ['用户ID', '昵称', '手机号', '状态', '报名时间', '核验', '核验时间', '取消时间', '备注'],
      rows.map((r) => [
        r.userId,
        r.userNickname,
        r.userMobile,
        r.signupStatusText,
        r.createdAt,
        signupCheckInText(r),
        r.checkedInAt || '',
        r.cancelledAt || '',
        r.remark || '',
      ]),
    )
    if (rows.length >= EXPORT_MAX_ROWS) {
      ElMessage.warning(`仅导出前 ${EXPORT_MAX_ROWS} 条，请缩小筛选范围`)
    }
  } catch (e) {
    showApiError(e, '导出失败')
  } finally {
    exporting.value = false
  }
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

.signup-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding-top: 10px;
}

.signup-checkin {
  display: flex;
  align-items: center;
  gap: 8px;
}

.signup-search {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
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

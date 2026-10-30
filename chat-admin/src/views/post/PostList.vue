<template>
  <div class="admin-page post-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <el-button v-if="pendingCount" type="warning" plain @click="showPending">
              待审核 {{ pendingCount }}
            </el-button>
            <el-button @click="mapVisible = true">全景地图</el-button>
          </div>
        </div>
      </template>

      <PostFilter
        v-model:id="query.id"
        v-model:keyword="query.keyword"
        v-model:publisher-id="query.publisherId"
        v-model:publisher-keyword="query.publisherKeyword"
        v-model:room-id="query.roomId"
        v-model:category="query.category"
        v-model:status="query.status"
        v-model:start-time="query.startTime"
        v-model:end-time="query.endTime"
        v-model:sort="query.sort"
        v-model:include-deleted="query.includeDeleted"
        @search="handleSearch"
        @reset="handleReset"
      />

      <div v-if="canBatch && selection.length" class="batch-bar">
        <span>已选 {{ selection.length }} 条</span>
        <el-button
          v-for="act in BATCH_ACTIONS"
          :key="act.status"
          size="small"
          :type="act.type"
          @click="batchStatus(act)"
        >
          批量{{ act.label }}
        </el-button>
      </div>

      <PostTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        :selectable="canBatch"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @detail="openDetail"
        @signup="openSignupList"
        @messages="goPostMessages"
        @go-room="goRoom"
        @status="onTableStatus"
        @selection-change="selection = $event"
        @page-change="fetchList"
      />
    </el-card>

    <PostDetailDrawer
      v-model="detailVisible"
      :current="current"
      @signup="openSignupList"
      @messages="goPostMessages"
      @go-room="goRoom"
      @filter-publisher="filterByPublisher"
      @status="onDrawerStatus"
    />

    <PostSignupDialog v-model="signupVisible" :post="signupPost" />

    <PostMapDialog
      v-model:visible="mapVisible"
      @filter-room="filterByRoom"
      @open-post="(id) => openDetail({ id })"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

import {
  batchUpdatePostStatus,
  getPostDetail,
  getPostList,
  updatePostStatus,
  type PostDetail,
  type PostItem,
} from '@/api/post'
import { useAdminBadges } from '@/composables/useAdminBadges'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { MessageScope } from '@/constants/message'
import { PERM } from '@/constants/permission'
import { PostStatus } from '@/constants/post'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { pushMenuRoute } from '@/utils/navigation'
import { postStatusActionLabel } from '@/utils/postDisplay'
import { confirmWithReason } from '@/utils/promptReason'
import PostDetailDrawer from './PostDetailDrawer.vue'
import PostFilter from './PostFilter.vue'
import PostMapDialog from './PostMapDialog.vue'
import PostSignupDialog from './PostSignupDialog.vue'
import PostTable from './PostTable.vue'

interface BatchAction {
  status: PostStatus
  from: PostStatus
  label: string
  type: 'success' | 'danger' | 'warning'
}

const BATCH_ACTIONS: BatchAction[] = [
  { status: PostStatus.Published, from: PostStatus.Pending, label: '通过', type: 'success' },
  { status: PostStatus.Rejected, from: PostStatus.Pending, label: '驳回', type: 'danger' },
  { status: PostStatus.Disabled, from: PostStatus.Published, label: '下架', type: 'warning' },
]

const router = useRouter()
const route = useRoute()
const employeeStore = useEmployeeStore()
const { can } = usePermission()
const { badgeFor, refreshBadges } = useAdminBadges()

const mapVisible = ref(false)
const selection = ref<PostItem[]>([])
const canBatch = computed(() => can(PERM.PostPostStatus))
const pendingCount = computed(() => badgeFor(route.path) ?? 0)

const { query, list, total, loading, error, fetchList, handleSearch, handleReset } = useListQuery({
  fetch: (q) => getPostList(q),
  defaults: {
    page: 1,
    pageSize: 20,
    id: undefined as number | undefined,
    keyword: '',
    publisherId: undefined as number | undefined,
    publisherKeyword: '',
    roomId: undefined as number | undefined,
    category: undefined as string | undefined,
    status: undefined as number | undefined,
    startTime: '',
    endTime: '',
    sort: 'desc' as 'asc' | 'desc',
    includeDeleted: false,
  },
  schema: {
    id: 'number',
    keyword: 'string',
    publisherId: 'number',
    publisherKeyword: 'string',
    roomId: 'number',
    category: 'string',
    status: 'number',
    startTime: 'string',
    endTime: 'string',
    sort: 'string',
    includeDeleted: 'boolean',
  },
  errorMessage: '获取帖子列表失败',
})

function replaceFilters(filters: Record<string, string>) {
  void router.replace({ path: route.path, query: filters })
}

function showPending() {
  replaceFilters({ status: String(PostStatus.Pending), sort: 'asc' })
}

function filterByRoom(roomId: number) {
  replaceFilters({ roomId: String(roomId) })
}

function filterByPublisher(userId: number) {
  detailVisible.value = false
  replaceFilters({ publisherId: String(userId) })
}

const detailVisible = ref(false)
const current = ref<PostDetail | null>(null)
let detailSeq = 0

async function openDetail(row: Pick<PostItem, 'id'> & Partial<PostItem>) {
  const seq = ++detailSeq
  const preview = row.title !== undefined
  if (preview) {
    current.value = row as PostDetail
    detailVisible.value = true
  }
  try {
    const { data } = await getPostDetail(row.id)
    if (seq !== detailSeq) return
    current.value = data
    detailVisible.value = true
  } catch (e) {
    if (!preview) showApiError(e, '获取详情失败')
  }
}

let autoOpenedId = 0
watch(list, (rows) => {
  if (!query.id) {
    autoOpenedId = 0
    return
  }
  const only = rows.length === 1 ? rows[0] : undefined
  if (only && only.id === query.id && autoOpenedId !== only.id) {
    autoOpenedId = only.id
    void openDetail(only)
  }
})

const signupVisible = ref(false)
const signupPost = ref<Pick<PostItem, 'id' | 'title'> | null>(null)

function openSignupList(row: Pick<PostItem, 'id' | 'title'>) {
  signupPost.value = { id: row.id, title: row.title }
  signupVisible.value = true
}

function needsReason(status: PostStatus) {
  return status === PostStatus.Rejected || status === PostStatus.Disabled
}

async function confirmStatusChange(
  status: PostStatus,
  action: string,
  message: string,
): Promise<string | null> {
  if (needsReason(status)) {
    return confirmWithReason(message, {
      title: `${action}帖子`,
      required: true,
      inputPlaceholder: `请输入${action}原因（作者可见）`,
    })
  }
  try {
    await ElMessageBox.confirm(message, `${action}帖子`, { type: 'warning' })
    return ''
  } catch {
    return null
  }
}

async function changeStatus(row: PostItem, status: PostStatus): Promise<boolean> {
  const action = postStatusActionLabel(row.status, status)
  const reason = await confirmStatusChange(status, action, `确定${action}「${row.title}」？`)
  if (reason === null) return false
  try {
    await updatePostStatus(row.id, status, reason)
  } catch (e) {
    showApiError(e, `${action}失败`)
    return false
  }
  ElMessage.success(`已${action}`)
  void refreshBadges({ force: true })
  await fetchList()
  return true
}

async function onTableStatus(row: PostItem, status: PostStatus) {
  if (!(await changeStatus(row, status))) return
  if (detailVisible.value && current.value?.id === row.id) await openDetail({ id: row.id })
}

async function onDrawerStatus(row: PostItem, status: PostStatus) {
  const wasPending = row.status === PostStatus.Pending
  if (!(await changeStatus(row, status))) return
  if (!wasPending) {
    await openDetail({ id: row.id })
    return
  }
  const next = list.value.find(
    (r) => r.status === PostStatus.Pending && r.id !== row.id && !r.deleted,
  )
  if (next) {
    ElMessage.info(`下一条待审核：${next.title}`)
    await openDetail(next)
  } else {
    detailVisible.value = false
  }
}

async function batchStatus(act: BatchAction) {
  const targets = selection.value.filter((r) => r.status === act.from && !r.deleted)
  if (!targets.length) {
    ElMessage.warning(`所选帖子中没有可${act.label}的`)
    return
  }
  const ignored = selection.value.length - targets.length
  const names = targets
    .slice(0, 3)
    .map((r) => `「${r.title}」`)
    .join('、')
  const message =
    `确定${act.label}${names}${targets.length > 3 ? ` 等 ${targets.length} 条帖子` : ''}？` +
    (ignored ? `另有 ${ignored} 条状态不符，将跳过。` : '')
  const reason = await confirmStatusChange(act.status, `批量${act.label}`, message)
  if (reason === null) return
  try {
    const { data } = await batchUpdatePostStatus(
      targets.map((r) => r.id),
      act.status,
      reason,
    )
    const skipped = data.skipped + ignored
    ElMessage.success(`已${act.label} ${data.updated} 条${skipped ? `，跳过 ${skipped} 条` : ''}`)
  } catch (e) {
    showApiError(e, `批量${act.label}失败`)
    return
  }
  void refreshBadges({ force: true })
  await fetchList()
}

function goPostMessages(postId?: number) {
  if (!postId) return
  if (!can(PERM.ChatMessage)) {
    ElMessage.warning('暂无消息菜单权限')
    return
  }
  pushMenuRoute(router, employeeStore.permissions, PERM.ChatMessage, {
    postId: String(postId),
    scope: String(MessageScope.Post),
  })
}

function goRoom(roomId: number) {
  if (!roomId) return
  pushMenuRoute(router, employeeStore.permissions, PERM.ChatRoom, { id: String(roomId) })
}
</script>

<style scoped>
.batch-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding: 8px 12px;
  border-radius: 4px;
  background: var(--el-color-primary-light-9);
  font-size: 13px;
}
</style>

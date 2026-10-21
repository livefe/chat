<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div v-if="can(PERM.UserAccountCreate)" class="card-header-actions">
            <el-button type="primary" @click="handleAdd">新增用户</el-button>
          </div>
        </div>
      </template>

      <UserFilter
        v-model:id="query.id"
        v-model:mobile="query.mobile"
        v-model:nickname="query.nickname"
        v-model:status="query.status"
        v-model:muted="query.muted"
        @search="handleSearch"
        @reset="handleReset"
      />

      <UserTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @detail="openDetail"
        @edit="handleEdit"
        @mute="handleMute"
        @toggle-status="handleToggleStatus"
        @deactivate="handleDeactivate"
        @page-change="fetchList"
      />
    </el-card>

    <UserFormDialog v-model="formVisible" :model="form" @success="fetchList" />

    <UserDetailDrawer
      v-model="detailVisible"
      :detail="detail"
      :loading="detailLoading"
      @messages="goUserMessages"
      @go-room="goRoom"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { showApiError } from '@/utils/apiError'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { useRouteQuery } from '@/composables/useRouteQuery'
import { useEmployeeStore } from '@/stores/employee'
import { confirmAndRun } from '@/utils/confirm'
import { confirmWithReason } from '@/utils/promptReason'
import { pushMenuRoute } from '@/utils/navigation'
import {
  getUserList,
  getUserDetail,
  updateUserStatus,
  deactivateUser,
  muteUser,
  type UserItem,
  type AdminUserDetail,
} from '@/api/user'
import { MUTE_LABELS } from '@/constants/user'
import UserFilter from './UserFilter.vue'
import UserTable from './UserTable.vue'
import UserDetailDrawer from './UserDetailDrawer.vue'
import UserFormDialog, { type UserFormModel } from './UserFormDialog.vue'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

const router = useRouter()
const route = useRoute()
const employeeStore = useEmployeeStore()
const { can } = usePermission()
const {
  loading,
  list,
  total,
  query,
  fetchList,
  handleSearch,
  handleReset: doReset,
} = usePaginatedList(
  (q) =>
    getUserList({
      page: q.page,
      pageSize: q.pageSize,
      id: q.id || undefined,
      mobile: q.mobile || undefined,
      nickname: q.nickname || undefined,
      status: q.status,
      muted: q.muted,
    }),
  {
    page: 1,
    pageSize: 10,
    id: undefined as number | undefined,
    mobile: '',
    nickname: '',
    status: undefined as number | undefined,
    muted: undefined as boolean | undefined,
  },
  '获取用户列表失败',
)

const { wrapReset } = useRouteQuery(
  query,
  [
    {
      key: 'id',
      apply: (v, q) => {
        const id = Number(v)
        if (!Number.isNaN(id) && id > 0) q.id = id
      },
      clear: (q) => {
        q.id = undefined
      },
    },
    {
      key: 'status',
      apply: (v, q) => {
        q.status = Number(v)
      },
      clear: (q) => {
        q.status = undefined
      },
    },
    {
      key: 'muted',
      apply: (v, q) => {
        q.muted = v === '1'
      },
      clear: (q) => {
        q.muted = undefined
      },
    },
  ],
  {
    onApplied: () => {
      query.page = 1
      fetchList()
    },
  },
)

const handleReset = wrapReset(doReset)

const detailVisible = ref(false)
const detailLoading = ref(false)
const detail = ref<AdminUserDetail | null>(null)
const formVisible = ref(false)
const form = reactive<UserFormModel>({ id: 0, mobile: '', nickname: '', status: 1 })

function goUserMessages(userId: number, roomId?: number) {
  detailVisible.value = false
  pushMenuRoute(router, employeeStore.permissions, PERM.ChatMessage, {
    userId: String(userId),
    ...(roomId ? { roomId: String(roomId) } : {}),
  })
}

function goRoom(roomId: number) {
  detailVisible.value = false
  pushMenuRoute(router, employeeStore.permissions, PERM.ChatRoom, {
    roomId: String(roomId),
  })
}

async function openDetail(row: UserItem) {
  detailVisible.value = true
  detailLoading.value = true
  detail.value = null
  try {
    const res = await getUserDetail(row.id)
    detail.value = res.data
  } catch (e) {
    showApiError(e, '获取用户详情失败')
    detailVisible.value = false
  } finally {
    detailLoading.value = false
  }
}

function handleAdd() {
  Object.assign(form, { id: 0, mobile: '', nickname: '', status: 1 })
  formVisible.value = true
}

function handleEdit(row: UserItem) {
  Object.assign(form, {
    id: row.id,
    mobile: row.mobile,
    nickname: row.nickname || '',
    status: row.status,
  })
  formVisible.value = true
}

async function handleDeactivate(row: UserItem) {
  const reason = await confirmWithReason(`确定要立即注销用户 ${row.mobile} 吗？此操作不可恢复。`, {
    title: '注销用户',
    required: true,
    inputPlaceholder: '请输入注销原因',
  })
  if (reason === null) return
  await confirmAndRun(
    `确认注销用户 ${row.mobile}？`,
    async () => {
      await deactivateUser(row.id, reason)
      await fetchList()
    },
    '注销成功',
  )
}

async function handleMute(row: UserItem, duration: number) {
  const label = MUTE_LABELS[duration] ?? '更新禁言'
  let reason = ''
  if (duration !== 0) {
    const input = await confirmWithReason(`确定要对用户 ${row.mobile} 执行「${label}」吗？`, {
      title: '禁言用户',
      inputPlaceholder: '请输入禁言原因（选填）',
    })
    if (input === null) return
    reason = input
  } else {
    try {
      await ElMessageBox.confirm(`确定对用户 ${row.mobile} 执行「${label}」吗？`, '提示', {
        type: 'warning',
      })
    } catch {
      return
    }
  }
  try {
    await muteUser(row.id, duration, reason)
    ElMessage.success('操作成功')
    await fetchList()
  } catch (e) {
    showApiError(e, '操作失败')
  }
}

async function handleToggleStatus(row: UserItem) {
  const next = row.status === 1 ? 0 : 1
  await confirmAndRun(`确定要${next === 1 ? '启用' : '禁用'}该用户吗？`, async () => {
    await updateUserStatus(row.id, next)
    await fetchList()
  })
}
</script>

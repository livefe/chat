<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
        </div>
      </template>

      <ConversationFilter
        v-model:user-id="query.userId"
        @search="handleSearch"
        @reset="handleReset"
      />

      <ConversationTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @messages="goPrivateMessages"
        @command="handleMoreCommand"
        @page-change="fetchList"
      />
    </el-card>

    <ConversationUserDrawer
      v-model="userDetailVisible"
      :detail="userDetail"
      :loading="userDetailLoading"
      @messages="goUserAllMessages"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getConversationList, type AdminConversationItem } from '@/api/conversation'
import { getUserDetail, muteUser, type AdminUserDetail } from '@/api/user'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { MessageScope } from '@/constants/message'
import { PERM } from '@/constants/permission'
import { MUTE_LABELS } from '@/constants/user'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { pushMenuRoute } from '@/utils/navigation'
import { confirmWithReason } from '@/utils/promptReason'
import ConversationFilter from './ConversationFilter.vue'
import ConversationTable from './ConversationTable.vue'
import ConversationUserDrawer from './ConversationUserDrawer.vue'

const router = useRouter()
const route = useRoute()
const employeeStore = useEmployeeStore()

const { loading, list, total, query, fetchList, handleSearch, handleReset } = usePaginatedList(
  (q) =>
    getConversationList({
      page: q.page,
      pageSize: q.pageSize,
      userId: q.userId || undefined,
    }),
  {
    page: 1,
    pageSize: 20,
    userId: undefined as number | undefined,
  },
  '获取会话列表失败',
)

const userDetailVisible = ref(false)
const userDetailLoading = ref(false)
const userDetail = ref<AdminUserDetail | null>(null)

function goPrivateMessages(row: AdminConversationItem) {
  pushMenuRoute(router, employeeStore.permissions, PERM.ChatMessage, {
    userId: String(row.userMinId),
    peerUserId: String(row.userMaxId),
    scope: String(MessageScope.Private),
  })
}

function goUserAllMessages(userId: number) {
  userDetailVisible.value = false
  pushMenuRoute(router, employeeStore.permissions, PERM.ChatMessage, {
    userId: String(userId),
  })
}

function handleMoreCommand(row: AdminConversationItem, command: string) {
  if (command.startsWith('detail:')) {
    openUserDetail(command === 'detail:min' ? row.userMinId : row.userMaxId)
    return
  }
  if (command.startsWith('mute:')) {
    const [, side, durationStr] = command.split(':')
    const userId = side === 'min' ? row.userMinId : row.userMaxId
    const name =
      side === 'min' ? row.userMinName || row.userMinMobile : row.userMaxName || row.userMaxMobile
    const duration = Number(durationStr)
    void runMute(userId, name, duration, MUTE_LABELS[duration] ?? '更新禁言')
  }
}

async function runMute(userId: number, name: string, duration: number, label: string) {
  let reason = ''
  if (duration !== 0) {
    const input = await confirmWithReason(`确定对用户 ${name} 执行「${label}」吗？`, {
      title: '禁言用户',
      inputPlaceholder: '请输入禁言原因（选填）',
    })
    if (input === null) return
    reason = input
  } else {
    try {
      await ElMessageBox.confirm(`确定对用户 ${name} 执行「${label}」吗？`, '提示', {
        type: 'warning',
      })
    } catch {
      return
    }
  }
  try {
    await muteUser(userId, duration, reason)
    ElMessage.success('操作成功')
    if (userDetailVisible.value && userDetail.value?.id === userId) {
      await refreshUserDetail(userId)
    }
  } catch (e) {
    showApiError(e, '操作失败')
  }
}

async function refreshUserDetail(userId: number) {
  userDetailLoading.value = true
  try {
    const res = await getUserDetail(userId)
    userDetail.value = res.data
  } catch (e) {
    showApiError(e, '刷新用户详情失败')
  } finally {
    userDetailLoading.value = false
  }
}

async function openUserDetail(userId: number) {
  userDetailVisible.value = true
  userDetailLoading.value = true
  userDetail.value = null
  try {
    const res = await getUserDetail(userId)
    userDetail.value = res.data
  } catch (e) {
    showApiError(e, '获取用户详情失败')
    userDetailVisible.value = false
  } finally {
    userDetailLoading.value = false
  }
}

onMounted(fetchList)
</script>

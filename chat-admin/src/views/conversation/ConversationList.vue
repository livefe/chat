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
        v-model:peer-user-id="query.peerUserId"
        v-model:start-time="query.startTime"
        v-model:end-time="query.endTime"
        @search="handleSearch"
        @reset="handleReset"
      />

      <ConversationTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @messages="goPrivateMessages"
        @mute="handleMute"
        @page-change="fetchList"
      />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router'

import { getConversationList, type AdminConversationItem } from '@/api/conversation'
import { useListQuery } from '@/composables/useListQuery'
import { useUserActions } from '@/composables/useUserActions'
import { MessageScope } from '@/constants/message'
import { PERM } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'
import { pushMenuRoute } from '@/utils/navigation'
import ConversationFilter from './ConversationFilter.vue'
import ConversationTable from './ConversationTable.vue'

const router = useRouter()
const route = useRoute()
const employeeStore = useEmployeeStore()
const actions = useUserActions()

const { loading, error, list, total, query, fetchList, handleSearch, handleReset } = useListQuery({
  fetch: (q) => getConversationList(q),
  defaults: {
    page: 1,
    pageSize: 20,
    userId: undefined as number | undefined,
    peerUserId: undefined as number | undefined,
    startTime: '',
    endTime: '',
  },
  schema: {
    userId: 'number',
    peerUserId: 'number',
    startTime: 'string',
    endTime: 'string',
  },
  errorMessage: '获取会话列表失败',
})

function goPrivateMessages(row: AdminConversationItem) {
  pushMenuRoute(router, employeeStore.permissions, PERM.ChatMessage, {
    userId: String(row.userMinId),
    peerUserId: String(row.userMaxId),
    scope: String(MessageScope.Private),
  })
}

function handleMute(row: AdminConversationItem, side: 'min' | 'max') {
  actions.openMute([
    side === 'min'
      ? { id: row.userMinId, nickname: row.userMinName }
      : { id: row.userMaxId, nickname: row.userMaxName },
  ])
}
</script>

<template>
  <div class="user-tab">
    <el-table :data="list" v-loading="loading" size="small">
      <template #empty>
        <TableEmpty :error="error" text="暂无私信会话" @retry="fetchList" />
      </template>
      <el-table-column label="对方" min-width="160">
        <template #default="{ row }">
          <UserCell v-bind="peerOf(row as AdminConversationItem)" />
        </template>
      </el-table-column>
      <el-table-column prop="lastContent" label="最后消息" min-width="180" show-overflow-tooltip />
      <el-table-column label="最后活跃" width="160">
        <template #default="{ row }">{{ formatUnixSeconds(row.lastActiveAt) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="90" fixed="right">
        <template #default="{ row }">
          <el-button
            v-if="can(PERM.ChatMessage)"
            size="small"
            link
            type="primary"
            @click="openMessages(row as AdminConversationItem)"
          >
            查看消息
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-if="total > query.pageSize"
      v-model:current-page="query.page"
      class="user-tab__pager"
      small
      layout="total, prev, pager, next"
      :page-size="query.pageSize"
      :total="total"
      @current-change="fetchList"
    />
  </div>
</template>

<script setup lang="ts">
import { getConversationList, type AdminConversationItem } from '@/api/conversation'
import TableEmpty from '@/components/common/TableEmpty.vue'
import UserCell from '@/components/common/UserCell.vue'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { MessageScope } from '@/constants/message'
import { PERM } from '@/constants/permission'
import { formatUnixSeconds } from '@/utils/datetime'
import { useDrawerNav } from './useDrawerNav'

const props = defineProps<{ userId: number }>()

const { can } = usePermission()
const { go } = useDrawerNav()

const { query, list, total, loading, error, fetchList } = useListQuery({
  fetch: (q) => getConversationList(q),
  defaults: { page: 1, pageSize: 10, userId: props.userId },
  errorMessage: '获取私信会话失败',
  syncUrl: false,
})

function peerOf(row: AdminConversationItem) {
  return row.userMinId === props.userId
    ? { userId: row.userMaxId, nickname: row.userMaxName, mobile: row.userMaxMobile }
    : { userId: row.userMinId, nickname: row.userMinName, mobile: row.userMinMobile }
}

function openMessages(row: AdminConversationItem) {
  go(PERM.ChatMessage, {
    userId: String(props.userId),
    peerUserId: String(peerOf(row).userId),
    scope: String(MessageScope.Private),
  })
}
</script>

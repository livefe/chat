<template>
  <el-table :data="list" v-loading="loading">
    <template #empty>
      <TableEmpty :error="error" text="暂无会话" @retry="emit('page-change')" />
    </template>
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column label="用户A" min-width="160">
      <template #default="{ row }">
        <UserCell
          :user-id="row.userMinId"
          :nickname="row.userMinName"
          :mobile="row.userMinMobile"
        />
      </template>
    </el-table-column>
    <el-table-column label="用户B" min-width="160">
      <template #default="{ row }">
        <UserCell
          :user-id="row.userMaxId"
          :nickname="row.userMaxName"
          :mobile="row.userMaxMobile"
        />
      </template>
    </el-table-column>
    <el-table-column prop="lastContent" label="最后消息" min-width="180" show-overflow-tooltip />
    <el-table-column label="最后活跃" width="170">
      <template #default="{ row }">{{ formatUnixSecondsLocale(row.lastActiveAt) }}</template>
    </el-table-column>
    <el-table-column label="操作" width="120" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.ChatMessage)"
            size="small"
            link
            type="primary"
            @click="emit('messages', row as AdminConversationItem)"
          >
            消息
          </el-button>
          <el-dropdown
            v-if="can(PERM.UserAccountMute)"
            trigger="click"
            @command="(cmd: 'min' | 'max') => emit('mute', row as AdminConversationItem, cmd)"
          >
            <el-button size="small" link type="primary">
              更多
              <el-icon class="action-icon"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="min">禁言 {{ row.userMinName }}</el-dropdown-item>
                <el-dropdown-item command="max">禁言 {{ row.userMaxName }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </template>
    </el-table-column>
  </el-table>

  <div class="pagination">
    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="PAGE_SIZES"
      :layout="PAGINATION_LAYOUT"
      @size-change="emit('page-change')"
      @current-change="emit('page-change')"
    />
  </div>
</template>

<script setup lang="ts">
import { ArrowDown } from '@element-plus/icons-vue'

import type { AdminConversationItem } from '@/api/conversation'
import TableEmpty from '@/components/common/TableEmpty.vue'
import UserCell from '@/components/common/UserCell.vue'
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'
import { formatUnixSecondsLocale } from '@/utils/datetime'

defineProps<{
  list: AdminConversationItem[]
  loading: boolean
  error?: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  messages: [row: AdminConversationItem]
  mute: [row: AdminConversationItem, side: 'min' | 'max']
  'page-change': []
}>()

const { can } = usePermission()
</script>

<style scoped>
.action-icon {
  margin-left: 2px;
  font-size: 12px;
}
</style>

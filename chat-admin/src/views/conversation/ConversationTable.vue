<template>
  <el-table :data="list" v-loading="loading" empty-text="暂无会话">
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column label="用户A" min-width="150">
      <template #default="{ row }">
        <div>{{ row.userMinName }}</div>
        <div class="sub">{{ row.userMinMobile }} · ID {{ row.userMinId }}</div>
      </template>
    </el-table-column>
    <el-table-column label="用户B" min-width="150">
      <template #default="{ row }">
        <div>{{ row.userMaxName }}</div>
        <div class="sub">{{ row.userMaxMobile }} · ID {{ row.userMaxId }}</div>
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
            v-if="canViewMessage"
            size="small"
            link
            type="primary"
            @click="emit('messages', row as AdminConversationItem)"
          >
            消息
          </el-button>
          <el-dropdown
            v-if="hasMoreActions"
            trigger="click"
            @command="(cmd: string) => emit('command', row as AdminConversationItem, cmd)"
          >
            <el-button size="small" link type="primary">
              更多
              <el-icon class="action-icon"><ArrowDown /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <template v-if="canViewUserDetail">
                  <el-dropdown-item command="detail:min">
                    {{ row.userMinName || row.userMinMobile }} 详情
                  </el-dropdown-item>
                  <el-dropdown-item command="detail:max">
                    {{ row.userMaxName || row.userMaxMobile }} 详情
                  </el-dropdown-item>
                </template>
                <template v-if="canMuteUser">
                  <el-dropdown-item
                    v-if="canViewUserDetail"
                    divided
                    disabled
                    class="dropdown-label"
                  >
                    禁言用户A
                  </el-dropdown-item>
                  <el-dropdown-item v-else disabled class="dropdown-label"
                    >禁言用户A</el-dropdown-item
                  >
                  <el-dropdown-item command="mute:min:1">1 小时</el-dropdown-item>
                  <el-dropdown-item command="mute:min:24">24 小时</el-dropdown-item>
                  <el-dropdown-item command="mute:min:168">7 天</el-dropdown-item>
                  <el-dropdown-item command="mute:min:-1">永久</el-dropdown-item>
                  <el-dropdown-item divided disabled class="dropdown-label"
                    >禁言用户B</el-dropdown-item
                  >
                  <el-dropdown-item command="mute:max:1">1 小时</el-dropdown-item>
                  <el-dropdown-item command="mute:max:24">24 小时</el-dropdown-item>
                  <el-dropdown-item command="mute:max:168">7 天</el-dropdown-item>
                  <el-dropdown-item command="mute:max:-1">永久</el-dropdown-item>
                </template>
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
import { computed } from 'vue'
import type { AdminConversationItem } from '@/api/conversation'
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'
import { formatUnixSecondsLocale } from '@/utils/datetime'

defineProps<{
  list: AdminConversationItem[]
  loading: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  messages: [row: AdminConversationItem]
  command: [row: AdminConversationItem, command: string]
  'page-change': []
}>()

const { can } = usePermission()
const canViewMessage = computed(() => can(PERM.ChatMessage))
const canViewUserDetail = computed(() => can(PERM.UserAccountDetail))
const canMuteUser = computed(() => can(PERM.UserAccountMute))
const hasMoreActions = computed(() => canViewUserDetail.value || canMuteUser.value)
</script>

<style scoped>
.action-icon {
  margin-left: 2px;
  font-size: 12px;
}
:deep(.dropdown-label) {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  cursor: default;
}
</style>

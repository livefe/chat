<template>
  <el-table
    :data="list"
    v-loading="loading"
    row-key="id"
    @selection-change="(rows: UserItem[]) => emit('selection-change', rows)"
  >
    <template #empty>
      <TableEmpty :error="error" text="暂无用户" @retry="emit('page-change')" />
    </template>
    <el-table-column v-if="canBatch" type="selection" width="44" :selectable="isSelectable" />
    <el-table-column label="用户" min-width="220">
      <template #default="{ row }">
        <div class="user-row">
          <el-avatar :src="row.avatarUrl || undefined" :size="36">
            {{ (row.nickname || '用').slice(0, 1) }}
          </el-avatar>
          <UserCell :user-id="row.id" :nickname="row.nickname" :mobile="row.mobile" />
        </div>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <el-tag v-if="row.deactivated" type="info">已注销</el-tag>
        <el-tag v-else :type="row.status === 1 ? 'success' : 'danger'">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="禁言" width="170">
      <template #default="{ row }">
        <el-tag v-if="row.muteText" type="warning">{{ row.muteText }}</el-tag>
        <span v-else class="sub">-</span>
      </template>
    </el-table-column>
    <el-table-column label="最近登录" width="170">
      <template #default="{ row }">{{ row.lastLoginAt || '-' }}</template>
    </el-table-column>
    <el-table-column prop="createdAt" label="注册时间" width="170" />
    <el-table-column label="操作" min-width="260" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.UserAccountDetail)"
            size="small"
            link
            type="primary"
            @click="drawer.open(row.id)"
          >
            详情
          </el-button>
          <template v-if="!row.deactivated">
            <el-button
              v-if="can(PERM.UserAccountUpdate)"
              size="small"
              link
              type="primary"
              @click="actions.openEdit(row as UserItem)"
            >
              编辑
            </el-button>
            <template v-if="can(PERM.UserAccountMute)">
              <el-button
                v-if="row.muteText"
                size="small"
                link
                type="warning"
                @click="actions.unmute(row as UserItem)"
              >
                解除禁言
              </el-button>
              <el-button
                v-else
                size="small"
                link
                type="warning"
                @click="actions.openMute([row as UserItem])"
              >
                禁言
              </el-button>
            </template>
            <template v-if="can(PERM.UserAccountStatus)">
              <el-button
                v-if="row.status === UserStatus.Enabled"
                size="small"
                link
                type="danger"
                @click="actions.disable(row as UserItem)"
              >
                禁用
              </el-button>
              <el-button
                v-else
                size="small"
                link
                type="success"
                @click="actions.enable(row as UserItem)"
              >
                启用
              </el-button>
            </template>
            <el-button
              v-if="can(PERM.UserAccountDeactivate)"
              size="small"
              link
              type="danger"
              @click="actions.deactivate(row as UserItem)"
            >
              注销
            </el-button>
          </template>
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

import type { UserItem } from '@/api/user'
import TableEmpty from '@/components/common/TableEmpty.vue'
import UserCell from '@/components/common/UserCell.vue'
import { usePermission } from '@/composables/usePermission'
import { useUserActions } from '@/composables/useUserActions'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'
import { UserStatus } from '@/constants/user'
import { useUserDrawerStore } from '@/stores/userDrawer'

defineProps<{
  list: UserItem[]
  loading: boolean
  error?: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  'selection-change': [rows: UserItem[]]
  'page-change': []
}>()

const { can } = usePermission()
const drawer = useUserDrawerStore()
const actions = useUserActions()

const canBatch = computed(() => can(PERM.UserAccountMute) || can(PERM.UserAccountStatus))
const isSelectable = (row: UserItem) => !row.deactivated
</script>

<style scoped>
.user-row {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
}
</style>

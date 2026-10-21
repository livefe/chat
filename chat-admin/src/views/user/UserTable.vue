<template>
  <el-table :data="list" v-loading="loading" empty-text="暂无用户">
    <el-table-column prop="id" label="ID" width="80" />
    <el-table-column prop="mobile" label="手机号" min-width="130" />
    <el-table-column prop="nickname" label="昵称" min-width="120" />
    <el-table-column label="状态" width="120">
      <template #default="{ row }">
        <el-tag v-if="row.deactivated" type="info">已注销</el-tag>
        <el-tag v-else :type="row.status === 1 ? 'success' : 'danger'">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="禁言" width="140">
      <template #default="{ row }">
        <el-tag v-if="row.muteText" type="warning">{{ row.muteText }}</el-tag>
        <span v-else class="sub">-</span>
      </template>
    </el-table-column>
    <el-table-column prop="createdAt" label="创建时间" width="180" />
    <el-table-column label="操作" min-width="340" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            size="small"
            link
            type="primary"
            @click="emit('detail', row as UserItem)"
            v-if="can(PERM.UserAccountDetail)"
          >
            详情
          </el-button>
          <el-button
            size="small"
            link
            type="primary"
            @click="emit('edit', row as UserItem)"
            v-if="!row.deactivated && can(PERM.UserAccountUpdate)"
          >
            编辑
          </el-button>
          <template v-if="!row.deactivated && can(PERM.UserAccountMute)">
            <el-button
              v-if="row.muteText"
              size="small"
              link
              type="warning"
              @click="emit('mute', row as UserItem, 0)"
            >
              解除禁言
            </el-button>
            <el-dropdown
              trigger="click"
              @command="(duration: string) => emit('mute', row as UserItem, Number(duration))"
            >
              <el-button size="small" link type="warning">禁言</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="1">1 小时</el-dropdown-item>
                  <el-dropdown-item command="24">24 小时</el-dropdown-item>
                  <el-dropdown-item command="168">7 天</el-dropdown-item>
                  <el-dropdown-item command="-1">永久</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </template>
          <el-button
            size="small"
            link
            :type="row.status === 1 ? 'danger' : 'success'"
            @click="emit('toggle-status', row as UserItem)"
            v-if="!row.deactivated && can(PERM.UserAccountStatus)"
          >
            {{ row.status === 1 ? '禁用' : '启用' }}
          </el-button>
          <el-button
            size="small"
            link
            type="danger"
            @click="emit('deactivate', row as UserItem)"
            v-if="!row.deactivated && can(PERM.UserAccountDeactivate)"
          >
            注销
          </el-button>
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
import type { UserItem } from '@/api/user'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

defineProps<{
  list: UserItem[]
  loading: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  detail: [row: UserItem]
  edit: [row: UserItem]
  mute: [row: UserItem, duration: number]
  'toggle-status': [row: UserItem]
  deactivate: [row: UserItem]
  'page-change': []
}>()

const { can } = usePermission()
</script>

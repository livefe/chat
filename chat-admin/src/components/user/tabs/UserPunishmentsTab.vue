<template>
  <div class="user-tab">
    <el-table :data="list" v-loading="loading" size="small">
      <template #empty>
        <TableEmpty :error="error" text="暂无处罚记录" @retry="fetchList" />
      </template>
      <el-table-column prop="createdAt" label="时间" width="160" />
      <el-table-column label="动作" width="130">
        <template #default="{ row }">
          <el-tag :type="punishActionTagType(row.action)" size="small">{{ row.actionText }}</el-tag>
          <div v-if="row.muteText" class="sub">{{ row.muteText }}</div>
        </template>
      </el-table-column>
      <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
      <el-table-column label="对象" width="110">
        <template #default="{ row }">
          <el-link
            v-if="row.action === PunishAction.DeleteMessage && can(PERM.ChatMessage)"
            type="primary"
            :underline="false"
            @click="go(PERM.ChatMessage, { contextId: String(row.targetId) })"
          >
            消息 #{{ row.targetId }}
          </el-link>
          <el-link
            v-else-if="row.action === PunishAction.DisablePost && can(PERM.PostPost)"
            type="primary"
            :underline="false"
            @click="go(PERM.PostPost, { id: String(row.targetId) })"
          >
            帖子 #{{ row.targetId }}
          </el-link>
          <span v-else-if="row.targetId" class="sub">#{{ row.targetId }}</span>
          <span v-else class="sub">-</span>
        </template>
      </el-table-column>
      <el-table-column label="操作人" width="100">
        <template #default="{ row }">{{ row.employeeName || '-' }}</template>
      </el-table-column>
      <el-table-column label="来源" width="100">
        <template #default="{ row }">
          <el-link
            v-if="row.reportId && can(PERM.UserReport)"
            type="primary"
            :underline="false"
            @click="openReport(row.reportId)"
          >
            举报 #{{ row.reportId }}
          </el-link>
          <span v-else-if="row.reportId" class="sub">举报 #{{ row.reportId }}</span>
          <span v-else class="sub">后台操作</span>
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
import { getUserPunishments } from '@/api/user'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { FILTER_ALL } from '@/constants'
import { PERM } from '@/constants/permission'
import { PunishAction, punishActionTagType } from '@/constants/user'
import { useDrawerNav } from './useDrawerNav'

const props = defineProps<{ userId: number }>()

const { can } = usePermission()
const { go } = useDrawerNav()

const { query, list, total, loading, error, fetchList } = useListQuery({
  fetch: (q) => getUserPunishments(q),
  defaults: { page: 1, pageSize: 10, userId: props.userId },
  errorMessage: '获取处罚记录失败',
  syncUrl: false,
})

function openReport(reportId: number) {
  go(PERM.UserReport, {
    targetUserId: String(props.userId),
    status: String(FILTER_ALL),
    reviewId: String(reportId),
  })
}
</script>

<template>
  <div class="user-tab">
    <el-table
      :data="list"
      v-loading="loading"
      size="small"
      row-class-name="user-tab__row"
      @row-click="openDetail"
    >
      <template #empty>
        <TableEmpty :error="error" text="暂无反馈" @retry="fetchList" />
      </template>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="typeText" label="类型" width="90" />
      <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="feedbackStatusTagType(row.status)" size="small">
            {{ row.statusText }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="提交时间" width="160" />
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
import { getFeedbackList, type FeedbackItem } from '@/api/feedback'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { FILTER_ALL } from '@/constants'
import { feedbackStatusTagType } from '@/constants/feedback'
import { PERM } from '@/constants/permission'
import { useDrawerNav } from './useDrawerNav'

const props = defineProps<{ userId: number }>()

const { can } = usePermission()
const { go } = useDrawerNav()

const { query, list, total, loading, error, fetchList } = useListQuery({
  fetch: (q) => getFeedbackList(q),
  defaults: { page: 1, pageSize: 10, userId: props.userId },
  errorMessage: '获取反馈失败',
  syncUrl: false,
})

function openDetail(row: FeedbackItem) {
  if (!can(PERM.UserFeedbackDetail)) return
  go(PERM.UserFeedback, {
    userId: String(props.userId),
    status: String(FILTER_ALL),
    detailId: String(row.id),
  })
}
</script>

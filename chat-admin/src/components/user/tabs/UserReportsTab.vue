<template>
  <div class="user-tab">
    <el-radio-group
      v-model="query.side"
      size="small"
      class="user-tab__toolbar"
      @change="handleSearch"
    >
      <el-radio-button value="target">被举报</el-radio-button>
      <el-radio-button value="reporter">举报他人</el-radio-button>
    </el-radio-group>
    <el-table
      :data="list"
      v-loading="loading"
      size="small"
      row-class-name="user-tab__row"
      @row-click="openReview"
    >
      <template #empty>
        <TableEmpty :error="error" text="暂无举报" @retry="fetchList" />
      </template>
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="targetTypeText" label="对象" width="60" />
      <el-table-column label="内容" min-width="180">
        <template #default="{ row }">
          <div>{{ reportTargetPreview(row as ReportItem, 40) }}</div>
          <div v-if="query.side === 'reporter' && row.targetUserId" class="sub">
            被举报：{{ row.targetUserName }}
          </div>
          <div v-if="query.side === 'target'" class="sub">举报人：{{ row.reporterNickname }}</div>
        </template>
      </el-table-column>
      <el-table-column prop="reasonText" label="原因" width="90" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="reportStatusTagType(row.status)" size="small">{{ row.statusText }}</el-tag>
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
import { getReportList, type ReportItem } from '@/api/report'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { FILTER_ALL } from '@/constants'
import { PERM } from '@/constants/permission'
import { reportStatusTagType, reportTargetPreview } from '@/utils/reportDisplay'
import { useDrawerNav } from './useDrawerNav'

const props = defineProps<{ userId: number }>()

const { can } = usePermission()
const { go } = useDrawerNav()

const { query, list, total, loading, error, fetchList, handleSearch } = useListQuery({
  fetch: ({ side, ...q }) =>
    getReportList(
      side === 'target' ? { ...q, targetUserId: props.userId } : { ...q, reporterId: props.userId },
    ),
  defaults: { page: 1, pageSize: 10, side: 'target' as 'target' | 'reporter' },
  errorMessage: '获取举报记录失败',
  syncUrl: false,
})

function openReview(row: ReportItem) {
  if (!can(PERM.UserReportDetail)) return
  const key = query.side === 'target' ? 'targetUserId' : 'reporterId'
  go(PERM.UserReport, {
    [key]: String(props.userId),
    status: String(FILTER_ALL),
    reviewId: String(row.id),
  })
}
</script>

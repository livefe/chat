<template>
  <el-table :data="list" v-loading="loading" row-class-name="report-row" @row-click="onRowClick">
    <template #empty>
      <TableEmpty :error="error" text="暂无举报" @retry="emit('page-change')" />
    </template>
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column label="举报人" min-width="140">
      <template #default="{ row }">
        <UserCell
          :user-id="row.reporterId"
          :nickname="row.reporterNickname"
          :mobile="row.reporterMobile"
        />
      </template>
    </el-table-column>
    <el-table-column prop="targetTypeText" label="对象" width="64" />
    <el-table-column label="被举报内容" min-width="200">
      <template #default="{ row }">
        <div class="target-preview">{{ reportTargetPreview(row as ReportItem) }}</div>
      </template>
    </el-table-column>
    <el-table-column label="被举报用户" min-width="140">
      <template #default="{ row }">
        <UserCell
          :user-id="row.targetUserId"
          :nickname="row.targetUserName"
          :mobile="row.targetUserMobile"
        />
      </template>
    </el-table-column>
    <el-table-column prop="reasonText" label="原因" width="90" />
    <el-table-column label="状态" width="120">
      <template #default="{ row }">
        <el-tag :type="reportStatusTagType(row.status)">{{ row.statusText }}</el-tag>
        <el-tooltip
          v-if="(row.targetReportCount ?? 0) > 1"
          content="查看同一对象的全部举报"
          placement="top"
        >
          <el-tag
            type="warning"
            effect="plain"
            size="small"
            class="repeat-tag"
            @click.stop="emit('same-target', row as ReportItem)"
          >
            ×{{ row.targetReportCount }}
          </el-tag>
        </el-tooltip>
      </template>
    </el-table-column>
    <el-table-column prop="createdAt" label="提交时间" width="170" />
    <el-table-column label="操作" width="80" fixed="right">
      <template #default="{ row }">
        <el-button
          v-if="can(PERM.UserReportDetail)"
          size="small"
          link
          type="primary"
          @click.stop="emit('review', row as ReportItem)"
        >
          {{ row.status === ReportStatus.Pending ? '审核' : '查看' }}
        </el-button>
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
import type { ReportItem } from '@/api/report'
import TableEmpty from '@/components/common/TableEmpty.vue'
import UserCell from '@/components/common/UserCell.vue'
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'
import { ReportStatus } from '@/constants/report'
import { reportStatusTagType, reportTargetPreview } from '@/utils/reportDisplay'

defineProps<{
  list: ReportItem[]
  loading: boolean
  error?: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  review: [ReportItem]
  'same-target': [ReportItem]
  'page-change': []
}>()

const { can } = usePermission()

function onRowClick(row: ReportItem) {
  if (!can(PERM.UserReportDetail)) return
  emit('review', row)
}
</script>

<style scoped>
.target-preview {
  line-height: 1.5;
  word-break: break-word;
}

.repeat-tag {
  margin-left: 6px;
  cursor: pointer;
}

:deep(.report-row) {
  cursor: pointer;
}
</style>

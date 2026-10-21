<template>
  <el-table
    :data="list"
    v-loading="loading"
    empty-text="暂无举报"
    row-class-name="report-row"
    @row-click="onRowClick"
  >
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column label="举报人" width="130">
      <template #default="{ row }">
        <div>{{ row.reporterNickname || '-' }}</div>
        <div class="sub">{{ row.reporterMobile }}</div>
      </template>
    </el-table-column>
    <el-table-column prop="targetTypeText" label="对象" width="72" />
    <el-table-column label="被举报内容" min-width="200">
      <template #default="{ row }">
        <div class="target-preview">{{ targetPreview(row as ReportItem) }}</div>
        <div v-if="row.targetUserName" class="sub">用户：{{ row.targetUserName }}</div>
      </template>
    </el-table-column>
    <el-table-column prop="reasonText" label="原因" width="96" />
    <el-table-column label="状态" width="108">
      <template #default="{ row }">
        <el-tag :type="reportStatusTagType(row.status)">{{ row.statusText }}</el-tag>
        <el-tag
          v-if="(row.targetReportCount ?? 0) > 1"
          type="warning"
          effect="plain"
          size="small"
          class="repeat-tag"
        >
          ×{{ row.targetReportCount }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="createdAt" label="提交时间" width="170" />
    <el-table-column label="操作" width="88" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            size="small"
            link
            type="primary"
            @click.stop="emit('review', row as ReportItem)"
            v-if="can(PERM.UserReportDetail)"
          >
            审核
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
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { type ReportItem } from '@/api/report'
import { reportStatusTagType } from '@/utils/reportDisplay'
import { formatMessageSummary } from '@/utils/messageDisplay'
import { MessageContentType } from '@/constants/message'
import { ReportTargetType } from '@/constants/report'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

defineProps<{
  list: ReportItem[]
  loading: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  review: [ReportItem]
  'page-change': []
}>()

const { can } = usePermission()

function targetPreview(row: ReportItem) {
  const ctx = row.context
  if (row.targetType === ReportTargetType.Post) {
    return ctx.postTitle || `内容 #${row.targetId}`
  }
  if (row.targetType === ReportTargetType.Message) {
    const text = formatMessageSummary(
      ctx.messageType ?? MessageContentType.Text,
      ctx.messageContent ?? '',
    )
    return text.length > 60 ? `${text.slice(0, 60)}…` : text || `消息 #${row.targetId}`
  }
  return ctx.nickname || row.targetUserName || `用户 #${row.targetId}`
}

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
}
.sub {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
:deep(.report-row) {
  cursor: pointer;
}
</style>

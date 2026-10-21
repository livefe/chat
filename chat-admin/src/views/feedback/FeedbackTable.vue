<template>
  <el-table :data="list" v-loading="loading" empty-text="暂无反馈">
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column label="用户" width="140">
      <template #default="{ row }">
        <div>{{ row.userNickname || '-' }}</div>
        <div class="sub">{{ row.userMobile }}</div>
      </template>
    </el-table-column>
    <el-table-column prop="typeText" label="类型" width="100" />
    <el-table-column prop="content" label="内容" min-width="200" show-overflow-tooltip />
    <el-table-column prop="contact" label="联系方式" width="120" show-overflow-tooltip />
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <el-tag :type="statusTag(row.status)">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="处理人" width="120">
      <template #default="{ row }">{{ row.employeeName || '-' }}</template>
    </el-table-column>
    <el-table-column prop="createdAt" label="提交时间" width="170" />
    <el-table-column label="操作" width="140" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.UserFeedbackDetail)"
            size="small"
            link
            type="primary"
            @click="emit('detail', row as FeedbackItem)"
          >
            详情
          </el-button>
          <el-button
            v-if="
              can(PERM.UserFeedbackStatus) &&
              row.status !== FeedbackStatus.Resolved &&
              row.status !== FeedbackStatus.Closed
            "
            size="small"
            link
            type="primary"
            @click="emit('handle', row as FeedbackItem)"
          >
            处理
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
import { type FeedbackItem } from '@/api/feedback'
import { FeedbackStatus } from '@/constants/feedback'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

defineProps<{
  list: FeedbackItem[]
  loading: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  detail: [FeedbackItem]
  handle: [FeedbackItem]
  'page-change': []
}>()

const { can } = usePermission()

function statusTag(status: number) {
  if (status === FeedbackStatus.Resolved) return 'success'
  if (status === FeedbackStatus.Closed) return 'info'
  return 'danger'
}
</script>

<style scoped>
.sub {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>

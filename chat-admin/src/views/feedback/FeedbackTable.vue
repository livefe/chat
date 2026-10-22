<template>
  <el-table
    :data="list"
    v-loading="loading"
    row-key="id"
    @selection-change="(rows: FeedbackItem[]) => emit('selection-change', rows)"
  >
    <template #empty>
      <TableEmpty :error="error" text="暂无反馈" @retry="emit('page-change')" />
    </template>
    <el-table-column
      v-if="can(PERM.UserFeedbackStatus)"
      type="selection"
      width="44"
      :selectable="(row: FeedbackItem) => row.status === FeedbackStatus.Pending"
    />
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column label="用户" min-width="150">
      <template #default="{ row }">
        <UserCell :user-id="row.userId" :nickname="row.userNickname" :mobile="row.userMobile" />
      </template>
    </el-table-column>
    <el-table-column prop="typeText" label="类型" width="100" />
    <el-table-column prop="content" label="内容" min-width="220" show-overflow-tooltip />
    <el-table-column prop="contact" label="联系方式" width="120" show-overflow-tooltip />
    <el-table-column label="状态" width="90">
      <template #default="{ row }">
        <el-tag :type="feedbackStatusTagType(row.status)">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="处理人" width="110">
      <template #default="{ row }">{{ row.employeeName || '-' }}</template>
    </el-table-column>
    <el-table-column prop="createdAt" label="提交时间" width="170" />
    <el-table-column label="操作" width="90" fixed="right">
      <template #default="{ row }">
        <el-button
          v-if="can(PERM.UserFeedbackDetail)"
          size="small"
          link
          type="primary"
          @click="emit('detail', row as FeedbackItem)"
        >
          {{
            row.status === FeedbackStatus.Pending && can(PERM.UserFeedbackStatus) ? '处理' : '详情'
          }}
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
import type { FeedbackItem } from '@/api/feedback'
import TableEmpty from '@/components/common/TableEmpty.vue'
import UserCell from '@/components/common/UserCell.vue'
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { FeedbackStatus, feedbackStatusTagType } from '@/constants/feedback'
import { PERM } from '@/constants/permission'

defineProps<{
  list: FeedbackItem[]
  loading: boolean
  error?: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  detail: [FeedbackItem]
  'selection-change': [rows: FeedbackItem[]]
  'page-change': []
}>()

const { can } = usePermission()
</script>

<template>
  <el-table
    :data="list"
    v-loading="loading"
    row-key="id"
    @selection-change="(rows: AdminMessageItem[]) => emit('selection-change', rows)"
  >
    <template #empty>
      <TableEmpty :error="error" text="暂无消息" @retry="emit('page-change')" />
    </template>
    <el-table-column
      v-if="can(PERM.ChatMessageDelete)"
      type="selection"
      width="44"
      :selectable="(row: AdminMessageItem) => !row.deleted"
    />
    <el-table-column label="ID" width="90">
      <template #default="{ row }"><CopyText :text="row.id" /></template>
    </el-table-column>
    <el-table-column prop="scopeText" label="范围" width="80" />
    <el-table-column label="发送者" min-width="150">
      <template #default="{ row }">
        <UserCell :user-id="row.fromUserId" :nickname="row.fromName" :mobile="row.fromMobile" />
      </template>
    </el-table-column>
    <el-table-column label="目标" min-width="150">
      <template #default="{ row }">
        <UserCell
          v-if="row.scope === MessageScope.Private"
          :user-id="row.toUserId"
          :nickname="row.toName"
          :mobile="row.toMobile"
        />
        <template v-else-if="row.scope === MessageScope.Room">
          <div>{{ row.roomName || '-' }}</div>
          <div class="sub">聊天区 ID {{ row.roomId }}</div>
        </template>
        <template v-else-if="row.scope === MessageScope.Post">
          <div>帖子讨论</div>
          <div class="sub">帖子 ID {{ row.postId }}</div>
        </template>
        <span v-else class="sub">附近广播</span>
      </template>
    </el-table-column>
    <el-table-column label="内容" min-width="240">
      <template #default="{ row }">
        <AdminReplyQuote :item="row as AdminMessageItem" />
        <MessageContent :item="row as AdminMessageItem" :muted="row.deleted || row.recalled" />
      </template>
    </el-table-column>
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <div class="status-cell">
          <el-tag v-if="row.recalled" type="warning" size="small">已撤回</el-tag>
          <el-tag v-if="row.deleted" type="danger" size="small">已删除</el-tag>
          <el-tag v-if="!row.recalled && !row.deleted" type="success" size="small">正常</el-tag>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="时间" width="170">
      <template #default="{ row }">{{ formatUnixSecondsLocale(row.createdAt) }}</template>
    </el-table-column>
    <el-table-column label="操作" width="120" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.ChatMessageContext)"
            size="small"
            link
            type="primary"
            @click="emit('context', row as AdminMessageItem)"
          >
            上下文
          </el-button>
          <el-button
            v-if="!row.deleted && can(PERM.ChatMessageDelete)"
            size="small"
            link
            type="danger"
            @click="emit('delete', row as AdminMessageItem)"
          >
            删除
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
import type { AdminMessageItem } from '@/api/message'
import AdminReplyQuote from '@/components/AdminReplyQuote.vue'
import CopyText from '@/components/common/CopyText.vue'
import TableEmpty from '@/components/common/TableEmpty.vue'
import UserCell from '@/components/common/UserCell.vue'
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { MessageScope } from '@/constants/message'
import { PERM } from '@/constants/permission'
import { formatUnixSecondsLocale } from '@/utils/datetime'
import MessageContent from './MessageContent.vue'

defineProps<{
  list: AdminMessageItem[]
  loading: boolean
  error?: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  context: [row: AdminMessageItem]
  delete: [row: AdminMessageItem]
  'selection-change': [rows: AdminMessageItem[]]
  'page-change': []
}>()

const { can } = usePermission()
</script>

<style scoped>
.status-cell {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}
</style>

<template>
  <el-table
    :data="list"
    v-loading="loading"
    row-key="id"
    @selection-change="(rows: PostItem[]) => emit('selection-change', rows)"
  >
    <template #empty>
      <TableEmpty :error="error" text="暂无帖子" @retry="emit('page-change')" />
    </template>
    <el-table-column
      v-if="selectable"
      type="selection"
      width="42"
      :selectable="(row: PostItem) => !row.deleted"
    />
    <el-table-column label="ID" width="90">
      <template #default="{ row }"><CopyText :text="row.id" /></template>
    </el-table-column>
    <el-table-column label="类型" width="90">
      <template #default="{ row }">
        <el-tag size="small">{{ postCategoryLabel(row.category) }}</el-tag>
        <el-tag v-if="row.pinned" size="small" type="warning" class="ml-4">置顶</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="标题" min-width="180" show-overflow-tooltip>
      <template #default="{ row }">
        <el-tag v-if="row.deleted" size="small" type="info" class="mr-4">已删除</el-tag>
        <span>{{ row.title }}</span>
      </template>
    </el-table-column>
    <el-table-column label="归属" width="140">
      <template #default="{ row }">
        <template v-if="row.roomId > 0">
          <el-button
            v-if="can(PERM.ChatRoom)"
            type="primary"
            link
            class="room-link"
            @click="emit('go-room', row.roomId)"
          >
            {{ row.roomName || `聊天区 #${row.roomId}` }}
          </el-button>
          <span v-else>{{ row.roomName || `聊天区 #${row.roomId}` }}</span>
          <div class="sub">#{{ row.roomId }}</div>
        </template>
        <span v-else class="sub">独立帖</span>
      </template>
    </el-table-column>
    <el-table-column label="发布者" min-width="150">
      <template #default="{ row }">
        <UserCell
          :user-id="row.publisherId"
          :nickname="row.publisher?.nickname"
          :mobile="row.publisherMobile"
        />
      </template>
    </el-table-column>
    <el-table-column label="附属信息" width="140">
      <template #default="{ row }">
        <el-tag
          v-if="formatExtraBrief(row as PostItem) !== '-'"
          size="small"
          :type="extraTagType(row as PostItem)"
          effect="plain"
        >
          {{ formatExtraBrief(row as PostItem) }}
        </el-tag>
        <span v-else class="sub">-</span>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="90">
      <template #default="{ row }">
        <el-tooltip
          v-if="row.auditRemark && row.status !== PostStatus.Published"
          :content="`原因：${row.auditRemark}`"
          placement="top"
        >
          <el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
        </el-tooltip>
        <el-tag v-else :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="发布时间" width="170">
      <template #default="{ row }">
        <div>{{ row.createdAt }}</div>
        <div v-if="row.startAt || row.endAt" class="sub">
          活动 {{ formatUnixSecondsShort(row.startAt) || '-' }} ~
          {{ formatUnixSecondsShort(row.endAt) || '-' }}
        </div>
      </template>
    </el-table-column>
    <el-table-column label="操作" min-width="240" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.PostPostDetail)"
            size="small"
            link
            type="primary"
            @click="emit('detail', row as PostItem)"
          >
            {{ row.status === PostStatus.Pending && canAudit(row as PostItem) ? '审核' : '详情' }}
          </el-button>
          <el-button
            v-if="row.extras?.needSignup && can(PERM.PostPostDetail)"
            size="small"
            link
            type="primary"
            @click="emit('signup', row as PostItem)"
          >
            报名名单
          </el-button>
          <el-button
            v-if="row.status === PostStatus.Published && can(PERM.ChatMessage)"
            size="small"
            link
            type="primary"
            @click="emit('messages', row.id)"
          >
            消息
          </el-button>
          <template v-if="can(PERM.PostPostStatus)">
            <el-button
              v-if="canChangePostStatus(row as PostItem, PostStatus.Disabled)"
              size="small"
              link
              type="danger"
              @click="emit('status', row as PostItem, PostStatus.Disabled)"
            >
              下架
            </el-button>
            <el-button
              v-if="
                row.status !== PostStatus.Pending &&
                canChangePostStatus(row as PostItem, PostStatus.Published)
              "
              size="small"
              link
              type="success"
              @click="emit('status', row as PostItem, PostStatus.Published)"
            >
              恢复上架
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
import { type PostItem } from '@/api/post'
import CopyText from '@/components/common/CopyText.vue'
import TableEmpty from '@/components/common/TableEmpty.vue'
import UserCell from '@/components/common/UserCell.vue'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { formatUnixSecondsShort } from '@/utils/datetime'
import {
  canChangePostStatus,
  extraTagType,
  formatExtraBrief,
  postCategoryLabel,
  statusLabel,
  statusTag,
} from '@/utils/postDisplay'
import { PostStatus } from '@/constants/post'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

defineProps<{
  list: PostItem[]
  loading: boolean
  error?: boolean
  total: number
  selectable?: boolean
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  detail: [row: PostItem]
  signup: [row: PostItem]
  messages: [postId: number]
  'go-room': [roomId: number]
  status: [row: PostItem, status: PostStatus]
  'selection-change': [rows: PostItem[]]
  'page-change': []
}>()

const { can } = usePermission()

function canAudit(row: PostItem) {
  return can(PERM.PostPostStatus) && !row.deleted
}
</script>

<style scoped>
.ml-4 {
  margin-left: 4px;
}

.mr-4 {
  margin-right: 4px;
}

.room-link {
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>

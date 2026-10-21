<template>
  <el-table :data="list" v-loading="loading" empty-text="暂无动态">
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column label="类型" width="90">
      <template #default="{ row }">
        <el-tag size="small">{{ postCategoryLabel(row.category) }}</el-tag>
        <el-tag v-if="row.pinned" size="small" type="warning" class="pin-tag">置顶</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="title" label="标题" min-width="160" show-overflow-tooltip />
    <el-table-column label="归属" width="120">
      <template #default="{ row }">
        <template v-if="row.roomId > 0">
          <el-button
            v-if="can(PERM.ChatRoom)"
            type="primary"
            link
            @click="emit('go-room', row.roomId)"
          >
            区 #{{ row.roomId }}
          </el-button>
          <span v-else class="sub">区 #{{ row.roomId }}</span>
        </template>
        <span v-else class="sub">独立帖</span>
      </template>
    </el-table-column>
    <el-table-column label="发布者" width="130">
      <template #default="{ row }">
        <div>{{ row.publisher?.nickname || '-' }}</div>
        <div class="sub">ID {{ row.publisherId }}</div>
      </template>
    </el-table-column>
    <el-table-column prop="placeName" label="地点" width="120" show-overflow-tooltip />
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
        <el-tag :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="时间" width="170">
      <template #default="{ row }">
        <template v-if="row.startAt || row.endAt">
          <div>{{ formatUnixSeconds(row.startAt) }}</div>
          <div class="sub">至 {{ formatUnixSeconds(row.endAt) }}</div>
        </template>
        <span v-else class="sub">{{ row.createdAt }}</span>
      </template>
    </el-table-column>
    <el-table-column label="操作" min-width="260" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.PostPostDetail)"
            size="small"
            link
            type="primary"
            @click="emit('detail', row as PostItem)"
          >
            详情
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
          <el-button
            v-if="row.status === PostStatus.Pending && can(PERM.PostPostStatus)"
            size="small"
            link
            type="warning"
            @click="emit('detail', row as PostItem)"
          >
            审核
          </el-button>
          <el-button
            v-if="row.status === PostStatus.Published && can(PERM.PostPostStatus)"
            size="small"
            link
            type="danger"
            @click="emit('status', row as PostItem, PostStatus.Disabled, '下架')"
          >
            下架
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
import { type PostItem } from '@/api/post'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { formatUnixSeconds } from '@/utils/datetime'
import {
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
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  detail: [row: PostItem]
  signup: [row: PostItem]
  messages: [postId: number]
  'go-room': [roomId: number]
  status: [row: PostItem, status: PostStatus, action: string]
  'page-change': []
}>()

const { can } = usePermission()
</script>

<style scoped>
.pin-tag {
  margin-left: 4px;
}
</style>

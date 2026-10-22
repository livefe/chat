<template>
  <el-table
    :data="list"
    v-loading="loading"
    row-key="id"
    @selection-change="(rows: RoomItem[]) => emit('selection-change', rows)"
  >
    <template #empty>
      <TableEmpty :error="error" text="暂无聊天区" @retry="emit('page-change')" />
    </template>
    <el-table-column
      v-if="selectable"
      type="selection"
      width="42"
      :selectable="(row: RoomItem) => auditable(row.status)"
    />
    <el-table-column label="ID" width="90">
      <template #default="{ row }"><CopyText :text="row.id" /></template>
    </el-table-column>
    <el-table-column prop="name" label="名称" min-width="140" show-overflow-tooltip />
    <el-table-column prop="placeTypeText" label="类型" width="90" />
    <el-table-column label="形状" width="110">
      <template #default="{ row }">
        <el-tag :type="row.shapeType === 1 ? 'primary' : 'warning'" size="small">
          {{ row.shapeType === 1 ? '圆形' : '多边形' }}
        </el-tag>
        <div class="sub">{{ row.radius }} 米</div>
      </template>
    </el-table-column>
    <el-table-column label="创建人" min-width="150">
      <template #default="{ row }">
        <UserCell
          v-if="row.source === RoomSource.User"
          :user-id="row.userId"
          :nickname="row.creatorNickname"
          :mobile="row.creatorMobile"
        />
        <template v-else>
          <div>{{ row.employeeName || '-' }}</div>
          <div class="sub">后台员工</div>
        </template>
      </template>
    </el-table-column>
    <el-table-column label="帖子数" width="80">
      <template #default="{ row }">
        <el-button
          v-if="can(PERM.PostPost)"
          link
          type="primary"
          @click="emit('posts', row as RoomItem)"
        >
          {{ row.postCount }}
        </el-button>
        <span v-else>{{ row.postCount }}</span>
      </template>
    </el-table-column>
    <el-table-column label="消息数" width="80">
      <template #default="{ row }">
        <el-button
          v-if="can(PERM.ChatMessage)"
          link
          type="primary"
          @click="emit('messages', row as RoomItem)"
        >
          {{ row.messageCount }}
        </el-button>
        <span v-else>{{ row.messageCount }}</span>
      </template>
    </el-table-column>
    <el-table-column label="围栏限制" width="90">
      <template #default="{ row }">
        <el-tag :type="row.requireGeofence ? 'warning' : 'info'" size="small">
          {{ row.requireGeofence ? '限围栏' : '不限制' }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="90">
      <template #default="{ row }">
        <el-tooltip
          v-if="row.status === RoomStatus.Rejected && row.auditRemark"
          :content="`驳回原因：${row.auditRemark}`"
          placement="top"
        >
          <el-tag :type="roomStatusTag(row.status)">{{ row.statusText }}</el-tag>
        </el-tooltip>
        <el-tag v-else :type="roomStatusTag(row.status)">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="createdAt" label="创建时间" width="170" />
    <el-table-column label="操作" min-width="180" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.ChatRoomUpdate)"
            size="small"
            link
            type="primary"
            @click="emit('edit', row as RoomItem)"
          >
            编辑
          </el-button>
          <el-button
            v-if="auditable(row.status) && can(PERM.ChatRoomAudit)"
            size="small"
            link
            type="warning"
            @click="emit('audit', row as RoomItem)"
          >
            {{ row.status === RoomStatus.Pending ? '审核' : '重新审核' }}
          </el-button>
          <el-button
            v-if="row.status === RoomStatus.Enabled && can(PERM.ChatRoomStatus)"
            size="small"
            link
            type="danger"
            @click="emit('toggle-status', row as RoomItem, RoomStatus.Disabled)"
          >
            禁用
          </el-button>
          <el-button
            v-if="row.status === RoomStatus.Disabled && can(PERM.ChatRoomStatus)"
            size="small"
            link
            type="success"
            @click="emit('toggle-status', row as RoomItem, RoomStatus.Enabled)"
          >
            启用
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
import type { RoomItem } from '@/api/room'
import CopyText from '@/components/common/CopyText.vue'
import TableEmpty from '@/components/common/TableEmpty.vue'
import UserCell from '@/components/common/UserCell.vue'
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'
import { RoomSource, RoomStatus, roomStatusTag } from '@/constants/room'

defineProps<{
  list: RoomItem[]
  loading: boolean
  error?: boolean
  total: number
  selectable?: boolean
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  posts: [row: RoomItem]
  messages: [row: RoomItem]
  edit: [row: RoomItem]
  audit: [row: RoomItem]
  'toggle-status': [row: RoomItem, status: RoomStatus]
  'selection-change': [rows: RoomItem[]]
  'page-change': []
}>()

const { can } = usePermission()

function auditable(status: number) {
  return status === RoomStatus.Pending || status === RoomStatus.Rejected
}
</script>

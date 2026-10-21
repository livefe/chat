<template>
  <el-table :data="list" v-loading="loading" empty-text="暂无聊天区">
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column prop="name" label="名称" />
    <el-table-column prop="placeTypeText" label="类型" width="100" />
    <el-table-column label="形状" width="90">
      <template #default="{ row }">
        <el-tag :type="row.shapeType === 1 ? 'primary' : 'warning'">
          {{ row.shapeType === 1 ? '圆形' : '多边形' }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="中心坐标" width="200">
      <template #default="{ row }">
        {{ row.centerLat?.toFixed(5) }}, {{ row.centerLng?.toFixed(5) }}
      </template>
    </el-table-column>
    <el-table-column label="半径(米)" width="100" prop="radius" />
    <el-table-column prop="sourceText" label="来源" width="80" />
    <el-table-column label="围栏限制" width="100">
      <template #default="{ row }">
        <el-tag :type="row.requireGeofence ? 'warning' : 'info'" size="small">
          {{ row.requireGeofence ? '限围栏' : '不限制' }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <el-tag :type="statusTag(row.status)">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="createdAt" label="创建时间" width="170" />
    <el-table-column label="操作" min-width="360" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.PostPost)"
            size="small"
            link
            type="primary"
            @click="emit('posts', row as RoomItem)"
          >
            查看帖子
          </el-button>
          <el-button
            v-if="can(PERM.ChatMessage)"
            size="small"
            link
            type="primary"
            @click="emit('messages', row as RoomItem)"
          >
            查看消息
          </el-button>
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
            v-if="row.status === 0 && can(PERM.ChatRoomAudit)"
            size="small"
            link
            type="warning"
            @click="emit('audit', row as RoomItem)"
          >
            审核
          </el-button>
          <el-button
            v-if="row.status === 1 && can(PERM.ChatRoomStatus)"
            size="small"
            link
            type="warning"
            @click="emit('toggle-status', row as RoomItem, 3)"
          >
            禁用
          </el-button>
          <el-button
            v-if="row.status === 3 && can(PERM.ChatRoomStatus)"
            size="small"
            link
            type="success"
            @click="emit('toggle-status', row as RoomItem, 1)"
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
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'

defineProps<{
  list: RoomItem[]
  loading: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  posts: [row: RoomItem]
  messages: [row: RoomItem]
  edit: [row: RoomItem]
  audit: [row: RoomItem]
  'toggle-status': [row: RoomItem, status: number]
  'page-change': []
}>()

const { can } = usePermission()

function statusTag(status: number) {
  if (status === 1) return 'success'
  if (status === 0) return 'warning'
  if (status === 2) return 'info'
  return 'danger'
}
</script>

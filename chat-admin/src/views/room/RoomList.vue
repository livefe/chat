<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <el-button @click="mapVisible = true">全景地图</el-button>
            <el-button v-if="can(PERM.ChatRoomCreate)" type="primary" @click="handleAdd">
              新增聊天区
            </el-button>
          </div>
        </div>
      </template>

      <RoomFilter
        v-model:id="query.id"
        v-model:name="query.name"
        v-model:status="query.status"
        v-model:source="query.source"
        v-model:place-type="query.placeType"
        @search="handleSearch"
        @reset="handleReset"
      />

      <RoomTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @posts="goRoomPosts"
        @messages="goRoomMessages"
        @edit="handleEdit"
        @audit="openAudit"
        @toggle-status="toggleStatus"
        @page-change="fetchList"
      />
    </el-card>

    <RoomFormDialog
      :visible="formVisible"
      :submitting="submitting"
      :form="form"
      v-model:name="form.name"
      v-model:description="form.description"
      v-model:place-type="form.placeType"
      v-model:require-geofence="form.requireGeofence"
      @update:visible="formVisible = $event"
      @submit="submit"
      @shape-change="onShapeChange"
    />

    <RoomAuditDialog
      :visible="auditVisible"
      :room-id="auditRoomId"
      :submitting="auditSubmitting"
      @update:visible="auditVisible = $event"
      @approve="submitAudit(true)"
      @reject="submitAudit(false)"
    />

    <RoomMapDialog v-model:visible="mapVisible" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { auditRoom, getRoomList, updateRoomStatus, type RoomItem } from '@/api/room'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { usePermission } from '@/composables/usePermission'
import { useRoomForm } from '@/composables/useRoomForm'
import { useRouteQuery } from '@/composables/useRouteQuery'
import { MessageScope } from '@/constants/message'
import { PERM } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { confirmAndRun } from '@/utils/confirm'
import { pushMenuRoute } from '@/utils/navigation'
import RoomAuditDialog from './RoomAuditDialog.vue'
import RoomFilter from './RoomFilter.vue'
import RoomFormDialog from './RoomFormDialog.vue'
import RoomMapDialog from './RoomMapDialog.vue'
import RoomTable from './RoomTable.vue'

const router = useRouter()
const route = useRoute()
const employeeStore = useEmployeeStore()
const { can } = usePermission()
const mapVisible = ref(false)

const {
  loading,
  list,
  total,
  query,
  fetchList,
  handleSearch,
  handleReset: doReset,
} = usePaginatedList(
  (q) =>
    getRoomList({
      page: q.page,
      pageSize: q.pageSize,
      id: q.id,
      name: q.name || undefined,
      status: q.status,
      source: q.source,
      placeType: q.placeType,
    }),
  {
    page: 1,
    pageSize: 20,
    id: undefined as number | undefined,
    name: '',
    status: undefined as number | undefined,
    source: undefined as number | undefined,
    placeType: undefined as number | undefined,
  },
  '获取聊天区列表失败',
)

const { wrapReset } = useRouteQuery(
  query,
  [
    {
      key: 'status',
      apply: (v, q) => {
        q.status = Number(v)
      },
      clear: (q) => {
        q.status = undefined
      },
    },
    {
      key: 'roomId',
      apply: (v, q) => {
        const id = Number(v)
        if (!Number.isNaN(id) && id > 0) q.id = id
      },
      clear: (q) => {
        q.id = undefined
      },
    },
  ],
  {
    onApplied: () => {
      query.page = 1
      fetchList()
    },
  },
)

const handleReset = wrapReset(doReset)

const { formVisible, submitting, form, handleAdd, handleEdit, submit, onShapeChange } =
  useRoomForm(fetchList)

const auditVisible = ref(false)
const auditRoomId = ref(0)
const auditSubmitting = ref(false)

function goRoomPosts(row: RoomItem) {
  pushMenuRoute(router, employeeStore.permissions, PERM.PostPost, {
    roomId: String(row.id),
  })
}

function goRoomMessages(row: RoomItem) {
  pushMenuRoute(router, employeeStore.permissions, PERM.ChatMessage, {
    roomId: String(row.id),
    scope: String(MessageScope.Room),
  })
}

function openAudit(row: RoomItem) {
  auditRoomId.value = row.id
  auditVisible.value = true
}

async function submitAudit(approve: boolean) {
  if (!auditRoomId.value) return
  auditSubmitting.value = true
  try {
    await auditRoom(auditRoomId.value, approve)
    ElMessage.success(approve ? '已通过' : '已驳回')
    auditVisible.value = false
    fetchList()
  } catch (e) {
    showApiError(e, approve ? '审核通过失败' : '驳回失败')
  } finally {
    auditSubmitting.value = false
  }
}

async function toggleStatus(row: RoomItem, target: number) {
  await confirmAndRun(target === 1 ? '确定启用？' : '确定禁用？', async () => {
    await updateRoomStatus(row.id, target)
    fetchList()
  })
}
</script>

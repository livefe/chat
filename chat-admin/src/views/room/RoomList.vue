<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <el-button v-if="pendingCount" type="warning" plain @click="showPending">
              待审核 {{ pendingCount }}
            </el-button>
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
        v-model:user-id="query.userId"
        @search="handleSearch"
        @reset="handleReset"
      />

      <div v-if="canAudit && selection.length" class="batch-bar">
        <span>已选 {{ selection.length }} 个</span>
        <el-button size="small" type="success" @click="batchAudit(true)">批量通过</el-button>
        <el-button size="small" type="danger" @click="batchAudit(false)">批量驳回</el-button>
      </div>

      <RoomTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        :selectable="canAudit"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @posts="goRoomPosts"
        @messages="goRoomMessages"
        @edit="handleEdit"
        @audit="openAudit"
        @toggle-status="toggleStatus"
        @selection-change="selection = $event"
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
      @reject="(reason) => submitAudit(false, reason)"
    />

    <RoomMapDialog v-model:visible="mapVisible" />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'

import { auditRoom, batchAuditRoom, getRoomList, updateRoomStatus, type RoomItem } from '@/api/room'
import { useAdminBadges } from '@/composables/useAdminBadges'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { useRoomForm } from '@/composables/useRoomForm'
import { MessageScope } from '@/constants/message'
import { PERM } from '@/constants/permission'
import { RoomStatus } from '@/constants/room'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { pushMenuRoute } from '@/utils/navigation'
import { confirmWithReason } from '@/utils/promptReason'
import RoomAuditDialog from './RoomAuditDialog.vue'
import RoomFilter from './RoomFilter.vue'
import RoomFormDialog from './RoomFormDialog.vue'
import RoomMapDialog from './RoomMapDialog.vue'
import RoomTable from './RoomTable.vue'

const router = useRouter()
const route = useRoute()
const employeeStore = useEmployeeStore()
const { can } = usePermission()
const { badgeFor, refreshBadges } = useAdminBadges()

const mapVisible = ref(false)
const selection = ref<RoomItem[]>([])
const canAudit = computed(() => can(PERM.ChatRoomAudit))
const pendingCount = computed(() => badgeFor(route.path) ?? 0)

const { query, list, total, loading, error, fetchList, handleSearch, handleReset } = useListQuery({
  fetch: (q) => getRoomList(q),
  defaults: {
    page: 1,
    pageSize: 20,
    id: undefined as number | undefined,
    name: '',
    status: undefined as number | undefined,
    source: undefined as number | undefined,
    placeType: undefined as number | undefined,
    userId: undefined as number | undefined,
  },
  schema: {
    id: 'number',
    name: 'string',
    status: 'number',
    source: 'number',
    placeType: 'number',
    userId: 'number',
  },
  errorMessage: '获取聊天区列表失败',
})

function showPending() {
  void router.replace({ path: route.path, query: { status: String(RoomStatus.Pending) } })
}

const { formVisible, submitting, form, handleAdd, handleEdit, submit, onShapeChange } =
  useRoomForm(fetchList)

const auditVisible = ref(false)
const auditRoomId = ref(0)
const auditSubmitting = ref(false)

function goRoomPosts(row: RoomItem) {
  pushMenuRoute(router, employeeStore.permissions, PERM.PostPost, { roomId: String(row.id) })
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

async function submitAudit(approve: boolean, reason = '') {
  if (!auditRoomId.value) return
  auditSubmitting.value = true
  try {
    await auditRoom(auditRoomId.value, approve, reason)
    ElMessage.success(approve ? '已通过' : '已驳回')
    auditVisible.value = false
    void refreshBadges({ force: true })
    await fetchList()
  } catch (e) {
    showApiError(e, approve ? '审核通过失败' : '驳回失败')
  } finally {
    auditSubmitting.value = false
  }
}

async function batchAudit(approve: boolean) {
  const rows = selection.value
  if (!rows.length) return
  const action = approve ? '通过' : '驳回'
  const names = rows
    .slice(0, 3)
    .map((r) => `「${r.name}」`)
    .join('、')
  const message = `确定批量${action}${names}${rows.length > 3 ? ` 等 ${rows.length} 个聊天区` : ''}？`
  let reason = ''
  if (approve) {
    try {
      await ElMessageBox.confirm(message, '批量通过', { type: 'warning' })
    } catch {
      return
    }
  } else {
    const input = await confirmWithReason(message, {
      title: '批量驳回',
      required: true,
      inputPlaceholder: '请输入驳回原因（申请人可见）',
    })
    if (input === null) return
    reason = input
  }
  try {
    const { data } = await batchAuditRoom(
      rows.map((r) => r.id),
      approve,
      reason,
    )
    ElMessage.success(
      `已${action} ${data.updated} 个${data.skipped ? `，跳过 ${data.skipped} 个` : ''}`,
    )
  } catch (e) {
    showApiError(e, `批量${action}失败`)
    return
  }
  void refreshBadges({ force: true })
  await fetchList()
}

async function toggleStatus(row: RoomItem, target: RoomStatus) {
  const enable = target === RoomStatus.Enabled
  const message = enable
    ? `确定启用「${row.name}」？`
    : `确定禁用「${row.name}」？禁用后该区在线用户会被移出频道，且无法再进入或发言。`
  try {
    await ElMessageBox.confirm(message, enable ? '启用聊天区' : '禁用聊天区', { type: 'warning' })
  } catch {
    return
  }
  try {
    await updateRoomStatus(row.id, target)
    ElMessage.success(enable ? '已启用' : '已禁用')
    await fetchList()
  } catch (e) {
    showApiError(e, enable ? '启用失败' : '禁用失败')
  }
}
</script>

<style scoped>
.batch-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
  padding: 8px 12px;
  border-radius: 4px;
  background: var(--el-color-primary-light-9);
  font-size: 13px;
}
</style>

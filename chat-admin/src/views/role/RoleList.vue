<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div v-if="can(PERM.OrgRoleCreate)" class="card-header-actions">
            <el-button type="primary" @click="handleAdd">新增角色</el-button>
          </div>
        </div>
      </template>

      <RoleFilter
        v-model:name="query.name"
        v-model:code="query.code"
        v-model:status="query.status"
        @search="handleSearch"
        @reset="handleReset"
      />

      <RoleTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @edit="handleEdit"
        @permissions="openPermissions"
        @members="openMembers"
        @delete="handleDelete"
        @page-change="fetchList"
      />
    </el-card>

    <RoleFormDialog v-model="formVisible" :model="form" @success="fetchList" />

    <RolePermissionDialog v-model="permVisible" :role="currentRole" />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { deleteRole, getRoleList, type RoleItem } from '@/api/role'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { DEFAULT_PAGE_SIZE } from '@/constants'
import { PERM } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'
import { confirmAndRun } from '@/utils/confirm'
import { pushMenuRoute } from '@/utils/navigation'
import RoleFilter from './RoleFilter.vue'
import RoleTable from './RoleTable.vue'
import RoleFormDialog, { type RoleFormModel } from './RoleFormDialog.vue'
import RolePermissionDialog from './RolePermissionDialog.vue'

const { can } = usePermission()
const route = useRoute()
const router = useRouter()
const employeeStore = useEmployeeStore()

const { loading, error, list, total, query, fetchList, handleSearch, handleReset, afterRemove } =
  useListQuery({
    fetch: (q) => getRoleList(q),
    defaults: {
      page: 1,
      pageSize: DEFAULT_PAGE_SIZE,
      name: '',
      code: '',
      status: undefined as number | undefined,
    },
    schema: { name: 'string', code: 'string', status: 'number' },
    errorMessage: '获取角色列表失败',
  })

const formVisible = ref(false)
const form = reactive<RoleFormModel>({ id: 0, name: '', code: '', description: '', status: 1 })
const permVisible = ref(false)
const currentRole = ref<RoleItem | null>(null)

function handleAdd() {
  Object.assign(form, { id: 0, name: '', code: '', description: '', status: 1 })
  formVisible.value = true
}

function handleEdit(row: RoleItem) {
  Object.assign(form, {
    id: row.id,
    name: row.name,
    code: row.code,
    description: row.description,
    status: row.status,
  })
  formVisible.value = true
}

function openPermissions(row: RoleItem) {
  currentRole.value = row
  permVisible.value = true
}

function openMembers(row: RoleItem) {
  pushMenuRoute(router, employeeStore.permissions, PERM.OrgEmployee, { roleId: String(row.id) })
}

async function handleDelete(row: RoleItem) {
  const impact =
    row.memberCount > 0
      ? `该角色当前有 ${row.memberCount} 名员工，删除后他们将失去此角色带来的全部权限。`
      : '该角色当前没有员工。'
  await confirmAndRun(
    `确定删除角色「${row.name}」吗？${impact}此操作不可恢复。`,
    async () => {
      await deleteRole(row.id)
      await afterRemove()
    },
    '已删除',
  )
}
</script>

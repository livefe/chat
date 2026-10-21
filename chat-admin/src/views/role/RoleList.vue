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
        v-model:status="query.status"
        @search="handleSearch"
        @reset="handleReset"
      />

      <RoleTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @edit="handleEdit"
        @permissions="openPermissions"
        @delete="handleDelete"
        @page-change="fetchList"
      />
    </el-card>

    <RoleFormDialog v-model="formVisible" :model="form" @success="fetchList" />

    <RolePermissionDialog v-model="permVisible" :role="currentRole" />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { deleteRole, getRoleList, type RoleItem } from '@/api/role'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { confirmAndRun } from '@/utils/confirm'
import RoleFilter from './RoleFilter.vue'
import RoleTable from './RoleTable.vue'
import RoleFormDialog, { type RoleFormModel } from './RoleFormDialog.vue'
import RolePermissionDialog from './RolePermissionDialog.vue'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

const { can } = usePermission()
const route = useRoute()
const { loading, list, total, query, fetchList, handleSearch, handleReset } = usePaginatedList(
  getRoleList,
  { page: 1, pageSize: 10, name: '', status: undefined as number | undefined },
  '获取角色列表失败',
)

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

async function handleDelete(row: RoleItem) {
  await confirmAndRun(
    '确定要删除该角色吗？',
    async () => {
      await deleteRole(row.id)
      fetchList()
    },
    '删除成功',
  )
}

onMounted(fetchList)
</script>

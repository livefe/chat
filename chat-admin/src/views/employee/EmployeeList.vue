<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div v-if="can(PERM.OrgEmployeeCreate)" class="card-header-actions">
            <el-button type="primary" @click="handleAdd">新增员工</el-button>
          </div>
        </div>
      </template>

      <EmployeeFilter
        v-model:mobile="query.mobile"
        v-model:name="query.name"
        v-model:department-id="query.departmentId"
        v-model:status="query.status"
        :department-options="departmentOptions"
        @search="handleSearch"
        @reset="handleReset"
      />

      <EmployeeTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @edit="handleEdit"
        @roles="openRoles"
        @toggle-status="handleToggleStatus"
        @page-change="fetchList"
      />
    </el-card>

    <EmployeeFormDialog
      v-model="formVisible"
      :model="form"
      :department-options="departmentOptions"
      @success="fetchList"
    />

    <EmployeeRoleDialog v-model="roleVisible" :employee="currentEmployee" @success="fetchList" />
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { getEmployeeList, updateEmployeeStatus, type EmployeeItem } from '@/api/employee'
import { getDepartmentOptions, type DepartmentOption } from '@/api/department'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { useRouteQuery } from '@/composables/useRouteQuery'
import { showApiError } from '@/utils/apiError'
import { confirmAndRun } from '@/utils/confirm'
import EmployeeFilter from './EmployeeFilter.vue'
import EmployeeTable from './EmployeeTable.vue'
import EmployeeFormDialog, { type EmployeeFormModel } from './EmployeeFormDialog.vue'
import EmployeeRoleDialog from './EmployeeRoleDialog.vue'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

const { can } = usePermission()
const route = useRoute()
const departmentOptions = ref<DepartmentOption[]>([])
const {
  loading,
  list,
  total,
  query,
  fetchList,
  handleSearch,
  handleReset: doReset,
} = usePaginatedList(
  getEmployeeList,
  {
    page: 1,
    pageSize: 10,
    mobile: '',
    name: '',
    departmentId: undefined as number | undefined,
    status: undefined as number | undefined,
  },
  '获取员工列表失败',
)

const { wrapReset } = useRouteQuery(
  query,
  [
    {
      key: 'departmentId',
      apply: (v, q) => {
        const id = Number(v)
        if (!Number.isNaN(id) && id > 0) q.departmentId = id
      },
      clear: (q) => {
        q.departmentId = undefined
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

const formVisible = ref(false)
const form = reactive<EmployeeFormModel>({
  id: 0,
  mobile: '',
  name: '',
  employeeNo: '',
  departmentId: undefined,
  remark: '',
  status: 1,
})
const roleVisible = ref(false)
const currentEmployee = ref<EmployeeItem | null>(null)

function handleAdd() {
  Object.assign(form, {
    id: 0,
    mobile: '',
    name: '',
    employeeNo: '',
    departmentId: undefined,
    remark: '',
    status: 1,
  })
  formVisible.value = true
}

function handleEdit(row: EmployeeItem) {
  Object.assign(form, {
    id: row.id,
    mobile: row.mobile,
    name: row.name,
    employeeNo: row.employeeNo,
    departmentId: row.departmentId || undefined,
    remark: row.remark,
    status: row.status,
  })
  formVisible.value = true
}

function openRoles(row: EmployeeItem) {
  currentEmployee.value = row
  roleVisible.value = true
}

async function handleToggleStatus(row: EmployeeItem) {
  const next = row.status === 1 ? 0 : 1
  await confirmAndRun(`确定要${next === 1 ? '启用' : '禁用'}该员工吗？`, async () => {
    await updateEmployeeStatus(row.id, next)
    await fetchList()
  })
}

async function loadDepartmentOptions() {
  try {
    const res = await getDepartmentOptions()
    departmentOptions.value = Array.isArray(res.data) ? res.data : []
  } catch (e) {
    showApiError(e, '加载部门列表失败')
  }
}

onMounted(() => {
  loadDepartmentOptions()
})
</script>

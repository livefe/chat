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
        v-model:employee-no="query.employeeNo"
        v-model:department-id="query.departmentId"
        v-model:role-id="query.roleId"
        v-model:status="query.status"
        v-model:time-range="timeRange"
        :department-options="departmentOptions"
        :role-options="roleOptions"
        @search="handleSearch"
        @reset="handleReset"
      />

      <el-alert
        v-if="query.id"
        type="info"
        show-icon
        :closable="false"
        class="list-scope-tip"
        :title="`仅显示员工 ID ${query.id}`"
      >
        <el-button link type="primary" @click="clearIdFilter">查看全部</el-button>
      </el-alert>

      <EmployeeTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @edit="handleEdit"
        @roles="openRoles"
        @login-records="openLoginRecords"
        @audit="openAudit"
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

    <EmployeeLoginRecordDrawer v-model="loginVisible" :employee="currentEmployee" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref, toRef } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  getEmployeeList,
  getEmployeeRoleOptions,
  updateEmployeeStatus,
  type EmployeeItem,
  type EmployeeRoleBrief,
} from '@/api/employee'
import { getDepartmentOptions, type DepartmentOption } from '@/api/department'
import { useDateRangeModel } from '@/composables/useDateRangeModel'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { DEFAULT_PAGE_SIZE } from '@/constants'
import { PERM } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { confirmAndRun } from '@/utils/confirm'
import { pushMenuRoute } from '@/utils/navigation'
import EmployeeFilter from './EmployeeFilter.vue'
import EmployeeTable from './EmployeeTable.vue'
import EmployeeFormDialog, { type EmployeeFormModel } from './EmployeeFormDialog.vue'
import EmployeeRoleDialog from './EmployeeRoleDialog.vue'
import EmployeeLoginRecordDrawer from './EmployeeLoginRecordDrawer.vue'

const { can } = usePermission()
const route = useRoute()
const router = useRouter()
const employeeStore = useEmployeeStore()
const departmentOptions = ref<DepartmentOption[]>([])
const roleOptions = ref<EmployeeRoleBrief[]>([])

const { loading, error, list, total, query, fetchList, handleSearch, handleReset } = useListQuery({
  fetch: (q) => getEmployeeList(q),
  defaults: {
    page: 1,
    pageSize: DEFAULT_PAGE_SIZE,
    id: undefined as number | undefined,
    mobile: '',
    name: '',
    employeeNo: '',
    departmentId: undefined as number | undefined,
    roleId: undefined as number | undefined,
    status: undefined as number | undefined,
    startTime: '',
    endTime: '',
  },
  schema: {
    id: 'number',
    mobile: 'string',
    name: 'string',
    employeeNo: 'string',
    departmentId: 'number',
    roleId: 'number',
    status: 'number',
    startTime: 'string',
    endTime: 'string',
  },
  errorMessage: '获取员工列表失败',
})

const timeRange = useDateRangeModel(toRef(query, 'startTime'), toRef(query, 'endTime'))

function clearIdFilter() {
  query.id = undefined
  void handleSearch()
}

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
const loginVisible = ref(false)
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

function openLoginRecords(row: EmployeeItem) {
  currentEmployee.value = row
  loginVisible.value = true
}

function openAudit(row: EmployeeItem) {
  pushMenuRoute(router, employeeStore.permissions, PERM.OpsAuditLog, {
    employeeId: String(row.id),
  })
}

async function handleToggleStatus(row: EmployeeItem) {
  const label = row.name || row.mobile
  if (row.status === 1) {
    await confirmAndRun(
      `确定禁用员工「${label}」吗？禁用后该员工会被立即强制下线，且无法再登录后台。`,
      async () => {
        await updateEmployeeStatus(row.id, 0)
        await fetchList()
      },
      '已禁用',
    )
    return
  }
  await confirmAndRun(
    `确定启用员工「${label}」吗？`,
    async () => {
      await updateEmployeeStatus(row.id, 1)
      await fetchList()
    },
    '已启用',
  )
}

async function loadOptions() {
  try {
    const [depts, roles] = await Promise.all([getDepartmentOptions(), getEmployeeRoleOptions()])
    departmentOptions.value = Array.isArray(depts.data) ? depts.data : []
    roleOptions.value = roles.data.options
  } catch (e) {
    showApiError(e, '加载筛选项失败')
  }
}

onMounted(loadOptions)
</script>

<style scoped>
.list-scope-tip {
  margin-bottom: 12px;
}
</style>

<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div v-if="can(PERM.OrgDepartmentCreate)" class="card-header-actions">
            <el-button type="primary" @click="formRef?.openCreate()">新增部门</el-button>
          </div>
        </div>
      </template>

      <DepartmentFilter
        v-model:name="query.name"
        v-model:status="query.status"
        @search="applyFilter"
        @reset="handleReset"
      />

      <DepartmentTable
        :data="displayTree"
        :loading="loading"
        :error="loadError"
        :expand-all="isFiltering"
        :table-key="tableKey"
        @retry="fetchList()"
        @detail="openDetail"
        @add-child="(row) => formRef?.openCreate(row)"
        @edit="(row) => formRef?.openEdit(row)"
        @toggle-status="handleToggleStatus"
      />
    </el-card>

    <DepartmentDetailDrawer
      v-model="detailVisible"
      :department="selected"
      @employees="goEmployees"
      @add-child="(row) => formRef?.openCreate(row)"
      @edit="(row) => formRef?.openEdit(row)"
    />

    <DepartmentFormDialog ref="formRef" :tree="allTree" @success="fetchList" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getDepartmentList, updateDepartmentStatus, type DepartmentItem } from '@/api/department'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { confirmAndRun } from '@/utils/confirm'
import { pushMenuRoute } from '@/utils/navigation'
import DepartmentDetailDrawer from './DepartmentDetailDrawer.vue'
import DepartmentFilter from './DepartmentFilter.vue'
import DepartmentFormDialog from './DepartmentFormDialog.vue'
import DepartmentTable from './DepartmentTable.vue'

const router = useRouter()
const route = useRoute()
const employeeStore = useEmployeeStore()
const { can } = usePermission()

const loading = ref(false)
const loadError = ref(false)
const allTree = ref<DepartmentItem[]>([])
const selected = ref<DepartmentItem | null>(null)
const detailVisible = ref(false)
const formRef = ref<InstanceType<typeof DepartmentFormDialog>>()
const filterActive = reactive({
  name: '',
  status: undefined as number | undefined,
})
const query = reactive({
  name: '',
  status: undefined as number | undefined,
})

function filterTree(
  nodes: DepartmentItem[],
  nameQ: string,
  status: number | undefined,
): DepartmentItem[] {
  const out: DepartmentItem[] = []
  for (const n of nodes) {
    const children = filterTree(n.children || [], nameQ, status)
    const nameOk = !nameQ || n.name.includes(nameQ)
    const statusOk = status === undefined || n.status === status
    if ((nameOk && statusOk) || children.length) {
      out.push({ ...n, children })
    }
  }
  return out
}

const isFiltering = computed(() => !!filterActive.name.trim() || filterActive.status !== undefined)

const displayTree = computed(() =>
  filterTree(allTree.value, filterActive.name.trim(), filterActive.status),
)

const tableKey = computed(
  () => `${filterActive.name}|${filterActive.status ?? 'all'}|${allTree.value.length}`,
)

function findNode(nodes: DepartmentItem[], id: number): DepartmentItem | null {
  for (const n of nodes) {
    if (n.id === id) return n
    if (n.children?.length) {
      const hit = findNode(n.children, id)
      if (hit) return hit
    }
  }
  return null
}

async function fetchList(preferId?: number) {
  loading.value = true
  loadError.value = false
  try {
    const { data } = await getDepartmentList()
    allTree.value = data.list || []
    const keepId = preferId ?? selected.value?.id
    if (keepId) {
      selected.value = findNode(allTree.value, keepId)
    }
  } catch (e) {
    allTree.value = []
    loadError.value = true
    showApiError(e, '获取部门列表失败')
  } finally {
    loading.value = false
  }
}

function applyFilter() {
  filterActive.name = query.name
  filterActive.status = query.status
}

function handleReset() {
  query.name = ''
  query.status = undefined
  applyFilter()
}

function openDetail(row: DepartmentItem) {
  if (!can(PERM.OrgDepartmentDetail)) return
  selected.value = row
  detailVisible.value = true
}

function goEmployees(row: DepartmentItem) {
  pushMenuRoute(router, employeeStore.permissions, PERM.OrgEmployee, {
    departmentId: String(row.id),
  })
}

async function handleToggleStatus(row: DepartmentItem) {
  const next = row.status === 1 ? 0 : 1
  if (next === 0 && row.employeeCount > 0) {
    ElMessage.warning(`该部门下仍有 ${row.employeeCount} 名员工，无法禁用`)
    return
  }
  if (next === 0 && row.children?.length) {
    ElMessage.warning('请先处理下级部门后再禁用')
    return
  }
  const message =
    next === 1
      ? `确定启用部门「${row.name}」吗？`
      : `确定禁用部门「${row.name}」吗？禁用后新增或编辑员工时将不能再选择该部门。`
  await confirmAndRun(
    message,
    async () => {
      await updateDepartmentStatus(row.id, next)
      await fetchList(row.id)
    },
    next === 1 ? '已启用' : '已禁用',
  )
}

onMounted(() => fetchList())
</script>

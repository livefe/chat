<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div v-if="can(PERM.OrgPermissionCreate)" class="card-header-actions">
            <el-button type="primary" @click="formRef?.openCreate()">新增权限</el-button>
          </div>
        </div>
      </template>

      <el-alert
        type="info"
        :closable="false"
        show-icon
        class="perm-model-hint"
        :title="PERMISSION_MODEL_HINT"
      />

      <PermissionFilter
        v-model:keyword="filter.keyword"
        v-model:type="filter.type"
        v-model:status="filter.status"
        @search="applyFilter"
        @reset="handleReset"
      />

      <PermissionTable
        :data="displayTree"
        :loading="loading"
        :error="loadError"
        :expand-all="isFiltering"
        :table-key="tableKey"
        @retry="fetchTree"
        @add-child="(row) => formRef?.openCreate(row)"
        @edit="(row) => formRef?.openEdit(row)"
        @delete="handleDelete"
      />
    </el-card>

    <PermissionFormDialog ref="formRef" :tree="allTree" @success="fetchTree" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { deletePermission, getPermissionTree, type PermissionTreeItem } from '@/api/permission'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { showApiError } from '@/utils/apiError'
import { confirmAndRun } from '@/utils/confirm'
import { PERMISSION_MODEL_HINT } from '@/utils/permission'
import PermissionFilter from './PermissionFilter.vue'
import PermissionFormDialog from './PermissionFormDialog.vue'
import PermissionTable from './PermissionTable.vue'

const { can } = usePermission()
const route = useRoute()
const formRef = ref<InstanceType<typeof PermissionFormDialog>>()

const loading = ref(false)
const loadError = ref(false)
const allTree = ref<PermissionTreeItem[]>([])

interface PermissionFilterState {
  keyword: string
  type?: number
  status?: number
}

const filter = reactive<PermissionFilterState>({ keyword: '' })
const active = reactive<PermissionFilterState>({ keyword: '' })

const isFiltering = computed(
  () => !!active.keyword.trim() || active.type !== undefined || active.status !== undefined,
)

function matches(n: PermissionTreeItem) {
  const kw = active.keyword.trim().toLowerCase()
  if (kw && !n.name.toLowerCase().includes(kw) && !n.code.toLowerCase().includes(kw)) {
    return false
  }
  if (active.type !== undefined && n.type !== active.type) return false
  if (active.status !== undefined && n.status !== active.status) return false
  return true
}

function filterTree(nodes: PermissionTreeItem[]): PermissionTreeItem[] {
  const out: PermissionTreeItem[] = []
  for (const n of nodes) {
    const children = filterTree(n.children || [])
    if (matches(n) || children.length) {
      out.push({ ...n, children: children.length ? children : undefined })
    }
  }
  return out
}

const displayTree = computed(() => (isFiltering.value ? filterTree(allTree.value) : allTree.value))

const tableKey = computed(
  () => `${active.keyword}|${active.type ?? ''}|${active.status ?? ''}|${allTree.value.length}`,
)

async function fetchTree() {
  loading.value = true
  loadError.value = false
  try {
    const { data } = await getPermissionTree()
    allTree.value = data
  } catch (e) {
    allTree.value = []
    loadError.value = true
    showApiError(e, '获取权限失败')
  } finally {
    loading.value = false
  }
}

function applyFilter() {
  Object.assign(active, filter)
}

function handleReset() {
  Object.assign(filter, { keyword: '', type: undefined, status: undefined })
  applyFilter()
}

async function handleDelete(row: PermissionTreeItem) {
  if (row.children?.length) {
    ElMessage.warning(`「${row.name}」下还有 ${row.children.length} 个子权限，请先删除子权限`)
    return
  }
  await confirmAndRun(
    `确定删除权限「${row.name}」（${row.code}）吗？已被角色引用的权限无法删除，请先从角色中移除。`,
    async () => {
      await deletePermission(row.id)
      await fetchTree()
    },
    '已删除',
  )
}

onMounted(fetchTree)
</script>

<style scoped>
.perm-model-hint {
  margin-bottom: 16px;
}
</style>

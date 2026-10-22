<template>
  <el-dialog
    v-model="visible"
    :title="`权限设置 - ${role?.name ?? ''}`"
    width="760px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <el-alert
      type="info"
      :closable="false"
      show-icon
      class="perm-hint"
      :title="PERMISSION_MODEL_HINT"
    />
    <div class="perm-toolbar">
      <div class="perm-filter">
        <el-input
          v-model="keyword"
          size="small"
          clearable
          placeholder="按名称或编码筛选"
          :prefix-icon="Search"
          style="width: 200px"
        />
        <el-checkbox-group v-model="typeFilter">
          <el-checkbox :value="PermissionType.Menu">菜单</el-checkbox>
          <el-checkbox :value="PermissionType.Button">按钮</el-checkbox>
        </el-checkbox-group>
      </div>
      <div class="perm-actions">
        <el-button size="small" link @click="toggleExpand">
          {{ expanded ? '全部收起' : '全部展开' }}
        </el-button>
        <el-button size="small" link @click="toggleCheckAll">
          {{ allChecked ? '取消全选' : '全选可见项' }}
        </el-button>
        <el-checkbox v-model="checkStrictly" size="small">父子独立</el-checkbox>
      </div>
    </div>
    <div v-if="outOfScopeCount > 0" class="form-tip perm-scope-tip">
      灰色项超出你自己的权限范围，不能修改；保存时这些权限会保持原样。
    </div>

    <el-tree
      ref="treeRef"
      v-loading="loading"
      :data="permissionTree"
      show-checkbox
      node-key="id"
      :props="treeProps"
      :check-strictly="effectiveStrict"
      :filter-node-method="filterNode"
      class="perm-tree"
      @check="refreshChecked"
    >
      <template #default="{ data }">
        <el-tooltip
          :disabled="inScope(data as PermissionTreeItem)"
          content="超出你的权限范围，不能修改"
          placement="right"
        >
          <span class="perm-node">
            <el-tag
              :type="permissionTypeTag(data.type)"
              size="small"
              effect="light"
              class="perm-tag"
            >
              {{ permissionTypeLabel(data.type) }}
            </el-tag>
            <span class="perm-name">{{ data.name }}</span>
            <span class="perm-code">{{ data.code }}</span>
          </span>
        </el-tooltip>
      </template>
    </el-tree>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" :disabled="loading" @click="handleSubmit">
          确定
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ElMessage, type FilterNodeMethodFunction, type TreeInstance } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { assignRolePermissions, getRolePermissions, type RoleItem } from '@/api/role'
import { getPermissionTree, type PermissionTreeItem } from '@/api/permission'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { PERMISSION_MODEL_HINT, permissionTypeLabel, permissionTypeTag } from '@/utils/permission'
import { PermissionType } from '@/constants/permission'

const visible = defineModel<boolean>({ required: true })
const props = defineProps<{ role: RoleItem | null }>()

const employeeStore = useEmployeeStore()
const loading = ref(false)
const saving = ref(false)
const treeRef = ref<TreeInstance>()
const permissionTree = ref<PermissionTreeItem[]>([])
const keyword = ref('')
const typeFilter = ref<number[]>([PermissionType.Menu, PermissionType.Button])
const expanded = ref(true)
const checkStrictly = ref(false)
const strictOverride = ref(false)
const effectiveStrict = computed(() => strictOverride.value || checkStrictly.value)
const checkedIds = ref<number[]>([])

/** 操作者自身拥有的权限；超出范围的节点不可勾选。 */
const scopeIds = computed(() => new Set(employeeStore.permissions.map((p) => p.id)))

function inScope(data: PermissionTreeItem) {
  return scopeIds.value.has(data.id)
}

const treeProps = {
  children: 'children',
  label: 'name',
  disabled: (data: Record<string, unknown>) => !scopeIds.value.has(data.id as number),
}

function flatten(nodes: PermissionTreeItem[]): PermissionTreeItem[] {
  return nodes.flatMap((n) => [n, ...(n.children ? flatten(n.children) : [])])
}

const allNodes = computed(() => flatten(permissionTree.value))
const outOfScopeCount = computed(() => allNodes.value.filter((n) => !inScope(n)).length)

function matches(data: PermissionTreeItem) {
  if (!typeFilter.value.includes(data.type)) return false
  const kw = keyword.value.trim().toLowerCase()
  if (!kw) return true
  return data.name.toLowerCase().includes(kw) || data.code.toLowerCase().includes(kw)
}

const filterNode: FilterNodeMethodFunction = (_value, data) => matches(data as PermissionTreeItem)

watch([keyword, typeFilter], () => {
  void nextTick(() => treeRef.value?.filter(keyword.value))
})

const selectableVisibleIds = computed(() =>
  allNodes.value.filter((n) => matches(n) && inScope(n)).map((n) => n.id),
)

const allChecked = computed(() => {
  const ids = selectableVisibleIds.value
  if (ids.length === 0) return false
  const checked = new Set(checkedIds.value)
  return ids.every((id) => checked.has(id))
})

function refreshChecked() {
  checkedIds.value = (treeRef.value?.getCheckedKeys() as number[] | undefined) ?? []
}

async function setChecked(ids: number[]) {
  strictOverride.value = true
  await nextTick()
  treeRef.value?.setCheckedKeys(ids)
  await nextTick()
  strictOverride.value = false
  refreshChecked()
}

watch(visible, async (v) => {
  if (!v || !props.role) return
  loading.value = true
  keyword.value = ''
  typeFilter.value = [PermissionType.Menu, PermissionType.Button]
  checkStrictly.value = false
  try {
    const [tree, rolePerms] = await Promise.all([
      getPermissionTree({ assignable: true }),
      getRolePermissions(props.role.id),
    ])
    permissionTree.value = tree.data
    await setChecked(rolePerms.data.map((p) => p.id))
    expandAll(true)
  } catch (e) {
    showApiError(e, '获取权限数据失败')
    visible.value = false
  } finally {
    loading.value = false
  }
})

function expandAll(value: boolean) {
  const nodesMap = treeRef.value?.store.nodesMap
  if (!nodesMap) return
  Object.values(nodesMap).forEach((n) => (n.expanded = value))
  expanded.value = value
}

function toggleExpand() {
  expandAll(!expanded.value)
}

async function toggleCheckAll() {
  const visibleIds = new Set(selectableVisibleIds.value)
  const current = new Set(checkedIds.value)
  if (allChecked.value) visibleIds.forEach((id) => current.delete(id))
  else visibleIds.forEach((id) => current.add(id))
  await setChecked([...current])
}

async function handleSubmit() {
  if (!props.role || !treeRef.value) return
  saving.value = true
  try {
    const ids = [
      ...treeRef.value.getCheckedKeys(),
      ...treeRef.value.getHalfCheckedKeys(),
    ] as number[]
    await assignRolePermissions(props.role.id, ids)
    ElMessage.success('权限已更新')
    visible.value = false
    await employeeStore.loadPermissions()
  } catch (e) {
    showApiError(e, '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.perm-hint {
  margin-bottom: 12px;
}

.perm-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 0 4px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  margin-bottom: 8px;
}

.perm-filter {
  display: flex;
  align-items: center;
  gap: 12px;
}

.perm-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.perm-scope-tip {
  margin: 0 4px 8px;
}

.perm-tree {
  max-height: 460px;
  overflow: auto;
}

.perm-node {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.perm-tag {
  flex-shrink: 0;
}

.perm-code {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
</style>

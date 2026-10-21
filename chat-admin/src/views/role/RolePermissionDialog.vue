<template>
  <el-dialog
    v-model="visible"
    :title="`权限设置 - ${role?.name ?? ''}`"
    width="720px"
    destroy-on-close
    :close-on-click-modal="false"
  >
    <div class="perm-toolbar">
      <el-checkbox-group v-model="typeFilter" class="perm-filter">
        <el-checkbox :label="PermissionType.Menu">菜单</el-checkbox>
        <el-checkbox :label="PermissionType.Button">按钮</el-checkbox>
        <el-checkbox :label="PermissionType.Data">数据</el-checkbox>
      </el-checkbox-group>
      <div class="perm-actions">
        <el-button size="small" link @click="toggleExpand">
          {{ expanded ? '全部收起' : '全部展开' }}
        </el-button>
        <el-button size="small" link @click="toggleCheckAll">
          {{ allChecked ? '取消全选' : '全部选中' }}
        </el-button>
        <el-checkbox v-model="checkStrictly" size="small">父子独立</el-checkbox>
      </div>
    </div>

    <el-tree
      ref="treeRef"
      :data="permissionTree"
      show-checkbox
      node-key="id"
      :props="{ children: 'children', label: 'name' }"
      :check-strictly="effectiveStrict"
      :filter-node-method="filterByType"
      class="perm-tree"
    >
      <template #default="{ data }">
        <span class="perm-node">
          <el-tag :type="permissionTypeTag(data.type)" size="small" effect="light" class="perm-tag">
            {{ permissionTypeLabel(data.type) }}
          </el-tag>
          <span class="perm-name">{{ data.name }}</span>
          <span class="perm-code">{{ data.code }}</span>
        </span>
      </template>
    </el-tree>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="handleSubmit">确定</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { ElMessage, type FilterNodeMethodFunction } from 'element-plus'
import { assignRolePermissions, getRolePermissions, type RoleItem } from '@/api/role'
import { getPermissionTree, type PermissionTreeItem } from '@/api/permission'
import { useEmployeeStore } from '@/stores/employee'
import { showApiError } from '@/utils/apiError'
import { permissionTypeLabel, permissionTypeTag } from '@/utils/permission'
import { PermissionType } from '@/constants/permission'

interface TreeNode {
  expanded: boolean
}

const visible = defineModel<boolean>({ required: true })
const props = defineProps<{ role: RoleItem | null }>()

const employeeStore = useEmployeeStore()
const saving = ref(false)
const treeRef = ref()
const permissionTree = ref<PermissionTreeItem[]>([])
const typeFilter = ref<number[]>([PermissionType.Menu, PermissionType.Button, PermissionType.Data])
const expanded = ref(true)
const checkStrictly = ref(false)
const strictOverride = ref(false)
const effectiveStrict = computed(() => strictOverride.value || checkStrictly.value)

const filterByType: FilterNodeMethodFunction = (_value, data) =>
  typeFilter.value.includes((data as PermissionTreeItem).type)

watch(typeFilter, () => {
  nextTick(() => treeRef.value?.filter(''))
})

function collectVisibleNodeIds(nodes: PermissionTreeItem[]): number[] {
  const out: number[] = []
  for (const n of nodes) {
    if (!typeFilter.value.includes(n.type)) continue
    out.push(n.id)
    if (n.children?.length) out.push(...collectVisibleNodeIds(n.children))
  }
  return out
}

const allChecked = computed(() => {
  if (!treeRef.value) return false
  const visibleIds = collectVisibleNodeIds(permissionTree.value)
  if (visibleIds.length === 0) return false
  const checked = new Set(treeRef.value.getCheckedKeys() as number[])
  return visibleIds.every((id) => checked.has(id))
})

watch(visible, async (v) => {
  if (!v || !props.role) return
  try {
    const [tree, rolePerms] = await Promise.all([
      getPermissionTree({ assignable: true }),
      getRolePermissions(props.role.id),
    ])
    permissionTree.value = tree.data
    const ids = rolePerms.data.map((p) => p.id)
    typeFilter.value = [PermissionType.Menu, PermissionType.Button, PermissionType.Data]
    expanded.value = true
    checkStrictly.value = false
    strictOverride.value = true
    await nextTick()
    treeRef.value?.setCheckedKeys(ids)
    await nextTick()
    strictOverride.value = false
    expandAll(true)
  } catch (e) {
    showApiError(e, '获取权限数据失败')
    visible.value = false
  }
})

function expandAll(value: boolean) {
  const store = treeRef.value?.store
  if (!store) return
  Object.values(store.nodesMap as Record<string, TreeNode>).forEach((n) => (n.expanded = value))
  expanded.value = value
}

function toggleExpand() {
  expandAll(!expanded.value)
}

async function toggleCheckAll() {
  if (!treeRef.value) return
  const visibleIds = new Set(collectVisibleNodeIds(permissionTree.value))
  strictOverride.value = true
  await nextTick()

  if (allChecked.value) {
    const keep = (treeRef.value.getCheckedKeys() as number[]).filter((id) => !visibleIds.has(id))
    treeRef.value.setCheckedKeys(keep)
  } else {
    const current = new Set(treeRef.value.getCheckedKeys() as number[])
    visibleIds.forEach((id) => current.add(id))
    treeRef.value.setCheckedKeys([...current])
  }

  await nextTick()
  strictOverride.value = false
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
    ElMessage.success('权限设置成功')
    visible.value = false
    await employeeStore.loadPermissions()
  } catch (e) {
    showApiError(e, '操作失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.perm-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 4px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
  margin-bottom: 8px;
}
.perm-filter {
  display: flex;
  gap: 8px;
}
.perm-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.perm-tree {
  max-height: 480px;
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

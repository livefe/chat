<template>
  <el-table
    :key="tableKey"
    :data="data"
    v-loading="loading"
    row-key="id"
    :default-expand-all="expandAll"
    :tree-props="{ children: 'children' }"
  >
    <template #empty>
      <TableEmpty :error="error" text="暂无权限" @retry="emit('retry')" />
    </template>
    <el-table-column prop="name" label="名称" min-width="220" />
    <el-table-column label="类型" width="80">
      <template #default="{ row }">
        <el-tag :type="permissionTypeTag(row.type)" size="small" effect="light">
          {{ row.typeText }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="code" label="编码" min-width="200" />
    <el-table-column label="路径" min-width="220" show-overflow-tooltip>
      <template #default="{ row }">
        <template v-if="row.path">
          <span v-if="row.method" class="perm-method">{{ row.method }}</span>
          {{ row.path }}
        </template>
        <span v-else class="sub">-</span>
      </template>
    </el-table-column>
    <el-table-column prop="sort" label="排序" width="70" />
    <el-table-column label="状态" width="80">
      <template #default="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
          {{ row.statusText }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="操作" width="170" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.OrgPermissionCreate) && row.type === PermissionType.Menu"
            size="small"
            link
            type="primary"
            @click="emit('add-child', row as PermissionTreeItem)"
          >
            下级
          </el-button>
          <el-button
            v-if="can(PERM.OrgPermissionUpdate)"
            size="small"
            link
            type="primary"
            @click="emit('edit', row as PermissionTreeItem)"
          >
            编辑
          </el-button>
          <el-button
            v-if="can(PERM.OrgPermissionDelete)"
            size="small"
            link
            type="danger"
            @click="emit('delete', row as PermissionTreeItem)"
          >
            删除
          </el-button>
        </div>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
import type { PermissionTreeItem } from '@/api/permission'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { usePermission } from '@/composables/usePermission'
import { PERM, PermissionType } from '@/constants/permission'
import { permissionTypeTag } from '@/utils/permission'

defineProps<{
  data: PermissionTreeItem[]
  loading: boolean
  error?: boolean
  expandAll: boolean
  tableKey: string
}>()

const emit = defineEmits<{
  retry: []
  'add-child': [row: PermissionTreeItem]
  edit: [row: PermissionTreeItem]
  delete: [row: PermissionTreeItem]
}>()

const { can } = usePermission()
</script>

<style scoped>
.perm-method {
  display: inline-block;
  margin-right: 4px;
  font-size: 12px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
}
</style>

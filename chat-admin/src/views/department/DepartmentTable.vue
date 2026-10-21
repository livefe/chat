<template>
  <el-table
    :key="tableKey"
    :data="data"
    v-loading="loading"
    row-key="id"
    empty-text="暂无部门"
    :default-expand-all="expandAll"
    :tree-props="{ children: 'children' }"
    @row-click="(row: DepartmentItem) => emit('detail', row)"
  >
    <el-table-column prop="name" label="部门名称" min-width="220" />
    <el-table-column prop="employeeCount" label="员工数" width="90" align="center" />
    <el-table-column prop="sort" label="排序" width="80" />
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
    <el-table-column label="操作" width="240" fixed="right">
      <template #default="{ row }">
        <div class="action-cell" @click.stop>
          <el-button
            v-if="can(PERM.OrgDepartmentDetail)"
            size="small"
            link
            type="primary"
            @click="emit('detail', row as DepartmentItem)"
          >
            详情
          </el-button>
          <el-button
            v-if="can(PERM.OrgDepartmentCreate) && row.level < 5"
            size="small"
            link
            type="primary"
            @click="emit('add-child', row as DepartmentItem)"
          >
            下级
          </el-button>
          <el-button
            v-if="can(PERM.OrgDepartmentUpdate)"
            size="small"
            link
            type="primary"
            @click="emit('edit', row as DepartmentItem)"
          >
            编辑
          </el-button>
          <el-button
            v-if="can(PERM.OrgDepartmentStatus)"
            size="small"
            link
            :type="row.status === 1 ? 'danger' : 'success'"
            @click="emit('toggle-status', row as DepartmentItem)"
          >
            {{ row.status === 1 ? '禁用' : '启用' }}
          </el-button>
        </div>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
import type { DepartmentItem } from '@/api/department'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'

defineProps<{
  data: DepartmentItem[]
  loading: boolean
  expandAll: boolean
  tableKey: string
}>()

const emit = defineEmits<{
  detail: [row: DepartmentItem]
  'add-child': [row: DepartmentItem]
  edit: [row: DepartmentItem]
  'toggle-status': [row: DepartmentItem]
}>()

const { can } = usePermission()
</script>

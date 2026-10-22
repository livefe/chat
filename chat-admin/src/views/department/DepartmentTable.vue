<template>
  <el-table
    :key="tableKey"
    :data="data"
    v-loading="loading"
    row-key="id"
    :default-expand-all="expandAll"
    :tree-props="{ children: 'children' }"
    @row-click="(row: DepartmentItem) => emit('detail', row)"
  >
    <template #empty>
      <TableEmpty :error="error" text="暂无部门" @retry="emit('retry')" />
    </template>
    <el-table-column prop="name" label="部门名称" min-width="220" />
    <el-table-column label="员工数" width="120" align="center">
      <template #header>
        <el-tooltip content="直属员工 / 含下级部门合计" placement="top">
          <span>员工数</span>
        </el-tooltip>
      </template>
      <template #default="{ row }">
        {{ row.employeeCount }}
        <span v-if="row.totalEmployeeCount !== row.employeeCount" class="sub">
          / {{ row.totalEmployeeCount }}
        </span>
      </template>
    </el-table-column>
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
import TableEmpty from '@/components/common/TableEmpty.vue'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'

defineProps<{
  data: DepartmentItem[]
  loading: boolean
  error?: boolean
  expandAll: boolean
  tableKey: string
}>()

const emit = defineEmits<{
  retry: []
  detail: [row: DepartmentItem]
  'add-child': [row: DepartmentItem]
  edit: [row: DepartmentItem]
  'toggle-status': [row: DepartmentItem]
}>()

const { can } = usePermission()
</script>

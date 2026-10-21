<template>
  <el-table :data="list" v-loading="loading" empty-text="暂无员工">
    <el-table-column prop="id" label="ID" width="80" />
    <el-table-column prop="mobile" label="手机号" width="130" />
    <el-table-column prop="name" label="姓名" width="120" />
    <el-table-column prop="employeeNo" label="工号" width="120" />
    <el-table-column prop="department" label="部门" width="140" />
    <el-table-column label="当前角色" min-width="180">
      <template #default="{ row }">
        <span v-if="!row.roles || row.roles.length === 0" class="sub">未分配</span>
        <el-space v-else wrap size="small">
          <el-tag v-for="r in row.roles" :key="r.id" type="info" effect="plain">{{
            r.name
          }}</el-tag>
        </el-space>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="createdAt" label="创建时间" width="180" />
    <el-table-column label="操作" width="200" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            size="small"
            link
            type="primary"
            @click="emit('edit', row as EmployeeItem)"
            v-if="can(PERM.OrgEmployeeUpdate)"
          >
            编辑
          </el-button>
          <el-button
            size="small"
            link
            type="primary"
            @click="emit('roles', row as EmployeeItem)"
            v-if="can(PERM.OrgEmployeeAssignRole)"
          >
            角色
          </el-button>
          <el-button
            size="small"
            link
            :type="row.status === 1 ? 'danger' : 'success'"
            @click="emit('toggle-status', row as EmployeeItem)"
            v-if="can(PERM.OrgEmployeeStatus)"
          >
            {{ row.status === 1 ? '禁用' : '启用' }}
          </el-button>
        </div>
      </template>
    </el-table-column>
  </el-table>

  <div class="pagination">
    <el-pagination
      v-model:current-page="page"
      v-model:page-size="pageSize"
      :total="total"
      :page-sizes="PAGE_SIZES"
      :layout="PAGINATION_LAYOUT"
      @size-change="emit('page-change')"
      @current-change="emit('page-change')"
    />
  </div>
</template>

<script setup lang="ts">
import type { EmployeeItem } from '@/api/employee'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

defineProps<{
  list: EmployeeItem[]
  loading: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  edit: [row: EmployeeItem]
  roles: [row: EmployeeItem]
  'toggle-status': [row: EmployeeItem]
  'page-change': []
}>()

const { can } = usePermission()
</script>

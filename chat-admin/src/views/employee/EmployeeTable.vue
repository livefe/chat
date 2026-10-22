<template>
  <el-table :data="list" v-loading="loading">
    <template #empty>
      <TableEmpty :error="error" text="暂无员工" @retry="emit('page-change')" />
    </template>
    <el-table-column label="ID" width="90">
      <template #default="{ row }"><CopyText :text="row.id" /></template>
    </el-table-column>
    <el-table-column label="员工" min-width="150">
      <template #default="{ row }">
        <div>{{ row.name || '-' }}</div>
        <div class="sub">{{ row.mobile }}</div>
      </template>
    </el-table-column>
    <el-table-column prop="employeeNo" label="工号" width="110">
      <template #default="{ row }">{{ row.employeeNo || '-' }}</template>
    </el-table-column>
    <el-table-column prop="department" label="部门" width="130" show-overflow-tooltip />
    <el-table-column label="当前角色" min-width="180">
      <template #default="{ row }">
        <span v-if="!row.roles || row.roles.length === 0" class="sub">未分配</span>
        <el-space v-else wrap size="small">
          <el-tooltip
            v-for="r in row.roles"
            :key="r.id"
            :disabled="r.status === 1"
            content="该角色已禁用，不再生效"
          >
            <el-tag :type="r.status === 1 ? 'info' : 'danger'" effect="plain">
              {{ r.name }}{{ r.status === 1 ? '' : '（已禁用）' }}
            </el-tag>
          </el-tooltip>
        </el-space>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="90">
      <template #default="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column label="最近登录" width="170">
      <template #default="{ row }">
        <template v-if="row.lastLoginAt">
          <div>{{ row.lastLoginAt }}</div>
          <div class="sub">{{ row.lastLoginIp }}</div>
        </template>
        <span v-else class="sub">从未登录</span>
      </template>
    </el-table-column>
    <el-table-column prop="createdAt" label="创建时间" width="170" />
    <el-table-column label="操作" width="230" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.OrgEmployeeUpdate)"
            size="small"
            link
            type="primary"
            @click="emit('edit', row as EmployeeItem)"
          >
            编辑
          </el-button>
          <el-button
            v-if="can(PERM.OrgEmployeeAssignRole)"
            size="small"
            link
            type="primary"
            @click="emit('roles', row as EmployeeItem)"
          >
            角色
          </el-button>
          <el-button
            size="small"
            link
            type="primary"
            @click="emit('login-records', row as EmployeeItem)"
          >
            登录记录
          </el-button>
          <el-button
            v-if="can(PERM.OpsAuditLog)"
            size="small"
            link
            type="primary"
            @click="emit('audit', row as EmployeeItem)"
          >
            操作审计
          </el-button>
          <el-button
            v-if="can(PERM.OrgEmployeeStatus)"
            size="small"
            link
            :type="row.status === 1 ? 'danger' : 'success'"
            @click="emit('toggle-status', row as EmployeeItem)"
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
import CopyText from '@/components/common/CopyText.vue'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

defineProps<{
  list: EmployeeItem[]
  loading: boolean
  error?: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  edit: [row: EmployeeItem]
  roles: [row: EmployeeItem]
  'login-records': [row: EmployeeItem]
  audit: [row: EmployeeItem]
  'toggle-status': [row: EmployeeItem]
  'page-change': []
}>()

const { can } = usePermission()
</script>

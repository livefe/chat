<template>
  <el-table :data="list" v-loading="loading">
    <template #empty>
      <TableEmpty :error="error" text="暂无角色" @retry="emit('page-change')" />
    </template>
    <el-table-column prop="id" label="ID" width="80" />
    <el-table-column prop="name" label="角色名称" min-width="140" />
    <el-table-column prop="code" label="角色编码" min-width="140" />
    <el-table-column prop="description" label="描述" min-width="160" show-overflow-tooltip />
    <el-table-column label="成员数" width="90" align="center">
      <template #default="{ row }">
        <el-button
          v-if="row.memberCount > 0 && can(PERM.OrgEmployee)"
          link
          type="primary"
          @click="emit('members', row as RoleItem)"
        >
          {{ row.memberCount }}
        </el-button>
        <span v-else>{{ row.memberCount }}</span>
      </template>
    </el-table-column>
    <el-table-column label="状态" width="90">
      <template #default="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="createdAt" label="创建时间" width="170" />
    <el-table-column label="操作" width="160" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.OrgRoleUpdate)"
            size="small"
            link
            type="primary"
            @click="emit('edit', row as RoleItem)"
          >
            编辑
          </el-button>
          <el-button
            v-if="can(PERM.OrgRoleAssignPerm)"
            size="small"
            link
            type="primary"
            @click="emit('permissions', row as RoleItem)"
          >
            权限
          </el-button>
          <el-button
            v-if="can(PERM.OrgRoleDelete) && row.code !== 'super_admin'"
            size="small"
            link
            type="danger"
            @click="emit('delete', row as RoleItem)"
          >
            删除
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
import type { RoleItem } from '@/api/role'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

defineProps<{
  list: RoleItem[]
  loading: boolean
  error?: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  edit: [row: RoleItem]
  permissions: [row: RoleItem]
  members: [row: RoleItem]
  delete: [row: RoleItem]
  'page-change': []
}>()

const { can } = usePermission()
</script>

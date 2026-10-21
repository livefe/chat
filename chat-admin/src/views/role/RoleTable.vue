<template>
  <el-table :data="list" v-loading="loading" empty-text="暂无角色">
    <el-table-column prop="id" label="ID" width="80" />
    <el-table-column prop="name" label="角色名称" />
    <el-table-column prop="code" label="角色编码" />
    <el-table-column prop="description" label="描述" show-overflow-tooltip />
    <el-table-column label="状态" width="100">
      <template #default="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="createdAt" label="创建时间" width="180" />
    <el-table-column label="操作" width="180" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            size="small"
            link
            type="primary"
            @click="emit('edit', row as RoleItem)"
            v-if="can(PERM.OrgRoleUpdate)"
          >
            编辑
          </el-button>
          <el-button
            size="small"
            link
            type="primary"
            @click="emit('permissions', row as RoleItem)"
            v-if="can(PERM.OrgRoleAssignPerm)"
          >
            权限
          </el-button>
          <el-button
            size="small"
            link
            type="danger"
            @click="emit('delete', row as RoleItem)"
            v-if="can(PERM.OrgRoleDelete)"
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
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

defineProps<{
  list: RoleItem[]
  loading: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  edit: [row: RoleItem]
  permissions: [row: RoleItem]
  delete: [row: RoleItem]
  'page-change': []
}>()

const { can } = usePermission()
</script>

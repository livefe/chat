<template>
  <el-table :data="list" v-loading="loading" empty-text="暂无权限">
    <el-table-column prop="id" label="ID" width="70" />
    <el-table-column prop="name" label="名称" />
    <el-table-column prop="code" label="编码" />
    <el-table-column prop="typeText" label="类型" width="80" />
    <el-table-column prop="path" label="路径" />
    <el-table-column prop="method" label="方法" width="80" />
    <el-table-column prop="sort" label="排序" width="80" />
    <el-table-column label="状态" width="80">
      <template #default="{ row }">
        <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.statusText }}</el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="createdAt" label="创建时间" width="180" />
    <el-table-column label="操作" width="120" fixed="right">
      <template #default="{ row }">
        <div class="action-cell">
          <el-button
            v-if="can(PERM.OrgPermissionUpdate)"
            size="small"
            link
            type="primary"
            @click="emit('edit', row as PermissionItem)"
          >
            编辑
          </el-button>
          <el-button
            v-if="can(PERM.OrgPermissionDelete)"
            size="small"
            link
            type="danger"
            @click="emit('delete', row as PermissionItem)"
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
import type { PermissionItem } from '@/api/permission'
import { usePermission } from '@/composables/usePermission'
import { PAGE_SIZES, PAGINATION_LAYOUT } from '@/constants'
import { PERM } from '@/constants/permission'

defineProps<{
  list: PermissionItem[]
  loading: boolean
  total: number
}>()

const page = defineModel<number>('page', { required: true })
const pageSize = defineModel<number>('pageSize', { required: true })

const emit = defineEmits<{
  edit: [row: PermissionItem]
  delete: [row: PermissionItem]
  'page-change': []
}>()

const { can } = usePermission()
</script>

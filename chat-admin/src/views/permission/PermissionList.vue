<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div v-if="can(PERM.OrgPermissionCreate)" class="card-header-actions">
            <el-button type="primary" @click="formRef?.openCreate()">新增权限</el-button>
          </div>
        </div>
      </template>

      <PermissionFilter
        v-model:name="query.name"
        v-model:type="query.type"
        v-model:status="query.status"
        @search="handleSearch"
        @reset="handleReset"
      />

      <PermissionTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @edit="(row) => formRef?.openEdit(row)"
        @delete="handleDelete"
        @page-change="fetchList"
      />
    </el-card>

    <PermissionFormDialog ref="formRef" @success="fetchList" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute } from 'vue-router'
import { deletePermission, getPermissionList, type PermissionItem } from '@/api/permission'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { confirmAndRun } from '@/utils/confirm'
import PermissionFilter from './PermissionFilter.vue'
import PermissionFormDialog from './PermissionFormDialog.vue'
import PermissionTable from './PermissionTable.vue'

const { can } = usePermission()
const route = useRoute()
const formRef = ref<InstanceType<typeof PermissionFormDialog>>()

const { loading, list, total, query, fetchList, handleSearch, handleReset } = usePaginatedList(
  getPermissionList,
  {
    page: 1,
    pageSize: 10,
    name: '',
    type: undefined as number | undefined,
    status: undefined as number | undefined,
  },
  '获取权限列表失败',
)

async function handleDelete(row: PermissionItem) {
  await confirmAndRun(
    '确定要删除该权限吗？',
    async () => {
      await deletePermission(row.id)
      fetchList()
    },
    '删除成功',
  )
}

onMounted(fetchList)
</script>

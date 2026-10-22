<template>
  <div class="user-tab">
    <el-table :data="list" v-loading="loading" size="small">
      <template #empty>
        <TableEmpty
          :error="error"
          :text="allowed ? '暂无帖子' : '暂无帖子菜单权限'"
          @retry="fetchList"
        />
      </template>
      <el-table-column label="ID" width="90">
        <template #default="{ row }"><CopyText :text="row.id" /></template>
      </el-table-column>
      <el-table-column label="标题" min-width="160" show-overflow-tooltip>
        <template #default="{ row }">
          <el-tag v-if="row.deleted" size="small" type="info" class="mr-4">已删除</el-tag>
          <span>{{ row.title }}</span>
        </template>
      </el-table-column>
      <el-table-column label="类型" width="70">
        <template #default="{ row }">{{ postCategoryLabel(row.category) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag size="small" :type="statusTag(row.status)">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="报名" width="110">
        <template #default="{ row }">
          <span v-if="row.extras?.needSignup">{{ formatExtraBrief(row as PostItem) }}</span>
          <span v-else class="sub">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="发布时间" width="150" />
      <el-table-column label="操作" width="70" fixed="right">
        <template #default="{ row }">
          <el-button size="small" link type="primary" @click="openPost(row.id)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div v-if="total > query.pageSize" class="user-tab__pager">
      <el-pagination
        v-model:current-page="query.page"
        :page-size="query.pageSize"
        :total="total"
        size="small"
        layout="total, prev, pager, next"
        @current-change="fetchList"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useRouter } from 'vue-router'

import { getPostList, type PostItem } from '@/api/post'
import CopyText from '@/components/common/CopyText.vue'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'
import { useUserDrawerStore } from '@/stores/userDrawer'
import { pushMenuRoute } from '@/utils/navigation'
import { formatExtraBrief, postCategoryLabel, statusLabel, statusTag } from '@/utils/postDisplay'

const props = defineProps<{ userId: number }>()

const router = useRouter()
const employeeStore = useEmployeeStore()
const drawer = useUserDrawerStore()
const { can } = usePermission()
const allowed = can(PERM.PostPost)

const { query, list, total, loading, error, fetchList, handleSearch } = useListQuery({
  fetch: (q) =>
    getPostList({
      page: q.page,
      pageSize: q.pageSize,
      publisherId: props.userId,
      includeDeleted: true,
    }),
  defaults: { page: 1, pageSize: 10 },
  errorMessage: '获取帖子失败',
  syncUrl: false,
  immediate: allowed,
})

watch(
  () => props.userId,
  () => {
    if (allowed) void handleSearch()
  },
)

function openPost(postId: number) {
  drawer.close()
  pushMenuRoute(router, employeeStore.permissions, PERM.PostPost, { id: String(postId) })
}
</script>

<style scoped>
.mr-4 {
  margin-right: 4px;
}

.user-tab__pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}
</style>

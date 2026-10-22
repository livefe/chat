<template>
  <div class="user-tab">
    <el-table :data="list" v-loading="loading" size="small">
      <template #empty>
        <TableEmpty
          :error="error"
          :text="allowed ? '暂无报名记录' : '暂无帖子菜单权限'"
          @retry="fetchList"
        />
      </template>
      <el-table-column label="活动" min-width="170" show-overflow-tooltip>
        <template #default="{ row }">
          <el-link type="primary" :underline="false" @click="openPost(row.postId)">
            {{ row.postTitle || `#${row.postId}` }}
          </el-link>
          <div class="sub">#{{ row.postId }} · {{ row.postStatusText }}</div>
        </template>
      </el-table-column>
      <el-table-column label="报名状态" width="90">
        <template #default="{ row }">
          <el-tag size="small" :type="signupStatusTag(row.signupStatus)">
            {{ row.signupStatusText }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="核验" width="150">
        <template #default="{ row }">
          <template v-if="row.checkedIn">
            <el-tag size="small" type="success">已核验</el-tag>
            <div class="sub">{{ row.checkedInAt }}</div>
          </template>
          <span v-else class="sub">-</span>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="报名时间" width="150" />
      <el-table-column label="取消时间" width="150">
        <template #default="{ row }">{{ row.cancelledAt || '-' }}</template>
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

import { getUserSignups } from '@/api/post'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { useListQuery } from '@/composables/useListQuery'
import { usePermission } from '@/composables/usePermission'
import { PERM } from '@/constants/permission'
import { useEmployeeStore } from '@/stores/employee'
import { useUserDrawerStore } from '@/stores/userDrawer'
import { pushMenuRoute } from '@/utils/navigation'
import { signupStatusTag } from '@/utils/postDisplay'

const props = defineProps<{ userId: number }>()

const router = useRouter()
const employeeStore = useEmployeeStore()
const drawer = useUserDrawerStore()
const { can } = usePermission()
const allowed = can(PERM.PostPost)

const { query, list, total, loading, error, fetchList, handleSearch } = useListQuery({
  fetch: (q) => getUserSignups({ userId: props.userId, page: q.page, pageSize: q.pageSize }),
  defaults: { page: 1, pageSize: 10 },
  errorMessage: '获取报名记录失败',
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
.user-tab__pager {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}
</style>

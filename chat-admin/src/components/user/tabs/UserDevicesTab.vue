<template>
  <div class="user-tab">
    <el-table :data="list" v-loading="loading" size="small">
      <template #empty>
        <TableEmpty
          :error="error"
          text="暂无设备（用户未开启推送或已退出登录）"
          @retry="fetchList"
        />
      </template>
      <el-table-column label="平台" width="90">
        <template #default="{ row }">{{ row.platform || '-' }}</template>
      </el-table-column>
      <el-table-column label="推送通道" width="90">
        <template #default="{ row }">{{ row.vendor || '-' }}</template>
      </el-table-column>
      <el-table-column label="App 版本" width="90">
        <template #default="{ row }">{{ row.appVersion || '-' }}</template>
      </el-table-column>
      <el-table-column label="系统版本" width="90">
        <template #default="{ row }">{{ row.osVersion || '-' }}</template>
      </el-table-column>
      <el-table-column label="推送" width="80">
        <template #default="{ row }">
          <el-tag :type="row.pushEnabled ? 'success' : 'info'" size="small">
            {{ row.pushEnabled ? '开启' : '关闭' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="最后活跃" width="150">
        <template #default="{ row }">{{ formatUnixSeconds(row.lastActiveAt) }}</template>
      </el-table-column>
      <el-table-column label="设备 ID" min-width="140">
        <template #default="{ row }"><CopyText :text="row.deviceId" /></template>
      </el-table-column>
    </el-table>
    <el-pagination
      v-if="total > query.pageSize"
      v-model:current-page="query.page"
      class="user-tab__pager"
      small
      layout="total, prev, pager, next"
      :page-size="query.pageSize"
      :total="total"
      @current-change="fetchList"
    />
  </div>
</template>

<script setup lang="ts">
import { getUserDevices } from '@/api/user'
import CopyText from '@/components/common/CopyText.vue'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { useListQuery } from '@/composables/useListQuery'
import { formatUnixSeconds } from '@/utils/datetime'

const props = defineProps<{ userId: number }>()

const { query, list, total, loading, error, fetchList } = useListQuery({
  fetch: (q) => getUserDevices(q),
  defaults: { page: 1, pageSize: 10, userId: props.userId },
  errorMessage: '获取设备失败',
  syncUrl: false,
})
</script>

<template>
  <el-drawer v-model="visible" :title="title" size="720px" destroy-on-close>
    <el-table :data="list" v-loading="loading" size="small">
      <template #empty>
        <TableEmpty :error="error" text="暂无登录记录" @retry="fetchList" />
      </template>
      <el-table-column prop="createdAt" label="时间" width="170" />
      <el-table-column label="结果" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">
            {{ row.statusText }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="ip" label="IP" width="140" />
      <el-table-column label="失败原因" min-width="140" show-overflow-tooltip>
        <template #default="{ row }">{{ row.remark || '-' }}</template>
      </el-table-column>
      <el-table-column prop="userAgent" label="设备" min-width="180" show-overflow-tooltip />
    </el-table>
    <div class="pagination">
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.pageSize"
        :total="total"
        :page-sizes="PAGE_SIZES"
        layout="total, sizes, prev, pager, next"
        size="small"
        @size-change="fetchList"
        @current-change="fetchList"
      />
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { getEmployeeLoginRecords, type EmployeeItem } from '@/api/employee'
import TableEmpty from '@/components/common/TableEmpty.vue'
import { useListQuery } from '@/composables/useListQuery'
import { DEFAULT_PAGE_SIZE, PAGE_SIZES } from '@/constants'

const visible = defineModel<boolean>({ required: true })
const props = defineProps<{ employee: EmployeeItem | null }>()

const title = computed(() => `登录记录 - ${props.employee?.name || props.employee?.mobile || ''}`)

const { query, list, total, loading, error, fetchList, handleSearch } = useListQuery({
  fetch: (q) => getEmployeeLoginRecords(q),
  defaults: { page: 1, pageSize: DEFAULT_PAGE_SIZE, employeeId: 0 },
  errorMessage: '获取登录记录失败',
  syncUrl: false,
  immediate: false,
})

watch(visible, (v) => {
  if (!v || !props.employee) return
  query.employeeId = props.employee.id
  void handleSearch()
})
</script>

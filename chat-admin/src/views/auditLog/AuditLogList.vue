<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
        </div>
      </template>

      <AuditLogFilter
        v-model:employee-id="query.employeeId"
        v-model:module="query.module"
        v-model:action="query.action"
        v-model:keyword="query.keyword"
        v-model:time-range="timeRange"
        :module-options="moduleOptions"
        :action-options="actionOptions"
        @search="handleSearch"
        @reset="handleReset"
      />

      <AuditLogTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @detail="(row) => detailRef?.open(row)"
        @page-change="fetchList"
      />
    </el-card>

    <AuditLogDetailDrawer ref="detailRef" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { getAuditLogList, type AuditActionOption, type AuditModuleOption } from '@/api/auditLog'
import { usePaginatedList } from '@/composables/usePaginatedList'
import AuditLogDetailDrawer from './AuditLogDetailDrawer.vue'
import AuditLogFilter from './AuditLogFilter.vue'
import AuditLogTable from './AuditLogTable.vue'

const route = useRoute()
const timeRange = ref<[string, string] | null>(null)
const moduleOptions = ref<AuditModuleOption[]>([])
const actionOptions = ref<AuditActionOption[]>([])
const detailRef = ref<InstanceType<typeof AuditLogDetailDrawer>>()

const {
  loading,
  list,
  total,
  query,
  fetchList,
  handleSearch: doSearch,
  handleReset: doReset,
} = usePaginatedList(
  async (q) => {
    const res = await getAuditLogList({
      page: q.page,
      pageSize: q.pageSize,
      employeeId: q.employeeId || undefined,
      module: q.module || undefined,
      action: q.action || undefined,
      keyword: q.keyword || undefined,
      startTime: q.startTime || undefined,
      endTime: q.endTime || undefined,
    })
    moduleOptions.value = res.data.modules ?? []
    actionOptions.value = res.data.actions ?? []
    return res
  },
  {
    page: 1,
    pageSize: 20,
    employeeId: undefined as number | undefined,
    module: '',
    action: '',
    keyword: '',
    startTime: '',
    endTime: '',
  },
  '获取审计日志失败',
)

function syncTimeRange() {
  if (timeRange.value?.length === 2) {
    query.startTime = timeRange.value[0]
    query.endTime = timeRange.value[1]
  } else {
    query.startTime = ''
    query.endTime = ''
  }
}

function handleSearch() {
  syncTimeRange()
  doSearch()
}

function handleReset() {
  timeRange.value = null
  doReset()
}

fetchList()
</script>

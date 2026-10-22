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
        v-model:target-type="query.targetType"
        v-model:target-id="query.targetId"
        v-model:request-path="query.requestPath"
        v-model:success="query.success"
        v-model:keyword="query.keyword"
        v-model:time-range="timeRange"
        :options="options"
        @search="handleSearch"
        @reset="handleReset"
      />

      <AuditLogTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @detail="(row) => detailRef?.open(row)"
        @filter-employee="filterByEmployee"
        @page-change="fetchList"
      />
    </el-card>

    <AuditLogDetailDrawer ref="detailRef" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, toRef } from 'vue'
import { useRoute } from 'vue-router'
import { getAuditLogList, getAuditLogOptions, type AuditOptions } from '@/api/auditLog'
import { useDateRangeModel } from '@/composables/useDateRangeModel'
import { useListQuery } from '@/composables/useListQuery'
import { DEFAULT_PAGE_SIZE } from '@/constants'
import { showApiError } from '@/utils/apiError'
import AuditLogDetailDrawer from './AuditLogDetailDrawer.vue'
import AuditLogFilter from './AuditLogFilter.vue'
import AuditLogTable from './AuditLogTable.vue'

const route = useRoute()
const detailRef = ref<InstanceType<typeof AuditLogDetailDrawer>>()
const options = ref<AuditOptions>({ modules: [], actions: [], targetTypes: [] })

const { loading, error, list, total, query, fetchList, handleSearch, handleReset } = useListQuery({
  fetch: (q) => getAuditLogList(q),
  defaults: {
    page: 1,
    pageSize: DEFAULT_PAGE_SIZE,
    employeeId: undefined as number | undefined,
    module: '',
    action: '',
    targetType: '',
    targetId: undefined as number | undefined,
    requestPath: '',
    success: undefined as boolean | undefined,
    keyword: '',
    startTime: '',
    endTime: '',
  },
  schema: {
    employeeId: 'number',
    module: 'string',
    action: 'string',
    targetType: 'string',
    targetId: 'number',
    requestPath: 'string',
    success: 'boolean',
    keyword: 'string',
    startTime: 'string',
    endTime: 'string',
  },
  errorMessage: '获取操作审计失败',
})

const timeRange = useDateRangeModel(toRef(query, 'startTime'), toRef(query, 'endTime'))

function filterByEmployee(employeeId: number) {
  query.employeeId = employeeId
  void handleSearch()
}

onMounted(async () => {
  try {
    const { data } = await getAuditLogOptions()
    options.value = data
  } catch (e) {
    showApiError(e, '加载筛选项失败')
  }
})
</script>

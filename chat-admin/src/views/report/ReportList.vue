<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
        </div>
      </template>

      <ReportFilter
        v-model:keyword="query.keyword"
        v-model:target-type="query.targetType"
        v-model:reason="query.reason"
        v-model:status="query.status"
        @search="handleSearch"
        @reset="handleReset"
      />
      <ReportTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @review="openReview"
        @page-change="fetchList"
      />
    </el-card>

    <ReportReviewDrawer ref="reviewDrawer" @handled="fetchList" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { useRouteQuery } from '@/composables/useRouteQuery'
import ReportFilter from './ReportFilter.vue'
import ReportTable from './ReportTable.vue'
import ReportReviewDrawer from './ReportReviewDrawer.vue'
import { getReportList, type ReportItem } from '@/api/report'
import { ReportStatus } from '@/constants/report'

const route = useRoute()
const reviewDrawer = ref<InstanceType<typeof ReportReviewDrawer> | null>(null)

const {
  loading,
  list,
  total,
  query,
  fetchList,
  handleSearch,
  handleReset: doReset,
} = usePaginatedList(
  (q) =>
    getReportList({
      page: q.page,
      pageSize: q.pageSize,
      keyword: q.keyword || undefined,
      targetType: q.targetType,
      reason: q.reason,
      status: q.status,
    }),
  {
    page: 1,
    pageSize: 10,
    keyword: '',
    targetType: undefined as number | undefined,
    reason: undefined as number | undefined,
    status: ReportStatus.Pending as number | undefined,
  },
  '获取举报列表失败',
)

const { wrapReset } = useRouteQuery(
  query,
  [
    {
      key: 'status',
      apply: (v, q) => {
        q.status = Number(v)
      },
      clear: (q) => {
        q.status = ReportStatus.Pending
      },
    },
  ],
  {
    onApplied: () => {
      query.page = 1
      fetchList()
    },
  },
)

const handleReset = wrapReset(doReset)

function openReview(row: ReportItem) {
  void reviewDrawer.value?.open(row)
}
</script>

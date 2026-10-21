<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
        </div>
      </template>

      <FeedbackFilter
        v-model:keyword="query.keyword"
        v-model:type="query.type"
        v-model:status="query.status"
        @search="handleSearch"
        @reset="handleReset"
      />
      <FeedbackTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @detail="(row) => detailRef?.open(row)"
        @handle="(row) => handleRef?.open(row)"
        @page-change="fetchList"
      />
    </el-card>

    <FeedbackDetailDialog ref="detailRef" />
    <FeedbackHandleDialog ref="handleRef" @success="fetchList" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRoute } from 'vue-router'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { useRouteQuery } from '@/composables/useRouteQuery'
import { getFeedbackList } from '@/api/feedback'
import FeedbackDetailDialog from './FeedbackDetailDialog.vue'
import FeedbackFilter from './FeedbackFilter.vue'
import FeedbackHandleDialog from './FeedbackHandleDialog.vue'
import FeedbackTable from './FeedbackTable.vue'

const route = useRoute()
const detailRef = ref<InstanceType<typeof FeedbackDetailDialog>>()
const handleRef = ref<InstanceType<typeof FeedbackHandleDialog>>()

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
    getFeedbackList({
      page: q.page,
      pageSize: q.pageSize,
      keyword: q.keyword || undefined,
      type: q.type,
      status: q.status,
    }),
  {
    page: 1,
    pageSize: 10,
    keyword: '',
    type: undefined as number | undefined,
    status: undefined as number | undefined,
  },
  '获取反馈列表失败',
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
        q.status = undefined
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
</script>

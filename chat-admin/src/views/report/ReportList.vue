<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <span v-if="query.status === ReportStatus.Pending" class="sub">
            待处理按提交时间从早到晚排列
          </span>
        </div>
      </template>

      <ReportFilter
        v-model:status="query.status"
        v-model:target-type="query.targetType"
        v-model:target-id="query.targetId"
        v-model:target-user-id="query.targetUserId"
        v-model:reporter-id="query.reporterId"
        v-model:reason="query.reason"
        v-model:keyword="query.keyword"
        v-model:start-time="query.startTime"
        v-model:end-time="query.endTime"
        @search="handleSearch"
        @reset="handleReset"
      />
      <ReportTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @review="(row) => openReview(row.id)"
        @same-target="showSameTarget"
        @page-change="fetchList"
      />
    </el-card>

    <ReportReviewDrawer ref="reviewDrawer" @handled="onHandled" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { getReportList, type ReportItem, type ReportListRequest } from '@/api/report'
import { useListQuery } from '@/composables/useListQuery'
import { FILTER_ALL } from '@/constants'
import { ReportStatus } from '@/constants/report'
import { useUserDrawerStore } from '@/stores/userDrawer'
import ReportFilter from './ReportFilter.vue'
import ReportReviewDrawer from './ReportReviewDrawer.vue'
import ReportTable from './ReportTable.vue'

const route = useRoute()
const router = useRouter()
const drawer = useUserDrawerStore()
const reviewDrawer = ref<InstanceType<typeof ReportReviewDrawer> | null>(null)

type ReportQuery = Omit<ReportListRequest, 'status'> & { status: number }

const toRequest = ({ status, ...q }: ReportQuery): ReportListRequest => ({
  ...q,
  status: status === FILTER_ALL ? undefined : status,
})

const { loading, error, list, total, query, fetchList, handleSearch, handleReset } = useListQuery({
  fetch: (q) => getReportList(toRequest(q)),
  defaults: {
    page: 1,
    pageSize: 20,
    status: ReportStatus.Pending as number,
    targetType: undefined as number | undefined,
    targetId: undefined as number | undefined,
    targetUserId: undefined as number | undefined,
    reporterId: undefined as number | undefined,
    reason: undefined as number | undefined,
    keyword: '',
    startTime: '',
    endTime: '',
  },
  schema: {
    status: 'number',
    targetType: 'number',
    targetId: 'number',
    targetUserId: 'number',
    reporterId: 'number',
    reason: 'number',
    keyword: 'string',
    startTime: 'string',
    endTime: 'string',
  },
  errorMessage: '获取举报列表失败',
})

function openReview(id: number) {
  void reviewDrawer.value?.open(id)
}

function showSameTarget(row: ReportItem) {
  Object.assign(query, {
    status: FILTER_ALL,
    targetType: row.targetType,
    targetId: row.targetId,
    targetUserId: undefined,
    reporterId: undefined,
    reason: undefined,
    keyword: '',
    startTime: '',
    endTime: '',
  })
  void handleSearch()
}

/** 处理完成后刷新列表；需要时按当前筛选条件打开下一条待处理（最早提交的一条）。 */
async function onHandled(openNext: boolean) {
  await fetchList()
  if (!openNext) return
  try {
    const { data } = await getReportList({
      ...toRequest({ ...query }),
      status: ReportStatus.Pending,
      page: 1,
      pageSize: 1,
    })
    const next = data.list[0]
    if (next) openReview(next.id)
  } catch {
    // 打开下一条失败不影响本次处理结果
  }
}

/** ?reviewId=<举报ID> 打开审核抽屉，打开后从地址栏移除。 */
function openReviewFromRoute() {
  const raw = route.query.reviewId
  const id = Number(Array.isArray(raw) ? raw[0] : raw)
  if (!Number.isInteger(id) || id <= 0) return
  const rest = { ...route.query }
  delete rest.reviewId
  void router.replace({ path: route.path, query: rest })
  openReview(id)
}

onMounted(openReviewFromRoute)
watch(() => route.query.reviewId, openReviewFromRoute)
watch(() => drawer.changed, fetchList)
</script>

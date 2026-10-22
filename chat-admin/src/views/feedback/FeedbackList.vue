<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div v-if="selected.length" class="card-header-actions">
            <span class="sub">已选 {{ selected.length }} 条</span>
            <el-button type="primary" plain @click="handleBatchClose">批量关闭</el-button>
          </div>
        </div>
      </template>

      <FeedbackFilter
        v-model:status="query.status"
        v-model:type="query.type"
        v-model:user-id="query.userId"
        v-model:keyword="query.keyword"
        v-model:start-time="query.startTime"
        v-model:end-time="query.endTime"
        @search="handleSearch"
        @reset="handleReset"
      />
      <FeedbackTable
        :list="list"
        :loading="loading"
        :error="error"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @detail="(row) => detailRef?.open(row.id)"
        @selection-change="(rows) => (selected = rows)"
        @page-change="fetchList"
      />
    </el-card>

    <FeedbackDetailDialog ref="detailRef" @success="fetchList" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'

import { batchCloseFeedback, getFeedbackList, type FeedbackItem } from '@/api/feedback'
import { useListQuery } from '@/composables/useListQuery'
import { FILTER_ALL } from '@/constants'
import { FeedbackStatus } from '@/constants/feedback'
import { showApiError } from '@/utils/apiError'
import { confirmWithReason } from '@/utils/promptReason'
import FeedbackDetailDialog from './FeedbackDetailDialog.vue'
import FeedbackFilter from './FeedbackFilter.vue'
import FeedbackTable from './FeedbackTable.vue'

const route = useRoute()
const router = useRouter()
const detailRef = ref<InstanceType<typeof FeedbackDetailDialog>>()
const selected = ref<FeedbackItem[]>([])

const { loading, error, list, total, query, fetchList, handleSearch, handleReset } = useListQuery({
  fetch: ({ status, ...q }) =>
    getFeedbackList({ ...q, status: status === FILTER_ALL ? undefined : status }),
  defaults: {
    page: 1,
    pageSize: 20,
    status: FeedbackStatus.Pending as number,
    type: undefined as number | undefined,
    userId: undefined as number | undefined,
    keyword: '',
    startTime: '',
    endTime: '',
  },
  schema: {
    status: 'number',
    type: 'number',
    userId: 'number',
    keyword: 'string',
    startTime: 'string',
    endTime: 'string',
  },
  errorMessage: '获取反馈列表失败',
})

async function handleBatchClose() {
  const rows = selected.value
  const handleResult = await confirmWithReason(
    `确定关闭选中的 ${rows.length} 条反馈吗？关闭后不可再处理。`,
    { title: '批量关闭', inputPlaceholder: '处理说明（选填，用户可见）' },
  )
  if (handleResult === null) return
  try {
    const { data } = await batchCloseFeedback(
      rows.map((r) => r.id),
      handleResult,
    )
    ElMessage.success(
      data.skipped > 0
        ? `已关闭 ${data.updated} 条，跳过 ${data.skipped} 条`
        : `已关闭 ${data.updated} 条`,
    )
    await fetchList()
  } catch (e) {
    showApiError(e, '批量关闭失败')
  }
}

/** ?detailId=<反馈ID> 打开详情弹窗，打开后从地址栏移除。 */
function openDetailFromRoute() {
  const raw = route.query.detailId
  const id = Number(Array.isArray(raw) ? raw[0] : raw)
  if (!Number.isInteger(id) || id <= 0) return
  const rest = { ...route.query }
  delete rest.detailId
  void router.replace({ path: route.path, query: rest })
  void detailRef.value?.open(id)
}

onMounted(openDetailFromRoute)
watch(() => route.query.detailId, openDetailFromRoute)
</script>

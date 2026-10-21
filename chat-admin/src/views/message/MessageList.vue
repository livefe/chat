<template>
  <div class="admin-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
        </div>
      </template>

      <MessageFilter
        v-model:user-id="query.userId"
        v-model:peer-user-id="query.peerUserId"
        v-model:room-id="query.roomId"
        v-model:post-id="query.postId"
        v-model:scope="query.scope"
        v-model:keyword="query.keyword"
        v-model:date-range="dateRange"
        @search="handleSearch"
        @reset="handleReset"
      />

      <MessageTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @context="(row) => contextRef?.open(row)"
        @delete="handleDelete"
        @page-change="fetchList"
      />
    </el-card>

    <MessageContextDialog ref="contextRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { deleteMessage, getMessageList, type AdminMessageItem } from '@/api/message'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { useRouteQuery } from '@/composables/useRouteQuery'
import { MessageScope } from '@/constants/message'
import { confirmAndRun } from '@/utils/confirm'
import MessageContextDialog from './MessageContextDialog.vue'
import MessageFilter from './MessageFilter.vue'
import MessageTable from './MessageTable.vue'

const route = useRoute()
const dateRange = ref<[string, string] | null>(null)
const contextRef = ref<InstanceType<typeof MessageContextDialog>>()

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
    getMessageList({
      page: q.page,
      pageSize: q.pageSize,
      userId: q.userId || undefined,
      peerUserId: q.peerUserId || undefined,
      roomId: q.roomId || undefined,
      postId: q.postId || undefined,
      scope: q.scope,
      keyword: q.keyword || undefined,
      startTime: q.startTime || undefined,
      endTime: q.endTime || undefined,
    }),
  {
    page: 1,
    pageSize: 20,
    userId: undefined as number | undefined,
    peerUserId: undefined as number | undefined,
    roomId: undefined as number | undefined,
    postId: undefined as number | undefined,
    scope: undefined as number | undefined,
    keyword: '',
    startTime: '',
    endTime: '',
  },
  '获取消息列表失败',
)

type IdKey = 'userId' | 'peerUserId' | 'roomId' | 'postId'

function idField(key: IdKey) {
  return {
    key,
    apply: (v: string, q: typeof query) => {
      const id = Number(v)
      if (!Number.isNaN(id) && id > 0) q[key] = id
    },
    clear: (q: typeof query) => {
      q[key] = undefined
    },
  }
}

const { wrapReset } = useRouteQuery(
  query,
  [
    idField('userId'),
    idField('peerUserId'),
    idField('roomId'),
    idField('postId'),
    {
      key: 'scope',
      apply: (v, q) => {
        const s = Number(v)
        if (!Number.isNaN(s)) q.scope = s
      },
      clear: (q) => {
        q.scope = undefined
      },
    },
    {
      key: 'keyword',
      apply: (v, q) => {
        q.keyword = v
      },
      clear: (q) => {
        q.keyword = ''
      },
    },
  ],
  {
    onApplied: () => {
      if (query.roomId && query.scope === undefined) {
        query.scope = MessageScope.Room
      }
      if (query.postId && query.scope === undefined) {
        query.scope = MessageScope.Post
      }
      query.page = 1
      fetchList()
    },
  },
)

watch(dateRange, (val) => {
  query.startTime = val?.[0] || ''
  query.endTime = val?.[1] || ''
})

const handleReset = wrapReset(() => {
  dateRange.value = null
  doReset()
})

async function handleDelete(row: AdminMessageItem) {
  await confirmAndRun(`确定删除消息 #${row.id} 吗？`, async () => {
    await deleteMessage(row.id)
    await fetchList()
  })
}
</script>

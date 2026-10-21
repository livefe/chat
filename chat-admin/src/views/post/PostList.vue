<template>
  <div class="admin-page post-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ route.meta.title }}</span>
          <div class="card-header-actions">
            <el-button @click="mapVisible = true">全景地图</el-button>
          </div>
        </div>
      </template>

      <PostFilter
        v-model:keyword="query.keyword"
        v-model:publisher-id="query.publisherId"
        v-model:room-id="query.roomId"
        v-model:category="query.category"
        v-model:status="query.status"
        @search="handleSearch"
        @reset="handleReset"
      />

      <PostTable
        :list="list"
        :loading="loading"
        :total="total"
        v-model:page="query.page"
        v-model:page-size="query.pageSize"
        @detail="openDetail"
        @signup="openSignupList"
        @messages="goPostMessages"
        @go-room="goRoom"
        @status="confirmStatus"
        @page-change="fetchList"
      />
    </el-card>

    <PostDetailDrawer
      v-model="detailVisible"
      :current="current"
      @signup="openSignupList"
      @messages="goPostMessages"
      @go-room="goRoom"
      @status="confirmStatus"
    />

    <PostSignupDialog
      v-model="signupVisible"
      :title="signupTitle"
      :list="signupList"
      :loading="signupLoading"
      :total="signupTotal"
      v-model:page="signupQuery.page"
      v-model:page-size="signupQuery.pageSize"
      v-model:status="signupQuery.status"
      v-model:check-in="signupQuery.checkIn"
      :need-check-in="signupNeedCheckIn"
      :counts="signupCounts"
      @fetch="fetchSignupList"
    />

    <PostMapDialog
      v-model:visible="mapVisible"
      @filter-room="filterByRoom"
      @open-post="(id) => openDetail({ id })"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { showApiError } from '@/utils/apiError'

import {
  getPostList,
  getPostDetail,
  getPostSignupList,
  updatePostStatus,
  type PostItem,
  type PostDetail,
  type PostSignupListCounts,
} from '@/api/post'
import { usePaginatedList } from '@/composables/usePaginatedList'
import { useRouteQuery } from '@/composables/useRouteQuery'
import { useEmployeeStore } from '@/stores/employee'
import { confirmAndRun } from '@/utils/confirm'
import { pushMenuRoute } from '@/utils/navigation'
import { emptySignupListCounts } from '@/utils/postDisplay'
import PostFilter from './PostFilter.vue'
import PostTable from './PostTable.vue'
import PostDetailDrawer from './PostDetailDrawer.vue'
import PostSignupDialog from './PostSignupDialog.vue'
import PostMapDialog from './PostMapDialog.vue'
import { PostStatus } from '@/constants/post'
import { MessageScope } from '@/constants/message'
import { PERM } from '@/constants/permission'
import { usePermission } from '@/composables/usePermission'

const router = useRouter()
const route = useRoute()
const employeeStore = useEmployeeStore()
const { can } = usePermission()
const mapVisible = ref(false)

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
    getPostList({
      page: q.page,
      pageSize: q.pageSize,
      status: q.status,
      publisherId: q.publisherId,
      roomId: q.roomId,
      category: q.category || undefined,
      keyword: q.keyword || undefined,
    }),
  {
    page: 1,
    pageSize: 20,
    keyword: '',
    publisherId: undefined as number | undefined,
    roomId: undefined as number | undefined,
    category: undefined as string | undefined,
    status: undefined as number | undefined,
  },
  '获取帖子列表失败',
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
    {
      key: 'publisherId',
      apply: (v, q) => {
        const id = Number(v)
        if (!Number.isNaN(id) && id > 0) q.publisherId = id
      },
      clear: (q) => {
        q.publisherId = undefined
      },
    },
    {
      key: 'roomId',
      apply: (v, q) => {
        const id = Number(v)
        if (!Number.isNaN(id) && id >= 0) q.roomId = id
      },
      clear: (q) => {
        q.roomId = undefined
      },
    },
    {
      key: 'category',
      apply: (v, q) => {
        q.category = v
      },
      clear: (q) => {
        q.category = undefined
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
      query.page = 1
      fetchList()
    },
  },
)

const handleReset = wrapReset(doReset)

const detailVisible = ref(false)
const current = ref<PostDetail | null>(null)

const signupVisible = ref(false)
const signupPost = ref<Pick<PostItem, 'id' | 'title'> | null>(null)
const signupNeedCheckIn = ref(false)
const signupCounts = ref<PostSignupListCounts>(emptySignupListCounts())

const {
  loading: signupLoading,
  list: signupList,
  total: signupTotal,
  query: signupQuery,
  fetchList: fetchSignupList,
} = usePaginatedList(
  (q) =>
    getPostSignupList({
      postId: signupPost.value?.id ?? 0,
      page: q.page,
      pageSize: q.pageSize,
      status: q.status,
      checkIn: q.checkIn,
    }).then((res) => {
      signupNeedCheckIn.value = !!res.data.needCheckIn
      signupCounts.value = res.data.counts ?? emptySignupListCounts()
      return res
    }),
  {
    page: 1,
    pageSize: 20,
    status: undefined as number | undefined,
    checkIn: undefined as number | undefined,
  },
  '获取报名名单失败',
)

const signupTitle = computed(() =>
  signupPost.value ? `报名名单 - ${signupPost.value.title}` : '报名名单',
)

async function openSignupList(row: Pick<PostItem, 'id' | 'title'>) {
  signupPost.value = row
  signupQuery.page = 1
  signupQuery.status = undefined
  signupQuery.checkIn = undefined
  signupCounts.value = emptySignupListCounts()
  signupVisible.value = true
  await fetchSignupList()
}

function filterByRoom(roomId: number) {
  query.roomId = roomId
  handleSearch()
}

async function openDetail(row: Pick<PostItem, 'id'> & Partial<PostItem>) {
  const preview = row.title !== undefined
  if (preview) {
    current.value = row as PostDetail
    detailVisible.value = true
  }
  try {
    const { data } = await getPostDetail(row.id)
    current.value = data
    detailVisible.value = true
  } catch (e) {
    if (!preview) showApiError(e, '获取详情失败')
  }
}

async function confirmStatus(row: PostItem, status: PostStatus, action: string) {
  await confirmAndRun(`确定要${action}该帖子吗？`, async () => {
    await updatePostStatus(row.id, status)
    if (current.value?.id === row.id) {
      const { data } = await getPostDetail(row.id)
      current.value = data
    }
    fetchList()
  })
}

function goPostMessages(postId?: number) {
  if (!postId) return
  if (!can(PERM.ChatMessage)) {
    ElMessage.warning('暂无消息菜单权限')
    return
  }
  pushMenuRoute(router, employeeStore.permissions, PERM.ChatMessage, {
    postId: String(postId),
    scope: String(MessageScope.Post),
  })
}

function goRoom(roomId: number) {
  if (!roomId) return
  pushMenuRoute(router, employeeStore.permissions, PERM.ChatRoom, {
    roomId: String(roomId),
  })
}
</script>

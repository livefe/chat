<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showConfirmDialog, showSuccessToast } from 'vant'

import PostListItem from '@/components/post/PostListItem.vue'
import PostJoinedActions from './components/PostJoinedActions.vue'
import PostListActionBar, { type PostListAction } from './components/PostListActionBar.vue'
import SubPageLayout from '@/components/SubPageLayout.vue'
import {
  postDelete,
  postJoinedRemove,
  postMyJoinedList,
  postMyList,
  postWithdraw,
  type PostItem,
} from '@/api/post'
import { PostStatus } from '@/constants/post'
import { useLocationStore } from '@/stores/location'
import {
  canDeleteJoinedPost,
  canDeletePublishedPost,
  canShowJoinedCheckInCode,
  canShowJoinedCheckInCodeIcon,
  hasCheckInTime,
} from '@/utils/post/post'
import { showCatchError } from '@/utils/toast'

const PAGE_SIZE = 30

const route = useRoute()
const router = useRouter()
const gps = useLocationStore().gps
const ready = ref(false)
const loading = ref(false)
const refreshing = ref(false)
const finished = ref(false)
const page = ref(1)
const list = ref<PostItem[]>([])
const actingId = ref(0)
let fetching = false

const mode = computed<'published' | 'joined'>(() =>
  route.query.mode === 'joined' ? 'joined' : 'published',
)
const title = computed(() => (mode.value === 'joined' ? '我参与的' : '我发布的'))
const emptyText = computed(() => (mode.value === 'joined' ? '暂无参与' : '暂无发布'))

function canEditPost(post: PostItem) {
  return post.status === PostStatus.Pending || post.status === PostStatus.Rejected
}

function canWithdrawPost(post: PostItem) {
  return post.status !== PostStatus.Disabled
}

function listParams() {
  return {
    page: page.value,
    pageSize: PAGE_SIZE,
    latitude: gps.valid ? gps.lat : undefined,
    longitude: gps.valid ? gps.lng : undefined,
  }
}

async function fetchPage() {
  const params = listParams()
  return mode.value === 'joined' ? postMyJoinedList(params) : postMyList(params)
}

async function loadMore() {
  if (finished.value || fetching) return
  fetching = true
  loading.value = true
  try {
    const res = await fetchPage()
    const items = res.data.list || []
    list.value.push(...items)
    if (list.value.length >= res.data.total || items.length < PAGE_SIZE) {
      finished.value = true
    } else {
      page.value += 1
    }
  } catch (e: unknown) {
    showCatchError(e, '加载失败', true)
    finished.value = true
  } finally {
    fetching = false
    loading.value = false
  }
}

async function onRefresh() {
  fetching = true
  page.value = 1
  finished.value = false
  await loadMoreReplace()
  refreshing.value = false
}

async function loadMoreReplace() {
  loading.value = true
  try {
    const res = await fetchPage()
    const items = res.data.list || []
    list.value = items
    if (list.value.length >= res.data.total || items.length < PAGE_SIZE) {
      finished.value = true
    } else {
      page.value += 1
    }
  } catch (e: unknown) {
    showCatchError(e, '加载失败', true)
    finished.value = true
  } finally {
    fetching = false
    loading.value = false
  }
}

async function reloadList() {
  page.value = 1
  finished.value = false
  list.value = []
  await loadMore()
}

function openDetail(post: PostItem) {
  router.push(`/post/${post.id}`)
}

function onEdit(post: PostItem) {
  router.push(`/post/${post.id}/edit`)
}

function needSignupList(post: PostItem) {
  return !!post.extras?.needSignup
}

function openSignupList(post: PostItem) {
  router.push(`/post/${post.id}/signups`)
}

function openScan() {
  router.push('/scan')
}

function openCheckInCode(post: PostItem) {
  router.push(`/post/${post.id}/check-in-code`)
}

function publishedActions(post: PostItem): PostListAction[] {
  const actions: PostListAction[] = []
  if (needSignupList(post)) {
    actions.push({ label: '报名名单', onClick: () => openSignupList(post) })
  }
  if (canEditPost(post)) {
    actions.push({ label: '编辑', onClick: () => onEdit(post) })
  }
  if (canWithdrawPost(post)) {
    actions.push({
      label: actingId.value === post.id ? '下架中…' : '下架',
      danger: true,
      disabled: actingId.value === post.id,
      onClick: () => onWithdraw(post),
    })
  }
  if (canDeletePublishedPost(post)) {
    actions.push({
      label: actingId.value === post.id ? '删除中…' : '删除',
      danger: true,
      disabled: actingId.value === post.id,
      onClick: () => onDeletePublished(post),
    })
  }
  return actions
}

function joinedActions(post: PostItem): PostListAction[] {
  const actions: PostListAction[] = []
  if (canShowJoinedCheckInCode(post) && hasCheckInTime(post.myCheckedInAt)) {
    actions.push({ label: '查看报名码', onClick: () => openCheckInCode(post) })
  }
  if (canDeleteJoinedPost(post)) {
    actions.push({
      label: actingId.value === post.id ? '移出中…' : '移出列表',
      danger: true,
      disabled: actingId.value === post.id,
      onClick: () => onDeleteJoined(post),
    })
  }
  return actions
}

function hasJoinedActions(post: PostItem) {
  return canShowJoinedCheckInCodeIcon(post) || joinedActions(post).length > 0
}

async function onWithdraw(post: PostItem) {
  if (actingId.value) return
  try {
    await showConfirmDialog({ title: '下架', message: '确定要下架吗？' })
  } catch {
    return
  }
  actingId.value = post.id
  try {
    await postWithdraw(post.id)
    showSuccessToast('已下架')
    await reloadList()
  } catch (e: unknown) {
    showCatchError(e, '下架失败', true)
  } finally {
    actingId.value = 0
  }
}

async function onDeletePublished(post: PostItem) {
  if (actingId.value) return
  try {
    await showConfirmDialog({
      title: '删除',
      message: '删除后不再出现在「我发布的」列表中，确定删除吗？',
    })
  } catch {
    return
  }
  actingId.value = post.id
  try {
    await postDelete(post.id)
    showSuccessToast('已删除')
    await reloadList()
  } catch (e: unknown) {
    showCatchError(e, '删除失败', true)
  } finally {
    actingId.value = 0
  }
}

async function onDeleteJoined(post: PostItem) {
  if (actingId.value) return
  try {
    await showConfirmDialog({
      title: '移出列表',
      message: '移出后不再出现在「我参与的」列表中，确定移出吗？',
    })
  } catch {
    return
  }
  actingId.value = post.id
  try {
    await postJoinedRemove(post.id)
    showSuccessToast('已移出')
    await reloadList()
  } catch (e: unknown) {
    showCatchError(e, '移出失败', true)
  } finally {
    actingId.value = 0
  }
}

onMounted(async () => {
  await gps.ensureFreshLocation()
  ready.value = true
})
</script>

<template>
  <SubPageLayout :title="title" full-height>
    <template v-if="mode === 'published'" #right>
      <button type="button" class="my-posts-scan" aria-label="扫一扫" @click="openScan">
        <svg class="my-posts-scan__icon" viewBox="0 0 24 24" aria-hidden="true">
          <path
            fill="none"
            stroke="currentColor"
            stroke-width="1.75"
            stroke-linecap="round"
            d="M5 9V5h4M15 5h4v4M19 15v4h-4M9 19H5v-4"
          />
          <path
            fill="none"
            stroke="currentColor"
            stroke-width="1.5"
            stroke-linecap="round"
            d="M7 12h10"
          />
        </svg>
      </button>
    </template>
    <van-pull-refresh v-model="refreshing" class="sub-page__scroll" @refresh="onRefresh">
      <div v-if="!ready" class="tab-page__status">
        <van-loading size="24px" vertical>加载中…</van-loading>
      </div>
      <van-list
        v-else
        :key="mode"
        :loading="loading && !refreshing"
        :finished="finished"
        :finished-text="list.length ? '没有更多了' : ''"
        :immediate-check="true"
        @update:loading="
          (v) => {
            if (!refreshing) loading = v
          }
        "
        @load="loadMore"
      >
        <van-empty v-if="finished && !list.length" :description="emptyText" />
        <div v-else class="feed-list">
          <div v-for="item in list" :key="item.id" class="feed-item">
            <div class="feed-item-wrap" @click="openDetail(item)">
              <PostListItem
                :post="item"
                variant="list"
                :badge-mode="mode === 'joined' ? 'joined' : 'owner'"
              >
                <template v-if="mode === 'published' && publishedActions(item).length" #actions>
                  <PostListActionBar :actions="publishedActions(item)" />
                </template>
                <template v-else-if="mode === 'joined' && hasJoinedActions(item)" #actions>
                  <PostJoinedActions
                    :show-check-in-code="canShowJoinedCheckInCodeIcon(item)"
                    :actions="joinedActions(item)"
                    @check-in-code="openCheckInCode(item)"
                  />
                </template>
              </PostListItem>
            </div>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>
  </SubPageLayout>
</template>

<style scoped>
.my-posts-scan {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border: none;
  background: transparent;
  padding: 0;
  color: var(--text-primary);
  cursor: pointer;
}

.my-posts-scan__icon {
  display: block;
  width: var(--icon-nav);
  height: var(--icon-nav);
}

.feed-item :deep(.post-item) {
  border-bottom: none;
}
</style>

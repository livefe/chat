<script setup lang="ts">
import { computed, onActivated, onMounted, ref } from 'vue'
import { showCatchError } from '@/utils/toast'
import { useRouter } from 'vue-router'

import PostListItem from '@/components/post/PostListItem.vue'
import SponsoredFeedCard from './components/SponsoredFeedCard.vue'
import {
  feedItemKey,
  normalizeFeedItem,
  postList,
  type FeedItem,
  type FeedItemDto,
} from '@/api/post'
import { useLocationStore } from '@/stores/location'

const PAGE_SIZE = 30

const router = useRouter()
const gps = useLocationStore().gps
const ready = ref(false)
const loading = ref(false)
const refreshing = ref(false)
const finished = ref(false)
const page = ref(1)
const feed = ref<FeedItem[]>([])
let listEpoch = 0

const leftItems = computed(() => feed.value.filter((_, i) => i % 2 === 0))
const rightItems = computed(() => feed.value.filter((_, i) => i % 2 === 1))
const postCount = computed(() => feed.value.filter((i) => i.type === 'post').length)

function listParams() {
  return {
    page: page.value,
    pageSize: PAGE_SIZE,
    latitude: gps.valid ? gps.lat : undefined,
    longitude: gps.valid ? gps.lng : undefined,
  }
}

function parseFeedList(list: FeedItemDto[] | undefined): FeedItem[] {
  const out: FeedItem[] = []
  for (const raw of list || []) {
    const item = normalizeFeedItem(raw)
    if (item) out.push(item)
  }
  return out
}

function mergeFeed(existing: FeedItem[], incoming: FeedItem[]): FeedItem[] {
  const seen = new Set(existing.map(feedItemKey))
  const next = [...existing]
  for (const item of incoming) {
    const key = feedItemKey(item)
    if (seen.has(key)) continue
    seen.add(key)
    next.push(item)
  }
  return next
}

async function loadMore() {
  if (finished.value || refreshing.value) return
  const epoch = listEpoch
  loading.value = true
  try {
    const res = await postList(listParams())
    if (epoch !== listEpoch) return
    const items = parseFeedList(res.data.list)
    const pagePostCount = items.filter((i) => i.type === 'post').length
    feed.value = mergeFeed(feed.value, items)
    if (postCount.value >= res.data.total || pagePostCount < PAGE_SIZE) {
      finished.value = true
    } else {
      page.value += 1
    }
  } catch (e: unknown) {
    if (epoch !== listEpoch) return
    showCatchError(e, '加载失败', true)
    finished.value = true
  } finally {
    if (epoch === listEpoch) loading.value = false
  }
}

async function onRefresh() {
  listEpoch += 1
  const epoch = listEpoch
  page.value = 1
  finished.value = false
  try {
    loading.value = true
    const res = await postList(listParams())
    if (epoch !== listEpoch) return
    const items = parseFeedList(res.data.list)
    const pagePostCount = items.filter((i) => i.type === 'post').length
    feed.value = items
    if (pagePostCount >= res.data.total || pagePostCount < PAGE_SIZE) {
      finished.value = true
    } else {
      page.value += 1
    }
  } catch (e: unknown) {
    if (epoch !== listEpoch) return
    showCatchError(e, '加载失败', true)
    finished.value = true
  } finally {
    if (epoch === listEpoch) {
      loading.value = false
      refreshing.value = false
    }
  }
}

onMounted(async () => {
  await gps.ensureFreshLocation()
  ready.value = true
})

onActivated(() => {
  if (!ready.value || loading.value || refreshing.value) return
  void onRefresh()
})

function openPublish() {
  router.push('/post/publish')
}

function openPost(id: number) {
  router.push(`/post/${id}`)
}
</script>

<template>
  <div class="tab-page">
    <div class="tab-page__toolbar">
      <div class="tab-page__title">发现</div>
      <button type="button" class="tab-page__header-btn" aria-label="发布" @click="openPublish">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
          <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z" />
        </svg>
      </button>
    </div>

    <van-pull-refresh v-model="refreshing" class="tab-page__scroll" @refresh="onRefresh">
      <div v-if="!ready" class="tab-page__status">
        <van-loading size="24px" vertical>加载中…</van-loading>
      </div>
      <van-list
        v-else
        :loading="loading && !refreshing"
        :finished="finished || refreshing"
        finished-text=""
        :immediate-check="true"
        @update:loading="
          (v) => {
            if (!refreshing) loading = v
          }
        "
        @load="loadMore"
      >
        <van-empty v-if="finished && !feed.length" description="暂无内容" />
        <div v-else class="waterfall">
          <div class="waterfall__col">
            <div v-for="item in leftItems" :key="feedItemKey(item)" class="waterfall__item">
              <div v-if="item.type === 'post'" @click="openPost(item.post.id)">
                <PostListItem :post="item.post" variant="waterfall" />
              </div>
              <SponsoredFeedCard v-else :item="item.sponsored" />
            </div>
          </div>
          <div class="waterfall__col">
            <div v-for="item in rightItems" :key="feedItemKey(item)" class="waterfall__item">
              <div v-if="item.type === 'post'" @click="openPost(item.post.id)">
                <PostListItem :post="item.post" variant="waterfall" />
              </div>
              <SponsoredFeedCard v-else :item="item.sponsored" />
            </div>
          </div>
        </div>
      </van-list>
    </van-pull-refresh>
  </div>
</template>

<style scoped>
.waterfall {
  display: flex;
  align-items: flex-start;
  gap: 6px;
  padding: 8px 12px 12px;
}

.waterfall__col {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.waterfall__item {
  width: 100%;
}

.tab-page__scroll :deep(.van-list__loading) {
  padding: 16px 0 28px;
}
</style>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showFailToast } from 'vant'

import PostListItem from '@/components/post/PostListItem.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { getUserProfile, type PublicUserProfile } from '@/api/user'
import { useLocationStore } from '@/stores/location'
import { formatDateChinese } from '@/utils/datetime'
import NavMoreButton from '@/components/NavMoreButton.vue'
import type { NavMoreAction } from '@/types/nav'
import {
  ACTION_REPORT_USER,
  openUserReportSheet,
  reportUserAction,
} from '@/composables/useReportAction'
import {
  ACTION_BLOCK_USER,
  ACTION_UNBLOCK_USER,
  blockUserAction,
  confirmBlockUser,
  confirmUnblockUser,
  unblockUserAction,
} from '@/composables/useBlockAction'
import { goBackOr } from '@/utils/nav'
import { showCatchError } from '@/utils/toast'
import ShareOpenAppBar from '@/components/ShareOpenAppBar.vue'
import { useShareLanding } from '@/composables/useShareLanding'
import { useUserPostList } from '@/composables/useUserPostList'

const route = useRoute()
const router = useRouter()
const gps = useLocationStore().gps

const profile = ref<PublicUserProfile | null>(null)
const loadingProfile = ref(false)
const ready = ref(false)

interface ProfilePreview {
  name: string
  avatarUrl: string
}

function queryName() {
  const name = route.query.name
  return typeof name === 'string' ? name.trim() : ''
}

function readRoutePreview(): ProfilePreview {
  const state = history.state as { profilePreview?: ProfilePreview } | null
  const fromState = state?.profilePreview
  return {
    name: (fromState?.name || queryName() || '').trim(),
    avatarUrl: (fromState?.avatarUrl || '').trim(),
  }
}

const preview = ref<ProfilePreview>(readRoutePreview())
const hasPreview = computed(() => !!preview.value.name)

const userId = computed(() => Number(route.params.id || 0))
const {
  posts,
  listLoading,
  refreshing,
  finished,
  reset,
  setListLoading,
  loadMorePosts,
  onRefresh,
} = useUserPostList(userId)
const { showOpenApp, downloadHref, openApp } = useShareLanding(() => ({
  type: 'user',
  id: userId.value,
}))
const joinLabel = computed(() =>
  profile.value?.createdAt ? formatDateChinese(profile.value.createdAt) : '',
)
const canMessage = computed(
  () =>
    !!profile.value && !profile.value.deactivated && !!profile.value.canMessage && userId.value > 0,
)
const canReport = computed(
  () => !!profile.value && !profile.value.isSelf && !profile.value.deactivated && userId.value > 0,
)
const showPosts = computed(
  () =>
    !!profile.value &&
    !profile.value.deactivated &&
    (!!profile.value.isSelf || !!profile.value.canMessage),
)
const headerName = computed(() => profile.value?.nickname || preview.value.name)
const headerAvatar = computed(() => profile.value?.avatarUrl || preview.value.avatarUrl)
const showHeader = computed(() => !!profile.value || hasPreview.value)
const showProfileSkeleton = computed(() => loadingProfile.value && !hasPreview.value)
const moreActions = computed(() => {
  if (!canReport.value) return []
  if (profile.value?.blockedByMe) {
    return [unblockUserAction, reportUserAction]
  }
  return [blockUserAction, reportUserAction]
})

async function loadProfile() {
  if (!userId.value) {
    showFailToast('用户不存在')
    goBackOr(router)
    return
  }
  loadingProfile.value = true
  try {
    const res = await getUserProfile(userId.value)
    profile.value = res.data
  } catch (e: unknown) {
    showCatchError(e, '加载失败', true)
    goBackOr(router)
  } finally {
    loadingProfile.value = false
  }
}

function onBack() {
  const from = route.query.from
  if (from === 'chat') {
    router.replace('/chat')
    return
  }
  if (from === 'messages') {
    router.replace({ name: 'ConversationList' })
    return
  }
  if (from === 'post') {
    const postId = route.query.postId
    if (postId) {
      router.replace(`/post/${postId}`)
      return
    }
  }
  if (from === 'private') {
    const peerRaw = route.query.peerId
    const peerId =
      typeof peerRaw === 'string' && Number(peerRaw) > 0 ? Number(peerRaw) : userId.value
    router.replace({
      name: 'PrivateChat',
      params: { peerId: String(peerId) },
      query: {
        from: typeof route.query.chatFrom === 'string' ? route.query.chatFrom : 'messages',
        name: typeof route.query.peerName === 'string' ? route.query.peerName : '',
        postId: route.query.postId,
      },
    })
    return
  }
  goBackOr(router)
}

function openPrivateChat() {
  if (!canMessage.value) return
  router.push({
    name: 'PrivateChat',
    params: { peerId: String(userId.value) },
    query: {
      from: 'profile',
      name: profile.value?.nickname || '',
      profileFrom: typeof route.query.from === 'string' ? route.query.from : '',
      postId: route.query.postId,
    },
  })
}

function onMoreSelect(action: NavMoreAction) {
  if (!profile.value) return
  if (action.name === ACTION_REPORT_USER) {
    openUserReportSheet(router, userId.value, profile.value.nickname, profile.value.avatarUrl)
    return
  }
  if (action.name === ACTION_BLOCK_USER) {
    void confirmBlockUser(userId.value, profile.value.nickname).then((ok) => {
      if (ok) void loadProfile()
    })
    return
  }
  if (action.name === ACTION_UNBLOCK_USER) {
    void confirmUnblockUser(userId.value, profile.value.nickname).then((ok) => {
      if (ok) void loadProfile()
    })
  }
}

async function bootstrap() {
  reset()
  preview.value = readRoutePreview()
  profile.value = null
  ready.value = false
  await gps.ensureFreshLocation()
  await loadProfile()
  ready.value = true
}

watch(
  userId,
  () => {
    void bootstrap()
  },
  { immediate: true },
)
</script>

<template>
  <div class="user-profile">
    <header class="profile-hero">
      <div class="profile-hero__toolbar">
        <button type="button" class="profile-hero__icon-btn" aria-label="返回" @click="onBack">
          <van-icon name="arrow-left" :size="22" />
        </button>
        <NavMoreButton
          v-if="canReport"
          class="profile-hero__more"
          :actions="moreActions"
          @select="onMoreSelect"
        />
      </div>

      <van-skeleton
        title
        avatar
        :row="3"
        :loading="showProfileSkeleton"
        class="profile-hero__skeleton"
      >
        <div v-if="showHeader" class="profile-hero__body">
          <UserAvatar :src="headerAvatar" :size="72" tone="profile" />
          <div class="profile-meta">
            <div class="profile-name">{{ headerName }}</div>
            <div v-if="joinLabel" class="profile-join">加入于 {{ joinLabel }}</div>
            <div v-if="profile?.deactivated" class="profile-hint">该用户已注销</div>
            <div v-else-if="profile?.blockedByMe" class="profile-hint">已拉黑该用户</div>
          </div>
          <van-button
            v-if="canMessage"
            class="message-btn"
            type="primary"
            size="small"
            @click="openPrivateChat"
          >
            发私信
          </van-button>
        </div>
      </van-skeleton>
    </header>

    <section v-if="showPosts" class="posts-section">
      <div class="section-title">发布的内容</div>
      <van-pull-refresh v-model="refreshing" class="posts-scroll" @refresh="onRefresh">
        <div v-if="!ready" class="tab-page__status">
          <van-loading size="24px" vertical>加载中…</van-loading>
        </div>
        <van-list
          v-else
          :loading="listLoading && !refreshing"
          :finished="finished || refreshing"
          :finished-text="posts.length ? '没有更多了' : ''"
          :immediate-check="true"
          @update:loading="setListLoading"
          @load="loadMorePosts"
        >
          <van-empty v-if="finished && !posts.length" description="暂无发布" />
          <div v-else class="feed-list">
            <div
              v-for="item in posts"
              :key="item.id"
              class="feed-item-wrap"
              @click="router.push(`/post/${item.id}`)"
            >
              <PostListItem :post="item" />
            </div>
          </div>
        </van-list>
      </van-pull-refresh>
    </section>
    <ShareOpenAppBar
      v-if="showOpenApp && userId > 0"
      :download-href="downloadHref"
      @open-app="openApp"
    />
  </div>
</template>

<style scoped>
.user-profile {
  height: var(--app-vh, 100dvh);
  display: flex;
  flex-direction: column;
  background: var(--page-bg);
  overflow: hidden;
}

.profile-hero {
  flex-shrink: 0;
  background: var(--card-bg);
  border-bottom: 1px solid var(--border-color);
}

.profile-hero__toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: calc(6px + env(safe-area-inset-top, 0px)) 8px 4px;
  color: var(--text-primary);
}

.profile-hero__icon-btn,
.profile-hero__more {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border: none;
  background: transparent;
  padding: 0;
  margin: 0;
  color: inherit;
  cursor: pointer;
}

.profile-hero__skeleton {
  padding: 0 16px 20px;
}

.profile-hero__body {
  padding: 4px 16px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.profile-meta {
  text-align: center;
}

.profile-name {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}

.profile-join {
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-tertiary);
}

.profile-hint {
  margin-top: 6px;
  font-size: 12px;
  color: var(--text-danger);
}

.message-btn {
  min-width: 96px;
}

.posts-section {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.section-title {
  flex-shrink: 0;
  padding: 12px 16px 8px;
  font-size: 14px;
  font-weight: 500;
  color: var(--text-secondary);
}

.posts-scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
}
</style>

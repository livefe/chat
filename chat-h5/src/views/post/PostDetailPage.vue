<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showSuccessToast } from 'vant'

import PostChatSheet from './components/PostChatSheet.vue'
import PostImageGrid from './components/PostImageGrid.vue'
import SubPageLayout from '@/components/SubPageLayout.vue'
import UserAvatar from '@/components/UserAvatar.vue'
import { usePostChat } from '@/composables/usePostChat'
import {
  postSignupCancel,
  postSignup,
  postDetail,
  type PostDetail,
  type PostDiscussPreview,
} from '@/api/post'
import { messagePreview } from '@/utils/chat/chatMessageContent'
import { useUserStore } from '@/stores/user'
import { SignupStatus } from '@/constants/post'
import { canDiscussPost, formatPostCreatedAt, signupStatusText } from '@/utils/post/post'
import NavMoreButton from '@/components/NavMoreButton.vue'
import type { NavMoreAction } from '@/types/nav'
import {
  ACTION_REPORT_CONTENT,
  openPostReportSheet,
  reportContentAction,
} from '@/composables/useReportAction'
import { openUserProfile } from '@/utils/user'
import ShareOpenAppBar from '@/components/ShareOpenAppBar.vue'
import { useShareLanding } from '@/composables/useShareLanding'
import { goBackOr } from '@/utils/nav'
import { buildShareUrl } from '@/utils/share/openApp'
import { openForwardPost } from '@/utils/chat/chatForwardNav'
import { showCatchError } from '@/utils/toast'
import PostDetailExtras from './components/PostDetailExtras.vue'
import PostDetailFooter from './components/PostDetailFooter.vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const detail = ref<PostDetail | null>(null)
const loading = ref(false)
const acting = ref(false)

const postId = computed(() => Number(route.params.id || 0))
const { showOpenApp, downloadHref, openApp } = useShareLanding(() => ({
  type: 'post',
  id: postId.value,
}))
const isOwner = computed(() => !!detail.value && detail.value.publisherId === userStore.info?.id)
const moreActions = [reportContentAction]

const {
  visible: chatVisible,
  opening: chatOpening,
  messages: chatMessages,
  historyLoading: chatHistoryLoading,
  historyHasMore: chatHistoryHasMore,
  previewText: chatPreviewText,
  historyLoaded: chatHistoryLoaded,
  onlineLabel: chatOnlineLabel,
  open: openChat,
  send: sendChat,
  loadMore: loadMoreChat,
  recall: recallChat,
} = usePostChat(postId)

const isWaitlisted = computed(() => detail.value?.mySignupStatus === SignupStatus.Waitlisted)
const hasJoined = computed(() => {
  const status = detail.value?.mySignupStatus
  return status === SignupStatus.Confirmed || status === SignupStatus.Waitlisted
})

function openPublisherProfile() {
  const publisher = detail.value?.publisher
  if (!publisher?.id) return
  openUserProfile(router, publisher.id, {
    from: 'post',
    postId: postId.value,
    name: publisher.nickname,
    avatarUrl: publisher.avatarUrl || '',
  })
}

function onBack() {
  if (route.query.from === 'share') {
    router.replace({ name: 'PostFeed' })
    return
  }
  goBackOr(router)
}

const canDiscuss = computed(() => !!detail.value && canDiscussPost(detail.value))

function discussPeek(preview?: PostDiscussPreview) {
  if (!preview) return ''
  const body = messagePreview(preview.contentType, preview.content).replace(/\s+/g, ' ').trim()
  if (!body) return ''
  const text = body.length > 28 ? `${body.slice(0, 28)}…` : body
  const name = preview.fromName.trim()
  return name ? `${name}：${text}` : text
}

const chatPeekText = computed(() => {
  if (!canDiscuss.value) return ''
  if (chatHistoryLoaded.value) return chatPreviewText.value
  return discussPeek(detail.value?.discussPreview)
})

async function load() {
  if (!postId.value) return
  loading.value = true
  try {
    const res = await postDetail(postId.value)
    detail.value = res.data
  } catch (e: unknown) {
    showCatchError(e, '加载失败', true)
  } finally {
    loading.value = false
  }
}

async function onSignup() {
  if (!postId.value || acting.value) return
  acting.value = true
  const wasJoined = hasJoined.value
  try {
    const res = await postSignup(postId.value)
    const status = res.data.status
    if (wasJoined) {
      showSuccessToast(signupStatusText(status) || '已报名')
    } else if (status === 2) {
      showSuccessToast('已加入候补')
    } else {
      showSuccessToast('报名成功')
    }
    await load()
  } catch (e: unknown) {
    showCatchError(e, '报名失败', true)
  } finally {
    acting.value = false
  }
}

async function onCancelSignup() {
  if (!postId.value || acting.value) return
  acting.value = true
  const wasWaitlisted = isWaitlisted.value
  try {
    await postSignupCancel(postId.value)
    showSuccessToast(wasWaitlisted ? '已取消候补' : '已取消报名')
    await load()
  } catch (e: unknown) {
    showCatchError(e, '取消失败', true)
  } finally {
    acting.value = false
  }
}

function onMoreSelect(action: NavMoreAction) {
  if (action.name !== ACTION_REPORT_CONTENT || !detail.value) return
  openPostReportSheet(router, postId.value, detail.value.title, detail.value.content)
}

function onShare() {
  if (!detail.value) return
  openForwardPost(router, {
    postId: postId.value,
    title: detail.value.title || '动态',
    cover: detail.value.images?.[0] || '',
    shareUrl: buildShareUrl({ type: 'post', id: postId.value }, 'link'),
    shareContent: detail.value.content || '',
  })
}

onMounted(() => {
  void load()
})
</script>

<template>
  <SubPageLayout full-height wide-title :on-back="onBack">
    <template #title>
      <div v-if="detail?.publisher" class="nav-author" @click.stop="openPublisherProfile">
        <UserAvatar :src="detail.publisher.avatarUrl" :size="28" tone="profile" />
        <span class="nav-author__name">{{ detail.publisher.nickname }}</span>
      </div>
    </template>
    <template v-if="detail && !isOwner" #right>
      <NavMoreButton :actions="moreActions" @select="onMoreSelect" />
    </template>

    <div class="post-detail">
      <div class="post-detail__scroll">
        <van-skeleton title :row="6" :loading="loading">
          <div v-if="detail" class="detail-stack">
            <section class="detail-card detail-card--hero">
              <h1 class="detail-title">{{ detail.title }}</h1>
            </section>

            <section class="detail-card">
              <PostImageGrid v-if="detail.images?.length" :images="detail.images" />
              <div class="detail-content">{{ detail.content }}</div>
              <div class="content-meta">
                <span v-if="detail.placeName" class="content-meta__place"
                  >位置 {{ detail.placeName }}</span
                >
                <span class="content-meta__time">{{ formatPostCreatedAt(detail.createdAt) }}</span>
              </div>
            </section>

            <PostDetailExtras :detail="detail" />
          </div>
        </van-skeleton>
      </div>

      <PostDetailFooter
        v-if="detail"
        :detail="detail"
        :has-joined="hasJoined"
        :is-waitlisted="isWaitlisted"
        :acting="acting"
        :chat-opening="chatOpening"
        :chat-peek-text="chatPeekText"
        @open-chat="openChat()"
        @share="onShare"
        @signup="onSignup"
        @cancel-signup="onCancelSignup"
      />
    </div>

    <PostChatSheet
      v-model:show="chatVisible"
      :post-id="postId"
      title="讨论"
      :online-label="chatOnlineLabel"
      :messages="chatMessages"
      :loading-more="chatHistoryLoading"
      :has-more="chatHistoryHasMore"
      :opening="chatOpening"
      @send="sendChat"
      @recall="recallChat"
      @load-more="loadMoreChat"
    />
    <ShareOpenAppBar
      v-if="showOpenApp && postId > 0"
      :download-href="downloadHref"
      @open-app="openApp"
    />
  </SubPageLayout>
</template>

<style scoped>
.nav-author {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  max-width: 100%;
  cursor: pointer;
}

.nav-author__name {
  font-size: var(--font-page-title);
  font-weight: 600;
  line-height: 1.3;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.post-detail {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.post-detail__scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  overscroll-behavior-y: none;
}

.detail-stack {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 12px 24px;
}

.detail-card {
  padding: 16px;
  border-radius: var(--radius-md);
  background: var(--card-bg);
}

.detail-card--hero {
  padding-bottom: 16px;
}

.detail-title {
  margin: 0;
  font-size: 20px;
  font-weight: 600;
  line-height: 1.45;
  color: var(--text-primary);
}

.detail-content {
  font-size: 15px;
  line-height: 1.75;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
}

.content-meta {
  margin-top: 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.content-meta__place {
  flex: 1;
  min-width: 0;
  font-size: 12px;
  line-height: 1.4;
  color: var(--text-secondary);
  word-break: break-word;
}

.content-meta__time {
  flex-shrink: 0;
  margin-left: auto;
  font-size: 12px;
  line-height: 1.4;
  color: var(--text-tertiary);
  white-space: nowrap;
}
</style>

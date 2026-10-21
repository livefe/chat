<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { showToast } from 'vant'

import MessagePanel from '@/components/chat/MessagePanel.vue'
import NavMoreButton from '@/components/NavMoreButton.vue'
import type { NavMoreAction } from '@/types/nav'
import type { MessageItem, MessageReplyQuote } from '@/api/message'
import { useChatStore } from '@/stores/chat'
import { useChatActions } from '@/composables/useChatActions'
import {
  ACTION_REPORT_USER,
  openMessageReportSheet,
  openUserReportSheet,
  reportUserAction,
} from '@/composables/useReportAction'
import { ACTION_BLOCK_USER, blockUserAction, confirmBlockUser } from '@/composables/useBlockAction'
import { useUserStore } from '@/stores/user'
import { openUserProfile } from '@/utils/user'
import { showCatchError } from '@/utils/toast'
import { isLocationPickNavigation } from '@/utils/map/locationPick'

const props = defineProps<{ peerId: string }>()
const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const chatStore = useChatStore()
const chatActions = useChatActions()
const entering = ref(true)

const peerIdNum = computed(() => Number(props.peerId))
const isSelfChat = computed(
  () => peerIdNum.value > 0 && peerIdNum.value === (userStore.info?.id ?? 0),
)
const moreActions = computed(() => (isSelfChat.value ? [] : [blockUserAction, reportUserAction]))

const peerTitle = computed(() => {
  const fromQuery = route.query.name
  if (typeof fromQuery === 'string' && fromQuery.trim()) return fromQuery.trim()

  const peerId = peerIdNum.value
  if (!peerId) return '私聊'

  const fromMsg = chatStore.getPrivateChatMessages(peerId).find((m) => m.fromUserId === peerId)
  if (fromMsg?.fromName) return fromMsg.fromName

  return `用户 ${peerId}`
})

const peerAvatar = computed(() => {
  const peerId = peerIdNum.value
  if (!peerId) return ''

  const fromMsg = chatStore.getPrivateChatMessages(peerId).find((m) => m.fromUserId === peerId)
  return fromMsg?.fromAvatar || ''
})

function profileQuery(name?: string, avatarUrl?: string) {
  return {
    from: 'private' as const,
    chatFrom: typeof route.query.from === 'string' ? route.query.from : 'messages',
    peerId: peerIdNum.value,
    peerName: peerTitle.value,
    name: name || peerTitle.value,
    avatarUrl: avatarUrl || peerAvatar.value,
    postId: typeof route.query.postId === 'string' ? route.query.postId : undefined,
  }
}

async function enterCurrentPeer() {
  if (!Number.isFinite(peerIdNum.value) || peerIdNum.value <= 0) {
    showToast('私聊对象无效')
    router.replace({ name: 'ConversationList' })
    return
  }
  chatStore.connect()
  entering.value = true
  try {
    await chatStore.enterPrivateChat(peerIdNum.value)
  } catch (e: unknown) {
    showCatchError(e, '加载私聊失败')
    router.replace(route.query.from === 'chat' ? '/chat' : { name: 'ConversationList' })
  } finally {
    entering.value = false
  }
}

onMounted(() => {
  void enterCurrentPeer()
})

watch(peerIdNum, (id, prev) => {
  if (!id || id === prev) return
  void enterCurrentPeer()
})

onUnmounted(() => {
  if (isLocationPickNavigation(router.currentRoute.value)) return
  chatStore.leavePrivateChat()
})

function onSend(content: string, contentType?: number, reply?: MessageReplyQuote | null) {
  void chatActions.send(content, contentType, reply)
}

function onBack() {
  const from = route.query.from
  if (from === 'profile') {
    router.replace({
      name: 'UserProfile',
      params: { id: String(peerIdNum.value) },
      query: {
        from: typeof route.query.profileFrom === 'string' ? route.query.profileFrom : 'messages',
        name: peerTitle.value,
        postId: route.query.postId,
      },
    })
    return
  }
  if (from === 'chat') {
    router.replace('/chat')
    return
  }
  if (from === 'post') {
    const postId = route.query.postId
    if (postId) {
      router.replace(`/post/${postId}`)
      return
    }
    router.back()
    return
  }
  router.replace({ name: 'ConversationList' })
}

function openPeerProfile() {
  openUserProfile(router, peerIdNum.value, profileQuery())
}

function onAvatarClick(msg: MessageItem) {
  if (!msg.fromUserId) return
  openUserProfile(router, msg.fromUserId, profileQuery(msg.fromName || '', msg.fromAvatar || ''))
}

function onMessageLongPress(msg: MessageItem) {
  openMessageReportSheet(router, msg)
}

function onMoreSelect(action: NavMoreAction) {
  if (action.name === ACTION_REPORT_USER) {
    openUserReportSheet(router, peerIdNum.value, peerTitle.value, peerAvatar.value)
    return
  }
  if (action.name === ACTION_BLOCK_USER) {
    void confirmBlockUser(peerIdNum.value, peerTitle.value)
  }
}
</script>

<template>
  <div class="private-page">
    <van-nav-bar left-arrow :border="false" @click-left="onBack">
      <template #title>
        <button type="button" class="peer-head" @click="openPeerProfile">
          <span class="peer-title">{{ peerTitle }}</span>
        </button>
      </template>
      <template #right>
        <NavMoreButton v-if="moreActions.length" :actions="moreActions" @select="onMoreSelect" />
      </template>
    </van-nav-bar>
    <MessagePanel
      class="private-panel"
      hide-sender-name
      :messages="chatStore.getPrivateChatMessages(peerIdNum)"
      :initial-loading="entering"
      :loading-more="chatStore.historyLoading"
      :has-more="chatStore.historyHasMore"
      :location-share="{ type: 'private', peerId: peerIdNum }"
      empty-text="开始聊天吧"
      @send="onSend"
      @recall="chatActions.recall"
      @load-more="chatActions.loadMoreHistory"
      @avatar-click="onAvatarClick"
      @message-long-press="onMessageLongPress"
    />
  </div>
</template>

<style scoped>
.private-page {
  display: flex;
  flex-direction: column;
  height: var(--app-vh, 100dvh);
  min-height: 0;
  overflow: hidden;
  box-sizing: border-box;
  background: var(--page-bg);

  --panel-bottom-inset: env(safe-area-inset-bottom, 0px);
}

.private-page :deep(.van-nav-bar) {
  flex-shrink: 0;
}

.peer-head {
  display: inline-flex;
  align-items: center;
  justify-content: flex-start;
  max-width: 100%;
  border: none;
  background: transparent;
  padding: 0;
  margin: 0;
  cursor: pointer;
  color: inherit;
  font: inherit;
}

.peer-title {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--font-page-title);
  font-weight: 600;
  line-height: 1.3;
  color: var(--text-primary);
}

.private-panel {
  flex: 1;
  min-height: 0;
}
</style>

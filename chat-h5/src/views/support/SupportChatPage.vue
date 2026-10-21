<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import MessagePanel from '@/components/chat/MessagePanel.vue'
import type { MessageItem } from '@/api/message'
import type { SupportMessage } from '@/api/support'
import { useSupportStore } from '@/stores/support'
import { useChatStore } from '@/stores/chat'
import { useUserStore } from '@/stores/user'
import { MessageContentTypes, MessageScope } from '@/constants/chat'
import { showCatchError } from '@/utils/toast'

const router = useRouter()
const support = useSupportStore()
const chatStore = useChatStore()
const userStore = useUserStore()
const loading = ref(true)

const SenderKindUser = 1
const SenderKindSystem = 3

function toChatMessage(m: SupportMessage): MessageItem {
  if (m.senderKind === SenderKindSystem) {
    return {
      id: m.id,
      scope: MessageScope.SystemHint,
      roomId: 0,
      postId: 0,
      fromUserId: 0,
      toUserId: 0,
      fromName: '',
      fromAvatar: '',
      contentType: MessageContentTypes.Text,
      content: m.content,
      latitude: 0,
      longitude: 0,
      createdAt: m.createdAt,
      recalled: m.recalled,
      clientMsgId: m.clientMsgId,
    }
  }

  const mine = m.senderKind === SenderKindUser
  const selfId = userStore.info?.id ?? 0
  return {
    id: m.id,
    scope: 0,
    roomId: 0,
    postId: 0,
    fromUserId: mine ? selfId || m.senderId : 0,
    toUserId: 0,
    fromName: m.senderName || '',
    fromAvatar: m.senderAvatarUrl || support.brandAvatar || '',
    contentType: m.contentType,
    content: m.content,
    latitude: 0,
    longitude: 0,
    createdAt: m.createdAt,
    recalled: m.recalled,
    clientMsgId: m.clientMsgId,
  }
}

const panelMessages = computed(() => support.messages.map(toChatMessage))

async function boot() {
  loading.value = true
  try {
    chatStore.connect()
    await support.loadProfile(true).catch(() => {})
    await support.enter()
  } catch (e) {
    showCatchError(e, '打开客服失败')
  } finally {
    loading.value = false
  }
}

async function onSend(content: string, contentType: number = MessageContentTypes.Text) {
  try {
    if (contentType === MessageContentTypes.Text) {
      await support.sendText(content)
    } else {
      await support.sendMedia(contentType, content)
    }
  } catch (e) {
    showCatchError(e, '发送失败')
  }
}

onMounted(() => {
  void boot()
})

onUnmounted(() => {
  support.leave()
})
</script>

<template>
  <div class="support-page">
    <van-nav-bar
      :title="support.brandName"
      left-arrow
      :border="false"
      @click-left="router.back()"
    />
    <MessagePanel
      class="support-panel"
      hide-sender-name
      :messages="panelMessages"
      :initial-loading="loading"
      :enable-actions="false"
      :enable-voice="false"
      :enable-location="false"
      placeholder="描述您的问题"
      empty-text="客服会尽快回复"
      @send="onSend"
    />
  </div>
</template>

<style scoped>
.support-page {
  display: flex;
  flex-direction: column;
  height: var(--app-vh, 100dvh);
  min-height: 0;
  overflow: hidden;
  box-sizing: border-box;
  background: var(--page-bg);
}

.support-panel {
  flex: 1;
  min-height: 0;
}
</style>

import { computed, onUnmounted, ref, watch, type Ref } from 'vue'

import type { MessageItem, MessageReplyQuote } from '@/api/message'
import { useChatActions } from '@/composables/useChatActions'
import { useChatStore } from '@/stores/chat'
import { messagePreview } from '@/utils/chat/chatMessageContent'
import { showCatchError } from '@/utils/toast'

export function usePostChat(postId: Ref<number>) {
  const chatStore = useChatStore()
  const chatActions = useChatActions()

  const visible = ref(false)
  const activePostId = ref(0)
  const opening = ref(false)
  const subscribed = ref(false)
  const historyLoaded = ref(false)

  const messages = computed(() => {
    return activePostId.value > 0 ? chatStore.getPostMessages(activePostId.value) : []
  })

  const historyLoading = computed(() =>
    activePostId.value > 0 ? chatStore.getPostHistory(activePostId.value).loading : false,
  )

  const historyHasMore = computed(() =>
    activePostId.value > 0 ? chatStore.getPostHistory(activePostId.value).hasMore : false,
  )

  const onlineCount = computed(() =>
    activePostId.value > 0 ? chatStore.onlineCountForPost(activePostId.value) : 0,
  )

  const onlineLabel = computed(() => `${onlineCount.value} 人在线`)

  const previewText = computed(() => {
    const msgs = messages.value
    for (let i = msgs.length - 1; i >= 0; i--) {
      const last = msgs[i]
      if (!last || last.id <= 0 || last.recalled) continue
      const name = last.fromName || ''
      const preview = messagePreview(last.contentType, last.content)
      const text = preview.length > 28 ? `${preview.slice(0, 28)}…` : preview
      return name ? `${name}：${text}` : text
    }
    return ''
  })

  async function attachFeed() {
    activePostId.value = postId.value
    chatStore.connect()
    if (!subscribed.value) {
      await chatStore.subscribePostFeed(activePostId.value)
      subscribed.value = true
    }
    await chatStore.refreshPostMessages(activePostId.value)
    historyLoaded.value = true
  }

  async function open() {
    if (!postId.value || opening.value) return
    opening.value = true
    try {
      await attachFeed()
      visible.value = true
    } catch (e: unknown) {
      showCatchError(e, '打开讨论失败', true)
      detachFeed()
    } finally {
      opening.value = false
    }
  }

  watch(postId, (id, prev) => {
    if (id === prev) return
    detachFeed()
    historyLoaded.value = false
    activePostId.value = 0
    visible.value = false
  })

  watch(visible, (open) => {
    if (!open) detachFeed()
  })

  function detachFeed() {
    if (!subscribed.value || activePostId.value <= 0) return
    chatStore.unsubscribePostFeed(activePostId.value)
    subscribed.value = false
  }

  function send(content: string, contentType?: number, reply?: MessageReplyQuote | null) {
    if (activePostId.value <= 0) return
    void chatActions.sendToPost(activePostId.value, content, contentType, reply)
  }

  async function loadMore() {
    if (activePostId.value <= 0) return
    await chatActions.loadMorePostMessages(activePostId.value)
  }

  function recall(msg: MessageItem) {
    void chatActions.recall(msg)
  }

  onUnmounted(() => {
    detachFeed()
  })

  return {
    visible,
    opening,
    messages,
    historyLoaded,
    onlineLabel,
    previewText,
    historyLoading,
    historyHasMore,
    open,
    send,
    loadMore,
    recall,
  }
}

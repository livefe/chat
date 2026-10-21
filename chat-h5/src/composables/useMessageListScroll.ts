import { nextTick, onDeactivated, onMounted, onUnmounted, ref, watch, type Ref } from 'vue'

import type { MessageItem } from '@/api/message'
import { messageStableKey } from '@/utils/chat/chatMessages'

const SCROLL_BOTTOM_THRESHOLD = 80
const SCROLL_TOP_THRESHOLD = 48
const INITIAL_STICK_LOCK_MS = 2000

function messageKey(m: MessageItem | undefined): string | number {
  if (!m) return ''
  return messageStableKey(m)
}

export interface MessageListScrollOptions {
  listRef: Ref<HTMLElement | null>
  messages: () => MessageItem[]
  hasMore: () => boolean
  loadingMore: () => boolean
  virtualEnabled: Ref<boolean>
  virtualTotal: () => number
  onVirtualScroll: () => void
  virtualScrollToBottom: () => void
  preserveScrollAfterPrepend: (previousTotal: number) => void
  getTotalHeight: () => number
  onLoadMore: () => void
  onDismiss: () => void
}

export function useMessageListScroll(opts: MessageListScrollOptions) {
  const stickToBottom = ref(true)

  let savedListScrollTop = 0
  let savedScrollHeight = 0
  let savedVirtualTotal = 0
  let loadMoreLocked = false

  let ignoreStickFromScroll = false
  let stickLockUntil = 0
  let stickPulseRaf = 0

  function isNearBottom(el: HTMLElement) {
    return el.scrollHeight - el.scrollTop - el.clientHeight < SCROLL_BOTTOM_THRESHOLD
  }

  function isStickLocked() {
    return Date.now() < stickLockUntil
  }

  function stopStickPulse() {
    if (!stickPulseRaf) return
    cancelAnimationFrame(stickPulseRaf)
    stickPulseRaf = 0
  }

  function releaseStickFromUser() {
    if (!isStickLocked() && !ignoreStickFromScroll) return
    stopStickPulse()
    stickLockUntil = 0
    ignoreStickFromScroll = false
  }

  function onPointerDown() {
    releaseStickFromUser()
    opts.onDismiss()
  }

  function scrollToBottom(scrollOpts?: { lockMs?: number }) {
    stickToBottom.value = true
    ignoreStickFromScroll = true
    if (scrollOpts?.lockMs && scrollOpts.lockMs > 0) {
      stickLockUntil = Math.max(stickLockUntil, Date.now() + scrollOpts.lockMs)
    }
    opts.virtualScrollToBottom()
    void nextTick(() => {
      opts.virtualScrollToBottom()
      if (!isStickLocked()) {
        requestAnimationFrame(() => {
          if (!isStickLocked()) ignoreStickFromScroll = false
        })
        return
      }
      stopStickPulse()
      const pulse = () => {
        opts.virtualScrollToBottom()
        if (isStickLocked()) {
          stickPulseRaf = requestAnimationFrame(pulse)
          return
        }
        stickPulseRaf = 0
        ignoreStickFromScroll = false
      }
      stickPulseRaf = requestAnimationFrame(pulse)
    })
  }

  function onScroll() {
    opts.onVirtualScroll()
    const el = opts.listRef.value
    if (!el) return

    if (isStickLocked() && !ignoreStickFromScroll && !isNearBottom(el)) {
      releaseStickFromUser()
      stickToBottom.value = false
    } else if (!ignoreStickFromScroll && !isStickLocked()) {
      stickToBottom.value = isNearBottom(el)
    }

    if (
      el.scrollTop <= SCROLL_TOP_THRESHOLD &&
      opts.hasMore() &&
      !opts.loadingMore() &&
      !loadMoreLocked
    ) {
      loadMoreLocked = true
      savedScrollHeight = el.scrollHeight
      savedVirtualTotal = opts.getTotalHeight()
      opts.onLoadMore()
    }
  }

  watch(opts.loadingMore, async (loading, wasLoading) => {
    if (loading) {
      savedScrollHeight = opts.listRef.value?.scrollHeight ?? 0
      savedVirtualTotal = opts.getTotalHeight()
      return
    }
    if (!wasLoading) return

    await nextTick()
    const el = opts.listRef.value
    if (!el) return

    if (opts.virtualEnabled.value) {
      opts.preserveScrollAfterPrepend(savedVirtualTotal)
    } else if (savedScrollHeight > 0) {
      el.scrollTop += el.scrollHeight - savedScrollHeight
    }

    savedScrollHeight = 0
    savedVirtualTotal = 0
    loadMoreLocked = false
  })

  watch(
    () => {
      const msgs = opts.messages()
      return {
        length: msgs.length,
        firstKey: messageKey(msgs[0]),
        lastKey: messageKey(msgs[msgs.length - 1]),
      }
    },
    async (curr, prev) => {
      await nextTick()
      if (!opts.listRef.value) return

      if (!prev?.length) {
        scrollToBottom({ lockMs: INITIAL_STICK_LOCK_MS })
        return
      }

      const prepended =
        curr.length > prev.length &&
        curr.firstKey !== prev.firstKey &&
        curr.lastKey === prev.lastKey
      if (prepended) return

      const appended = curr.lastKey !== prev.lastKey
      const reset = curr.firstKey !== prev.firstKey

      if ((appended || reset) && (stickToBottom.value || isStickLocked())) {
        scrollToBottom(isStickLocked() ? { lockMs: stickLockUntil - Date.now() } : undefined)
      }
    },
  )

  watch(opts.virtualTotal, () => {
    if (stickToBottom.value || isStickLocked()) {
      opts.virtualScrollToBottom()
    }
  })

  watch(opts.virtualEnabled, async (on) => {
    if (!on) return
    await nextTick()
    if (stickToBottom.value || isStickLocked()) {
      scrollToBottom({ lockMs: INITIAL_STICK_LOCK_MS })
    }
  })

  function onViewportResize() {
    if (!stickToBottom.value && !isStickLocked()) return
    void nextTick(() => scrollToBottom())
  }

  onMounted(() => {
    void nextTick(() => {
      if (opts.messages().length === 0) return
      scrollToBottom({ lockMs: INITIAL_STICK_LOCK_MS })
    })

    window.visualViewport?.addEventListener('resize', onViewportResize)
  })

  onUnmounted(() => {
    window.visualViewport?.removeEventListener('resize', onViewportResize)
    stopStickPulse()
  })

  onDeactivated(() => {
    stopStickPulse()
    const el = opts.listRef.value
    if (el) savedListScrollTop = el.scrollTop
  })

  function syncAfterActivate(syncOpts?: { forceStick?: boolean; lockMs?: number }) {
    void nextTick(async () => {
      if (opts.messages().length === 0) return
      await nextTick()
      const el = opts.listRef.value
      if (!el) return
      if (syncOpts?.forceStick || isStickLocked() || stickToBottom.value) {
        scrollToBottom({
          lockMs:
            syncOpts?.lockMs ??
            (isStickLocked()
              ? Math.max(INITIAL_STICK_LOCK_MS, stickLockUntil - Date.now())
              : INITIAL_STICK_LOCK_MS),
        })
        return
      }
      el.scrollTop = savedListScrollTop
      opts.onVirtualScroll()
    })
  }

  function requestStick() {
    if (stickToBottom.value || isStickLocked()) scrollToBottom()
  }

  return {
    messageKey,
    onScroll,
    onPointerDown,
    scrollToBottom,
    syncAfterActivate,
    requestStick,
    isStickLocked,
    stickLockUntil: () => stickLockUntil,
  }
}

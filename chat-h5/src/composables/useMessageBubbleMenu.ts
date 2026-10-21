import { computed, nextTick, ref, toValue, watch, type MaybeRefOrGetter } from 'vue'
import { showToast } from 'vant'

import type { MessageItem } from '@/api/message'
import { useLongPress } from '@/composables/useLongPress'
import { RECALL_WINDOW_SEC } from '@/constants/chat'
import { resolveMediaUrl } from '@/utils/media'
import { parseChatMedia } from '@/utils/chat/chatMedia'
import {
  isImageMessage,
  isVideoMessage,
  messageClipboardText,
} from '@/utils/chat/chatMessageContent'
import { copyText } from '@/utils/share/shareCore'

export interface MessageBubbleMenuEmitters {
  reply: (msg: MessageItem) => void
  forward: (msg: MessageItem) => void
  report: (msg: MessageItem) => void
  recall: (msg: MessageItem) => void
}

export interface MessageBubbleGestures {
  enabled: boolean
  onTouchStart: (e: TouchEvent) => void
  onTouchEnd: () => void
  onTouchMove: (e: TouchEvent) => void
  consumeClickIfLongPressed: (e?: Event) => boolean
}

export function useMessageBubbleMenu(
  message: MaybeRefOrGetter<MessageItem>,
  self: MaybeRefOrGetter<boolean>,
  emit: MessageBubbleMenuEmitters,
) {
  const menuOpen = ref(false)
  const stackRef = ref<HTMLElement | null>(null)
  const menuStyle = ref<Record<string, string>>({})
  let menuOpenedAt = 0

  function placeMenu() {
    const el = stackRef.value
    if (!el) return
    const r = el.getBoundingClientRect()
    const gap = 6
    const menuH = 40
    const spaceAbove = r.top
    const placeBelow = spaceAbove < menuH + gap + 8
    menuStyle.value = {
      left: `${r.left + r.width / 2}px`,
      top: placeBelow ? `${r.bottom + gap}px` : `${r.top - gap}px`,
      transform: placeBelow ? 'translate(-50%, 0)' : 'translate(-50%, -100%)',
    }
  }

  function closeMenu() {
    if (Date.now() - menuOpenedAt < 400) return
    menuOpen.value = false
  }

  const { onTouchStart, onTouchEnd, onTouchMove, consumeClickIfLongPressed } = useLongPress(() => {
    const msg = toValue(message)
    if (!(msg.id > 0) || msg.recalled) return
    menuOpenedAt = Date.now()
    menuOpen.value = true
    void nextTick(() => placeMenu())
  })

  const mediaSrc = computed(() => {
    const msg = toValue(message)
    if (!isImageMessage(msg.contentType) && !isVideoMessage(msg.contentType)) return ''
    return resolveMediaUrl(parseChatMedia(msg.content)?.url || '')
  })

  const canRecall = computed(() => {
    const msg = toValue(message)
    if (!toValue(self) || msg.recalled || !(msg.id > 0)) return false
    return Math.floor(Date.now() / 1000) - msg.createdAt <= RECALL_WINDOW_SEC
  })

  function onQuote() {
    menuOpen.value = false
    emit.reply(toValue(message))
  }

  function onForward() {
    menuOpen.value = false
    emit.forward(toValue(message))
  }

  async function onCopy() {
    menuOpen.value = false
    const msg = toValue(message)
    const text = mediaSrc.value || messageClipboardText(msg.contentType, msg.content)
    if (!text) {
      showToast('无可复制内容')
      return
    }
    try {
      await copyText(text)
      showToast('已复制')
    } catch {
      showToast('复制失败')
    }
  }

  function onReport() {
    menuOpen.value = false
    emit.report(toValue(message))
  }

  function onRecall() {
    menuOpen.value = false
    emit.recall(toValue(message))
  }

  watch(menuOpen, (open, _prev, onCleanup) => {
    if (!open) return
    const onResize = () => placeMenu()
    window.addEventListener('resize', onResize)
    onCleanup(() => window.removeEventListener('resize', onResize))
  })

  return {
    menuOpen,
    menuStyle,
    stackRef,
    closeMenu,
    onTouchStart,
    onTouchEnd,
    onTouchMove,
    consumeClickIfLongPressed,
    onCopy,
    onQuote,
    onForward,
    onReport,
    onRecall,
    canRecall,
  }
}

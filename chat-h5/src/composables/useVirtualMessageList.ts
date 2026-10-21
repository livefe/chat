import {
  computed,
  nextTick,
  onMounted,
  onUnmounted,
  ref,
  watch,
  type ComponentPublicInstance,
  type Ref,
} from 'vue'

import type { MessageItem } from '@/api/message'
import { messageStableKey } from '@/utils/chat/chatMessages'
import { isImageMessage, isLocationMessage, isVideoMessage } from '@/utils/chat/chatMessageContent'
import { shouldShowTimeDivider } from '@/utils/chat/chatTime'

const VIRTUAL_THRESHOLD = 36
const ESTIMATED_HEIGHT = 78
const ESTIMATED_PRIVATE_MESSAGE_HEIGHT = 56
const ESTIMATED_IMAGE_HEIGHT = 180
const ESTIMATED_LOCATION_HEIGHT = 72
const ESTIMATED_DIVIDER_HEIGHT = 37
const ROW_GAP = 10
const PRIVATE_ROW_GAP = 4
const OVERSCAN = 6
export const VIRTUAL_HEADER_HEIGHT = 32

type VirtualRow =
  | { kind: 'message'; message: MessageItem; messageIndex: number }
  | { kind: 'divider'; ts: number; messageIndex: number }

function messageKey(m: MessageItem | undefined): string | number {
  if (!m) return ''
  return messageStableKey(m)
}

export function virtualRowKey(row: VirtualRow): string | number {
  if (row.kind === 'message') return messageKey(row.message)
  return `divider-${row.messageIndex}-${row.ts}`
}

export function useVirtualMessageList(
  containerRef: Ref<HTMLElement | null>,
  messages: Ref<MessageItem[]>,
  options?: {
    showTimeDividers?: Ref<boolean>

    compactRows?: Ref<boolean>
  },
) {
  const scrollTop = ref(0)
  const viewportHeight = ref(0)
  const heightByKey = ref(new Map<string | number, number>())

  const enabled = computed(() => messages.value.length >= VIRTUAL_THRESHOLD)

  const rows = computed((): VirtualRow[] => {
    const msgs = messages.value
    const showDividers = options?.showTimeDividers?.value ?? true
    const out: VirtualRow[] = []

    for (let i = 0; i < msgs.length; i++) {
      const m = msgs[i]
      if (!m) continue

      if (
        showDividers &&
        shouldShowTimeDivider(msgs[i - 1]?.createdAt, m.createdAt, { isFirst: i === 0 })
      ) {
        out.push({ kind: 'divider', ts: m.createdAt, messageIndex: i })
      }
      out.push({ kind: 'message', message: m, messageIndex: i })
    }
    return out
  })

  const keys = computed(() => rows.value.map(virtualRowKey))

  function estimatedRowSize(row: VirtualRow): number {
    if (row.kind === 'divider') return ESTIMATED_DIVIDER_HEIGHT + PRIVATE_ROW_GAP
    const compact = options?.compactRows?.value ?? false
    const gap = compact ? PRIVATE_ROW_GAP : ROW_GAP
    if (isImageMessage(row.message.contentType) || isVideoMessage(row.message.contentType)) {
      return ESTIMATED_IMAGE_HEIGHT + gap
    }
    if (isLocationMessage(row.message.contentType)) return ESTIMATED_LOCATION_HEIGHT + gap
    const body = compact ? ESTIMATED_PRIVATE_MESSAGE_HEIGHT : ESTIMATED_HEIGHT
    return body + gap
  }

  function rowHeightAt(index: number) {
    const key = keys.value[index]
    const row = rows.value[index]
    if (key === undefined || row === undefined) return ESTIMATED_HEIGHT + ROW_GAP
    return heightByKey.value.get(key) ?? estimatedRowSize(row)
  }

  const layout = computed(() => {
    const tops: number[] = []
    let total = VIRTUAL_HEADER_HEIGHT
    for (let i = 0; i < rows.value.length; i++) {
      tops.push(total)
      total += rowHeightAt(i)
    }
    return { tops, total }
  })

  const visibleRange = computed(() => {
    const count = rows.value.length
    if (count === 0) return { start: 0, end: -1 }

    if (!enabled.value) {
      return { start: 0, end: count - 1 }
    }

    const { tops } = layout.value
    const viewTop = scrollTop.value
    const viewBottom = scrollTop.value + viewportHeight.value

    let start = 0
    let lo = 0
    let hi = count - 1
    while (lo < hi) {
      const mid = Math.floor((lo + hi) / 2)
      const top = tops[mid] ?? VIRTUAL_HEADER_HEIGHT
      const bottom = top + rowHeightAt(mid)
      if (bottom <= viewTop) lo = mid + 1
      else hi = mid
    }
    start = Math.max(0, lo - OVERSCAN)

    lo = start
    hi = count - 1
    while (lo < hi) {
      const mid = Math.ceil((lo + hi) / 2)
      if ((tops[mid] ?? VIRTUAL_HEADER_HEIGHT) >= viewBottom) hi = mid - 1
      else lo = mid
    }
    const end = Math.min(count - 1, lo + OVERSCAN)

    return { start, end }
  })

  const visibleRows = computed(() => {
    const { start, end } = visibleRange.value
    if (end < start) return []
    return rows.value.slice(start, end + 1).map((row, i) => ({
      row,
      index: start + i,
    }))
  })

  function setRowHeight(key: string | number, height: number, row: VirtualRow) {
    const compact = options?.compactRows?.value ?? false
    const gap = row.kind === 'divider' || compact ? PRIVATE_ROW_GAP : ROW_GAP
    const next = Math.round(height) + gap
    if (heightByKey.value.get(key) === next) return
    heightByKey.value.set(key, next)
    heightByKey.value = new Map(heightByKey.value)
  }

  const rowRoByKey = new Map<string | number, ResizeObserver>()
  const rowElByKey = new Map<string | number, HTMLElement>()
  const rowMetaByKey = new Map<string | number, VirtualRow>()

  function unbindRow(key: string | number) {
    rowRoByKey.get(key)?.disconnect()
    rowRoByKey.delete(key)
    rowElByKey.delete(key)
    rowMetaByKey.delete(key)
  }

  function unbindAllRows() {
    for (const key of [...rowRoByKey.keys()]) unbindRow(key)
  }

  function measureRow(key: string | number, node: HTMLElement, row: VirtualRow) {
    const h = node.offsetHeight
    if (h > 0) setRowHeight(key, h, row)
  }

  function setRowRef(
    key: string | number,
    row: VirtualRow,
    el: Element | ComponentPublicInstance | null,
  ) {
    if (!enabled.value) return
    if (!el) {
      unbindRow(key)
      return
    }
    const node = el instanceof Element ? el : el.$el
    if (!(node instanceof HTMLElement)) return

    rowMetaByKey.set(key, row)
    if (rowElByKey.get(key) === node) {
      measureRow(key, node, row)
      return
    }

    unbindRow(key)
    rowElByKey.set(key, node)
    rowMetaByKey.set(key, row)
    const ro = new ResizeObserver(() => {
      const current = rowElByKey.get(key)
      const meta = rowMetaByKey.get(key)
      if (current && meta) measureRow(key, current, meta)
    })
    ro.observe(node)
    rowRoByKey.set(key, ro)
    measureRow(key, node, row)
  }

  function syncScrollTop() {
    const el = containerRef.value
    if (!el) return
    scrollTop.value = el.scrollTop
  }

  function onScroll() {
    syncScrollTop()
  }

  function updateViewport() {
    const el = containerRef.value
    if (!el) return
    viewportHeight.value = el.clientHeight
  }

  function scrollToBottom() {
    const el = containerRef.value
    if (!el) return
    const apply = () => {
      el.scrollTop = el.scrollHeight
      syncScrollTop()
    }
    apply()
    void nextTick(() => {
      apply()
      requestAnimationFrame(apply)
    })
  }

  function preserveScrollAfterPrepend(prevTotal: number) {
    const el = containerRef.value
    if (!el) return
    void nextTick(() => {
      const delta = layout.value.total - prevTotal
      if (delta > 0) {
        el.scrollTop += delta
        syncScrollTop()
      }
    })
  }

  function getTotalHeight() {
    return layout.value.total
  }

  function rowTop(index: number) {
    return layout.value.tops[index] ?? VIRTUAL_HEADER_HEIGHT
  }

  let resizeObserver: ResizeObserver | null = null

  onMounted(() => {
    updateViewport()
    resizeObserver = new ResizeObserver(updateViewport)
    if (containerRef.value) resizeObserver.observe(containerRef.value)
  })

  onUnmounted(() => {
    resizeObserver?.disconnect()
    unbindAllRows()
  })

  watch(enabled, (on) => {
    if (!on) {
      unbindAllRows()
      return
    }
    void nextTick(updateViewport)
  })

  return {
    enabled,
    visibleRows,
    layout,
    bindRowRef: setRowRef,
    onScroll,
    scrollToBottom,
    preserveScrollAfterPrepend,
    getTotalHeight,
    rowTop,
  }
}

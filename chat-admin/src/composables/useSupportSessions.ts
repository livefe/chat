import { computed, onUnmounted, reactive, ref, type Ref } from 'vue'

import type { SupportAdminSession } from '@/api/support'
import { showApiError } from '@/utils/apiError'

const SUPPORT_PAGE_SIZE = 50

type SessionPage = { data: { total: number; list: SupportAdminSession[] } }
type LoadOptions = { silent?: boolean; append?: boolean }

export function supportSessionTitle(s: SupportAdminSession) {
  return s.userNickname || s.userMobile || `用户#${s.userId}`
}

export function useNowSec() {
  const now = ref(Math.floor(Date.now() / 1000))
  const timer = window.setInterval(() => {
    now.value = Math.floor(Date.now() / 1000)
  }, 1000)
  onUnmounted(() => window.clearInterval(timer))
  return now
}

export function supportWaitSec(s: SupportAdminSession, now: number) {
  return s.queueAt > 0 ? Math.max(0, now - s.queueAt) : 0
}

function sameSessionIds(a: SupportAdminSession[], b: SupportAdminSession[]) {
  return a.length === b.length && a.every((x, i) => x.id === b[i]?.id)
}

function applySessionPatch(cur: SupportAdminSession, item: SupportAdminSession) {
  Object.assign(cur, item)
}

function softMergeSessions(
  current: SupportAdminSession[],
  page1: SupportAdminSession[],
): SupportAdminSession[] {
  const byId = new Map(current.map((x) => [x.id, x]))
  const head = page1.map((item) => {
    const cur = byId.get(item.id)
    if (!cur) return item
    applySessionPatch(cur, item)
    return cur
  })
  if (current.length <= SUPPORT_PAGE_SIZE) return head
  const page1Ids = new Set(page1.map((x) => x.id))
  const oldTopIds = new Set(current.slice(0, SUPPORT_PAGE_SIZE).map((x) => x.id))
  const tail = current.filter((x) => !page1Ids.has(x.id) && !oldTopIds.has(x.id))
  return [...head, ...tail]
}

export function reselectActiveSession(
  active: Ref<SupportAdminSession | null>,
  list: SupportAdminSession[],
  when: (s: SupportAdminSession) => boolean,
) {
  const cur = active.value
  if (!cur?.id || !when(cur)) return
  const still = list.find((x) => x.id === cur.id)
  if (still) active.value = still
}

export function useSupportSessionList(options: {
  fetch: (params: { page: number; pageSize: number }) => Promise<SessionPage>
  errorText: string
  onReplaced?: (list: SupportAdminSession[]) => void
  onFetched?: (page: SupportAdminSession[]) => void
}) {
  const list = ref<SupportAdminSession[]>([])
  const total = ref(0)
  const loading = ref(false)
  const loadingMore = ref(false)
  const loaded = ref(false)
  const hasMore = computed(() => list.value.length < total.value)
  let seq = 0

  async function load(opts?: LoadOptions) {
    const silent = opts?.silent === true
    const append = opts?.append === true
    const cur = ++seq
    if (append) loadingMore.value = true
    else if (!silent) loading.value = true
    try {
      const page = append ? Math.floor(list.value.length / SUPPORT_PAGE_SIZE) + 1 : 1
      const { data } = await options.fetch({ page, pageSize: SUPPORT_PAGE_SIZE })
      if (cur !== seq && !append) return
      const fetched = data.list || []
      total.value = data.total || 0
      if (append) {
        const seen = new Set(list.value.map((x) => x.id))
        list.value.push(...fetched.filter((x) => !seen.has(x.id)))
      } else if (silent) {
        list.value = softMergeSessions(list.value, fetched)
        options.onReplaced?.(list.value)
      } else if (!sameSessionIds(list.value, fetched)) {
        list.value = fetched
        options.onReplaced?.(list.value)
      } else {
        list.value.forEach((x, i) => {
          const next = fetched[i]
          if (next) applySessionPatch(x, next)
        })
      }
      loaded.value = true
      options.onFetched?.(fetched)
    } catch (e) {
      if (!silent) showApiError(e, options.errorText)
    } finally {
      if (append) loadingMore.value = false
      else if (!silent) loading.value = false
    }
  }

  function loadMore() {
    if (!hasMore.value || loadingMore.value) return
    void load({ append: true })
  }

  function reset() {
    seq++
    list.value = []
    total.value = 0
    loaded.value = false
    loading.value = false
  }

  return reactive({ list, total, loading, loadingMore, loaded, hasMore, load, loadMore, reset })
}

export type SupportSessionList = ReturnType<typeof useSupportSessionList>

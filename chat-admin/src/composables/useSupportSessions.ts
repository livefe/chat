import { computed, reactive, ref, type Ref } from 'vue'

import type { SupportAdminSession } from '@/api/support'
import { showApiError } from '@/utils/apiError'

const SUPPORT_PAGE_SIZE = 50

type SessionPage = { data: { total: number; list: SupportAdminSession[] } }
type LoadOptions = { silent?: boolean; append?: boolean }

export function supportSessionTitle(s: SupportAdminSession) {
  return s.userNickname || s.userMobile || `用户#${s.userId}`
}

function sameSessionList(a: SupportAdminSession[], b: SupportAdminSession[]) {
  if (a.length !== b.length) return false
  for (let i = 0; i < a.length; i++) {
    const left = a[i]
    const right = b[i]
    if (!left || !right || left.id !== right.id) return false
  }
  return true
}

function mergeSessionFields(target: SupportAdminSession[], list: SupportAdminSession[]) {
  list.forEach((item, i) => {
    const cur = target[i]
    if (!cur) return
    cur.queueWaitSec = item.queueWaitSec
    cur.lastContent = item.lastContent
    cur.lastActiveAt = item.lastActiveAt
    cur.unreadCount = item.unreadCount
    cur.status = item.status
    cur.statusText = item.statusText
    cur.closeAt = item.closeAt
    cur.assigneeEmployeeId = item.assigneeEmployeeId
    cur.assigneeName = item.assigneeName
  })
}

function applySessionPatch(cur: SupportAdminSession, item: SupportAdminSession) {
  cur.queueWaitSec = item.queueWaitSec
  cur.lastContent = item.lastContent
  cur.lastActiveAt = item.lastActiveAt
  cur.unreadCount = item.unreadCount
  cur.status = item.status
  cur.statusText = item.statusText
  cur.closeAt = item.closeAt
  cur.userNickname = item.userNickname
  cur.userMobile = item.userMobile
  cur.userAvatarUrl = item.userAvatarUrl
  cur.assigneeEmployeeId = item.assigneeEmployeeId
  cur.assigneeName = item.assigneeName
}

function softMergeSessions(
  current: SupportAdminSession[],
  page1: SupportAdminSession[],
): SupportAdminSession[] {
  if (current.length <= SUPPORT_PAGE_SIZE) {
    return page1
  }
  const byId = new Map(current.map((x) => [x.id, x]))
  const page1Ids = new Set(page1.map((x) => x.id))
  const oldTopIds = new Set(current.slice(0, SUPPORT_PAGE_SIZE).map((x) => x.id))

  const head: SupportAdminSession[] = []
  for (const item of page1) {
    const cur = byId.get(item.id)
    if (cur) {
      applySessionPatch(cur, item)
      head.push(cur)
    } else {
      head.push(item)
    }
  }

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
  mode: 'merge' | 'replace'
  onReplaced?: (list: SupportAdminSession[]) => void
  onFetched?: (page: SupportAdminSession[]) => void
}) {
  const list = ref<SupportAdminSession[]>([])
  const total = ref(0)
  const loading = ref(false)
  const loadingMore = ref(false)
  const hasMore = computed(() => list.value.length < total.value)

  async function load(opts?: LoadOptions) {
    const silent = opts?.silent === true
    const append = opts?.append === true
    if (append) loadingMore.value = true
    else if (!silent) loading.value = true
    try {
      const page = append ? Math.floor(list.value.length / SUPPORT_PAGE_SIZE) + 1 : 1
      const { data } = await options.fetch({ page, pageSize: SUPPORT_PAGE_SIZE })
      const fetched = data.list || []
      total.value = data.total || 0
      if (append) {
        const seen = new Set(list.value.map((x) => x.id))
        list.value.push(...fetched.filter((x) => !seen.has(x.id)))
      } else if (options.mode === 'replace') {
        list.value = fetched
        options.onReplaced?.(list.value)
      } else if (silent) {
        list.value = softMergeSessions(list.value, fetched)
        options.onReplaced?.(list.value)
      } else if (!sameSessionList(list.value, fetched)) {
        list.value = fetched
        options.onReplaced?.(list.value)
      } else {
        mergeSessionFields(list.value, fetched)
      }
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

  return reactive({ list, total, loading, loadingMore, hasMore, load, loadMore })
}

export type SupportSessionList = ReturnType<typeof useSupportSessionList>

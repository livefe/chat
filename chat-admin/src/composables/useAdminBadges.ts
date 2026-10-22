import { computed, onBeforeUnmount, onMounted, reactive } from 'vue'

import { getPendingCounts, type AdminPendingCounts } from '@/api/ops'
import { useEmployeeStore } from '@/stores/employee'
import { PERM, type PermissionCode } from '@/constants/permission'
import { findMenuPath } from '@/utils/navigation'

/** 菜单有角标时，点击进入列表所带的筛选条件。 */
export const PENDING_QUERY: Partial<Record<PermissionCode, Record<string, string>>> = {
  [PERM.PostPost]: { status: '0', sort: 'asc' },
  [PERM.ChatRoom]: { status: '0' },
  [PERM.UserReport]: { status: '0' },
  [PERM.UserFeedback]: { status: '0' },
}

const counts = reactive<AdminPendingCounts>({
  imUnread: 0,
  supportQueue: 0,
  supportUnread: 0,
  reportPending: 0,
  feedbackPending: 0,
  roomAuditPending: 0,
  postAuditPending: 0,
})

/** 菜单编码 → 角标数字；客服工作台显示「排队 + 我的未读」。 */
const BADGE_SOURCES: [PermissionCode, () => number][] = [
  [PERM.ImChat, () => counts.imUnread],
  [PERM.ImSupport, () => counts.supportQueue + counts.supportUnread],
  [PERM.ImSupportOverview, () => counts.supportQueue],
  [PERM.UserReport, () => counts.reportPending],
  [PERM.UserFeedback, () => counts.feedbackPending],
  [PERM.ChatRoom, () => counts.roomAuditPending],
  [PERM.PostPost, () => counts.postAuditPending],
]

let subscribers = 0
let badgeTimer: ReturnType<typeof setInterval> | null = null
let inflight: Promise<void> | null = null
let lastFetchedAt = 0

const POLL_MS = 45_000
const MIN_REFRESH_GAP_MS = 8_000

function clearAll() {
  for (const key of Object.keys(counts) as (keyof AdminPendingCounts)[]) counts[key] = 0
}

async function refreshBadges(opts?: { force?: boolean }) {
  if (!useEmployeeStore().isAuthenticated) {
    clearAll()
    return
  }
  if (document.hidden && !opts?.force) return
  const now = Date.now()
  if (!opts?.force && lastFetchedAt > 0 && now - lastFetchedAt < MIN_REFRESH_GAP_MS) return
  if (inflight) return inflight

  inflight = (async () => {
    try {
      const { data } = await getPendingCounts()
      Object.assign(counts, data)
      lastFetchedAt = Date.now()
    } catch {
      // 角标刷新失败不打扰用户，下个轮询周期重试
    } finally {
      inflight = null
    }
  })()
  return inflight
}

function onFocus() {
  void refreshBadges()
}

function onVisibility() {
  if (!document.hidden) void refreshBadges()
}

function startPolling() {
  if (badgeTimer) return
  void refreshBadges({ force: true })
  badgeTimer = setInterval(() => void refreshBadges(), POLL_MS)
  window.addEventListener('focus', onFocus)
  document.addEventListener('visibilitychange', onVisibility)
}

function stopPolling() {
  if (!badgeTimer) return
  clearInterval(badgeTimer)
  badgeTimer = null
  window.removeEventListener('focus', onFocus)
  document.removeEventListener('visibilitychange', onVisibility)
}

export function useAdminBadges() {
  const employeeStore = useEmployeeStore()

  const badgeByPath = computed(() => {
    const out = new Map<string, { code: PermissionCode; count: number }>()
    for (const [code, read] of BADGE_SOURCES) {
      const path = findMenuPath(employeeStore.permissions, code)
      if (path) out.set(path, { code, count: read() })
    }
    return out
  })

  function badgeFor(path: string | undefined) {
    if (!path) return undefined
    return badgeByPath.value.get(path)?.count || undefined
  }

  /** 菜单有待办角标时返回对应的筛选条件，否则返回 undefined。 */
  function pendingQueryFor(path: string | undefined) {
    if (!path) return undefined
    const hit = badgeByPath.value.get(path)
    if (!hit?.count) return undefined
    return PENDING_QUERY[hit.code]
  }

  onMounted(() => {
    subscribers += 1
    startPolling()
  })

  onBeforeUnmount(() => {
    subscribers = Math.max(0, subscribers - 1)
    if (subscribers === 0) stopPolling()
  })

  return {
    badgeFor,
    pendingQueryFor,
    refreshBadges,
  }
}

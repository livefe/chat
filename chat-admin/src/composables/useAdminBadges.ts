import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

import { getPendingCounts } from '@/api/ops'
import { useEmployeeStore } from '@/stores/employee'
import { PERM, type PermissionCode } from '@/constants/permission'
import { findMenuPath } from '@/utils/navigation'

const imUnread = ref(0)
const supportQueue = ref(0)
const reportPending = ref(0)
const feedbackPending = ref(0)
const roomPending = ref(0)
const postPending = ref(0)

let subscribers = 0
let badgeTimer: ReturnType<typeof setInterval> | null = null
let focusBound = false
let visibilityBound = false

let inflight: Promise<void> | null = null
let lastFetchedAt = 0

const POLL_MS = 45_000

const MIN_REFRESH_GAP_MS = 8_000

export function useAdminBadges() {
  const employeeStore = useEmployeeStore()

  function pathByPermissionCode(code: PermissionCode) {
    return findMenuPath(employeeStore.permissions, code)
  }

  const imChatPath = computed(() => pathByPermissionCode(PERM.ImChat))
  const imSupportPath = computed(() => pathByPermissionCode(PERM.ImSupport))
  const imSupportOverviewPath = computed(() => pathByPermissionCode(PERM.ImSupportOverview))
  const reportPath = computed(() => pathByPermissionCode(PERM.UserReport))
  const feedbackPath = computed(() => pathByPermissionCode(PERM.UserFeedback))
  const roomPath = computed(() => pathByPermissionCode(PERM.ChatRoom))
  const postPath = computed(() => pathByPermissionCode(PERM.PostPost))

  function badgeFor(path: string | undefined) {
    if (!path) return undefined
    if (imChatPath.value && path === imChatPath.value) return imUnread.value || undefined
    if (imSupportPath.value && path === imSupportPath.value) return supportQueue.value || undefined
    if (imSupportOverviewPath.value && path === imSupportOverviewPath.value) {
      return supportQueue.value || undefined
    }
    if (reportPath.value && path === reportPath.value) return reportPending.value || undefined
    if (feedbackPath.value && path === feedbackPath.value) return feedbackPending.value || undefined
    if (roomPath.value && path === roomPath.value) return roomPending.value || undefined
    if (postPath.value && path === postPath.value) return postPending.value || undefined
    return undefined
  }

  function clearAll() {
    imUnread.value = 0
    supportQueue.value = 0
    reportPending.value = 0
    feedbackPending.value = 0
    roomPending.value = 0
    postPending.value = 0
  }

  async function refreshBadges(opts?: { force?: boolean }) {
    if (!employeeStore.isAuthenticated) {
      clearAll()
      return
    }

    if (typeof document !== 'undefined' && document.hidden && !opts?.force) {
      return
    }
    const now = Date.now()
    if (!opts?.force && lastFetchedAt > 0 && now - lastFetchedAt < MIN_REFRESH_GAP_MS) {
      return
    }
    if (inflight) return inflight

    inflight = (async () => {
      try {
        const { data } = await getPendingCounts()
        imUnread.value = data?.imUnread || 0
        supportQueue.value = data?.supportQueue || 0
        reportPending.value = data?.reportPending || 0
        feedbackPending.value = data?.feedbackPending || 0
        roomPending.value = data?.roomAuditPending || 0
        postPending.value = data?.postAuditPending || 0
        lastFetchedAt = Date.now()
      } catch {
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
    if (!focusBound) {
      window.addEventListener('focus', onFocus)
      focusBound = true
    }
    if (!visibilityBound) {
      document.addEventListener('visibilitychange', onVisibility)
      visibilityBound = true
    }
  }

  function stopPolling() {
    if (badgeTimer) {
      clearInterval(badgeTimer)
      badgeTimer = null
    }
    if (focusBound) {
      window.removeEventListener('focus', onFocus)
      focusBound = false
    }
    if (visibilityBound) {
      document.removeEventListener('visibilitychange', onVisibility)
      visibilityBound = false
    }
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
    refreshBadges,
  }
}
